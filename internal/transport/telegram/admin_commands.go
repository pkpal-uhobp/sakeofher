package telegramtransport

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"sakeofher/internal/domain"
)

type adminState string

const (
	adminStateNone                 adminState = ""
	adminStateWaitingCheckTGID     adminState = "waiting_check_tg_id"
	adminStateWaitingGrant         adminState = "waiting_grant"
	adminStateWaitingExtendDays    adminState = "waiting_extend_days"
	adminStateWaitingEnableTGID    adminState = "waiting_enable_tg_id"
	adminStateWaitingDisableTGID   adminState = "waiting_disable_tg_id"
	adminStateWaitingCancelTGID    adminState = "waiting_cancel_tg_id"
	adminStateWaitingTraffic       adminState = "waiting_traffic"
	adminStateWaitingAlias         adminState = "waiting_alias"
	adminStateWaitingMessage       adminState = "waiting_message"
	adminStateWaitingBlockTGID     adminState = "waiting_block_tg_id"
	adminStateWaitingUnblockTGID   adminState = "waiting_unblock_tg_id"
	adminStateWaitingBroadcastText adminState = "waiting_broadcast_text"
)

func (b *Bot) showAdminMenu(ctx context.Context, upd update) error {
	telegramID := updateTelegramID(upd)
	if !b.settings.isAdmin(telegramID) {
		return b.replyOrEdit(ctx, upd, "Доступ запрещён. Команда только для админов.", backToMenuKeyboard())
	}
	b.clearAdminState(telegramID)
	return b.replyOrEdit(ctx, upd, adminMenuMessage(), adminMenuKeyboard())
}

func (b *Bot) handleAdminCallback(ctx context.Context, upd update, action string) error {
	telegramID := updateTelegramID(upd)
	if !b.settings.isAdmin(telegramID) {
		return b.replyOrEdit(ctx, upd, "Доступ запрещён.", backToMenuKeyboard())
	}

	switch action {
	case "menu":
		return b.showAdminMenu(ctx, upd)
	case "stats":
		return b.adminStats(ctx, upd)
	case "users":
		return b.adminListUsers(ctx, upd, "")
	case "subs":
		return b.adminListSubscriptions(ctx, upd, "all")
	case "tariffs":
		return b.adminListTariffs(ctx, upd)
	case "stars":
		return b.adminStars(ctx, upd)
	case "check":
		b.setAdminState(telegramID, adminStateWaitingCheckTGID)
		return b.replyOrEdit(ctx, upd, "Введите Telegram ID пользователя.\nПример: 970706613", adminBackKeyboard())
	case "grant":
		b.setAdminState(telegramID, adminStateWaitingGrant)
		return b.replyOrEdit(ctx, upd, "Введите: TG_ID [tariff_id] [alias]\nПример: 970706613 1 @nickname\n\nЕсли tariff_id не указать, бот возьмёт первый активный тариф.", adminBackKeyboard())
	case "extend":
		b.setAdminState(telegramID, adminStateWaitingExtendDays)
		return b.replyOrEdit(ctx, upd, "Введите: TG_ID DAYS\nПример: 970706613 30", adminBackKeyboard())
	case "enable":
		b.setAdminState(telegramID, adminStateWaitingEnableTGID)
		return b.replyOrEdit(ctx, upd, "Введите TG ID пользователя, которому нужно включить последнюю подписку.", adminBackKeyboard())
	case "disable":
		b.setAdminState(telegramID, adminStateWaitingDisableTGID)
		return b.replyOrEdit(ctx, upd, "Введите TG ID пользователя, которому нужно отключить последнюю подписку.", adminBackKeyboard())
	case "cancel":
		b.setAdminState(telegramID, adminStateWaitingCancelTGID)
		return b.replyOrEdit(ctx, upd, "Введите TG ID пользователя, которому нужно отменить последнюю подписку.", adminBackKeyboard())
	case "traffic":
		b.setAdminState(telegramID, adminStateWaitingTraffic)
		return b.replyOrEdit(ctx, upd, "Введите: TG_ID GB\nПример: 970706613 300", adminBackKeyboard())
	case "alias":
		b.setAdminState(telegramID, adminStateWaitingAlias)
		return b.replyOrEdit(ctx, upd, "Введите: TG_ID alias\nПример: 970706613 @nickname", adminBackKeyboard())
	case "message":
		b.setAdminState(telegramID, adminStateWaitingMessage)
		return b.replyOrEdit(ctx, upd, "Введите: TG_ID текст сообщения\nПример: 970706613 Добрый день, ваша подписка обновлена.", adminBackKeyboard())
	case "broadcast":
		b.setAdminState(telegramID, adminStateWaitingBroadcastText)
		return b.replyOrEdit(ctx, upd, "Введите текст рассылки для всех активных пользователей.\n\nМожно отправлять обычный текст и ссылки. После ввода бот покажет предпросмотр и попросит подтверждение.", adminBackKeyboard())
	case "broadcast_confirm":
		return b.adminBroadcastConfirm(ctx, upd)
	case "broadcast_cancel":
		b.clearBroadcastDraft(telegramID)
		return b.replyOrEdit(ctx, upd, "Рассылка отменена.", adminMenuKeyboard())
	case "block":
		b.setAdminState(telegramID, adminStateWaitingBlockTGID)
		return b.replyOrEdit(ctx, upd, "Введите TG ID пользователя для блокировки.", adminBackKeyboard())
	case "unblock":
		b.setAdminState(telegramID, adminStateWaitingUnblockTGID)
		return b.replyOrEdit(ctx, upd, "Введите TG ID пользователя для разблокировки.", adminBackKeyboard())
	case "reset":
		b.clearAdminState(telegramID)
		return b.replyOrEdit(ctx, upd, "Действие сброшено.", adminMenuKeyboard())
	case "close":
		b.clearAdminState(telegramID)
		return b.replyOrEdit(ctx, upd, "Админ-панель закрыта.", backToMenuKeyboard())
	default:
		return b.showAdminMenu(ctx, upd)
	}
}

func (b *Bot) handleAdminStateInput(ctx context.Context, upd update) error {
	msg := upd.Message
	if msg == nil || msg.From == nil {
		return nil
	}
	adminID := msg.From.ID
	if !b.settings.isAdmin(adminID) {
		b.clearAdminState(adminID)
		return b.sendMessage(ctx, msg.Chat.ID, "Доступ запрещён.", backToMenuKeyboard())
	}
	text := strings.TrimSpace(msg.Text)
	state := b.getAdminState(adminID)
	b.clearAdminState(adminID)

	switch state {
	case adminStateWaitingCheckTGID:
		return b.adminCheckByText(ctx, upd, text)
	case adminStateWaitingGrant:
		return b.adminGrantByText(ctx, upd, text)
	case adminStateWaitingExtendDays:
		return b.adminExtendByText(ctx, upd, text)
	case adminStateWaitingEnableTGID:
		return b.adminSetSubscriptionStateByText(ctx, upd, text, "enable")
	case adminStateWaitingDisableTGID:
		return b.adminSetSubscriptionStateByText(ctx, upd, text, "disable")
	case adminStateWaitingCancelTGID:
		return b.adminSetSubscriptionStateByText(ctx, upd, text, "cancel")
	case adminStateWaitingTraffic:
		return b.adminTrafficByText(ctx, upd, text)
	case adminStateWaitingAlias:
		return b.adminAliasByText(ctx, upd, text)
	case adminStateWaitingMessage:
		return b.adminMessageByText(ctx, upd, text)
	case adminStateWaitingBlockTGID:
		return b.adminSetUserBlockStateByText(ctx, upd, text, true)
	case adminStateWaitingUnblockTGID:
		return b.adminSetUserBlockStateByText(ctx, upd, text, false)
	case adminStateWaitingBroadcastText:
		return b.adminBroadcastByText(ctx, upd, text)
	default:
		return b.showAdminMenu(ctx, upd)
	}
}

func (b *Bot) handleGrantCommand(ctx context.Context, upd update) error {
	return b.adminCommandWithArgs(ctx, upd, "/grant", "Пример: /grant 970706613 1 @nickname", b.adminGrantByText)
}

func (b *Bot) handleCheckCommand(ctx context.Context, upd update) error {
	return b.adminCommandWithArgs(ctx, upd, "/check", "Пример: /check 970706613", b.adminCheckByText)
}

func (b *Bot) handleExtendCommand(ctx context.Context, upd update) error {
	return b.adminCommandWithArgs(ctx, upd, "/extend", "Пример: /extend 970706613 30", b.adminExtendByText)
}

func (b *Bot) handleTrafficCommand(ctx context.Context, upd update) error {
	return b.adminCommandWithArgs(ctx, upd, "/traffic", "Пример: /traffic 970706613 300", b.adminTrafficByText)
}

func (b *Bot) handleEnableCommand(ctx context.Context, upd update) error {
	return b.adminCommandWithArgs(ctx, upd, "/enable", "Пример: /enable 970706613", func(ctx context.Context, upd update, args string) error {
		return b.adminSetSubscriptionStateByText(ctx, upd, args, "enable")
	})
}

func (b *Bot) handleDisableCommand(ctx context.Context, upd update) error {
	return b.adminCommandWithArgs(ctx, upd, "/disable", "Пример: /disable 970706613", func(ctx context.Context, upd update, args string) error {
		return b.adminSetSubscriptionStateByText(ctx, upd, args, "disable")
	})
}

func (b *Bot) handleCancelCommand(ctx context.Context, upd update) error {
	return b.adminCommandWithArgs(ctx, upd, "/cancel", "Пример: /cancel 970706613", func(ctx context.Context, upd update, args string) error {
		return b.adminSetSubscriptionStateByText(ctx, upd, args, "cancel")
	})
}

func (b *Bot) handleBlockCommand(ctx context.Context, upd update) error {
	return b.adminCommandWithArgs(ctx, upd, "/block", "Пример: /block 970706613", func(ctx context.Context, upd update, args string) error {
		return b.adminSetUserBlockStateByText(ctx, upd, args, true)
	})
}

func (b *Bot) handleUnblockCommand(ctx context.Context, upd update) error {
	return b.adminCommandWithArgs(ctx, upd, "/unblock", "Пример: /unblock 970706613", func(ctx context.Context, upd update, args string) error {
		return b.adminSetUserBlockStateByText(ctx, upd, args, false)
	})
}

func (b *Bot) handleAliasCommand(ctx context.Context, upd update) error {
	return b.adminCommandWithArgs(ctx, upd, "/alias", "Пример: /alias 970706613 @nickname", b.adminAliasByText)
}

func (b *Bot) handleBroadcastCommand(ctx context.Context, upd update) error {
	return b.adminCommandWithArgs(ctx, upd, "/broadcast", "Пример: /broadcast Текст рассылки для всех активных пользователей", b.adminBroadcastByText)
}

func (b *Bot) handleMessageCommand(ctx context.Context, upd update) error {
	return b.adminCommandWithArgs(ctx, upd, "/msg", "Пример: /msg 970706613 Текст сообщения", b.adminMessageByText)
}

func (b *Bot) handleUsersCommand(ctx context.Context, upd update) error {
	msg := upd.Message
	if msg == nil || msg.From == nil {
		return nil
	}
	if !b.settings.isAdmin(msg.From.ID) {
		return b.sendMessage(ctx, msg.Chat.ID, "Доступ запрещён. Команда только для админов.", backToMenuKeyboard())
	}
	query := strings.TrimSpace(strings.TrimPrefix(msg.Text, "/users"))
	return b.adminListUsers(ctx, upd, query)
}

func (b *Bot) handleSubsCommand(ctx context.Context, upd update) error {
	msg := upd.Message
	if msg == nil || msg.From == nil {
		return nil
	}
	if !b.settings.isAdmin(msg.From.ID) {
		return b.sendMessage(ctx, msg.Chat.ID, "Доступ запрещён. Команда только для админов.", backToMenuKeyboard())
	}
	status := strings.TrimSpace(strings.TrimPrefix(msg.Text, "/subs"))
	if status == "" {
		status = "all"
	}
	return b.adminListSubscriptions(ctx, upd, status)
}

func (b *Bot) handleAdminTariffsCommand(ctx context.Context, upd update) error {
	msg := upd.Message
	if msg == nil || msg.From == nil {
		return nil
	}
	if !b.settings.isAdmin(msg.From.ID) {
		return b.sendMessage(ctx, msg.Chat.ID, "Доступ запрещён. Команда только для админов.", backToMenuKeyboard())
	}
	return b.adminListTariffs(ctx, upd)
}

func (b *Bot) handleStarsCommand(ctx context.Context, upd update) error {
	msg := upd.Message
	if msg == nil || msg.From == nil {
		return nil
	}
	if !b.settings.isAdmin(msg.From.ID) {
		return b.sendMessage(ctx, msg.Chat.ID, "Доступ запрещён. Команда только для админов.", backToMenuKeyboard())
	}
	return b.adminStars(ctx, upd)
}

type adminTextHandler func(context.Context, update, string) error

func (b *Bot) adminCommandWithArgs(ctx context.Context, upd update, command string, usage string, handler adminTextHandler) error {
	msg := upd.Message
	if msg == nil || msg.From == nil {
		return nil
	}
	if !b.settings.isAdmin(msg.From.ID) {
		return b.sendMessage(ctx, msg.Chat.ID, "Доступ запрещён. Команда только для админов.", backToMenuKeyboard())
	}
	args := strings.TrimSpace(strings.TrimPrefix(msg.Text, command))
	if args == "" {
		return b.sendMessage(ctx, msg.Chat.ID, usage, adminMenuKeyboard())
	}
	return handler(ctx, upd, args)
}

func (b *Bot) adminStats(ctx context.Context, upd update) error {
	text, err := b.adminStatsText(ctx)
	if err != nil {
		text = "Не удалось собрать статистику: " + err.Error()
	}
	return b.replyOrEdit(ctx, upd, text, adminMenuKeyboard())
}

func (b *Bot) adminStatsText(ctx context.Context) (string, error) {
	usersAll, err := b.services.Users.List(ctx, domain.UserListInput{Limit: 1})
	if err != nil {
		return "", err
	}
	usersActive, _ := b.services.Users.List(ctx, domain.UserListInput{Status: domain.UserStatusActive, Limit: 1})
	usersBlocked, _ := b.services.Users.List(ctx, domain.UserListInput{Status: domain.UserStatusBlocked, Limit: 1})
	usersDeleted, _ := b.services.Users.List(ctx, domain.UserListInput{Status: domain.UserStatusDeleted, Limit: 1})

	subsAll, err := b.services.Subscriptions.List(ctx, domain.SubscriptionListInput{Limit: 1})
	if err != nil {
		return "", err
	}
	subsActive, _ := b.services.Subscriptions.List(ctx, domain.SubscriptionListInput{Status: domain.SubscriptionStatusActive, Limit: 1})
	subsExpired, _ := b.services.Subscriptions.List(ctx, domain.SubscriptionListInput{Status: domain.SubscriptionStatusExpired, Limit: 1})
	subsCancelled, _ := b.services.Subscriptions.List(ctx, domain.SubscriptionListInput{Status: domain.SubscriptionStatusCancelled, Limit: 1})

	return fmt.Sprintf(strings.TrimSpace(` Статистика

Пользователи:
• всего: %d
• активные: %d
• заблокированные: %d
• удалённые: %d

Подписки:
• всего: %d
• активные: %d
• истёкшие: %d
• отменённые: %d`),
		usersAll.Total,
		totalOrZero(usersActive),
		totalOrZero(usersBlocked),
		totalOrZero(usersDeleted),
		subsAll.Total,
		subsTotalOrZero(subsActive),
		subsTotalOrZero(subsExpired),
		subsTotalOrZero(subsCancelled),
	), nil
}

func (b *Bot) adminListUsers(ctx context.Context, upd update, query string) error {
	resp, err := b.services.Users.List(ctx, domain.UserListInput{Query: strings.TrimSpace(query), Limit: 10})
	if err != nil {
		return b.replyOrEdit(ctx, upd, "Не удалось загрузить пользователей: "+err.Error(), adminMenuKeyboard())
	}
	return b.replyOrEdit(ctx, upd, adminUsersListText(resp), adminMenuKeyboard())
}

func (b *Bot) adminListSubscriptions(ctx context.Context, upd update, statusRaw string) error {
	statusRaw = strings.ToLower(strings.TrimSpace(statusRaw))
	if statusRaw == "" || statusRaw == "all" || statusRaw == "все" {
		statusRaw = "all"
	}
	input := domain.SubscriptionListInput{Limit: 10}
	if statusRaw != "all" {
		input.Status = domain.SubscriptionStatus(statusRaw)
	}
	resp, err := b.services.Subscriptions.List(ctx, input)
	if err != nil {
		return b.replyOrEdit(ctx, upd, "Не удалось загрузить подписки: "+err.Error(), adminMenuKeyboard())
	}
	return b.replyOrEdit(ctx, upd, adminSubsListText(resp), adminMenuKeyboard())
}

func (b *Bot) adminListTariffs(ctx context.Context, upd update) error {
	items, err := b.services.Tariffs.ListAllWithPrices(ctx)
	if err != nil {
		return b.replyOrEdit(ctx, upd, "Не удалось загрузить тарифы: "+err.Error(), adminMenuKeyboard())
	}
	return b.replyOrEdit(ctx, upd, adminTariffsText(items), adminMenuKeyboard())
}

func (b *Bot) adminStars(ctx context.Context, upd update) error {
	balance, err := b.getStarBalance(ctx)
	if err != nil {
		return b.replyOrEdit(ctx, upd, "Не удалось получить баланс Stars через Bot API: "+err.Error(), adminMenuKeyboard())
	}
	return b.replyOrEdit(ctx, upd, fmt.Sprintf("Баланс Telegram Stars: %d", balance.Amount), adminMenuKeyboard())
}

func (b *Bot) adminCheckByText(ctx context.Context, upd update, text string) error {
	msg := upd.Message
	if msg == nil {
		return nil
	}
	tgID, err := firstInt64(text)
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "TG ID должен быть числом. Пример: 970706613", adminMenuKeyboard())
	}
	return b.adminCheckByTGID(ctx, upd, tgID)
}

func (b *Bot) adminCheckByTGID(ctx context.Context, upd update, tgID int64) error {
	text := ""
	user, userErr := b.services.Users.GetByTelegramID(ctx, tgID)
	if userErr == nil && user != nil {
		text += adminUserText(user) + "\n\n"
	} else if !errors.Is(userErr, domain.ErrNotFound) {
		text += "Пользователь: ошибка чтения: " + userErr.Error() + "\n\n"
	}

	sub, err := b.services.Subscriptions.GetLatestByTelegramID(ctx, tgID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			text += fmt.Sprintf("Подписка для %d не найдена.", tgID)
		} else {
			text += "Ошибка проверки подписки: " + err.Error()
		}
	} else {
		text += publicSubscriptionText(sub)
		if url := b.siteSubscriptionURL(ctx, sub); url != "" {
			text += "\n\nСсылка сайта:\n" + url
		}
	}
	return b.replyOrEdit(ctx, upd, text, adminUserActionsKeyboard(tgID))
}

func (b *Bot) adminGrantByText(ctx context.Context, upd update, text string) error {
	msg := upd.Message
	if msg == nil {
		return nil
	}
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return b.sendMessage(ctx, msg.Chat.ID, "Введите: TG_ID [tariff_id] [alias]", adminMenuKeyboard())
	}
	tgID, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil || tgID <= 0 {
		return b.sendMessage(ctx, msg.Chat.ID, "TG ID должен быть числом. Пример: 970706613 1 @nickname", adminMenuKeyboard())
	}

	var tariffID int64
	alias := ""
	for _, f := range fields[1:] {
		if id, err := strconv.ParseInt(f, 10, 64); err == nil && id > 0 && tariffID == 0 {
			tariffID = id
			continue
		}
		if alias == "" {
			alias = f
		}
	}

	sub, err := b.adminGrantOrRenew(ctx, tgID, tariffID, alias)
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "Ошибка выдачи: "+err.Error(), adminMenuKeyboard())
	}
	return b.sendMessage(ctx, msg.Chat.ID, fmt.Sprintf("Подписка для %d выдана/продлена.\n\n%s", tgID, publicSubscriptionText(sub)), adminUserActionsKeyboard(tgID))
}

func (b *Bot) adminGrantOrRenew(ctx context.Context, telegramID int64, tariffID int64, alias string) (*domain.PublicSubscription, error) {
	user, err := b.services.Users.GetOrCreateTelegramUser(ctx, domain.TelegramUserInput{TelegramID: telegramID})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(alias) != "" {
		cleanAlias := strings.TrimPrefix(strings.TrimSpace(alias), "@")
		user, err = b.services.Users.Update(ctx, user.ID, domain.UpdateUserInput{Alias: &cleanAlias, TelegramUsername: &cleanAlias})
		if err != nil {
			return nil, err
		}
	}

	var tariff *domain.Tariff
	if tariffID > 0 {
		tariff, err = b.services.Tariffs.GetByID(ctx, tariffID)
		if err != nil {
			return nil, err
		}
	} else {
		items, err := b.services.Tariffs.ListActiveWithPrices(ctx)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return nil, domain.ErrNotFound
		}
		tariff = &items[0].Tariff
	}

	trafficGB := domain.TrafficBytesToGB(tariff.TrafficLimitBytes)
	if trafficGB <= 0 {
		trafficGB = 100
	}
	sub, err := b.services.Subscriptions.CreateManual(ctx, domain.CreateManualSubscriptionInput{
		UserID:               user.ID,
		TariffID:             tariff.ID,
		TrafficLimitGB:       trafficGB,
		ActiveInternalSquads: b.settings.DefaultRemnaSquads,
	})
	if err != nil {
		return nil, err
	}
	if fixed, err := b.services.Subscriptions.Enable(ctx, sub.Subscription.ID); err == nil && fixed != nil {
		sub = fixed
	}
	return sub, nil
}

func (b *Bot) adminExtendByText(ctx context.Context, upd update, text string) error {
	msg := upd.Message
	if msg == nil {
		return nil
	}
	fields := strings.Fields(text)
	if len(fields) < 2 {
		return b.sendMessage(ctx, msg.Chat.ID, "Введите: TG_ID DAYS\nПример: 970706613 30", adminMenuKeyboard())
	}
	tgID, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil || tgID <= 0 {
		return b.sendMessage(ctx, msg.Chat.ID, "TG ID должен быть числом.", adminMenuKeyboard())
	}
	days, err := strconv.Atoi(fields[1])
	if err != nil || days <= 0 {
		return b.sendMessage(ctx, msg.Chat.ID, "DAYS должен быть положительным числом. Пример: 30", adminMenuKeyboard())
	}
	sub, err := b.services.Subscriptions.GetLatestByTelegramID(ctx, tgID)
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "Подписка не найдена: "+err.Error(), adminMenuKeyboard())
	}
	updated, err := b.services.Subscriptions.Extend(ctx, sub.Subscription.ID, domain.ExtendSubscriptionInput{Days: &days, ActiveInternalSquads: b.settings.DefaultRemnaSquads})
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "Ошибка продления: "+err.Error(), adminMenuKeyboard())
	}
	if fixed, err := b.services.Subscriptions.Enable(ctx, updated.Subscription.ID); err == nil && fixed != nil {
		updated = fixed
	}
	return b.sendMessage(ctx, msg.Chat.ID, fmt.Sprintf("Подписка %d продлена на %d дн.\n\n%s", tgID, days, publicSubscriptionText(updated)), adminUserActionsKeyboard(tgID))
}

func (b *Bot) adminTrafficByText(ctx context.Context, upd update, text string) error {
	msg := upd.Message
	if msg == nil {
		return nil
	}
	fields := strings.Fields(text)
	if len(fields) < 2 {
		return b.sendMessage(ctx, msg.Chat.ID, "Введите: TG_ID GB\nПример: 970706613 300", adminMenuKeyboard())
	}
	tgID, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil || tgID <= 0 {
		return b.sendMessage(ctx, msg.Chat.ID, "TG ID должен быть числом.", adminMenuKeyboard())
	}
	gb, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil || gb <= 0 {
		return b.sendMessage(ctx, msg.Chat.ID, "GB должен быть положительным числом.", adminMenuKeyboard())
	}
	sub, err := b.services.Subscriptions.GetLatestByTelegramID(ctx, tgID)
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "Подписка не найдена: "+err.Error(), adminMenuKeyboard())
	}
	updated, err := b.services.Subscriptions.UpdateTrafficLimit(ctx, sub.Subscription.ID, domain.UpdateTrafficLimitInput{TrafficLimitGB: gb})
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "Ошибка изменения трафика: "+err.Error(), adminMenuKeyboard())
	}
	return b.sendMessage(ctx, msg.Chat.ID, fmt.Sprintf("Лимит трафика для %d изменён на %d ГБ.\n\n%s", tgID, gb, publicSubscriptionText(updated)), adminUserActionsKeyboard(tgID))
}

func (b *Bot) adminSetSubscriptionStateByText(ctx context.Context, upd update, text string, action string) error {
	msg := upd.Message
	if msg == nil {
		return nil
	}
	tgID, err := firstInt64(text)
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "TG ID должен быть числом.", adminMenuKeyboard())
	}
	sub, err := b.services.Subscriptions.GetLatestByTelegramID(ctx, tgID)
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "Подписка не найдена: "+err.Error(), adminMenuKeyboard())
	}
	var updated *domain.PublicSubscription
	switch action {
	case "enable":
		updated, err = b.services.Subscriptions.Enable(ctx, sub.Subscription.ID)
	case "disable":
		updated, err = b.services.Subscriptions.Disable(ctx, sub.Subscription.ID)
	case "cancel":
		updated, err = b.services.Subscriptions.Cancel(ctx, sub.Subscription.ID)
	default:
		err = domain.ErrInvalidInput
	}
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "Ошибка операции: "+err.Error(), adminMenuKeyboard())
	}
	return b.sendMessage(ctx, msg.Chat.ID, fmt.Sprintf("Операция %s выполнена для %d.\n\n%s", action, tgID, publicSubscriptionText(updated)), adminUserActionsKeyboard(tgID))
}

func (b *Bot) adminSetUserBlockStateByText(ctx context.Context, upd update, text string, block bool) error {
	msg := upd.Message
	if msg == nil {
		return nil
	}
	tgID, err := firstInt64(text)
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "TG ID должен быть числом.", adminMenuKeyboard())
	}
	user, err := b.services.Users.GetByTelegramID(ctx, tgID)
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "Пользователь не найден: "+err.Error(), adminMenuKeyboard())
	}
	if block {
		user, err = b.services.Users.Block(ctx, user.ID)
	} else {
		user, err = b.services.Users.Unblock(ctx, user.ID)
	}
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "Ошибка изменения статуса пользователя: "+err.Error(), adminMenuKeyboard())
	}
	action := "разблокирован"
	if block {
		action = "заблокирован"
	}
	return b.sendMessage(ctx, msg.Chat.ID, fmt.Sprintf("Пользователь %d %s.\n\n%s", tgID, action, adminUserText(user)), adminUserActionsKeyboard(tgID))
}

func (b *Bot) adminAliasByText(ctx context.Context, upd update, text string) error {
	msg := upd.Message
	if msg == nil {
		return nil
	}
	fields := strings.Fields(text)
	if len(fields) < 2 {
		return b.sendMessage(ctx, msg.Chat.ID, "Введите: TG_ID alias", adminMenuKeyboard())
	}
	tgID, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil || tgID <= 0 {
		return b.sendMessage(ctx, msg.Chat.ID, "TG ID должен быть числом.", adminMenuKeyboard())
	}
	alias := strings.TrimPrefix(strings.TrimSpace(fields[1]), "@")
	if alias == "" {
		return b.sendMessage(ctx, msg.Chat.ID, "Alias не должен быть пустым.", adminMenuKeyboard())
	}
	user, err := b.services.Users.GetByTelegramID(ctx, tgID)
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "Пользователь не найден: "+err.Error(), adminMenuKeyboard())
	}
	user, err = b.services.Users.Update(ctx, user.ID, domain.UpdateUserInput{Alias: &alias, TelegramUsername: &alias})
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "Ошибка обновления alias: "+err.Error(), adminMenuKeyboard())
	}
	return b.sendMessage(ctx, msg.Chat.ID, " Alias обновлён.\n\n"+adminUserText(user), adminUserActionsKeyboard(tgID))
}

func (b *Bot) adminMessageByText(ctx context.Context, upd update, text string) error {
	msg := upd.Message
	if msg == nil {
		return nil
	}
	fields := strings.Fields(text)
	if len(fields) < 2 {
		return b.sendMessage(ctx, msg.Chat.ID, "Введите: TG_ID текст сообщения", adminMenuKeyboard())
	}
	tgID, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil || tgID <= 0 {
		return b.sendMessage(ctx, msg.Chat.ID, "TG ID должен быть числом.", adminMenuKeyboard())
	}
	body := strings.TrimSpace(strings.TrimPrefix(text, fields[0]))
	if body == "" {
		return b.sendMessage(ctx, msg.Chat.ID, "Текст сообщения пустой.", adminMenuKeyboard())
	}
	if err := b.sendMessage(ctx, tgID, body, nil); err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, "Не удалось отправить сообщение пользователю: "+err.Error(), adminMenuKeyboard())
	}
	return b.sendMessage(ctx, msg.Chat.ID, fmt.Sprintf("Сообщение отправлено пользователю %d.", tgID), adminUserActionsKeyboard(tgID))
}

func (b *Bot) adminBroadcastByText(ctx context.Context, upd update, text string) error {
	msg := upd.Message
	if msg == nil || msg.From == nil {
		return nil
	}
	adminID := msg.From.ID
	body := strings.TrimSpace(text)
	if body == "" {
		return b.sendMessage(ctx, msg.Chat.ID, "Текст рассылки пустой.", adminMenuKeyboard())
	}
	if len([]rune(body)) > 3500 {
		return b.sendMessage(ctx, msg.Chat.ID, "Текст слишком длинный. Для безопасной отправки сделай до 3500 символов.", adminMenuKeyboard())
	}
	b.setBroadcastDraft(adminID, broadcastDraft{Text: body, Audience: "active", CreatedAt: time.Now()})
	preview := fmt.Sprintf(strings.TrimSpace(` Предпросмотр рассылки

Аудитория: все активные пользователи

Сообщение:
%s

Отправить?`), body)
	return b.sendMessage(ctx, msg.Chat.ID, preview, broadcastConfirmKeyboard())
}

func (b *Bot) adminBroadcastConfirm(ctx context.Context, upd update) error {
	adminID := updateTelegramID(upd)
	if !b.settings.isAdmin(adminID) {
		return b.replyOrEdit(ctx, upd, "Доступ запрещён.", backToMenuKeyboard())
	}
	draft, ok := b.getBroadcastDraft(adminID)
	if !ok || strings.TrimSpace(draft.Text) == "" {
		return b.replyOrEdit(ctx, upd, "Черновик рассылки не найден или устарел. Создайте рассылку заново через /broadcast текст.", adminMenuKeyboard())
	}
	if time.Since(draft.CreatedAt) > 30*time.Minute {
		b.clearBroadcastDraft(adminID)
		return b.replyOrEdit(ctx, upd, "Черновик рассылки устарел. Создайте рассылку заново.", adminMenuKeyboard())
	}
	b.clearBroadcastDraft(adminID)
	chatID := updateChatID(upd)
	if err := b.replyOrEdit(ctx, upd, " Рассылка запущена. Я пришлю итог после завершения.", adminMenuKeyboard()); err != nil {
		return err
	}
	go b.runBroadcast(context.Background(), adminID, chatID, draft)
	return nil
}

type broadcastResult struct {
	Total   int64
	Sent    int64
	Failed  int64
	Skipped int64
}

func (b *Bot) runBroadcast(ctx context.Context, adminID int64, adminChatID int64, draft broadcastDraft) {
	result := broadcastResult{}
	limit := 200
	offset := 0
	for {
		resp, err := b.services.Users.List(ctx, domain.UserListInput{Status: domain.UserStatusActive, Limit: limit, Offset: offset})
		if err != nil {
			_ = b.sendMessage(ctx, adminChatID, " Рассылка остановлена: не удалось получить пользователей: "+err.Error(), adminMenuKeyboard())
			return
		}
		if result.Total == 0 {
			result.Total = resp.Total
		}
		if len(resp.Items) == 0 {
			break
		}
		for _, user := range resp.Items {
			if user.TelegramID <= 0 || user.TelegramID == adminID {
				result.Skipped++
				continue
			}
			if err := b.sendMessage(ctx, user.TelegramID, draft.Text, nil); err != nil {
				result.Failed++
				b.log.Warn("broadcast send failed", zap.Int64("telegram_id", user.TelegramID), zap.Error(err))
			} else {
				result.Sent++
			}
			if b.settings.BroadcastDelay > 0 {
				time.Sleep(b.settings.BroadcastDelay)
			}
		}
		offset += len(resp.Items)
		if offset >= int(resp.Total) {
			break
		}
		if result.Sent > 0 && result.Sent%50 == 0 {
			_ = b.sendMessage(ctx, adminChatID, fmt.Sprintf("Рассылка в процессе: отправлено %d, ошибок %d.", result.Sent, result.Failed), nil)
		}
	}
	text := fmt.Sprintf(strings.TrimSpace(` Рассылка завершена

Аудитория: активные пользователи
Всего в базе: %d
Отправлено: %d
Ошибок: %d
Пропущено: %d`), result.Total, result.Sent, result.Failed, result.Skipped)
	_ = b.sendMessage(ctx, adminChatID, text, adminMenuKeyboard())
}

func firstInt64(text string) (int64, error) {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return 0, errors.New("empty")
	}
	id, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func totalOrZero(resp *domain.UserListResponse) int64 {
	if resp == nil {
		return 0
	}
	return resp.Total
}

func subsTotalOrZero(resp *domain.SubscriptionListResponse) int64 {
	if resp == nil {
		return 0
	}
	return resp.Total
}

func userDisplayName(u domain.User) string {
	parts := make([]string, 0, 3)
	if u.TelegramUsername != nil && *u.TelegramUsername != "" {
		parts = append(parts, "@"+strings.TrimPrefix(*u.TelegramUsername, "@"))
	}
	if u.Alias != nil && *u.Alias != "" {
		parts = append(parts, "alias:"+*u.Alias)
	}
	name := strings.TrimSpace(strings.TrimSpace(ptrString(u.TelegramFirstName)) + " " + strings.TrimSpace(ptrString(u.TelegramLastName)))
	if name != "" {
		parts = append(parts, name)
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, " / ")
}

func adminUserText(u *domain.User) string {
	if u == nil {
		return "Пользователь не найден."
	}
	remnaUUID := "—"
	if u.RemnaUUID != nil && strings.TrimSpace(*u.RemnaUUID) != "" {
		remnaUUID = *u.RemnaUUID
	}
	remnaUsername := "—"
	if u.RemnaUsername != nil && strings.TrimSpace(*u.RemnaUsername) != "" {
		remnaUsername = *u.RemnaUsername
	}
	return fmt.Sprintf(strings.TrimSpace(` Пользователь

TG ID: %d
Имя: %s
Статус: %s
Remnawave: %s
Remna username: %s
Remna UUID: %s
Создан: %s
Последний вход: %s`),
		u.TelegramID,
		userDisplayName(*u),
		u.Status,
		u.RemnaStatus,
		remnaUsername,
		remnaUUID,
		u.CreatedAt.Format("02.01.2006 15:04"),
		formatTimePtr(u.LastSeenAt),
	)
}

func adminUsersListText(resp *domain.UserListResponse) string {
	if resp == nil || len(resp.Items) == 0 {
		return "Пользователи не найдены."
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Пользователи: %d всего\n\n", resp.Total))
	for _, u := range resp.Items {
		b.WriteString(fmt.Sprintf("• %d — %s — %s — remna:%s\n", u.TelegramID, userDisplayName(u), u.Status, u.RemnaStatus))
	}
	b.WriteString("\nКоманды: /check TG_ID, /grant TG_ID [tariff_id], /block TG_ID")
	return b.String()
}

func adminSubsListText(resp *domain.SubscriptionListResponse) string {
	if resp == nil || len(resp.Items) == 0 {
		return "Подписки не найдены."
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Подписки: %d всего\n\n", resp.Total))
	for _, item := range resp.Items {
		b.WriteString(fmt.Sprintf("• sub#%d / TG %d — %s — до %s — %s — %s\n",
			item.Subscription.ID,
			item.User.TelegramID,
			item.Subscription.Status,
			item.Subscription.ExpiresAt.Format("02.01.2006"),
			trafficLabel(item.Subscription.TrafficLimitBytes),
			valueOrDash(item.Tariff.Title),
		))
	}
	b.WriteString("\nКоманды: /enable TG_ID, /disable TG_ID, /extend TG_ID DAYS")
	return b.String()
}

func adminTariffsText(items []domain.TariffWithPrices) string {
	if len(items) == 0 {
		return "Тарифы не найдены."
	}
	var b strings.Builder
	b.WriteString("Тарифы\n\n")
	for _, item := range items {
		state := "выкл"
		if item.IsActive {
			state = "активен"
		}
		b.WriteString(fmt.Sprintf("• ID %d — %s — %s — %s — %s\n", item.ID, item.Title, state, durationLabel(item.DurationDays), trafficLabel(item.TrafficLimitBytes)))
		if len(item.Prices) > 0 {
			b.WriteString("" + tariffPricesLabel(item) + "\n")
		}
	}
	b.WriteString("\nДля выдачи: /grant TG_ID tariff_id")
	return b.String()
}

func ptrString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func formatTimePtr(t *time.Time) string {
	if t == nil || t.IsZero() {
		return "—"
	}
	return t.Format("02.01.2006 15:04")
}

func (b *Bot) hasAdminState(telegramID int64) bool {
	return b.getAdminState(telegramID) != adminStateNone
}

func (b *Bot) getAdminState(telegramID int64) adminState {
	b.stateMu.RLock()
	defer b.stateMu.RUnlock()
	return b.states[telegramID]
}

func (b *Bot) setAdminState(telegramID int64, state adminState) {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	if state == adminStateNone {
		delete(b.states, telegramID)
		return
	}
	b.states[telegramID] = state
}

func (b *Bot) clearAdminState(telegramID int64) {
	b.setAdminState(telegramID, adminStateNone)
}

func (b *Bot) handleAdminUserCallback(ctx context.Context, upd update, raw string) error {
	telegramID := updateTelegramID(upd)
	if !b.settings.isAdmin(telegramID) {
		return b.replyOrEdit(ctx, upd, "Доступ запрещён.", backToMenuKeyboard())
	}
	parts := strings.Split(strings.TrimSpace(raw), ":")
	if len(parts) != 2 {
		return b.showAdminMenu(ctx, upd)
	}
	action := parts[0]
	tgID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || tgID <= 0 {
		return b.replyOrEdit(ctx, upd, "Некорректный TG ID.", adminMenuKeyboard())
	}
	switch action {
	case "check":
		return b.adminCheckByTGID(ctx, upd, tgID)
	case "grant":
		sub, err := b.adminGrantOrRenew(ctx, tgID, 0, "")
		if err != nil {
			return b.replyOrEdit(ctx, upd, "Ошибка выдачи: "+err.Error(), adminUserActionsKeyboard(tgID))
		}
		return b.replyOrEdit(ctx, upd, fmt.Sprintf("Подписка для %d выдана/продлена.\n\n%s", tgID, publicSubscriptionText(sub)), adminUserActionsKeyboard(tgID))
	case "enable", "disable", "cancel":
		sub, err := b.services.Subscriptions.GetLatestByTelegramID(ctx, tgID)
		if err != nil {
			return b.replyOrEdit(ctx, upd, "Подписка не найдена: "+err.Error(), adminUserActionsKeyboard(tgID))
		}
		var updated *domain.PublicSubscription
		switch action {
		case "enable":
			updated, err = b.services.Subscriptions.Enable(ctx, sub.Subscription.ID)
		case "disable":
			updated, err = b.services.Subscriptions.Disable(ctx, sub.Subscription.ID)
		case "cancel":
			updated, err = b.services.Subscriptions.Cancel(ctx, sub.Subscription.ID)
		}
		if err != nil {
			return b.replyOrEdit(ctx, upd, "Ошибка операции: "+err.Error(), adminUserActionsKeyboard(tgID))
		}
		return b.replyOrEdit(ctx, upd, fmt.Sprintf("Операция %s выполнена для %d.\n\n%s", action, tgID, publicSubscriptionText(updated)), adminUserActionsKeyboard(tgID))
	default:
		return b.showAdminMenu(ctx, upd)
	}
}

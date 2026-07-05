package telegramtransport

import (
	"context"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"sakeofher/internal/domain"
)

type Router struct {
	bot *Bot
}

func NewRouter(bot *Bot) *Router {
	return &Router{bot: bot}
}

func (r *Router) Handle(ctx context.Context, upd update) error {
	if upd.PreCheckoutQuery != nil {
		return r.handlePreCheckout(ctx, upd)
	}
	if upd.Message != nil {
		return r.handleMessage(ctx, upd)
	}
	if upd.CallbackQuery != nil {
		return r.handleCallback(ctx, upd)
	}

	return nil
}

func (r *Router) handlePreCheckout(ctx context.Context, upd update) error {
	q := upd.PreCheckoutQuery
	if q == nil {
		return nil
	}

	return r.bot.answerPreCheckout(ctx, q.ID, true, "")
}

func (r *Router) handleMessage(ctx context.Context, upd update) error {
	msg := upd.Message
	if msg == nil || msg.From == nil {
		return nil
	}

	if _, err := r.ensureUser(ctx, *msg.From); err != nil {
		return err
	}

	if msg.SuccessfulPayment != nil {
		return r.bot.handleSuccessfulPayment(ctx, upd)
	}

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return r.bot.showMainMenu(ctx, upd)
	}

	if r.bot.hasAdminState(msg.From.ID) {
		return r.bot.handleAdminStateInput(ctx, upd)
	}

	cmd := text
	if idx := strings.Index(cmd, " "); idx >= 0 {
		cmd = cmd[:idx]
	}

	switch cmd {
	case "/start":
		payload := strings.TrimSpace(strings.TrimPrefix(text, "/start"))
		return r.bot.handleStart(ctx, upd, payload)
	case "/menu":
		return r.bot.showMainMenu(ctx, upd)
	case "/tariffs", "/plans", "/buy", "/renew":
		return r.bot.showTariffs(ctx, upd)
	case "/status", "/sub":
		return r.bot.showStatus(ctx, upd)
	case "/help":
		return r.bot.showHelp(ctx, upd)
	case "/admin":
		return r.bot.showAdminMenu(ctx, upd)
	case "/grant":
		return r.bot.handleGrantCommand(ctx, upd)
	case "/check":
		return r.bot.handleCheckCommand(ctx, upd)
	case "/stars":
		return r.bot.handleStarsCommand(ctx, upd)
	case "/broadcast":
		return r.bot.handleBroadcastCommand(ctx, upd)
	case "/extend":
		return r.bot.handleExtendCommand(ctx, upd)
	case "/traffic":
		return r.bot.handleTrafficCommand(ctx, upd)
	case "/enable":
		return r.bot.handleEnableCommand(ctx, upd)
	case "/disable":
		return r.bot.handleDisableCommand(ctx, upd)
	case "/cancel":
		return r.bot.handleCancelCommand(ctx, upd)
	case "/block":
		return r.bot.handleBlockCommand(ctx, upd)
	case "/unblock":
		return r.bot.handleUnblockCommand(ctx, upd)
	case "/alias":
		return r.bot.handleAliasCommand(ctx, upd)
	case "/msg":
		return r.bot.handleMessageCommand(ctx, upd)
	case "/users":
		return r.bot.handleUsersCommand(ctx, upd)
	case "/subs":
		return r.bot.handleSubsCommand(ctx, upd)
	case "/admin_tariffs":
		return r.bot.handleAdminTariffsCommand(ctx, upd)
	default:
		switch text {
		case "Купить VPN / Продлить", "Купить / Продлить":
			return r.bot.showTariffs(ctx, upd)
		case "Личный кабинет":
			return r.bot.showStatus(ctx, upd)
		case "Как подключить":
			return r.bot.showInstructions(ctx, upd)
		case "Поддержка":
			return r.bot.showSupport(ctx, upd)
		default:
			return r.bot.showMainMenu(ctx, upd)
		}
	}
}

func (r *Router) handleCallback(ctx context.Context, upd update) error {
	cb := upd.CallbackQuery
	if cb == nil {
		return nil
	}

	if _, err := r.ensureUser(ctx, cb.From); err != nil {
		_ = r.bot.answerCallback(ctx, cb.ID, "Не удалось сохранить пользователя", true)
		return err
	}

	data := strings.TrimSpace(cb.Data)

	if err := r.bot.answerCallback(ctx, cb.ID, "", false); err != nil {
		r.bot.log.Debug("telegram answerCallback failed", zap.Error(err))
	}

	switch {
	case data == "menu:home":
		return r.bot.showMainMenu(ctx, upd)
	case data == "main:connect" || data == "profile:pay_options":
		return r.bot.showTariffs(ctx, upd)
	case data == "main:profile" || data == "profile:refresh":
		return r.bot.showStatus(ctx, upd)
	case data == "main:instructions":
		return r.bot.showInstructions(ctx, upd)
	case data == "main:support":
		return r.bot.showSupport(ctx, upd)
	case strings.HasPrefix(data, "tariff:"):
		return r.bot.showTariffPaymentOptions(ctx, upd, strings.TrimPrefix(data, "tariff:"))
	case strings.HasPrefix(data, "pay:stars:"):
		return r.bot.handleStarsPayment(ctx, upd, strings.TrimPrefix(data, "pay:stars:"))
	case strings.HasPrefix(data, "pay:crypto_check:"):
		return r.bot.handleCryptoBotPaymentCheck(ctx, upd, strings.TrimPrefix(data, "pay:crypto_check:"))
	case strings.HasPrefix(data, "pay:crypto:"):
		return r.bot.handleCryptoBotPayment(ctx, upd, strings.TrimPrefix(data, "pay:crypto:"))
	case strings.HasPrefix(data, "free:activate:"):
		return r.bot.activateWithoutPayment(ctx, upd, strings.TrimPrefix(data, "free:activate:"))
	case strings.HasPrefix(data, "dev:activate:"):
		return r.bot.activateWithoutPayment(ctx, upd, strings.TrimPrefix(data, "dev:activate:"))
	case strings.HasPrefix(data, "admin:"):
		return r.bot.handleAdminCallback(ctx, upd, strings.TrimPrefix(data, "admin:"))
	case strings.HasPrefix(data, "admin_user:"):
		return r.bot.handleAdminUserCallback(ctx, upd, strings.TrimPrefix(data, "admin_user:"))
	default:
		return r.bot.showMainMenu(ctx, upd)
	}
}

func (r *Router) ensureUser(ctx context.Context, user tgUser) (*domain.User, error) {
	saved, err := r.bot.services.Users.GetOrCreateTelegramUser(ctx, domain.TelegramUserInput{
		TelegramID:        user.ID,
		TelegramUsername:  optionalString(user.Username),
		TelegramFirstName: optionalString(user.FirstName),
		TelegramLastName:  optionalString(user.LastName),
		LanguageCode:      optionalString(user.LanguageCode),
	})
	if err != nil || saved == nil {
		return saved, err
	}

	if saved.Alias == nil || strings.TrimSpace(*saved.Alias) == "" {
		alias := defaultTelegramAlias(user)
		if alias != "" {
			updated, updateErr := r.bot.services.Users.Update(ctx, saved.ID, domain.UpdateUserInput{Alias: &alias})
			if updateErr == nil && updated != nil {
				saved = updated
			} else if updateErr != nil {
				r.bot.log.Debug("telegram user alias update failed", zap.Int64("telegram_id", user.ID), zap.Error(updateErr))
			}
		}
	}

	return saved, nil
}

func defaultTelegramAlias(user tgUser) string {
	username := strings.TrimSpace(user.Username)
	if username != "" {
		return "@" + strings.TrimPrefix(username, "@")
	}

	fullName := strings.TrimSpace(strings.Join([]string{
		strings.TrimSpace(user.FirstName),
		strings.TrimSpace(user.LastName),
	}, " "))
	if fullName != "" {
		return fullName
	}

	if user.ID > 0 {
		return "tg_" + strconv.FormatInt(user.ID, 10)
	}

	return ""
}

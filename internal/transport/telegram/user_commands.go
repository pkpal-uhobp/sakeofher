package telegramtransport

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"sakeofher/internal/domain"
)

func (b *Bot) handleStart(ctx context.Context, upd update, payload string) error {
	payload = strings.TrimSpace(payload)
	if strings.HasPrefix(payload, "buy_t") {
		if tariffID, ok := parseBuyPayload(payload); ok {
			return b.showTariffPaymentOptions(ctx, upd, strconv.FormatInt(tariffID, 10))
		}
	}
	return b.showMainMenu(ctx, upd)
}

func (b *Bot) showMainMenu(ctx context.Context, upd update) error {
	return b.replyOrEdit(ctx, upd, welcomeMessage(updateFirstName(upd)), b.mainMenuKeyboard(updateTelegramID(upd)))
}

func (b *Bot) showHelp(ctx context.Context, upd update) error {
	return b.replyOrEdit(ctx, upd, helpMessage(b.settings.isAdmin(updateTelegramID(upd))), backToMenuKeyboard())
}

func (b *Bot) showTariffs(ctx context.Context, upd update) error {
	items, err := b.services.Tariffs.ListActiveWithPrices(ctx)
	if err != nil {
		return b.replyOrEdit(ctx, upd, "Не удалось загрузить тарифы. Попробуйте позже.", backToMenuKeyboard())
	}
	return b.replyOrEdit(ctx, upd, tariffsMessage(items), tariffsKeyboard(items))
}

func (b *Bot) showStatus(ctx context.Context, upd update) error {
	telegramID := updateTelegramID(upd)
	if telegramID <= 0 {
		return nil
	}
	sub, err := b.services.Subscriptions.GetLatestByTelegramID(ctx, telegramID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return b.replyOrEdit(ctx, upd, "У вас пока нет подписки. Выберите тариф в разделе «Купить VPN / Продлить».", b.mainMenuKeyboard(telegramID))
		}
		return b.replyOrEdit(ctx, upd, "Не удалось получить статус подписки. Попробуйте позже.", backToMenuKeyboard())
	}
	active := sub.Subscription.Status == domain.SubscriptionStatusActive
	if active {
		// Refresh button should also reconcile Remnawave with the current DB state:
		// final expiration date + preferred/default squads. This fixes already-created
		// local test subscriptions without extending them again.
		if fixed, err := b.services.Subscriptions.Enable(ctx, sub.Subscription.ID); err == nil && fixed != nil {
			sub = fixed
		}
	}
	siteURL := b.siteSubscriptionURL(ctx, sub)
	return b.replyOrEdit(ctx, upd, statusMessage(sub, siteURL), profileKeyboard(active, siteURL))
}

func (b *Bot) showInstructions(ctx context.Context, upd update) error {
	return b.replyOrEdit(ctx, upd, instructionsMessage(), instructionsKeyboard(b.settings))
}

func (b *Bot) showSupport(ctx context.Context, upd update) error {
	return b.replyOrEdit(ctx, upd, supportMessage(b.settings), supportKeyboard(b.settings))
}

func updateTelegramID(upd update) int64 {
	if upd.Message != nil && upd.Message.From != nil {
		return upd.Message.From.ID
	}
	if upd.CallbackQuery != nil {
		return upd.CallbackQuery.From.ID
	}
	if upd.PreCheckoutQuery != nil {
		return upd.PreCheckoutQuery.From.ID
	}
	return 0
}

func updateFirstName(upd update) string {
	if upd.Message != nil && upd.Message.From != nil {
		return upd.Message.From.FirstName
	}
	if upd.CallbackQuery != nil {
		return upd.CallbackQuery.From.FirstName
	}
	return ""
}

func updateChatID(upd update) int64 {
	if upd.Message != nil {
		return upd.Message.Chat.ID
	}
	if upd.CallbackQuery != nil && upd.CallbackQuery.Message != nil {
		return upd.CallbackQuery.Message.Chat.ID
	}
	return 0
}

func parseBuyPayload(payload string) (int64, bool) {
	payload = strings.TrimPrefix(payload, "buy_t")
	parts := strings.Split(payload, "_")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	return id, err == nil && id > 0
}

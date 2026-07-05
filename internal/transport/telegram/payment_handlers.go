package telegramtransport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"sakeofher/internal/domain"
)

func (b *Bot) showTariffPaymentOptions(ctx context.Context, upd update, rawTariffID string) error {
	item, err := b.findActiveTariffWithPrices(ctx, rawTariffID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return b.replyOrEdit(ctx, upd, "Тариф не найден или отключён.", backToMenuKeyboard())
		}
		return b.replyOrEdit(ctx, upd, "Не удалось загрузить тариф. Попробуйте позже.", backToMenuKeyboard())
	}

	return b.replyOrEdit(ctx, upd, tariffPaymentMessage(*item), paymentMethodsKeyboard(*item, b.settings.AllowFreePurchase))
}

func (b *Bot) handleStarsPayment(ctx context.Context, upd update, rawPriceID string) error {
	chatID := updateChatID(upd)
	telegramID := updateTelegramID(upd)
	if chatID <= 0 || telegramID <= 0 {
		return nil
	}

	priceID, err := strconv.ParseInt(rawPriceID, 10, 64)
	if err != nil || priceID <= 0 {
		return b.replyOrEdit(ctx, upd, "Некорректный способ оплаты.", backToMenuKeyboard())
	}

	if _, err := b.services.Users.GetByTelegramID(ctx, telegramID); err != nil {
		return b.replyOrEdit(ctx, upd, "Не удалось найти пользователя. Нажмите /start и попробуйте снова.", backToMenuKeyboard())
	}

	payment, err := b.services.Payments.CreatePayment(ctx, domain.CreatePaymentInput{
		TelegramID:     telegramID,
		TariffPriceID: priceID,
	})
	if err != nil {
		return b.replyOrEdit(ctx, upd, "Не удалось создать платёж. Попробуйте позже.", backToMenuKeyboard())
	}

	if payment.StarsAmount == nil || *payment.StarsAmount <= 0 {
		return b.replyOrEdit(ctx, upd, "Для этого тарифа не настроена цена в Telegram Stars.", backToMenuKeyboard())
	}

	payload := fmt.Sprintf("stars:%d:%d:%d", payment.ID, payment.TariffID, telegramID)
	description := fmt.Sprintf(
		"SakeOfHer VPN на %s. Лимит: %s. После оплаты доступ будет выдан автоматически.",
		durationLabel(payment.DurationDays),
		trafficLabel(payment.TrafficLimitBytes),
	)

	if err := b.sendInvoice(ctx, chatID, "SakeOfHer VPN", description, payload, *payment.StarsAmount); err != nil {
		return b.replyOrEdit(ctx, upd, "Не удалось создать счёт Telegram Stars. Попробуйте позже.", backToMenuKeyboard())
	}

	return b.replyOrEdit(ctx, upd, invoiceCreatedMessage(), backToMenuKeyboard())
}

func (b *Bot) handleSuccessfulPayment(ctx context.Context, upd update) error {
	msg := upd.Message
	if msg == nil || msg.SuccessfulPayment == nil {
		return nil
	}

	paymentID, ok := parseStarsPayload(msg.SuccessfulPayment.InvoicePayload)
	if !ok {
		return b.sendMessage(ctx, msg.Chat.ID, "Оплата прошла, но payload платежа не распознан. Напишите в поддержку.", supportKeyboard(b.settings))
	}

	providerPaymentID := strings.TrimSpace(msg.SuccessfulPayment.TelegramPaymentChargeID)
	if providerPaymentID == "" {
		providerPaymentID = strings.TrimSpace(msg.SuccessfulPayment.ProviderPaymentChargeID)
	}
	if providerPaymentID == "" {
		providerPaymentID = fmt.Sprintf("telegram-stars-%d", paymentID)
	}

	rawPayload, _ := json.Marshal(msg.SuccessfulPayment)

	payment, err := b.services.Payments.ConfirmTelegramStarsPayment(ctx, domain.TelegramStarsPaidInput{
		PaymentID:  paymentID,
		ChargeID:   providerPaymentID,
		EventID:    providerPaymentID,
		PaidAt:     time.Now(),
		RawPayload: rawPayload,
	})
	if err != nil {
		b.log.Error("telegram stars payment activation failed", zap.Error(err), zap.Int64("payment_id", paymentID))
		return b.sendMessage(ctx, msg.Chat.ID, "Оплата прошла, но не удалось активировать подписку автоматически. Нажмите «Личный кабинет» или напишите в поддержку.", b.mainMenuKeyboard(msg.From.ID))
	}

	sub, err := b.services.Subscriptions.GetLatestByTelegramID(ctx, msg.From.ID)
	if err != nil {
		return b.sendMessage(ctx, msg.Chat.ID, fmt.Sprintf("Оплата #%d подтверждена, но статус пока не обновился. Нажмите «Обновить данные» через 5–10 секунд.", payment.ID), b.mainMenuKeyboard(msg.From.ID))
	}

	siteURL := b.siteSubscriptionURL(ctx, sub)
	return b.sendMessage(ctx, msg.Chat.ID, paymentActivatedMessage(sub, siteURL), profileKeyboard(true, siteURL))
}

func (b *Bot) handleExternalPaymentNotReady(ctx context.Context, upd update, provider string) error {
	return b.replyOrEdit(ctx, upd, cryptoNotReadyMessage(provider), supportKeyboard(b.settings))
}

func (b *Bot) handleCryptoBotPayment(ctx context.Context, upd update, rawPriceID string) error {
	telegramID := updateTelegramID(upd)
	if telegramID <= 0 {
		return nil
	}

	priceID, err := strconv.ParseInt(rawPriceID, 10, 64)
	if err != nil || priceID <= 0 {
		return b.replyOrEdit(ctx, upd, "Некорректный способ оплаты.", backToMenuKeyboard())
	}

	payment, err := b.services.Payments.CreateCryptoBotPayment(ctx, domain.CreateCryptoBotPaymentInput{
		TelegramID:     telegramID,
		TariffPriceID: priceID,
	})
	if err != nil {
		b.log.Error(
			"cryptobot create payment failed",
			zap.Error(err),
			zap.Int64("telegram_id", telegramID),
			zap.Int64("tariff_price_id", priceID),
		)

		return b.replyOrEdit(ctx, upd, "Не удалось создать счёт CryptoBot.\n\nОшибка:\n"+err.Error(), backToMenuKeyboard())
	}

	if payment.PaymentURL == nil || strings.TrimSpace(*payment.PaymentURL) == "" {
		b.log.Error(
			"cryptobot invoice created without payment url",
			zap.Int64("telegram_id", telegramID),
			zap.Int64("tariff_price_id", priceID),
			zap.Int64("payment_id", payment.ID),
		)

		return b.replyOrEdit(ctx, upd, "CryptoBot создал платёж без ссылки. Напишите в поддержку.", supportKeyboard(b.settings))
	}

	return b.replyOrEdit(
		ctx,
		upd,
		"Счёт CryptoBot создан.\n\nПосле оплаты нажмите «Проверить оплату». Также оплату автоматически проверяет воркер.",
		&inlineKeyboardMarkup{InlineKeyboard: [][]inlineKeyboardButton{
			{
				{Text: "Открыть оплату CryptoBot", URL: *payment.PaymentURL},
			},
			{
				{Text: "Проверить оплату", CallbackData: fmt.Sprintf("pay:crypto_check:%d", payment.ID)},
			},
			{
				{Text: "Личный кабинет", CallbackData: "profile:refresh"},
			},
			{
				{Text: "В меню", CallbackData: "menu:home"},
			},
		}},
	)
}

func (b *Bot) handleCryptoBotPaymentCheck(ctx context.Context, upd update, rawPaymentID string) error {
	telegramID := updateTelegramID(upd)
	if telegramID <= 0 {
		return nil
	}

	paymentID, err := strconv.ParseInt(rawPaymentID, 10, 64)
	if err != nil || paymentID <= 0 {
		return b.replyOrEdit(ctx, upd, "Некорректный платёж.", backToMenuKeyboard())
	}

	payment, err := b.services.Payments.CheckCryptoBotPayment(ctx, paymentID)
	if err != nil {
		b.log.Error(
			"cryptobot manual check failed",
			zap.Error(err),
			zap.Int64("telegram_id", telegramID),
			zap.Int64("payment_id", paymentID),
		)

		return b.replyOrEdit(ctx, upd, "Не удалось проверить оплату.\n\nОшибка:\n"+err.Error(), backToMenuKeyboard())
	}

	if payment.Status != domain.PaymentStatusActivated {
		return b.replyOrEdit(
			ctx,
			upd,
			"Оплата пока не найдена.\n\nЕсли вы уже оплатили счёт, подождите 10–30 секунд и нажмите «Проверить оплату» ещё раз.",
			&inlineKeyboardMarkup{InlineKeyboard: [][]inlineKeyboardButton{
				{
					{Text: "Проверить оплату", CallbackData: fmt.Sprintf("pay:crypto_check:%d", payment.ID)},
				},
				{
					{Text: "Личный кабинет", CallbackData: "profile:refresh"},
				},
				{
					{Text: "В меню", CallbackData: "menu:home"},
				},
			}},
		)
	}

	sub, err := b.services.Subscriptions.GetLatestByTelegramID(ctx, telegramID)
	if err != nil {
		return b.replyOrEdit(ctx, upd, "Оплата подтверждена, но подписка ещё обновляется. Нажмите «Личный кабинет» через несколько секунд.", b.mainMenuKeyboard(telegramID))
	}

	siteURL := b.siteSubscriptionURL(ctx, sub)
	return b.replyOrEdit(ctx, upd, paymentActivatedMessage(sub, siteURL), profileKeyboard(true, siteURL))
}

func (b *Bot) activateWithoutPayment(ctx context.Context, upd update, rawTariffID string) error {
	if !b.settings.AllowFreePurchase {
		return b.replyOrEdit(ctx, upd, "Покупка без оплаты выключена.", backToMenuKeyboard())
	}

	telegramID := updateTelegramID(upd)
	if telegramID <= 0 {
		return nil
	}

	tariffID, err := strconv.ParseInt(rawTariffID, 10, 64)
	if err != nil || tariffID <= 0 {
		return b.replyOrEdit(ctx, upd, "Некорректный тариф.", backToMenuKeyboard())
	}

	user, err := b.services.Users.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return b.replyOrEdit(ctx, upd, "Не удалось найти пользователя. Нажмите /start и попробуйте снова.", backToMenuKeyboard())
	}

	tariff, err := b.services.Tariffs.GetByID(ctx, tariffID)
	if err != nil {
		return b.replyOrEdit(ctx, upd, "Тариф не найден или отключён.", backToMenuKeyboard())
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
		return b.replyOrEdit(ctx, upd, "Не удалось активировать подписку. Проверь Remnawave env, сквады и логи API/worker.", backToMenuKeyboard())
	}

	if fixed, err := b.services.Subscriptions.Enable(ctx, sub.Subscription.ID); err == nil && fixed != nil {
		sub = fixed
	}

	siteURL := b.siteSubscriptionURL(ctx, sub)
	return b.replyOrEdit(ctx, upd, freeActivatedMessage(sub, siteURL), profileKeyboard(true, siteURL))
}

func (b *Bot) findActiveTariffWithPrices(ctx context.Context, rawTariffID string) (*domain.TariffWithPrices, error) {
	tariffID, err := strconv.ParseInt(rawTariffID, 10, 64)
	if err != nil || tariffID <= 0 {
		return nil, domain.ErrInvalidInput
	}

	items, err := b.services.Tariffs.ListActiveWithPrices(ctx)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.ID == tariffID {
			return &item, nil
		}
	}

	return nil, domain.ErrNotFound
}

func parseStarsPayload(payload string) (int64, bool) {
	parts := strings.Split(strings.TrimSpace(payload), ":")
	if len(parts) < 2 || parts[0] != "stars" {
		return 0, false
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	return id, err == nil && id > 0
}

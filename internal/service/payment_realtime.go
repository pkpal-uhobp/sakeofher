package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"sakeofher/internal/domain"
)

func (s *paymentService) ConfirmTelegramStarsPayment(
	ctx context.Context,
	input domain.TelegramStarsPaidInput,
) (*domain.Payment, error) {
	if input.PaymentID <= 0 || strings.TrimSpace(input.ChargeID) == "" {
		return nil, domain.ErrInvalidInput
	}
	if input.EventID == "" {
		input.EventID = input.ChargeID
	}
	if input.PaidAt.IsZero() {
		input.PaidAt = time.Now()
	}

	var paymentID int64

	err := s.repo.Tx.WithinTransaction(ctx, func(ctx context.Context) error {
		created, err := s.repo.PaymentEvents.CreateOnce(ctx, domain.PaymentEvent{
			Provider:   domain.PaymentProviderTelegramStars,
			EventID:    input.EventID,
			EventType:  "successful_payment",
			RawPayload: input.RawPayload,
		})
		if err != nil {
			return err
		}
		if !created {
			return domain.ErrPaymentEventAlreadyProcessed
		}

		p, err := s.repo.Payments.GetByIDForUpdate(ctx, input.PaymentID)
		if err != nil {
			return err
		}
		if p.Provider != domain.PaymentProviderTelegramStars {
			return domain.ErrInvalidInput
		}
		if p.Status == domain.PaymentStatusActivated {
			paymentID = p.ID
			return domain.ErrPaymentAlreadyActivated
		}

		if err := s.repo.Payments.MarkPaid(ctx, p.ID, input.ChargeID, input.PaidAt, input.RawPayload); err != nil {
			return err
		}

		paymentID = p.ID
		return nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrPaymentEventAlreadyProcessed) ||
			errors.Is(err, domain.ErrPaymentAlreadyActivated) {
			return s.repo.Payments.GetByID(ctx, input.PaymentID)
		}
		return nil, err
	}

	if err := s.subscriptions.ActivateAfterPayment(ctx, paymentID); err != nil {
		return nil, err
	}

	return s.repo.Payments.GetByID(ctx, paymentID)
}

func (s *paymentService) CreateCryptoBotPayment(
	ctx context.Context,
	input domain.CreateCryptoBotPaymentInput,
) (*domain.Payment, error) {
	if input.TelegramID <= 0 || input.TariffPriceID <= 0 {
		return nil, domain.ErrInvalidInput
	}

	priceWithTariff, err := s.repo.TariffPrices.GetWithTariffByID(ctx, input.TariffPriceID)
	if err != nil {
		return nil, err
	}

	if !priceWithTariff.Price.IsActive || !priceWithTariff.Tariff.IsActive {
		return nil, domain.ErrInactiveTariffPrice
	}
	if priceWithTariff.Price.Provider != domain.PaymentProviderCryptoBot ||
		priceWithTariff.Price.PaymentMethod != domain.PaymentMethodCrypto {
		return nil, domain.ErrInvalidInput
	}
	if priceWithTariff.Price.AmountMinor == nil || *priceWithTariff.Price.AmountMinor <= 0 {
		return nil, domain.ErrInvalidInput
	}

	payment, err := s.CreatePayment(ctx, domain.CreatePaymentInput{
		TelegramID:     input.TelegramID,
		TariffPriceID: input.TariffPriceID,
	})
	if err != nil {
		return nil, err
	}

	amountRub := fmt.Sprintf("%.2f", float64(*payment.AmountMinor)/100.0)
	description := fmt.Sprintf(
		"SakeOfHer VPN: %s, %d дней",
		priceWithTariff.Tariff.Title,
		priceWithTariff.Tariff.DurationDays,
	)
	payload := fmt.Sprintf("cryptobot:%d:%d:%d", payment.ID, payment.TariffID, input.TelegramID)

	invoice, err := s.gates.CryptoBot.CreateInvoice(ctx, domain.CryptoBotCreateInvoiceRequest{
		Amount:         amountRub,
		CurrencyType:   "fiat",
		Fiat:           "RUB",
		AcceptedAssets: priceWithTariff.Price.AcceptedAssets,
		Description:    description,
		Payload:        payload,
		ExpiresIn:      30 * 60,
	})
	if err != nil {
		return nil, err
	}

	providerPaymentID := strconv.FormatInt(invoice.InvoiceID, 10)
	paymentURL := bestCryptoBotInvoiceURL(invoice)

	if err := s.repo.Payments.MarkWaitingPayment(
		ctx,
		payment.ID,
		&providerPaymentID,
		&paymentURL,
		invoice.ExpirationDate,
	); err != nil {
		return nil, err
	}

	return s.repo.Payments.GetByID(ctx, payment.ID)
}

func (s *paymentService) CheckCryptoBotPayment(ctx context.Context, paymentID int64) (*domain.Payment, error) {
	if paymentID <= 0 {
		return nil, domain.ErrInvalidInput
	}

	payment, err := s.repo.Payments.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if payment.Provider != domain.PaymentProviderCryptoBot {
		return nil, domain.ErrInvalidInput
	}
	if payment.Status == domain.PaymentStatusActivated {
		return payment, nil
	}
	if payment.ProviderPaymentID == nil || strings.TrimSpace(*payment.ProviderPaymentID) == "" {
		return payment, nil
	}

	invoice, err := s.gates.CryptoBot.GetInvoice(ctx, *payment.ProviderPaymentID)
	if err != nil {
		return nil, err
	}
	if invoice.Status != "paid" {
		return payment, nil
	}

	rawPayload, _ := json.Marshal(invoice)
	paidAt := time.Now()
	if invoice.PaidAt != nil {
		paidAt = *invoice.PaidAt
	}

	err = s.HandlePaymentPaid(ctx, domain.PaymentPaidInput{
		Provider:          domain.PaymentProviderCryptoBot,
		EventID:           "cryptobot_invoice_paid_" + strconv.FormatInt(invoice.InvoiceID, 10),
		EventType:         "invoice_paid",
		ProviderPaymentID: strconv.FormatInt(invoice.InvoiceID, 10),
		PaidAt:            paidAt,
		RawPayload:        rawPayload,
	})
	if err != nil {
		if !errors.Is(err, domain.ErrPaymentEventAlreadyProcessed) &&
			!errors.Is(err, domain.ErrPaymentAlreadyActivated) {
			return nil, err
		}
	}

	return s.repo.Payments.GetByID(ctx, paymentID)
}

func (s *paymentService) PollCryptoBotPayments(ctx context.Context, limit int) error {
	payments, err := s.repo.Payments.FindWaitingByProvider(ctx, domain.PaymentProviderCryptoBot, limit)
	if err != nil {
		return err
	}

	for _, payment := range payments {
		if payment.ProviderPaymentID == nil || strings.TrimSpace(*payment.ProviderPaymentID) == "" {
			continue
		}

		invoice, err := s.gates.CryptoBot.GetInvoice(ctx, *payment.ProviderPaymentID)
		if err != nil {
			return err
		}
		if invoice.Status != "paid" {
			continue
		}

		rawPayload, _ := json.Marshal(invoice)
		paidAt := time.Now()
		if invoice.PaidAt != nil {
			paidAt = *invoice.PaidAt
		}

		err = s.HandlePaymentPaid(ctx, domain.PaymentPaidInput{
			Provider:          domain.PaymentProviderCryptoBot,
			EventID:           "cryptobot_invoice_paid_" + strconv.FormatInt(invoice.InvoiceID, 10),
			EventType:         "invoice_paid",
			ProviderPaymentID: strconv.FormatInt(invoice.InvoiceID, 10),
			PaidAt:            paidAt,
			RawPayload:        rawPayload,
		})
		if err != nil {
			if errors.Is(err, domain.ErrPaymentEventAlreadyProcessed) ||
				errors.Is(err, domain.ErrPaymentAlreadyActivated) {
				continue
			}
			return err
		}
	}

	return nil
}

func (s *paymentService) HandleCryptoBotWebhook(ctx context.Context, rawPayload json.RawMessage) error {
	var update struct {
		UpdateID   int64           `json:"update_id"`
		UpdateType string          `json:"update_type"`
		Payload    json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(rawPayload, &update); err != nil {
		return err
	}

	if update.UpdateType != "" && update.UpdateType != "invoice_paid" {
		return nil
	}

	var invoice struct {
		InvoiceID int64  `json:"invoice_id"`
		Status    string `json:"status"`
	}
	if err := json.Unmarshal(update.Payload, &invoice); err != nil {
		return err
	}
	if invoice.InvoiceID <= 0 {
		return domain.ErrInvalidInput
	}
	if invoice.Status != "" && strings.ToLower(invoice.Status) != "paid" {
		return nil
	}

	eventID := fmt.Sprintf("cryptobot_update_%d", update.UpdateID)
	if update.UpdateID <= 0 {
		eventID = fmt.Sprintf("cryptobot_invoice_paid_%d", invoice.InvoiceID)
	}

	return s.HandlePaymentPaid(ctx, domain.PaymentPaidInput{
		Provider:          domain.PaymentProviderCryptoBot,
		EventID:           eventID,
		EventType:         "invoice_paid",
		ProviderPaymentID: strconv.FormatInt(invoice.InvoiceID, 10),
		PaidAt:            time.Now(),
		RawPayload:        rawPayload,
	})
}

func bestCryptoBotInvoiceURL(invoice *domain.CryptoBotInvoice) string {
	if invoice == nil {
		return ""
	}
	if strings.TrimSpace(invoice.MiniAppInvoiceURL) != "" {
		return invoice.MiniAppInvoiceURL
	}
	if strings.TrimSpace(invoice.WebAppInvoiceURL) != "" {
		return invoice.WebAppInvoiceURL
	}
	return strings.TrimSpace(invoice.BotInvoiceURL)
}

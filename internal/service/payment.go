package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"sakeofher/internal/domain"
	"sakeofher/internal/gateway"
	"sakeofher/internal/repository"
)

type paymentService struct {
	repo          *repository.Repositories
	gates         gateway.Gateways
	subscriptions SubscriptionService
}

func NewPaymentService(repo *repository.Repositories, gates gateway.Gateways, subscriptions SubscriptionService) PaymentService {
	return &paymentService{repo: repo, gates: gates, subscriptions: subscriptions}
}

func (s *paymentService) CreatePayment(ctx context.Context, input domain.CreatePaymentInput) (*domain.Payment, error) {
	if input.TelegramID <= 0 || input.TariffPriceID <= 0 {
		return nil, domain.ErrInvalidInput
	}

	var result *domain.Payment
	err := s.repo.Tx.WithinTransaction(ctx, func(ctx context.Context) error {
		user, err := s.repo.Users.GetByTelegramID(ctx, input.TelegramID)
		if err != nil {
			return err
		}

		priceWithTariff, err := s.repo.TariffPrices.GetWithTariffByID(ctx, input.TariffPriceID)
		if err != nil {
			return err
		}
		if !priceWithTariff.Price.IsActive || !priceWithTariff.Tariff.IsActive {
			return domain.ErrInactiveTariffPrice
		}

		priceID := priceWithTariff.Price.ID
		payment := &domain.Payment{
			UserID:            user.ID,
			TariffID:          priceWithTariff.Tariff.ID,
			TariffPriceID:     &priceID,
			Provider:          priceWithTariff.Price.Provider,
			PaymentMethod:     priceWithTariff.Price.PaymentMethod,
			Currency:          priceWithTariff.Price.Currency,
			AmountMinor:       priceWithTariff.Price.AmountMinor,
			StarsAmount:       priceWithTariff.Price.StarsAmount,
			DurationDays:      priceWithTariff.Tariff.DurationDays,
			PeriodDays:        priceWithTariff.Tariff.PeriodDays,
			TrafficLimitBytes: priceWithTariff.Tariff.TrafficLimitBytes,
			Status:            domain.PaymentStatusCreated,
		}

		if err := s.repo.Payments.Create(ctx, payment); err != nil {
			return err
		}
		result = payment
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// MarkPaidForDev is a temporary local helper for Stage 2.
// Later real paid state will come from Telegram Stars / Tribute / CryptoBot webhooks.
func (s *paymentService) MarkPaidForDev(ctx context.Context, paymentID int64, providerPaymentID string) (*domain.Payment, error) {
	if paymentID <= 0 {
		return nil, domain.ErrInvalidInput
	}
	if providerPaymentID == "" {
		providerPaymentID = fmt.Sprintf("dev-%d-%d", paymentID, time.Now().Unix())
	}

	var payment *domain.Payment
	err := s.repo.Tx.WithinTransaction(ctx, func(ctx context.Context) error {
		p, err := s.repo.Payments.GetByIDForUpdate(ctx, paymentID)
		if err != nil {
			return err
		}
		if p.Status == domain.PaymentStatusActivated {
			payment = p
			return nil
		}
		raw, _ := json.Marshal(map[string]any{"source": "dev", "payment_id": paymentID})
		if err := s.repo.Payments.MarkPaid(ctx, p.ID, providerPaymentID, time.Now(), raw); err != nil {
			return err
		}
		payment = p
		return nil
	})
	if err != nil {
		return nil, err
	}

	if payment != nil && payment.Status != domain.PaymentStatusActivated {
		if err := s.subscriptions.ActivateAfterPayment(ctx, payment.ID); err != nil {
			return nil, err
		}
	}
	return s.repo.Payments.GetByID(ctx, paymentID)
}

func (s *paymentService) HandlePaymentPaid(ctx context.Context, input domain.PaymentPaidInput) error {
	if input.Provider == "" || input.EventID == "" || input.ProviderPaymentID == "" {
		return domain.ErrInvalidInput
	}
	if input.PaidAt.IsZero() {
		input.PaidAt = time.Now()
	}

	var paymentID int64
	err := s.repo.Tx.WithinTransaction(ctx, func(ctx context.Context) error {
		created, err := s.repo.PaymentEvents.CreateOnce(ctx, domain.PaymentEvent{
			Provider:   input.Provider,
			EventID:    input.EventID,
			EventType:  input.EventType,
			RawPayload: input.RawPayload,
		})
		if err != nil {
			return err
		}
		if !created {
			return domain.ErrPaymentEventAlreadyProcessed
		}

		p, err := s.repo.Payments.GetByProviderPaymentIDForUpdate(ctx, input.Provider, input.ProviderPaymentID)
		if err != nil {
			return err
		}
		if p.Status == domain.PaymentStatusActivated {
			return domain.ErrPaymentAlreadyActivated
		}
		if err := s.repo.Payments.MarkPaid(ctx, p.ID, input.ProviderPaymentID, input.PaidAt, input.RawPayload); err != nil {
			return err
		}
		paymentID = p.ID
		return nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrPaymentEventAlreadyProcessed) || errors.Is(err, domain.ErrPaymentAlreadyActivated) {
			return nil
		}
		return err
	}

	return s.subscriptions.ActivateAfterPayment(ctx, paymentID)
}

func (s *paymentService) RetryFailedActivations(ctx context.Context, limit int) error {
	payments, err := s.repo.Payments.FindActivationFailed(ctx, limit)
	if err != nil {
		return err
	}
	for _, p := range payments {
		if err := s.subscriptions.ActivateAfterPayment(ctx, p.ID); err != nil {
			return err
		}
	}
	return nil
}

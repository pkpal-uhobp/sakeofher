package service

import (
	"context"
	"time"

	"sakeofher/internal/domain"
	"sakeofher/internal/gateway"
	"sakeofher/internal/repository"
)

type PaymentService struct {
	repo          *repository.Repositories
	gates         gateway.Gateways
	subscriptions *SubscriptionService
}

func NewPaymentService(repo *repository.Repositories, gates gateway.Gateways, subscriptions *SubscriptionService) *PaymentService {
	return &PaymentService{repo: repo, gates: gates, subscriptions: subscriptions}
}

func (s *PaymentService) CreatePayment(ctx context.Context, userID int64, tariffPriceID int64) (*domain.Payment, error) {
	var result *domain.Payment
	err := s.repo.Tx.WithinTransaction(ctx, func(ctx context.Context) error {
		price, err := s.repo.TariffPrices.GetByID(ctx, tariffPriceID)
		if err != nil {
			return err
		}
		p := &domain.Payment{
			UserID:        userID,
			TariffID:      price.TariffID,
			TariffPriceID: price.ID,
			Provider:      price.Provider,
			PaymentMethod: price.PaymentMethod,
			Status:        domain.PaymentStatusCreated,
			Currency:      price.Currency,
			AmountMinor:   price.AmountMinor,
			StarsAmount:   price.StarsAmount,
		}
		if err := s.repo.Payments.Create(ctx, p); err != nil {
			return err
		}
		result = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PaymentService) MarkPaid(ctx context.Context, paymentID int64, providerPaymentID string, rawPayload []byte) error {
	return s.repo.Tx.WithinTransaction(ctx, func(ctx context.Context) error {
		return s.repo.Payments.MarkPaid(ctx, paymentID, providerPaymentID, time.Now(), rawPayload)
	})
}

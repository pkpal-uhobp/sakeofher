package service

import (
	"context"
	"time"

	"sakeofher/internal/domain"
)

func (s *subscriptionService) List(ctx context.Context, input domain.SubscriptionListInput) (*domain.SubscriptionListResponse, error) {
	if input.Status != "" && input.Status != domain.SubscriptionStatusActive && input.Status != domain.SubscriptionStatusExpired && input.Status != domain.SubscriptionStatusCancelled {
		return nil, domain.ErrInvalidInput
	}

	items, total, err := s.repo.Subscriptions.ListPublic(ctx, input)
	if err != nil {
		return nil, err
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	return &domain.SubscriptionListResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *subscriptionService) GetByID(ctx context.Context, id int64) (*domain.PublicSubscription, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Subscriptions.GetPublicByID(ctx, id)
}

func (s *subscriptionService) CreateManual(ctx context.Context, input domain.CreateManualSubscriptionInput) (*domain.PublicSubscription, error) {
	if input.UserID <= 0 || input.TariffID <= 0 || input.TrafficLimitGB <= 0 {
		return nil, domain.ErrInvalidInput
	}

	tariff, err := s.repo.Tariffs.GetByID(ctx, input.TariffID)
	if err != nil {
		return nil, err
	}

	trafficLimitBytes := domain.TrafficGBToBytes(input.TrafficLimitGB)
	now := time.Now()

	if err := s.repo.Tx.WithinTransaction(ctx, func(ctx context.Context) error {
		user, err := s.repo.Users.GetByIDForUpdate(ctx, input.UserID)
		if err != nil {
			return err
		}

		return s.createOrExtendSiteSubscription(ctx, user.ID, tariff, trafficLimitBytes, now)
	}); err != nil {
		return nil, err
	}

	active, err := s.repo.Subscriptions.GetActiveByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return s.repo.Subscriptions.GetPublicByID(ctx, active.ID)
}

func (s *subscriptionService) Extend(ctx context.Context, id int64, input domain.ExtendSubscriptionInput) (*domain.PublicSubscription, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	var out *domain.PublicSubscription

	if err := s.repo.Tx.WithinTransaction(ctx, func(ctx context.Context) error {
		sub, err := s.repo.Subscriptions.GetByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}

		tariffID := int64(0)
		if sub.TariffID != nil {
			tariffID = *sub.TariffID
		}
		if input.TariffID != nil && *input.TariffID > 0 {
			tariffID = *input.TariffID
		}
		if tariffID <= 0 {
			return domain.ErrInvalidInput
		}

		tariff, err := s.repo.Tariffs.GetByID(ctx, tariffID)
		if err != nil {
			return err
		}

		days := tariff.DurationDays
		if input.Days != nil && *input.Days > 0 {
			days = *input.Days
		}

		now := time.Now()
		base := sub.ExpiresAt
		if base.Before(now) {
			base = now
		}

		nextExpiresAt := base.AddDate(0, 0, days)
		periodStart := sub.CurrentPeriodStart
		periodEnd := sub.CurrentPeriodEnd
		if periodEnd.Before(now) || periodEnd.After(nextExpiresAt) {
			periodStart = now
			periodEnd = now.AddDate(0, 0, tariff.PeriodDays)
			if periodEnd.After(nextExpiresAt) {
				periodEnd = nextExpiresAt
			}
		}

		if err := s.repo.Subscriptions.ExtendByID(ctx, id, tariff.ID, nextExpiresAt, periodStart, periodEnd, sub.TrafficLimitBytes); err != nil {
			return err
		}

		out, err = s.repo.Subscriptions.GetPublicByID(ctx, id)
		return err
	}); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *subscriptionService) Update(ctx context.Context, id int64, input domain.UpdateSubscriptionInput) (*domain.PublicSubscription, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	if input.Status != nil {
		switch *input.Status {
		case domain.SubscriptionStatusActive, domain.SubscriptionStatusExpired, domain.SubscriptionStatusCancelled:
		default:
			return nil, domain.ErrInvalidInput
		}
	}

	if input.PeriodStatus != nil {
		switch *input.PeriodStatus {
		case domain.PeriodStatusActive, domain.PeriodStatusTrafficExhausted, domain.PeriodStatusFinished:
		default:
			return nil, domain.ErrInvalidInput
		}
	}

	if err := s.repo.Subscriptions.UpdateManual(ctx, id, input); err != nil {
		return nil, err
	}

	return s.repo.Subscriptions.GetPublicByID(ctx, id)
}

func (s *subscriptionService) UpdateTrafficLimit(ctx context.Context, id int64, input domain.UpdateTrafficLimitInput) (*domain.PublicSubscription, error) {
	if id <= 0 || input.TrafficLimitGB <= 0 {
		return nil, domain.ErrInvalidInput
	}

	if err := s.repo.Subscriptions.UpdateTrafficLimit(ctx, id, domain.TrafficGBToBytes(input.TrafficLimitGB)); err != nil {
		return nil, err
	}

	return s.repo.Subscriptions.GetPublicByID(ctx, id)
}

func (s *subscriptionService) Disable(ctx context.Context, id int64) (*domain.PublicSubscription, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	if err := s.repo.Subscriptions.SetStatus(ctx, id, domain.SubscriptionStatusExpired, domain.PeriodStatusFinished); err != nil {
		return nil, err
	}

	return s.repo.Subscriptions.GetPublicByID(ctx, id)
}

func (s *subscriptionService) Enable(ctx context.Context, id int64) (*domain.PublicSubscription, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	if err := s.repo.Subscriptions.SetStatus(ctx, id, domain.SubscriptionStatusActive, domain.PeriodStatusActive); err != nil {
		return nil, err
	}

	return s.repo.Subscriptions.GetPublicByID(ctx, id)
}

func (s *subscriptionService) Cancel(ctx context.Context, id int64) (*domain.PublicSubscription, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	if err := s.repo.Subscriptions.SetStatus(ctx, id, domain.SubscriptionStatusCancelled, domain.PeriodStatusFinished); err != nil {
		return nil, err
	}

	return s.repo.Subscriptions.GetPublicByID(ctx, id)
}

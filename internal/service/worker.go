package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository"
)

type workerService struct {
	repo          *repository.Repositories
	subscriptions SubscriptionService
	payments      PaymentService
	notifications NotificationService
	log           *zap.Logger
}

func NewWorkerService(
	repo *repository.Repositories,
	subscriptions SubscriptionService,
	payments PaymentService,
	notifications NotificationService,
	log *zap.Logger,
) WorkerService {
	return &workerService{
		repo:           repo,
		subscriptions: subscriptions,
		payments:       payments,
		notifications:  notifications,
		log:            log,
	}
}

func (s *workerService) ExpireSubscriptions(ctx context.Context) error {
	return s.subscriptions.DisableExpiredSubscriptions(ctx, 100)
}

func (s *workerService) DeleteOldDisabledUsers(ctx context.Context) error {
	return s.subscriptions.DeleteOldDisabledUsers(ctx, 100)
}

func (s *workerService) RetryFailedActivations(ctx context.Context) error {
	return s.payments.RetryFailedActivations(ctx, 50)
}

func (s *workerService) SyncUsage(ctx context.Context) error {
	return s.subscriptions.SyncRemnaUsage(ctx, 100)
}

func (s *workerService) ResetTrafficPeriods(ctx context.Context) error {
	return s.subscriptions.ResetTrafficPeriods(ctx, 100)
}

func (s *workerService) NotifyExpiringAndTraffic(ctx context.Context) error {
	now := time.Now()

	if err := s.notifyExpiring(ctx, now); err != nil {
		return err
	}

	if err := s.notifyLowTraffic(ctx); err != nil {
		return err
	}

	return nil
}

func (s *workerService) notifyExpiring(ctx context.Context, now time.Time) error {
	items, err := s.repo.Subscriptions.FindExpiringForNotifications(ctx, now, now.Add(72*time.Hour), 500)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		s.log.Info("worker notify expiring: no subscriptions found")
		return nil
	}

	sent := 0

	for _, item := range items {
		remaining := time.Until(item.Subscription.ExpiresAt)
		label, key := expirationLabelAndKey(remaining)
		if key == "" {
			continue
		}

		alreadySent, err := s.repo.Subscriptions.WasNotificationSent(ctx, item.Subscription.ID, key)
		if err != nil {
			return err
		}
		if alreadySent {
			continue
		}

		text := fmt.Sprintf("Подписка скоро закончится: осталось %s. Продлите доступ в боте.", label)

		s.log.Warn(
			"subscription expiration warning",
			zap.Int64("subscription_id", item.Subscription.ID),
			zap.Int64("telegram_id", item.User.TelegramID),
			zap.String("username", optionalString(item.User.TelegramUsername)),
			zap.String("notification_key", key),
			zap.String("remaining", label),
			zap.Time("expires_at", item.Subscription.ExpiresAt),
		)

		if err := s.notifications.Send(ctx, item.User.TelegramID, text); err != nil {
			s.log.Error("send expiration warning failed", zap.Int64("telegram_id", item.User.TelegramID), zap.Error(err))
		}

		if err := s.repo.Subscriptions.MarkNotificationSent(ctx, item.Subscription.ID, key); err != nil {
			return err
		}

		sent++
	}

	s.log.Info("worker notify expiring finished", zap.Int("found", len(items)), zap.Int("sent", sent))
	return nil
}

func (s *workerService) notifyLowTraffic(ctx context.Context) error {
	items, err := s.repo.Subscriptions.FindLowTrafficForNotifications(ctx, 15*domain.BytesInGiB, 500)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		s.log.Info("worker notify traffic: no subscriptions found")
		return nil
	}

	sent := 0

	for _, item := range items {
		remaining := item.Subscription.TrafficLimitBytes - item.Subscription.TrafficUsedBytes
		if remaining < 0 {
			remaining = 0
		}

		label, key := trafficLabelAndKey(remaining)
		if key == "" {
			continue
		}

		alreadySent, err := s.repo.Subscriptions.WasNotificationSent(ctx, item.Subscription.ID, key)
		if err != nil {
			return err
		}
		if alreadySent {
			continue
		}

		text := fmt.Sprintf("Заканчивается трафик: осталось примерно %s. После исчерпания доступ будет остановлен до нового периода или продления.", label)

		s.log.Warn(
			"subscription traffic warning",
			zap.Int64("subscription_id", item.Subscription.ID),
			zap.Int64("telegram_id", item.User.TelegramID),
			zap.String("username", optionalString(item.User.TelegramUsername)),
			zap.String("notification_key", key),
			zap.Int64("remaining_bytes", remaining),
			zap.String("remaining", label),
			zap.Int64("traffic_limit_bytes", item.Subscription.TrafficLimitBytes),
			zap.Int64("traffic_used_bytes", item.Subscription.TrafficUsedBytes),
		)

		if err := s.notifications.Send(ctx, item.User.TelegramID, text); err != nil {
			s.log.Error("send traffic warning failed", zap.Int64("telegram_id", item.User.TelegramID), zap.Error(err))
		}

		if err := s.repo.Subscriptions.MarkNotificationSent(ctx, item.Subscription.ID, key); err != nil {
			return err
		}

		sent++
	}

	s.log.Info("worker notify traffic finished", zap.Int("found", len(items)), zap.Int("sent", sent))
	return nil
}

func expirationLabelAndKey(remaining time.Duration) (string, string) {
	if remaining <= 0 {
		return "", ""
	}

	switch {
	case remaining <= 6*time.Hour:
		return "6 часов", "expire_6h"
	case remaining <= 24*time.Hour:
		return "1 день", "expire_1d"
	case remaining <= 48*time.Hour:
		return "2 дня", "expire_2d"
	case remaining <= 72*time.Hour:
		return "3 дня", "expire_3d"
	default:
		return "", ""
	}
}

func trafficLabelAndKey(remainingBytes int64) (string, string) {
	switch {
	case remainingBytes <= 5*domain.BytesInGiB:
		return "5 ГБ", "traffic_5gb"
	case remainingBytes <= 10*domain.BytesInGiB:
		return "10 ГБ", "traffic_10gb"
	case remainingBytes <= 15*domain.BytesInGiB:
		return "15 ГБ", "traffic_15gb"
	default:
		return "", ""
	}
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

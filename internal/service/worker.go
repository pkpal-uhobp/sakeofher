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
	if log == nil {
		log = zap.NewNop()
	}
	return &workerService{
		repo:          repo,
		subscriptions: subscriptions,
		payments:      payments,
		notifications: notifications,
		log:           log,
	}
}

func (s *workerService) ExpireSubscriptions(ctx context.Context) error {
	return s.subscriptions.DisableExpiredSubscriptions(ctx, 100)
}

func (s *workerService) DeleteOldDisabledUsers(ctx context.Context) error {
	// Step 1: after delete_after <= now(), remove disabled users from Remnawave
	// and mark them as remna_status='deleted'. This keeps the existing safe flow.
	if err := s.subscriptions.DeleteOldDisabledUsers(ctx, 100); err != nil {
		return err
	}

	// Step 2: after Remnawave deletion succeeded, hard-delete those users from the site DB.
	// subscriptions/payments/broadcast_recipients are removed by FK cascade.
	deleted, err := s.repo.Users.DeleteRemnaDeleted(ctx, 100)
	if err != nil {
		return err
	}
	if deleted > 0 {
		s.log.Info("old disabled users deleted from site db", zap.Int64("deleted", deleted))
	}
	return nil
}

func (s *workerService) RetryFailedActivations(ctx context.Context) error {
	return s.payments.RetryFailedActivations(ctx, 50)
}

func (s *workerService) SyncUsage(ctx context.Context) error {
	return s.subscriptions.SyncRemnaUsage(ctx, 100)
}

func (s *workerService) ResetTrafficPeriods(ctx context.Context) error {
	// First handle users who exhausted the traffic quota before the scheduled
	// period end. For multi-month subscriptions this consumes the next paid
	// traffic period immediately instead of keeping the user blocked until the
	// calendar period ends.
	if advancer, ok := s.subscriptions.(interface {
		AdvanceTrafficExhaustedPeriods(context.Context, int) error
	}); ok {
		if err := advancer.AdvanceTrafficExhaustedPeriods(ctx, 100); err != nil {
			return err
		}
	}

	return s.subscriptions.ResetTrafficPeriods(ctx, 100)
}

func (s *workerService) ReconcileRemnaState(ctx context.Context) error {
	return s.subscriptions.ReconcileRemnaState(ctx, 100)
}

func (s *workerService) NotifyExpiringAndTraffic(ctx context.Context) error {
	now := time.Now()
	if err := s.notifyExpiring(ctx, now); err != nil {
		return err
	}
	if err := s.notifyLowTraffic(ctx); err != nil {
		return err
	}
	if err := s.notifyTrafficExhaustedDaily(ctx, now); err != nil {
		return err
	}
	return nil
}

func (s *workerService) notifyExpiring(ctx context.Context, now time.Time) error {
	items, err := s.repo.Subscriptions.FindExpiringSoonForWorker(ctx, now, now.Add(72*time.Hour), 500)
	if err != nil {
		return err
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

		text := fmt.Sprintf("Подписка скоро закончится: осталось %s.\n\nПродлите доступ в Telegram-боте, чтобы VPN не отключился.", label)
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
	items, err := s.repo.Subscriptions.FindLowTrafficForWorker(ctx, 10*domain.BytesInGiB, 500)
	if err != nil {
		return err
	}

	sent := 0
	for _, item := range items {
		remaining := item.Subscription.TrafficLimitBytes - item.Subscription.TrafficUsedBytes
		if remaining < 0 {
			remaining = 0
		}
		label, key := trafficLabelAndKey(remaining)
		if key != "" {
			key = trafficNotificationPeriodKey(item.Subscription, key)
		}
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

		text := fmt.Sprintf("Заканчивается трафик в текущем периоде: осталось примерно %s.\n\nДата обновления трафика: %s. Если трафик закончится раньше, доступ восстановится после начала нового периода или после продления.", label, item.Subscription.CurrentPeriodEnd.Format("02.01.2006"))
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

func (s *workerService) notifyTrafficExhaustedDaily(ctx context.Context, now time.Time) error {
	items, err := s.repo.Subscriptions.FindTrafficExhaustedForNotifications(ctx, 500)
	if err != nil {
		return err
	}

	dateKey := now.Format("20060102")
	sent := 0
	for _, item := range items {
		key := trafficNotificationPeriodKey(item.Subscription, "traffic_exhausted_daily_"+dateKey)
		alreadySent, err := s.repo.Subscriptions.WasNotificationSent(ctx, item.Subscription.ID, key)
		if err != nil {
			return err
		}
		if alreadySent {
			continue
		}

		text := fmt.Sprintf("Трафик в текущем периоде исчерпан.\n\nДоступ временно остановлен. Он восстановится после обновления периода: %s, либо после продления подписки в боте.", item.Subscription.CurrentPeriodEnd.Format("02.01.2006"))
		if err := s.notifications.Send(ctx, item.User.TelegramID, text); err != nil {
			s.log.Error("send daily traffic exhausted warning failed", zap.Int64("telegram_id", item.User.TelegramID), zap.Error(err))
		}
		if err := s.repo.Subscriptions.MarkNotificationSent(ctx, item.Subscription.ID, key); err != nil {
			return err
		}
		sent++
	}

	s.log.Info("worker notify traffic exhausted daily finished", zap.Int("found", len(items)), zap.Int("sent", sent))
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

func trafficNotificationPeriodKey(sub domain.Subscription, baseKey string) string {
	if baseKey == "" {
		return ""
	}
	periodStart := sub.CurrentPeriodStart
	periodEnd := sub.CurrentPeriodEnd
	if periodStart.IsZero() && periodEnd.IsZero() {
		return baseKey
	}
	if periodStart.IsZero() {
		return fmt.Sprintf("%s_period_end_%s", baseKey, periodEnd.Format("20060102"))
	}
	if periodEnd.IsZero() {
		return fmt.Sprintf("%s_period_start_%s", baseKey, periodStart.Format("20060102"))
	}
	return fmt.Sprintf("%s_period_%s_%s", baseKey, periodStart.Format("20060102"), periodEnd.Format("20060102"))
}

func trafficLabelAndKey(remainingBytes int64) (string, string) {
	switch {
	case remainingBytes <= 5*domain.BytesInGiB:
		return "5 ГБ", "traffic_5gb"
	case remainingBytes <= 10*domain.BytesInGiB:
		return "10 ГБ", "traffic_10gb"
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

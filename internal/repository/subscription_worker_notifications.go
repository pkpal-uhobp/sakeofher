package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"sakeofher/internal/domain"
)

func (r *SubscriptionRepository) WasNotificationSent(ctx context.Context, subscriptionID int64, key string) (bool, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	var exists bool
	err := r.tx.Querier(ctx).QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM subscription_notifications
			WHERE subscription_id = $1
			  AND notification_key = $2
		)
	`, subscriptionID, key).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check subscription notification: %w", err)
	}

	return exists, nil
}

func (r *SubscriptionRepository) MarkNotificationSent(ctx context.Context, subscriptionID int64, key string) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		INSERT INTO subscription_notifications (subscription_id, notification_key, sent_at)
		VALUES ($1, $2, now())
		ON CONFLICT (subscription_id, notification_key) DO NOTHING
	`, subscriptionID, key)
	if err != nil {
		return fmt.Errorf("mark subscription notification sent: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) FindExpiringForNotifications(
	ctx context.Context,
	now time.Time,
	until time.Time,
	limit int,
) ([]domain.PublicSubscription, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	query := publicSubscriptionSelect() + `
		WHERE s.status = 'active'
		  AND s.expires_at > $1
		  AND s.expires_at <= $2
		ORDER BY s.expires_at ASC
		LIMIT $3
	`

	rows, err := r.tx.Querier(ctx).Query(ctx, query, now, until, normalizeLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("find expiring subscriptions for notifications: %w", err)
	}
	defer rows.Close()

	items := make([]domain.PublicSubscription, 0)
	for rows.Next() {
		item, err := scanWorkerPublicSubscriptionRow(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expiring subscriptions for notifications: %w", err)
	}

	return items, nil
}

func (r *SubscriptionRepository) FindLowTrafficForNotifications(
	ctx context.Context,
	remainingBytes int64,
	limit int,
) ([]domain.PublicSubscription, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	query := publicSubscriptionSelect() + `
		WHERE s.status = 'active'
		  AND s.period_status = 'active'
		  AND s.traffic_limit_bytes > 0
		  AND s.traffic_used_bytes > 0
		  AND (s.traffic_limit_bytes - s.traffic_used_bytes) >= 0
		  AND (s.traffic_limit_bytes - s.traffic_used_bytes) <= $1
		ORDER BY (s.traffic_limit_bytes - s.traffic_used_bytes) ASC
		LIMIT $2
	`

	rows, err := r.tx.Querier(ctx).Query(ctx, query, remainingBytes, normalizeLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("find low traffic subscriptions for notifications: %w", err)
	}
	defer rows.Close()

	items := make([]domain.PublicSubscription, 0)
	for rows.Next() {
		item, err := scanWorkerPublicSubscriptionRow(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate low traffic subscriptions for notifications: %w", err)
	}

	return items, nil
}

func scanWorkerPublicSubscriptionRow(row subscriptionScanner) (*domain.PublicSubscription, error) {
	var out domain.PublicSubscription
	err := row.Scan(
		&out.Subscription.ID,
		&out.Subscription.UserID,
		&out.Subscription.TariffID,
		&out.Subscription.LastPaymentID,
		&out.Subscription.Status,
		&out.Subscription.StartedAt,
		&out.Subscription.ExpiresAt,
		&out.Subscription.CurrentPeriodStart,
		&out.Subscription.CurrentPeriodEnd,
		&out.Subscription.TrafficLimitBytes,
		&out.Subscription.TrafficUsedBytes,
		&out.Subscription.PeriodStatus,
		&out.Subscription.PublicToken,
		&out.Subscription.LastRemnaCheckAt,
		&out.Subscription.LastExpireNotificationAt,
		&out.Subscription.LastTrafficNotificationAt,
		&out.Subscription.Notified3Days,
		&out.Subscription.Notified1Day,
		&out.Subscription.NotifiedExpired,
		&out.Subscription.Traffic80Notified,
		&out.Subscription.Traffic95Notified,
		&out.Subscription.TrafficExhaustedNotified,
		&out.Subscription.CreatedAt,
		&out.Subscription.UpdatedAt,

		&out.User.ID,
		&out.User.TelegramID,
		&out.User.TelegramUsername,
		&out.User.TelegramFirstName,
		&out.User.TelegramLastName,
		&out.User.LanguageCode,
		&out.User.Alias,
		&out.User.RemnaUUID,
		&out.User.RemnaUsername,
		&out.User.SubscriptionURL,
		&out.User.Status,
		&out.User.RemnaStatus,
		&out.User.DisabledAt,
		&out.User.DeleteAfter,
		&out.User.DeletedAt,
		&out.User.LastSeenAt,
		&out.User.CreatedAt,
		&out.User.UpdatedAt,

		&out.Tariff.ID,
		&out.Tariff.Code,
		&out.Tariff.Title,
		&out.Tariff.Description,
		&out.Tariff.DurationDays,
		&out.Tariff.PeriodDays,
		&out.Tariff.TrafficLimitBytes,
		&out.Tariff.IsActive,
		&out.Tariff.SortOrder,
		&out.Tariff.CreatedAt,
		&out.Tariff.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scan worker public subscription: %w", err)
	}

	out.SubscriptionURL = out.User.SubscriptionURL
	return &out, nil
}

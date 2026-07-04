package repository

import (
	"context"
	"fmt"
	"time"

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
		item, err := r.scanPublicSubscriptionFromRows(rows)
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
		item, err := r.scanPublicSubscriptionFromRows(rows)
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

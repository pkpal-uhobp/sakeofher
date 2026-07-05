package repository

import (
	"context"
	"fmt"
	"time"

	"sakeofher/internal/domain"
)

// FindExpiringSoonForWorker returns active subscriptions that expire inside the requested window.
// It intentionally does not depend on old boolean notification flags; Worker de-duplicates via
// subscription_notifications keys such as expire_3d / expire_1d.
func (r *SubscriptionRepository) FindExpiringSoonForWorker(
	ctx context.Context,
	now time.Time,
	until time.Time,
	limit int,
) ([]domain.PublicSubscription, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	query := publicSubscriptionSelect() + `
		WHERE s.status = 'active'
		  AND u.status = 'active'
		  AND s.expires_at > $1
		  AND s.expires_at <= $2
		ORDER BY s.expires_at ASC, s.id ASC
		LIMIT $3
	`

	rows, err := r.tx.Querier(ctx).Query(ctx, query, now, until, normalizeLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("find expiring subscriptions for worker notifications: %w", err)
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
		return nil, fmt.Errorf("iterate expiring subscriptions for worker notifications: %w", err)
	}
	return items, nil
}

// FindLowTrafficForWorker returns active subscriptions with positive traffic left but less than threshold.
// Exhausted traffic is handled separately by FindTrafficExhaustedForNotifications.
func (r *SubscriptionRepository) FindLowTrafficForWorker(
	ctx context.Context,
	thresholdBytes int64,
	limit int,
) ([]domain.PublicSubscription, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	query := publicSubscriptionSelect() + `
		WHERE s.status = 'active'
		  AND u.status = 'active'
		  AND s.traffic_limit_bytes > 0
		  AND s.current_period_start <= now()
		  AND s.current_period_end > now()
		  -- Send low-traffic warnings only in the final paid traffic period.
		  -- If current_period_end < expires_at, the user still has another paid period,
		  -- so Worker must silently roll over/reset traffic instead of asking to renew.
		  AND s.current_period_end >= s.expires_at
		  AND GREATEST(s.traffic_limit_bytes - s.traffic_used_bytes, 0) > 0
		  AND GREATEST(s.traffic_limit_bytes - s.traffic_used_bytes, 0) <= $1
		ORDER BY GREATEST(s.traffic_limit_bytes - s.traffic_used_bytes, 0) ASC, s.id ASC
		LIMIT $2
	`

	rows, err := r.tx.Querier(ctx).Query(ctx, query, thresholdBytes, normalizeLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("find low traffic subscriptions for worker notifications: %w", err)
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
		return nil, fmt.Errorf("iterate low traffic subscriptions for worker notifications: %w", err)
	}
	return items, nil
}

// FindTrafficExhaustedForNotifications returns active subscriptions whose traffic is already exhausted.
// Worker uses a date-based notification key, so each user receives no more than one such message per day.
func (r *SubscriptionRepository) FindTrafficExhaustedForNotifications(
	ctx context.Context,
	limit int,
) ([]domain.PublicSubscription, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	query := publicSubscriptionSelect() + `
		WHERE s.status = 'active'
		  AND u.status = 'active'
		  AND s.current_period_start <= now()
		  AND s.current_period_end > now()
		  -- Notify about exhausted traffic only if there is no next paid period.
		  -- If current_period_end < expires_at, AdvanceTrafficExhaustedPeriods
		  -- will open the next paid period and reset traffic silently.
		  AND s.current_period_end >= s.expires_at
		  AND (
		    s.period_status = 'traffic_exhausted'
		    OR (s.traffic_limit_bytes > 0 AND s.traffic_used_bytes >= s.traffic_limit_bytes)
		  )
		ORDER BY s.updated_at ASC, s.id ASC
		LIMIT $1
	`

	rows, err := r.tx.Querier(ctx).Query(ctx, query, normalizeLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("find traffic exhausted subscriptions for notifications: %w", err)
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
		return nil, fmt.Errorf("iterate traffic exhausted subscriptions for notifications: %w", err)
	}

	return items, nil
}

// FindTrafficExhaustedReadyForAdvance returns active subscriptions where the current
// traffic period is exhausted before its scheduled end and there is still paid time left.
func (r *SubscriptionRepository) FindTrafficExhaustedReadyForAdvance(
	ctx context.Context,
	now time.Time,
	limit int,
) ([]domain.PublicSubscription, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	query := publicSubscriptionSelect() + `
		WHERE s.status = 'active'
		  AND u.status = 'active'
		  AND s.traffic_limit_bytes > 0
		  AND s.current_period_start <= $1
		  AND s.current_period_end > $1
		  AND s.current_period_end < s.expires_at
		  AND s.expires_at > $1
		  AND (
		    s.period_status = 'traffic_exhausted'
		    OR s.traffic_used_bytes >= s.traffic_limit_bytes
		  )
		ORDER BY s.current_period_end ASC, s.id ASC
		LIMIT $2
	`

	rows, err := r.tx.Querier(ctx).Query(ctx, query, now, normalizeLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("find traffic exhausted subscriptions ready for period advance: %w", err)
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
		return nil, fmt.Errorf("iterate traffic exhausted subscriptions ready for period advance: %w", err)
	}
	return items, nil
}

// AdvanceTrafficPeriodAfterExhaustion closes the exhausted current period, opens the
// next paid traffic period and resets local traffic counters.
func (r *SubscriptionRepository) AdvanceTrafficPeriodAfterExhaustion(
	ctx context.Context,
	subscriptionID int64,
	nextStart time.Time,
	nextEnd time.Time,
) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE subscriptions
		SET current_period_start = $2,
		    current_period_end = $3,
		    traffic_used_bytes = 0,
		    period_status = 'active',
		    last_traffic_notification_at = NULL,
		    traffic_80_notified = false,
		    traffic_95_notified = false,
		    traffic_exhausted_notified = false,
		    updated_at = now()
		WHERE id = $1
		  AND status = 'active'
	`, subscriptionID, nextStart, nextEnd)
	if err != nil {
		return fmt.Errorf("advance traffic period after exhaustion: %w", err)
	}
	return nil
}

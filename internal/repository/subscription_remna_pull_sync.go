package repository

import (
	"context"
	"fmt"
	"time"
)

// UpdateRemnaPulledState stores values pulled from Remnawave.
// Remnawave can be edited manually, so Worker must be able to pull traffic limit
// and expiration date back into the site DB before reconcile pushes state back.
func (r *SubscriptionRepository) UpdateRemnaPulledState(
	ctx context.Context,
	subscriptionID int64,
	usedBytes int64,
	limitBytes int64,
	expiresAt time.Time,
	checkedAt time.Time,
) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	if usedBytes < 0 {
		usedBytes = 0
	}

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE subscriptions
		SET
			traffic_used_bytes = $2,
			traffic_limit_bytes = CASE
				WHEN $3 > 0 THEN $3
				ELSE traffic_limit_bytes
			END,
			expires_at = CASE
				WHEN NOT $4::timestamptz IS NULL THEN $4
				ELSE expires_at
			END,
			current_period_end = CASE
				WHEN NOT $4::timestamptz IS NULL AND current_period_end > $4 THEN $4
				ELSE current_period_end
			END,
			period_status = CASE
				WHEN (
					CASE WHEN $3 > 0 THEN $3 ELSE traffic_limit_bytes END
				) > 0
				AND $2 < (
					CASE WHEN $3 > 0 THEN $3 ELSE traffic_limit_bytes END
				)
				AND period_status = 'traffic_exhausted'
					THEN 'active'
				ELSE period_status
			END,
			last_remna_check_at = $5,
			updated_at = now()
		WHERE id = $1
	`, subscriptionID, usedBytes, limitBytes, expiresAt, checkedAt)
	if err != nil {
		return fmt.Errorf("update subscription remnawave pulled state: %w", err)
	}
	return nil
}

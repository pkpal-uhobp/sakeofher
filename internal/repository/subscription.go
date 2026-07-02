package repository

import (
	"context"
	"fmt"
	"time"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository/tx"
)

type SubscriptionRepository struct{ tx *tx.Manager }

func NewSubscriptionRepository(txManager *tx.Manager) *SubscriptionRepository {
	return &SubscriptionRepository{tx: txManager}
}

func (r *SubscriptionRepository) CreateActive(ctx context.Context, s *domain.Subscription) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	q := `
        INSERT INTO subscriptions (
            user_id, tariff_id, status, started_at, expires_at, current_period_start,
            current_period_end, traffic_limit_bytes, traffic_used_bytes
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
        RETURNING id, created_at, updated_at
    `
	err := r.tx.Querier(ctx).QueryRow(ctx, q,
		s.UserID, s.TariffID, s.Status, s.StartedAt, s.ExpiresAt, s.CurrentPeriodStart,
		s.CurrentPeriodEnd, s.TrafficLimitBytes, s.TrafficUsedBytes,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create active subscription: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) FindExpiredActive(ctx context.Context, now time.Time, limit int) ([]domain.Subscription, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	rows, err := r.tx.Querier(ctx).Query(ctx, `
        SELECT id, user_id, tariff_id, status, started_at, expires_at, current_period_start, current_period_end,
               traffic_limit_bytes, traffic_used_bytes, notified_80, notified_95, notified_3_days, notified_1_day, created_at, updated_at
        FROM subscriptions
        WHERE status = 'active' AND expires_at <= $1
        ORDER BY expires_at ASC
        LIMIT $2
    `, now, limit)
	if err != nil {
		return nil, fmt.Errorf("find expired active subscriptions: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Subscription, 0)
	for rows.Next() {
		var s domain.Subscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.TariffID, &s.Status, &s.StartedAt, &s.ExpiresAt, &s.CurrentPeriodStart, &s.CurrentPeriodEnd,
			&s.TrafficLimitBytes, &s.TrafficUsedBytes, &s.Notified80, &s.Notified95, &s.Notified3Days, &s.Notified1Day, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan expired subscription: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

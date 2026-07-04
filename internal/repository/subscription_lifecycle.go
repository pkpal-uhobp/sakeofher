package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"sakeofher/internal/domain"
)

func (r *SubscriptionRepository) CreateLifecycleEvent(ctx context.Context, event domain.SubscriptionLifecycleEvent) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	if event.Success == false && event.ErrorText == nil {
		empty := ""
		event.ErrorText = &empty
	}

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		INSERT INTO subscription_lifecycle_events (
			subscription_id,
			user_id,
			payment_id,
			event_type,
			from_status,
			to_status,
			from_period_status,
			to_period_status,
			reason,
			success,
			error_text,
			details
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`,
		event.SubscriptionID,
		event.UserID,
		event.PaymentID,
		event.EventType,
		event.FromStatus,
		event.ToStatus,
		event.FromPeriodStatus,
		event.ToPeriodStatus,
		event.Reason,
		event.Success,
		event.ErrorText,
		event.Details,
	)
	if err != nil {
		return fmt.Errorf("create subscription lifecycle event: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) FindRemnaReconcileCandidates(
	ctx context.Context,
	now time.Time,
	limit int,
) ([]domain.PublicSubscription, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	query := publicSubscriptionSelect() + `
		WHERE u.remna_uuid IS NOT NULL
		  AND u.remna_uuid::text <> ''
		  AND (
		      s.status IN ('active', 'expired', 'cancelled')
		      OR s.period_status IN ('active', 'finished', 'traffic_exhausted')
		      OR s.expires_at <= $1
		  )
		ORDER BY
		  CASE
		    WHEN s.status = 'active' AND s.expires_at <= $1 THEN 0
		    WHEN s.period_status = 'traffic_exhausted' THEN 1
		    WHEN s.status = 'active' THEN 2
		    ELSE 3
		  END,
		  s.updated_at DESC
		LIMIT $2
	`

	rows, err := r.tx.Querier(ctx).Query(ctx, query, now, normalizeLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("find remnawave reconcile candidates: %w", err)
	}
	defer rows.Close()

	items := make([]domain.PublicSubscription, 0)
	for rows.Next() {
		item, err := scanLifecyclePublicSubscriptionRow(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate remnawave reconcile candidates: %w", err)
	}

	return items, nil
}

func scanLifecyclePublicSubscriptionRow(row subscriptionScanner) (*domain.PublicSubscription, error) {
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

		return nil, fmt.Errorf("scan lifecycle public subscription: %w", err)
	}

	out.SubscriptionURL = out.User.SubscriptionURL
	return &out, nil
}

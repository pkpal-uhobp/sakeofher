package repository

import (
	"context"
	"fmt"
	"time"

	"sakeofher/internal/domain"
)

func (r *SubscriptionRepository) FindActiveWithRemna(ctx context.Context, limit int) ([]domain.SubscriptionWithUser, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	rows, err := r.tx.Querier(ctx).Query(ctx, `
		SELECT
			s.id, s.user_id, s.tariff_id, s.last_payment_id, s.status,
			s.started_at, s.expires_at, s.current_period_start, s.current_period_end,
			s.traffic_limit_bytes, s.traffic_used_bytes, s.period_status,
			s.public_token, s.last_remna_check_at, s.last_expire_notification_at,
			s.last_traffic_notification_at, s.notified_3_days, s.notified_1_day,
			s.notified_expired, s.traffic_80_notified, s.traffic_95_notified,
			s.traffic_exhausted_notified, s.created_at, s.updated_at,

			u.id, u.telegram_id, u.telegram_username, u.telegram_first_name,
			u.telegram_last_name, u.language_code, u.alias, u.remna_uuid,
			u.remna_username, u.subscription_url, u.status, u.remna_status,
			u.disabled_at, u.delete_after, u.deleted_at, u.last_seen_at,
			u.created_at, u.updated_at
		FROM subscriptions s
		JOIN users u ON u.id = s.user_id
		WHERE s.status = 'active'
		  AND u.remna_uuid IS NOT NULL
		  AND u.remna_uuid::text <> ''
		ORDER BY COALESCE(s.last_remna_check_at, '1970-01-01'::timestamp) ASC
		LIMIT $1
	`, normalizeLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("find active with remna: %w", err)
	}
	defer rows.Close()

	out := make([]domain.SubscriptionWithUser, 0)
	for rows.Next() {
		var item domain.SubscriptionWithUser
		if err := scanSubscriptionWithUser(rows, &item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active with remna: %w", err)
	}

	return out, nil
}

func (r *SubscriptionRepository) FindReadyForTrafficReset(ctx context.Context, now time.Time, limit int) ([]domain.SubscriptionWithUserAndTariff, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	rows, err := r.tx.Querier(ctx).Query(ctx, `
		SELECT
			s.id, s.user_id, s.tariff_id, s.last_payment_id, s.status,
			s.started_at, s.expires_at, s.current_period_start, s.current_period_end,
			s.traffic_limit_bytes, s.traffic_used_bytes, s.period_status,
			s.public_token, s.last_remna_check_at, s.last_expire_notification_at,
			s.last_traffic_notification_at, s.notified_3_days, s.notified_1_day,
			s.notified_expired, s.traffic_80_notified, s.traffic_95_notified,
			s.traffic_exhausted_notified, s.created_at, s.updated_at,

			u.id, u.telegram_id, u.telegram_username, u.telegram_first_name,
			u.telegram_last_name, u.language_code, u.alias, u.remna_uuid,
			u.remna_username, u.subscription_url, u.status, u.remna_status,
			u.disabled_at, u.delete_after, u.deleted_at, u.last_seen_at,
			u.created_at, u.updated_at,

			t.id, t.code, t.title, t.description, t.duration_days, t.period_days,
			t.traffic_limit_bytes, t.is_active, t.sort_order, t.created_at, t.updated_at
		FROM subscriptions s
		JOIN users u ON u.id = s.user_id
		JOIN tariffs t ON t.id = s.tariff_id
		WHERE s.status = 'active'
		  AND s.expires_at > $1
		  AND s.current_period_end <= $1
		  AND u.remna_uuid IS NOT NULL
		  AND u.remna_uuid::text <> ''
		ORDER BY s.current_period_end ASC
		LIMIT $2
	`, now, normalizeLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("find ready for traffic reset: %w", err)
	}
	defer rows.Close()

	out := make([]domain.SubscriptionWithUserAndTariff, 0)
	for rows.Next() {
		var item domain.SubscriptionWithUserAndTariff

		if err := rows.Scan(
			&item.Subscription.ID,
			&item.Subscription.UserID,
			&item.Subscription.TariffID,
			&item.Subscription.LastPaymentID,
			&item.Subscription.Status,
			&item.Subscription.StartedAt,
			&item.Subscription.ExpiresAt,
			&item.Subscription.CurrentPeriodStart,
			&item.Subscription.CurrentPeriodEnd,
			&item.Subscription.TrafficLimitBytes,
			&item.Subscription.TrafficUsedBytes,
			&item.Subscription.PeriodStatus,
			&item.Subscription.PublicToken,
			&item.Subscription.LastRemnaCheckAt,
			&item.Subscription.LastExpireNotificationAt,
			&item.Subscription.LastTrafficNotificationAt,
			&item.Subscription.Notified3Days,
			&item.Subscription.Notified1Day,
			&item.Subscription.NotifiedExpired,
			&item.Subscription.Traffic80Notified,
			&item.Subscription.Traffic95Notified,
			&item.Subscription.TrafficExhaustedNotified,
			&item.Subscription.CreatedAt,
			&item.Subscription.UpdatedAt,

			&item.User.ID,
			&item.User.TelegramID,
			&item.User.TelegramUsername,
			&item.User.TelegramFirstName,
			&item.User.TelegramLastName,
			&item.User.LanguageCode,
			&item.User.Alias,
			&item.User.RemnaUUID,
			&item.User.RemnaUsername,
			&item.User.SubscriptionURL,
			&item.User.Status,
			&item.User.RemnaStatus,
			&item.User.DisabledAt,
			&item.User.DeleteAfter,
			&item.User.DeletedAt,
			&item.User.LastSeenAt,
			&item.User.CreatedAt,
			&item.User.UpdatedAt,

			&item.Tariff.ID,
			&item.Tariff.Code,
			&item.Tariff.Title,
			&item.Tariff.Description,
			&item.Tariff.DurationDays,
			&item.Tariff.PeriodDays,
			&item.Tariff.TrafficLimitBytes,
			&item.Tariff.IsActive,
			&item.Tariff.SortOrder,
			&item.Tariff.CreatedAt,
			&item.Tariff.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan reset item: %w", err)
		}

		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate traffic reset items: %w", err)
	}

	return out, nil
}

func (r *SubscriptionRepository) UpdateRemnaUsage(ctx context.Context, subscriptionID int64, usedBytes int64, checkedAt time.Time) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE subscriptions
		SET traffic_used_bytes = $2,
		    last_remna_check_at = $3,
		    updated_at = now()
		WHERE id = $1
	`, subscriptionID, usedBytes, checkedAt)
	if err != nil {
		return fmt.Errorf("update remna usage: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) MarkTrafficExhausted(ctx context.Context, subscriptionID int64, usedBytes int64, checkedAt time.Time) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE subscriptions
		SET traffic_used_bytes = $2,
		    period_status = 'traffic_exhausted',
		    traffic_exhausted_notified = true,
		    last_remna_check_at = $3,
		    updated_at = now()
		WHERE id = $1
	`, subscriptionID, usedBytes, checkedAt)
	if err != nil {
		return fmt.Errorf("mark traffic exhausted: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) ResetTrafficPeriod(ctx context.Context, subscriptionID int64, start time.Time, end time.Time) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE subscriptions
		SET current_period_start = $2,
		    current_period_end = $3,
		    traffic_used_bytes = 0,
		    period_status = 'active',
		    traffic_80_notified = false,
		    traffic_95_notified = false,
		    traffic_exhausted_notified = false,
		    last_traffic_notification_at = NULL,
		    last_remna_check_at = now(),
		    updated_at = now()
		WHERE id = $1
	`, subscriptionID, start, end)
	if err != nil {
		return fmt.Errorf("reset traffic period: %w", err)
	}

	return nil
}

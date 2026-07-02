package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

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
			user_id, tariff_id, last_payment_id, status, started_at, expires_at, current_period_start,
			current_period_end, traffic_limit_bytes, traffic_used_bytes, period_status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, public_token, created_at, updated_at
	`
	err := r.tx.Querier(ctx).QueryRow(ctx, q,
		s.UserID, s.TariffID, s.LastPaymentID, s.Status, s.StartedAt, s.ExpiresAt, s.CurrentPeriodStart,
		s.CurrentPeriodEnd, s.TrafficLimitBytes, s.TrafficUsedBytes, s.PeriodStatus,
	).Scan(&s.ID, &s.PublicToken, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create active subscription: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) GetActiveByUserID(ctx context.Context, userID int64) (*domain.Subscription, error) {
	return r.getOne(ctx, baseSubscriptionSelect()+` WHERE user_id = $1 AND status = 'active'`, userID)
}

func (r *SubscriptionRepository) GetLatestByUserID(ctx context.Context, userID int64) (*domain.Subscription, error) {
	return r.getOne(ctx, baseSubscriptionSelect()+` WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, userID)
}

func (r *SubscriptionRepository) GetActiveByUserIDForUpdate(ctx context.Context, userID int64) (*domain.Subscription, error) {
	return r.getOne(ctx, baseSubscriptionSelect()+` WHERE user_id = $1 AND status = 'active' FOR UPDATE`, userID)
}

func (r *SubscriptionRepository) GetLatestByUserIDForUpdate(ctx context.Context, userID int64) (*domain.Subscription, error) {
	return r.getOne(ctx, baseSubscriptionSelect()+` WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1 FOR UPDATE`, userID)
}

func (r *SubscriptionRepository) GetPublicByToken(ctx context.Context, token string) (*domain.PublicSubscription, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	q := publicSubscriptionSelect() + ` WHERE s.public_token = $1`
	out, err := r.scanPublicSubscription(ctx, q, token)
	if err != nil {
		return nil, fmt.Errorf("get public subscription by token: %w", err)
	}
	return out, nil
}

func (r *SubscriptionRepository) GetActivePublicByTelegramID(ctx context.Context, telegramID int64) (*domain.PublicSubscription, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	q := publicSubscriptionSelect() + `
		WHERE u.telegram_id = $1 AND s.status = 'active'
		ORDER BY s.created_at DESC
		LIMIT 1
	`
	out, err := r.scanPublicSubscription(ctx, q, telegramID)
	if err != nil {
		return nil, fmt.Errorf("get active public subscription by telegram id: %w", err)
	}
	return out, nil
}

func (r *SubscriptionRepository) GetLatestPublicByTelegramID(ctx context.Context, telegramID int64) (*domain.PublicSubscription, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	q := publicSubscriptionSelect() + `
		WHERE u.telegram_id = $1
		ORDER BY s.created_at DESC
		LIMIT 1
	`
	out, err := r.scanPublicSubscription(ctx, q, telegramID)
	if err != nil {
		return nil, fmt.Errorf("get latest public subscription by telegram id: %w", err)
	}
	return out, nil
}

func (r *SubscriptionRepository) ExtendActive(ctx context.Context, s *domain.Subscription) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE subscriptions
		SET tariff_id = $2,
			last_payment_id = $3,
			expires_at = $4,
			current_period_start = $5,
			current_period_end = $6,
			traffic_limit_bytes = $7,
			period_status = 'active',
			status = 'active',
			updated_at = now()
		WHERE id = $1
	`, s.ID, s.TariffID, s.LastPaymentID, s.ExpiresAt, s.CurrentPeriodStart, s.CurrentPeriodEnd, s.TrafficLimitBytes)
	if err != nil {
		return fmt.Errorf("extend active subscription: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) MarkExpired(ctx context.Context, subscriptionID int64) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE subscriptions
		SET status = 'expired',
			period_status = 'finished',
			updated_at = now()
		WHERE id = $1
	`, subscriptionID)
	if err != nil {
		return fmt.Errorf("mark subscription expired: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) FindExpiredActiveWithUsers(ctx context.Context, now time.Time, limit int) ([]domain.SubscriptionWithUser, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	rows, err := r.tx.Querier(ctx).Query(ctx, `
		SELECT
			s.id, s.user_id, s.tariff_id, s.last_payment_id, s.status, s.started_at, s.expires_at,
			s.current_period_start, s.current_period_end, s.traffic_limit_bytes, s.traffic_used_bytes,
			s.period_status, s.public_token, s.last_remna_check_at, s.last_expire_notification_at,
			s.last_traffic_notification_at, s.notified_3_days, s.notified_1_day, s.notified_expired,
			s.traffic_80_notified, s.traffic_95_notified, s.traffic_exhausted_notified,
			s.created_at, s.updated_at,
			u.id, u.telegram_id, u.telegram_username, u.telegram_first_name, u.telegram_last_name, u.language_code,
			u.alias, u.remna_uuid, u.remna_username, u.subscription_url, u.status, u.remna_status,
			u.disabled_at, u.delete_after, u.deleted_at, u.last_seen_at, u.created_at, u.updated_at
		FROM subscriptions s
		JOIN users u ON u.id = s.user_id
		WHERE s.status = 'active' AND s.expires_at <= $1
		ORDER BY s.expires_at ASC
		LIMIT $2
	`, now, limit)
	if err != nil {
		return nil, fmt.Errorf("find expired active subscriptions: %w", err)
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
		return nil, fmt.Errorf("iterate expired subscriptions: %w", err)
	}
	return out, nil
}

func (r *SubscriptionRepository) scanPublicSubscription(ctx context.Context, query string, args ...any) (*domain.PublicSubscription, error) {
	var out domain.PublicSubscription
	err := r.tx.Querier(ctx).QueryRow(ctx, query, args...).Scan(
		&out.Subscription.ID, &out.Subscription.UserID, &out.Subscription.TariffID, &out.Subscription.LastPaymentID, &out.Subscription.Status, &out.Subscription.StartedAt, &out.Subscription.ExpiresAt,
		&out.Subscription.CurrentPeriodStart, &out.Subscription.CurrentPeriodEnd, &out.Subscription.TrafficLimitBytes, &out.Subscription.TrafficUsedBytes,
		&out.Subscription.PeriodStatus, &out.Subscription.PublicToken, &out.Subscription.LastRemnaCheckAt, &out.Subscription.LastExpireNotificationAt,
		&out.Subscription.LastTrafficNotificationAt, &out.Subscription.Notified3Days, &out.Subscription.Notified1Day, &out.Subscription.NotifiedExpired,
		&out.Subscription.Traffic80Notified, &out.Subscription.Traffic95Notified, &out.Subscription.TrafficExhaustedNotified,
		&out.Subscription.CreatedAt, &out.Subscription.UpdatedAt,
		&out.User.ID, &out.User.TelegramID, &out.User.TelegramUsername, &out.User.TelegramFirstName, &out.User.TelegramLastName, &out.User.LanguageCode,
		&out.User.Alias, &out.User.RemnaUUID, &out.User.RemnaUsername, &out.User.SubscriptionURL, &out.User.Status, &out.User.RemnaStatus,
		&out.User.DisabledAt, &out.User.DeleteAfter, &out.User.DeletedAt, &out.User.LastSeenAt, &out.User.CreatedAt, &out.User.UpdatedAt,
		&out.Tariff.ID, &out.Tariff.Code, &out.Tariff.Title, &out.Tariff.Description, &out.Tariff.DurationDays, &out.Tariff.PeriodDays,
		&out.Tariff.TrafficLimitBytes, &out.Tariff.IsActive, &out.Tariff.SortOrder, &out.Tariff.CreatedAt, &out.Tariff.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	out.SubscriptionURL = out.User.SubscriptionURL
	return &out, nil
}

func (r *SubscriptionRepository) getOne(ctx context.Context, query string, args ...any) (*domain.Subscription, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	s, err := scanSubscription(r.tx.Querier(ctx).QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return s, nil
}

func baseSubscriptionSelect() string {
	return `
		SELECT id, user_id, tariff_id, last_payment_id, status, started_at, expires_at, current_period_start, current_period_end,
			traffic_limit_bytes, traffic_used_bytes, period_status, public_token, last_remna_check_at,
			last_expire_notification_at, last_traffic_notification_at, notified_3_days, notified_1_day,
			notified_expired, traffic_80_notified, traffic_95_notified, traffic_exhausted_notified, created_at, updated_at
		FROM subscriptions
	`
}

func publicSubscriptionSelect() string {
	return `
		SELECT
			s.id, s.user_id, s.tariff_id, s.last_payment_id, s.status, s.started_at, s.expires_at,
			s.current_period_start, s.current_period_end, s.traffic_limit_bytes, s.traffic_used_bytes,
			s.period_status, s.public_token, s.last_remna_check_at, s.last_expire_notification_at,
			s.last_traffic_notification_at, s.notified_3_days, s.notified_1_day, s.notified_expired,
			s.traffic_80_notified, s.traffic_95_notified, s.traffic_exhausted_notified,
			s.created_at, s.updated_at,
			u.id, u.telegram_id, u.telegram_username, u.telegram_first_name, u.telegram_last_name, u.language_code,
			u.alias, u.remna_uuid, u.remna_username, u.subscription_url, u.status, u.remna_status,
			u.disabled_at, u.delete_after, u.deleted_at, u.last_seen_at, u.created_at, u.updated_at,
			t.id, t.code, t.title, t.description, t.duration_days, t.period_days, t.traffic_limit_bytes, t.is_active, t.sort_order, t.created_at, t.updated_at
		FROM subscriptions s
		JOIN users u ON u.id = s.user_id
		LEFT JOIN payments p ON p.id = s.last_payment_id
		JOIN tariffs t ON t.id = COALESCE(s.tariff_id, p.tariff_id)
	`
}

type subscriptionScanner interface{ Scan(dest ...any) error }

func scanSubscription(row subscriptionScanner) (*domain.Subscription, error) {
	var s domain.Subscription
	if err := row.Scan(
		&s.ID, &s.UserID, &s.TariffID, &s.LastPaymentID, &s.Status, &s.StartedAt, &s.ExpiresAt, &s.CurrentPeriodStart, &s.CurrentPeriodEnd,
		&s.TrafficLimitBytes, &s.TrafficUsedBytes, &s.PeriodStatus, &s.PublicToken, &s.LastRemnaCheckAt,
		&s.LastExpireNotificationAt, &s.LastTrafficNotificationAt, &s.Notified3Days, &s.Notified1Day,
		&s.NotifiedExpired, &s.Traffic80Notified, &s.Traffic95Notified, &s.TrafficExhaustedNotified, &s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan subscription: %w", err)
	}
	return &s, nil
}

func scanSubscriptionWithUser(row subscriptionScanner, item *domain.SubscriptionWithUser) error {
	return row.Scan(
		&item.Subscription.ID, &item.Subscription.UserID, &item.Subscription.TariffID, &item.Subscription.LastPaymentID, &item.Subscription.Status, &item.Subscription.StartedAt, &item.Subscription.ExpiresAt,
		&item.Subscription.CurrentPeriodStart, &item.Subscription.CurrentPeriodEnd, &item.Subscription.TrafficLimitBytes, &item.Subscription.TrafficUsedBytes,
		&item.Subscription.PeriodStatus, &item.Subscription.PublicToken, &item.Subscription.LastRemnaCheckAt, &item.Subscription.LastExpireNotificationAt,
		&item.Subscription.LastTrafficNotificationAt, &item.Subscription.Notified3Days, &item.Subscription.Notified1Day, &item.Subscription.NotifiedExpired,
		&item.Subscription.Traffic80Notified, &item.Subscription.Traffic95Notified, &item.Subscription.TrafficExhaustedNotified,
		&item.Subscription.CreatedAt, &item.Subscription.UpdatedAt,
		&item.User.ID, &item.User.TelegramID, &item.User.TelegramUsername, &item.User.TelegramFirstName, &item.User.TelegramLastName, &item.User.LanguageCode,
		&item.User.Alias, &item.User.RemnaUUID, &item.User.RemnaUsername, &item.User.SubscriptionURL, &item.User.Status, &item.User.RemnaStatus,
		&item.User.DisabledAt, &item.User.DeleteAfter, &item.User.DeletedAt, &item.User.LastSeenAt, &item.User.CreatedAt, &item.User.UpdatedAt,
	)
}

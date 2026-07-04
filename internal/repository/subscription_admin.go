package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sakeofher/internal/domain"
)

func (r *SubscriptionRepository) ListPublic(ctx context.Context, input domain.SubscriptionListInput) ([]domain.PublicSubscription, int64, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	limit := normalizeLimit(input.Limit)
	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	where := make([]string, 0)
	args := make([]any, 0)

	if input.UserID > 0 {
		args = append(args, input.UserID)
		where = append(where, fmt.Sprintf("s.user_id = $%d", len(args)))
	}

	if input.TelegramID > 0 {
		args = append(args, input.TelegramID)
		where = append(where, fmt.Sprintf("u.telegram_id = $%d", len(args)))
	}

	if input.Status != "" {
		args = append(args, input.Status)
		where = append(where, fmt.Sprintf("s.status = $%d", len(args)))
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	countQuery := `
		SELECT count(*)
		FROM subscriptions s
		JOIN users u ON u.id = s.user_id
		LEFT JOIN payments p ON p.id = s.last_payment_id
		JOIN tariffs t ON t.id = COALESCE(s.tariff_id, p.tariff_id)
	` + whereSQL

	var total int64
	if err := r.tx.Querier(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count subscriptions: %w", err)
	}

	args = append(args, limit, offset)
	query := publicSubscriptionSelect() + whereSQL + fmt.Sprintf(" ORDER BY s.created_at DESC, s.id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.tx.Querier(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	items := make([]domain.PublicSubscription, 0)
	for rows.Next() {
		item, err := r.scanPublicSubscriptionFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate subscriptions: %w", err)
	}

	return items, total, nil
}

func (r *SubscriptionRepository) GetPublicByID(ctx context.Context, id int64) (*domain.PublicSubscription, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	return r.scanPublicSubscription(ctx, publicSubscriptionSelect()+" WHERE s.id = $1", id)
}

func (r *SubscriptionRepository) GetByIDForUpdate(ctx context.Context, id int64) (*domain.Subscription, error) {
	return r.getOne(ctx, baseSubscriptionSelect()+" WHERE id = $1 FOR UPDATE", id)
}

func (r *SubscriptionRepository) UpdateTrafficLimit(ctx context.Context, id int64, trafficLimitBytes int64) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE subscriptions
		SET traffic_limit_bytes = $2,
		    updated_at = now()
		WHERE id = $1
	`, id, trafficLimitBytes)
	if err != nil {
		return fmt.Errorf("update subscription traffic limit: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) UpdateTrafficUsed(ctx context.Context, id int64, trafficUsedBytes int64) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE subscriptions
		SET traffic_used_bytes = $2,
		    updated_at = now()
		WHERE id = $1
	`, id, trafficUsedBytes)
	if err != nil {
		return fmt.Errorf("update subscription traffic used: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) ExtendByID(ctx context.Context, id int64, tariffID int64, expiresAt time.Time, currentPeriodStart time.Time, currentPeriodEnd time.Time, trafficLimitBytes int64) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE subscriptions
		SET tariff_id = $2,
		    expires_at = $3,
		    current_period_start = $4,
		    current_period_end = $5,
		    traffic_limit_bytes = $6,
		    status = 'active',
		    period_status = 'active',
		    updated_at = now()
		WHERE id = $1
	`, id, tariffID, expiresAt, currentPeriodStart, currentPeriodEnd, trafficLimitBytes)
	if err != nil {
		return fmt.Errorf("extend subscription by id: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) SetStatus(ctx context.Context, id int64, status domain.SubscriptionStatus, periodStatus domain.PeriodStatus) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE subscriptions
		SET status = $2,
		    period_status = $3,
		    updated_at = now()
		WHERE id = $1
	`, id, status, periodStatus)
	if err != nil {
		return fmt.Errorf("set subscription status: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) UpdateManual(ctx context.Context, id int64, input domain.UpdateSubscriptionInput) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	var trafficLimitBytes *int64
	if input.TrafficLimitGB != nil {
		value := domain.TrafficGBToBytes(*input.TrafficLimitGB)
		trafficLimitBytes = &value
	}

	var trafficUsedBytes *int64
	if input.TrafficUsedGB != nil {
		value := domain.TrafficGBToBytes(*input.TrafficUsedGB)
		trafficUsedBytes = &value
	}

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE subscriptions
		SET status = COALESCE($2, status),
		    period_status = COALESCE($3, period_status),
		    expires_at = COALESCE($4, expires_at),
		    current_period_start = COALESCE($5, current_period_start),
		    current_period_end = COALESCE($6, current_period_end),
		    traffic_limit_bytes = COALESCE($7, traffic_limit_bytes),
		    traffic_used_bytes = COALESCE($8, traffic_used_bytes),
		    updated_at = now()
		WHERE id = $1
	`, id, input.Status, input.PeriodStatus, input.ExpiresAt, input.CurrentPeriodStart, input.CurrentPeriodEnd, trafficLimitBytes, trafficUsedBytes)
	if err != nil {
		return fmt.Errorf("update subscription manual: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) scanPublicSubscriptionFromRows(row subscriptionScanner) (*domain.PublicSubscription, error) {
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
		return nil, fmt.Errorf("scan public subscription: %w", err)
	}

	out.SubscriptionURL = out.User.SubscriptionURL
	return &out, nil
}

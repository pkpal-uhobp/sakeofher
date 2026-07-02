package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository/tx"
)

type PaymentRepository struct{ tx *tx.Manager }

func NewPaymentRepository(txManager *tx.Manager) *PaymentRepository {
	return &PaymentRepository{tx: txManager}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	q := `
		INSERT INTO payments (
			user_id, tariff_id, tariff_price_id, provider, payment_method,
			currency, amount_minor, stars_amount, duration_days, period_days, traffic_limit_bytes,
			status, provider_payment_id, payment_url, expires_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id, created_at, updated_at
	`
	err := r.tx.Querier(ctx).QueryRow(ctx, q,
		p.UserID, p.TariffID, p.TariffPriceID, p.Provider, p.PaymentMethod,
		p.Currency, p.AmountMinor, p.StarsAmount, p.DurationDays, p.PeriodDays, p.TrafficLimitBytes,
		p.Status, p.ProviderPaymentID, p.PaymentURL, p.ExpiresAt,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create payment: %w", err)
	}
	return nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, paymentID int64) (*domain.Payment, error) {
	return r.getOne(ctx, basePaymentSelect()+` WHERE id = $1`, paymentID)
}

func (r *PaymentRepository) GetByIDForUpdate(ctx context.Context, paymentID int64) (*domain.Payment, error) {
	return r.getOne(ctx, basePaymentSelect()+` WHERE id = $1 FOR UPDATE`, paymentID)
}

func (r *PaymentRepository) GetByProviderPaymentIDForUpdate(ctx context.Context, provider domain.PaymentProvider, providerPaymentID string) (*domain.Payment, error) {
	return r.getOne(ctx, basePaymentSelect()+` WHERE provider = $1 AND provider_payment_id = $2 FOR UPDATE`, provider, providerPaymentID)
}

func (r *PaymentRepository) MarkWaitingPayment(ctx context.Context, paymentID int64, providerPaymentID *string, paymentURL *string, expiresAt *time.Time) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE payments
		SET status = $2,
			provider_payment_id = COALESCE($3, provider_payment_id),
			payment_url = COALESCE($4, payment_url),
			expires_at = COALESCE($5, expires_at),
			updated_at = now()
		WHERE id = $1
	`, paymentID, domain.PaymentStatusWaitingPayment, providerPaymentID, paymentURL, expiresAt)
	if err != nil {
		return fmt.Errorf("mark payment waiting: %w", err)
	}
	return nil
}

func (r *PaymentRepository) MarkPaid(ctx context.Context, paymentID int64, providerPaymentID string, paidAt time.Time, rawPayload json.RawMessage) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE payments
		SET status = $2,
			provider_payment_id = $3,
			paid_at = $4,
			raw_payload = $5::jsonb,
			updated_at = now()
		WHERE id = $1
	`, paymentID, domain.PaymentStatusPaid, providerPaymentID, paidAt, jsonPayload(rawPayload))
	if err != nil {
		return fmt.Errorf("mark payment paid: %w", err)
	}
	return nil
}

func (r *PaymentRepository) MarkActivated(ctx context.Context, paymentID int64, activatedAt time.Time) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE payments
		SET status = $2,
			activated_at = $3,
			updated_at = now()
		WHERE id = $1
	`, paymentID, domain.PaymentStatusActivated, activatedAt)
	if err != nil {
		return fmt.Errorf("mark payment activated: %w", err)
	}
	return nil
}

func (r *PaymentRepository) MarkActivationFailed(ctx context.Context, paymentID int64, rawErr error) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	payload := map[string]string{"error": rawErr.Error()}
	b, _ := json.Marshal(payload)

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE payments
		SET status = $2,
			raw_payload = $3::jsonb,
			updated_at = now()
		WHERE id = $1
	`, paymentID, domain.PaymentStatusActivationFailed, string(b))
	if err != nil {
		return fmt.Errorf("mark payment activation failed: %w", err)
	}
	return nil
}

func (r *PaymentRepository) FindActivationFailed(ctx context.Context, limit int) ([]domain.Payment, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	rows, err := r.tx.Querier(ctx).Query(ctx, basePaymentSelect()+`
		WHERE status IN ('paid', 'activation_failed')
		ORDER BY updated_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("find activation failed payments: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Payment, 0)
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate activation failed payments: %w", err)
	}
	return items, nil
}

func (r *PaymentRepository) getOne(ctx context.Context, query string, args ...any) (*domain.Payment, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	p, err := scanPayment(r.tx.Querier(ctx).QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func basePaymentSelect() string {
	return `
		SELECT id, user_id, tariff_id, tariff_price_id, provider, payment_method,
			currency, amount_minor, stars_amount::bigint, duration_days, period_days, traffic_limit_bytes,
			status, provider_payment_id, payment_url, paid_asset, paid_amount::text, fee_asset, fee_amount::text,
			expires_at, paid_at, activated_at, raw_payload, created_at, updated_at
		FROM payments
	`
}

type paymentScanner interface{ Scan(dest ...any) error }

func scanPayment(row paymentScanner) (*domain.Payment, error) {
	var p domain.Payment
	if err := row.Scan(
		&p.ID, &p.UserID, &p.TariffID, &p.TariffPriceID, &p.Provider, &p.PaymentMethod,
		&p.Currency, &p.AmountMinor, &p.StarsAmount, &p.DurationDays, &p.PeriodDays, &p.TrafficLimitBytes,
		&p.Status, &p.ProviderPaymentID, &p.PaymentURL, &p.PaidAsset, &p.PaidAmount, &p.FeeAsset, &p.FeeAmount,
		&p.ExpiresAt, &p.PaidAt, &p.ActivatedAt, &p.RawPayload, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan payment: %w", err)
	}
	return &p, nil
}

func jsonPayload(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "null"
	}
	return string(raw)
}

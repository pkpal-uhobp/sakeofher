package repository

import (
	"context"
	"fmt"
	"time"

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
            user_id, tariff_id, tariff_price_id, provider, payment_method, status,
            currency, amount_minor, stars_amount, provider_payment_id, payment_url, expires_at
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
        RETURNING id, created_at, updated_at
    `
	err := r.tx.Querier(ctx).QueryRow(ctx, q,
		p.UserID, p.TariffID, p.TariffPriceID, p.Provider, p.PaymentMethod, p.Status,
		p.Currency, p.AmountMinor, p.StarsAmount, p.ProviderPaymentID, p.PaymentURL, p.ExpiresAt,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create payment: %w", err)
	}
	return nil
}

func (r *PaymentRepository) MarkPaid(ctx context.Context, paymentID int64, providerPaymentID string, paidAt time.Time, rawPayload []byte) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	q := `
        UPDATE payments
        SET status = $2,
            provider_payment_id = $3,
            paid_at = $4,
            raw_payload = $5,
            updated_at = now()
        WHERE id = $1
    `
	_, err := r.tx.Querier(ctx).Exec(ctx, q, paymentID, domain.PaymentStatusPaid, providerPaymentID, paidAt, rawPayload)
	if err != nil {
		return fmt.Errorf("mark payment paid: %w", err)
	}
	return nil
}

func (r *PaymentRepository) MarkActivated(ctx context.Context, paymentID int64) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `UPDATE payments SET status = $2, updated_at = now() WHERE id = $1`, paymentID, domain.PaymentStatusActivated)
	if err != nil {
		return fmt.Errorf("mark payment activated: %w", err)
	}
	return nil
}

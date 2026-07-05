package repository

import (
	"context"
	"fmt"

	"sakeofher/internal/domain"
)

func (r *PaymentRepository) FindWaitingByProvider(
	ctx context.Context,
	provider domain.PaymentProvider,
	limit int,
) ([]domain.Payment, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	rows, err := r.tx.Querier(ctx).Query(ctx, basePaymentSelect()+`
		WHERE provider = $1
		  AND status = 'waiting_payment'
		  AND provider_payment_id IS NOT NULL
		  AND (expires_at IS NULL OR expires_at > now())
		ORDER BY updated_at ASC
		LIMIT $2
	`, provider, limit)
	if err != nil {
		return nil, fmt.Errorf("find waiting payments by provider: %w", err)
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
		return nil, fmt.Errorf("iterate waiting payments by provider: %w", err)
	}

	return items, nil
}

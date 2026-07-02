package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository/tx"
)

type PaymentEventRepository struct{ tx *tx.Manager }

func NewPaymentEventRepository(txManager *tx.Manager) *PaymentEventRepository {
	return &PaymentEventRepository{tx: txManager}
}

func (r *PaymentEventRepository) CreateOnce(ctx context.Context, e domain.PaymentEvent) (bool, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	q := `
        INSERT INTO payment_events (provider, event_id, payment_id, event_type, raw_payload)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.tx.Querier(ctx).Exec(ctx, q, e.Provider, e.EventID, e.PaymentID, e.EventType, e.RawPayload)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return false, nil
		}
		return false, fmt.Errorf("create payment event: %w", err)
	}
	return true, nil
}

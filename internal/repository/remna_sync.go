package repository

import (
	"context"
	"fmt"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository/tx"
)

type RemnaSyncRepository struct{ tx *tx.Manager }

func NewRemnaSyncRepository(txManager *tx.Manager) *RemnaSyncRepository {
	return &RemnaSyncRepository{tx: txManager}
}

func (r *RemnaSyncRepository) Create(ctx context.Context, l domain.RemnaSyncLog) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		INSERT INTO remna_sync_logs (
			user_id, subscription_id, payment_id, action, success, error_text, request_payload, response_payload
		) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8::jsonb)
	`, l.UserID, l.SubscriptionID, l.PaymentID, l.Action, l.Success, l.ErrorText, jsonPayload(l.RequestPayload), jsonPayload(l.ResponsePayload))
	if err != nil {
		return fmt.Errorf("create remna sync log: %w", err)
	}
	return nil
}

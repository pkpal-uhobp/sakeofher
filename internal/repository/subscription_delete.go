package repository

import (
	"context"
	"fmt"

	"sakeofher/internal/domain"
)

func (r *SubscriptionRepository) Delete(ctx context.Context, id int64) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	tag, err := r.tx.Querier(ctx).Exec(ctx, `
		DELETE FROM subscriptions
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

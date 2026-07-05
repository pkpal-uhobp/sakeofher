package repository

import (
	"context"
	"fmt"
)

// DeleteRemnaDeleted permanently removes users that were already deleted from Remnawave.
// Related subscriptions, payments and broadcast recipients are removed by ON DELETE CASCADE.
func (r *UserRepository) DeleteRemnaDeleted(ctx context.Context, limit int) (int64, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	tag, err := r.tx.Querier(ctx).Exec(ctx, `
		DELETE FROM users
		WHERE id IN (
			SELECT id
			FROM users
			WHERE remna_status = 'deleted'
			  AND deleted_at IS NOT NULL
			ORDER BY deleted_at ASC, id ASC
			LIMIT $1
		)
	`, normalizeLimit(limit))
	if err != nil {
		return 0, fmt.Errorf("delete remna deleted users from site db: %w", err)
	}

	return tag.RowsAffected(), nil
}

package repository

import (
	"context"
	"fmt"

	"sakeofher/internal/repository/tx"
)

type AdminRepository struct{ tx *tx.Manager }

func NewAdminRepository(txManager *tx.Manager) *AdminRepository {
	return &AdminRepository{tx: txManager}
}

func (r *AdminRepository) IsActiveByTelegramID(ctx context.Context, telegramID int64) (bool, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	var exists bool
	err := r.tx.Querier(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM admins
			WHERE telegram_id = $1
				AND is_active = true
		)
	`, telegramID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check active admin: %w", err)
	}
	return exists, nil
}

func (r *AdminRepository) MarkLogin(ctx context.Context, telegramID int64) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE admins
		SET last_login_at = now(),
			updated_at = now()
		WHERE telegram_id = $1
	`, telegramID)
	if err != nil {
		return fmt.Errorf("mark admin login: %w", err)
	}
	return nil
}

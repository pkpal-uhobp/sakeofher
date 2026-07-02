package repository

import (
	"context"
	"fmt"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository/tx"
)

type UserRepository struct{ tx *tx.Manager }

func NewUserRepository(txManager *tx.Manager) *UserRepository { return &UserRepository{tx: txManager} }

func (r *UserRepository) CreateTelegramUser(ctx context.Context, input domain.TelegramUserInput) (*domain.User, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	q := `
        INSERT INTO users (telegram_id, telegram_username, first_name, last_name)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (telegram_id) DO UPDATE SET
            telegram_username = EXCLUDED.telegram_username,
            first_name = EXCLUDED.first_name,
            last_name = EXCLUDED.last_name,
            updated_at = now()
        RETURNING id, telegram_id, telegram_username, first_name, last_name, remna_uuid, remna_username,
                  subscription_url, public_token, remna_status, disabled_at, delete_after, deleted_at, created_at, updated_at
    `

	var u domain.User
	err := r.tx.Querier(ctx).QueryRow(ctx, q, input.TelegramID, input.TelegramUsername, input.FirstName, input.LastName).Scan(
		&u.ID, &u.TelegramID, &u.TelegramUsername, &u.FirstName, &u.LastName, &u.RemnaUUID, &u.RemnaUsername,
		&u.SubscriptionURL, &u.PublicToken, &u.RemnaStatus, &u.DisabledAt, &u.DeleteAfter, &u.DeletedAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create telegram user: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) SetRemnaData(ctx context.Context, userID int64, data domain.RemnaUserData) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	q := `
        UPDATE users
        SET remna_uuid = $2,
            remna_username = $3,
            subscription_url = $4,
            remna_status = $5,
            disabled_at = NULL,
            delete_after = NULL,
            deleted_at = NULL,
            updated_at = now()
        WHERE id = $1
    `
	_, err := r.tx.Querier(ctx).Exec(ctx, q, userID, data.UUID, data.Username, data.SubscriptionURL, data.Status)
	if err != nil {
		return fmt.Errorf("set remna data: %w", err)
	}
	return nil
}

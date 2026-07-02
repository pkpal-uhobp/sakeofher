package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository/tx"
)

type UserRepository struct{ tx *tx.Manager }

func NewUserRepository(txManager *tx.Manager) *UserRepository { return &UserRepository{tx: txManager} }

func (r *UserRepository) CreateOrUpdateTelegramUser(ctx context.Context, input domain.TelegramUserInput) (*domain.User, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	q := `
		INSERT INTO users (telegram_id, telegram_username, telegram_first_name, telegram_last_name, language_code, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (telegram_id) DO UPDATE SET
			telegram_username = EXCLUDED.telegram_username,
			telegram_first_name = EXCLUDED.telegram_first_name,
			telegram_last_name = EXCLUDED.telegram_last_name,
			language_code = EXCLUDED.language_code,
			last_seen_at = now(),
			updated_at = now()
		RETURNING id, telegram_id, telegram_username, telegram_first_name, telegram_last_name, language_code,
			alias, remna_uuid, remna_username, subscription_url, status, remna_status,
			disabled_at, delete_after, deleted_at, last_seen_at, created_at, updated_at
	`

	var u domain.User
	err := r.tx.Querier(ctx).QueryRow(ctx, q,
		input.TelegramID,
		input.TelegramUsername,
		input.TelegramFirstName,
		input.TelegramLastName,
		input.LanguageCode,
	).Scan(
		&u.ID, &u.TelegramID, &u.TelegramUsername, &u.TelegramFirstName, &u.TelegramLastName, &u.LanguageCode,
		&u.Alias, &u.RemnaUUID, &u.RemnaUsername, &u.SubscriptionURL, &u.Status, &u.RemnaStatus,
		&u.DisabledAt, &u.DeleteAfter, &u.DeletedAt, &u.LastSeenAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create or update telegram user: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return r.getOne(ctx, `
		SELECT id, telegram_id, telegram_username, telegram_first_name, telegram_last_name, language_code,
			alias, remna_uuid, remna_username, subscription_url, status, remna_status,
			disabled_at, delete_after, deleted_at, last_seen_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id)
}

func (r *UserRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	return r.getOne(ctx, `
		SELECT id, telegram_id, telegram_username, telegram_first_name, telegram_last_name, language_code,
			alias, remna_uuid, remna_username, subscription_url, status, remna_status,
			disabled_at, delete_after, deleted_at, last_seen_at, created_at, updated_at
		FROM users
		WHERE telegram_id = $1
	`, telegramID)
}

func (r *UserRepository) GetByIDForUpdate(ctx context.Context, id int64) (*domain.User, error) {
	return r.getOne(ctx, `
		SELECT id, telegram_id, telegram_username, telegram_first_name, telegram_last_name, language_code,
			alias, remna_uuid, remna_username, subscription_url, status, remna_status,
			disabled_at, delete_after, deleted_at, last_seen_at, created_at, updated_at
		FROM users
		WHERE id = $1
		FOR UPDATE
	`, id)
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

func (r *UserRepository) MarkRemnaDisabled(ctx context.Context, userID int64, disabledAt, deleteAfter time.Time) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE users
		SET remna_status = $2,
			disabled_at = $3,
			delete_after = $4,
			updated_at = now()
		WHERE id = $1
	`, userID, domain.RemnaStatusDisabled, disabledAt, deleteAfter)
	if err != nil {
		return fmt.Errorf("mark remna disabled: %w", err)
	}
	return nil
}

func (r *UserRepository) MarkRemnaDeleted(ctx context.Context, userID int64, deletedAt time.Time) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE users
		SET remna_status = $2,
			deleted_at = $3,
			updated_at = now()
		WHERE id = $1
	`, userID, domain.RemnaStatusDeleted, deletedAt)
	if err != nil {
		return fmt.Errorf("mark remna deleted: %w", err)
	}
	return nil
}

func (r *UserRepository) FindDisabledReadyForDelete(ctx context.Context, now time.Time, limit int) ([]domain.User, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	rows, err := r.tx.Querier(ctx).Query(ctx, `
		SELECT id, telegram_id, telegram_username, telegram_first_name, telegram_last_name, language_code,
			alias, remna_uuid, remna_username, subscription_url, status, remna_status,
			disabled_at, delete_after, deleted_at, last_seen_at, created_at, updated_at
		FROM users
		WHERE remna_status = 'disabled'
			AND delete_after IS NOT NULL
			AND delete_after <= $1
		ORDER BY delete_after ASC
		LIMIT $2
	`, now, limit)
	if err != nil {
		return nil, fmt.Errorf("find disabled ready for delete: %w", err)
	}
	defer rows.Close()

	items := make([]domain.User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate disabled users: %w", err)
	}
	return items, nil
}

func (r *UserRepository) getOne(ctx context.Context, query string, args ...any) (*domain.User, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	u, err := scanUser(r.tx.Querier(ctx).QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(row userScanner) (*domain.User, error) {
	var u domain.User
	if err := row.Scan(
		&u.ID, &u.TelegramID, &u.TelegramUsername, &u.TelegramFirstName, &u.TelegramLastName, &u.LanguageCode,
		&u.Alias, &u.RemnaUUID, &u.RemnaUsername, &u.SubscriptionURL, &u.Status, &u.RemnaStatus,
		&u.DisabledAt, &u.DeleteAfter, &u.DeletedAt, &u.LastSeenAt, &u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}

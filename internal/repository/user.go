package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository/tx"
)

type UserRepository struct {
	tx *tx.Manager
}

func NewUserRepository(txManager *tx.Manager) *UserRepository {
	return &UserRepository{tx: txManager}
}

func (r *UserRepository) CreateOrUpdateTelegramUser(ctx context.Context, input domain.TelegramUserInput) (*domain.User, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	query := `
		INSERT INTO users (
			telegram_id,
			telegram_username,
			telegram_first_name,
			telegram_last_name,
			language_code,
			last_seen_at
		)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (telegram_id) DO UPDATE
		SET
			telegram_username = EXCLUDED.telegram_username,
			telegram_first_name = EXCLUDED.telegram_first_name,
			telegram_last_name = EXCLUDED.telegram_last_name,
			language_code = EXCLUDED.language_code,
			last_seen_at = now(),
			updated_at = now()
		RETURNING
			id,
			telegram_id,
			telegram_username,
			telegram_first_name,
			telegram_last_name,
			language_code,
			alias,
			remna_uuid::text,
			remna_username,
			subscription_url,
			status,
			remna_status,
			disabled_at,
			delete_after,
			deleted_at,
			last_seen_at,
			created_at,
			updated_at
	`

	user, err := scanUser(r.tx.Querier(ctx).QueryRow(
		ctx,
		query,
		input.TelegramID,
		input.TelegramUsername,
		input.TelegramFirstName,
		input.TelegramLastName,
		input.LanguageCode,
	))
	if err != nil {
		return nil, fmt.Errorf("create or update telegram user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) List(ctx context.Context, input domain.UserListInput) ([]domain.User, int64, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	limit := normalizeLimit(input.Limit)
	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	where := make([]string, 0)
	args := make([]any, 0)

	if input.Status != "" {
		args = append(args, input.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}

	queryText := strings.TrimSpace(input.Query)
	if queryText != "" {
		args = append(args, "%"+queryText+"%")
		idx := len(args)
		where = append(where, fmt.Sprintf("(telegram_username ILIKE $%d OR telegram_first_name ILIKE $%d OR telegram_last_name ILIKE $%d OR alias ILIKE $%d OR telegram_id::text ILIKE $%d)", idx, idx, idx, idx, idx))
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	var total int64
	if err := r.tx.Querier(ctx).QueryRow(ctx, "SELECT count(*) FROM users"+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	args = append(args, limit, offset)
	query := baseUserSelect() + whereSQL + fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.tx.Querier(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	items := make([]domain.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate users: %w", err)
	}

	return items, total, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return r.getOne(ctx, baseUserSelect()+" WHERE id = $1", id)
}

func (r *UserRepository) GetByIDForUpdate(ctx context.Context, id int64) (*domain.User, error) {
	return r.getOne(ctx, baseUserSelect()+" WHERE id = $1 FOR UPDATE", id)
}

func (r *UserRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	return r.getOne(ctx, baseUserSelect()+" WHERE telegram_id = $1", telegramID)
}

func (r *UserRepository) Update(ctx context.Context, id int64, input domain.UpdateUserInput) (*domain.User, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	query := `
		UPDATE users
		SET
			telegram_username = COALESCE($2, telegram_username),
			telegram_first_name = COALESCE($3, telegram_first_name),
			telegram_last_name = COALESCE($4, telegram_last_name),
			language_code = COALESCE($5, language_code),
			alias = COALESCE($6, alias),
			status = COALESCE($7, status),
			updated_at = now()
		WHERE id = $1
		RETURNING
			id,
			telegram_id,
			telegram_username,
			telegram_first_name,
			telegram_last_name,
			language_code,
			alias,
			remna_uuid::text,
			remna_username,
			subscription_url,
			status,
			remna_status,
			disabled_at,
			delete_after,
			deleted_at,
			last_seen_at,
			created_at,
			updated_at
	`

	user, err := scanUser(r.tx.Querier(ctx).QueryRow(
		ctx,
		query,
		id,
		input.TelegramUsername,
		input.TelegramFirstName,
		input.TelegramLastName,
		input.LanguageCode,
		input.Alias,
		input.Status,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) SetStatus(ctx context.Context, id int64, status domain.UserStatus) (*domain.User, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	user, err := scanUser(r.tx.Querier(ctx).QueryRow(
		ctx,
		baseUserUpdateReturning()+`
			SET status = $2,
			    updated_at = now()
			WHERE id = $1
		`,
		id,
		status,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("set user status: %w", err)
	}

	return user, nil
}

func (r *UserRepository) MarkDeleted(ctx context.Context, id int64, deletedAt time.Time) (*domain.User, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	user, err := scanUser(r.tx.Querier(ctx).QueryRow(
		ctx,
		baseUserUpdateReturning()+`
			SET status = 'deleted',
			    deleted_at = $2,
			    updated_at = now()
			WHERE id = $1
		`,
		id,
		deletedAt,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("mark user deleted: %w", err)
	}

	return user, nil
}

func (r *UserRepository) SetRemnaData(ctx context.Context, userID int64, data domain.RemnaUserData) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(
		ctx,
		`
			UPDATE users
			SET
				remna_uuid = $2,
				remna_username = $3,
				subscription_url = $4,
				remna_status = $5,
				disabled_at = NULL,
				delete_after = NULL,
				deleted_at = NULL,
				updated_at = now()
			WHERE id = $1
		`,
		userID,
		data.UUID,
		data.Username,
		data.SubscriptionURL,
		data.Status,
	)
	if err != nil {
		return fmt.Errorf("set remna data: %w", err)
	}

	return nil
}

func (r *UserRepository) MarkRemnaDisabled(ctx context.Context, userID int64, disabledAt time.Time, deleteAfter time.Time) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(
		ctx,
		`
			UPDATE users
			SET
				remna_status = 'disabled',
				disabled_at = $2,
				delete_after = $3,
				updated_at = now()
			WHERE id = $1
		`,
		userID,
		disabledAt,
		deleteAfter,
	)
	if err != nil {
		return fmt.Errorf("mark remna disabled: %w", err)
	}

	return nil
}

func (r *UserRepository) MarkRemnaDeleted(ctx context.Context, userID int64, deletedAt time.Time) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	_, err := r.tx.Querier(ctx).Exec(
		ctx,
		`
			UPDATE users
			SET
				remna_status = 'deleted',
				deleted_at = $2,
				updated_at = now()
			WHERE id = $1
		`,
		userID,
		deletedAt,
	)
	if err != nil {
		return fmt.Errorf("mark remna deleted: %w", err)
	}

	return nil
}

func (r *UserRepository) FindDisabledReadyForDelete(ctx context.Context, now time.Time, limit int) ([]domain.User, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	rows, err := r.tx.Querier(ctx).Query(
		ctx,
		baseUserSelect()+`
			WHERE remna_status = 'disabled'
			  AND delete_after IS NOT NULL
			  AND delete_after <= $1
			ORDER BY delete_after ASC
			LIMIT $2
		`,
		now,
		normalizeLimit(limit),
	)
	if err != nil {
		return nil, fmt.Errorf("find disabled ready for delete: %w", err)
	}
	defer rows.Close()

	items := make([]domain.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate disabled users: %w", err)
	}

	return items, nil
}

func (r *UserRepository) getOne(ctx context.Context, query string, args ...any) (*domain.User, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	user, err := scanUser(r.tx.Querier(ctx).QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 200 {
		return 200
	}
	return limit
}

func baseUserSelect() string {
	return `
		SELECT
			id,
			telegram_id,
			telegram_username,
			telegram_first_name,
			telegram_last_name,
			language_code,
			alias,
			remna_uuid::text,
			remna_username,
			subscription_url,
			status,
			remna_status,
			disabled_at,
			delete_after,
			deleted_at,
			last_seen_at,
			created_at,
			updated_at
		FROM users
	`
}

func baseUserUpdateReturning() string {
	return `
		UPDATE users
	`
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(row userScanner) (*domain.User, error) {
	var user domain.User

	err := row.Scan(
		&user.ID,
		&user.TelegramID,
		&user.TelegramUsername,
		&user.TelegramFirstName,
		&user.TelegramLastName,
		&user.LanguageCode,
		&user.Alias,
		&user.RemnaUUID,
		&user.RemnaUsername,
		&user.SubscriptionURL,
		&user.Status,
		&user.RemnaStatus,
		&user.DisabledAt,
		&user.DeleteAfter,
		&user.DeletedAt,
		&user.LastSeenAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}

	return &user, nil
}

package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository/tx"
)

type TariffRepository struct {
	tx *tx.Manager
}

func NewTariffRepository(txManager *tx.Manager) *TariffRepository {
	return &TariffRepository{tx: txManager}
}

func (r *TariffRepository) ListActive(ctx context.Context) ([]domain.Tariff, error) {
	return r.list(ctx, true)
}

func (r *TariffRepository) ListAll(ctx context.Context) ([]domain.Tariff, error) {
	return r.list(ctx, false)
}

func (r *TariffRepository) list(ctx context.Context, onlyActive bool) ([]domain.Tariff, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	where := "WHERE code NOT LIKE '__deleted_%'"
	if onlyActive {
		where = "WHERE is_active = true AND code NOT LIKE '__deleted_%'"
	}

	rows, err := r.tx.Querier(ctx).Query(
		ctx,
		baseTariffSelect()+" "+where+" ORDER BY sort_order ASC, duration_days ASC, id ASC",
	)
	if err != nil {
		return nil, fmt.Errorf("list tariffs: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Tariff, 0)

	for rows.Next() {
		tariff, err := scanTariff(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *tariff)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tariffs: %w", err)
	}

	return items, nil
}

func (r *TariffRepository) ListActiveWithPrices(ctx context.Context) ([]domain.TariffWithPrices, error) {
	return r.listWithPrices(ctx, true)
}

func (r *TariffRepository) ListAllWithPrices(ctx context.Context) ([]domain.TariffWithPrices, error) {
	return r.listWithPrices(ctx, false)
}

func (r *TariffRepository) listWithPrices(ctx context.Context, onlyActive bool) ([]domain.TariffWithPrices, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	where := "WHERE t.code NOT LIKE '__deleted_%'"
	if onlyActive {
		where = "WHERE t.is_active = true AND t.code NOT LIKE '__deleted_%'"
	}

	rows, err := r.tx.Querier(ctx).Query(ctx, `
		SELECT
			t.id, t.code, t.title, t.description, t.duration_days, t.period_days,
			t.traffic_limit_bytes, t.price_rub, t.is_active, t.sort_order, t.created_at, t.updated_at,
			p.id, p.tariff_id, p.provider, p.payment_method, p.currency, p.amount_minor,
			p.stars_amount::bigint, p.accepted_assets, p.is_active, p.sort_order, p.created_at, p.updated_at
		FROM tariffs t
		LEFT JOIN tariff_prices p ON p.tariff_id = t.id AND p.is_active = true
		`+where+`
		ORDER BY t.sort_order ASC, t.duration_days ASC, p.sort_order ASC, p.id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list tariffs with prices: %w", err)
	}
	defer rows.Close()

	byID := make(map[int64]*domain.TariffWithPrices)
	order := make([]int64, 0)

	for rows.Next() {
		var tariff domain.Tariff
		var price domain.TariffPrice
		var priceID *int64

		err := rows.Scan(
			&tariff.ID,
			&tariff.Code,
			&tariff.Title,
			&tariff.Description,
			&tariff.DurationDays,
			&tariff.PeriodDays,
			&tariff.TrafficLimitBytes,
			&tariff.PriceRub,
			&tariff.IsActive,
			&tariff.SortOrder,
			&tariff.CreatedAt,
			&tariff.UpdatedAt,
			&priceID,
			&price.TariffID,
			&price.Provider,
			&price.PaymentMethod,
			&price.Currency,
			&price.AmountMinor,
			&price.StarsAmount,
			&price.AcceptedAssets,
			&price.IsActive,
			&price.SortOrder,
			&price.CreatedAt,
			&price.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan tariff with price: %w", err)
		}

		item, ok := byID[tariff.ID]
		if !ok {
			item = &domain.TariffWithPrices{
				Tariff: tariff,
				Prices: make([]domain.TariffPrice, 0),
			}
			byID[tariff.ID] = item
			order = append(order, tariff.ID)
		}

		if priceID != nil {
			price.ID = *priceID
			item.Prices = append(item.Prices, price)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tariffs with prices: %w", err)
	}

	out := make([]domain.TariffWithPrices, 0, len(order))
	for _, id := range order {
		out = append(out, *byID[id])
	}

	return out, nil
}

func (r *TariffRepository) GetByID(ctx context.Context, id int64) (*domain.Tariff, error) {
	return r.getOne(ctx, baseTariffSelect()+" WHERE id = $1 AND is_active = true AND code NOT LIKE '__deleted_%'", id)
}

func (r *TariffRepository) GetAnyByID(ctx context.Context, id int64) (*domain.Tariff, error) {
	return r.getOne(ctx, baseTariffSelect()+" WHERE id = $1 AND code NOT LIKE '__deleted_%'", id)
}

func (r *TariffRepository) Create(ctx context.Context, input domain.CreateTariffInput) (*domain.Tariff, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	tariff, err := scanTariff(r.tx.Querier(ctx).QueryRow(ctx, `
		INSERT INTO tariffs (
			code, title, description, duration_days, period_days,
			traffic_limit_bytes, price_rub, is_active, sort_order
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, code, title, description, duration_days, period_days,
		          traffic_limit_bytes, price_rub, is_active, sort_order, created_at, updated_at
	`,
		input.Code,
		input.Title,
		input.Description,
		input.DurationDays,
		input.PeriodDays,
		domain.TrafficGBToBytes(input.TrafficLimitGB),
		input.PriceRub,
		isActive,
		input.SortOrder,
	))
	if err != nil {
		return nil, fmt.Errorf("create tariff: %w", err)
	}

	return tariff, nil
}

func (r *TariffRepository) Update(ctx context.Context, id int64, input domain.UpdateTariffInput) (*domain.Tariff, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	var trafficLimitBytes *int64
	if input.TrafficLimitGB != nil {
		value := domain.TrafficGBToBytes(*input.TrafficLimitGB)
		trafficLimitBytes = &value
	}

	tariff, err := scanTariff(r.tx.Querier(ctx).QueryRow(ctx, `
		UPDATE tariffs
		SET code = COALESCE($2, code),
		    title = COALESCE($3, title),
		    description = COALESCE($4, description),
		    duration_days = COALESCE($5, duration_days),
		    period_days = COALESCE($6, period_days),
		    traffic_limit_bytes = COALESCE($7, traffic_limit_bytes),
		    price_rub = COALESCE($8, price_rub),
		    is_active = COALESCE($9, is_active),
		    sort_order = COALESCE($10, sort_order),
		    updated_at = now()
		WHERE id = $1
		  AND code NOT LIKE '__deleted_%'
		RETURNING id, code, title, description, duration_days, period_days,
		          traffic_limit_bytes, price_rub, is_active, sort_order, created_at, updated_at
	`,
		id,
		input.Code,
		input.Title,
		input.Description,
		input.DurationDays,
		input.PeriodDays,
		trafficLimitBytes,
		input.PriceRub,
		input.IsActive,
		input.SortOrder,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("update tariff: %w", err)
	}

	return tariff, nil
}

func (r *TariffRepository) SetActive(ctx context.Context, id int64, isActive bool) (*domain.Tariff, error) {
	return r.Update(ctx, id, domain.UpdateTariffInput{IsActive: &isActive})
}

func (r *TariffRepository) Delete(ctx context.Context, id int64) (*domain.Tariff, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	tariff, err := scanTariff(r.tx.Querier(ctx).QueryRow(ctx, `
		UPDATE tariffs
		SET is_active = false,
		    code = '__deleted_' || id::text || '_' || code,
		    title = title || ' (удален)',
		    sort_order = 2147483647,
		    updated_at = now()
		WHERE id = $1
		  AND code NOT LIKE '__deleted_%'
		RETURNING id, code, title, description, duration_days, period_days,
		          traffic_limit_bytes, price_rub, is_active, sort_order, created_at, updated_at
	`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("delete tariff: %w", err)
	}

	return tariff, nil
}

func (r *TariffRepository) getOne(ctx context.Context, query string, args ...any) (*domain.Tariff, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	tariff, err := scanTariff(r.tx.Querier(ctx).QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("get tariff: %w", err)
	}

	return tariff, nil
}

func baseTariffSelect() string {
	return `
		SELECT id, code, title, description, duration_days, period_days,
		       traffic_limit_bytes, price_rub, is_active, sort_order, created_at, updated_at
		FROM tariffs
	`
}

type tariffScanner interface {
	Scan(dest ...any) error
}

func scanTariff(row tariffScanner) (*domain.Tariff, error) {
	var tariff domain.Tariff

	if err := row.Scan(
		&tariff.ID,
		&tariff.Code,
		&tariff.Title,
		&tariff.Description,
		&tariff.DurationDays,
		&tariff.PeriodDays,
		&tariff.TrafficLimitBytes,
		&tariff.PriceRub,
		&tariff.IsActive,
		&tariff.SortOrder,
		&tariff.CreatedAt,
		&tariff.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan tariff: %w", err)
	}

	return &tariff, nil
}

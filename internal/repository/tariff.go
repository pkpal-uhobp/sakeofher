package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository/tx"
)

type TariffRepository struct{ tx *tx.Manager }

func NewTariffRepository(txManager *tx.Manager) *TariffRepository {
	return &TariffRepository{tx: txManager}
}

func (r *TariffRepository) ListActive(ctx context.Context) ([]domain.Tariff, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	rows, err := r.tx.Querier(ctx).Query(ctx, `
		SELECT id, code, title, description, duration_days, period_days, traffic_limit_bytes, is_active, sort_order, created_at, updated_at
		FROM tariffs
		WHERE is_active = true
		ORDER BY sort_order ASC, duration_days ASC, id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list active tariffs: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Tariff, 0)
	for rows.Next() {
		var t domain.Tariff
		if err := rows.Scan(&t.ID, &t.Code, &t.Title, &t.Description, &t.DurationDays, &t.PeriodDays, &t.TrafficLimitBytes, &t.IsActive, &t.SortOrder, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan tariff: %w", err)
		}
		items = append(items, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tariffs: %w", err)
	}
	return items, nil
}

func (r *TariffRepository) ListActiveWithPrices(ctx context.Context) ([]domain.TariffWithPrices, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	rows, err := r.tx.Querier(ctx).Query(ctx, `
		SELECT
			t.id, t.code, t.title, t.description, t.duration_days, t.period_days, t.traffic_limit_bytes, t.is_active, t.sort_order, t.created_at, t.updated_at,
			p.id, p.tariff_id, p.provider, p.payment_method, p.currency, p.amount_minor, p.stars_amount::bigint, p.accepted_assets, p.is_active, p.sort_order, p.created_at, p.updated_at
		FROM tariffs t
		LEFT JOIN tariff_prices p ON p.tariff_id = t.id AND p.is_active = true
		WHERE t.is_active = true
		ORDER BY t.sort_order ASC, t.duration_days ASC, p.sort_order ASC, p.id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list active tariffs with prices: %w", err)
	}
	defer rows.Close()

	byID := make(map[int64]*domain.TariffWithPrices)
	order := make([]int64, 0)

	for rows.Next() {
		var t domain.Tariff
		var p domain.TariffPrice
		var priceID *int64

		err := rows.Scan(
			&t.ID, &t.Code, &t.Title, &t.Description, &t.DurationDays, &t.PeriodDays, &t.TrafficLimitBytes, &t.IsActive, &t.SortOrder, &t.CreatedAt, &t.UpdatedAt,
			&priceID, &p.TariffID, &p.Provider, &p.PaymentMethod, &p.Currency, &p.AmountMinor, &p.StarsAmount, &p.AcceptedAssets, &p.IsActive, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan tariff with price: %w", err)
		}

		item, ok := byID[t.ID]
		if !ok {
			item = &domain.TariffWithPrices{Tariff: t, Prices: make([]domain.TariffPrice, 0)}
			byID[t.ID] = item
			order = append(order, t.ID)
		}
		if priceID != nil {
			p.ID = *priceID
			item.Prices = append(item.Prices, p)
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
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	var t domain.Tariff
	err := r.tx.Querier(ctx).QueryRow(ctx, `
		SELECT id, code, title, description, duration_days, period_days, traffic_limit_bytes, is_active, sort_order, created_at, updated_at
		FROM tariffs
		WHERE id = $1 AND is_active = true
	`, id).Scan(&t.ID, &t.Code, &t.Title, &t.Description, &t.DurationDays, &t.PeriodDays, &t.TrafficLimitBytes, &t.IsActive, &t.SortOrder, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get tariff by id: %w", err)
	}
	return &t, nil
}

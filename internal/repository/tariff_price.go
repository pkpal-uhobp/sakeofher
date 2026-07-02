package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository/tx"
)

type TariffPriceRepository struct{ tx *tx.Manager }

func NewTariffPriceRepository(txManager *tx.Manager) *TariffPriceRepository {
	return &TariffPriceRepository{tx: txManager}
}

func (r *TariffPriceRepository) GetByID(ctx context.Context, id int64) (*domain.TariffPrice, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	q := `
		SELECT id, tariff_id, provider, payment_method, currency, amount_minor, stars_amount::bigint, accepted_assets, is_active, sort_order, created_at, updated_at
		FROM tariff_prices
		WHERE id = $1
	`
	var p domain.TariffPrice
	err := r.tx.Querier(ctx).QueryRow(ctx, q, id).Scan(
		&p.ID, &p.TariffID, &p.Provider, &p.PaymentMethod, &p.Currency, &p.AmountMinor, &p.StarsAmount,
		&p.AcceptedAssets, &p.IsActive, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get tariff price by id: %w", err)
	}
	return &p, nil
}

func (r *TariffPriceRepository) GetWithTariffByID(ctx context.Context, id int64) (*domain.TariffPriceWithTariff, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	q := `
		SELECT
			p.id, p.tariff_id, p.provider, p.payment_method, p.currency, p.amount_minor, p.stars_amount::bigint, p.accepted_assets, p.is_active, p.sort_order, p.created_at, p.updated_at,
			t.id, t.code, t.title, t.description, t.duration_days, t.period_days, t.traffic_limit_bytes, t.is_active, t.sort_order, t.created_at, t.updated_at
		FROM tariff_prices p
		JOIN tariffs t ON t.id = p.tariff_id
		WHERE p.id = $1
	`
	var out domain.TariffPriceWithTariff
	err := r.tx.Querier(ctx).QueryRow(ctx, q, id).Scan(
		&out.Price.ID, &out.Price.TariffID, &out.Price.Provider, &out.Price.PaymentMethod, &out.Price.Currency, &out.Price.AmountMinor, &out.Price.StarsAmount,
		&out.Price.AcceptedAssets, &out.Price.IsActive, &out.Price.SortOrder, &out.Price.CreatedAt, &out.Price.UpdatedAt,
		&out.Tariff.ID, &out.Tariff.Code, &out.Tariff.Title, &out.Tariff.Description, &out.Tariff.DurationDays, &out.Tariff.PeriodDays,
		&out.Tariff.TrafficLimitBytes, &out.Tariff.IsActive, &out.Tariff.SortOrder, &out.Tariff.CreatedAt, &out.Tariff.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get tariff price with tariff by id: %w", err)
	}
	return &out, nil
}

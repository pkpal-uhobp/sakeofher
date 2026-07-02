package repository

import (
	"context"
	"fmt"

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
        SELECT id, tariff_id, provider, payment_method, currency, amount_minor, stars_amount, accepted_assets, is_active, created_at, updated_at
        FROM tariff_prices
        WHERE id = $1
    `
	var p domain.TariffPrice
	err := r.tx.Querier(ctx).QueryRow(ctx, q, id).Scan(
		&p.ID, &p.TariffID, &p.Provider, &p.PaymentMethod, &p.Currency, &p.AmountMinor, &p.StarsAmount,
		&p.AcceptedAssets, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get tariff price by id: %w", err)
	}
	return &p, nil
}

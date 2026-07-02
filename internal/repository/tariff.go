package repository

import (
	"context"
	"fmt"

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
        SELECT id, code, name, description, duration_days, traffic_limit_bytes, is_active, created_at, updated_at
        FROM tariffs
        WHERE is_active = true
        ORDER BY duration_days ASC
    `)
	if err != nil {
		return nil, fmt.Errorf("list active tariffs: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Tariff, 0)
	for rows.Next() {
		var t domain.Tariff
		if err := rows.Scan(&t.ID, &t.Code, &t.Name, &t.Description, &t.DurationDays, &t.TrafficLimitBytes, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan tariff: %w", err)
		}
		items = append(items, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tariffs: %w", err)
	}
	return items, nil
}

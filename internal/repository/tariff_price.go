package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository/tx"
)

type TariffPriceRepository struct {
	tx *tx.Manager
}

func NewTariffPriceRepository(txManager *tx.Manager) *TariffPriceRepository {
	return &TariffPriceRepository{tx: txManager}
}

func (r *TariffPriceRepository) GetByID(ctx context.Context, id int64) (*domain.TariffPrice, error) {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	q := `
		SELECT id, tariff_id, provider, payment_method, currency, amount_minor,
		       stars_amount::bigint, accepted_assets, is_active, sort_order, created_at, updated_at
		FROM tariff_prices
		WHERE id = $1
	`

	var p domain.TariffPrice
	err := r.tx.Querier(ctx).QueryRow(ctx, q, id).Scan(
		&p.ID,
		&p.TariffID,
		&p.Provider,
		&p.PaymentMethod,
		&p.Currency,
		&p.AmountMinor,
		&p.StarsAmount,
		&p.AcceptedAssets,
		&p.IsActive,
		&p.SortOrder,
		&p.CreatedAt,
		&p.UpdatedAt,
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
		    p.id, p.tariff_id, p.provider, p.payment_method, p.currency,
		    p.amount_minor, p.stars_amount::bigint, p.accepted_assets,
		    p.is_active, p.sort_order, p.created_at, p.updated_at,

		    t.id, t.code, t.title, t.description, t.duration_days, t.period_days,
		    t.traffic_limit_bytes, t.price_rub, t.is_active, t.sort_order,
		    t.created_at, t.updated_at
		FROM tariff_prices p
		JOIN tariffs t ON t.id = p.tariff_id
		WHERE p.id = $1
	`

	var out domain.TariffPriceWithTariff
	err := r.tx.Querier(ctx).QueryRow(ctx, q, id).Scan(
		&out.Price.ID,
		&out.Price.TariffID,
		&out.Price.Provider,
		&out.Price.PaymentMethod,
		&out.Price.Currency,
		&out.Price.AmountMinor,
		&out.Price.StarsAmount,
		&out.Price.AcceptedAssets,
		&out.Price.IsActive,
		&out.Price.SortOrder,
		&out.Price.CreatedAt,
		&out.Price.UpdatedAt,

		&out.Tariff.ID,
		&out.Tariff.Code,
		&out.Tariff.Title,
		&out.Tariff.Description,
		&out.Tariff.DurationDays,
		&out.Tariff.PeriodDays,
		&out.Tariff.TrafficLimitBytes,
		&out.Tariff.PriceRub,
		&out.Tariff.IsActive,
		&out.Tariff.SortOrder,
		&out.Tariff.CreatedAt,
		&out.Tariff.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("get tariff price with tariff by id: %w", err)
	}

	return &out, nil
}

func (r *TariffPriceRepository) ReplaceManagedForTariff(
	ctx context.Context,
	tariffID int64,
	settings domain.TariffPaymentSettingsInput,
) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	assets := normalizeAcceptedAssets(settings.CryptoBotCrypto.AcceptedAssets)
	if len(assets) == 0 {
		assets = []string{"USDT", "TON"}
	}

	if err := r.upsertTelegramStars(ctx, tariffID, settings.TelegramStars); err != nil {
		return err
	}

	if err := r.upsertCryptoBotCrypto(ctx, tariffID, settings.CryptoBotCrypto, assets); err != nil {
		return err
	}

	if err := r.upsertTributeRub(ctx, tariffID, settings.TributeRub); err != nil {
		return err
	}

	// Historical cleanup: this project used to display "CryptoBot — рубли".
	// CryptoBot remains only crypto; separate RUB payment is Tribute.
	_, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE tariff_prices
		SET is_active = false,
		    updated_at = now()
		WHERE tariff_id = $1
		  AND provider = 'crypto_bot'
		  AND payment_method = 'rub'
	`, tariffID)
	if err != nil {
		return fmt.Errorf("disable legacy cryptobot rub price: %w", err)
	}

	return nil
}

func (r *TariffPriceRepository) upsertTelegramStars(
	ctx context.Context,
	tariffID int64,
	settings domain.TariffTelegramStarsSettings,
) error {
	var starsAmount *int64
	if settings.Enabled {
		if settings.StarsAmount <= 0 {
			return domain.ErrInvalidInput
		}

		starsAmount = &settings.StarsAmount
	}

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		INSERT INTO tariff_prices (
		    tariff_id,
		    provider,
		    payment_method,
		    currency,
		    amount_minor,
		    stars_amount,
		    accepted_assets,
		    is_active,
		    sort_order
		)
		VALUES ($1, 'telegram_stars', 'stars', 'XTR', NULL, $2, '{}'::TEXT[], $3, 10)
		ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE
		SET currency = EXCLUDED.currency,
		    amount_minor = EXCLUDED.amount_minor,
		    stars_amount = EXCLUDED.stars_amount,
		    accepted_assets = EXCLUDED.accepted_assets,
		    is_active = EXCLUDED.is_active,
		    sort_order = EXCLUDED.sort_order,
		    updated_at = now()
	`, tariffID, starsAmount, settings.Enabled)
	if err != nil {
		return fmt.Errorf("replace telegram stars tariff price: %w", err)
	}

	return nil
}

func (r *TariffPriceRepository) upsertCryptoBotCrypto(
	ctx context.Context,
	tariffID int64,
	settings domain.TariffCryptoBotCryptoSettings,
	assets []string,
) error {
	var amountMinor *int64
	if settings.Enabled {
		if settings.PriceRub <= 0 {
			return domain.ErrInvalidInput
		}

		value := settings.PriceRub * 100
		amountMinor = &value
	}

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		INSERT INTO tariff_prices (
		    tariff_id,
		    provider,
		    payment_method,
		    currency,
		    amount_minor,
		    stars_amount,
		    accepted_assets,
		    is_active,
		    sort_order
		)
		VALUES ($1, 'crypto_bot', 'crypto', 'RUB', $2, NULL, $3, $4, 20)
		ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE
		SET currency = EXCLUDED.currency,
		    amount_minor = EXCLUDED.amount_minor,
		    stars_amount = EXCLUDED.stars_amount,
		    accepted_assets = EXCLUDED.accepted_assets,
		    is_active = EXCLUDED.is_active,
		    sort_order = EXCLUDED.sort_order,
		    updated_at = now()
	`, tariffID, amountMinor, assets, settings.Enabled)
	if err != nil {
		return fmt.Errorf("replace cryptobot crypto tariff price: %w", err)
	}

	return nil
}

func (r *TariffPriceRepository) upsertTributeRub(
	ctx context.Context,
	tariffID int64,
	settings domain.TariffTributeRubSettings,
) error {
	var amountMinor *int64
	if settings.Enabled {
		if settings.PriceRub <= 0 {
			return domain.ErrInvalidInput
		}

		value := settings.PriceRub * 100
		amountMinor = &value
	}

	_, err := r.tx.Querier(ctx).Exec(ctx, `
		INSERT INTO tariff_prices (
		    tariff_id,
		    provider,
		    payment_method,
		    currency,
		    amount_minor,
		    stars_amount,
		    accepted_assets,
		    is_active,
		    sort_order
		)
		VALUES ($1, 'tribute', 'rub', 'RUB', $2, NULL, '{}'::TEXT[], $3, 30)
		ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE
		SET currency = EXCLUDED.currency,
		    amount_minor = EXCLUDED.amount_minor,
		    stars_amount = EXCLUDED.stars_amount,
		    accepted_assets = EXCLUDED.accepted_assets,
		    is_active = EXCLUDED.is_active,
		    sort_order = EXCLUDED.sort_order,
		    updated_at = now()
	`, tariffID, amountMinor, settings.Enabled)
	if err != nil {
		return fmt.Errorf("replace tribute rub tariff price: %w", err)
	}

	return nil
}

func normalizeAcceptedAssets(items []string) []string {
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{})

	for _, item := range items {
		item = strings.ToUpper(strings.TrimSpace(item))
		if item == "" {
			continue
		}

		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		out = append(out, item)
	}

	return out
}

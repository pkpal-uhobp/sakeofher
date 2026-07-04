package repository

import (
	"context"
	"fmt"
	"strings"

	"sakeofher/internal/domain"
)

func (r *TariffPriceRepository) ReplaceManagedForTariff(ctx context.Context, tariffID int64, settings domain.TariffPaymentSettingsInput) error {
	ctx, cancel := r.tx.WithTimeout(ctx)
	defer cancel()

	if _, err := r.tx.Querier(ctx).Exec(ctx, `
		UPDATE tariff_prices
		SET is_active = false,
		    updated_at = now()
		WHERE tariff_id = $1
		  AND (
		    (provider = 'telegram' AND payment_method = 'stars')
		    OR provider = 'cryptobot'
		  )
	`, tariffID); err != nil {
		return fmt.Errorf("disable old tariff payment settings: %w", err)
	}

	sortOrder := 10

	if settings.TelegramStars.Enabled {
		if settings.TelegramStars.StarsAmount <= 0 {
			return domain.ErrInvalidInput
		}

		starsAmount := settings.TelegramStars.StarsAmount
		if err := r.insertManagedPrice(ctx, tariffID, "telegram", "stars", "XTR", nil, &starsAmount, []string{}, sortOrder); err != nil {
			return err
		}
		sortOrder += 10
	}

	if settings.CryptoBotCrypto.Enabled {
		if settings.CryptoBotCrypto.PriceRub <= 0 {
			return domain.ErrInvalidInput
		}

		assets := normalizeAssets(settings.CryptoBotCrypto.AcceptedAssets)
		if len(assets) == 0 {
			assets = []string{"USDT", "TON", "BTC", "ETH", "LTC", "BNB", "TRX", "USDC"}
		}

		amountMinor := settings.CryptoBotCrypto.PriceRub * 100
		if err := r.insertManagedPrice(ctx, tariffID, "cryptobot", "crypto", "RUB", &amountMinor, nil, assets, sortOrder); err != nil {
			return err
		}
		sortOrder += 10
	}

	if settings.CryptoBotRub.Enabled {
		if settings.CryptoBotRub.PriceRub <= 0 {
			return domain.ErrInvalidInput
		}

		amountMinor := settings.CryptoBotRub.PriceRub * 100
		if err := r.insertManagedPrice(ctx, tariffID, "cryptobot", "rub", "RUB", &amountMinor, nil, []string{}, sortOrder); err != nil {
			return err
		}
	}

	return nil
}

func (r *TariffPriceRepository) insertManagedPrice(
	ctx context.Context,
	tariffID int64,
	provider string,
	method string,
	currency string,
	amountMinor *int64,
	starsAmount *int64,
	acceptedAssets []string,
	sortOrder int,
) error {
	if _, err := r.tx.Querier(ctx).Exec(ctx, `
		INSERT INTO tariff_prices (
			tariff_id, provider, payment_method, currency, amount_minor,
			stars_amount, accepted_assets, is_active, sort_order
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,true,$8)
	`, tariffID, provider, method, currency, amountMinor, starsAmount, acceptedAssets, sortOrder); err != nil {
		return fmt.Errorf("insert tariff payment setting %s/%s: %w", provider, method, err)
	}

	return nil
}

func normalizeAssets(items []string) []string {
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

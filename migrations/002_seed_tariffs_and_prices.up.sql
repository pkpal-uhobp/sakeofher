-- 002_seed_tariffs_and_prices.up.sql
-- Default tariffs and prices.
-- Traffic: 300 GiB per 30-day period.
--
-- amount_minor / price_rub are stored in kopecks:
-- 6500 = 65.00 RUB

BEGIN;

INSERT INTO tariffs (
    code,
    title,
    description,
    duration_days,
    period_days,
    traffic_limit_bytes,
    price_rub,
    is_active,
    sort_order
)
VALUES
    (
        'vpn_1m_300gb',
        '1 месяц',
        'Доступ на 30 дней, 300 ГБ на период',
        30,
        30,
        322122547200,
        6500,
        true,
        10
    ),
    (
        'vpn_2m_300gb',
        '2 месяца',
        'Доступ на 60 дней, 300 ГБ на период',
        60,
        30,
        322122547200,
        14000,
        true,
        20
    ),
    (
        'vpn_3m_300gb',
        '3 месяца',
        'Доступ на 90 дней, 300 ГБ на период',
        90,
        30,
        322122547200,
        21000,
        true,
        30
    )
ON CONFLICT (code) DO UPDATE
SET title = EXCLUDED.title,
    description = EXCLUDED.description,
    duration_days = EXCLUDED.duration_days,
    period_days = EXCLUDED.period_days,
    traffic_limit_bytes = EXCLUDED.traffic_limit_bytes,
    price_rub = EXCLUDED.price_rub,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

-- Telegram Stars.

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
SELECT id, 'telegram_stars', 'stars', 'XTR', NULL, 50, '{}'::TEXT[], true, 10
FROM tariffs
WHERE code = 'vpn_1m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE
SET currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

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
SELECT id, 'telegram_stars', 'stars', 'XTR', NULL, 100, '{}'::TEXT[], true, 10
FROM tariffs
WHERE code = 'vpn_2m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE
SET currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

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
SELECT id, 'telegram_stars', 'stars', 'XTR', NULL, 150, '{}'::TEXT[], true, 10
FROM tariffs
WHERE code = 'vpn_3m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE
SET currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

-- CryptoBot crypto: RUB-denominated invoice, paid by selected crypto assets.

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
SELECT id, 'crypto_bot', 'crypto', 'RUB', 6500, NULL,
       ARRAY['USDT','TON','BTC','ETH','LTC','BNB','TRX','USDC']::TEXT[],
       true, 20
FROM tariffs
WHERE code = 'vpn_1m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE
SET currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

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
SELECT id, 'crypto_bot', 'crypto', 'RUB', 13000, NULL,
       ARRAY['USDT','TON','BTC','ETH','LTC','BNB','TRX','USDC']::TEXT[],
       true, 20
FROM tariffs
WHERE code = 'vpn_2m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE
SET currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

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
SELECT id, 'crypto_bot', 'crypto', 'RUB', 19500, NULL,
       ARRAY['USDT','TON','BTC','ETH','LTC','BNB','TRX','USDC']::TEXT[],
       true, 20
FROM tariffs
WHERE code = 'vpn_3m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE
SET currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

-- Tribute RUB.
-- 1 month Tribute is intentionally inactive because the requested Tribute prices
-- are specified only for 2 and 3 months.

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
SELECT id, 'tribute', 'rub', 'RUB', 6500, NULL, '{}'::TEXT[], false, 30
FROM tariffs
WHERE code = 'vpn_1m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE
SET currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

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
SELECT id, 'tribute', 'rub', 'RUB', 14000, NULL, '{}'::TEXT[], true, 30
FROM tariffs
WHERE code = 'vpn_2m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE
SET currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

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
SELECT id, 'tribute', 'rub', 'RUB', 21000, NULL, '{}'::TEXT[], true, 30
FROM tariffs
WHERE code = 'vpn_3m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE
SET currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

COMMIT;

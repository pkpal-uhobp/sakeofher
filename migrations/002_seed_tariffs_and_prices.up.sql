-- 002_seed_tariffs_and_prices.up.sql
-- Initial tariffs and prices.
-- Edit prices here or later through admin panel.
-- Traffic: 300 GiB per 30-day period.

BEGIN;

INSERT INTO tariffs
(code, title, description, duration_days, period_days, traffic_limit_bytes, is_active, sort_order)
VALUES
('vpn_1m_300gb', '1 месяц', 'VPN на 30 дней, 300 GB трафика', 30, 30, 322122547200, true, 10),
('vpn_3m_300gb', '3 месяца', 'VPN на 90 дней, 300 GB трафика каждые 30 дней', 90, 30, 322122547200, true, 20)
ON CONFLICT (code) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    duration_days = EXCLUDED.duration_days,
    period_days = EXCLUDED.period_days,
    traffic_limit_bytes = EXCLUDED.traffic_limit_bytes,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

-- Telegram Stars.
INSERT INTO tariff_prices
(tariff_id, provider, payment_method, currency, amount_minor, stars_amount, accepted_assets, is_active, sort_order)
SELECT id, 'telegram_stars', 'stars', 'XTR', NULL, 50, '{}'::TEXT[], true, 10
FROM tariffs WHERE code = 'vpn_1m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE SET
    currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

INSERT INTO tariff_prices
(tariff_id, provider, payment_method, currency, amount_minor, stars_amount, accepted_assets, is_active, sort_order)
SELECT id, 'telegram_stars', 'stars', 'XTR', NULL, 150, '{}'::TEXT[], true, 20
FROM tariffs WHERE code = 'vpn_3m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE SET
    currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

-- Tribute RUB. amount_minor is in kopecks.
INSERT INTO tariff_prices
(tariff_id, provider, payment_method, currency, amount_minor, stars_amount, accepted_assets, is_active, sort_order)
SELECT id, 'tribute', 'rub', 'RUB', 7000, NULL, '{}'::TEXT[], true, 30
FROM tariffs WHERE code = 'vpn_1m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE SET
    currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

INSERT INTO tariff_prices
(tariff_id, provider, payment_method, currency, amount_minor, stars_amount, accepted_assets, is_active, sort_order)
SELECT id, 'tribute', 'rub', 'RUB', 21000, NULL, '{}'::TEXT[], true, 40
FROM tariffs WHERE code = 'vpn_3m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE SET
    currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

-- CryptoBot Crypto Pay. Fiat invoice in RUB, paid by accepted assets.
INSERT INTO tariff_prices
(tariff_id, provider, payment_method, currency, amount_minor, stars_amount, accepted_assets, is_active, sort_order)
SELECT id, 'crypto_bot', 'crypto', 'RUB', 6500, NULL, ARRAY['USDT', 'TON']::TEXT[], true, 50
FROM tariffs WHERE code = 'vpn_1m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE SET
    currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

INSERT INTO tariff_prices
(tariff_id, provider, payment_method, currency, amount_minor, stars_amount, accepted_assets, is_active, sort_order)
SELECT id, 'crypto_bot', 'crypto', 'RUB', 19500, NULL, ARRAY['USDT', 'TON']::TEXT[], true, 60
FROM tariffs WHERE code = 'vpn_3m_300gb'
ON CONFLICT (tariff_id, provider, payment_method) DO UPDATE SET
    currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    stars_amount = EXCLUDED.stars_amount,
    accepted_assets = EXCLUDED.accepted_assets,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

COMMIT;

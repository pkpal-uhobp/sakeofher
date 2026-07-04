-- 002_seed_tariffs_and_prices.down.sql

BEGIN;

DELETE FROM tariff_prices
WHERE tariff_id IN (
    SELECT id
    FROM tariffs
    WHERE code IN ('vpn_1m_300gb', 'vpn_2m_300gb', 'vpn_3m_300gb')
);

DELETE FROM tariffs
WHERE code IN ('vpn_1m_300gb', 'vpn_2m_300gb', 'vpn_3m_300gb');

COMMIT;

BEGIN;

DELETE FROM tariff_prices
WHERE tariff_id IN (
    SELECT id
    FROM tariffs
    WHERE code IN ('vpn_1m_300gb', 'vpn_2m_300gb', 'vpn_3m_300gb')
)
AND provider IN ('telegram_stars', 'crypto_bot', 'tribute');

UPDATE tariffs
SET is_active = false,
    updated_at = now()
WHERE code = 'vpn_2m_300gb';

COMMIT;

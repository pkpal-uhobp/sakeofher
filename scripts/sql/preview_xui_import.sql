-- Preview x-ui import candidates against current SakeOfHer database.
-- Safe: does not modify anything.

WITH params AS (
    SELECT
        :target_tariff_id::bigint AS target_tariff_id,
        322122547200::bigint AS target_traffic_limit_bytes,
        TIMESTAMPTZ '2026-07-20 23:59:59+03' AS min_expires_at
),
old_import(telegram_id, base_username, expires_at_ms, traffic_used_bytes) AS (
    VALUES
    (1839835493, 'an_lffy', 1784364187370, 109433253),
    (1136445296, 'Dertoo1', 1785864695916, 0),
    (998952661, 'firefly_4ik', 1784364155174, 3867447377),
    (984724049, 'Firstness1', 1784883576332, 648533),
    (6075877693, 'LwvwvvL', 1785118869800, 1336391600),
    (7992413092, 'monkxze', 1783520997816, 349245750),
    (970706613, 'pkpal_uhobp', 1784364231291, 117027280355),
    (802389929, 'snebj0', 1785328602049, 0),
    (889789573, 'tabletka_49', 1784711374751, 3534752936),
    (884621667, 'TheHappyPrince', 1785861640699, 0),
    (1238134452, 'TotsamyDanya', 1784795173269, 49822696),
    (1766479687, 'trvssdav', 1784549285020, 2917703914),
    (1203698798, 'Vital4ik52', 1783627074813, 2957336),
    (991438068, 'whitebotik', 1784272698806, 603375)
),
prepared AS (
    SELECT
        telegram_id,
        base_username,
        GREATEST(to_timestamp(expires_at_ms / 1000.0), params.min_expires_at) AS target_expires_at,
        params.target_traffic_limit_bytes AS target_traffic_limit_bytes,
        traffic_used_bytes
    FROM old_import
    CROSS JOIN params
    WHERE expires_at_ms > 0
      AND telegram_id > 1
),
latest_subscription AS (
    SELECT DISTINCT ON (s.user_id)
        s.user_id,
        s.status,
        s.expires_at,
        s.traffic_limit_bytes,
        s.traffic_used_bytes
    FROM subscriptions s
    ORDER BY s.user_id, s.expires_at DESC, s.id DESC
)
SELECT
    p.base_username,
    p.telegram_id,
    CASE WHEN u.id IS NULL THEN 'will_create_user' ELSE 'user_exists' END AS user_state,
    u.id AS user_id,
    u.telegram_username AS current_username,
    u.alias AS current_alias,
    ls.status AS current_sub_status,
    ls.expires_at AS current_expires_at,
    p.target_expires_at,
    p.target_traffic_limit_bytes,
    p.traffic_used_bytes AS imported_traffic_used_bytes
FROM prepared p
LEFT JOIN users u ON u.telegram_id = p.telegram_id
LEFT JOIN latest_subscription ls ON ls.user_id = u.id
ORDER BY p.base_username;

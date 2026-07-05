-- Import x-ui users into SakeOfHer bot database.
-- Rules:
-- 1) Every imported user gets exactly 300 GiB traffic limit.
-- 2) If old subscription expires before 2026-07-20 23:59:59+03,
--    it is extended to 2026-07-20 23:59:59+03.
-- 3) Existing active subscriptions are not shortened.
-- 4) target_tariff_id is passed by scripts/import-xui-users.ps1.
-- 5) Run Preview first.

BEGIN;

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
        GREATEST(to_timestamp(expires_at_ms / 1000.0), params.min_expires_at) AS expires_at,
        params.target_traffic_limit_bytes AS traffic_limit_bytes,
        traffic_used_bytes
    FROM old_import
    CROSS JOIN params
    WHERE expires_at_ms > 0
      AND telegram_id > 1
),
upsert_users AS (
    INSERT INTO users (
        telegram_id,
        telegram_username,
        alias,
        status,
        created_at,
        updated_at
    )
    SELECT
        telegram_id,
        base_username,
        '@' || base_username,
        'active',
        now(),
        now()
    FROM prepared
    ON CONFLICT (telegram_id) DO UPDATE
    SET
        telegram_username = COALESCE(NULLIF(users.telegram_username, ''), EXCLUDED.telegram_username),
        alias = COALESCE(NULLIF(users.alias, ''), EXCLUDED.alias),
        status = 'active',
        updated_at = now()
    RETURNING id, telegram_id
),
src AS (
    SELECT
        u.id AS user_id,
        p.telegram_id,
        p.expires_at,
        p.traffic_limit_bytes,
        p.traffic_used_bytes,
        params.target_tariff_id AS tariff_id
    FROM prepared p
    JOIN users u ON u.telegram_id = p.telegram_id
    CROSS JOIN params
),
latest_active AS (
    SELECT DISTINCT ON (s.user_id)
        s.id,
        s.user_id,
        s.expires_at
    FROM subscriptions s
    JOIN src ON src.user_id = s.user_id
    WHERE s.status = 'active'
    ORDER BY s.user_id, s.expires_at DESC, s.id DESC
),
updated AS (
    UPDATE subscriptions s
    SET
        tariff_id = src.tariff_id,
        status = 'active',
        period_status = 'active',
        expires_at = GREATEST(s.expires_at, src.expires_at),
        current_period_start = COALESCE(s.current_period_start, now()),
        current_period_end = LEAST(now() + interval '30 days', GREATEST(s.expires_at, src.expires_at)),
        traffic_limit_bytes = src.traffic_limit_bytes,
        traffic_used_bytes = GREATEST(s.traffic_used_bytes, src.traffic_used_bytes),
        updated_at = now()
    FROM latest_active la
    JOIN src ON src.user_id = la.user_id
    WHERE s.id = la.id
    RETURNING s.user_id
)
INSERT INTO subscriptions (
    user_id,
    tariff_id,
    status,
    period_status,
    started_at,
    expires_at,
    current_period_start,
    current_period_end,
    traffic_limit_bytes,
    traffic_used_bytes,
    created_at,
    updated_at
)
SELECT
    src.user_id,
    src.tariff_id,
    'active',
    'active',
    now(),
    src.expires_at,
    now(),
    LEAST(now() + interval '30 days', src.expires_at),
    src.traffic_limit_bytes,
    src.traffic_used_bytes,
    now(),
    now()
FROM src
WHERE NOT EXISTS (
    SELECT 1 FROM updated u WHERE u.user_id = src.user_id
);

COMMIT;

\echo ''
\echo 'Import finished. Check result:'
SELECT
    u.telegram_id,
    u.telegram_username,
    u.alias,
    s.status,
    s.expires_at,
    s.traffic_limit_bytes,
    s.traffic_used_bytes
FROM subscriptions s
JOIN users u ON u.id = s.user_id
WHERE u.telegram_id IN (1839835493, 1136445296, 998952661, 984724049, 6075877693, 7992413092, 970706613, 802389929, 889789573, 884621667, 1238134452, 1766479687, 1203698798, 991438068)
ORDER BY u.telegram_username;

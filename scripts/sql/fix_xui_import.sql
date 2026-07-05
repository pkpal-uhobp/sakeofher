-- Fix already imported x-ui users in SakeOfHer database.
-- Rules:
-- 1) Set all imported users to 300 GiB traffic limit.
-- 2) Extend everyone whose subscription ends before 2026-07-20 23:59:59+03 to this date.
-- 3) Do not shorten anyone whose subscription already ends later.
-- 4) target_tariff_id is passed by scripts/import-xui-users.ps1.

BEGIN;

WITH imported_users(telegram_id) AS (
    VALUES
    (1839835493),
    (1136445296),
    (998952661),
    (984724049),
    (6075877693),
    (7992413092),
    (970706613),
    (802389929),
    (889789573),
    (884621667),
    (1238134452),
    (1766479687),
    (1203698798),
    (991438068)
),
params AS (
    SELECT
        :target_tariff_id::bigint AS target_tariff_id,
        322122547200::bigint AS target_traffic_limit_bytes,
        TIMESTAMPTZ '2026-07-20 23:59:59+03' AS min_expires_at
),
latest_active AS (
    SELECT DISTINCT ON (s.user_id)
        s.id
    FROM subscriptions s
    JOIN users u ON u.id = s.user_id
    JOIN imported_users iu ON iu.telegram_id = u.telegram_id
    WHERE s.status = 'active'
    ORDER BY s.user_id, s.expires_at DESC, s.id DESC
)
UPDATE subscriptions s
SET
    tariff_id = params.target_tariff_id,
    expires_at = GREATEST(s.expires_at, params.min_expires_at),
    current_period_end = LEAST(now() + interval '30 days', GREATEST(s.expires_at, params.min_expires_at)),
    traffic_limit_bytes = params.target_traffic_limit_bytes,
    period_status = 'active',
    updated_at = now()
FROM latest_active la
CROSS JOIN params
WHERE s.id = la.id;

COMMIT;

\echo ''
\echo 'Fix finished. Check result:'
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

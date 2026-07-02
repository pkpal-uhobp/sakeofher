-- 003_add_tariff_to_subscriptions.down.sql

BEGIN;

ALTER TABLE subscriptions
    DROP CONSTRAINT IF EXISTS subscriptions_tariff_or_payment_check;

DROP INDEX IF EXISTS idx_subscriptions_tariff_id;

ALTER TABLE subscriptions
    DROP COLUMN IF EXISTS tariff_id;

COMMIT;

-- 003_add_tariff_to_subscriptions.up.sql
-- Adds direct tariff reference to subscriptions so website/manual purchases can work without payment mechanics.

BEGIN;

ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS tariff_id BIGINT REFERENCES tariffs(id);

CREATE INDEX IF NOT EXISTS idx_subscriptions_tariff_id ON subscriptions (tariff_id);

-- Backfill old subscriptions that were created through payments.
UPDATE subscriptions s
SET tariff_id = p.tariff_id
FROM payments p
WHERE s.last_payment_id = p.id
  AND s.tariff_id IS NULL;

-- During transition we allow old payment-based records, but every new website subscription must set tariff_id.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'subscriptions_tariff_or_payment_check'
    ) THEN
        ALTER TABLE subscriptions
            ADD CONSTRAINT subscriptions_tariff_or_payment_check
            CHECK (tariff_id IS NOT NULL OR last_payment_id IS NOT NULL);
    END IF;
END $$;

COMMIT;

BEGIN;

CREATE TABLE IF NOT EXISTS subscription_lifecycle_events (
    id BIGSERIAL PRIMARY KEY,
    subscription_id BIGINT REFERENCES subscriptions(id) ON DELETE SET NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    payment_id BIGINT REFERENCES payments(id) ON DELETE SET NULL,
    event_type VARCHAR(80) NOT NULL,
    from_status VARCHAR(40),
    to_status VARCHAR(40),
    from_period_status VARCHAR(40),
    to_period_status VARCHAR(40),
    reason TEXT,
    success BOOLEAN NOT NULL DEFAULT true,
    error_text TEXT,
    details JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_subscription_lifecycle_events_subscription_id
    ON subscription_lifecycle_events(subscription_id);

CREATE INDEX IF NOT EXISTS idx_subscription_lifecycle_events_user_id
    ON subscription_lifecycle_events(user_id);

CREATE INDEX IF NOT EXISTS idx_subscription_lifecycle_events_type_created
    ON subscription_lifecycle_events(event_type, created_at DESC);

CREATE TABLE IF NOT EXISTS user_remna_squads (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    active_internal_squads TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE user_remna_squads
    ADD COLUMN IF NOT EXISTS desired_internal_squads TEXT[] NOT NULL DEFAULT '{}';

ALTER TABLE user_remna_squads
    ADD COLUMN IF NOT EXISTS sync_status VARCHAR(40) NOT NULL DEFAULT 'unknown';

ALTER TABLE user_remna_squads
    ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMPTZ;

ALTER TABLE user_remna_squads
    ADD COLUMN IF NOT EXISTS last_error TEXT;

UPDATE user_remna_squads
SET desired_internal_squads = active_internal_squads
WHERE desired_internal_squads = '{}'
  AND active_internal_squads <> '{}';

COMMIT;

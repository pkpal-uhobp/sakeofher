BEGIN;

CREATE TABLE IF NOT EXISTS subscription_notifications (
    id BIGSERIAL PRIMARY KEY,
    subscription_id BIGINT NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    notification_key VARCHAR(80) NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (subscription_id, notification_key)
);

CREATE INDEX IF NOT EXISTS idx_subscription_notifications_subscription_id
    ON subscription_notifications(subscription_id);

COMMIT;

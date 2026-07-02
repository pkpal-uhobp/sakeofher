-- 001_initial_schema.up.sql
-- PostgreSQL schema for VPN subscription service:
-- Telegram bot + website/backend + Remnawave + payments via Telegram Stars, Tribute, CryptoBot.
-- PostgreSQL 13+

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Auto-update updated_at column.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =========================
-- 1. USERS
-- =========================
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,

    -- Telegram user data.
    telegram_id BIGINT UNIQUE NOT NULL,
    telegram_username VARCHAR(255),
    telegram_first_name VARCHAR(255),
    telegram_last_name VARCHAR(255),
    language_code VARCHAR(20),

    alias VARCHAR(255),

    -- Remnawave data. Filled only after successful payment/activation.
    remna_uuid UUID UNIQUE,
    remna_username VARCHAR(255),
    subscription_url TEXT,

    -- Local user status in our system.
    status VARCHAR(50) NOT NULL DEFAULT 'active',

    -- Remnawave user lifecycle in our system:
    -- not_created -> active -> disabled -> deleted.
    remna_status VARCHAR(50) NOT NULL DEFAULT 'not_created',

    -- Business rule:
    -- subscription expired -> disable Remnawave user immediately;
    -- 7 days without payment -> delete Remnawave user.
    disabled_at TIMESTAMPTZ,
    delete_after TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT users_status_check
        CHECK (status IN ('active', 'blocked', 'deleted')),

    CONSTRAINT users_remna_status_check
        CHECK (remna_status IN ('not_created', 'active', 'disabled', 'deleted')),

    CONSTRAINT users_delete_after_check
        CHECK (delete_after IS NULL OR disabled_at IS NOT NULL)
);

CREATE INDEX idx_users_telegram_username ON users (telegram_username);
CREATE INDEX idx_users_remna_uuid ON users (remna_uuid);
CREATE INDEX idx_users_remna_status ON users (remna_status);
CREATE INDEX idx_users_last_seen_at ON users (last_seen_at);
CREATE INDEX idx_users_disabled_delete_after
    ON users (delete_after)
    WHERE remna_status = 'disabled' AND delete_after IS NOT NULL;

CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- =========================
-- 2. TARIFFS
-- =========================
-- Tariff = subscription product itself. Prices are stored separately.
CREATE TABLE tariffs (
    id BIGSERIAL PRIMARY KEY,

    code VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,

    -- Subscription duration.
    duration_days INT NOT NULL,

    -- Monthly traffic period. For 3 months this means 300 GB each 30-day period.
    period_days INT NOT NULL DEFAULT 30,
    traffic_limit_bytes BIGINT NOT NULL DEFAULT 322122547200, -- 300 GiB

    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT tariffs_duration_check
        CHECK (duration_days > 0),

    CONSTRAINT tariffs_period_days_check
        CHECK (period_days > 0),

    CONSTRAINT tariffs_traffic_limit_check
        CHECK (traffic_limit_bytes > 0)
);

CREATE INDEX idx_tariffs_active_sort ON tariffs (is_active, sort_order);
CREATE INDEX idx_tariffs_code ON tariffs (code);

CREATE TRIGGER trg_tariffs_updated_at
BEFORE UPDATE ON tariffs
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- =========================
-- 3. TARIFF PRICES
-- =========================
-- One tariff can have several prices: Stars, RUB via Tribute, crypto via CryptoBot.
CREATE TABLE tariff_prices (
    id BIGSERIAL PRIMARY KEY,

    tariff_id BIGINT NOT NULL REFERENCES tariffs(id) ON DELETE CASCADE,

    provider VARCHAR(50) NOT NULL,
    payment_method VARCHAR(50) NOT NULL,

    -- XTR for Telegram Stars, RUB for Tribute/CryptoBot fiat invoice.
    currency VARCHAR(20) NOT NULL,

    -- For RUB payments: kopecks. Example: 19900 = 199.00 RUB.
    amount_minor BIGINT,

    -- For Telegram Stars: exact stars count.
    stars_amount INT,

    -- For CryptoBot: accepted assets, for example {USDT,TON}.
    accepted_assets TEXT[] NOT NULL DEFAULT '{}'::TEXT[],

    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (tariff_id, provider, payment_method),

    CONSTRAINT tariff_prices_provider_check
        CHECK (provider IN ('telegram_stars', 'tribute', 'crypto_bot')),

    CONSTRAINT tariff_prices_method_check
        CHECK (payment_method IN ('stars', 'rub', 'crypto')),

    CONSTRAINT tariff_prices_amount_check
        CHECK (
            (provider = 'telegram_stars' AND currency = 'XTR' AND stars_amount IS NOT NULL AND stars_amount > 0 AND amount_minor IS NULL)
            OR
            (provider IN ('tribute', 'crypto_bot') AND currency = 'RUB' AND amount_minor IS NOT NULL AND amount_minor > 0 AND stars_amount IS NULL)
        ),

    CONSTRAINT tariff_prices_provider_method_check
        CHECK (
            (provider = 'telegram_stars' AND payment_method = 'stars')
            OR (provider = 'tribute' AND payment_method = 'rub')
            OR (provider = 'crypto_bot' AND payment_method = 'crypto')
        )
);

CREATE INDEX idx_tariff_prices_tariff_id ON tariff_prices (tariff_id);
CREATE INDEX idx_tariff_prices_provider ON tariff_prices (provider, payment_method);
CREATE INDEX idx_tariff_prices_active_sort ON tariff_prices (is_active, sort_order);

CREATE TRIGGER trg_tariff_prices_updated_at
BEFORE UPDATE ON tariff_prices
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- =========================
-- 4. PAYMENTS
-- =========================
CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,

    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tariff_id BIGINT NOT NULL REFERENCES tariffs(id),
    tariff_price_id BIGINT REFERENCES tariff_prices(id),

    provider VARCHAR(50) NOT NULL,
    payment_method VARCHAR(50) NOT NULL,

    -- Snapshot copied at payment creation. Tariff changes must not affect old payments.
    currency VARCHAR(20) NOT NULL,
    amount_minor BIGINT,
    stars_amount INT,
    duration_days INT NOT NULL,
    period_days INT NOT NULL DEFAULT 30,
    traffic_limit_bytes BIGINT NOT NULL,

    status VARCHAR(50) NOT NULL DEFAULT 'created',

    -- External payment system identifiers:
    -- Telegram Stars charge id / CryptoBot invoice id / Tribute order id.
    provider_payment_id VARCHAR(255),
    payment_url TEXT,

    -- CryptoBot actual paid details.
    paid_asset VARCHAR(20),
    paid_amount NUMERIC(18, 8),
    fee_asset VARCHAR(20),
    fee_amount NUMERIC(18, 8),

    expires_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    activated_at TIMESTAMPTZ,

    -- Last raw payload from provider, useful for support/debug.
    raw_payload JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT payments_status_check
        CHECK (status IN (
            'created',
            'waiting_payment',
            'paid',
            'activation_failed',
            'activated',
            'failed',
            'cancelled',
            'expired',
            'refunded'
        )),

    CONSTRAINT payments_provider_check
        CHECK (provider IN ('telegram_stars', 'tribute', 'crypto_bot')),

    CONSTRAINT payments_method_check
        CHECK (payment_method IN ('stars', 'rub', 'crypto')),

    CONSTRAINT payments_provider_method_check
        CHECK (
            (provider = 'telegram_stars' AND payment_method = 'stars')
            OR (provider = 'tribute' AND payment_method = 'rub')
            OR (provider = 'crypto_bot' AND payment_method = 'crypto')
        ),

    CONSTRAINT payments_amount_check
        CHECK (
            (provider = 'telegram_stars' AND currency = 'XTR' AND stars_amount IS NOT NULL AND stars_amount > 0 AND amount_minor IS NULL)
            OR
            (provider IN ('tribute', 'crypto_bot') AND currency = 'RUB' AND amount_minor IS NOT NULL AND amount_minor > 0 AND stars_amount IS NULL)
        ),

    CONSTRAINT payments_duration_check
        CHECK (duration_days > 0),

    CONSTRAINT payments_period_check
        CHECK (period_days > 0),

    CONSTRAINT payments_traffic_check
        CHECK (traffic_limit_bytes > 0)
);

-- Allows repeated NULL provider_payment_id, but prevents duplicates once it exists.
CREATE UNIQUE INDEX uq_payments_provider_payment
ON payments (provider, provider_payment_id)
WHERE provider_payment_id IS NOT NULL;

CREATE INDEX idx_payments_user_id ON payments (user_id);
CREATE INDEX idx_payments_status ON payments (status);
CREATE INDEX idx_payments_created_at ON payments (created_at);
CREATE INDEX idx_payments_paid_at ON payments (paid_at);
CREATE INDEX idx_payments_provider_status ON payments (provider, status);
CREATE INDEX idx_payments_waiting_expires
    ON payments (expires_at)
    WHERE status = 'waiting_payment' AND expires_at IS NOT NULL;

CREATE TRIGGER trg_payments_updated_at
BEFORE UPDATE ON payments
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- =========================
-- 5. PAYMENT EVENTS
-- =========================
-- Idempotency table for Telegram Stars successful_payment, Tribute webhooks, CryptoBot webhooks.
CREATE TABLE payment_events (
    id BIGSERIAL PRIMARY KEY,

    provider VARCHAR(50) NOT NULL,
    event_id VARCHAR(255) NOT NULL,

    payment_id BIGINT REFERENCES payments(id) ON DELETE SET NULL,

    event_type VARCHAR(100) NOT NULL,
    raw_payload JSONB NOT NULL,

    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT payment_events_provider_check
        CHECK (provider IN ('telegram_stars', 'tribute', 'crypto_bot')),

    UNIQUE (provider, event_id)
);

CREATE INDEX idx_payment_events_payment_id ON payment_events (payment_id);
CREATE INDEX idx_payment_events_provider_type ON payment_events (provider, event_type);
CREATE INDEX idx_payment_events_created_at ON payment_events (created_at);

-- =========================
-- 6. SUBSCRIPTIONS
-- =========================
CREATE TABLE subscriptions (
    id BIGSERIAL PRIMARY KEY,

    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_payment_id BIGINT REFERENCES payments(id) ON DELETE SET NULL,

    status VARCHAR(50) NOT NULL DEFAULT 'active',

    started_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,

    -- Current traffic period.
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,

    traffic_limit_bytes BIGINT NOT NULL,
    traffic_used_bytes BIGINT NOT NULL DEFAULT 0,

    -- active = normal period, traffic_exhausted = disabled until next period/payment, finished = subscription ended.
    period_status VARCHAR(50) NOT NULL DEFAULT 'active',

    -- Pretty subscription page: site.ru/s/{public_token}
    public_token VARCHAR(100) UNIQUE NOT NULL DEFAULT encode(gen_random_bytes(24), 'hex'),

    -- Remnawave usage sync.
    last_remna_check_at TIMESTAMPTZ,

    -- Notifications.
    last_expire_notification_at TIMESTAMPTZ,
    last_traffic_notification_at TIMESTAMPTZ,

    notified_3_days BOOLEAN NOT NULL DEFAULT false,
    notified_1_day BOOLEAN NOT NULL DEFAULT false,
    notified_expired BOOLEAN NOT NULL DEFAULT false,

    traffic_80_notified BOOLEAN NOT NULL DEFAULT false,
    traffic_95_notified BOOLEAN NOT NULL DEFAULT false,
    traffic_exhausted_notified BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT subscriptions_status_check
        CHECK (status IN ('active', 'expired', 'cancelled')),

    CONSTRAINT subscriptions_period_status_check
        CHECK (period_status IN ('active', 'traffic_exhausted', 'finished')),

    CONSTRAINT subscriptions_dates_check
        CHECK (expires_at > started_at),

    CONSTRAINT subscriptions_period_dates_check
        CHECK (current_period_end > current_period_start),

    CONSTRAINT subscriptions_traffic_limit_check
        CHECK (traffic_limit_bytes > 0),

    CONSTRAINT subscriptions_traffic_used_check
        CHECK (traffic_used_bytes >= 0)
);

-- Only one active subscription per user.
CREATE UNIQUE INDEX uq_subscriptions_one_active_per_user
ON subscriptions (user_id)
WHERE status = 'active';

CREATE INDEX idx_subscriptions_user_id ON subscriptions (user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions (status);
CREATE INDEX idx_subscriptions_expires_at ON subscriptions (expires_at);
CREATE INDEX idx_subscriptions_public_token ON subscriptions (public_token);
CREATE INDEX idx_subscriptions_period_end ON subscriptions (current_period_end);
CREATE INDEX idx_subscriptions_last_remna_check ON subscriptions (last_remna_check_at);
CREATE INDEX idx_subscriptions_active_expiry
    ON subscriptions (expires_at)
    WHERE status = 'active';
CREATE INDEX idx_subscriptions_period_reset
    ON subscriptions (current_period_end)
    WHERE status = 'active';

CREATE TRIGGER trg_subscriptions_updated_at
BEFORE UPDATE ON subscriptions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- =========================
-- 7. ADMINS
-- =========================
CREATE TABLE admins (
    id BIGSERIAL PRIMARY KEY,

    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),

    role VARCHAR(50) NOT NULL DEFAULT 'admin',
    is_active BOOLEAN NOT NULL DEFAULT true,

    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT admins_role_check
        CHECK (role IN ('owner', 'admin', 'support'))
);

CREATE INDEX idx_admins_telegram_id ON admins (telegram_id);
CREATE INDEX idx_admins_active ON admins (is_active);

CREATE TRIGGER trg_admins_updated_at
BEFORE UPDATE ON admins
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- =========================
-- 8. ADMIN ACTIONS
-- =========================
CREATE TABLE admin_actions (
    id BIGSERIAL PRIMARY KEY,

    admin_id BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    target_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,

    action VARCHAR(100) NOT NULL,
    details JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_admin_actions_admin_id ON admin_actions (admin_id);
CREATE INDEX idx_admin_actions_target_user_id ON admin_actions (target_user_id);
CREATE INDEX idx_admin_actions_created_at ON admin_actions (created_at);
CREATE INDEX idx_admin_actions_action ON admin_actions (action);

-- =========================
-- 9. BROADCASTS
-- =========================
CREATE TABLE broadcasts (
    id BIGSERIAL PRIMARY KEY,

    admin_id BIGINT REFERENCES admins(id) ON DELETE SET NULL,

    message_text TEXT NOT NULL,
    parse_mode VARCHAR(50),

    status VARCHAR(50) NOT NULL DEFAULT 'draft',

    target_filter JSONB,
    total_count INT NOT NULL DEFAULT 0,
    sent_count INT NOT NULL DEFAULT 0,
    failed_count INT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT broadcasts_status_check
        CHECK (status IN ('draft', 'queued', 'sending', 'finished', 'failed', 'cancelled')),

    CONSTRAINT broadcasts_counts_check
        CHECK (
            total_count >= 0
            AND sent_count >= 0
            AND failed_count >= 0
        )
);

CREATE INDEX idx_broadcasts_status ON broadcasts (status);
CREATE INDEX idx_broadcasts_created_at ON broadcasts (created_at);

CREATE TRIGGER trg_broadcasts_updated_at
BEFORE UPDATE ON broadcasts
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- =========================
-- 10. BROADCAST RECIPIENTS
-- =========================
CREATE TABLE broadcast_recipients (
    id BIGSERIAL PRIMARY KEY,

    broadcast_id BIGINT NOT NULL REFERENCES broadcasts(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_text TEXT,

    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (broadcast_id, user_id),

    CONSTRAINT broadcast_recipients_status_check
        CHECK (status IN ('pending', 'sent', 'failed', 'skipped'))
);

CREATE INDEX idx_broadcast_recipients_broadcast ON broadcast_recipients (broadcast_id);
CREATE INDEX idx_broadcast_recipients_status ON broadcast_recipients (status);
CREATE INDEX idx_broadcast_recipients_user ON broadcast_recipients (user_id);

CREATE TRIGGER trg_broadcast_recipients_updated_at
BEFORE UPDATE ON broadcast_recipients
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- =========================
-- 11. REMNAWAVE SYNC LOGS
-- =========================
CREATE TABLE remna_sync_logs (
    id BIGSERIAL PRIMARY KEY,

    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    subscription_id BIGINT REFERENCES subscriptions(id) ON DELETE SET NULL,
    payment_id BIGINT REFERENCES payments(id) ON DELETE SET NULL,

    action VARCHAR(100) NOT NULL,
    success BOOLEAN NOT NULL DEFAULT false,
    error_text TEXT,
    request_payload JSONB,
    response_payload JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_remna_sync_logs_user_id ON remna_sync_logs (user_id);
CREATE INDEX idx_remna_sync_logs_payment_id ON remna_sync_logs (payment_id);
CREATE INDEX idx_remna_sync_logs_action ON remna_sync_logs (action);
CREATE INDEX idx_remna_sync_logs_success ON remna_sync_logs (success);
CREATE INDEX idx_remna_sync_logs_created_at ON remna_sync_logs (created_at);

COMMIT;

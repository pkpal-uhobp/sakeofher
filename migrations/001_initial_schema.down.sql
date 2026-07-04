-- 001_initial_schema.down.sql
-- Rollback grouped initial schema.

BEGIN;

DROP TABLE IF EXISTS subscription_lifecycle_events;
DROP TABLE IF EXISTS subscription_notifications;
DROP TABLE IF EXISTS user_remna_squads;
DROP TABLE IF EXISTS remna_sync_logs;
DROP TABLE IF EXISTS broadcast_recipients;
DROP TABLE IF EXISTS broadcasts;
DROP TABLE IF EXISTS admin_actions;
DROP TABLE IF EXISTS admins;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS payment_events;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS tariff_prices;
DROP TABLE IF EXISTS tariffs;
DROP TABLE IF EXISTS users;

DROP FUNCTION IF EXISTS set_updated_at();

COMMIT;

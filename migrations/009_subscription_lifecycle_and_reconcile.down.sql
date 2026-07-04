BEGIN;

ALTER TABLE user_remna_squads
    DROP COLUMN IF EXISTS last_error;

ALTER TABLE user_remna_squads
    DROP COLUMN IF EXISTS last_synced_at;

ALTER TABLE user_remna_squads
    DROP COLUMN IF EXISTS sync_status;

ALTER TABLE user_remna_squads
    DROP COLUMN IF EXISTS desired_internal_squads;

DROP TABLE IF EXISTS subscription_lifecycle_events;

COMMIT;

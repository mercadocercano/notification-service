DROP INDEX IF EXISTS uq_notifications_dedup;
ALTER TABLE notifications DROP COLUMN IF EXISTS dedup_key;

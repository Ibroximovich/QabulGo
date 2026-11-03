-- 000010_add_max_expires_at_to_refresh_sessions.down.sql
DROP INDEX IF EXISTS idx_refresh_sessions_max_expires_at;
ALTER TABLE refresh_sessions DROP COLUMN IF EXISTS max_expires_at;


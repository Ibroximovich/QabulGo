-- 000010_add_max_expires_at_to_refresh_sessions.up.sql
ALTER TABLE refresh_sessions ADD COLUMN IF NOT EXISTS max_expires_at TIMESTAMPTZ;

-- Dastlabki mavjud sessiyalar uchun max_expires_at ni expires_at ga tenglashtiramiz
UPDATE refresh_sessions SET max_expires_at = expires_at WHERE max_expires_at IS NULL;

-- Ustunni NOT NULL qilamiz
ALTER TABLE refresh_sessions ALTER COLUMN max_expires_at SET NOT NULL;

-- Indeks yaratish
CREATE INDEX IF NOT EXISTS idx_refresh_sessions_max_expires_at ON refresh_sessions(max_expires_at);


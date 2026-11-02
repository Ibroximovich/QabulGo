-- ============================================================================
-- 000009_create_telegram_tables.up.sql
-- QabulGo Telegram Bot Integration (Clinic Bot & Superadmin Bot)
-- ============================================================================

-- 1. Clinic Telegram Notification Recipients
CREATE TABLE IF NOT EXISTS clinic_telegram_recipients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    telegram_user_id BIGINT NOT NULL,
    telegram_chat_id BIGINT NOT NULL,
    telegram_username VARCHAR(255),
    telegram_first_name VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_clinic_telegram_recipient UNIQUE (clinic_id, telegram_chat_id)
);

CREATE INDEX IF NOT EXISTS idx_clinic_telegram_recipients_clinic_id ON clinic_telegram_recipients(clinic_id);
CREATE INDEX IF NOT EXISTS idx_clinic_telegram_recipients_tg_user_id ON clinic_telegram_recipients(telegram_user_id);
CREATE INDEX IF NOT EXISTS idx_clinic_telegram_recipients_active ON clinic_telegram_recipients(is_active, notifications_enabled);

-- 2. One-Time Cryptographic Linking Tokens
CREATE TABLE IF NOT EXISTS telegram_linking_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_telegram_linking_tokens_hash ON telegram_linking_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_telegram_linking_tokens_clinic_id ON telegram_linking_tokens(clinic_id);
CREATE INDEX IF NOT EXISTS idx_telegram_linking_tokens_active ON telegram_linking_tokens(is_used, expires_at);

-- 3. Reliable Notification Outbox (Transactional Outbox Pattern)
CREATE TABLE IF NOT EXISTS notification_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL, -- 'appointment_created', 'appointment_cancelled', 'clinic_submitted'
    entity_type VARCHAR(50) NOT NULL, -- 'appointment', 'clinic'
    entity_id UUID NOT NULL,
    target_clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    target_role VARCHAR(50) NOT NULL, -- 'clinic', 'admin'
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'delivered', 'failed', 'skipped'
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 5,
    next_retry_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    delivered_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notification_outbox_poll ON notification_outbox(status, next_retry_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_notification_outbox_entity ON notification_outbox(entity_type, entity_id);

-- 4. Telegram Update Deduplication Table
CREATE TABLE IF NOT EXISTS telegram_update_dedup (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bot_type VARCHAR(20) NOT NULL, -- 'clinic', 'admin'
    update_id BIGINT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_telegram_update_dedup UNIQUE (bot_type, update_id)
);

CREATE INDEX IF NOT EXISTS idx_telegram_update_dedup_time ON telegram_update_dedup(processed_at);

-- 5. Admin Bot Interactive Session State (e.g. awaiting rejection reason)
CREATE TABLE IF NOT EXISTS telegram_admin_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_user_id BIGINT NOT NULL UNIQUE,
    awaiting_action VARCHAR(50), -- 'reject_clinic'
    target_clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


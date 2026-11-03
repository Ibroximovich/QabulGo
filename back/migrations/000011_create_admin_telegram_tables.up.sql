-- ============================================================================
-- 000011_create_admin_telegram_tables.up.sql
-- Superadmin Telegram Bot Integration & Dynamic Account Linking
-- ============================================================================

-- 1. Superadmin Telegram Connected Recipients
CREATE TABLE IF NOT EXISTS admin_telegram_recipients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    telegram_user_id BIGINT NOT NULL,
    telegram_chat_id BIGINT NOT NULL,
    telegram_username VARCHAR(255),
    telegram_first_name VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_admin_telegram_recipient_user UNIQUE (user_id),
    CONSTRAINT uq_admin_telegram_recipient_chat UNIQUE (telegram_chat_id)
);

CREATE INDEX IF NOT EXISTS idx_admin_telegram_recipients_user ON admin_telegram_recipients(user_id);
CREATE INDEX IF NOT EXISTS idx_admin_telegram_recipients_tg_user ON admin_telegram_recipients(telegram_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_telegram_recipients_active ON admin_telegram_recipients(is_active);

-- 2. One-Time Cryptographic Linking Tokens for Superadmin
CREATE TABLE IF NOT EXISTS admin_telegram_linking_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_telegram_linking_tokens_hash ON admin_telegram_linking_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_admin_telegram_linking_tokens_user ON admin_telegram_linking_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_admin_telegram_linking_tokens_active ON admin_telegram_linking_tokens(is_used, expires_at);

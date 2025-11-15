-- Migration: Token blacklist for logout functionality
-- Created: 2025-11-14
-- Description: Таблица для хранения инвалидированных токенов (logout)

CREATE TABLE IF NOT EXISTS token_blacklist (
    id BIGSERIAL PRIMARY KEY,

    -- SHA256 hash токена (для безопасности не храним сам токен)
    token_hash VARCHAR(64) NOT NULL UNIQUE,

    -- Время истечения токена (для автоматической очистки)
    expires_at TIMESTAMPTZ NOT NULL,

    -- Время добавления в blacklist
    revoked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- ID пользователя (для быстрой очистки при logout-all)
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_token_blacklist_hash ON token_blacklist(token_hash);
CREATE INDEX IF NOT EXISTS idx_token_blacklist_user_id ON token_blacklist(user_id);
CREATE INDEX IF NOT EXISTS idx_token_blacklist_expires_at ON token_blacklist(expires_at);

-- Функция для автоматической очистки истекших токенов (опционально, можно запускать по cron)
-- DELETE FROM token_blacklist WHERE expires_at < NOW();


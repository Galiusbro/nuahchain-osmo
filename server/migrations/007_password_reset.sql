-- Migration: Password reset tokens
-- Created: 2025-11-14
-- Description: Таблица для хранения токенов сброса пароля

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id BIGSERIAL PRIMARY KEY,

    -- Связь с пользователем
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Токен для сброса пароля (случайная строка)
    token VARCHAR(255) NOT NULL UNIQUE,

    -- Время истечения токена (обычно 1 час)
    expires_at TIMESTAMPTZ NOT NULL,

    -- Использован ли токен
    used BOOLEAN NOT NULL DEFAULT FALSE,

    -- IP адрес с которого был запрошен сброс
    ip_address VARCHAR(45),

    -- Временные метки
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_active ON password_reset_tokens(token, expires_at, used) WHERE used = FALSE;

-- Автоматическая очистка истекших токенов (можно запускать по cron)
-- DELETE FROM password_reset_tokens WHERE expires_at < NOW() OR (used = TRUE AND used_at < NOW() - INTERVAL '7 days');


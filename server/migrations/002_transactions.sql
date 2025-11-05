-- Migration: Transactions tracking
-- Created: 2025-11-06
-- Description: Таблица для записи всех операций пользователей (токены, активы)

-- Типы операций:
-- TOKEN_CREATE - создание токена
-- TOKEN_BUY - покупка токена
-- TOKEN_SELL - продажа токена
-- ASSET_ENSURE - создание/обеспечение актива
-- ASSET_BUY - покупка актива
-- ASSET_SELL - продажа актива

-- Статусы транзакций:
-- PENDING - транзакция отправлена, ожидает подтверждения
-- SUCCESS - транзакция успешно выполнена
-- FAILED - транзакция не удалась

CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,

    -- Связь с пользователем
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Тип операции (см. выше)
    operation_type VARCHAR(50) NOT NULL,

    -- Хеш транзакции в блокчейне
    tx_hash VARCHAR(255) NOT NULL,

    -- Статус транзакции (PENDING, SUCCESS, FAILED)
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',

    -- Данные операции (JSON для гибкости)
    -- Для токенов: denom, symbol, name, amount, payment_denom, tokens_out, price_paid
    -- Для активов: symbol, denom, amount, base_amount, payout_ndollar
    operation_data JSONB,

    -- Сообщение об ошибке (если статус FAILED)
    error_message TEXT,

    -- Временные метки
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Для быстрого поиска
    CONSTRAINT valid_operation_type CHECK (
        operation_type IN (
            'TOKEN_CREATE', 'TOKEN_BUY', 'TOKEN_SELL',
            'ASSET_ENSURE', 'ASSET_BUY', 'ASSET_SELL'
        )
    ),
    CONSTRAINT valid_status CHECK (
        status IN ('PENDING', 'SUCCESS', 'FAILED')
    )
);

-- Индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_operation_type ON transactions(operation_type);
CREATE INDEX IF NOT EXISTS idx_transactions_tx_hash ON transactions(tx_hash);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_user_type ON transactions(user_id, operation_type);
CREATE INDEX IF NOT EXISTS idx_transactions_user_created ON transactions(user_id, created_at DESC);

-- Комментарии к таблице и полям (для документации)
COMMENT ON TABLE transactions IS 'Таблица для хранения всех операций пользователей с токенами и активами';
COMMENT ON COLUMN transactions.operation_type IS 'Тип операции: TOKEN_CREATE, TOKEN_BUY, TOKEN_SELL, ASSET_ENSURE, ASSET_BUY, ASSET_SELL';
COMMENT ON COLUMN transactions.status IS 'Статус: PENDING (отправлена), SUCCESS (успешно), FAILED (ошибка)';
COMMENT ON COLUMN transactions.operation_data IS 'JSON данные операции (denom, symbol, amount и т.д.)';
COMMENT ON COLUMN transactions.error_message IS 'Сообщение об ошибке, если транзакция не удалась';


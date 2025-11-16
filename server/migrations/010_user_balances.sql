-- Migration: Create user balances tracking tables
-- Created: 2025-11-15
-- Purpose: Track user balances in database with history for real-time updates

BEGIN;

-- 1. Current balances table
CREATE TABLE IF NOT EXISTS user_balances (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address VARCHAR(255) NOT NULL,  -- Cosmos address
    denom VARCHAR(255) NOT NULL,    -- Token denom
    amount NUMERIC NOT NULL,        -- Balance amount (as string for precision)
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by_tx_hash VARCHAR(64),  -- Last transaction that changed this balance
    updated_by_height BIGINT,       -- Block height of last update

    UNIQUE(user_id, denom),
    UNIQUE(address, denom)
);

CREATE INDEX IF NOT EXISTS idx_user_balances_user_id ON user_balances(user_id);
CREATE INDEX IF NOT EXISTS idx_user_balances_address ON user_balances(address);
CREATE INDEX IF NOT EXISTS idx_user_balances_denom ON user_balances(denom);
CREATE INDEX IF NOT EXISTS idx_user_balances_updated ON user_balances(updated_at);
CREATE INDEX IF NOT EXISTS idx_user_balances_height ON user_balances(updated_by_height);

-- 2. Balance history table
CREATE TABLE IF NOT EXISTS balance_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address VARCHAR(255) NOT NULL,
    denom VARCHAR(255) NOT NULL,
    amount_before NUMERIC,          -- Balance before change (NULL for first entry)
    amount_after NUMERIC NOT NULL,  -- Balance after change
    amount_delta NUMERIC NOT NULL,  -- Change amount (positive = increase, negative = decrease)
    tx_hash VARCHAR(64) NOT NULL,   -- Transaction that caused the change
    height BIGINT NOT NULL,         -- Block height
    event_type VARCHAR(50),         -- 'transfer', 'coin_received', 'coin_spent', 'sync'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_balance_history_user_id ON balance_history(user_id);
CREATE INDEX IF NOT EXISTS idx_balance_history_address ON balance_history(address);
CREATE INDEX IF NOT EXISTS idx_balance_history_denom ON balance_history(denom);
CREATE INDEX IF NOT EXISTS idx_balance_history_tx_hash ON balance_history(tx_hash);
CREATE INDEX IF NOT EXISTS idx_balance_history_height ON balance_history(height);
CREATE INDEX IF NOT EXISTS idx_balance_history_created ON balance_history(created_at);
CREATE INDEX IF NOT EXISTS idx_balance_history_user_denom ON balance_history(user_id, denom, created_at DESC);

-- 3. Balance indexer state table (tracks last processed block)
CREATE TABLE IF NOT EXISTS balance_indexer_state (
    id INT PRIMARY KEY DEFAULT 1,
    last_processed_height BIGINT NOT NULL DEFAULT 0,
    last_processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT single_row CHECK (id = 1)
);

-- Initialize indexer state
INSERT INTO balance_indexer_state (id, last_processed_height, last_processed_at)
VALUES (1, 0, NOW())
ON CONFLICT (id) DO NOTHING;

-- 4. Failed balance updates (Dead Letter Queue)
CREATE TABLE IF NOT EXISTS failed_balance_updates (
    id BIGSERIAL PRIMARY KEY,
    tx_hash VARCHAR(64) NOT NULL,
    height BIGINT NOT NULL,
    address VARCHAR(255) NOT NULL,
    denom VARCHAR(255) NOT NULL,
    amount_delta NUMERIC NOT NULL,
    error_message TEXT NOT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_retry_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_failed_balance_updates_tx_hash ON failed_balance_updates(tx_hash);
CREATE INDEX IF NOT EXISTS idx_failed_balance_updates_address ON failed_balance_updates(address);
CREATE INDEX IF NOT EXISTS idx_failed_balance_updates_retry_count ON failed_balance_updates(retry_count);
CREATE INDEX IF NOT EXISTS idx_failed_balance_updates_created ON failed_balance_updates(created_at);

COMMENT ON TABLE user_balances IS 'Current balances for all users, updated in real-time from blockchain events';
COMMENT ON TABLE balance_history IS 'History of all balance changes for audit and analytics';
COMMENT ON TABLE balance_indexer_state IS 'Tracks the last processed block height for balance indexer';
COMMENT ON TABLE failed_balance_updates IS 'Dead letter queue for failed balance updates';

COMMIT;


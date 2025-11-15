-- Migration: Create tokens table for storing token metadata
-- Created: 2025-11-14
-- Purpose: Store token metadata (name, symbol, image, etc.) to avoid blockchain queries

CREATE TABLE IF NOT EXISTS tokens (
    id BIGSERIAL PRIMARY KEY,
    denom VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    symbol VARCHAR(100) NOT NULL,
    image VARCHAR(500),
    description TEXT,
    creator_address VARCHAR(255) NOT NULL,
    creator_user_id BIGINT,
    decimals INTEGER DEFAULT 6,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_tokens_denom ON tokens(denom);
CREATE INDEX IF NOT EXISTS idx_tokens_creator_user_id ON tokens(creator_user_id);
CREATE INDEX IF NOT EXISTS idx_tokens_symbol ON tokens(symbol);
CREATE INDEX IF NOT EXISTS idx_tokens_name ON tokens(name);

-- Add comment
COMMENT ON TABLE tokens IS 'Stores token metadata to avoid blockchain queries when listing user tokens';


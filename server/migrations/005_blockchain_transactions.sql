-- 005_blockchain_transactions.sql
-- Create table for storing all blockchain transactions from WebSocket monitoring

BEGIN;

CREATE TABLE IF NOT EXISTS blockchain_transactions (
    tx_hash VARCHAR(64) PRIMARY KEY,
    height BIGINT NOT NULL,
    code INT NOT NULL,
    codespace VARCHAR(50),
    success BOOLEAN NOT NULL,
    log TEXT,
    raw_log TEXT,
    info TEXT,

    -- Gas information
    gas_wanted BIGINT,
    gas_used BIGINT,

    -- Fee information
    fee_amount JSONB,  -- {denom: "unuah", amount: "1000"}

    -- Transaction details
    messages JSONB,   -- [{type: "MsgBuyAsset", value: {...}}]
    signers TEXT[],   -- ["cosmos1...", "cosmos1..."]
    memo TEXT,

    -- Events and responses
    events JSONB NOT NULL,  -- All events with attributes
    msg_responses JSONB,   -- Responses from modules

    -- Extracted from events (for quick search)
    operation_type VARCHAR(200),  -- From message.action
    sender TEXT,                  -- From message.sender
    module_name VARCHAR(50),      -- From events

    -- Timestamps
    block_timestamp TIMESTAMPTZ,  -- Block time from GetTx
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),  -- When we received the event

    -- Enrichment status
    enriched BOOLEAN DEFAULT FALSE,  -- Whether we fetched full data via GetTx
    enriched_at TIMESTAMPTZ,         -- When we enriched

    -- Additional data
    data TEXT,  -- Base64 encoded result data
    tx_bytes TEXT  -- Base64 encoded transaction bytes (from WebSocket event)
);

-- Indexes for fast queries
CREATE INDEX idx_blockchain_tx_height ON blockchain_transactions(height);
CREATE INDEX idx_blockchain_tx_sender ON blockchain_transactions(sender) WHERE sender IS NOT NULL;
CREATE INDEX idx_blockchain_tx_operation ON blockchain_transactions(operation_type) WHERE operation_type IS NOT NULL;
CREATE INDEX idx_blockchain_tx_module ON blockchain_transactions(module_name) WHERE module_name IS NOT NULL;
CREATE INDEX idx_blockchain_tx_created ON blockchain_transactions(created_at);
CREATE INDEX idx_blockchain_tx_enriched ON blockchain_transactions(enriched) WHERE enriched = FALSE;
CREATE INDEX idx_blockchain_tx_success ON blockchain_transactions(success);

-- GIN index for JSONB fields (for complex queries)
CREATE INDEX idx_blockchain_tx_events_gin ON blockchain_transactions USING GIN (events);
CREATE INDEX idx_blockchain_tx_messages_gin ON blockchain_transactions USING GIN (messages) WHERE messages IS NOT NULL;

COMMIT;


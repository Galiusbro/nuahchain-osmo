-- Migration: Extend operation types for exchange and stablecoin
-- Created: 2025-11-11
-- Description: Adds EXCHANGE, STABLECOIN_BUY, STABLECOIN_SELL to allowed transaction operation types.

BEGIN;

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS valid_operation_type;

ALTER TABLE transactions
    ADD CONSTRAINT valid_operation_type CHECK (
        operation_type IN (
            'TOKEN_CREATE', 'TOKEN_BUY', 'TOKEN_SELL',
            'ASSET_ENSURE', 'ASSET_BUY', 'ASSET_SELL',
            'ASSET_MARGIN_OPEN', 'ASSET_MARGIN_CLOSE',
            'EXCHANGE', 'STABLECOIN_BUY', 'STABLECOIN_SELL'
        )
    );

COMMENT ON CONSTRAINT valid_operation_type ON transactions IS 'Allowed values for transactions.operation_type including exchange and stablecoin operations';

COMMIT;

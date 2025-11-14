# Exchange Module

## Overview

The Exchange module provides API endpoints for exchanging supported cryptocurrencies (ETH, BTC, USDC, USDT, ATOM, OSMO, SOL) for unuah using real-time price feeds from the USD Oracle. Exchanges now respond immediately with `status: "PENDING"`; a background tracker finalises each transaction and updates the database.

## Architecture

```
server/exchange/
├── models.go      # Request/Response structures
├── service.go     # Business logic
├── handlers.go    # HTTP handlers
└── README.md      # This file
```

## Components

### 1. Models (`models.go`)

Defines the API request and response structures:

- **ExchangeTokensRequest**: Input for token exchange
  - `token_denom`: Token to exchange (e.g., "uusdc", "ueth")
  - `amount`: Amount in base units
  - `min_output`: Minimum unuah output (slippage protection)

- **ExchangeTokensResponse**: Result of token exchange
  - `status`: `PENDING`, `SUCCESS`, or `FAILED`
  - `tx_hash`: Blockchain transaction hash
  - `unuah_out`: Amount of unuah received (populated when known)
  - `error_msg`: Error message if failed
  - `message`: Optional human-friendly status

### 2. Service (`service.go`)

Implements the core business logic:

- **NewService**: Constructor for Exchange service
- **ExchangeTokens**: Main exchange logic
  - Authenticates the user and decrypts wallet
  - Parses and validates amounts
  - Records the transaction as `PENDING` in the database (operation type `EXCHANGE`)
  - Broadcasts the blockchain exchange via `blockchain.ExchangeTokensWithKey`
  - Enqueues the transaction hash for background tracking

### 3. Handlers (`handlers.go`)

HTTP request handlers:

- **HandleExchangeTokens**: POST `/api/exchange/tokens`
  - Authenticates user via JWT
  - Parses request body
  - Calls service layer
  - Returns JSON response

### 4. Blockchain Client (`server/blockchain/exchange.go`)

Low-level blockchain interaction:

- **ExchangeTokensWithKey**: Signs and broadcasts exchange transaction
- **extractExchangeOutput**: Parses transaction response for output amount

## Supported Tokens

The following tokens are supported by the exchange module (if configured in x/usdoracle):

- **ueth** - Ethereum
- **ubtc** - Bitcoin
- **uusdc** - USD Coin
- **uusdt** - Tether USD
- **uatom** - Cosmos ATOM
- **uosmo** - Osmosis
- **usol** - Solana

## Exchange Process

1. User sends exchange request with token and amount
2. Server validates authentication and request
3. Exchange module fetches current USD price from USD Oracle
4. Price is validated against TWAP for security
5. Exchange fee is applied (default: 0.1%)
6. unuah is minted and sent to user
7. Input tokens are sent to treasury or community pool
8. Transaction is recorded in database

## Rate Limits and Validation

- **Min exchange amount**: $10 USD equivalent
- **Max exchange amount**: $100,000 USD equivalent
- **Daily limit**: $1,000,000 USD equivalent per user
- **Price deviation threshold**: 2% (Oracle vs TWAP)

## Database Recording

All exchange transactions are recorded in the shared `transactions` table with:

- `operation_type`: `EXCHANGE`
- `status`: `PENDING` → `SUCCESS` or `FAILED`
- `tx_hash`: Blockchain transaction hash
- `operation_data_exchange`: JSON blob with request metadata (token_denom, amount, min_output)
- `error_message`: Failure reason (on error)

The transaction tracker polls the blockchain (`GetTx`) and updates the record once the transaction is finalised. Clients can poll `/api/tx/<tx_hash>` to retrieve raw logs and events.

## Error Handling

The exchange can fail for several reasons:

1. **Token not supported**: Token not registered in x/usdoracle
2. **Insufficient balance**: User doesn't have enough tokens
3. **Price deviation**: Oracle price deviates more than 2% from TWAP
4. **Daily limit exceeded**: User exceeded daily limit
5. **Exchange disabled**: Module is disabled by governance
6. **Invalid amount**: Below minimum or above maximum

All errors are:
- Returned in the immediate API response with `status = FAILED` (HTTP 500) if broadcast fails; otherwise they surface when the tracker resolves the transaction
- Recorded in the database with `status: FAILED`
- Include detailed `error_msg` for debugging

## Configuration Requirements

For the exchange module to work properly, the following must be configured:

1. **x/usdoracle module**:
   - Supported tokens must be registered
   - Price sources must be configured (Yahoo Finance v8, CoinGecko, etc.)
   - Token decimals and symbols must be set

2. **x/exchange module**:
   - Module must be enabled (`enabled: true`)
   - Exchange parameters must be set (limits, fees, etc.)
   - Treasury addresses or community pool configured

3. **Database**:
   - Transactions table must exist (via migrations)
   - Database connection configured in `config.yaml`

## Testing

Use the provided test script to verify exchange functionality:

```bash
./server/scripts/test_exchange.sh
```

The script will:
1. Authenticate a test user
2. Check x/usdoracle configuration
3. Check x/exchange module status
4. Attempt a test exchange
5. Verify transaction on blockchain

## API Documentation

See `server/API_DOCUMENTATION.md` for detailed API endpoint documentation.

## Future Enhancements

Potential improvements for the exchange module:

1. **Multi-source price aggregation**: Average prices from multiple oracles
2. **Dynamic fee calculation**: Adjust fees based on volatility
3. **Token whitelisting UI**: Admin panel for managing supported tokens
4. **Exchange rate caching**: Cache rates for a short period to reduce oracle queries
5. **Advanced slippage protection**: Multi-step validation
6. **Exchange history API**: Query user's exchange history
7. **Exchange analytics**: Volume, fees collected, popular tokens
8. **Liquidity pools integration**: Use DEX pools as secondary price source

## Dependencies

- **blockchain**: Cosmos SDK client for transaction signing and broadcasting
- **auth**: User authentication and wallet management
- **transactions**: Database recording of all operations
- **x/exchange**: Blockchain module for token exchange
- **x/usdoracle**: Price feed provider

## Integration Points

1. **Router** (`server/api/router.go`): Registers `/api/exchange/tokens` endpoint
2. **Main** (`server/main.go`): Initializes ExchangeService
3. **Database**: Uses shared `transactions` table
4. **Blockchain**: Interacts with x/exchange and x/usdoracle modules


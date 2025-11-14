# Assets Trading Module

This module provides functionality for trading assets from the `x/assets` blockchain module.

## Features

- **Ensure Asset**: Create or ensure an asset exists on the blockchain
- **Buy Asset**: Purchase assets using payment denoms (NDOLLAR, unuah, or factory denoms)
- **Sell Asset**: Sell assets and receive NDOLLAR in return
- **Margin Trading**: Open and close leveraged positions via the `x/leverage` module

All endpoints broadcast Cosmos transactions synchronously but now return immediately with `status: "PENDING"` and the `tx_hash`. A background tracker resolves final outcomes (`SUCCESS` / `FAILED`).

## API Endpoints

### POST /api/assets/ensure

Ensure an asset exists, creating it if necessary.

**Request:**
```json
{
  "symbol": "GOLD"
}
```

**Response:**
```json
{
  "tx_hash": "ABC123...",
  "status": "PENDING",
  "message": "Asset ensure broadcast, awaiting confirmation"
}
```

### POST /api/assets/buy

Buy an asset using payment denom.

**Request (new format):**
```json
{
  "symbol": "GOLD",
  "denom": "NDOLLAR",
  "amount": "1000000"
}
```

**Request (deprecated format):**
```json
{
  "symbol": "GOLD",
  "amount_ndollar": "1000000"
}
```

**Response:**
```json
{
  "tx_hash": "ABC123...",
  "status": "PENDING",
  "base_amount": "",
  "message": "Asset purchase broadcast, awaiting confirmation"
}
```

### POST /api/assets/sell

Sell an asset and receive NDOLLAR.

**Request:**
```json
{
  "symbol": "GOLD",
  "base_amount": "0.5"
}
```

**Response:**
```json
{
  "tx_hash": "ABC123...",
  "status": "PENDING",
  "payout_ndollar": "",
  "message": "Asset sale broadcast, awaiting confirmation"
}
```

### POST /api/assets/margin/open

Open a leveraged (margin) position on an underlying asset.

**Request:**
```json
{
  "symbol": "GOLD",
  "side": "long",
  "quote_amount": "1000000",
  "leverage": "3"
}
```

**Response:**
```json
{
  "tx_hash": "ABC123...",
  "status": "PENDING",
  "position_id": "",
  "message": "Margin position opening broadcast, awaiting confirmation"
}
```

### POST /api/assets/margin/close

Close an existing leveraged position by its on-chain identifier.

**Request:**
```json
{
  "position_id": "42"
}
```

**Response:**
```json
{
  "tx_hash": "DEF456...",
  "status": "PENDING",
  "pnl": "",
  "message": "Margin position closure broadcast, awaiting confirmation"
}
```

## Authentication

All endpoints require authentication via Bearer token in the Authorization header:

```
Authorization: Bearer <token>
```

## Payment Denom Selection

When buying assets, if `denom` is not provided, the system automatically selects:
1. `undollar` (if balance > 0)
2. `unuah` (if balance > 0)
3. Falls back to `undollar` (transaction will fail if insufficient funds)

## Architecture

- `models.go`: Request/response models
- `service.go`: Service with business logic and wallet management
- `ensure.go`: Ensure asset implementation
- `buy.go`: Buy asset implementation
- `sell.go`: Sell asset implementation
- `handlers.go`: HTTP handlers for API endpoints

## Integration with Blockchain

The module uses the `blockchain.Client` to:
- Sign transactions with user's private key
- Broadcast transactions to the blockchain
- Record the transaction as `PENDING` in the database and enqueue it for background tracking
- Update application state once the tracker resolves the final status

For confirmed results, clients should query `GET /api/tx/<tx_hash>` or the transactions API (if exposed) to retrieve events and final status.


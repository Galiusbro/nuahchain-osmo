# Assets Trading Module

This module provides functionality for trading assets from the `x/assets` blockchain module.

## Features

- **Ensure Asset**: Create or ensure an asset exists on the blockchain
- **Buy Asset**: Purchase assets using payment denoms (NDOLLAR, unuah, or factory denoms)
- **Sell Asset**: Sell assets and receive NDOLLAR in return

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
  "success": true,
  "message": "Asset ensure initiated"
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
  "base_amount": "0.5",
  "success": true,
  "message": "Asset purchase initiated"
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
  "payout_ndollar": "1000000",
  "success": true,
  "message": "Asset sale initiated"
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
- Extract transaction results and events


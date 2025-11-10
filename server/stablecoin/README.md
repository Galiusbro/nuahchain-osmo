# Stablecoin Module API

## Overview

The stablecoin module provides endpoints for converting between `unuah` and `NDOLLAR` at a 1:1 ratio. This module handles the lifecycle of the NDOLLAR stablecoin, including minting and burning operations.

## File Structure

```
server/stablecoin/
├── models.go      - Request/response models
├── service.go     - Service interface and initialization
├── buy.go         - Buy NDOLLAR implementation (unuah → NDOLLAR)
├── sell.go        - Sell NDOLLAR implementation (NDOLLAR → unuah)
├── handlers.go    - HTTP request handlers
└── README.md      - This file
```

## Endpoints

### POST `/api/stablecoin/buy-ndollar`

Converts `unuah` to `NDOLLAR` at 1:1 ratio.

**Authentication:** Required (JWT token in `Authorization` header)

**Request Body:**
```json
{
  "amount": "1000000"  // Amount of unuah to convert (micro-units)
}
```

**Response (Success):**
```json
{
  "success": true,
  "tx_hash": "ABC123...",
  "ndollar_amount": "1000000",
  "ndollar_denom": "factory/nuah1.../ndollar",
  "error": ""
}
```

**Response (Failure):**
```json
{
  "success": false,
  "tx_hash": "ABC123...",
  "ndollar_amount": "",
  "ndollar_denom": "",
  "error": "transaction failed: insufficient funds"
}
```

**Example (curl):**
```bash
curl -X POST http://localhost:8080/api/stablecoin/buy-ndollar \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"amount": "1000000"}'
```

### POST `/api/stablecoin/sell-ndollar`

Converts `NDOLLAR` back to `unuah` at 1:1 ratio.

**Authentication:** Required (JWT token in `Authorization` header)

**Request Body:**
```json
{
  "amount": "1000000"  // Amount of NDOLLAR to convert (micro-units)
}
```

**Response (Success):**
```json
{
  "success": true,
  "tx_hash": "ABC123...",
  "unuah_amount": "1000000",
  "error": ""
}
```

**Response (Failure):**
```json
{
  "success": false,
  "tx_hash": "ABC123...",
  "unuah_amount": "",
  "error": "transaction failed: insufficient NDOLLAR balance"
}
```

**Example (curl):**
```bash
curl -X POST http://localhost:8080/api/stablecoin/sell-ndollar \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"amount": "1000000"}'
```

## Architecture

### Blockchain Layer
- **Module:** `x/stablecoin`
- **Messages:** `MsgBuyNDollar`, `MsgSellNDollar`
- **Keeper:** Handles conversion, minting, burning, and statistics

### Server Layer
- **Service:** `server/stablecoin/service.go` - Business logic
- **Blockchain Client:** `server/blockchain/stablecoin.go` - Blockchain interactions
- **Handlers:** `server/stablecoin/handlers.go` - HTTP request/response handling

## Implementation Details

### Buy NDOLLAR Process
1. User sends `unuah` amount
2. Server validates request and gets user's wallet
3. Creates `MsgBuyNDollar` transaction
4. Blockchain transfers `unuah` from user to `stablecoin` module
5. Blockchain mints equivalent `NDOLLAR` to module
6. Blockchain sends `NDOLLAR` from module to user
7. Statistics are recorded (total minted)

### Sell NDOLLAR Process
1. User sends `NDOLLAR` amount
2. Server validates request and gets user's wallet
3. Creates `MsgSellNDollar` transaction
4. Blockchain transfers `NDOLLAR` from user to `stablecoin` module
5. Blockchain burns `NDOLLAR` from module
6. Blockchain sends equivalent `unuah` from module to user
7. Statistics are recorded (total burned)

## Conversion Rate

- **Fixed Rate:** 1 `unuah` = 1 `NDOLLAR` (1:1)
- **Decimal Places:** Both tokens use 6 decimal places (micro-units)
- **Example:** 1,000,000 unuah = 1.0 NUAH = 1.0 NDOLLAR

## Security

- **Authentication:** JWT-based authentication required
- **Private Keys:** Encrypted at rest using `AUTH_MASTER_KEY`
- **Transaction Signing:** Server-side signing (not exposed to client)
- **Blockchain Validation:** All operations validated by blockchain consensus

## Error Handling

- **Invalid Amount:** Returns 400 Bad Request
- **Insufficient Balance:** Returns blockchain error in response
- **Authentication Failed:** Returns 401 Unauthorized
- **Transaction Failures:** Captured in response with `success: false`

## Related Endpoints

- `/api/assets/buy` - Buy external assets (GOLD, SILVER) with NDOLLAR
- `/api/assets/sell` - Sell assets for NDOLLAR
- `/api/tokens/buy` - Buy user tokens with NDOLLAR

## Statistics

The `x/stablecoin` module tracks:
- Total NDOLLAR minted
- Total NDOLLAR burned
- Net supply (minted - burned)

Access statistics via blockchain query (not yet exposed in REST API).

# API Documentation

Complete API reference for frontend integration.

**Base URL:** `http://localhost:8080` (development)
**Content-Type:** `application/json`
**Authentication:** Bearer token in `Authorization` header

---

## Table of Contents

1. [Authentication](#authentication)
2. [Health Checks](#health-checks)
3. [User Profile](#user-profile)
4. [Balances](#balances)
5. [Tokens](#tokens)
6. [Marketplace](#marketplace)
7. [Quotes](#quotes)
8. [Assets](#assets)
9. [Stablecoin](#stablecoin)
10. [Exchange](#exchange)
11. [Transactions](#transactions)
12. [Admin](#admin)
13. [WebSocket](#websocket)
14. [Error Handling](#error-handling)

---

## Authentication

### Register (Web)

**POST** `/api/auth/register`

Register a new user with email and password.

**Request:**
```json
{
  "email": "user@example.com",
  "username": "username",
  "password": "password123"
}
```

**Response:** `201 Created`
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "username",
    "created_at": "2025-01-01T00:00:00Z",
    "is_active": true
  },
  "wallet": {
    "id": 1,
    "user_id": 1,
    "address": "nuah1...",
    "created_at": "2025-01-01T00:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "message": "Registration successful"
}
```

---

### Login (Web)

**POST** `/api/auth/login`

Login with email and password.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:** `200 OK`
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "username",
    "last_login_at": "2025-01-01T00:00:00Z"
  },
  "wallet": {
    "id": 1,
    "address": "nuah1...",
    "created_at": "2025-01-01T00:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "message": "Login successful"
}
```

---

### Telegram Authentication

**POST** `/api/auth/telegram`

Authenticate via Telegram.

**Request:**
```json
{
  "id": 123456789,
  "first_name": "John",
  "last_name": "Doe",
  "username": "johndoe",
  "auth_date": 1234567890,
  "hash": "abc123..."
}
```

**Response:** `200 OK` (same as Register)

---

### Refresh Token

**POST** `/api/auth/refresh`

Refresh access token using refresh token.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response:** `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "message": "Token refreshed successfully"
}
```

---

### Logout

**POST** `/api/auth/logout`

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "message": "Logged out successfully"
}
```

---

### Logout All

**POST** `/api/auth/logout-all`

Logout from all devices.

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "message": "Logged out from all devices successfully"
}
```

---

### Get Current User

**GET** `/api/auth/me`

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "username",
    "image_url": "/uploads/images/user_1_123.jpg",
    "created_at": "2025-01-01T00:00:00Z",
    "last_login_at": "2025-01-01T00:00:00Z"
  },
  "wallet": {
    "id": 1,
    "address": "nuah1..."
  }
}
```

---

### Get Sessions

**GET** `/api/auth/sessions`

Get all active sessions for the current user.

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "sessions": [
    {
      "id": 1,
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2025-01-01T00:00:00Z",
      "last_used_at": "2025-01-01T00:00:00Z",
      "expires_at": "2025-01-08T00:00:00Z",
      "is_current": true
    }
  ],
  "total": 1
}
```

---

### Forgot Password

**POST** `/api/auth/web/forgot-password`

**Request:**
```json
{
  "email": "user@example.com"
}
```

**Response:** `200 OK`
```json
{
  "message": "If an account with that email exists, a password reset link has been sent",
  "token": "reset_token_here" // Only in development
}
```

---

### Reset Password

**POST** `/api/auth/web/reset-password`

**Request:**
```json
{
  "token": "reset_token_here",
  "new_password": "newpassword123"
}
```

**Response:** `200 OK`
```json
{
  "message": "Password has been reset successfully"
}
```

---

## Health Checks

### Health Check

**GET** `/health`

**Response:** `200 OK`
```json
{
  "status": "ok",
  "timestamp": "2025-01-01T00:00:00Z"
}
```

---

### Database Health Check

**GET** `/health/db`

**Response:** `200 OK`
```json
{
  "status": "ok",
  "database": "connected",
  "timestamp": "2025-01-01T00:00:00Z"
}
```

---

## User Profile

### Get User Profile

**GET** `/api/users/me`

Get full user profile with statistics and wallet information.

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "username",
    "image_url": "/uploads/images/user_1_123.jpg",
    "created_at": "2025-01-01T00:00:00Z",
    "last_login_at": "2025-01-01T00:00:00Z"
  },
  "wallet": {
    "address": "nuah1...",
    "created_at": "2025-01-01T00:00:00Z",
    "balances": [
      {
        "denom": "unuah",
        "amount": "1000000000"
      }
    ]
  },
  "tokens_count": 5,
  "active_sessions": 2,
  "assets_owned": 3,
  "stats": {
    "total_transactions": 100,
    "successful_transactions": 95,
    "failed_transactions": 5,
    "pending_transactions": 0,
    "transactions_by_type": {
      "TOKEN_CREATE": 5,
      "TOKEN_BUY": 30,
      "TOKEN_SELL": 20,
      "ASSET_BUY": 25,
      "ASSET_SELL": 15,
      "ASSET_MARGIN_OPEN": 3,
      "ASSET_MARGIN_CLOSE": 2
    },
    "last_transaction_at": "2025-01-01T12:00:00Z",
    "tokens_created": 5,
    "margin_positions_open": 1
  }
}
```

---

### Get User Info Summary

**GET** `/api/users/me/info`

Get lightweight user profile summary.

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "id": 1,
  "username": "username",
  "email": "user@example.com",
  "image_url": "/uploads/images/user_1_123.jpg",
  "wallet_address": "nuah1...",
  "total_transactions": 100,
  "tokens_created": 5,
  "created_at": "2025-01-01T00:00:00Z"
}
```

---

### Get User Tokens

**GET** `/api/users/me/tokens`

Get list of tokens owned by the user.

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "tokens": [
    {
      "denom": "factory/nuah1.../token",
      "amount": "1000000",
      "name": "My Token",
      "symbol": "MTK",
      "image": "https://...",
      "decimals": 6
    }
  ],
  "count": 1
}
```

---

### Upload User Image

**POST** `/api/users/me/upload-image`

Upload user profile image.

**Headers:** `Authorization: Bearer <token>`
**Content-Type:** `multipart/form-data`

**Form Data:**
- `image`: File (max 10MB, supported: jpg, jpeg, png, gif, webp)

**Response:** `200 OK`
```json
{
  "message": "Image uploaded successfully",
  "image_url": "/uploads/images/user_1_1234567890.jpg"
}
```

---

### Update Username

**PATCH** `/api/users/username` or **PUT** `/api/users/username`

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "username": "newusername"
}
```

**Response:** `200 OK`
```json
{
  "message": "Username updated successfully",
  "username": "newusername"
}
```

---

## Balances

### Get Balances from Database

**GET** `/api/users/balances-db`

Get user balances from database (fast, potentially slightly stale).

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
- `denom` (optional): Filter by specific denom
- `tokenMint` (optional): Alias for `denom` (for compatibility)

**Response:** `200 OK`
```json
{
  "balances": [
    {
      "denom": "unuah",
      "amount": "1000000000"
    },
    {
      "denom": "factory/nuah1.../token",
      "amount": "5000000"
    }
  ],
  "count": 2,
  "source": "database"
}
```

---

### Get Balances from Blockchain

**GET** `/api/users/balances`

Get user balances directly from blockchain (slower, always fresh).

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "balances": [
    {
      "denom": "unuah",
      "amount": "1000000000"
    }
  ],
  "count": 1,
  "source": "blockchain"
}
```

---

### Sync Balances

**POST** `/api/users/balances/sync`

Synchronize balances from blockchain to database.

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "message": "Balances synchronized successfully"
}
```

---

### Get Balance History

**GET** `/api/users/balances/history`

Get balance change history.

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
- `denom` (optional): Filter by specific denom
- `limit` (optional): Number of records (default: 100)

**Response:** `200 OK`
```json
{
  "history": [
    {
      "id": 1,
      "user_id": 1,
      "address": "nuah1...",
      "denom": "unuah",
      "amount_before": "1000000000",
      "amount_after": "999000000",
      "amount_delta": "-1000000",
      "tx_hash": "abc123...",
      "height": 12345,
      "event_type": "coin_spent",
      "created_at": "2025-01-01T00:00:00Z"
    }
  ],
  "count": 1
}
```

---

### Balance WebSocket

**GET** `/api/users/balances/ws` (WebSocket)

Real-time balance updates via WebSocket.

**Headers:** `Authorization: Bearer <token>`

**Connection:** WebSocket upgrade

**Initial Message:**
```json
{
  "type": "connected",
  "user_id": 1,
  "message": "Connected to balance updates"
}
```

**Subscribe to specific denoms:**
```json
{
  "action": "subscribe",
  "denoms": ["unuah", "factory/nuah1.../token"]
}
```

**Unsubscribe:**
```json
{
  "action": "unsubscribe"
}
```

**Balance Update Message:**
```json
{
  "type": "balance_update",
  "user_id": 1,
  "address": "nuah1...",
  "denom": "unuah",
  "amount": "1000000000",
  "tx_hash": "abc123...",
  "height": 12345,
  "timestamp": "2025-01-01T00:00:00Z"
}
```

---

## Tokens

### Create Token

**POST** `/api/tokens/create`

Create a new user token.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "name": "My Token",
  "symbol": "MTK",
  "image": "https://example.com/image.png",
  "description": "Token description"
}
```

**Response:** `202 Accepted` (or `200 OK` if immediate success)
```json
{
  "denom": "factory/nuah1.../mtk",
  "tx_hash": "abc123...",
  "status": "PENDING",
  "message": "Token creation broadcast, awaiting confirmation"
}
```

**Status Values:**
- `PENDING`: Transaction broadcast, awaiting confirmation
- `SUCCESS`: Transaction confirmed and successful
- `FAILED`: Transaction failed

---

### Buy Token

**POST** `/api/tokens/buy`

Buy tokens from bonding curve.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "denom": "factory/nuah1.../mtk",
  "payment_amount": "1000000",
  "payment_denom": "unuah",
  "min_tokens_out": "500000"
}
```

**Response:** `202 Accepted`
```json
{
  "tx_hash": "abc123...",
  "tokens_out": "550000",
  "price_paid": "1.818",
  "status": "PENDING",
  "message": "Token purchase broadcast, awaiting confirmation"
}
```

---

### Sell Token

**POST** `/api/tokens/sell`

Sell tokens to bonding curve.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "denom": "factory/nuah1.../mtk",
  "token_amount": "500000",
  "payment_denom": "unuah",
  "min_payment_out": "800000"
}
```

**Response:** `202 Accepted`
```json
{
  "tx_hash": "abc123...",
  "payment_out": "900000",
  "price_received": "1.8",
  "status": "PENDING",
  "message": "Token sale broadcast, awaiting confirmation"
}
```

---

## Marketplace

### Get Marketplace Tokens

**GET** `/api/tokens/market`

Get list of all tokens in marketplace.

**Query Parameters:**
- `limit` (optional): Number of tokens (default: 100)
- `offset` (optional): Pagination offset (default: 0)

**Response:** `200 OK`
```json
{
  "tokens": [
    {
      "denom": "factory/nuah1.../mtk",
      "name": "My Token",
      "symbol": "MTK",
      "image": "https://...",
      "description": "Token description",
      "creator": "nuah1...",
      "current_price": "1.5",
      "tokens_sold": "1000000",
      "curve_completed": false,
      "decimals": 6,
      "stats": {
        "total_supply": "10000000",
        "current_price": "1.5",
        "tokens_sold": "1000000"
      }
    }
  ],
  "count": 1,
  "limit": 100,
  "offset": 0
}
```

---

### Search Tokens

**GET** `/api/tokens/search`

Search tokens by name, symbol, or denom.

**Query Parameters:**
- `query` (required): Search query
- `limit` (optional): Number of results (default: 50)
- `offset` (optional): Pagination offset (default: 0)

**Response:** `200 OK`
```json
{
  "tokens": [...],
  "count": 5,
  "query": "MTK",
  "limit": 50,
  "offset": 0
}
```

---

### Get Token Details

**GET** `/api/tokens/{denom}/details`

Get detailed information about a specific token.

**Note:** Denom can contain slashes (e.g., `factory/creator/symbol`). URL encode if needed.

**Response:** `200 OK`
```json
{
  "denom": "factory/nuah1.../mtk",
  "name": "My Token",
  "symbol": "MTK",
  "image": "https://...",
  "description": "Token description",
  "creator": "nuah1...",
  "decimals": 6,
  "stats": {
    "total_supply": "10000000",
    "current_price": "1.5",
    "tokens_sold": "1000000",
    "volume_24h": "5000000"
  }
}
```

---

## Quotes

### Get Trade Quote

**GET** `/api/quote/trade`

Get quote for buying or selling tokens on bonding curve.

**Query Parameters:**
- `denom` (required): Token denom
- `operation` (required): `buy` or `sell`
- `amount` (required): Payment amount (buy) or token amount (sell)
- `payment_denom` (optional): Payment currency (default: `unuah`)

**Response:** `200 OK`
```json
{
  "denom": "factory/nuah1.../mtk",
  "operation": "buy",
  "input_amount": "1000000",
  "input_denom": "unuah",
  "output_amount": "550000",
  "output_denom": "factory/nuah1.../mtk",
  "price": "1.818",
  "price_impact": "0.5",
  "fee": "1000",
  "min_output": "500000"
}
```

---

### Get Swap Quote

**GET** `/api/quote/swap`

Get quote for swapping tokens via exchange module.

**Query Parameters:**
- `token_in` (required): Input token denom (e.g., `ueth`, `ubtc`)
- `amount_in` (required): Amount of input token

**Response:** `200 OK`
```json
{
  "token_in": "ueth",
  "amount_in": "1000000",
  "token_out": "unuah",
  "amount_out": "3500000000",
  "exchange_rate": "3500",
  "fee": "35000",
  "min_output": "3400000000"
}
```

**Error Response:** `400 Bad Request`
```json
{
  "error": "Failed to get swap quote",
  "message": "Exchange rate not found for token: ueth. Supported tokens: [ueth, ubtc]. Exchange rates need to be updated via UpdateExchangeRate.",
  "token_in": "ueth"
}
```

---

### Get Supported Tokens

**GET** `/api/quote/supported-tokens`

Get list of tokens supported for exchange.

**Response:** `200 OK`
```json
{
  "supported_tokens": ["ueth", "ubtc", "uusdc"],
  "available_rates": ["ueth", "ubtc"]
}
```

---

## Assets

### Ensure Asset

**POST** `/api/assets/ensure`

Ensure an asset exists (create if necessary).

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "symbol": "GOLD"
}
```

**Response:** `202 Accepted`
```json
{
  "tx_hash": "abc123...",
  "status": "PENDING",
  "message": "Asset ensure broadcast, awaiting confirmation"
}
```

---

### Buy Asset

**POST** `/api/assets/buy`

Buy an asset using payment denom.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "symbol": "GOLD",
  "denom": "unuah",
  "amount": "1000000"
}
```

**Or (deprecated):**
```json
{
  "symbol": "GOLD",
  "amount_ndollar": "1000000"
}
```

**Response:** `202 Accepted`
```json
{
  "tx_hash": "abc123...",
  "base_amount": "0.5",
  "status": "PENDING",
  "message": "Asset purchase broadcast, awaiting confirmation"
}
```

---

### Sell Asset

**POST** `/api/assets/sell`

Sell an asset and receive NDOLLAR.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "symbol": "GOLD",
  "base_amount": "0.5"
}
```

**Response:** `202 Accepted`
```json
{
  "tx_hash": "abc123...",
  "payout_ndollar": "950000",
  "status": "PENDING",
  "message": "Asset sale broadcast, awaiting confirmation"
}
```

---

### Open Margin Position

**POST** `/api/assets/margin/open`

Open a leveraged (margin) position.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "symbol": "GOLD",
  "side": "long",
  "quote_amount": "1000000",
  "leverage": "3"
}
```

**Response:** `202 Accepted`
```json
{
  "tx_hash": "abc123...",
  "position_id": "123",
  "base_quantity": "1.5",
  "entry_price": "666666.67",
  "leverage": "3",
  "status": "PENDING",
  "message": "Margin position opened, awaiting confirmation"
}
```

---

### Close Margin Position

**POST** `/api/assets/margin/close`

Close a leveraged position.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "position_id": "123"
}
```

**Response:** `202 Accepted`
```json
{
  "tx_hash": "abc123...",
  "pnl": "50000",
  "status": "PENDING",
  "message": "Margin position closed, awaiting confirmation"
}
```

---

## Stablecoin

### Buy NDOLLAR

**POST** `/api/stablecoin/buy-ndollar`

Buy NDOLLAR stablecoin with unuah (1:1 exchange).

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "amount": "1000000"
}
```

**Response:** `202 Accepted`
```json
{
  "status": "PENDING",
  "tx_hash": "abc123...",
  "ndollar_amount": "1000000",
  "ndollar_denom": "factory/nuah1.../ndollar"
}
```

---

### Sell NDOLLAR

**POST** `/api/stablecoin/sell-ndollar`

Sell NDOLLAR back to unuah (1:1 exchange).

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "amount": "1000000"
}
```

**Response:** `202 Accepted`
```json
{
  "status": "PENDING",
  "tx_hash": "abc123...",
  "unuah_amount": "1000000"
}
```

---

## Exchange

### Exchange Tokens

**POST** `/api/exchange/tokens`

Exchange tokens for unuah via exchange module.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "token_denom": "ueth",
  "amount": "1000000",
  "min_output": "3400000000"
}
```

**Response:** `202 Accepted`
```json
{
  "status": "PENDING",
  "tx_hash": "abc123...",
  "unuah_out": "3500000000"
}
```

---

## Transactions

### Get Transaction Status

**GET** `/api/tx/{hash}`

Get status of any transaction by hash.

**Response:** `200 OK`
```json
{
  "tx_hash": "abc123...",
  "found": true,
  "success": true,
  "code": 0,
  "codespace": "",
  "height": 12345,
  "gas_used": 100000,
  "gas_wanted": 150000,
  "log": "",
  "error": "",
  "events": [
    {
      "type": "message",
      "attributes": [
        {"key": "action", "value": "buy_token"}
      ]
    }
  ]
}
```

**Error Response:** `400 Bad Request`
```json
{
  "tx_hash": "abc123...",
  "found": false,
  "success": false,
  "error": "Transaction not found"
}
```

---

## Admin

### Get Recent Transactions

**GET** `/api/admin/transactions`

Get recent blockchain transactions (admin only).

**Query Parameters:**
- `limit` (optional): Number of transactions (default: 100)

**Response:** `200 OK`
```json
{
  "transactions": [
    {
      "tx_hash": "abc123...",
      "height": 12345,
      "success": true,
      "code": 0,
      "log": "",
      "time": "2025-01-01T00:00:00Z",
      "events": ["message", "transfer"]
    }
  ],
  "count": 1
}
```

---

### Get Transaction by Hash

**GET** `/api/admin/transactions?hash={hash}`

Get transaction details by hash.

**Response:** `200 OK`
```json
{
  "tx_hash": "abc123...",
  "height": 12345,
  "code": 0,
  "success": true,
  "log": "",
  "events": {...},
  "messages": [...],
  "gas_used": 100000,
  "gas_wanted": 150000
}
```

---

### Get Stats

**GET** `/api/admin/stats`

Get monitoring statistics.

**Response:** `200 OK`
```json
{
  "total_transactions": 1000,
  "successful_transactions": 950,
  "failed_transactions": 50,
  "pending_transactions": 0,
  "last_block_height": 12345,
  "uptime_seconds": 86400
}
```

---

### Admin WebSocket

**GET** `/api/admin/transactions/ws` (WebSocket)

Real-time transaction updates for admin panel.

**Connection:** WebSocket upgrade

**Message Format:**
```json
{
  "tx_hash": "abc123...",
  "height": 12345,
  "success": true,
  "code": 0,
  "log": "",
  "time": "2025-01-01T00:00:00Z",
  "events": ["message", "transfer"]
}
```

---

## Error Handling

### Standard Error Response

All endpoints return consistent error responses:

**400 Bad Request:**
```json
{
  "error": "Invalid request",
  "message": "Field 'denom' is required"
}
```

**401 Unauthorized:**
```json
{
  "error": "Unauthorized",
  "message": "Invalid or expired token"
}
```

**404 Not Found:**
```json
{
  "error": "Not found",
  "message": "Token not found"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Internal server error",
  "message": "Failed to process request"
}
```

---

## Transaction Status

All transaction endpoints return a `status` field with one of these values:

- **`PENDING`**: Transaction broadcast, awaiting blockchain confirmation
- **`SUCCESS`**: Transaction confirmed and successful
- **`FAILED`**: Transaction failed (check `error` field for details)

After receiving a `PENDING` status, you can:
1. Poll `/api/tx/{hash}` to check status
2. Use WebSocket to receive real-time updates
3. Wait for the transaction to be confirmed (usually within a few seconds)

---

## Notes

### Authentication

All protected endpoints require a Bearer token in the `Authorization` header:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### Amounts

All amounts are returned as strings to preserve precision. Use `BigNumber` or similar libraries for calculations.

### Denoms

Token denoms can contain slashes (e.g., `factory/creator/symbol`). URL encode when using in paths:
- `/api/tokens/factory%2Fcreator%2Fsymbol/details`

### Pagination

Endpoints supporting pagination use `limit` and `offset` query parameters.

### WebSocket

WebSocket connections require authentication via `Authorization` header during the initial HTTP upgrade request.

### Image Uploads

User images are served statically at `/uploads/images/{filename}`. The full URL is returned in responses.

---

## Base Denoms

Common base denoms:
- `unuah`: Base currency (1 unuah = 10^-6 N$)
- `undollar`: NDOLLAR stablecoin (1 undollar = 10^-6 NDOLLAR)
- `factory/{creator}/{symbol}`: User-created tokens
- `asset/{symbol}`: Assets (e.g., `asset/GOLD`)

---

## Rate Limits

Currently no rate limits are enforced. In production, consider implementing rate limiting for:
- Authentication endpoints
- Transaction endpoints
- Quote endpoints

---

## Versioning

API versioning is not currently implemented. All endpoints are under `/api/` prefix.

---

## Support

For issues or questions, contact the backend team.


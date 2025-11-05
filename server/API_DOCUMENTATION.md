# REST API Documentation

This document describes the REST API endpoints for the NuahChain token trading server. All endpoints return JSON responses.

## Base URL

```
http://localhost:8080
```

## Authentication

Most endpoints require authentication using a Bearer token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

## Content Type

All requests should include:
```
Content-Type: application/json
```

## Response Format

All responses are in JSON format. Success responses include relevant data, while error responses include an error message.

---

## Health Check Endpoints

### GET /health

Check if the server is running.

**Request:**
```http
GET /health
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "timestamp": "2025-11-05T23:35:25.254978+07:00"
}
```

---

### GET /health/db

Check database connection health.

**Request:**
```http
GET /health/db
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "timestamp": "2025-11-05T23:35:25.254978+07:00"
}
```

**Response (503 Service Unavailable) - Database Error:**
```json
{
  "status": "error",
  "timestamp": "2025-11-05T23:35:25.254978+07:00",
  "error": "connection failed"
}
```

---

## Authentication Endpoints

### POST /api/auth/register

Register a new user account with email and password.

**Request:**
```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "username": "username",  // optional
  "password": "secure_password"
}
```

**Request Body:**
- `email` (string, required): User's email address
- `username` (string, optional): User's username
- `password` (string, required): User's password

**Response (201 Created):**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "username",
    "created_at": "2025-11-05T23:35:25Z",
    "updated_at": "2025-11-05T23:35:25Z",
    "is_active": true
  },
  "wallet": {
    "id": 1,
    "user_id": 1,
    "address": "nuah1x9nc3nse2skvtdqxutrfk3pg6rtckd3qskvc5n",
    "created_at": "2025-11-05T23:35:25Z",
    "updated_at": "2025-11-05T23:35:25Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Registration successful"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body or missing required fields
- `500 Internal Server Error`: Server error during registration

---

### POST /api/auth/login

Login with email and password to get an authentication token.

**Request:**
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "secure_password"
}
```

**Request Body:**
- `email` (string, required): User's email address
- `password` (string, required): User's password

**Response (200 OK):**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "username",
    "created_at": "2025-11-05T23:35:25Z",
    "updated_at": "2025-11-05T23:35:25Z",
    "last_login_at": "2025-11-05T23:35:25Z",
    "is_active": true
  },
  "wallet": {
    "id": 1,
    "user_id": 1,
    "address": "nuah1x9nc3nse2skvtdqxutrfk3pg6rtckd3qskvc5n",
    "created_at": "2025-11-05T23:35:25Z",
    "updated_at": "2025-11-05T23:35:25Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Login successful"
}
```

**Error Responses:**
- `400 Bad Request`: Missing email or password
- `401 Unauthorized`: Invalid credentials
- `500 Internal Server Error`: Server error

---

### POST /api/auth/telegram

Authenticate or register using Telegram credentials.

**Request:**
```http
POST /api/auth/telegram
Content-Type: application/json

{
  "id": 123456789,
  "first_name": "John",
  "last_name": "Doe",
  "username": "johndoe",
  "auth_date": 1699123456,
  "hash": "abc123def456..."
}
```

**Request Body:**
- `id` (int64, required): Telegram user ID
- `first_name` (string, required): User's first name
- `last_name` (string, optional): User's last name
- `username` (string, optional): Telegram username
- `auth_date` (int64, required): Authentication timestamp
- `hash` (string, required): Telegram authentication hash

**Response (200 OK):**
```json
{
  "user": {
    "id": 1,
    "telegram_id": 123456789,
    "telegram_username": "johndoe",
    "created_at": "2025-11-05T23:35:25Z",
    "updated_at": "2025-11-05T23:35:25Z",
    "is_active": true
  },
  "wallet": {
    "id": 1,
    "user_id": 1,
    "address": "nuah1x9nc3nse2skvtdqxutrfk3pg6rtckd3qskvc5n",
    "created_at": "2025-11-05T23:35:25Z",
    "updated_at": "2025-11-05T23:35:25Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Telegram authentication successful"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body
- `500 Internal Server Error`: Server error

---

### GET /api/auth/me

Get current authenticated user information.

**Authentication:** Required

**Request:**
```http
GET /api/auth/me
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "username",
    "created_at": "2025-11-05T23:35:25Z",
    "updated_at": "2025-11-05T23:35:25Z",
    "is_active": true
  },
  "wallet": {
    "id": 1,
    "address": "nuah1x9nc3nse2skvtdqxutrfk3pg6rtckd3qskvc5n"
  }
}
```

**Error Responses:**
- `401 Unauthorized`: Missing or invalid token
- `500 Internal Server Error`: Server error

---
---

## Asset Trading Endpoints

Asset endpoints integrate with the Cosmos `x/assets` module. All operations fetch live prices from the Yahoo Finance v8 API. Symbols are passed verbatim from the frontend; the server normalises and, when necessary, auto-resolves common aliases via Yahoo symbol search (e.g. `GOLD` → `GC=F`, `BTC` → `BTC-USD`). No manual symbol mapping is required.

Every asset endpoint requires authentication (`Authorization: Bearer <token>`).

### POST /api/assets/ensure

Ensure that an asset exists (creates it on-chain if missing).

**Request:**
```http
POST /api/assets/ensure
Authorization: Bearer <token>
Content-Type: application/json

{
  "symbol": "GOLD"
}
```

**Request Body:**
- `symbol` (string, required): Asset symbol as typed by the user.

**Response (200 OK):**
```json
{
  "tx_hash": "952a20f9f5b50ce1bbde925af17975f70b6f6d4d6cb90597d394347cd82fb374",
  "success": true,
  "message": "Asset ensure initiated"
}
```

**Notes:**
- Re-ensuring an existing asset is idempotent and still returns a transaction hash.
- All operations are recorded in the `transactions` table for later audit.

---

### POST /api/assets/buy

Buy an asset using `unuah`, `NDOLLAR`, or the factory denom for NDOLLAR. Pricing is fetched on demand from the oracle.

**Request:**
```http
POST /api/assets/buy
Authorization: Bearer <token>
Content-Type: application/json

{
  "symbol": "GOLD",
  "denom": "unuah",
  "amount": "1000000"
}
```

**Request Body:**
- `symbol` (string, required): Asset symbol.
- `denom` (string, optional): Payment denom. Supports `unuah`, `NDOLLAR`, or `factory/.../ndollar`. If omitted the backend auto-selects based on balances.
- `amount` (string, optional): Payment amount (base units). Required when `denom` is provided.
- `amount_ndollar` (string, optional): Deprecated legacy field; still supported for backwards compatibility.

**Response (200 OK):**
```json
{
  "tx_hash": "3a2da984a4a381ca732a910d2b7e1468947d38437584f1a747fd228b546ff80f",
  "success": true,
  "message": "Asset purchase initiated"
}
```

**Notes:**
- The synchronous response does not include the final `base_amount`. Use `/api/tx/:hash` to inspect events (`assets.asset_bought`) once the transaction is finalised.
- When paying with `unuah` the module mints and burns NDOLLAR internally.

---

### POST /api/assets/sell

Sell a previously purchased asset back for NDOLLAR.

**Request:**
```http
POST /api/assets/sell
Authorization: Bearer <token>
Content-Type: application/json

{
  "symbol": "GOLD",
  "base_amount": "0.1"
}
```

**Request Body:**
- `symbol` (string, required): Asset symbol.
- `base_amount` (string, required): Amount of asset to sell (human-readable units, e.g. `0.1`). The backend converts to micro units internally.

**Response (200 OK):**
```json
{
  "tx_hash": "a1fd3b388b138f96c6f565eebe9b853268cfba8a42b27a02243d9adf683afdf1",
  "success": true,
  "message": "Asset sale initiated"
}
```

**Notes:**
- Detailed payout can be retrieved through `/api/tx/:hash` (event `assets.asset_sold`).
- The sale fails if the user lacks balance or if the oracle cannot refresh the price.

---

### Oracle behaviour & error handling
- Live prices are fetched via Yahoo Finance v8. If a direct request yields HTTP 404/400 or empty data, the backend automatically performs a Yahoo symbol search and retries with the best candidate.
- If no valid price is found after fallback the transaction fails (e.g. `failed to fetch price from API: HTTP error: 404`). The error is saved to the `transactions` table (`status: failed`).
- Symbols are sanitised but otherwise unaltered, allowing the frontend to send arbitrary tickers.

---

## Token Management Endpoints

All token endpoints require authentication.

### POST /api/tokens/create

Create a new token on the blockchain.

**Authentication:** Required

**Request:**
```http
POST /api/tokens/create
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "My Token",
  "symbol": "MTK",
  "image": "https://example.com/token.png",  // optional
  "description": "Token description"  // optional
}
```

**Request Body:**
- `name` (string, required): Token name (e.g., "My Token")
- `symbol` (string, required): Token symbol (e.g., "MTK")
- `image` (string, optional): URL to token image
- `description` (string, optional): Token description

**Response (201 Created):**
```json
{
  "denom": "factory/nuah1x9nc3nse2skvtdqxutrfk3pg6rtckd3qskvc5n/mtk",
  "tx_hash": "f4b6ea890e56d0f96b066c0ce77230b284ada899b6724d8612b955290d58d676",
  "success": true,
  "message": "Token creation initiated"
}
```

**Response Fields:**
- `denom` (string): Token denomination (used for trading)
- `tx_hash` (string): Blockchain transaction hash
- `success` (boolean): Whether the transaction was initiated successfully
- `message` (string): Status message
- `error` (string, optional): Error message if transaction failed

**Error Responses:**
- `400 Bad Request`: Missing name or symbol
- `401 Unauthorized`: Missing or invalid token
- `500 Internal Server Error`: Transaction failed or server error

**Note:** Token creation mints tokens and distributes them according to configured percentages:
- 30% to bonding curve wallet (for trading)
- 10% to platform wallet
- 10% to referral wallet
- 40% to AI CEO wallet
- 10% remaining (if applicable)

---

### POST /api/tokens/buy

Buy tokens from the bonding curve.

**Authentication:** Required

**Request:**
```http
POST /api/tokens/buy
Authorization: Bearer <token>
Content-Type: application/json

{
  "denom": "factory/nuah1x9nc3nse2skvtdqxutrfk3pg6rtckd3qskvc5n/mtk",
  "payment_amount": "10000000",
  "payment_denom": "unuah",  // optional, defaults to unuah
  "min_tokens_out": "20000000"  // optional, slippage protection
}
```

**Request Body:**
- `denom` (string, required): Token denomination to buy (e.g., `factory/creator/symbol`)
- `payment_amount` (string, required): Amount to pay in base units (e.g., `10000000` = 10 tokens if display exponent is 6)
- `payment_denom` (string, optional): Payment currency (`unuah` or `undollar`). If not provided, system auto-selects based on user balance (prefers `undollar`, falls back to `unuah`)
- `min_tokens_out` (string, optional): Minimum tokens to receive (slippage protection)

**Response (200 OK):**
```json
{
  "tx_hash": "4700a4a49d5afc46caca6526bd377a3e88955c5c95f54069d7763c3cfe46b0eb",
  "tokens_out": "24491346",
  "price_paid": "10000000",
  "success": true,
  "message": "Token purchase initiated"
}
```

**Response Fields:**
- `tx_hash` (string): Blockchain transaction hash
- `tokens_out` (string, optional): Actual tokens received (in base units)
- `price_paid` (string, optional): Total price paid
- `success` (boolean): Whether the transaction was initiated successfully
- `message` (string): Status message
- `error` (string, optional): Error message if transaction failed

**Error Responses:**
- `400 Bad Request`: Missing required fields
- `401 Unauthorized`: Missing or invalid token
- `500 Internal Server Error`: Transaction failed or server error

**Note:**
- Amounts are specified in base units (e.g., `10000000` = 10 tokens if display exponent is 6)
- The system automatically selects payment denomination if not specified
- Transaction is initiated asynchronously; check `tx_hash` status on blockchain

---

### POST /api/tokens/sell

Sell tokens back to the bonding curve.

**Authentication:** Required

**Request:**
```http
POST /api/tokens/sell
Authorization: Bearer <token>
Content-Type: application/json

{
  "denom": "factory/nuah1x9nc3nse2skvtdqxutrfk3pg6rtckd3qskvc5n/mtk",
  "token_amount": "5000000",
  "payment_denom": "unuah",  // optional, defaults to unuah
  "min_payment_out": "3000000"  // optional, slippage protection
}
```

**Request Body:**
- `denom` (string, required): Token denomination to sell (e.g., `factory/creator/symbol`)
- `token_amount` (string, required): Amount of tokens to sell in base units (e.g., `5000000` = 5 tokens if display exponent is 6)
- `payment_denom` (string, optional): Payment currency to receive (`unuah` or `undollar`). If not provided, system auto-selects based on user balance (prefers `undollar`, falls back to `unuah`)
- `min_payment_out` (string, optional): Minimum payment to receive (slippage protection)

**Response (200 OK):**
```json
{
  "tx_hash": "3041d3feda2795e35d7fa131ae29839127440323927d79465375b1f329d41df4",
  "payment_out": "3652495",
  "price_received": "3652495",
  "success": true,
  "message": "Token sale initiated"
}
```

**Response Fields:**
- `tx_hash` (string): Blockchain transaction hash
- `payment_out` (string, optional): Actual payment received (in base units)
- `price_received` (string, optional): Price received per token
- `success` (boolean): Whether the transaction was initiated successfully
- `message` (string): Status message
- `error` (string, optional): Error message if transaction failed

**Error Responses:**
- `400 Bad Request`: Missing required fields
- `401 Unauthorized`: Missing or invalid token
- `500 Internal Server Error`: Transaction failed or server error

**Note:**
- Amounts are specified in base units (e.g., `5000000` = 5 tokens if display exponent is 6)
- The system automatically selects payment denomination if not specified
- Transaction is initiated asynchronously; check `tx_hash` status on blockchain

---

## Transaction Status Endpoint

### GET /api/tx/:hash

Get the status of any transaction by its hash. This endpoint queries the blockchain directly using REST API and provides full transaction details including logs and events.

**Authentication:** Not required (public endpoint)

**Request:**
```http
GET /api/tx/4700a4a49d5afc46caca6526bd377a3e88955c5c95f54069d7763c3cfe46b0eb
```

**URL Parameters:**
- `hash` (string, required): Transaction hash (hex string, 64 characters, can be in any case)

**Response (200 OK) - Transaction Found (Success):**
```json
{
  "tx_hash": "4700a4a49d5afc46caca6526bd377a3e88955c5c95f54069d7763c3cfe46b0eb",
  "found": true,
  "success": true,
  "code": 0,
  "codespace": "",
  "height": 45297,
  "gas_used": 227350,
  "gas_wanted": 500000,
  "log": "",
  "events": [
    {
      "type": "transfer",
      "attributes": {
        "recipient": "nuah1x9nc3nse2skvtdqxutrfk3pg6rtckd3qskvc5n",
        "sender": "nuah1fx6f78mh5pa7qq9xkttemu3gx49r9pnvq8nwpd",
        "amount": "200000000unuah"
      }
    }
  ]
}
```

**Response (200 OK) - Transaction Failed:**
```json
{
  "tx_hash": "4700a4a49d5afc46caca6526bd377a3e88955c5c95f54069d7763c3cfe46b0eb",
  "found": true,
  "success": false,
  "code": 5,
  "codespace": "sdk",
  "height": 45297,
  "gas_used": 130091,
  "gas_wanted": 500000,
  "log": "failed to execute message; message index: 0: insufficient funds",
  "error": "failed to execute message; message index: 0: insufficient funds",
  "timestamp": "2025-11-05T23:35:25Z"
}
```

**Response (200 OK) - Transaction Not Found:**
```json
{
  "tx_hash": "4700a4a49d5afc46caca6526bd377a3e88955c5c95f54069d7763c3cfe46b0eb",
  "found": false,
  "success": false,
  "error": "transaction not found"
}
```

**Response Fields:**
- `tx_hash` (string): Transaction hash (normalized to lowercase)
- `found` (boolean): Whether the transaction was found on the blockchain
- `success` (boolean): Whether the transaction succeeded. This checks both `code == 0` AND logs for error messages (even if code is 0, errors in logs will mark success as false)
- `code` (int, optional): Transaction result code (0 = success, non-zero = error)
- `codespace` (string, optional): Error codespace
- `height` (int64, optional): Block height where transaction was included
- `gas_used` (int64, optional): Gas actually used
- `gas_wanted` (int64, optional): Gas requested
- `log` (string, optional): Full transaction log (may contain error details even if code is 0)
- `error` (string, optional): Error message if transaction failed
- `events` (array, optional): Array of transaction events with type and attributes

**Error Responses:**
- `400 Bad Request`: Invalid transaction hash format

**Important Notes:**
- **Works for any transaction type**: This endpoint queries the blockchain directly and works for all transaction types (bank transfers, token operations, staking, governance, etc.)
- **Full log inspection**: Even if `code` is 0, the endpoint checks logs for error messages (like "insufficient funds", "invalid", etc.) and sets `success: false` if errors are found
- **REST API**: Uses blockchain REST API (`http://localhost:26657/tx`) for complete transaction information, including all events and logs
- Transaction hash must be a valid 64-character hexadecimal string
- If transaction is not yet included in a block, `found` will be `false`
- Hash can be in any case (upper, lower, or mixed) - it will be normalized to lowercase
- Use this endpoint to check transaction status after any blockchain operation

---

## Error Handling

All endpoints may return standard HTTP error codes:

- `400 Bad Request`: Invalid request parameters or missing required fields
- `401 Unauthorized`: Missing or invalid authentication token
- `405 Method Not Allowed`: Wrong HTTP method used
- `500 Internal Server Error`: Server error or blockchain transaction failure

Error responses include an error message in the response body or in the `error` field of the JSON response.

---

## Token Amounts and Units

### Base Units vs Display Units

All amounts in API requests and responses are specified in **base units** (exponent 0).

For display purposes, tokens typically have a display exponent of 6:
- Base unit: `10000000` = Display unit: `10` tokens
- Base unit: `5000000` = Display unit: `5` tokens
- Base unit: `1000000` = Display unit: `1` token

### Payment Denominations

Supported payment denominations:
- `unuah`: Base currency (default)
- `undollar`: Alternative payment currency

If `payment_denom` is not specified in buy/sell requests, the system automatically selects based on user balance:
1. First tries `undollar` if balance is sufficient
2. Falls back to `unuah` if `undollar` is insufficient or unavailable

---

## Example Workflow

### 1. Register a new user
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "secure_password"
  }'
```

### 2. Login to get token
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "secure_password"
  }'
```

### 3. Create a token
```bash
curl -X POST http://localhost:8080/api/tokens/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "My Token",
    "symbol": "MTK",
    "description": "My first token"
  }'
```

### 4. Buy tokens
```bash
curl -X POST http://localhost:8080/api/tokens/buy \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "denom": "factory/nuah1x9nc3nse2skvtdqxutrfk3pg6rtckd3qskvc5n/mtk",
    "payment_amount": "10000000"
  }'
```

### 5. Sell tokens
```bash
curl -X POST http://localhost:8080/api/tokens/sell \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "denom": "factory/nuah1x9nc3nse2skvtdqxutrfk3pg6rtckd3qskvc5n/mtk",
    "token_amount": "5000000"
  }'
```

### 6. Check transaction status (works for any transaction)
```bash
curl -X GET http://localhost:8080/api/tx/4700a4a49d5afc46caca6526bd377a3e88955c5c95f54069d7763c3cfe46b0eb
```

---

## Notes for Frontend Developers

1. **Token Storage**: Store the JWT token securely (e.g., in localStorage or secure cookies) after login/registration.

2. **Token Expiration**: JWT tokens expire after a set period. Handle token expiration by prompting users to re-login.

3. **Transaction Status**: After initiating a transaction, use the `tx_hash` to query the blockchain for transaction status. The API returns immediately with the transaction hash, but the transaction may take a few seconds to be included in a block.

4. **Error Handling**: Always check the `success` field in responses. Even if HTTP status is 200, `success: false` indicates the transaction failed.

5. **Amount Formatting**: Convert between base units and display units for user-friendly display:
   - To display: `displayAmount = baseAmount / 1000000`
   - To submit: `baseAmount = displayAmount * 1000000`

6. **Async Operations**: Token creation, buying, and selling are asynchronous operations. The API returns immediately with a transaction hash. Poll the blockchain or wait a few seconds before checking balances.

7. **Payment Denom Selection**: If you don't specify `payment_denom`, the system will automatically select the best available option. You can check user balances first to suggest the preferred denomination.

---

## Support

For issues or questions, please refer to the project documentation or contact the development team.


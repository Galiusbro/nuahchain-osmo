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

### Transaction Status Lifecycle

Most state-changing endpoints (assets, user tokens, stablecoin, exchange) initiate on-chain transactions via `BroadcastTxSync`. The HTTP response is **pessimistic**—it returns immediately with:

- `tx_hash`: the Cosmos transaction hash
- `status`: one of
  - `PENDING` – transaction submitted, waiting for block inclusion
  - `SUCCESS` – transaction confirmed on-chain (typically returned by background polling)
  - `FAILED` – transaction rejected on-chain (code ≠ 0, insufficient funds, etc.)
- Optional contextual fields (e.g. `message`, `unuah_out`)

HTTP status codes follow this state: `202 Accepted` for `PENDING`, `200 OK` for `SUCCESS`, `500 Internal Server Error` for `FAILED`.

A background tracker polls Tendermint/gRPC and updates the server-side database once the final status is known. Clients should poll the dedicated endpoint:

```
GET /api/tx/<tx_hash>
```

The response mirrors on-chain data (code, raw log, events). Scripts/clients can use this to wait for completion, update UI, or surface errors.

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

In addition to spot trading, the server exposes leverage trading through the `x/leverage` module (margin positions backed by NDOLLAR collateral).

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

**Response (202 Accepted):**
```json
{
  "tx_hash": "812ddfd2cdec3f0ec14e92adcf5082d66afe648509eeefe0a8baea1d4fde8fca",
  "status": "PENDING",
  "message": "Asset ensure transaction broadcast, awaiting confirmation"
}
```

**Notes:**
- Re-ensuring an existing asset is idempotent and still returns a transaction hash.
- The operation is logged to the `transactions` table with `status = PENDING`, then updated by the tracker once finalized.
- Call `GET /api/tx/<tx_hash>` to obtain on-chain success/failure details (event `assets.asset_created`).

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

**Response (202 Accepted):**
```json
{
  "tx_hash": "ebcfe2c4ee86b2bb88c33765c9f157b8918dced029b93100fef650251e736392",
  "status": "PENDING",
  "base_amount": "",
  "message": "Asset purchase broadcast, awaiting confirmation"
}
```

**Notes:**
- `base_amount` will be populated once the tracker resolves the transaction; inspect `/api/tx/<tx_hash>` for `assets.asset_bought` events.
- When paying with `unuah` the module mints and burns NDOLLAR internally.
- Database entry moves from `PENDING` → `SUCCESS`/`FAILED` automatically.

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

**Response (202 Accepted):**
```json
{
  "tx_hash": "b379c8c73fe72d6833bc03dca16e1b0aa7097544bcd6aa1a40ec505e28759a74",
  "status": "PENDING",
  "payout_ndollar": "",
  "message": "Asset sale broadcast, awaiting confirmation"
}
```

**Notes:**
- Detailed payout appears in `/api/tx/<tx_hash>` (event `assets.asset_sold`).
- Failures (e.g. insufficient asset balance) are surfaced as `status = FAILED` once the tracker resolves the transaction.

---

### POST /api/assets/margin/open

Open a leveraged (margin) position on an asset using NDOLLAR collateral.

**Request:**
```http
POST /api/assets/margin/open
Authorization: Bearer <token>
Content-Type: application/json

{
  "symbol": "GOLD",
  "side": "long",
  "quote_amount": "1000000",
  "leverage": "2"
}
```

**Request Body:**
- `symbol` (string, required): Asset symbol.
- `side` (string, required): Either `long` or `short` (case-insensitive).
- `quote_amount` (string, required): Margin supplied in micro NDOLLAR.
- `leverage` (string, required): Desired leverage multiplier expressed as a decimal string (e.g., `2`, `3.5`).

**Response (202 Accepted):**
```json
{
  "tx_hash": "426913ce0daed179d57d3c8b3ff73907e7757841e0d1b9db412bebaf0d2af905",
  "status": "PENDING",
  "position_id": "",
  "base_quantity": "",
  "entry_price": "",
  "leverage": "1.0",
  "message": "Margin position opening broadcast, awaiting confirmation"
}
```

**Notes:**
- Ensure the wallet holds sufficient NDOLLAR for the quoted margin; the module borrows/repays automatically.
- Position metadata (ID, base quantity, entry price) becomes available once the tracker resolves the transaction; fetch `/api/tx/<tx_hash>` (event `leverage.position_opened`).
- Transactions are recorded in the `transactions` table under `ASSET_MARGIN_OPEN` and transition from `PENDING` to `SUCCESS`/`FAILED` automatically.

---

### POST /api/assets/margin/close

Close an existing leveraged position, realising PnL.

**Request:**
```http
POST /api/assets/margin/close
Authorization: Bearer <token>
Content-Type: application/json

{
  "position_id": "7"
}
```

**Request Body:**
- `position_id` (string, required): ID returned when the position was opened.

**Response (202 Accepted):**
```json
{
  "tx_hash": "92a5fea273f064184fc8ccb5be223a9d14ec89adc9cdccec7b7ec9a930941969",
  "status": "PENDING",
  "pnl": "",
  "message": "Margin position closure broadcast, awaiting confirmation"
}
```

**Notes:**
- After confirmation the tracker enriches the DB record; fetch `/api/tx/<tx_hash>` for final PnL and events (`leverage.position_closed`).
- Database entry transitions from `PENDING` to `SUCCESS`/`FAILED` automatically based on on-chain result.

---

### Oracle behaviour & error handling
- Live prices are fetched via Yahoo Finance v8. If a direct request yields HTTP 404/400 or empty data, the backend automatically performs a Yahoo symbol search and retries with the best candidate.
- If no valid price is found after fallback the transaction fails (e.g. `failed to fetch price from API: HTTP error: 404`). The error is saved to the `transactions` table (`status: failed`).
- Symbols are sanitised but otherwise unaltered, allowing the frontend to send arbitrary tickers.

---

## Stablecoin Endpoints

These endpoints interact with the `x/stablecoin` module to swap between `unuah` and NDOLLAR at a 1:1 rate. Both require authentication.

### POST /api/stablecoin/buy-ndollar

Convert `unuah` into NDOLLAR.

**Request:**
```http
POST /api/stablecoin/buy-ndollar
Authorization: Bearer <token>
Content-Type: application/json

{
  "amount": "1000000"
}
```

**Request Body:**
- `amount` (string, required): Amount of `unuah` to convert (base units).

**Response (202 Accepted):**
```json
{
  "tx_hash": "cab0fdb536e06fb0773cfe55a8633b60bd015106fb296cf02c62dc7e9abfccb2",
  "status": "PENDING",
  "ndollar_amount": "",
  "ndollar_denom": "",
  "error": ""
}
```

**Notes:**
- The tracker fills `ndollar_amount`/`ndollar_denom` once the transaction confirms (see `/api/tx/<tx_hash>` event `buy_ndollar`).
- Insufficient balances or module constraints will produce `status = FAILED` with `error` populated.

### POST /api/stablecoin/sell-ndollar

Convert NDOLLAR back into `unuah`.

**Request:**
```http
POST /api/stablecoin/sell-ndollar
Authorization: Bearer <token>
Content-Type: application/json

{
  "amount": "500000"
}
```

**Request Body:**
- `amount` (string, required): Amount of NDOLLAR to burn (base units).

**Response (202 Accepted):**
```json
{
  "tx_hash": "92a5fea273f064184fc8ccb5be223a9d14ec89adc9cdccec7b7ec9a930941969",
  "status": "PENDING",
  "unuah_amount": "",
  "error": ""
}
```

**Notes:**
- Final redeemed amount appears in `/api/tx/<tx_hash>` (event `sell_ndollar`).
- Background tracker updates the corresponding row in `transactions` (operation type `STABLECOIN_SELL`).

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

**Response (202 Accepted):**
```json
{
  "denom": "factory/nuah10us33fwsvajr57pgjxw638xzqjsfntqxk6yw56/test",
  "tx_hash": "0a6bc0af332d6e4adbe000523b48128fffb994098357a1d5926034222c0097c3",
  "status": "PENDING",
  "message": "Token creation broadcast, awaiting confirmation"
}
```

**Response Fields:**
- `denom` (string): Token denomination (used for trading)
- `tx_hash` (string): Blockchain transaction hash
- `status` (string): `PENDING`, `SUCCESS`, or `FAILED`
- `message` (string): Status message
- `error` (string, optional): Error message if transaction failed

**Error Responses:**
- `400 Bad Request`: Missing name or symbol
- `401 Unauthorized`: Missing or invalid token
- `500 Internal Server Error`: Transaction failed or server error

**Notes:**
- The transaction record is created with `status = PENDING` and updated automatically; inspect `/api/tx/<tx_hash>` for final state and detailed events (`usertoken.token_created`).
- Token creation mints tokens and distributes them according to configured percentages:
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

**Response (202 Accepted):**
```json
{
  "tx_hash": "35c78d5bfcb488013c7fcb0ba7706167d84fff4d0524763561426bfee0801114",
  "status": "PENDING",
  "tokens_out": "",
  "price_paid": "",
  "message": "Token purchase broadcast, awaiting confirmation"
}
```

**Notes:**
- Once the transaction lands on-chain, the tracker fills `tokens_out`/`price_paid` and flips `status` to `SUCCESS`/`FAILED` in the database.
- Use `/api/tx/<tx_hash>` to inspect bonding curve events for exact fills.

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

**Response (202 Accepted):**
```json
{
  "tx_hash": "edb4ae45b9b3fde9b9aae9990701bb0368d1374c8b43aa4a23357b735ede27f8",
  "status": "PENDING",
  "payment_out": "",
  "price_received": "",
  "message": "Token sale broadcast, awaiting confirmation"
}
```

**Notes:**
- Final payout metrics are visible after confirmation via `/api/tx/<tx_hash>` (event `usertoken.token_sold`).
- Database status transitions automatically; clients should treat immediate responses as intent acknowledgements.

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

## Exchange Endpoints

The Exchange module allows users to exchange supported cryptocurrencies (ETH, BTC, USDC, USDT, ATOM, OSMO, SOL) for unuah using real-time price feeds from the USD Oracle.

### POST /api/exchange/tokens

Exchange supported tokens for unuah.

**Authentication:** Required

**Request:**
```http
POST /api/exchange/tokens
Content-Type: application/json
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "token_denom": "uusdc",
  "amount": "100000000",
  "min_output": "95000000"
}
```

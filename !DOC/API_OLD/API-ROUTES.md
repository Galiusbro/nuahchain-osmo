# 🧩 API Endpoint Reference

> **Base URL:** `https://api.ndollar.org/api/v1`  
> All endpoints are relative to this base.  
> Authenticated routes require a valid `Authorization: Bearer <token>` header.

---

## 🟢 **POST Endpoints**

| Endpoint | Description |
|-----------|--------------|
| **POST** `/api/v1/auth/web/register` | Register a new user with `{ email, username, password }` |
| **POST** `/api/v1/auth/web/login` | Log in user with `{ email, password }` |
| **POST** `/api/v1/auth/farcaster/login` | Farcaster quick-login with `{ token: quickAuthJwt }` |
| **POST** `/api/v1/auth/refresh` | Refresh access token with `{ refreshToken }` |
| **POST** `/api/v1/auth/logout` | Log out current user (invalidate current access token) |
| **POST** `/api/v1/auth/logout-all` | Log out from all sessions (invalidate all tokens) |
| **POST** `/api/v1/auth/web/forgot-password` | Request password reset link for `{ email }` |
| **POST** `/api/v1/auth/web/reset-password` | Reset password with `{ token, newPassword }` |
| **POST** `/api/v1/create-token` | Create a new token `{ name, symbol, metadata }` |
| **POST** `/api/v1/tokens/batch-names` | Generate a batch of token names |
| **POST** `/api/v1/buy` | Buy a token (`IBuyTokenPayload`) |
| **POST** `/api/v1/sell` | Sell a token (`ISellTokenPayload`) |
| **POST** `/api/v1/n-dollar/swap-sol` | Swap SOL for NDollar `{ solAmount }` |
| **POST** `/api/v1/n-dollar/swap-ndollar` | Swap NDollar for SOL (`ISwapNDollarToSolPayload`) |
| **POST** `/api/v1/tokens/creator-buy` | Creator buy-in `{ tokenMintAddress }` |
| **POST** `/api/v1/tokens/skip-creator-buy` | Skip creator buy `{ tokenMintAddress }` |
| **POST** `/api/v1/users/me/link-telegram-token` | Generate Telegram link token `{ refreshToken }` |
| **POST** `/api/v1/referrals/claim` | Claim referral `{ IClaimReferralPayload }` |
| **POST** `/api/v1/referrals/user/claim` | Claim user referral `{ IClaimUserReferralBody }` |
| **POST** `/api/v1/tokens/{tokenMintAddress}/candlestick-interval` | Set candlestick interval `{ interval }` |
| **POST** `/api/v1/auth/wallet/start` | Start wallet authentication `{ chain, address, publicKeyBase58? }` |
| **POST** `/api/v1/auth/wallet/verify` | Verify wallet authentication `{ chain, address, signature, pubkey?, nonce, walletType? }` |
| **POST** `/api/v1/users/me/upload-image` | Upload user image (FormData: `image`) |

---

## 🔵 **GET Endpoints**

| Endpoint | Description |
|-----------|--------------|
| **GET** `/api/v1/health` | Health check for the API service |
| **GET** `/api/v1/health/db` | Database health check |
| **GET** `/api/v1/auth/sessions` | Get all active sessions for authenticated user |
| **GET** `/api/v1/users/me` | Get authenticated user profile |
| **GET** `/api/v1/users/me/info` | Get user info summary |
| **GET** `/api/v1/users/me/tokens` | List all tokens owned by the current user |
| **GET** `/api/v1/users/balances-db` | Get user balances (optional `tokenMint` param) |
| **GET** `/api/v1/tokens/market` | List all tokens in the marketplace |
| **GET** `/api/v1/tokens/search?query={query}` | Search tokens by name or symbol |
| **GET** `/api/v1/tokens/{mintAddress}/details` | Get detailed info for a token by mint address |
| **GET** `/api/v1/tokens/{mintAddress}/available-supply` | Get available supply of a token |
| **GET** `/api/v1/tokens/{mintAddress}/creator-telegram-id` | Get creator’s Telegram ID for a token |
| **GET** `/api/v1/tokens/{mintAddress}/tx-history?limit={limit}` | Get transaction history (default limit 50) |
| **GET** `/api/v1/tokens/{mintAddress}/history` | Get token price history |
| **GET** `/api/v1/tokens/{mintAddress}/ohlc` | Get OHLC (Open, High, Low, Close) price data |
| **GET** `/api/v1/tokens/{mintAddress}/holders?limit={limit}` | Get token holders (default 20) |
| **GET** `/api/v1/tokens/creation-status` | Get latest token creation status |
| **GET** `/api/v1/tokens/{mintAddress}/creation-status` | Get creation status for specific token |
| **GET** `/api/v1/tokens/check-name?name={name}` | Check if a token name exists |
| **GET** `/api/v1/tokens/{mintAddress}/referral-payload` | Get referral payload for a token |
| **GET** `/api/v1/quote/trade?tokenMintAddress={mintAddress}&amount={amount}&operation={buy|sell}&inputType={type}` | Get quote for trading token |
| **GET** `/api/v1/quote/swap-sol?solAmount={solAmount}` | Get quote for swapping SOL to NDollar |
| **GET** `/api/v1/n-dollar/quote-reverse?ndollarAmount={amount}` | Get quote for NDollar → SOL |
| **GET** `/api/v1/users/me/activity/points?limit={limit}` | Return user’s activity points (default 10) |
| **GET** `/api/v1/referrals/me/link` | Get user’s referral link `{ code, url }` |
| **GET** `/api/v1/referrals/me/users` | Get referred users (count + list) |
| **GET** `/api/v1/users/me/referral-links` | Get all referral links for user |
| **GET** `/api/v1/tokens/{mintAddress}/referral-payload` | Fetch referral payload for token |
| **GET** `/api/v1/tokens/{mintAddress}/creator-telegram-id` | Fetch creator’s Telegram ID for token |
| **GET** `/api/v1/tokens/{mintAddress}/holders?limit={limit}` | Fetch token holders |
| **GET** `/api/v1/users/me/tokens` | Fetch all user tokens |
| **GET** `/api/v1/users/me/activity/points` | Return activity points of user |
| **GET** `/api/v1/users/me/info` | Fetch user profile summary |

---

## 🟠 **PATCH Endpoints**

| Endpoint | Description |
|-----------|--------------|
| **PATCH** `/api/v1/users/username` | Update current user’s username `{ username }` |

---

## 🧠 Notes

- `{mintAddress}` and other placeholders should be replaced with actual IDs when making API calls.  
- `Content-Type` defaults to `application/json` unless otherwise specified.  
- `multipart/form-data` is used for image uploads.  
- Query parameters like `limit`, `name`, `query`, etc., should be URL-encoded.  
- Endpoints under `/users/me/...` require authentication (`Bearer token`).

---

> **Author:** Auto-generated from `api.ts` definitions  
> **Last Updated:** {{12/11/25}}
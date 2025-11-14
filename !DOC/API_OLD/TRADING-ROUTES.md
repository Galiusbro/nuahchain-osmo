# ⚙️ Trade API Endpoint Reference

> **Base URL:** `https://api.ndollar.org/api/v1`  
> All endpoints are relative to this base.  
> Authenticated routes require a valid `Authorization: Bearer <token>` header.

---

## 🟢 **POST Endpoints**

| Endpoint | Description |
|-----------|--------------|
| **POST** `/api/v1/autotrade/bots` | Create a new trading bot with payload `{ ICreateBotPayload }` |
| **POST** `/api/v1/autotrade/bots/{id}/start` | Start a specific trading bot |
| **POST** `/api/v1/autotrade/bots/{id}/stop` | Stop a specific trading bot |
| **POST** `/api/v1/autotrade/strategies/{name}/backtest` | Run a backtest for a given strategy `{ IRunBacktestPayload }` |
| **POST** `/api/v1/autotrade/set-status` | Set or update autotrade system status `{ ISetAutotradeStatusPayload }` |

---

## 🔵 **GET Endpoints**

| Endpoint | Description |
|-----------|--------------|
| **GET** `/api/v1/autotrade/bots` | List all existing bots for the user |
| **GET** `/api/v1/autotrade/bots/{id}` | Retrieve details for a specific bot |
| **GET** `/api/v1/autotrade/strategies` | List all available trading strategies |
| **GET** `/api/v1/autotrade/strategies/{name}` | Get configuration/details for a specific strategy |
| **GET** `/api/v1/autotrade/market/analysis/{token}` | Retrieve technical analysis data for a given token |
| **GET** `/api/v1/autotrade/market/candlesticks/{token}` | Get candlestick data for a token (optional query params: timeframe, limit, etc.) |
| **GET** `/api/v1/autotrade/market/conditions/{token}` | Fetch current market conditions for a token |
| **GET** `/api/v1/autotrade/performance/overview` | Get overall trading performance overview (accepts filters via query params) |
| **GET** `/api/v1/autotrade/performance/bots/{id}` | Get performance data for a specific bot (params: timeframe, metrics, etc.) |
| **GET** `/api/v1/autotrade/performance/history` | Get full trading history with optional filters (date range, botId, etc.) |
| **GET** `/api/v1/autotrade/details` | Retrieve global autotrade account details and metadata |
| **GET** `/api/v1/autotrade/get-status` | Get current status of the autotrade system (running, paused, etc.) |

---

## 🟡 **PUT Endpoints**

| Endpoint | Description |
|-----------|--------------|
| **PUT** `/api/v1/autotrade/bots/{id}` | Update an existing trading bot `{ IUpdateBotPayload }` |

---

## 🔴 **DELETE Endpoints**

| Endpoint | Description |
|-----------|--------------|
| **DELETE** `/api/v1/autotrade/bots/{id}` | Delete a trading bot permanently by its ID |

---

## 🧠 Notes

- Replace `{id}`, `{name}`, and `{token}` with actual values when calling each endpoint.
- All `/autotrade/...` routes are authenticated — ensure a valid JWT token is passed via the `Authorization` header.
- Payloads are JSON unless specified otherwise.
- Query params (like filters, timeframe, or limit) are optional and can refine results.
- For `backtest`, `createBot`, and `updateBot` endpoints, follow your respective interface types (`ICreateBotPayload`, `IRunBacktestPayload`, etc.).

---

> **Author:** Auto-generated from `tradeApi` definitions  
> **Last Updated:** {{12/11/25}}
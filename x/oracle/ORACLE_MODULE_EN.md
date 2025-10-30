## Oracle Module

This document describes the `x/oracle` module: its responsibilities, data model, messages, queries, and integration points.

### Overview
- The oracle stores asset prices and maintains optional price history.
- Prices can be set by authority via message or fetched automatically from external APIs (Yahoo Finance via `APIIKeeper`).
- The module exposes gRPC/REST/CLI queries for current price and historical entries.

### Key Components
- `AppModule` (`x/oracle/module.go`)
  - Registers Msg/Query servers and gRPC-Gateway routes.
  - Genesis import/export of stored prices.
  - Periodically updates prices in `BeginBlock` (every 300 blocks ~ 5 minutes) using `APIIKeeper`, and persists history entries.
- `Keeper` (`x/oracle/keeper/keeper.go`)
  - Core state access: `SetPrice`, `GetPrice`, `GetAllPrices`.
  - Price history: `SetPriceHistory`, `GetPriceHistory`, `GetLatestPriceHistory`.
- `APIIKeeper` (`x/oracle/keeper/sources.go`)
  - Extends `Keeper` with external data-source capabilities.
  - Fetches prices and history from Yahoo Finance (`sources.APIClient`).
  - Provides defaults: symbol categories and lists; helper methods and stats.
- `Query Server` (`x/oracle/keeper/query_server.go`)
  - gRPC handlers for price and price history queries with validation and sane limits.

### Data Model
- Proto types (`proto/osmosis/oracle/v1/oracle.proto`):
  - `Price { symbol, value, source, timestamp, confidence }`
  - `PriceHistoryEntry { symbol, value, source, timestamp, confidence, block_height }`
- Store keys (`x/oracle/types/keys.go`):
  - `PriceKeyPrefix = 0x01` → `PriceKey(symbol)`
  - `PriceHistoryKeyPrefix = 0x02` → `PriceHistoryKey(symbol, timestamp)`; time-sorted by big-endian timestamp; prefix helper for iteration.

### Messages
- `Msg SetPrice` (authority-only):
  - `MsgSetPrice { authority, symbol, value } → MsgSetPriceResponse`
  - Route: `POST /osmosis/oracle/v1/set-price`
  - Triggers `EventTypePriceUpdated` with attributes `symbol`, `value`, `authority`.

### Queries
- Service: `Query` (`x/oracle/types` via proto).
  - `Price(QueryPriceRequest{symbol}) → QueryPriceResponse{price}`
    - REST: `GET /osmosis/oracle/v1/price/{symbol}`
    - Example: `curl http://localhost:1317/osmosis/oracle/v1/price/BTC-USD`
  - `PriceHistory(QueryPriceHistoryRequest{symbol, start_time?, end_time?, limit?}) → QueryPriceHistoryResponse{entries}`
    - REST: `GET /osmosis/oracle/v1/price-history/{symbol}`
    - Query params: `start_time` (Unix timestamp), `end_time` (Unix timestamp), `limit` (default: 100, max: 1000)
    - Example: `curl http://localhost:1317/osmosis/oracle/v1/price-history/BTC-USD?start_time=1700000000&end_time=1700086400&limit=100`
    - If no time range provided, returns latest entries; default `limit=100`, max `1000`.

### CLI
- Root: `osmosisd q oracle`
- Current price:
  - `osmosisd q oracle price [symbol]`
- History (optional flags `--start-time`, `--end-time`, `--limit`):
  - `osmosisd q oracle price-history [symbol] --limit 200`

### Genesis
- Import: `InitGenesis` stores provided `prices` after validation.
- Export: `ExportGenesis` emits all stored prices.

### Automatic Updates and History
- `BeginBlock` every 300 blocks:
  - Resolves default symbol categories from `APIIKeeper.GetDefaultSymbols()`.
  - Updates a small subset per category (up to 3 symbols) to avoid rate limits.
  - On success, stores a `PriceHistoryEntry` with the current `block_height`.

### External Data Source
- Yahoo Finance via `sources.APIClient`:
  - `UpdatePriceFromAPI(ctx, symbol)` updates store with `Price` including `source`, `timestamp`, `confidence`.
  - `GetHistoricalData(ctx, symbol, period)` optional helper for integrations.
  - `SearchSymbols(ctx, query)` optional helper.
  - Defaults: `GetDefaultSymbols()` returns categories and curated symbol lists.

### Validation and Limits
- Symbol is sanitized via `EnsureSymbol(strings.TrimSpace)`.
- Query server validates inputs, enforces sensible `limit` (default 100, max 1000).

### Events
- `oracle.price_updated` with attributes: `symbol`, `value`, `authority`.

### Integration Notes
- Updating logic in `BeginBlock` is best-effort; failures are logged and skipped.
- To use external updates elsewhere, depend on `APIIKeeper` methods, or call `GetPriceWithFallback` to fetch/store on-demand.
- When running in production, consider API rate limiting and batching strategies; adjust symbol lists accordingly.

### Testing Aids
- See `x/oracle/keeper/*_test.go` and `test_api_keeper.go` for examples of fetching real API data and category-driven updates.

### Security Considerations
- Authority-gated `MsgSetPrice` prevents arbitrary writes.
- External data is tagged with `source` and `confidence` for downstream filtering.
- Keep API credentials and rate limits in mind if sources change in future.



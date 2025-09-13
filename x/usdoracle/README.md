# USD Oracle Module

The USD Oracle module provides real-time USD price data for the Osmosis blockchain, enabling accurate price feeds for stablecoins and other financial applications.

## Overview

This module implements a decentralized price oracle system that:
- Tracks USD prices from multiple sources
- Maintains historical price data
- Provides gRPC and CLI interfaces for price queries
- Supports configurable price deviation thresholds
- Enables governance-controlled parameter updates

## Architecture

### Core Components

1. **Keeper**: Core business logic for price management
2. **Types**: Message and query definitions
3. **Client**: CLI commands for interaction
4. **Genesis**: Module initialization and state export

### Key Features

- **Price Tracking**: Real-time USD price updates with validation
- **Historical Data**: Configurable price history storage
- **Multiple Sources**: Support for multiple price feed sources
- **Deviation Monitoring**: Automatic price deviation detection
- **Governance Integration**: Parameter updates via governance proposals

## State

### Current Price
```go
type CurrentPrice struct {
    Price     sdk.Dec
    Timestamp time.Time
    Source    string
}
```

### Price History
```go
type PriceHistory struct {
    Prices []PriceEntry
}

type PriceEntry struct {
    Price     sdk.Dec
    Timestamp time.Time
    Source    string
}
```

### Price Sources
```go
type PriceSource struct {
    Name      string
    Url       string
    Active    bool
    Weight    sdk.Dec
}
```

## Messages

### MsgUpdateUSDPrice
Updates the current USD price.

```go
type MsgUpdateUSDPrice struct {
    Authority string
    Price     sdk.Dec
    Source    string
}
```

### MsgSetPriceSources
Configures price feed sources.

```go
type MsgSetPriceSources struct {
    Authority string
    Sources   []PriceSource
}
```

### MsgUpdateParams
Updates module parameters via governance.

```go
type MsgUpdateParams struct {
    Authority string
    Params    Params
}
```

## Queries

### Current Price
```bash
osmosisd query usdoracle price
```

### Price History
```bash
osmosisd query usdoracle history
```

### Price Deviation
```bash
osmosisd query usdoracle deviation
```

### Parameters
```bash
osmosisd query usdoracle params
```

## Parameters

```go
type Params struct {
    MaxPriceAge           time.Duration // Maximum age for price data
    PriceDeviationLimit   sdk.Dec       // Maximum allowed price deviation
    HistorySize          uint64        // Number of historical entries to keep
    MinSources           uint64        // Minimum number of active sources
}
```

### Default Parameters
- `MaxPriceAge`: 5 minutes
- `PriceDeviationLimit`: 5% (0.05)
- `HistorySize`: 1000 entries
- `MinSources`: 1

## CLI Usage

### Query Commands

```bash
# Get current USD price
osmosisd query usdoracle price

# Get price history
osmosisd query usdoracle history

# Get price deviation
osmosisd query usdoracle deviation

# Get module parameters
osmosisd query usdoracle params
```

### Transaction Commands

```bash
# Update USD price (authority only)
osmosisd tx usdoracle update-price 1.00 "coinbase" --from authority

# Set price sources (authority only)
osmosisd tx usdoracle set-sources sources.json --from authority

# Update parameters via governance
osmosisd tx gov submit-proposal param-change proposal.json --from proposer
```

## gRPC API

### Query Service

```protobuf
service Query {
  // Get current USD price
  rpc GetUSDPrice(QueryGetUSDPriceRequest) returns (QueryGetUSDPriceResponse);

  // Get price history
  rpc GetPriceHistory(QueryGetPriceHistoryRequest) returns (QueryGetPriceHistoryResponse);

  // Get price deviation
  rpc GetPriceDeviation(QueryGetPriceDeviationRequest) returns (QueryGetPriceDeviationResponse);

  // Get module parameters
  rpc Params(QueryParamsRequest) returns (QueryParamsResponse);
}
```

### Message Service

```protobuf
service Msg {
  // Update USD price
  rpc UpdateUSDPrice(MsgUpdateUSDPrice) returns (MsgUpdateUSDPriceResponse);

  // Set price sources
  rpc SetPriceSources(MsgSetPriceSources) returns (MsgSetPriceSourcesResponse);

  // Update parameters
  rpc UpdateParams(MsgUpdateParams) returns (MsgUpdateParamsResponse);
}
```

## Integration

### Adding to App

1. Import the module:
```go
import "github.com/osmosis-labs/osmosis/v30/x/usdoracle"
```

2. Add to module manager:
```go
app.mm = module.NewManager(
    // other modules...
    usdoracle.NewAppModule(appCodec, app.USDOracleKeeper),
)
```

3. Add store key:
```go
keys := sdk.NewKVStoreKeys(
    // other keys...
    usdoracletypes.StoreKey,
)
```

4. Initialize keeper:
```go
app.USDOracleKeeper = usdoraclekeeper.NewKeeper(
    appCodec,
    keys[usdoracletypes.StoreKey],
    app.GetSubspace(usdoracletypes.ModuleName),
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
)
```

### Using in Other Modules

```go
// Get current USD price
price, found := app.USDOracleKeeper.GetCurrentPrice(ctx)
if !found {
    return sdkerrors.Wrap(types.ErrPriceNotFound, "USD price not available")
}

// Check if price is within acceptable deviation
if !app.USDOracleKeeper.IsWithinThreshold(ctx, newPrice) {
    return sdkerrors.Wrap(types.ErrPriceDeviation, "price deviation too high")
}
```

## Security Considerations

1. **Authority Control**: Only authorized addresses can update prices
2. **Deviation Limits**: Automatic rejection of prices outside acceptable ranges
3. **Source Validation**: Multiple source requirement for price updates
4. **Governance**: Parameter changes require governance approval

## Testing

Run module tests:
```bash
go test ./x/usdoracle/...
```

## Contributing

Contributions are welcome! Please ensure:
1. All tests pass
2. Code follows Go conventions
3. Documentation is updated
4. Security considerations are addressed

## License

This module is part of the Osmosis project and follows the same license terms.

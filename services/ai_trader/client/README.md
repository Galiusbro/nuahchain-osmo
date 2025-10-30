# AI Trader Client Library

This library provides a comprehensive client for AI trading operations on the Nuah blockchain, including oracle price queries, asset trading, and delegated operations through authz.

## Features

- **Oracle Client**: Query real-time prices from the oracle module
- **Trading Client**: Execute buy/sell operations for assets
- **Authz Client**: Execute delegated trading operations using MsgExec
- **Unified Client**: Combined interface for all trading operations
- **Comprehensive Testing**: Unit tests with mocked gRPC calls

## Architecture

```
services/ai_trader/client/
├── oracle/          # Oracle price queries
├── trading/         # Direct asset trading
├── authz/           # Delegated operations
├── client.go        # Unified client interface
└── example/         # Usage examples
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/osmosis-labs/osmosis/v30/services/ai_trader/client"
)

func main() {
    // Create client
    aiClient, err := client.NewClient("http://localhost:26657")
    if err != nil {
        log.Fatal(err)
    }
    defer aiClient.Close()

    ctx := context.Background()

    // Get price data
    price, err := aiClient.GetPriceData(ctx, "BTC")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("BTC Price: %s\n", price.Value)

    // Execute trading decision
    decision := &client.TradingDecision{
        Symbol:     "BTC",
        Action:     "buy",
        Amount:     "1000000",
        Price:      "50000.00",
        Reason:     "Price drop detected",
        Confidence: 0.8,
    }

    result, err := aiClient.ExecuteTradingDecision(ctx, decision,
        "cosmos1grantee", "cosmos1granter")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Trade executed: %v\n", result.Success)
}
```

## API Reference

### Oracle Client

```go
// Get single price
price, err := oracleClient.GetPrice(ctx, "BTC")

// Get multiple prices
prices, err := oracleClient.GetPrices(ctx, []string{"BTC", "ETH"})

// Validate price data
err := oracleClient.ValidatePrice(price)

// Check if price is stale
isStale := oracleClient.IsPriceStale(price, 30*time.Minute)
```

### Trading Client

```go
// Buy asset
result, err := tradingClient.BuyAsset(ctx, buyer, "BTC", "1000000")

// Sell asset
result, err := tradingClient.SellAsset(ctx, seller, "BTC", "0.02")

// Execute trade request against the assets module
req := &trading.TradeRequest{
    Symbol: "BTC",
    Amount: "1000000",
    Type:   "buy",
    Market: trading.MarketAssets,
}
result, err := tradingClient.ExecuteTrade(ctx, trader, req)

// Execute trade request against the bonding curve module
curveReq := &trading.TradeRequest{
    Symbol:       "factory/nuah1example/token",
    Amount:       "250.0",               // payment amount for buys
    Type:         "buy",
    Market:       trading.MarketBondingCurve,
    PaymentDenom: "NDOLLAR",             // optional, defaults to NDOLLAR
    MinOutput:    "120.0",               // optional slippage protection
}
curveResult, err := tradingClient.ExecuteTrade(ctx, trader, curveReq)
```

### Authz Client

```go
// Execute delegated buy
result, err := authzClient.ExecuteBuyAsset(ctx, grantee, granter, "BTC", "1000000")

// Execute delegated sell
result, err := authzClient.ExecuteSellAsset(ctx, grantee, granter, "BTC", "0.02")

// Execute delegated bonding curve trades
curveBuy, err := authzClient.ExecuteBuyFromCurve(ctx, grantee, granter, "factory/nuah1example/token", "NDOLLAR", "250.0", "120.0")
curveSell, err := authzClient.ExecuteSellToCurve(ctx, grantee, granter, "factory/nuah1example/token", "75.0", "NDOLLAR", "70.0")

// Execute multiple operations
msgs := []sdk.Msg{
    types.NewMsgBuyAsset(granter, "BTC", "1000000"),
    types.NewMsgSellAsset(granter, "ETH", "0.5"),
}
result, err := authzClient.ExecuteMultipleOperations(ctx, grantee, msgs)
```

## Data Structures

### PriceData

```go
type PriceData struct {
    Symbol     string    `json:"symbol"`
    Value      string    `json:"value"`
    Source     string    `json:"source"`
    Timestamp  time.Time `json:"timestamp"`
    Confidence float32   `json:"confidence"`
}
```

### TradingDecision

```go
type TradingDecision struct {
    Symbol       string              `json:"symbol"`
    Action       string              `json:"action"`     // "buy", "sell", "hold"
    Amount       string              `json:"amount"`
    Price        string              `json:"price"`
    Reason       string              `json:"reason"`
    Confidence   float32             `json:"confidence"`
    Market       trading.TradeMarket `json:"market,omitempty"`
    PaymentDenom string              `json:"payment_denom,omitempty"` // optional for bonding curve trades
    MinOutput    string              `json:"min_output,omitempty"`    // optional slippage control
}

### TradeMarket

```go
const (
    MarketAssets       trading.TradeMarket = "assets"
    MarketBondingCurve trading.TradeMarket = "bondingcurve"
)
```
```

### TradeResponse

```go
type TradeResponse struct {
    Symbol       string              `json:"symbol"`
    Type         string              `json:"type"`
    Amount       string              `json:"amount"`
    Result       string              `json:"result"`
    ResultDenom  string              `json:"result_denom,omitempty"`
    PaymentDenom string              `json:"payment_denom,omitempty"`
    Market       trading.TradeMarket `json:"market"`
    TxHash       string              `json:"tx_hash"`
    Timestamp    time.Time           `json:"timestamp"`
    Success      bool                `json:"success"`
    Error        string              `json:"error,omitempty"`
}
```

## Error Handling

All client methods return errors that should be checked:

```go
price, err := client.GetPriceData(ctx, "BTC")
if err != nil {
    // Handle error
    log.Printf("Failed to get price: %v", err)
    return
}
```

Common error types:
- `InvalidArgument`: Invalid parameters (empty symbol, invalid address)
- `NotFound`: Price not found for symbol
- `Unauthorized`: Insufficient permissions for authz operations
- `NetworkError`: gRPC connection issues

## Testing

The library includes comprehensive unit tests:

```bash
go test ./services/ai_trader/client/... -v
```

Tests cover:
- Price validation and staleness checks
- Trade request validation
- Authz request validation
- Error handling scenarios

## Configuration

The client requires a node URL for gRPC connections:

```go
client, err := client.NewClient("http://localhost:26657")
```

For production use, consider:
- Using TLS connections (`https://`)
- Implementing connection pooling
- Adding retry logic for network failures
- Setting appropriate timeouts

## Dependencies

- `github.com/cosmos/cosmos-sdk`: Core Cosmos SDK types and interfaces
- `google.golang.org/grpc`: gRPC client implementation
- `github.com/osmosis-labs/osmosis/v30/x/oracle`: Oracle module types
- `github.com/osmosis-labs/osmosis/v30/x/assets`: Assets module types

## License

This library is part of the Osmosis project and follows the same license terms.

# Exchange Module

The Exchange module provides functionality for exchanging various tokens for N$ (Nuah Dollar) stablecoin within the Osmosis ecosystem. This module integrates with both the USD Oracle and TWAP modules to obtain real-time price feeds with comprehensive validation, and implements robust security measures including transaction limits, daily limits, and advanced price deviation checks.

## Overview

The Exchange module enables users to:
- Exchange supported tokens (ETH, BTC, USDC, USDT, ATOM, OSMO, SOL) for N$ stablecoin
- Configure exchange parameters through governance
- Monitor exchange rates and transaction limits
- Manage daily exchange limits per address

## Architecture

### Core Components

1. **Keeper**: Core business logic for token exchanges and parameter management
2. **Types**: Message types, parameters, and data structures
3. **Client/CLI**: Command-line interface for user interactions
4. **Genesis**: Module initialization and state export/import

### Exchange Flow

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   Client    │───▶│   Exchange   │───▶│ USD Oracle  │
│ (User/CLI)  │    │   Module     │    │   Module    │
└─────────────┘    └──────────────┘    └─────────────┘
                           │                     │
                           ▼                     ▼
                   ┌──────────────┐    ┌─────────────┐
                   │     TWAP     │    │   Price     │
                   │   Module     │    │ Validation  │
                   └──────────────┘    └─────────────┘
                           │                     │
                           ▼                     ▼
                   ┌──────────────┐    ┌─────────────┐
                   │   Security   │    │   Token     │
                   │   Checks     │    │  Transfer   │
                   └──────────────┘    └─────────────┘
```

### Processing Pipeline

1. **Input Validation**: Validate token denomination, amount, and addresses
2. **Price Fetching**: Get current price from USD Oracle module
3. **TWAP Verification**: Cross-check price with TWAP module
4. **Security Checks**: Verify transaction and daily limits
5. **Price Deviation**: Check if price deviation is within acceptable range
6. **Token Transfer**: Execute the actual token exchange
7. **Event Emission**: Emit exchange events for monitoring

### Key Features

- **Multi-token Support**: Supports major cryptocurrencies and stablecoins (ETH, BTC, USDC, USDT, ATOM, OSMO, SOL)
- **Dual Oracle Integration**: Uses USD Oracle for real-time prices and TWAP for validation
- **Advanced Security**: Multi-layered protection with transaction limits, daily limits, and price deviation detection
- **Price Validation**: Cross-verification between Oracle and TWAP prices with configurable deviation thresholds
- **Anti-Whale Protection**: Daily limits per address to prevent market manipulation
- **Governance Control**: All parameters configurable through governance proposals
- **Treasury Integration**: Collected tokens and fees sent to community pool

## Parameters

The module supports the following configurable parameters:

| Parameter | Type | Description | Default |
|-----------|------|-------------|----------|
| `enabled` | bool | Enable/disable the exchange functionality | true |
| `admin` | string | Admin address for emergency controls | "" |
| `min_exchange_amount_usd` | Dec | Minimum exchange amount in USD | 10.0 |
| `max_exchange_amount_usd` | Dec | Maximum exchange amount in USD | 100000.0 |
| `daily_limit_usd` | Dec | Daily exchange limit per address in USD | 1000000.0 |
| `price_deviation_threshold` | Dec | Maximum allowed price deviation (%) | 0.02 |
| `exchange_fee` | Dec | Exchange fee percentage | 0.001 |
| `treasury_addresses` | []string | Treasury addresses for fee collection | [] |

## Messages

### MsgExchangeTokens

Exchange tokens for N$ stablecoin.

```protobuf
message MsgExchangeTokens {
  string sender = 1;
  cosmos.base.v1beta1.Coin amount = 2;
  string recipient = 3;
}
```

**Parameters:**
- `sender`: Address initiating the exchange
- `amount`: Token amount to exchange
- `recipient`: Address to receive N$ tokens

### MsgUpdateParams

Update module parameters (governance only).

```protobuf
message MsgUpdateParams {
  string authority = 1;
  Params params = 2;
}
```

**Parameters:**
- `authority`: Governance module address
- `params`: New parameter values

## Queries

### Params

Query current module parameters.

```bash
osmosisd query exchange params
```

### ExchangeRate

Query exchange rate for a specific token.

```bash
osmosisd query exchange exchange-rate [denom]
```

### DailyLimit

Query daily exchange limit for an address.

```bash
osmosisd query exchange daily-limit [address]
```

### AllExchangeRates

Query all available exchange rates.

```bash
osmosisd query exchange all-exchange-rates
```

## CLI Commands

### Exchange Tokens

Exchange tokens for N$ stablecoin.

```bash
osmosisd tx exchange exchange-tokens [amount] [recipient] --from [sender]
```

**Examples:**
```bash
# Exchange 100 OSMO for N$
osmosisd tx exchange exchange-tokens 100000000uosmo osmo1abc... --from mykey

# Exchange 0.1 ETH for N$
osmosisd tx exchange exchange-tokens 100000000000000000ibc/EA1D43981D5C9A1C4AAEA9C23BB1D4FA126BA9BC7020A25E0AE4AA841EA25DC5 osmo1abc... --from mykey

# Exchange 1000 USDC for N$
osmosisd tx exchange exchange-tokens 1000000000ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858 osmo1abc... --from mykey
```

### Update Parameters

Update module parameters (governance proposal required).

```bash
osmosisd tx exchange update-params [enabled] [admin] [min-amount] [max-amount] [daily-limit] [price-deviation] [exchange-fee] [treasury-addresses] --from [authority]
```

**Example:**
```bash
# Update daily limit to $2M via governance
osmosisd tx exchange update-params true "" 10.0 100000.0 2000000.0 0.02 0.001 "osmo1treasury..." --from governance
```

## Security Features

### Transaction Limits

- **Minimum Amount**: Configurable minimum exchange amount (default: $10)
- **Maximum Amount**: Configurable maximum exchange amount (default: $100,000)
- **Daily Limits**: Per-address daily exchange limits (default: $1,000,000)

### Price Protection

- **Dual Oracle System**: Real-time price feeds from USD Oracle module with TWAP validation
- **Price Deviation Detection**: Configurable threshold (default 2%) for detecting suspicious price movements
- **TWAP Cross-Verification**: All Oracle prices validated against TWAP prices before execution
- **Automatic Rejection**: Transactions automatically rejected if price deviation exceeds threshold
- **Governance Configurable**: Price deviation thresholds adjustable through governance proposals

### Anti-Whale Protection

- **Daily Limits**: Prevents large single-address accumulation
- **Rate Limiting**: Built-in transaction rate limiting
- **Governance Control**: All limits adjustable through governance

## Integration

### USD Oracle Module

The Exchange module depends on the USD Oracle module for:
- Real-time price feeds for supported tokens (ETH, BTC, USDC, USDT, ATOM, OSMO, SOL)
- Primary price source for exchange rate calculations
- Price validation and deviation checks

### TWAP Module

Integration with the TWAP module provides:
- Time-weighted average price calculations for validation
- Cross-verification of Oracle prices to detect anomalies
- Historical price data for trend analysis
- Enhanced security through dual price source validation

### Distribution Module

Integration with the distribution module for:
- Sending collected fees to community pool
- Treasury management
- Reward distribution

### Governance Module

Governance integration for:
- Parameter updates
- Emergency controls
- Token registry management

## Error Handling

The module implements comprehensive error handling for:

- **Invalid Parameters**: Validation of all input parameters and token denominations
- **Insufficient Funds**: Balance checks before exchanges with clear error messages
- **Price Deviations**: Detection and automatic rejection of transactions with suspicious price movements
- **Limit Violations**: Enforcement of minimum/maximum transaction limits and daily limits per address
- **Oracle Failures**: Graceful handling of USD Oracle unavailability with fallback mechanisms
- **TWAP Validation Failures**: Proper error handling when TWAP prices are unavailable or invalid
- **Unsupported Tokens**: Clear rejection of exchanges for non-supported token denominations
- **Daily Limit Exceeded**: Informative errors when users exceed their daily exchange limits

## Events

The module emits the following events:

### TokenExchanged

Emitted when tokens are successfully exchanged.

```json
{
  "type": "token_exchanged",
  "attributes": [
    {"key": "sender", "value": "osmo1..."},
    {"key": "recipient", "value": "osmo1..."},
    {"key": "input_amount", "value": "100000000uosmo"},
    {"key": "output_amount", "value": "150000000nuah"},
    {"key": "exchange_rate", "value": "1.5"},
    {"key": "usd_value", "value": "150.0"},
    {"key": "oracle_price", "value": "1.5"},
    {"key": "twap_price", "value": "1.48"},
    {"key": "price_deviation", "value": "0.013"},
    {"key": "fee", "value": "150000nuah"}
  ]
}
```

### DailyLimitUpdated

Emitted when daily limit is updated for an address.

```json
{
  "type": "daily_limit_updated",
  "attributes": [
    {"key": "address", "value": "osmo1..."},
    {"key": "date", "value": "2024-01-15"},
    {"key": "previous_amount", "value": "100000.0"},
    {"key": "new_amount", "value": "250000.0"},
    {"key": "limit_threshold", "value": "1000000.0"}
  ]
}
```

### ParamsUpdated

Emitted when module parameters are updated.

```json
{
  "type": "params_updated",
  "attributes": [
    {"key": "authority", "value": "osmo1..."},
    {"key": "updated_params", "value": "..."}
  ]
}
```

## Usage Examples

### Basic Token Exchange

1. **Check current exchange rates:**
```bash
osmosisd query exchange all-exchange-rates
```

2. **Check your daily limit status:**
```bash
osmosisd query exchange daily-limit osmo1youraddress...
```

3. **Exchange OSMO for N$:**
```bash
osmosisd tx exchange exchange-tokens 50000000uosmo osmo1youraddress... --from mykey --gas auto --gas-adjustment 1.5
```

### Advanced Usage

1. **Large exchange with limit checking:**
```bash
# First check if amount is within limits
osmosisd query exchange params

# Check current daily usage
osmosisd query exchange daily-limit osmo1youraddress...

# Execute exchange
osmosisd tx exchange exchange-tokens 100000000000uosmo osmo1youraddress... --from mykey
```

2. **Monitor exchange events:**
```bash
# Query recent exchange transactions
osmosisd query txs --events 'token_exchanged.sender=osmo1youraddress...'
```

### Governance Operations

1. **Submit parameter update proposal:**
```bash
osmosisd tx gov submit-proposal param-change proposal.json --from mykey
```

2. **Emergency disable (admin only):**
```bash
osmosisd tx exchange update-params false "osmo1admin..." 10.0 100000.0 1000000.0 0.02 0.001 "osmo1treasury..." --from admin
```

## Development

### Building

```bash
go build ./x/exchange/...
```

### Testing

```bash
# Run all tests
go test ./x/exchange/...

# Run specific test
go test ./x/exchange/keeper/... -run TestExchangeTokens -v

# Run with coverage
go test ./x/exchange/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Linting

```bash
golangci-lint run ./x/exchange/...
```

### Integration Testing

```bash
# Test with local chain
make test-integration

# Test specific scenarios
go test ./x/exchange/keeper/... -run TestKeeperTestSuite -v

# Test daily limit functionality
go test ./x/exchange/keeper/... -run TestExchangeTokensDailyLimitError -v

# Test price deviation checks
go test ./x/exchange/keeper/... -run TestExchangeTokensPriceDeviation -v
```

## Technical Implementation

### Supported Token Denominations

| Token | Denomination | IBC Path | Precision |
|-------|-------------|----------|----------|
| OSMO | uosmo | native | 6 |
| ETH | ibc/EA1D43981D5C9A1C4AAEA9C23BB1D4FA126BA9BC7020A25E0AE4AA841EA25DC5 | IBC | 18 |
| BTC | ibc/D1542AA8762DB13087D8364F3EA6509FD6F009A34F00426AF9E4F9FA85CBBF1F | IBC | 8 |
| USDC | ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858 | IBC | 6 |
| USDT | ibc/8242AD24008032E457D2E12D46588FD39FB54FB29680C6C7663D296B383C37C4 | IBC | 6 |
| ATOM | ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 | IBC | 6 |
| SOL | ibc/1E43D59E565D41FB4E54CA639B838FFD5BCACB3B5B83FF31D7C838F2A2AA2108 | IBC | 9 |

### State Management

#### Daily Limits Storage

```go
// Key format: "daily_limit/{address}/{date}"
// Value: DailyLimit protobuf message
type DailyLimit struct {
    Address           string
    Date              string  // YYYY-MM-DD format
    TotalExchangedUsd sdk.Dec
}
```

#### Parameters Storage

```go
// Key: "params"
// Value: Params protobuf message
type Params struct {
    Enabled                   bool
    Admin                     string
    MinExchangeAmountUsd      sdk.Dec
    MaxExchangeAmountUsd      sdk.Dec
    DailyLimitUsd            sdk.Dec
    PriceDeviationThreshold   sdk.Dec
    ExchangeFee              sdk.Dec
    TreasuryAddresses        []string
}
```

### Price Validation Logic

```go
// Pseudo-code for price validation
func ValidatePrice(oraclePrice, twapPrice sdk.Dec, threshold sdk.Dec) error {
    deviation := oraclePrice.Sub(twapPrice).Abs().Quo(twapPrice)
    if deviation.GT(threshold) {
        return fmt.Errorf("price deviation %.4f exceeds threshold %.4f",
                         deviation, threshold)
    }
    return nil
}
```

### Security Considerations

- **Reentrancy Protection**: All state changes are atomic
- **Input Validation**: Comprehensive validation of all inputs
- **Access Control**: Admin functions protected by governance
- **Rate Limiting**: Daily limits prevent abuse
- **Price Manipulation**: TWAP validation prevents oracle attacks
- **Overflow Protection**: Safe math operations throughout

## Implementation Status

### ✅ Completed Features

- **Core Exchange Functionality**: Token to N$ exchange with full validation
- **Multi-token Support**: ETH, BTC, USDC, USDT, ATOM, OSMO, SOL support
- **Dual Oracle Integration**: USD Oracle + TWAP price validation
- **Security Measures**: Transaction limits, daily limits, price deviation checks
- **Comprehensive Testing**: Full test suite with edge case coverage
- **Error Handling**: Robust error handling for all failure scenarios

### 🔄 Pending Implementation

- **Token Registry**: Governance-managed registry of supported tokens
- **Oracle Update Frequency**: Configurable update intervals (Oracle: 1min, TWAP: 5-15min)

### 🚀 Future Enhancements

- **Additional Token Support**: Expand supported token list through governance
- **Advanced Price Feeds**: Multiple oracle integration for enhanced reliability
- **Liquidity Pools**: Direct integration with AMM pools for better pricing
- **Cross-chain Support**: IBC token exchange capabilities
- **Advanced Analytics**: Enhanced monitoring and reporting dashboard

## License


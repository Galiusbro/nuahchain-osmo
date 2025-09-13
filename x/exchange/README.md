# Exchange Module

The Exchange module provides functionality for exchanging various tokens for N$ (Nuah Dollar) stablecoin within the Osmosis ecosystem. This module integrates with the USD Oracle module to obtain real-time price feeds and implements comprehensive security measures including transaction limits, daily limits, and price deviation checks.

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

### Key Features

- **Multi-token Support**: Supports major cryptocurrencies and stablecoins
- **Oracle Integration**: Uses USD Oracle module for real-time price feeds
- **Security Measures**: Transaction limits, daily limits, and price deviation checks
- **Governance Control**: All parameters configurable through governance proposals
- **Treasury Integration**: Collected tokens sent to community pool

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

**Example:**
```bash
osmosisd tx exchange exchange-tokens 100uosmo osmo1abc... --from mykey
```

### Update Parameters

Update module parameters (governance proposal required).

```bash
osmosisd tx exchange update-params [enabled] [admin] [min-amount] [max-amount] [daily-limit] [price-deviation] [exchange-fee] [treasury-addresses] --from [authority]
```

## Security Features

### Transaction Limits

- **Minimum Amount**: Configurable minimum exchange amount (default: $10)
- **Maximum Amount**: Configurable maximum exchange amount (default: $100,000)
- **Daily Limits**: Per-address daily exchange limits (default: $1,000,000)

### Price Protection

- **Oracle Integration**: Real-time price feeds from USD Oracle module
- **Price Deviation Check**: Configurable threshold for price deviation detection
- **TWAP Verification**: Integration with TWAP module for price validation

### Anti-Whale Protection

- **Daily Limits**: Prevents large single-address accumulation
- **Rate Limiting**: Built-in transaction rate limiting
- **Governance Control**: All limits adjustable through governance

## Integration

### USD Oracle Module

The Exchange module depends on the USD Oracle module for:
- Real-time price feeds for supported tokens
- Price validation and deviation checks
- Historical price data for TWAP calculations

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

- **Invalid Parameters**: Validation of all input parameters
- **Insufficient Funds**: Balance checks before exchanges
- **Price Deviations**: Detection and rejection of suspicious price movements
- **Limit Violations**: Enforcement of transaction and daily limits
- **Oracle Failures**: Graceful handling of oracle unavailability

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
    {"key": "input_amount", "value": "100uosmo"},
    {"key": "output_amount", "value": "150nuah"},
    {"key": "exchange_rate", "value": "1.5"},
    {"key": "fee", "value": "0.15nuah"}
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

## Development

### Building

```bash
go build ./x/exchange/...
```

### Testing

```bash
go test ./x/exchange/...
```

### Linting

```bash
golangci-lint run ./x/exchange/...
```

## Future Enhancements

- **Additional Token Support**: Expand supported token list
- **Advanced Price Feeds**: Multiple oracle integration
- **Liquidity Pools**: Direct integration with AMM pools
- **Cross-chain Support**: IBC token exchange capabilities
- **Advanced Analytics**: Enhanced monitoring and reporting

## License

This module is part of the Osmosis project and is licensed under the Apache 2.0 License.
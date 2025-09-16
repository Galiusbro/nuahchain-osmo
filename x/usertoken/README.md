# UserToken Module

The UserToken module is a comprehensive token management system built on top of the Osmosis blockchain that enables users to create, trade, and manage custom tokens with advanced features including bonding curves, liquidity bootstrapping pools (LBPs), founder token distribution, and vesting functionality.

## Table of Contents

- [Overview](#overview)
- [Core Concepts](#core-concepts)
- [Architecture](#architecture)
- [Messages](#messages)
- [Queries](#queries)
- [CLI Commands](#cli-commands)
- [Events](#events)
- [Parameters](#parameters)
- [State](#state)
- [Integration](#integration)

## Overview

The UserToken module provides a complete ecosystem for token creation and management with the following key features:

- **Token Creation**: Create custom tokens with metadata (name, symbol, decimals)
- **Bonding Curve Trading**: Automated market making with linear bonding curves
- **Founder Token Distribution**: Controlled token distribution for project founders
- **Liquidity Bootstrapping Pools (LBPs)**: Fair token distribution mechanism
- **Vesting Accounts**: Time-locked token distribution for teams and investors
- **Automatic Pool Creation**: Seamless integration with Osmosis AMM

## Core Concepts

### Token Lifecycle

1. **Creation Phase**: A user creates a token with basic metadata
2. **Trading Phase**: Tokens can be bought/sold via bonding curve
3. **Founder Distribution**: Founders can claim allocated tokens
4. **LBP Launch**: Optional liquidity bootstrapping pool for fair distribution
5. **Mature Trading**: Full AMM trading on Osmosis DEX

### Bonding Curve Mechanism

The module implements a linear bonding curve for token pricing:

```
Price = 0.0002 + (currentSupply / 30M) * (1.0 - 0.0002)
```

- **Minimum Price**: 0.0002 N$ (Nuah Dollar)
- **Maximum Price**: 1.0 N$ at 30M token supply
- **Linear Progression**: Price increases proportionally with supply

### Founder Token Economics

- **Total Allocation**: 30% of maximum supply (9M tokens)
- **Distribution**: Released in tranches over time
- **Vesting**: Can be combined with vesting accounts for time-locked distribution

### Liquidity Bootstrapping Pools (LBPs)

- **Purpose**: Fair price discovery and distribution
- **Mechanism**: Gradual weight shifting from project token to N$
- **Duration**: Configurable time period for price discovery
- **Integration**: Seamless transition to regular AMM pools

## Architecture

### Module Structure

```
x/usertoken/
├── client/cli/          # CLI commands
├── keeper/              # Business logic
├── types/               # Type definitions
├── simulation/          # Simulation logic
└── docs/               # Documentation
```

### Dependencies

The module integrates with several Osmosis modules:

- **TokenFactory**: Token creation and management
- **Bank**: Token transfers and balances
- **Account**: Account management and vesting
- **GAMM**: Automated market maker pools
- **PoolManager**: Pool creation and management

### Key Components

#### Keeper
The keeper manages all module state and business logic:
- Token metadata storage
- Bonding curve calculations
- Founder token distribution
- Pool creation and management
- Vesting account creation

#### Message Server
Handles all transaction messages:
- `CreateUserToken`: Token creation
- `BuyTokens`: Purchase tokens via bonding curve
- `SellTokens`: Sell tokens via bonding curve
- `ClaimFounderTokens`: Founder token distribution
- `StartLBP`: Launch liquidity bootstrapping pool
- `CreateVestingAccount`: Create time-locked distributions

## Messages

### MsgCreateUserToken

Creates a new user token with specified metadata.

```go
type MsgCreateUserToken struct {
    Creator  string // Token creator address
    Subdenom string // Token subdenom (max 44 chars, alphanumeric + hyphens)
    Name     string // Token name (max 128 chars)
    Symbol   string // Token symbol (max 32 chars, alphanumeric)
    Decimals uint32 // Token decimals (max 18)
}
```

**Validation Rules:**
- Subdenom: 1-44 characters, alphanumeric and hyphens only
- Name: 1-128 characters
- Symbol: 1-32 characters, alphanumeric only
- Decimals: 0-18

**Effects:**
- Creates token via TokenFactory
- Stores token metadata
- Automatically creates N$/UserToken pool
- Emits `create_user_token` event

### MsgBuyTokens

Purchases tokens using the bonding curve mechanism.

```go
type MsgBuyTokens struct {
    Buyer     string   // Buyer address
    Denom     string   // Token denomination to buy
    Amount    sdk.Coin // Payment amount in N$
    MinTokens string   // Minimum tokens to receive (slippage protection)
}
```

**Process:**
1. Validates payment amount and token existence
2. Calculates tokens to mint based on bonding curve
3. Transfers payment from buyer to module
4. Mints tokens to buyer
5. Updates token supply
6. Emits `buy_tokens` event

### MsgSellTokens

Sells tokens back to the bonding curve.

```go
type MsgSellTokens struct {
    Seller   string   // Seller address
    Denom    string   // Token denomination to sell
    Amount   sdk.Coin // Token amount to sell
    MinPrice string   // Minimum N$ to receive (slippage protection)
}
```

**Process:**
1. Validates token amount and ownership
2. Calculates payout based on bonding curve
3. Burns tokens from seller
4. Transfers N$ from module to seller
5. Updates token supply
6. Emits `sell_tokens` event

### MsgClaimFounderTokens

Allows founders to claim their allocated tokens.

```go
type MsgClaimFounderTokens struct {
    Claimer string // Founder address
    Denom   string // Token denomination
    Amount  string // Amount to claim
}
```

**Validation:**
- Claimer must be token creator
- Amount must not exceed remaining allocation
- Respects vesting schedules if applicable

### MsgStartLBP

Launches a liquidity bootstrapping pool for the token.

```go
type MsgStartLBP struct {
    Creator string // Token creator address
    Denom   string // Token denomination
}
```

**Process:**
1. Validates token exists and creator permissions
2. Creates balancer pool with initial weights
3. Sets up weight shifting schedule
4. Marks LBP as active
5. Emits `start_lbp` event

### MsgCreateVestingAccount

Creates a vesting account for time-locked token distribution.

```go
type MsgCreateVestingAccount struct {
    Creator   string    // Account creator
    ToAddress string    // Vesting account recipient
    Amount    sdk.Coins // Tokens to vest
    EndTime   int64     // Vesting end time (Unix timestamp)
    Delayed   bool      // Delayed vs continuous vesting
}
```

**Vesting Types:**
- **Continuous**: Tokens unlock gradually over time
- **Delayed**: All tokens unlock at end time

## Queries

The module provides several query endpoints:

### Query User Token
```bash
osmosisd query usertoken show-user-token [denom]
```

Returns token metadata and current state.

### Query All User Tokens
```bash
osmosisd query usertoken list-user-tokens
```

Returns all user tokens with pagination.

### Query Token Supply
```bash
osmosisd query usertoken token-supply [denom]
```

Returns current token supply.

### Query Bonding Curve Price
```bash
osmosisd query usertoken bonding-curve-price [denom]
```

Returns current token price based on supply.

### Query Founder Tokens Remaining
```bash
osmosisd query usertoken founder-tokens-remaining [denom]
```

Returns remaining founder token allocation.

## CLI Commands

### Transaction Commands

#### Create User Token
```bash
osmosisd tx usertoken create-user-token [subdenom] [name] [symbol] [decimals] \
  --from [creator] \
  --chain-id [chain-id] \
  --fees [fees]
```

**Example:**
```bash
osmosisd tx usertoken create-user-token "mytoken" "My Token" "MTK" 6 \
  --from creator \
  --chain-id osmosis-1 \
  --fees 5000uosmo
```

#### Buy Tokens
```bash
osmosisd tx usertoken buy-tokens [denom] [amount] [min-tokens] \
  --from [buyer] \
  --chain-id [chain-id] \
  --fees [fees]
```

**Example:**
```bash
osmosisd tx usertoken buy-tokens "factory/osmo1.../mytoken" "1000000unuah" "900000" \
  --from buyer \
  --chain-id osmosis-1 \
  --fees 5000uosmo
```

#### Sell Tokens
```bash
osmosisd tx usertoken sell-tokens [denom] [amount] [min-price] \
  --from [seller] \
  --chain-id [chain-id] \
  --fees [fees]
```

#### Claim Founder Tokens
```bash
osmosisd tx usertoken claim-founder-tokens [denom] [amount] \
  --from [founder] \
  --chain-id [chain-id] \
  --fees [fees]
```

#### Start LBP
```bash
osmosisd tx usertoken start-lbp [denom] \
  --from [creator] \
  --chain-id [chain-id] \
  --fees [fees]
```

#### Create Vesting Account
```bash
osmosisd tx usertoken create-vesting-account [to-address] [amount] [end-time] [delayed] \
  --from [creator] \
  --chain-id [chain-id] \
  --fees [fees]
```

**Example:**
```bash
osmosisd tx usertoken create-vesting-account \
  "osmo1recipient..." \
  "1000000factory/osmo1creator.../mytoken" \
  "1735689600" \
  "false" \
  --from creator \
  --chain-id osmosis-1 \
  --fees 5000uosmo
```

## Events

The module emits the following events:

### create_user_token
```json
{
  "type": "create_user_token",
  "attributes": [
    {"key": "creator", "value": "osmo1..."},
    {"key": "denom", "value": "factory/osmo1.../subdenom"},
    {"key": "subdenom", "value": "subdenom"},
    {"key": "name", "value": "Token Name"},
    {"key": "symbol", "value": "SYMBOL"},
    {"key": "decimals", "value": "6"}
  ]
}
```

### buy_tokens
```json
{
  "type": "buy_tokens",
  "attributes": [
    {"key": "buyer", "value": "osmo1..."},
    {"key": "denom", "value": "factory/osmo1.../subdenom"},
    {"key": "payment_amount", "value": "1000000unuah"},
    {"key": "tokens_minted", "value": "950000"},
    {"key": "new_supply", "value": "1950000"}
  ]
}
```

### sell_tokens
```json
{
  "type": "sell_tokens",
  "attributes": [
    {"key": "seller", "value": "osmo1..."},
    {"key": "denom", "value": "factory/osmo1.../subdenom"},
    {"key": "tokens_burned", "value": "100000"},
    {"key": "payout_amount", "value": "95000unuah"},
    {"key": "new_supply", "value": "1850000"}
  ]
}
```

### claim_founder_tokens
```json
{
  "type": "claim_founder_tokens",
  "attributes": [
    {"key": "claimer", "value": "osmo1..."},
    {"key": "denom", "value": "factory/osmo1.../subdenom"},
    {"key": "amount", "value": "500000"},
    {"key": "total_cost", "value": "100000unuah"},
    {"key": "founder_price", "value": "0.2unuah"}
  ]
}
```

### start_lbp
```json
{
  "type": "start_lbp",
  "attributes": [
    {"key": "creator", "value": "osmo1..."},
    {"key": "denom", "value": "factory/osmo1.../subdenom"},
    {"key": "pool_id", "value": "123"},
    {"key": "start_time", "value": "1640995200"}
  ]
}
```

### create_vesting_account
```json
{
  "type": "create_vesting_account",
  "attributes": [
    {"key": "creator", "value": "osmo1..."},
    {"key": "to_address", "value": "osmo1..."},
    {"key": "amount", "value": "1000000factory/osmo1.../subdenom"},
    {"key": "end_time", "value": "1735689600"},
    {"key": "delayed", "value": "false"}
  ]
}
```

## Parameters

The module uses the following parameters (stored in module params):

```go
type Params struct {
    // Add module parameters here as needed
}
```

Default parameters can be set during genesis or updated via governance.

## State

The module maintains the following state:

### UserToken

Stores metadata and state for each user token:

```go
type UserToken struct {
    Denom                string    // Token denomination
    Creator              string    // Token creator address
    CurrentSupply        math.Int  // Current token supply
    FounderTokensClaimed math.Int  // Founder tokens already claimed
    LbpActive           bool      // Whether LBP is active
    LbpStartTime        int64     // LBP start timestamp
}
```

**Storage Key**: `UserTokenKeyPrefix + denom`

### FounderTranche

Tracks founder token distribution:

```go
type FounderTranche struct {
    Denom           string    // Token denomination
    TotalAllocation math.Int  // Total founder allocation
    ClaimedAmount   math.Int  // Amount already claimed
    VestingSchedule []VestingPeriod // Vesting periods
}
```

**Storage Key**: `FounderTrancheKeyPrefix + denom`

## Integration

### Adding to App

To integrate the UserToken module into your Cosmos SDK application:

1. **Import the module**:
```go
import "github.com/osmosis-labs/osmosis/v30/x/usertoken"
```

2. **Add to module manager**:
```go
app.mm = module.NewManager(
    // other modules...
    usertoken.NewAppModule(appCodec, app.UserTokenKeeper),
)
```

3. **Add keeper to app**:
```go
app.UserTokenKeeper = usertokenkeeper.NewKeeper(
    appCodec,
    keys[usertokentypes.StoreKey],
    keys[usertokentypes.MemStoreKey],
    app.TokenFactoryKeeper,
    app.BankKeeper,
    app.AccountKeeper,
    app.GAMMKeeper,
    app.PoolManagerKeeper,
)
```

4. **Add to genesis**:
```go
genesisModuleOrder := []string{
    // other modules...
    usertokentypes.ModuleName,
}
```

### Dependencies

Ensure the following modules are available:
- `x/tokenfactory` - Token creation and management
- `x/bank` - Token transfers
- `x/auth` - Account management
- `x/gamm` - AMM functionality
- `x/poolmanager` - Pool management

### Permissions

The module requires the following permissions:
- **Minter**: To mint tokens via bonding curve
- **Burner**: To burn tokens when selling
- **Module Account**: To hold N$ reserves

## Security Considerations

1. **Bonding Curve Integrity**: Ensure mathematical precision in price calculations
2. **Founder Token Limits**: Validate founder allocations don't exceed limits
3. **Vesting Security**: Prevent premature token unlocking
4. **Pool Creation**: Validate pool parameters and permissions
5. **Access Control**: Ensure only authorized users can perform privileged operations

## Future Enhancements

- **Multiple Bonding Curves**: Support for different curve types
- **Governance Integration**: Token-based governance for user tokens
- **Advanced Vesting**: More complex vesting schedules
- **Cross-Chain Support**: IBC integration for multi-chain tokens
- **Fee Customization**: Configurable fees for different operations

## Contributing

Contributions to the UserToken module are welcome. Please ensure:

1. All code follows Go best practices
2. Tests are included for new functionality
3. Documentation is updated accordingly
4. Security considerations are addressed

For more information, see the [contributing guidelines](../../CONTRIBUTING.md).

## License

This module is part of the Osmosis project and is licensed under the Apache 2.0 License.
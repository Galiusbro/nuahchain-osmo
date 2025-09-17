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

- **Start Price**: 0.0002 N$ (Nuah Dollar)
- **End Price**: 1.0 N$ at 30M token supply
- **Linear Progression**: Price increases proportionally with supply
- **Average Price**: ~0.5001 N$ across the full curve
- **Total Cost**: ~15,003,000 N$ to purchase all 30M tokens
- **Supply Logic**: 30M public tokens + 10M founder tokens = 40M total supply

### Founder Token Economics

- **Total Allocation**: 10M tokens (25% of 40M total supply)
- **Founder Price**: 0.00005 N$ per token (configurable via governance)
- **Total Cost**: 500 N$ for all 10M founder tokens
- **Purchase Logic**:
  - Creator can buy all 10M founder tokens for 500 N$ total
  - If creator purchases: tokens are locked for 1 year (vesting)
  - If creator doesn't purchase: tokens automatically go to AI CEO wallet
- **Minimum Threshold**: The 500 N$ represents the minimum purchase amount to claim founder tokens
- **Governance**: Founder token price can be updated through governance proposals
- **Vesting**: Purchased founder tokens are automatically locked for 1 year

### Token Distribution

When a new token is created, the total supply of 100M tokens is distributed as follows:

- **30M tokens** → Bonding curve (for public trading, remain in module)
- **10M tokens** → Platform wallet
- **10M tokens** → Referral wallet
- **40M tokens** → AI CEO wallet
- **10M tokens** → Reserved for founder (can be purchased for 500 N$, remain in module until purchased)

If the founder doesn't purchase their 10M tokens within 7 days, those tokens are automatically transferred to the AI CEO wallet, bringing their total to 50M tokens.

**Module Balance**: The usertoken module holds 40M tokens initially (30M for bonding curve + 10M founder reserve). After founder token expiration, it holds only 30M tokens for the bonding curve.

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

### Special Wallets

The module uses several special wallet addresses for token distribution:

- **AI CEO Wallet**: Receives founder tokens when creators don't meet minimum purchase requirements
- **Referral Wallet**: Handles referral program rewards and incentives
- **Platform Fee Wallet**: Collects platform fees from various operations

These wallet addresses are configured during chain initialization and can be updated via governance.

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
    {"key": "amount", "value": "10000000"},
    {"key": "total_cost", "value": "500000000unuah"},
    {"key": "founder_price", "value": "0.00005unuah"}
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
    FounderTranchePrice    math.LegacyDec // Price per founder token in N$
    FounderTrancheAmount   math.Int       // Total founder tokens available
    BondingCurveStartPrice math.LegacyDec // Starting price for bonding curve
    BondingCurveEndPrice   math.LegacyDec // Ending price for bonding curve
    BondingCurveMaxSupply  math.Int       // Maximum supply for bonding curve
    MinCreatorPurchase     math.LegacyDec // Minimum N$ amount creator must purchase
    AiCeoWallet            string         // AI CEO wallet address for redirected tokens
    ReferralWallet         string         // Referral program wallet address
    PlatformFeeWallet      string         // Platform fee wallet address
}
```

### Default Parameter Values

| Parameter | Default Value | Description |
|-----------|---------------|-------------|
| `FounderTranchePrice` | `0.00005` N$ | Price per founder token |
| `FounderTrancheAmount` | `10,000,000` tokens | Total founder tokens available (10M) |
| `BondingCurveStartPrice` | `0.0002` N$ | Starting price for public bonding curve |
| `BondingCurveEndPrice` | `1.0` N$ | Maximum price at full supply |
| `BondingCurveMaxSupply` | `30,000,000` tokens | Maximum supply for bonding curve (30M) |
| `MinCreatorPurchase` | `500` N$ | Minimum purchase amount - if creator doesn't buy, tokens go to AI CEO |
| `AiCeoWallet` | `""` | Set during chain initialization |
| `ReferralWallet` | `""` | Set during chain initialization |
| `PlatformFeeWallet` | `""` | Set during chain initialization |

### Parameter Economics

- **Founder Token Opportunity**: 10M tokens × 0.00005 N$ = 500 N$ total
- **Creator Choice**:
  - **Buy**: Pay 500 N$, get 10M tokens locked for 1 year
  - **Don't Buy**: Tokens automatically go to AI CEO wallet
- **Public Curve Range**: 0.0002 N$ → 1.0 N$ across 30M tokens
- **Average Public Price**: ~0.5001 N$ per token
- **Full Curve Cost**: ~15,003,000 N$ for all 30M public tokens
- **Minimum Threshold**: 500 N$ is the fixed purchase amount (not a minimum requirement)

Parameters can be updated via governance proposals.

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
2. **Founder Token Distribution**:
   - Validate minimum purchase requirements (500 N$)
   - Secure redirection to AI CEO wallet when requirements not met
   - Prevent double-claiming of founder tokens
3. **Wallet Security**:
   - Protect AI CEO, Referral, and Platform Fee wallet addresses
   - Validate wallet configurations during initialization
   - Secure governance updates of wallet addresses
4. **Vesting Security**: Prevent premature token unlocking
5. **Pool Creation**: Validate pool parameters and permissions
6. **Access Control**:
   - Only token creators can claim founder tokens
   - Validate all message signers and permissions
   - Prevent unauthorized parameter updates
7. **Economic Security**:
   - Validate all price calculations and token amounts
   - Prevent overflow/underflow in mathematical operations
   - Ensure proper handling of edge cases in token distribution

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

### Testing

The module includes comprehensive tests covering:

- **Message Validation**: All message types and their validation logic
- **Token Economics**: Bonding curve calculations and founder token distribution
- **Token Distribution**: Verification of correct token allocation (30M bonding curve, 10M platform, 10M referral, 40M AI CEO, 10M founder reserve)
- **Module Balance**: Testing that module holds correct amounts (40M initially, 30M after founder expiration)
- **Wallet Configuration**: Proper setup and validation of platform, referral, and AI CEO wallets
- **Edge Cases**: Minimum purchase requirements, wallet redirections, founder token expiration
- **Integration**: Cross-module interactions with TokenFactory, Bank, etc.

Run tests with:
```bash
go test ./x/usertoken/... -v
```

### Recent Updates

- **Fixed founder token distribution logic** with minimum purchase requirements
- **Added AI CEO wallet redirection** for insufficient purchases
- **Updated test cases** to reflect new tokenomics (500 N$ minimum)
- **Enhanced parameter validation** and economic security
- **Added comprehensive token distribution testing** to verify correct allocation of 100M tokens
- **Updated documentation** to clearly describe token distribution scheme and module balance logic
- **Improved test coverage** for wallet configuration and balance verification

For more information, see the [contributing guidelines](../../CONTRIBUTING.md).

## License

This module is part of the Osmosis project and is licensed under the Apache 2.0 License.

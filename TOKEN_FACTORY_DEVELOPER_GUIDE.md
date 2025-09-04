# TokenFactory Developer Guide 🔧

Technical documentation for developers working with Nuah Chain's TokenFactory module.

## 🏗️ Architecture Overview

TokenFactory is a Cosmos SDK module forked from Osmosis that enables permissionless token creation. Each token is identified by a unique denomination string following the pattern: `factory/{creator_address}/{subdenom}`.

### Core Components

```
x/tokenfactory/
├── keeper/          # Business logic and state management
├── types/           # Message types and codec definitions
├── client/cli/      # CLI commands and queries
└── module.go        # Module definition and configuration
```

### Message Types

| Message | Purpose | Parameters |
|---------|---------|------------|
| `MsgCreateDenom` | Create new token | sender, subdenom |
| `MsgMint` | Mint tokens | sender, amount, mintToAddress |
| `MsgBurn` | Burn tokens | sender, amount, burnFromAddress |
| `MsgChangeAdmin` | Transfer admin rights | sender, denom, newAdmin |
| `MsgSetDenomMetadata` | Set/update metadata | sender, metadata |
| `MsgSetBeforeSendHook` | Set transfer hooks | sender, denom, cosmwasmAddress |
| `MsgForceTransfer` | Admin force transfer | sender, amount, transferFromAddress, transferToAddress |

## 🔧 Integration Guide

### Setting Up the Module

The TokenFactory module is automatically included in Nuah Chain. No additional setup required.

### Querying TokenFactory Data

#### gRPC Queries

```go
import (
    "github.com/osmosis-labs/osmosis/v30/x/tokenfactory/types"
)

// Query denoms created by an address
func QueryDenomsFromCreator(client types.QueryClient, creator string) (*types.QueryDenomsFromCreatorResponse, error) {
    return client.DenomsFromCreator(context.Background(), &types.QueryDenomsFromCreatorRequest{
        Creator: creator,
    })
}

// Query denom authority metadata
func QueryDenomAuthorityMetadata(client types.QueryClient, denom string) (*types.QueryDenomAuthorityMetadataResponse, error) {
    return client.DenomAuthorityMetadata(context.Background(), &types.QueryDenomAuthorityMetadataRequest{
        Creator:  creator,
        SubDenom: subdenom,
    })
}
```

#### REST API

```bash
# Query denoms from creator
GET /osmosis/tokenfactory/v1beta1/denoms_from_creator/{creator}

# Query denom authority
GET /osmosis/tokenfactory/v1beta1/denoms/{creator}/{subdenom}/authority_metadata

# Query module parameters
GET /osmosis/tokenfactory/v1beta1/params
```

### Creating Messages Programmatically

```go
import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/osmosis-labs/osmosis/v30/x/tokenfactory/types"
)

// Create a new denomination
func CreateDenomMsg(creator, subdenom string) *types.MsgCreateDenom {
    return &types.MsgCreateDenom{
        Sender:   creator,
        Subdenom: subdenom,
    }
}

// Mint tokens
func MintTokensMsg(sender, amount, mintTo string) *types.MsgMint {
    coins, _ := sdk.ParseCoinsNormalized(amount)
    return &types.MsgMint{
        Sender:        sender,
        Amount:        coins[0],
        MintToAddress: mintTo,
    }
}

// Set metadata
func SetMetadataMsg(sender string, metadata banktypes.Metadata) *types.MsgSetDenomMetadata {
    return &types.MsgSetDenomMetadata{
        Sender:   sender,
        Metadata: metadata,
    }
}
```

## 🐛 Troubleshooting Fixed Issues

During implementation, we encountered and fixed several critical issues:

### 1. Bech32 Prefix Mismatch

**Problem:** `hrp does not match bech32 prefix: expected 'osmo' got 'nuah'`

**Root Cause:** Multiple configuration files contained hardcoded `osmo` prefix.

**Files Fixed:**
```
osmoutils/noapptest/cdc.go              # Line 28
x/txfees/types/msgs_test.go            # Line 47  
x/epochs/keeper/keeper_test.go         # Line 39
app/params/proto.go                    # Line 15
app/params/amino.go                    # Line 15
```

**Solution:**
```go
// Before
interfaceRegistry := testutil.CodecOptions{AccAddressPrefix: "osmo", ValAddressPrefix: "nuahvaloper"}

// After  
interfaceRegistry := testutil.CodecOptions{AccAddressPrefix: "nuah", ValAddressPrefix: "nuahvaloper"}
```

### 2. Hardcoded Burn Address

**Problem:** Fee decorator used hardcoded Osmosis burn address.

**Files Fixed:**
```
x/txfees/keeper/feedecorator.go        # Line 298
x/txfees/keeper/feedecorator_test.go   # Line 166
```

**Solution:**
```go
// Before
burnAcctAddr, _ := sdk.AccAddressFromBech32("osmo1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqmcn030")

// After
burnAcctAddr, _ := sdk.AccAddressFromBech32("nuah1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqvthuyl")
```

### 3. Hardcoded Message Types

**Problem:** Amino codec registration used hardcoded `osmosis` prefix.

**File Fixed:**
```
x/tokenfactory/types/codec.go          # Lines 14-20
```

**Solution:**
```go
// Before
legacy.RegisterAminoMsg(cdc, &MsgCreateDenom{}, "osmosis/tokenfactory/create-denom")

// After
legacy.RegisterAminoMsg(cdc, &MsgCreateDenom{}, "nuah/tokenfactory/create-denom")
```

### 4. Incorrect DenomCreationFee

**Problem:** Module configuration used wrong base denomination.

**File Fixed:**
```
x/tokenfactory/module.go               # Line 90
```

**Solution:**
```go
// Before  
DenomCreationFee: sdk.NewCoins(sdk.NewInt64Coin("nuah", 2000)),

// After
DenomCreationFee: sdk.NewCoins(sdk.NewInt64Coin("unuah", 2000)),
```

## 🧪 Testing

### Unit Tests

Run TokenFactory-specific tests:
```bash
go test ./x/tokenfactory/...
```

### Integration Tests

Test end-to-end token creation:
```bash
# Create token
nuahd tx tokenfactory create-denom testtoken --from test-account --gas 1500000 --fees 16000unuah -y

# Verify creation
nuahd query tokenfactory denoms-from-creator $(nuahd keys show test-account -a)

# Test minting
nuahd tx tokenfactory mint 1000000factory/$(nuahd keys show test-account -a)/testtoken $(nuahd keys show test-account -a) --from test-account --gas 200000 --fees 2000unuah -y
```

### Performance Testing

Benchmark token operations:
```go
func BenchmarkCreateDenom(b *testing.B) {
    for i := 0; i < b.N; i++ {
        subdenom := fmt.Sprintf("token%d", i)
        _, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
            Sender:   creator,
            Subdenom: subdenom,
        })
        require.NoError(b, err)
    }
}
```

## 🔐 Security Considerations

### Admin Privileges

Token creators have significant control:
- **Minting:** Can create unlimited supply
- **Burning:** Can burn from any address (if enabled)  
- **Force Transfer:** Can move tokens between addresses
- **Metadata:** Can change token information

### Validation Rules

The module enforces several constraints:
```go
// Subdenom validation
MaxSubdenomLength = 44
ValidSubdenomRegex = `^[a-zA-Z0-9./]{1,44}$`

// Creator validation  
MaxCreatorLength = 59
```

### Gas Consumption

| Operation | Approximate Gas |
|-----------|----------------|
| CreateDenom | ~1,000,000 |
| SetMetadata | ~300,000 |
| Mint | ~200,000 |
| Burn | ~200,000 |

## 📊 Module Parameters

Current TokenFactory parameters:

```go
type Params struct {
    DenomCreationFee        sdk.Coins  // Fee to create denomination
    DenomCreationGasConsume uint64     // Gas consumed for denom creation
}
```

Default values:
```go
DenomCreationFee: sdk.NewCoins(sdk.NewInt64Coin("unuah", 2000))
DenomCreationGasConsume: 1_000_000
```

## 🎯 CosmWasm Integration

TokenFactory can be integrated with CosmWasm contracts:

### Setting Hooks

```go
msg := &types.MsgSetBeforeSendHook{
    Sender:           admin,
    Denom:           tokenDenom,
    CosmwasmAddress: contractAddress,
}
```

### Hook Contract Interface

```rust
#[cw_serde]
pub struct BeforeSendHookMsg {
    pub from: String,
    pub to: String,
    pub amount: Coin,
}

pub fn before_send_hook(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: BeforeSendHookMsg,
) -> Result<Response, ContractError> {
    // Implement custom logic
    // Return error to block transfer
    Ok(Response::new())
}
```

## 🔄 Migration Guide

### From Osmosis TokenFactory

If migrating from Osmosis:

1. **Update imports:**
```go
// Old
import "github.com/osmosis-labs/osmosis/v20/x/tokenfactory/types"

// New
import "github.com/osmosis-labs/osmosis/v30/x/tokenfactory/types"
```

2. **Update addresses:** All `osmo` prefixed addresses need conversion to `nuah`.

3. **Update configurations:** Check all hardcoded references to Osmosis.

### State Migration

The module state is compatible. No manual migration needed.

## 📈 Monitoring & Metrics

### Key Metrics to Track

```go
// Number of denoms created
totalDenoms := len(k.GetAllDenoms(ctx))

// Total fees collected
feeCollectorBalance := bankKeeper.GetBalance(ctx, feeCollectorAddress, "unuah")

// Gas usage statistics
avgGasPerOperation := totalGasUsed / totalOperations
```

### Prometheus Metrics

```go
var (
    denomsCreated = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "tokenfactory_denoms_created_total",
            Help: "Total number of denoms created",
        },
        []string{"creator"},
    )
)
```

## 🔗 API Reference

### CLI Commands

```bash
# Transactions
nuahd tx tokenfactory create-denom <subdenom>
nuahd tx tokenfactory mint <amount> <mint-to-address>  
nuahd tx tokenfactory burn <amount> <burn-from-address>
nuahd tx tokenfactory change-admin <denom> <new-admin>
nuahd tx tokenfactory set-denom-metadata <metadata-json>
nuahd tx tokenfactory set-before-send-hook <denom> <contract-address>
nuahd tx tokenfactory force-transfer <amount> <from> <to>

# Queries
nuahd query tokenfactory params
nuahd query tokenfactory denoms-from-creator <creator-address>
nuahd query tokenfactory denom-authority-metadata <creator> <subdenom>
```

### gRPC Services

```protobuf
service Query {
    rpc Params(QueryParamsRequest) returns (QueryParamsResponse);
    rpc DenomAuthorityMetadata(QueryDenomAuthorityMetadataRequest) returns (QueryDenomAuthorityMetadataResponse);
    rpc DenomsFromCreator(QueryDenomsFromCreatorRequest) returns (QueryDenomsFromCreatorResponse);
    rpc BeforeSendHookAddress(QueryBeforeSendHookAddressRequest) returns (QueryBeforeSendHookAddressResponse);
}

service Msg {
    rpc CreateDenom(MsgCreateDenom) returns (MsgCreateDenomResponse);
    rpc Mint(MsgMint) returns (MsgMintResponse);
    rpc Burn(MsgBurn) returns (MsgBurnResponse);
    rpc ChangeAdmin(MsgChangeAdmin) returns (MsgChangeAdminResponse);
    rpc SetDenomMetadata(MsgSetDenomMetadata) returns (MsgSetDenomMetadataResponse);
    rpc SetBeforeSendHook(MsgSetBeforeSendHook) returns (MsgSetBeforeSendHookResponse);
    rpc ForceTransfer(MsgForceTransfer) returns (MsgForceTransferResponse);
}
```

## 🚀 Advanced Use Cases

### 1. Governance Tokens

```go
// Create governance token with voting power
governanceToken := types.MsgCreateDenom{
    Sender:   daoAddress,
    Subdenom: "govtoken",
}

// Set metadata for governance
metadata := banktypes.Metadata{
    Name:        "DAO Governance Token",
    Symbol:      "GOV",
    Display:     "gov",
    Base:        fullDenom,
    Description: "Voting power for DAO governance",
}
```

### 2. Staking Derivatives

```go
// Create liquid staking token
lstToken := types.MsgCreateDenom{
    Sender:   stakingProtocolAddress,
    Subdenom: "stnuah",  // Staked NUAH
}

// Mint LST when users stake
mintMsg := types.MsgMint{
    Sender:        stakingProtocolAddress,
    Amount:        stakeAmount,
    MintToAddress: userAddress,
}
```

### 3. Yield Farming Tokens

```go
// Create yield farm token
farmToken := types.MsgCreateDenom{
    Sender:   farmContractAddress,
    Subdenom: "farmtoken",
}

// Burn tokens when claiming rewards
burnMsg := types.MsgBurn{
    Sender:           farmContractAddress,
    Amount:           rewardTokens,
    BurnFromAddress:  userAddress,
}
```

## 📚 Resources

### Official Documentation
- [Cosmos SDK Modules](https://docs.cosmos.network/main/modules)
- [Osmosis TokenFactory](https://docs.osmosis.zone/osmosis-core/modules/tokenfactory)

### Community Resources  
- [Nuah Chain GitHub](https://github.com/osmosis-labs/osmosis)
- [Developer Discord](https://discord.gg/osmosis)
- [Forum Discussions](https://forum.osmosis.zone)

### Example Implementations
- [Token Launchpad](https://github.com/example/token-launchpad)
- [Governance Integration](https://github.com/example/governance-tokens)
- [DeFi Protocols](https://github.com/example/defi-tokenfactory)

---

🔧 **Happy Building!** This guide covers everything you need to integrate with and extend the TokenFactory module. For specific questions, reach out to the developer community.
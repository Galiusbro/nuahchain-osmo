# Vesting Functionality in UserToken Module

The UserToken module provides vesting functionality that allows token creators to create vesting accounts for token distribution with time-based release schedules.

## Overview

Vesting accounts are special accounts that hold tokens that are released over time according to a predefined schedule. This is commonly used for:

- Team token allocations
- Investor token distributions
- Community rewards
- Founder token releases

## Message Types

### MsgCreateVestingAccount

Creates a new vesting account with specified parameters.

```protobuf
message MsgCreateVestingAccount {
  string creator = 1;           // Address of the account creator
  string to_address = 2;        // Address of the vesting account recipient
  repeated Coin amount = 3;     // Tokens to be vested
  int64 end_time = 4;          // Unix timestamp when vesting ends
  bool delayed = 5;            // Whether this is a delayed vesting account
}
```

#### Parameters

- **creator**: The address creating the vesting account (must have sufficient balance)
- **to_address**: The recipient address that will receive the vested tokens
- **amount**: Array of coins to be vested (supports multiple denominations)
- **end_time**: Unix timestamp when all tokens will be fully vested
- **delayed**: 
  - `false`: Continuous vesting (tokens unlock gradually over time)
  - `true`: Delayed vesting (all tokens unlock at end_time)

## Usage Examples

### Creating a Continuous Vesting Account

```bash
# Vest 1,000,000 user tokens over 1 year (continuous release)
osmosisd tx usertoken create-vesting-account \
  --to-address osmo1recipient... \
  --amount 1000000factory/osmo1creator.../mytoken \
  --end-time 1735689600 \
  --delayed false \
  --from creator
```

### Creating a Delayed Vesting Account

```bash
# Vest 500,000 user tokens with cliff (all unlock at end time)
osmosisd tx usertoken create-vesting-account \
  --to-address osmo1recipient... \
  --amount 500000factory/osmo1creator.../mytoken \
  --end-time 1735689600 \
  --delayed true \
  --from creator
```

## Vesting Types

### Continuous Vesting

- Tokens are released linearly over time
- Recipients can access vested tokens at any point
- Formula: `vested_amount = total_amount * (current_time - start_time) / (end_time - start_time)`

### Delayed Vesting (Cliff)

- All tokens are locked until the end time
- No tokens are accessible before the cliff date
- All tokens become available at once when end_time is reached

## Integration with Cosmos SDK Vesting

The UserToken vesting functionality leverages the Cosmos SDK's native vesting account types:

- **Continuous Vesting**: Uses `ContinuousVestingAccount`
- **Delayed Vesting**: Uses `DelayedVestingAccount`

## Events

When a vesting account is created, the following events are emitted:

```json
{
  "type": "create_vesting_account",
  "attributes": [
    {
      "key": "creator",
      "value": "osmo1creator..."
    },
    {
      "key": "to_address",
      "value": "osmo1recipient..."
    },
    {
      "key": "amount",
      "value": "1000000factory/osmo1creator.../mytoken"
    },
    {
      "key": "end_time",
      "value": "1735689600"
    },
    {
      "key": "delayed",
      "value": "false"
    }
  ]
}
```

## Validation Rules

1. **Creator Balance**: Creator must have sufficient balance of all tokens being vested
2. **Valid Addresses**: Both creator and to_address must be valid bech32 addresses
3. **Positive Amount**: Vesting amount must be positive for all denominations
4. **Future End Time**: end_time must be in the future
5. **Unique Recipient**: Cannot create multiple vesting accounts for the same recipient address

## Error Handling

Common errors and their meanings:

- `insufficient funds`: Creator doesn't have enough tokens to vest
- `invalid address`: Malformed creator or recipient address
- `invalid end time`: End time is in the past or invalid
- `empty amount`: No tokens specified for vesting
- `account already exists`: Vesting account already exists for the recipient

## Query Commands

### Check Vesting Account Status

```bash
# Query vesting account details
osmosisd query auth account osmo1recipient...

# Query vesting account balances
osmosisd query bank balances osmo1recipient...
```

### Check Vested vs Locked Tokens

```bash
# Query spendable (vested) tokens
osmosisd query bank spendable-balances osmo1recipient...
```

## Best Practices

1. **Test on Testnet**: Always test vesting parameters on testnet first
2. **Clear Communication**: Ensure recipients understand the vesting schedule
3. **Proper End Times**: Use realistic end times with proper timezone considerations
4. **Multiple Denominations**: You can vest multiple token types in a single account
5. **Documentation**: Keep records of all vesting schedules for transparency

## Security Considerations

1. **Irreversible**: Vesting account creation is irreversible
2. **Token Transfer**: Tokens are immediately transferred from creator to vesting account
3. **Access Control**: Only the recipient can access vested tokens
4. **Time Dependency**: Vesting is based on block time, not wall clock time

## Integration with UserToken Lifecycle

Vesting accounts work seamlessly with other UserToken features:

- **Token Creation**: Create user tokens first, then set up vesting
- **Pool Creation**: Vested tokens can be used in liquidity pools once unlocked
- **Trading**: Recipients can trade vested tokens on DEX once unlocked
- **Governance**: Vested tokens may participate in governance (depending on implementation)

## Technical Implementation

The vesting functionality is implemented in:

- **Message Handler**: `x/usertoken/keeper/msg_server.go`
- **Types**: `x/usertoken/types/tx.pb.go`
- **Tests**: `x/usertoken/keeper/msg_server_vesting_test.go`

For developers integrating with this functionality, refer to the test files for usage examples and the keeper implementation for technical details.
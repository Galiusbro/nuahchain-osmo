# LimitedAccount Module Documentation

## Overview

The LimitedAccount module provides a fee management system that allows specific accounts to perform a limited number of free transactions per day. This system is designed to support two types of accounts:

1. **FreeAccount** - Unlimited free transactions (existing module)
2. **LimitedAccount** - 3 free transactions per day (new module)

## Architecture Changes

### Modified Files

#### 1. `app/ante.go` - Enhanced Ante Handler

**Key Changes:**
- Added `CombinedFreeAccountChecker` struct to handle both free and limited accounts
- Modified `NewAnteHandler` function to use combined checker
- Implemented `IsFreeAccount` method that checks both account types

**Code Structure:**
```go
// New combined checker structure
type CombinedFreeAccountChecker struct {
    freeAccountKeeper    *freeaccountkeeper.Keeper
    limitedAccountKeeper *limitedaccountkeeper.Keeper
}

// Enhanced fee decorator creation
if freeAccountKeeper != nil || limitedAccountKeeper != nil {
    combinedChecker := &CombinedFreeAccountChecker{
        freeAccountKeeper:    freeAccountKeeper,
        limitedAccountKeeper: limitedAccountKeeper,
    }
    mempoolFeeDecorator = txfeeskeeper.NewMempoolFeeDecoratorWithFreeAccounts(
        *txFeesKeeper, mempoolFeeOptions, combinedChecker)
}
```

#### 2. `x/limitedaccount/` - Complete Module Implementation

**Module Components:**
- **Proto definitions** (`proto/osmosis/limitedaccount/`)
- **Keeper logic** (`keeper/`)
- **Types and messages** (`types/`)
- **Ante decorator** (`ante/`)
- **Module registration** (`module.go`)

## Module Logic

### How LimitedAccount Works

1. **Account Registration:**
   - Accounts are registered as limited accounts through transactions
   - Each account has a daily transaction counter and reset timestamp
   - Maximum daily transactions is set to 3

2. **Transaction Processing:**
   - When a transaction is submitted, the ante handler checks if the sender is a limited account
   - If yes, it verifies if the account can still transact for free (hasn't exceeded daily limit)
   - The `CombinedFreeAccountChecker.IsFreeAccount()` method returns `true` if:
     - Account is a registered free account (unlimited), OR
     - Account is a limited account with remaining free transactions for the day

3. **Daily Reset Logic:**
   - Transaction counters reset every 24 hours
   - Reset time is tracked per account
   - When a new day begins, the counter resets to 0

4. **Fee Enforcement:**
   - First 3 transactions per day: **FREE** (0 fees required)
   - 4th+ transactions per day: **PAID** (normal fees required)
   - After daily limit: transactions are rejected with "daily transaction limit exceeded"

### Transaction Flow

```
Transaction Submitted
        |
        v
Ante Handler Checks
        |
        v
Is FreeAccount? ──YES──> Allow Free Transaction
        |
        NO
        v
Is LimitedAccount? ──NO──> Require Normal Fees
        |
        YES
        v
Check Daily Limit
        |
        v
Within Limit? ──YES──> Allow Free Transaction + Increment Counter
        |
        NO
        v
Reject Transaction ("daily transaction limit exceeded")
```

## Setup Instructions

### Prerequisites
- Go 1.21+
- Make
- Basic understanding of Cosmos SDK

### Step 1: Initialize the Chain

```bash
# Remove existing data
rm -rf ~/.nuahd

# Initialize node
./build/nuahd init validator --chain-id nuahchain-1 --home ~/.nuahd

# Create validator key
./build/nuahd keys add validator --keyring-backend test --home ~/.nuahd
```

### Step 2: Configure Genesis

```bash
# Get validator address
VALIDATOR_ADDR=$(./build/nuahd keys show validator -a --keyring-backend test --home ~/.nuahd)

# Update genesis.json - replace stake with unuah
sed -i '' 's/"stake"/"unuah"/g' ~/.nuahd/config/genesis.json

# Add genesis account with 10M NUAH
./build/nuahd genesis add-genesis-account $VALIDATOR_ADDR 10000000000000unuah --home ~/.nuahd

# Generate genesis transaction
./build/nuahd genesis gentx validator 5000000000000unuah --chain-id nuahchain-1 --keyring-backend test --home ~/.nuahd

# Collect genesis transactions
./build/nuahd genesis collect-gentxs --home ~/.nuahd

# Validate genesis
./build/nuahd genesis validate --home ~/.nuahd
```

### Step 3: Build and Start

```bash
# Build the binary
make build

# Start the node
./build/nuahd start --rpc.laddr=tcp://0.0.0.0:26657 --home ~/.nuahd
```

### Step 4: Create Test Accounts

```bash
# Create test user
./build/nuahd keys add testuser --keyring-backend test --home ~/.nuahd

# Get test user address
TEST_ADDR=$(./build/nuahd keys show testuser -a --keyring-backend test --home ~/.nuahd)

# Send tokens to test user
./build/nuahd tx bank send validator $TEST_ADDR 1000000unuah \
  --gas auto --gas-adjustment 1.5 --fees 350unuah \
  --chain-id nuahchain-1 --keyring-backend test --home ~/.nuahd --yes
```

### Step 5: Register Limited Account

```bash
# Register testuser as limited account
./build/nuahd tx limitedaccount register-limited-account \
  --from testuser --keyring-backend test --home ~/.nuahd \
  --chain-id nuahchain-1 --gas auto --gas-adjustment 1.5 \
  --fees 350unuah --yes
```

### Step 6: Test the System

```bash
# Test 1: First free transaction (should succeed with 0 fees)
./build/nuahd tx bank send testuser $VALIDATOR_ADDR 100unuah \
  --gas auto --gas-adjustment 1.5 --fees 0unuah \
  --chain-id nuahchain-1 --keyring-backend test --home ~/.nuahd --yes

# Test 2: Second free transaction (should succeed with 0 fees)
./build/nuahd tx bank send testuser $VALIDATOR_ADDR 100unuah \
  --gas auto --gas-adjustment 1.5 --fees 0unuah \
  --chain-id nuahchain-1 --keyring-backend test --home ~/.nuahd --yes

# Test 3: Third free transaction (should succeed with 0 fees)
./build/nuahd tx bank send testuser $VALIDATOR_ADDR 100unuah \
  --gas auto --gas-adjustment 1.5 --fees 0unuah \
  --chain-id nuahchain-1 --keyring-backend test --home ~/.nuahd --yes

# Test 4: Fourth transaction (should fail - limit exceeded)
./build/nuahd tx bank send testuser $VALIDATOR_ADDR 100unuah \
  --gas auto --gas-adjustment 1.5 --fees 0unuah \
  --chain-id nuahchain-1 --keyring-backend test --home ~/.nuahd --yes

# Test 5: Fourth transaction with proper fees (should fail - limit exceeded)
./build/nuahd tx bank send testuser $VALIDATOR_ADDR 100unuah \
  --gas auto --gas-adjustment 1.5 --fees 350unuah \
  --chain-id nuahchain-1 --keyring-backend test --home ~/.nuahd --yes
```

### Step 7: Query Account Status

```bash
# Check limited account status
./build/nuahd query limitedaccount limited-account $TEST_ADDR

# List all limited accounts
./build/nuahd query limitedaccount all-limited-accounts
```

## Expected Results

### Successful Setup Indicators:

1. **Node starts without errors**
2. **First 2 transactions succeed** with `code: 0` and `txhash` present
3. **Third transaction fails** with `insufficient fee` error
4. **Fourth transaction fails** with `daily transaction limit exceeded` error
5. **Account query shows** `daily_tx_count: 3` and `max_daily_txs: 3`

### Transaction Output Examples:

**Successful Free Transaction:**
```
gas estimate: 120096
code: 0
txhash: E769485BDF755DB24D2FE410145C5F69DB7B4208DD818AA9957D1C195D664CF6
```

**Failed - Insufficient Fee:**
```
code: 13
codespace: sdk
raw_log: 'Expected 1 fee denom attached, got 0: insufficient fee'
```

**Failed - Limit Exceeded:**
```
code: 2
codespace: limitedaccount
raw_log: daily transaction limit exceeded
```

## Troubleshooting

### Common Issues:

1. **"account not found" error:**
   - Ensure account has been funded with tokens
   - Check account address is correct

2. **"daily transaction limit exceeded" on first transaction:**
   - Account may have been used before
   - Wait 24 hours for reset or create new account

3. **All transactions require fees:**
   - Check if account is registered as limited account
   - Verify ante handler modifications are correct
   - Ensure binary was rebuilt after code changes

4. **Node fails to start:**
   - Check genesis.json is valid
   - Ensure all modules are properly registered
   - Verify no syntax errors in code

### Debug Commands:

```bash
# Check account balance
./build/nuahd query bank balances $TEST_ADDR

# Check if account exists
./build/nuahd query auth account $TEST_ADDR

# View node logs
tail -f ~/.nuahd/logs/node.log
```

## Configuration Options

### Customizing Daily Limits:

To change the daily transaction limit from 3 to another number:

1. Edit `x/limitedaccount/types/params.go`:
```go
const DefaultMaxDailyTxs = 5  // Change from 3 to 5
```

2. Rebuild and restart:
```bash
make build
# Restart node
```

### Gas Price Configuration:

The system uses `ConsensusMinFee` from `x/txfees/types/constants.go`:
```go
var ConsensusMinFee osmomath.Dec = osmomath.NewDecWithPrec(25, 4)  // 0.0025 unuah/gas
```

## Security Considerations

1. **Rate Limiting:** The 3-transaction daily limit prevents spam
2. **Reset Logic:** Daily resets prevent indefinite free usage
3. **Fee Enforcement:** After limit, normal fees apply
4. **Account Registration:** Only registered accounts get free transactions

## Future Enhancements

1. **Configurable Limits:** Make daily limits governance-controlled
2. **Account Types:** Different limits for different account types
3. **Time Windows:** Configurable reset periods (hourly, weekly)
4. **Fee Tiers:** Graduated fee structure after free transactions

This documentation provides a complete guide for understanding, implementing, and testing the LimitedAccount module in NuahChain.
# Gas Price Configuration Guide for NuahChain

## Overview

This guide explains how to modify gas prices in NuahChain. Gas prices determine the minimum fee users must pay for transactions.

## Understanding Gas Price Components

NuahChain uses multiple layers to determine minimum gas prices:

1. **ConsensusMinFee** - Hard-coded minimum in source code (highest priority)
2. **app.toml settings** - Configuration file settings
3. **Command line parameters** - Runtime parameters

## Step-by-Step Instructions

### Step 1: Modify the Source Code Constant

**Location:** `x/txfees/types/constants.go`

**What to change:**
```go
// BEFORE (0.01 unuah per gas unit)
var ConsensusMinFee osmomath.Dec = osmomath.NewDecWithPrec(1, 2)

// AFTER (0.0001 unuah per gas unit)
var ConsensusMinFee osmomath.Dec = osmomath.NewDecWithPrec(1, 4)
```

**Explanation:**
- `NewDecWithPrec(1, 2)` = 1 × 10^(-2) = 0.01
- `NewDecWithPrec(1, 4)` = 1 × 10^(-4) = 0.0001

### Step 2: Configure app.toml Settings

**Location:** `~/.nuahd/config/app.toml`

**Settings to modify:**
```toml
# Basic minimum gas price
minimum-gas-prices = "0.0001unuah"

# Osmosis mempool settings
[osmosis-mempool]
arbitrage-min-gas-fee = "0.0001"
min-gas-price-for-high-gas-tx = "0.0001"
```

### Step 3: Rebuild the Binary

```bash
make build
```

### Step 4: Restart the Node

```bash
# Stop current node (if running)
# Then start with new settings
./build/nuahd start --rpc.laddr=tcp://0.0.0.0:26657 --home ~/.nuahd --minimum-gas-prices=0.0001unuah
```

## Testing Your Changes

### Test with Low Fee Transaction

```bash
# This should work with the new settings
./build/nuahd tx bank send validator [recipient] 100unuah \
  --gas auto --gas-adjustment 1.5 --fees 12unuah \
  --home ~/.nuahd --keyring-backend test --chain-id nuahchain-1 --yes
```

### Test with Very Low Fee (Should Fail)

```bash
# This should fail - fee too low
./build/nuahd tx bank send validator [recipient] 100unuah \
  --gas auto --gas-adjustment 1.5 --fees 1unuah \
  --home ~/.nuahd --keyring-backend test --chain-id nuahchain-1 --yes
```

## Gas Price Calculation

**Formula:** `Required Fee = Gas Used × Gas Price`

**Example:**
- Gas Used: 118,000 units
- Gas Price: 0.0001 unuah/gas
- Required Fee: 118,000 × 0.0001 = 11.8 unuah (rounded up to 12 unuah)

## Common Gas Price Values

| Description | Value | Fee for 100k Gas |
|-------------|-------|-------------------|
| Very Low | 0.0001 | 10 unuah |
| Low | 0.001 | 100 unuah |
| Medium | 0.01 | 1,000 unuah |
| High | 0.1 | 10,000 unuah |

## Troubleshooting

### Problem: Changes Don't Take Effect

**Solution:**
1. Ensure you modified the source code constant
2. Rebuild the binary with `make build`
3. Restart the node completely
4. Check that app.toml settings match

### Problem: Transactions Still Rejected

**Possible Causes:**
1. ConsensusMinFee not changed in source code
2. Binary not rebuilt after code changes
3. Node not restarted with new binary
4. Fee calculation error

### Problem: Node Won't Start

**Check:**
1. Syntax errors in app.toml
2. Invalid gas price format
3. Missing dependencies after rebuild

## Important Notes

⚠️ **Warning:** Always test changes on a development network first!

📝 **Remember:** 
- Source code changes require rebuilding the binary
- Configuration changes require node restart
- The ConsensusMinFee has the highest priority

🔍 **Verification:**
Always test with actual transactions to confirm your changes work as expected.

## Advanced Configuration

### Different Fees for Different Transaction Types

```toml
[osmosis-mempool]
# Regular transactions
minimum-gas-prices = "0.0001unuah"

# Arbitrage transactions (higher fee)
arbitrage-min-gas-fee = "0.001"

# High gas transactions (higher fee)
min-gas-price-for-high-gas-tx = "0.0005"
```

### EIP-1559 Style Dynamic Fees

```toml
[osmosis-mempool]
# Enable dynamic fee adjustment
adaptive-fee-enabled = "true"
```

This enables automatic fee adjustment based on network congestion.
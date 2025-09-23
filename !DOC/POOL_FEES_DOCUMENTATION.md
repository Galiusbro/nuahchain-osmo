# Pool Fees in Osmosis/GAMM: Complete Guide

## Overview

This document explains how pool fees (spread factors) work in Osmosis/GAMM and how to modify them. Based on our investigation of the codebase, here's everything you need to know.

## How Pool Fees Work

### Fee Collection Process

1. **Fee Deduction**: When a swap occurs, a fee (typically 0.02% or 0.2%) is deducted from the input token
2. **Fee Storage**: The deducted fees are sent to a special address (`GetSpreadRewardsAddress`) for each pool
3. **Fee Accumulation**: Fees accumulate in this address and are tracked using accumulators
4. **Fee Distribution**: Fees are distributed proportionally to liquidity providers based on their share of the pool

### Fee Collection by Liquidity Providers

Liquidity providers can collect their accumulated fees using:
- **Transaction**: `MsgCollectSpreadRewards`
- **CLI Command**: `osmosisd tx concentratedliquidity collect-spread-rewards`

## How to Change Pool Fees

### Method 1: Set Fees When Creating a New Pool (✅ Available Now)

**For Beginners - Step by Step:**

1. **Create a pool configuration file** (e.g., `my-pool.json`):
   ```json
   {
     "weights": "5utoken1,5utoken2",
     "initial-deposit": "1000000utoken1,1000000utoken2",
     "swap-fee": "0.003",
     "exit-fee": "0.000",
     "future-governor": ""
   }
   ```

2. **Understand the swap-fee parameter**:
   - `"0.003"` = 0.3% fee
   - `"0.002"` = 0.2% fee (default)
   - `"0.001"` = 0.1% fee
   - `"0.005"` = 0.5% fee

3. **Create the pool**:
   ```bash
   osmosisd tx gamm create-pool my-pool.json --from your-wallet --chain-id osmosis-1
   ```

**Important Notes:**
- ⚠️ **Once a pool is created, you CANNOT change the swap fee directly**
- Choose your fee carefully before creating the pool
- Higher fees = more rewards for liquidity providers, but less attractive for traders
- Lower fees = more trading volume, but less rewards per trade

### Method 2: Future Pool Governor (🚧 Planned Feature)

The codebase includes a "Future Pool Governor" mechanism that would allow changing pool parameters after creation:

**Three governance options:**
1. **No governance**: Pool parameters are immutable
2. **Address-based governance**: A specific address can modify parameters
3. **Token-based DAO governance**: Token holders can vote on parameter changes

**Current Status**: This feature is defined in the code but not yet implemented.

### Method 3: Governance Proposals (❌ Not Available for Swap Fees)

While Osmosis supports governance proposals for various parameters, swap fees are not currently modifiable through governance.

## Practical Recommendations

### For Pool Creators

1. **Research similar pools**: Check what fees other pools with similar tokens are using
2. **Consider your strategy**:
   - **High volume tokens**: Use lower fees (0.1-0.2%) to attract more trades
   - **Exotic/risky tokens**: Use higher fees (0.3-0.5%) to compensate liquidity providers
3. **Test with small amounts**: Create a test pool first if possible

### For Existing Pools

If you need to change fees for an existing pool:

1. **Wait for Future Pool Governor**: Monitor Osmosis updates for this feature
2. **Create a new pool**: 
   - Create a new pool with desired fees
   - Migrate liquidity from old pool to new pool
   - Coordinate with other liquidity providers
3. **Community coordination**: Discuss with the community about creating a replacement pool

## Code References

### Key Files
- `concentrated-liquidity/swaps.go`: Fee calculation and distribution
- `concentrated-liquidity/spread_rewards.go`: Fee accumulation logic
- `gamm/pool-models/balancer/pool_params.go`: Pool parameter definitions
- `gamm/README.md`: Documentation on pool parameters

### Key Functions
- `GetSpreadRewardsAddress()`: Returns the fee collection address for a pool
- `CollectSpreadRewards()`: Allows liquidity providers to claim fees
- `NewPoolParams()`: Creates pool parameters including swap fees

## Frequently Asked Questions

**Q: Can I change the swap fee after creating a pool?**
A: Currently, no. Swap fees are set at pool creation and cannot be modified.

**Q: What's a good swap fee percentage?**
A: It depends on your tokens and strategy. Common ranges are 0.1% to 0.5%. Research similar pools for guidance.

**Q: How do liquidity providers collect fees?**
A: Use the `MsgCollectSpreadRewards` transaction or the CLI command `collect-spread-rewards`.

**Q: When will Future Pool Governor be available?**
A: This feature is planned but not yet implemented. Monitor Osmosis development updates.

**Q: Can governance change swap fees?**
A: No, governance proposals cannot currently modify individual pool swap fees.

## Summary

- **Current reality**: Swap fees can only be set when creating a pool
- **Future possibility**: Future Pool Governor will allow post-creation modifications
- **Workaround**: Create new pools with desired fees and migrate liquidity
- **Best practice**: Choose fees carefully during pool creation

This documentation reflects the current state of the Osmosis/GAMM codebase and may change as new features are implemented.
# PegKeeper Module Documentation

## What is PegKeeper?

Imagine you have a special coin that should always be worth exactly $1. But sometimes, due to market forces, this coin might be worth $1.05 or $0.95. The PegKeeper module is like a smart robot that watches your coin's price and automatically adjusts the supply to keep it as close to $1 as possible.

**Simple analogy**: Think of it like a thermostat in your house. When it gets too hot, the AC turns on. When it gets too cold, the heater turns on. PegKeeper does the same thing but with coin supply and price!

## How Does It Work?

### The Basic Concept

1. **Target Price**: The price we want our coin to maintain (e.g., $1.00)
2. **Current Price**: The actual market price of our coin
3. **Deviation**: How far the current price is from the target price
4. **Action**: If the deviation is too big, PegKeeper takes action

### What Actions Can PegKeeper Take?

- **Price too HIGH** (e.g., $1.10): Create more coins (increase supply) → Price goes down
- **Price too LOW** (e.g., $0.90): Remove coins from circulation (decrease supply) → Price goes up

## Module Parameters

These are the "settings" that control how PegKeeper behaves:

### Core Parameters

| Parameter | What It Does | Example Value | Simple Explanation |
|-----------|--------------|---------------|--------------------|
| `target_denom` | The coin we're stabilizing | `factory/nuah1.../ndollar` | Which coin to watch |
| `reference_denom` | What we compare against | `uusd` | What currency to compare to |
| `target_price` | The ideal price | `1.0` | We want 1 coin = $1 |
| `max_deviation_threshold` | When to take action | `0.05` | Act if price moves 5% away |
| `adjustment_factor` | How much to adjust | `0.1` | Adjust supply by 10% max |
| `min_adjustment_interval` | How often to check | `3600` | Wait 1 hour between actions |
| `max_supply_change_per_adjustment` | Safety limit | `0.02` | Never change supply by more than 2% |
| `oracle_module` | Where to get price data | `usdoracle` | Which module provides prices |
| `enabled` | Turn on/off | `true` | Is PegKeeper active? |

## Real-World Example

Let's say we have a stablecoin called "N$" that should be worth $1:

### Scenario 1: Price Too High
```
Current situation:
- N$ market price: $1.08
- Target price: $1.00
- Deviation: 8% (above threshold of 5%)

PegKeeper's action:
1. "Price is too high, people are paying too much!"
2. Create new N$ tokens (increase supply)
3. More tokens in market → price comes down
4. Goal: Bring price back to ~$1.00
```

### Scenario 2: Price Too Low
```
Current situation:
- N$ market price: $0.93
- Target price: $1.00
- Deviation: 7% (above threshold of 5%)

PegKeeper's action:
1. "Price is too low, tokens are too cheap!"
2. Remove N$ tokens from circulation (decrease supply)
3. Fewer tokens available → price goes up
4. Goal: Bring price back to ~$1.00
```

### Scenario 3: Price Just Right
```
Current situation:
- N$ market price: $1.02
- Target price: $1.00
- Deviation: 2% (below threshold of 5%)

PegKeeper's action:
1. "Price is close enough, no action needed"
2. Wait and monitor
3. Check again after minimum interval
```

## CLI Commands

### Query Commands (Check Information)

#### 1. Check Current Parameters
```bash
./build/nuahd query pegkeeper params
```
**What it shows**: All the current settings of PegKeeper

#### 2. Check Current Peg State
```bash
./build/nuahd query pegkeeper peg-state
```
**What it shows**:
- Current price of your coin
- How far it is from target (deviation)
- Whether PegKeeper is actively working
- When it last made an adjustment

#### 3. Check Adjustment History
```bash
./build/nuahd query pegkeeper adjustment-history
```
**What it shows**: A list of all the times PegKeeper took action, including:
- When it happened
- How much supply was changed
- What the price was before and after

### Transaction Commands (Make Changes)

#### Update Parameters (Governance Only)
```bash
./build/nuahd tx pegkeeper update-params \
  [target_denom] \
  [reference_denom] \
  [max_deviation_threshold] \
  [adjustment_factor] \
  [min_adjustment_interval] \
  [max_supply_change_per_adjustment] \
  [oracle_module] \
  [enabled] \
  [target_price] \
  --from [your-key] \
  --chain-id [chain-id]
```

**Example**:
```bash
./build/nuahd tx pegkeeper update-params \
  "factory/nuah1tvnraf3rgdajgtfu2rz9jxatxxkp8qqquzq7dp/ndollar" \
  "uusd" \
  "0.05" \
  "0.1" \
  "3600" \
  "0.02" \
  "usdoracle" \
  true \
  "1.0" \
  --from validator \
  --chain-id nuahchain
```

## Integration with Other Modules

### USDOracle Module
PegKeeper needs to know the current price of your coin. It gets this information from the USDOracle module, which:
- Fetches price data from external sources
- Provides reliable price feeds
- Updates prices regularly

### TokenFactory Module
When PegKeeper needs to adjust supply, it uses the TokenFactory module to:
- Mint new tokens (increase supply)
- Burn existing tokens (decrease supply)

## Safety Features

### 1. Maximum Deviation Threshold
- PegKeeper only acts when price moves significantly
- Prevents unnecessary adjustments for small price movements
- Example: Only act if price moves more than 5%

### 2. Adjustment Factor Limits
- Limits how much supply can be changed at once
- Prevents dramatic market disruptions
- Example: Never adjust supply by more than 10% at once

### 3. Time Intervals
- Prevents too frequent adjustments
- Gives market time to react to changes
- Example: Wait at least 1 hour between adjustments

### 4. Maximum Supply Change
- Additional safety limit on supply changes
- Even stricter than adjustment factor
- Example: Never change total supply by more than 2%

### 5. Enable/Disable Switch
- Can turn PegKeeper on or off completely
- Useful for maintenance or emergencies

## Common Use Cases

### 1. Stablecoin Management
- Most common use case
- Maintain $1 peg for stablecoins
- Automatic supply adjustments

### 2. Algorithmic Central Banking
- Act like a central bank for your token
- Maintain price stability
- Respond to market conditions

### 3. DeFi Protocol Integration
- Provide stable assets for DeFi protocols
- Enable reliable lending/borrowing
- Support liquidity provision

## Troubleshooting

### Problem: PegKeeper Not Working
**Check**:
1. Is `enabled` set to `true`?
2. Is the oracle providing price data?
3. Has enough time passed since last adjustment?
4. Is the deviation above the threshold?

### Problem: Too Frequent Adjustments
**Solution**:
1. Increase `min_adjustment_interval`
2. Increase `max_deviation_threshold`
3. Decrease `adjustment_factor`

### Problem: Price Still Unstable
**Solution**:
1. Decrease `max_deviation_threshold` (act sooner)
2. Increase `adjustment_factor` (stronger actions)
3. Check oracle price accuracy
4. Verify sufficient market liquidity

## Best Practices

### 1. Start Conservative
- Begin with small adjustment factors
- Use higher deviation thresholds
- Monitor closely before making aggressive changes

### 2. Monitor Regularly
- Check peg state frequently
- Review adjustment history
- Watch for unusual market conditions

### 3. Test Thoroughly
- Test on testnet first
- Start with small amounts
- Gradually increase parameters

### 4. Have Emergency Plans
- Know how to disable PegKeeper quickly
- Have governance procedures ready
- Monitor external price feeds

## Example Workflow

### Setting Up a New Stablecoin

1. **Create the Token**
   ```bash
   # Create N$ token using tokenfactory
   ./build/nuahd tx tokenfactory create-denom ndollar --from creator
   ```

2. **Configure USDOracle**
   ```bash
   # Set up price feeds for your token
   ./build/nuahd tx usdoracle update-params true admin 60 0.05
   ```

3. **Configure PegKeeper**
   ```bash
   # Set PegKeeper parameters for N$ token
   ./build/nuahd tx pegkeeper update-params \
     "factory/creator/ndollar" "uusd" "0.05" "0.1" "3600" "0.02" "usdoracle" true "1.0"
   ```

4. **Monitor and Adjust**
   ```bash
   # Check if it's working
   ./build/nuahd query pegkeeper peg-state
   ./build/nuahd query pegkeeper adjustment-history
   ```

## Conclusion

The PegKeeper module is a powerful tool for maintaining price stability in blockchain tokens. By automatically adjusting supply based on price deviations, it helps create reliable stablecoins and stable assets for DeFi ecosystems.

Remember:
- Start with conservative settings
- Monitor regularly
- Test thoroughly
- Have safety measures in place

With proper configuration and monitoring, PegKeeper can help maintain stable token prices automatically, reducing the need for manual intervention and providing a better user experience for your token holders.

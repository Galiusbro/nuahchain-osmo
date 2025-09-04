# Token Trading Guide 🚀

**Quick start guide for trading custom tokens on Nuah Chain**

## 🎯 Getting Started

### What You Need
- NUAH tokens for trading and fees
- Access to a Nuah Chain node
- Basic understanding of AMM trading

### Available Trading Pairs
- **NUAH ↔ MyToken2**: Pool ID 1, 0.3% fee
- *More pools coming soon...*

## 💱 How to Trade

### Buy MyToken2 with NUAH

```bash
# Buy MyToken2 with 0.05 NUAH (50,000 unuah)
nuahd tx poolmanager swap-exact-amount-in \
  50000unuah \
  1 \
  --swap-route-pool-ids=1 \
  --swap-route-denoms=factory/nuah14k38ajalnef2yauznt4q7ep893djkl4vm54mcs/mytoken2 \
  --from YOUR_ACCOUNT \
  --chain-id nuahchain-1 \
  --keyring-backend test \
  --gas 400000 \
  --fees 4000unuah \
  -y
```

**Expected Result:**
- **Input**: 0.05 NUAH (50,000 unuah)
- **Output**: ~23,741 MyToken2
- **Fee**: 0.3% (150 unuah)
- **Total Cost**: 54,000 unuah

### Sell MyToken2 for NUAH

```bash
# Sell 10,000 MyToken2 for NUAH
nuahd tx poolmanager swap-exact-amount-in \
  10000factory/nuah14k38ajalnef2yauznt4q7ep893djkl4vm54mcs/mytoken2 \
  1 \
  --swap-route-pool-ids=1 \
  --swap-route-denoms=unuah \
  --from YOUR_ACCOUNT \
  --chain-id nuahchain-1 \
  --keyring-backend test \
  --gas 400000 \
  --fees 4000unuah \
  -y
```

**Expected Result:**
- **Input**: 10,000 MyToken2
- **Output**: ~21,529 unuah
- **Current Price**: ~2.15 NUAH per MyToken2

## 📊 Live Market Data

### Current Pool Status (Pool ID 1)

```bash
# Check current pool state
nuahd query gamm pool 1
```

**Live Stats:**
- **MyToken2 Liquidity**: 486,259 tokens
- **NUAH Liquidity**: 1,028,471 unuah  
- **Current Price**: ~2.11 NUAH per MyToken2
- **24h Volume**: Active trading
- **Swap Fee**: 0.3%

### Check Your Balance

```bash
# View your token balances
nuahd query bank balances YOUR_ADDRESS
```

## 💡 Trading Tips

### 🎯 Slippage Protection
Always set a reasonable minimum output:
```bash
# For 1% slippage tolerance on 50,000 unuah trade:
# Expected output: ~23,741 MyToken2
# Min output (1% slippage): 23,504 MyToken2
--swap-exact-amount-in 50000unuah 23504
```

### 💰 Fee Optimization
- **Small trades**: Use higher gas to ensure execution
- **Large trades**: Consider splitting to reduce slippage
- **Frequent trading**: Monitor cumulative fees

### 📈 Price Monitoring
```bash
# Get current exchange rate
POOL_DATA=$(nuahd query gamm pool 1 -o json)
NUAH_AMT=$(echo $POOL_DATA | jq -r '.pool.pool_assets[1].token.amount')
TOKEN_AMT=$(echo $POOL_DATA | jq -r '.pool.pool_assets[0].token.amount')
PRICE=$(echo "scale=6; $NUAH_AMT / $TOKEN_AMT" | bc)
echo "Current price: $PRICE NUAH per MyToken2"
```

## 🔄 Advanced Trading

### Multi-Hop Trading
*Coming soon: Trade through multiple pools*

### Limit Orders
*Future feature: Set buy/sell orders at specific prices*

### Arbitrage Opportunities
Watch for price differences between:
- Different pools
- Cross-chain bridges
- Centralized exchanges

## 📱 Web Interface

### Using Keplr Wallet
1. Connect Keplr to Nuah Chain
2. Navigate to DEX interface  
3. Select NUAH/MyToken2 pair
4. Enter trade amount
5. Confirm transaction

### API Integration
```javascript
// Example swap transaction
const swapMsg = {
  typeUrl: "/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn",
  value: {
    sender: userAddress,
    routes: [{
      poolId: "1",
      tokenOutDenom: "factory/nuah14k38ajalnef2yauznt4q7ep893djkl4vm54mcs/mytoken2"
    }],
    tokenIn: { denom: "unuah", amount: "50000" },
    tokenOutMinAmount: "23000"
  }
};
```

## 🏆 Success Examples

### Real Trading Results

**Trade #1: NUAH → MyToken2**
```
Input:  100,000 unuah (0.1 NUAH)
Output: 47,619 MyToken2  
Price:  2.10 NUAH per MyToken2
Fee:    300 unuah (0.3%)
Status: ✅ Success
```

**Trade #2: NUAH → MyToken2**
```  
Input:  50,000 unuah (0.05 NUAH)
Output: 23,741 MyToken2
Price:  2.11 NUAH per MyToken2  
Fee:    150 unuah (0.3%)
Status: ✅ Success
```

**Trade #3: MyToken2 → NUAH**
```
Input:  10,000 MyToken2
Output: 21,529 unuah
Price:  2.15 NUAH per MyToken2
Fee:    65 MyToken2 (0.3%)
Status: ✅ Success
```

## ⚠️ Risk Management

### Common Pitfalls
- **High slippage**: Use reasonable trade sizes
- **Front-running**: Consider mempool conditions  
- **Price impact**: Large trades affect prices
- **Gas estimation**: Always have extra NUAH for fees

### Safety Checklist
- [ ] Double-check token addresses
- [ ] Set appropriate slippage tolerance
- [ ] Have sufficient gas fees
- [ ] Start with small test trades
- [ ] Monitor market conditions

## 🆘 Troubleshooting

### Failed Transactions

**"insufficient fees"**
```bash
# Increase fee amount
--fees 8000unuah  # Try double the original fee
```

**"out of gas"**
```bash  
# Increase gas limit
--gas 600000  # Try 50% more gas
```

**"swap slippage exceeded"**
```bash
# Lower minimum output or increase slippage tolerance
# For 5% slippage: min_out = expected_out * 0.95
```

### Price Issues

**"Pool not found"**
- Verify pool ID exists: `nuahd query gamm pools`
- Check pool is active and has liquidity

**"Token not found"**  
- Confirm token denomination spelling
- Use full factory address format

## 📞 Support

### Community Resources
- **Discord**: Real-time trading discussion
- **Telegram**: Price alerts and updates
- **Forum**: Strategy sharing
- **Documentation**: Technical guides

### Developer Support
- **GitHub Issues**: Bug reports
- **API Docs**: Integration help
- **SDK Examples**: Code samples

---

## 🎉 Start Trading!

**Your trading journey begins now:**

1. **Check your balance**: Ensure you have NUAH
2. **Start small**: Try a 0.01 NUAH trade first  
3. **Monitor prices**: Watch for good entry points
4. **Trade responsibly**: Only risk what you can afford
5. **Join the community**: Share strategies and tips

**Happy Trading!** 🚀💰

*Current Market: NUAH/MyToken2 @ ~2.11 - Active liquidity, low slippage*
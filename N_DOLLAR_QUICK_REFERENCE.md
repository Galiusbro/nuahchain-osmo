# N$ Token Quick Reference Card

**Essential commands for N$ token management on Nuah Chain**

## 🚀 Quick Setup Commands

```bash
# 1. Build node
make build

# 2. Start node (background)
./build/nuahd start --rpc.laddr=tcp://0.0.0.0:26657 --home ~/.nuahd &

# 3. Create N$ token
./build/nuahd tx tokenfactory create-denom ndollar --from validator --chain-id nuahchain-1 --keyring-backend test --gas 1500000 --fees 16000unuah -y

# 4. Set metadata
./build/nuahd tx tokenfactory set-denom-metadata '{"description":"N Dollar - algorithmic stablecoin","denom_units":[{"denom":"factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar","exponent":0,"aliases":["ndollar"]},{"denom":"NDOLLAR","exponent":6,"aliases":["ndollar"]}],"base":"factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar","display":"NDOLLAR","name":"N Dollar","symbol":"N$"}' --from validator --chain-id nuahchain-1 --keyring-backend test --gas 300000 --fees 3000unuah -y

# 5. Mint tokens
./build/nuahd tx tokenfactory mint 1000000factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d --from validator --chain-id nuahchain-1 --keyring-backend test --gas 300000 --fees 3000unuah -y

# 6. Create pool
./build/nuahd tx gamm create-pool --pool-file=ndollar-pool-config.json --from validator --chain-id nuahchain-1 --keyring-backend test --gas 1500000 --fees 16000unuah -y
```

## 📊 Query Commands

```bash
# Check node status
curl http://localhost:26657/status

# List keys
./build/nuahd keys list --keyring-backend test

# Check balances
./build/nuahd query bank balances nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d

# List created tokens
./build/nuahd query tokenfactory denoms-from-creator nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d

# Check token metadata
./build/nuahd query bank denom-metadata factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar

# List pools
./build/nuahd query gamm pools

# Check specific pool
./build/nuahd query gamm pool 1

# Check pool params
./build/nuahd query gamm pool-params 1

# Get spot price
./build/nuahd query gamm spot-price 1 unuah factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar
```

## 💱 Trading Commands

```bash
# Swap NUAH for N$
./build/nuahd tx poolmanager swap-exact-amount-in 100000unuah 90000 --swap-route-pool-ids=1 --swap-route-denoms=factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar --from validator --chain-id nuahchain-1 --keyring-backend test --gas 1000000 --fees 10000unuah -y

# Swap N$ for NUAH
./build/nuahd tx poolmanager swap-exact-amount-in 100000factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar 90000 --swap-route-pool-ids=1 --swap-route-denoms=unuah --from validator --chain-id nuahchain-1 --keyring-backend test --gas 1000000 --fees 10000unuah -y

# Add liquidity
./build/nuahd tx gamm join-pool 1 "50000unuah,50000factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar" 1000000000000000000 --from validator --chain-id nuahchain-1 --keyring-backend test --gas 1000000 --fees 10000unuah -y
```

## 📈 TWAP Oracle Commands

```bash
# Query current TWAP price (N$ to USD)
./build/nuahd query twap arithmetic 1 factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar $(date -v-30S +%s) 30s

# Query reverse TWAP (NUAH to N$)
./build/nuahd query twap arithmetic 1 unuah $(date -v-30S +%s) 30s

# Check TWAP module parameters
./build/nuahd query twap params

# Example output:
# arithmetic_twap: "1.001025908322333333"
```

## 🌐 API Monitoring Endpoints

```bash
# Start API server
go run cmd/api-server/main.go

# Available endpoints:
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/ndollar/price
curl http://localhost:8080/api/v1/ndollar/twap?period=1h
curl http://localhost:8080/api/v1/ndollar/metrics
curl http://localhost:8080/api/v1/ndollar/supply
```

## 💳 Fee Abstraction Setup

```bash
# Run fee abstraction setup script
./scripts/setup_fee_abstraction.sh

# Example fee payment with N$ (after governance approval)
./build/nuahd tx bank send [from] [to] 1000unuah \
  --fees 1000factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar \
  --chain-id nuahchain-1 --keyring-backend test -y
```

## 🔧 Pool Configuration Template

**File: `ndollar-pool-config.json`**
```json
{
  "weights": "1000000unuah,1000000factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar",
  "initial-deposit": "1000000unuah,1000000factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar",
  "swap-fee": "0.005",
  "exit-fee": "0.000",
  "future-governor": ""
}
```

## 📋 Key Information

| Parameter | Value |
|-----------|-------|
| **Token Denom** | `factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar` |
| **Display Name** | N Dollar |
| **Symbol** | N$ |
| **Pool ID** | 1 |
| **Initial Ratio** | 1:1 (NUAH:N$) |
| **Swap Fee** | 0.5% |
| **Current TWAP** | ~1.001 USD |
| **API Server** | http://localhost:8080 |
| **Validator Address** | `nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d` |

## 🚨 Common Issues

| Problem | Solution |
|---------|----------|
| `command not found: nuahd` | Run `make build` first |
| `connection refused` | Start node with `./build/nuahd start` |
| `insufficient funds` | Check balance with query commands |
| `invalid denom` | Use full factory denom path |
| `pool not found` | Verify pool creation with `query gamm pools` |

## 🎯 Success Indicators

- ✅ Node running on port 26657
- ✅ Validator key exists and funded
- ✅ N$ token created with metadata
- ✅ 1M N$ tokens minted
- ✅ Pool ID 1 with NUAH/N$ pair
- ✅ Trading functionality active
- ✅ TWAP oracle operational
- ✅ API monitoring server running
- ✅ Fee abstraction configured
- ✅ Documentation complete

---
**Quick Start**: Copy commands → Replace addresses → Execute in order
**Full Guide**: See `N_DOLLAR_SETUP_GUIDE.md` for detailed explanations

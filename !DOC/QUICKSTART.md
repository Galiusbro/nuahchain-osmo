# Nuah Chain Quick Start

## 🚀 Fast Track Setup

### 1. Build & Initialize
```bash
# Build the binary
make build

# Initialize node
./build/nuahd init test-node --chain-id nuahchain-1

# Create validator key (SAVE THE MNEMONIC!)
./build/nuahd keys add validator --keyring-backend test
```

### 2. Configure NUAH Token
```bash
# Replace stake with unuah
sed -i '' 's/"stake"/"unuah"/g' ~/.nuahd/config/genesis.json

# Add genesis account (replace ADDRESS with your validator address)
./build/nuahd add-genesis-account [VALIDATOR_ADDRESS] 10000000000000unuah --keyring-backend test

# Create genesis transaction
./build/nuahd gentx validator 5000000000000unuah --chain-id nuahchain-1 --keyring-backend test

# Collect transactions
./build/nuahd collect-gentxs
```

### 3. Start Node
```bash
# Start in background
./build/nuahd start --rpc.laddr=tcp://0.0.0.0:26657 &
```

### 4. Verify
```bash
# Wait 5 seconds then check status
sleep 5 && curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height'

# Check balance
./build/nuahd query bank balances [VALIDATOR_ADDRESS]
```

## 🎉 Success!

If you see:
- ✅ Block height increasing
- ✅ Balance: `5000000000000 unuah` (5M NUAH)
- ✅ RPC responding on port 26657

Your NUAH node is running! 

**Token Info**: 1 NUAH = 1,000,000 unuah

## 📍 Endpoints
- **RPC**: http://localhost:26657
- **REST**: http://localhost:1317  
- **gRPC**: localhost:9090

## 🆘 Problems?
See detailed [SETUP_GUIDE.md](./SETUP_GUIDE.md) for troubleshooting.
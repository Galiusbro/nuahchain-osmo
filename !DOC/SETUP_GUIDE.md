# Nuah Chain Setup Guide

## Overview

This guide walks you through setting up and running an Osmosis-based blockchain node with the NUAH native token. The node will run locally for development and testing purposes.

## Prerequisites

- **Go**: Version 1.21+ installed
- **Make**: Build tools
- **Git**: For cloning repositories
- **jq**: JSON processor for testing (optional but recommended)

## Project Structure

```
nuahchain/
├── .vscode/           # VS Code configuration
├── build/             # Build artifacts
├── ~/.nuahd/         # Node data directory
├── token_config.json  # Token configuration
└── SETUP_GUIDE.md    # This file
```

## 1. Initial Setup

### Clone and Build

```bash
# Navigate to project directory
cd /path/to/nuahchain

# Build the binary
make build
```

The binary will be created at `./build/nuahd`.

### VS Code Configuration

The project includes VS Code tasks for development:

- **Build**: `Cmd+Shift+P` → `Tasks: Run Task` → `build`
- **Test**: `Cmd+Shift+P` → `Tasks: Run Task` → `test`
- **Lint**: `Cmd+Shift+P` → `Tasks: Run Task` → `lint`

## 2. Node Initialization

### Initialize the Node

```bash
# Initialize node with chain ID
./build/nuahd init test-node --chain-id nuahchain-1
```

This creates the initial configuration and genesis file.

### Create Validator Key

```bash
# Create validator key (save the mnemonic safely!)
./build/nuahd keys add validator --keyring-backend test
```

**Example output:**
```
- address: nuah14k38ajalnef2yauznt4q7ep893djkl4vm54mcs
  name: validator
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A5F64yN7aDbvkUjUkSxOLppjlhrz2J88wpM4eB69dq09"}'
  type: local

**Important** write this mnemonic phrase in a safe place.
vintage coconut sick drive muffin sort solar occur valid daring play melt gasp pill banner clown learn nominee hold live bachelor ecology solve hunt
```

### Configure NUAH Token

Replace the default "stake" denomination with "unuah":

```bash
# Backup genesis file
cp ~/.nuahd/config/genesis.json ~/.nuahd/config/genesis.json.backup

# Replace stake with unuah denomination
sed -i '' 's/"stake"/"unuah"/g' ~/.nuahd/config/genesis.json
```

### Add Genesis Account

```bash
# Add genesis account with 10 million NUAH (10 trillion unuah)
./build/nuahd add-genesis-account [VALIDATOR_ADDRESS] 10000000000000unuah --keyring-backend test
```

Replace `[VALIDATOR_ADDRESS]` with the address from the previous step.

### Create Genesis Transaction

```bash
# Create genesis transaction with 5 million NUAH stake
./build/nuahd gentx validator 5000000000000unuah --chain-id nuahchain-1 --keyring-backend test
```

### Collect Genesis Transactions

```bash
# Collect all genesis transactions
./build/nuahd collect-gentxs
```

## 3. Start the Node

### Background Mode

```bash
# Start node in background
./build/nuahd start --rpc.laddr=tcp://0.0.0.0:26657 &
```

### Foreground Mode

```bash
# Start node in foreground (for debugging)
./build/nuahd start --rpc.laddr=tcp://0.0.0.0:26657
```

## 4. Verify Installation

### Check Node Status

```bash
# Check if node is running
curl -s http://localhost:26657/status | jq

# Check current block height
curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height'
```

### Check Validator Balance

```bash
# Query validator balance
./build/nuahd query bank balances [VALIDATOR_ADDRESS]
```

**Expected output:**
```yaml
balances:
- amount: "5000000000000"
  denom: unuah
pagination:
  total: "1"
```

### Check Validator Information

```bash
# Check validator status
./build/nuahd query staking validators
```

## 5. API Endpoints

Once the node is running, the following endpoints are available:

| Service | URL | Description |
|---------|-----|-------------|
| **RPC** | `http://localhost:26657` | Tendermint RPC |
| **REST API** | `http://localhost:1317` | Cosmos REST API |
| **gRPC** | `localhost:9090` | gRPC endpoint |

### Example API Calls

```bash
# Node info
curl -s http://localhost:26657/status

# Health check
curl -s http://localhost:26657/health

# Account balance via REST API
curl -s http://localhost:1317/cosmos/bank/v1beta1/balances/[ADDRESS]

# Validators
curl -s http://localhost:26657/validators

# Network info
curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info
```

## 6. Token Configuration

### NUAH Token Details

| Parameter | Value |
|-----------|-------|
| **Symbol** | NUAH |
| **Name** | Nuah |
| **Base Denomination** | unuah |
| **Display Denomination** | NUAH |
| **Decimals** | 6 |
| **Conversion** | 1 NUAH = 1,000,000 unuah |

### Wallet Integration

The `token_config.json` file contains wallet integration parameters:

```json
{
  "chain_id": "nuahchain-1",
  "network_name": "Nuah Chain",
  "native_token": {
    "symbol": "NUAH",
    "display_name": "Nuah",
    "base_denom": "unuah",
    "display_denom": "NUAH",
    "decimals": 6,
    "conversion_rate": "1000000"
  }
}
```

## 7. Common Commands

### Node Management

```bash
# Stop the node
pkill nuahd

# Restart the node
./build/nuahd start --rpc.laddr=tcp://0.0.0.0:26657

# Reset node data (WARNING: destroys all data)
./build/nuahd unsafe-reset-all
```

### Key Management

```bash
# List all keys
./build/nuahd keys list --keyring-backend test

# Show specific key
./build/nuahd keys show validator --keyring-backend test

# Export key
./build/nuahd keys export validator --keyring-backend test
```

### Query Commands

```bash
# Check account balance
./build/nuahd query bank balances [ADDRESS]

# Check validator info
./build/nuahd query staking validator [VALIDATOR_ADDRESS]

# Check delegation
./build/nuahd query staking delegation [DELEGATOR_ADDRESS] [VALIDATOR_ADDRESS]

# Check network parameters
./build/nuahd query staking params
```

### Transaction Commands

```bash
# Send tokens
./build/nuahd tx bank send [FROM_ADDRESS] [TO_ADDRESS] [AMOUNT]unuah --chain-id nuahchain-1 --keyring-backend test

# Delegate tokens
./build/nuahd tx staking delegate [VALIDATOR_ADDRESS] [AMOUNT]unuah --from validator --chain-id nuahchain-1 --keyring-backend test
```

## 8. Troubleshooting

### Common Issues

#### Node Won't Start

1. **Check if port is in use:**
   ```bash
   lsof -i :26657
   ```

2. **Check genesis file validity:**
   ```bash
   ./build/nuahd validate-genesis
   ```

3. **Reset and reinitialize:**
   ```bash
   rm -rf ~/.nuahd
   # Follow initialization steps again
   ```

#### Empty Validator Set Error

This typically occurs when the staking amount is too low. Ensure you:
1. Added genesis account with sufficient tokens
2. Created gentx with adequate stake amount
3. Collected gentxs properly

#### Connection Issues

- Verify node is running: `ps aux | grep nuahd`
- Check logs in `~/.nuahd/logs/`
- Ensure firewall allows connections on ports 26657, 1317, 9090

### Reset Everything

If you need to start completely fresh:

```bash
# Stop node
pkill nuahd

# Remove all data
rm -rf ~/.nuahd

# Rebuild if needed
make build

# Follow setup steps from beginning
```

## 9. Development Workflow

### Using VS Code

1. Open project in VS Code
2. Use `Ctrl+Shift+P` → `Tasks: Run Task` → `build` to build
3. Use built-in terminal for running commands
4. Utilize Go extension for code editing and debugging

### Testing Changes

1. Stop the node
2. Make code changes
3. Build: `make build`
4. Restart node
5. Test with API calls

### Log Monitoring

```bash
# Follow logs
tail -f ~/.nuahd/logs/nuahd.log

# Or use journalctl if running as service
journalctl -u nuahd -f
```

## 10. Production Considerations

**⚠️ This setup is for development only**

For production deployment:
- Use secure key management (not `--keyring-backend test`)
- Configure proper networking and firewall
- Set up monitoring and alerting
- Use systemd service for automatic restart
- Configure backup procedures
- Use separate machines for different components

## Success Indicators

Your node is working correctly when:

✅ **Node Status**: `curl -s http://localhost:26657/status` returns valid JSON  
✅ **Block Production**: Block height increases over time  
✅ **Validator Active**: Validator appears in active set  
✅ **Token Balance**: Correct NUAH balance displayed  
✅ **API Access**: All endpoints respond correctly  

## Support

For issues or questions:
1. Check this documentation first
2. Verify all prerequisites are met
3. Check troubleshooting section
4. Review terminal output for error messages

---

**Chain ID**: `nuahchain-1`  
**Native Token**: `NUAH` (`unuah`)  
**RPC**: `http://localhost:26657`  
**REST**: `http://localhost:1317`  
**gRPC**: `localhost:9090`
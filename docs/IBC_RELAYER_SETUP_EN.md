# IBC Relayer Setup Guide (English)

## Table of Contents
1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Setting up Keys](#setting-up-keys)
6. [Creating IBC Connections](#creating-ibc-connections)
7. [Creating IBC Channels](#creating-ibc-channels)
8. [Testing IBC Transfers](#testing-ibc-transfers)
9. [Troubleshooting](#troubleshooting)
10. [Best Practices](#best-practices)

## Overview

This guide provides step-by-step instructions for setting up an IBC (Inter-Blockchain Communication) relayer using Hermes to connect your NUAH blockchain with external networks like Osmosis testnet.

IBC relayers are essential components that facilitate cross-chain communication by:
- Creating and maintaining IBC clients
- Establishing connections between chains
- Creating channels for specific protocols (like token transfers)
- Relaying packets between connected chains

## Prerequisites

Before starting, ensure you have:

1. **Running NUAH node** with IBC enabled
2. **Hermes relayer** installed (v1.7.4 or later)
3. **Access to target network** (e.g., Osmosis testnet)
4. **Sufficient funds** in both networks for transaction fees
5. **Basic understanding** of IBC concepts

### System Requirements
- Operating System: macOS, Linux, or Windows (WSL)
- RAM: Minimum 4GB, Recommended 8GB+
- Storage: At least 10GB free space
- Network: Stable internet connection

## Installation

### Installing Hermes

#### Option 1: Using Cargo (Recommended)
```bash
# Install Rust if not already installed
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source ~/.cargo/env

# Install Hermes
cargo install ibc-relayer-cli --bin hermes --locked
```

#### Option 2: Download Pre-built Binary
```bash
# Download latest release
wget https://github.com/informalsystems/hermes/releases/download/v1.7.4/hermes-v1.7.4-x86_64-unknown-linux-gnu.tar.gz

# Extract and install
tar -xzf hermes-v1.7.4-x86_64-unknown-linux-gnu.tar.gz
sudo mv hermes /usr/local/bin/
```

### Verify Installation
```bash
hermes --version
# Should output: hermes 1.7.4+ab73266
```

## Configuration

### 1. Create Configuration Directory
```bash
mkdir -p ~/.hermes_test
```

### 2. Create Configuration File
Create `~/.hermes_test/config.toml` with the following content:

```toml
[global]
log_level = 'info'

[mode]

[mode.clients]
enabled = true
refresh = true
misbehaviour = true

[mode.connections]
enabled = false

[mode.channels]
enabled = false

[mode.packets]
enabled = false

# NUAH Chain Configuration
[[chains]]
id = 'nuahchain-1'
type = 'CosmosSdk'
rpc_addr = 'http://localhost:26657'
grpc_addr = 'http://localhost:9090'
event_source = { mode = 'push', url = 'ws://localhost:26657/websocket', batch_delay = '500ms' }
rpc_timeout = '10s'
account_prefix = 'nuah'
key_name = 'test-relayer'
address_type = { derivation = 'cosmos' }
store_prefix = 'ibc'
default_gas = 100000
max_gas = 400000
gas_price = { price = 0.001, denom = 'unuah' }
gas_multiplier = 1.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
max_block_time = '30s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }

# Osmosis Testnet Configuration
[[chains]]
id = 'osmo-test-5'
type = 'CosmosSdk'
rpc_addr = 'https://rpc.testnet.osmosis.zone:443'
grpc_addr = 'https://grpc.testnet.osmosis.zone:443'
event_source = { mode = 'push', url = 'wss://rpc.testnet.osmosis.zone:443/websocket', batch_delay = '500ms' }
rpc_timeout = '10s'
account_prefix = 'osmo'
key_name = 'osmosis-relayer'
address_type = { derivation = 'cosmos' }
store_prefix = 'ibc'
default_gas = 100000
max_gas = 400000
gas_price = { price = 0.0025, denom = 'uosmo' }
gas_multiplier = 1.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
max_block_time = '30s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
```

### 3. Validate Configuration
```bash
hermes --config ~/.hermes_test/config.toml health-check
```

## Setting up Keys

### 1. Create Relayer Keys

For NUAH chain:
```bash
# Generate or import mnemonic for NUAH relayer
hermes --config ~/.hermes_test/config.toml keys add \
  --chain nuahchain-1 \
  --mnemonic-file <(echo "your mnemonic phrase here")
```

For Osmosis testnet:
```bash
# Generate or import mnemonic for Osmosis relayer
hermes --config ~/.hermes_test/config.toml keys add \
  --chain osmo-test-5 \
  --mnemonic-file <(echo "your mnemonic phrase here")
```

### 2. Verify Keys
```bash
# List keys for NUAH chain
hermes --config ~/.hermes_test/config.toml keys list --chain nuahchain-1

# List keys for Osmosis testnet
hermes --config ~/.hermes_test/config.toml keys list --chain osmo-test-5
```

### 3. Fund Relayer Accounts

**NUAH Chain:**
```bash
# Check balance
./build/nuahd query bank balances [relayer-address]

# Send funds if needed
./build/nuahd tx bank send [sender] [relayer-address] 1000000unuah --fees 1000unuah
```

**Osmosis Testnet:**
- Use the Osmosis testnet faucet: https://faucet.testnet.osmosis.zone/
- Request tokens for your relayer address

## Creating IBC Connections

### 1. Create IBC Clients

Create client for NUAH chain on Osmosis:
```bash
hermes --config ~/.hermes_test/config.toml create client \
  --host-chain osmo-test-5 \
  --reference-chain nuahchain-1
```

Create client for Osmosis on NUAH chain:
```bash
hermes --config ~/.hermes_test/config.toml create client \
  --host-chain nuahchain-1 \
  --reference-chain osmo-test-5
```

### 2. Create Connection
```bash
hermes --config ~/.hermes_test/config.toml create connection \
  --a-chain nuahchain-1 \
  --b-chain osmo-test-5
```

### 3. Verify Connection
```bash
# Check connections on NUAH chain
./build/nuahd query ibc connection connections

# Check connections on Osmosis (if you have access)
osmosisd query ibc connection connections
```

## Creating IBC Channels

### 1. Create Transfer Channel
```bash
hermes --config ~/.hermes_test/config.toml create channel \
  --a-chain nuahchain-1 \
  --a-connection connection-0 \
  --a-port transfer \
  --b-port transfer \
  --channel-version ics20-1 \
  --order unordered
```

### 2. Verify Channel
```bash
# Check channels on NUAH chain
./build/nuahd query ibc channel channels

# List all channels via Hermes
hermes --config ~/.hermes_test/config.toml query channels --chain nuahchain-1
```

## Testing IBC Transfers

### 1. Transfer from NUAH to Osmosis
```bash
./build/nuahd tx ibc-transfer transfer \
  transfer \
  channel-0 \
  [osmosis-receiver-address] \
  1000unuah \
  --from [nuah-sender] \
  --fees 1000unuah \
  --timeout-height 0-0 \
  --timeout-timestamp 0
```

### 2. Transfer from Osmosis to NUAH
```bash
osmosisd tx ibc-transfer transfer \
  transfer \
  channel-X \
  [nuah-receiver-address] \
  1000uosmo \
  --from [osmosis-sender] \
  --fees 1000uosmo \
  --chain-id osmo-test-5 \
  --node https://rpc.testnet.osmosis.zone:443
```

### 3. Start Relayer (for packet relaying)
```bash
hermes --config ~/.hermes_test/config.toml start
```

## Troubleshooting

### Common Issues

#### 1. "Insufficient funds" Error
**Problem:** Relayer account doesn't have enough tokens for transaction fees.
**Solution:** 
- Fund the relayer account using faucet or direct transfer
- Check minimum required balance for the network

#### 2. "Client state type not supported" Error
**Problem:** Incompatible client types between chains.
**Solution:**
- Ensure both chains support the same IBC version
- Check Hermes compatibility with chain versions

#### 3. "Connection not found" Error
**Problem:** Trying to create channel with non-existent connection.
**Solution:**
- Verify connection exists: `hermes query connections --chain [chain-id]`
- Create connection first if it doesn't exist

#### 4. Health Check Warnings
**Problem:** SDK version compatibility warnings.
**Solution:**
- These are usually non-critical warnings
- Ensure you're using compatible Hermes version
- Update chain software if needed

### Debugging Commands

```bash
# Check Hermes logs
hermes --config ~/.hermes_test/config.toml health-check

# Query specific client
hermes --config ~/.hermes_test/config.toml query client state --chain [chain-id] --client [client-id]

# Query connection details
hermes --config ~/.hermes_test/config.toml query connection end --chain [chain-id] --connection [connection-id]

# Query channel details
hermes --config ~/.hermes_test/config.toml query channel end --chain [chain-id] --port [port-id] --channel [channel-id]
```

## Best Practices

### Security
1. **Use dedicated relayer accounts** - Don't use validator or main accounts
2. **Secure key storage** - Store mnemonics securely, consider hardware wallets for production
3. **Monitor balances** - Set up alerts for low balances
4. **Regular updates** - Keep Hermes and chain software updated

### Performance
1. **Optimize gas settings** - Adjust gas prices based on network conditions
2. **Monitor packet delays** - Set appropriate timeout values
3. **Use multiple relayers** - For high-traffic channels, consider multiple relayer instances
4. **Resource allocation** - Ensure adequate CPU and memory for relayer operations

### Monitoring
1. **Set up logging** - Configure appropriate log levels for monitoring
2. **Health checks** - Regular health check monitoring
3. **Metrics collection** - Use Prometheus/Grafana for advanced monitoring
4. **Alert systems** - Set up alerts for relayer failures or stuck packets

### Maintenance
1. **Regular client updates** - Update IBC clients before they expire
2. **Connection monitoring** - Monitor connection health and status
3. **Channel maintenance** - Monitor channel state and packet flow
4. **Backup procedures** - Regular backup of relayer configuration and keys

---

For more detailed information, refer to:
- [Hermes Documentation](https://hermes.informal.systems/)
- [IBC Protocol Specification](https://github.com/cosmos/ibc)
- [Cosmos SDK IBC Module](https://ibc.cosmos.network/)
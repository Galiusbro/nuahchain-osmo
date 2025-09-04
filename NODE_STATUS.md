# Nuah Chain Node Status

## Current Configuration

- **Node Status**: ✅ RUNNING
- **Chain ID**: nuahchain-1  
- **Node Moniker**: test-node
- **Block Height**: 67+ (and increasing)
- **Validator Address**: `nuah14k38ajalnef2yauznt4q7ep893djkl4vm54mcs`
- **Validator Status**: ACTIVE (BOND_STATUS_BONDED)

## Token Details

- **Native Token**: NUAH
- **Base Denomination**: unuah
- **Validator Balance**: 5,000,000 NUAH (5,000,000,000,000 unuah)
- **Staked Amount**: 5,000,000 NUAH  
- **Total Supply**: 10,000,000 NUAH

## Network Endpoints

- **RPC**: http://localhost:26657
- **REST API**: http://localhost:1317
- **gRPC**: localhost:9090

## Quick Status Checks

```bash
# Node health
curl -s http://localhost:26657/health

# Current block
curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height'

# Validator balance  
./build/nuahd query bank balances nuah14k38ajalnef2yauznt4q7ep893djkl4vm54mcs
```

## Process Information

```bash
# Find node process
ps aux | grep nuahd

# Node logs
tail -f ~/.nuahd/logs/nuahd.log
```

---
*Last updated: Node successfully initialized and running*
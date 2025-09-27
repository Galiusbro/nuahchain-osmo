# N$ (Ndollar) Token Setup Guide 🏦

This guide provides comprehensive instructions for setting up and managing the N$ (Ndollar) algorithmic stablecoin on Nuah Chain.

## 📋 Overview

N$ (Ndollar) is an algorithmic stablecoin without collateral backing, designed to maintain a 1 N$ ≈ 1 USD parity. It serves as:

- **Base trading pair** for all user tokens created through tokenfactory
- **Payment medium** through fee abstraction
- **Ecosystem glue** connecting various assets in Nuah Chain

## 🚀 Quick Start

### Prerequisites

1. **Running Nuah Chain node** - Ensure your node is operational
2. **Validator key** - Created through `setup_proper_tokenomics.sh`
3. **NUAH tokens** - For transaction fees during setup

### Step 1: Setup N$ Token

Run the main setup script:

```bash
cd /path/to/nuahchain_osmosis
./scripts/setup_ndollar.sh
```

This script will:
- ✅ Create N$ token using TokenFactory
- ✅ Set comprehensive token metadata
- ✅ Mint initial supply (1M N$)
- ✅ Create NUAH/N$ liquidity pool (1:1 ratio)
- ✅ Configure fee abstraction
- ✅ Verify complete setup

### Step 2: Verify Installation

Check N$ status:

```bash
./scripts/ndollar_helper.sh status
```

## 🔧 Management Tools

### N$ Helper Script

The `ndollar_helper.sh` script provides utilities for managing N$ operations:

```bash
# Check balance
./scripts/ndollar_helper.sh balance <address>

# Transfer tokens
./scripts/ndollar_helper.sh transfer validator <recipient> <amount>

# Mint tokens (admin only)
./scripts/ndollar_helper.sh mint validator <amount> <recipient>

# Check token metadata
./scripts/ndollar_helper.sh metadata

# View pool information
./scripts/ndollar_helper.sh pool

# Swap NUAH to N$
./scripts/ndollar_helper.sh swap-to-ndollar validator <nuah_amount> <min_ndollar> <pool_id>

# Full status report
./scripts/ndollar_helper.sh status
```

### Manual Commands

#### Token Operations

```bash
# Check N$ balance
./build/nuahd query bank balance <address> factory/<creator>/ndollar

# Transfer N$ tokens
./build/nuahd tx bank send validator <recipient> 1000000factory/<creator>/ndollar \
  --keyring-backend test --chain-id nuahchain-1 --fees 2000unuah -y

# Mint N$ (admin only)
./build/nuahd tx tokenfactory mint 1000000factory/<creator>/ndollar <recipient> \
  --from validator --keyring-backend test --chain-id nuahchain-1 --fees 2000unuah -y
```

#### Fee Abstraction

```bash
# Use N$ for transaction fees
./build/nuahd tx bank send validator <recipient> 1000000unuah \
  --fees 1000factory/<creator>/ndollar --keyring-backend test --chain-id nuahchain-1 -y

# Check fee abstraction parameters
./build/nuahd query txfees params
```

#### Pool Operations

```bash
# Query pools containing N$
./build/nuahd query poolmanager pools

# Swap in pool
./build/nuahd tx poolmanager swap-exact-amount-in \
  1000000unuah 900000 \
  --swap-route-pool-ids 1 \
  --swap-route-denoms factory/<creator>/ndollar \
  --from validator --keyring-backend test --chain-id nuahchain-1 -y
```

## 📁 File Structure

```
nuahchain_osmosis/
├── scripts/
│   ├── setup_ndollar.sh          # Main setup script
│   ├── ndollar_helper.sh          # Management utilities
│   └── setup_fee_abstraction.sh   # Fee abstraction setup
├── configs/
│   └── ndollar_pool_config.json   # Pool configuration
└── README_NDOLLAR.md              # This documentation
```

## 🔍 Technical Details

### Token Specification

- **Name**: Ndollar
- **Symbol**: N$
- **Denomination**: `factory/{creator_address}/ndollar`
- **Decimals**: 6
- **Type**: Algorithmic stablecoin
- **Target**: 1 N$ ≈ 1 USD

### Pool Configuration

- **Type**: Balancer pool
- **Assets**: NUAH/N$ (1:1 ratio)
- **Swap Fee**: 0.3%
- **Exit Fee**: 0%
- **Initial Liquidity**: 1M NUAH + 1M N$

### Features

1. **TokenFactory Integration**: Uses Osmosis-based tokenfactory module
2. **Fee Abstraction**: Pay transaction fees with N$
3. **Base Trading Pair**: All user tokens trade against N$
4. **Algorithmic Supply**: Supply managed algorithmically
5. **Price Stability**: Targeting USD parity

## 🛠️ Troubleshooting

### Common Issues

#### 1. "nuahd binary not found"
```bash
# Build the binary
make build
```

#### 2. "Node is not running"
```bash
# Start the node or run tokenomics setup
./scripts/setup_proper_tokenomics.sh
```

#### 3. "Validator key not found"
```bash
# Create validator key first
./build/nuahd keys add validator --keyring-backend test
```

#### 4. "Insufficient fees"
```bash
# Ensure you have NUAH tokens for fees
./build/nuahd query bank balance <address> unuah
```

### Validation Commands

```bash
# Check if N$ token exists
./build/nuahd query tokenfactory denoms-from-creator <creator_address>

# Verify metadata
./build/nuahd query bank denom-metadata factory/<creator>/ndollar

# Check pool exists
./build/nuahd query poolmanager pools | grep ndollar

# Verify fee abstraction
./build/nuahd query txfees params | grep ndollar
```

## 📊 Monitoring

### Key Metrics to Monitor

1. **Total Supply**: Track N$ token supply
2. **Pool Liquidity**: Monitor NUAH/N$ pool depth
3. **Price Stability**: Watch N$/USD parity
4. **Fee Usage**: Track fee abstraction usage
5. **Trading Volume**: Monitor swap activity

### Monitoring Commands

```bash
# Total N$ supply
./build/nuahd query bank total --denom factory/<creator>/ndollar

# Pool information
./scripts/ndollar_helper.sh pool

# Recent transactions
./build/nuahd query tx <tx_hash>

# Account balances
./scripts/ndollar_helper.sh balance <address>
```

## 🔐 Security Considerations

### Admin Privileges

The N$ token creator has significant control:
- **Minting**: Can create new N$ tokens
- **Burning**: Can destroy N$ tokens
- **Metadata**: Can update token information
- **Admin Transfer**: Can change token admin

### Best Practices

1. **Secure Key Management**: Protect validator/admin keys
2. **Regular Monitoring**: Watch for unusual activity
3. **Supply Management**: Monitor total supply changes
4. **Pool Health**: Ensure adequate liquidity
5. **Fee Validation**: Verify fee abstraction works correctly

## 📚 Additional Resources

- [Technical Specification v.0.1.1](./!DOC/technical%20specification%20v.0.1.1.md)
- [TokenFactory Developer Guide](./!DOC/TOKEN_FACTORY_DEVELOPER_GUIDE.md)
- [TokenFactory User Guide](./!DOC/TOKEN_FACTORY_USER_GUIDE.md)
- [Nuah Chain Documentation](./!DOC/)

## 🆘 Support

For issues or questions:

1. Check this documentation first
2. Review the technical specification
3. Examine script logs for error details
4. Test with small amounts first
5. Reach out to the development team

---

**⚠️ Important**: N$ is an algorithmic stablecoin without collateral backing. Use appropriate risk management and monitor the system closely.

🏦 **Happy Trading with N$!** 🏦
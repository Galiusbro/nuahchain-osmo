# TokenFactory User Guide 🏭

Welcome to the Nuah Chain TokenFactory! This powerful feature allows you to create, manage, and trade custom tokens directly on the blockchain.

## 🌟 What is TokenFactory?

TokenFactory is a module that enables **anyone** to create their own tokens on Nuah Chain. Unlike other blockchains where token creation requires complex smart contracts, TokenFactory makes it simple and secure.

### Key Features:
- **🚀 Easy Token Creation** - Create tokens with a single command
- **📋 Rich Metadata** - Set name, symbol, description, and display units
- **🛠️ Full Control** - Mint, burn, and transfer tokens as the admin
- **💰 Low Fees** - Cost-effective token operations
- **🔒 Secure** - Built on battle-tested Osmosis codebase

## 🎯 Quick Start

### Prerequisites
- Nuah Chain node running
- Account with NUAH tokens for fees
- Basic command-line knowledge

### Step 1: Create Your Token

```bash
nuahd tx tokenfactory create-denom YOUR_TOKEN_NAME \
  --from YOUR_WALLET \
  --chain-id nuahchain-1 \
  --keyring-backend test \
  --gas 1500000 \
  --fees 16000unuah \
  -y
```

**Example:**
```bash
nuahd tx tokenfactory create-denom mytoken \
  --from validator \
  --chain-id nuahchain-1 \
  --keyring-backend test \
  --gas 1500000 \
  --fees 16000unuah \
  -y
```

This creates a token with the full denomination:
`factory/YOUR_ADDRESS/YOUR_TOKEN_NAME`

### Step 2: Set Token Metadata

```bash
nuahd tx tokenfactory set-denom-metadata '{
  "description": "My awesome token on Nuah Chain",
  "denom_units": [
    {
      "denom": "factory/YOUR_ADDRESS/YOUR_TOKEN_NAME",
      "exponent": 0,
      "aliases": ["base-unit"]
    },
    {
      "denom": "YOUR_TOKEN_NAME", 
      "exponent": 6,
      "aliases": ["TOKEN"]
    }
  ],
  "base": "factory/YOUR_ADDRESS/YOUR_TOKEN_NAME",
  "display": "YOUR_TOKEN_NAME",
  "name": "My Token",
  "symbol": "MTK"
}' \
  --from YOUR_WALLET \
  --chain-id nuahchain-1 \
  --keyring-backend test \
  --gas 300000 \
  --fees 3000unuah \
  -y
```

### Step 3: Mint Your Tokens

```bash
nuahd tx tokenfactory mint AMOUNT+DENOM RECIPIENT_ADDRESS \
  --from YOUR_WALLET \
  --chain-id nuahchain-1 \
  --keyring-backend test \
  --gas 200000 \
  --fees 2000unuah \
  -y
```

**Example:**
```bash
nuahd tx tokenfactory mint 1000000factory/nuah14k38ajalnef2yauznt4q7ep893djkl4vm54mcs/mytoken nuah14k38ajalnef2yauznt4q7ep893djkl4vm54mcs \
  --from validator \
  --chain-id nuahchain-1 \
  --keyring-backend test \
  --gas 200000 \
  --fees 2000unuah \
  -y
```

## 📖 Complete Example Walkthrough

Let's create a token called "SuperCoin" (SUP):

### 1. Create the Token
```bash
nuahd tx tokenfactory create-denom supercoin \
  --from myaccount \
  --chain-id nuahchain-1 \
  --keyring-backend test \
  --gas 1500000 \
  --fees 16000unuah \
  -y
```

### 2. Set Rich Metadata
```bash
nuahd tx tokenfactory set-denom-metadata '{
  "description": "SuperCoin - The ultimate digital currency for the future",
  "denom_units": [
    {
      "denom": "factory/nuah1abc123.../supercoin",
      "exponent": 0,
      "aliases": ["usup", "microsup"]
    },
    {
      "denom": "msup",
      "exponent": 3,
      "aliases": ["millisup"]
    },
    {
      "denom": "sup",
      "exponent": 6,
      "aliases": ["SUP", "SuperCoin"]
    }
  ],
  "base": "factory/nuah1abc123.../supercoin",
  "display": "sup",
  "name": "SuperCoin",
  "symbol": "SUP",
  "uri": "https://supercoin.com",
  "uri_hash": ""
}' \
  --from myaccount \
  --chain-id nuahchain-1 \
  --keyring-backend test \
  --gas 300000 \
  --fees 3000unuah \
  -y
```

### 3. Mint Initial Supply
```bash
# Mint 1 million SUP (1,000,000 * 10^6 base units)
nuahd tx tokenfactory mint 1000000000000factory/nuah1abc123.../supercoin nuah1abc123... \
  --from myaccount \
  --chain-id nuahchain-1 \
  --keyring-backend test \
  --gas 200000 \
  --fees 2000unuah \
  -y
```

## 🔍 Query Commands

### Check Your Created Tokens
```bash
nuahd query tokenfactory denoms-from-creator YOUR_ADDRESS
```

### View Token Metadata
```bash
nuahd query bank denom-metadata FULL_DENOM_NAME
```

### Check Token Balances
```bash
nuahd query bank balances YOUR_ADDRESS
```

### Check Token Authority
```bash
nuahd query tokenfactory denom-authority-metadata FULL_DENOM_NAME
```

## 🛠️ Advanced Operations

### Burn Tokens
```bash
nuahd tx tokenfactory burn AMOUNT+DENOM \
  --from YOUR_WALLET \
  --chain-id nuahchain-1 \
  --keyring-backend test \
  --gas 200000 \
  --fees 2000unuah \
  -y
```

### Change Token Admin
```bash
nuahd tx tokenfactory change-admin FULL_DENOM_NAME NEW_ADMIN_ADDRESS \
  --from CURRENT_ADMIN \
  --chain-id nuahchain-1 \
  --keyring-backend test \
  --gas 200000 \
  --fees 2000unuah \
  -y
```

### Force Transfer (Admin only)
```bash
nuahd tx tokenfactory force-transfer AMOUNT+DENOM FROM_ADDRESS TO_ADDRESS \
  --from ADMIN \
  --chain-id nuahchain-1 \
  --keyring-backend test \
  --gas 200000 \
  --fees 2000unuah \
  -y
```

## 💡 Best Practices

### 🎨 Token Design
- **Choose meaningful names** - Make your token easily identifiable
- **Set appropriate decimals** - Usually 6 or 18 for compatibility
- **Add comprehensive metadata** - Description, website, social links

### 💰 Cost Management
- **Token creation fee**: ~16,000 unuah (~$0.01)
- **Metadata setting**: ~3,000 unuah
- **Mint operations**: ~2,000 unuah per transaction
- **Use appropriate gas limits** to avoid failures

### 🔐 Security
- **Secure your admin keys** - Admin can mint/burn tokens
- **Consider multi-sig** for valuable tokens
- **Test on testnet first** before mainnet deployment

### 📊 Economics
- **Plan your tokenomics** - Total supply, distribution, utility
- **Consider deflationary mechanisms** - Burn tokens to reduce supply
- **Implement governance** for community-driven tokens

## 🌐 Integration Examples

### Web3 Integration
```javascript
// Query token metadata
const metadata = await client.query.bank.denomMetadata({
  denom: "factory/nuah1abc.../mytoken"
});

// Check balance
const balance = await client.query.bank.balance({
  address: "nuah1abc...",
  denom: "factory/nuah1abc.../mytoken"
});
```

### Trading Integration
Your tokens are automatically tradeable on:
- **Osmosis DEX** (if bridged)
- **Local AMM pools**
- **Custom trading interfaces**

## 🔗 Useful Resources

- **Nuah Chain Explorer**: Browse tokens and transactions
- **Node Status**: Check network health and block height  
- **Community Discord**: Get help and share your tokens
- **Developer Docs**: Technical implementation details

## ❓ FAQ

**Q: How much does it cost to create a token?**
A: Approximately 16,000 unuah (~$0.01 USD) plus network fees.

**Q: Can I create tokens with different decimal places?**
A: Yes! Set the `exponent` in your metadata. Common values are 6, 8, or 18.

**Q: Can I change token metadata after creation?**
A: Yes, as long as you're the admin. Use `set-denom-metadata` again.

**Q: How do I make my token tradeable?**
A: Once created, your token can be used in any DEX or trading protocol that supports Nuah Chain.

**Q: What's the difference between base and display denominations?**
A: Base is the smallest unit (like satoshis), display is the human-readable unit (like BTC).

**Q: Can I create NFTs with TokenFactory?**
A: TokenFactory is for fungible tokens. For NFTs, use the CosmWasm NFT modules.

## 🆘 Troubleshooting

### Common Errors

**"out of gas"**
- Increase `--gas` parameter (try 1500000 for token creation)

**"insufficient fees"**
- Increase `--fees` parameter (try 16000unuah for token creation)

**"denom already exists"**
- Choose a different subdenom name

**"unauthorized"**
- Make sure you're using the correct admin account

### Getting Help
- Check the error message carefully
- Verify your account has sufficient NUAH balance
- Ensure the node is running and synced
- Ask in our community channels

---

🎉 **Congratulations!** You now have everything you need to create and manage tokens on Nuah Chain. Start building the future of decentralized finance!
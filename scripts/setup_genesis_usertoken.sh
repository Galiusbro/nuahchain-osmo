#!/bin/bash

# Script to setup genesis file with proper usertoken module parameters
# This script should be run BEFORE starting the blockchain for the first time

set -e

CHAIN_ID="nuahchain-1"
MONIKER="nuah-validator"
KEYRING_BACKEND="test"
HOME_DIR="$HOME/.nuahd"

echo "🚀 Setting up genesis file with usertoken module parameters..."

# Function to create a wallet if it doesn't exist
create_wallet_if_not_exists() {
    local wallet_name=$1
    if ! ./build/nuahd keys show $wallet_name --keyring-backend $KEYRING_BACKEND --home $HOME_DIR >/dev/null 2>&1; then
        echo "Creating wallet: $wallet_name"
        ./build/nuahd keys add $wallet_name --keyring-backend $KEYRING_BACKEND --home $HOME_DIR
    else
        echo "Wallet $wallet_name already exists"
    fi
}

# Create special wallets for usertoken module
echo "📝 Creating special wallets..."
create_wallet_if_not_exists "ai-ceo-wallet"
create_wallet_if_not_exists "referral-wallet"  
create_wallet_if_not_exists "platform-fee-wallet"

# Get wallet addresses
AI_CEO_WALLET=$(./build/nuahd keys show ai-ceo-wallet -a --keyring-backend $KEYRING_BACKEND --home $HOME_DIR)
REFERRAL_WALLET=$(./build/nuahd keys show referral-wallet -a --keyring-backend $KEYRING_BACKEND --home $HOME_DIR)
PLATFORM_FEE_WALLET=$(./build/nuahd keys show platform-fee-wallet -a --keyring-backend $KEYRING_BACKEND --home $HOME_DIR)

echo "✅ Wallet addresses:"
echo "AI CEO Wallet: $AI_CEO_WALLET"
echo "Referral Wallet: $REFERRAL_WALLET"
echo "Platform Fee Wallet: $PLATFORM_FEE_WALLET"

# Create custom genesis file with usertoken parameters
echo "🔧 Creating custom genesis file..."

# Initialize the chain if not already done
if [ ! -f "$HOME_DIR/config/genesis.json" ]; then
    echo "Initializing chain..."
    ./build/nuahd init $MONIKER --chain-id $CHAIN_ID --home $HOME_DIR
fi

# Create a temporary JSON file with usertoken parameters
cat > /tmp/usertoken_params.json << EOF
{
  "params": {
    "founder_tranche_price": "0.00005",
    "founder_tranche_amount": "10000000",
    "bonding_curve_start_price": "0.0002",
    "bonding_curve_end_price": "1.0",
    "bonding_curve_max_supply": "30000000",
    "min_creator_purchase": "500",
    "ai_ceo_wallet": "$AI_CEO_WALLET",
    "referral_wallet": "$REFERRAL_WALLET",
    "platform_fee_wallet": "$PLATFORM_FEE_WALLET"
  },
  "user_tokens": [],
  "referral_programs": [],
  "referral_activations": [],
  "user_referral_quotas": []
}
EOF

# Update genesis file with usertoken module parameters
echo "📄 Updating genesis file with usertoken parameters..."
jq --argjson usertoken "$(cat /tmp/usertoken_params.json)" '.app_state.usertoken = $usertoken' $HOME_DIR/config/genesis.json > /tmp/genesis_updated.json

# Replace the original genesis file
mv /tmp/genesis_updated.json $HOME_DIR/config/genesis.json

# Clean up temporary file
rm /tmp/usertoken_params.json

echo "✅ Genesis file updated successfully!"
echo "🔍 Usertoken module parameters in genesis:"
jq '.app_state.usertoken.params' $HOME_DIR/config/genesis.json

echo ""
echo "🎉 Setup complete! Your blockchain is now configured with:"
echo "   - AI CEO Wallet: $AI_CEO_WALLET"
echo "   - Referral Wallet: $REFERRAL_WALLET" 
echo "   - Platform Fee Wallet: $PLATFORM_FEE_WALLET"
echo "   - Founder tranche price: 0.00005 N$"
echo "   - Founder tranche amount: 10M tokens"
echo "   - Bonding curve start price: 0.0002 N$"
echo "   - Bonding curve end price: 1.0 N$"
echo "   - Bonding curve max supply: 30M tokens"
echo "   - Min creator purchase: 500 N$"
echo ""
echo "🚀 You can now start your blockchain with: ./build/nuahd start"
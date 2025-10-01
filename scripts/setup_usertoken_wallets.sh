#!/bin/bash

# Script to setup special wallets for usertoken module
# This script creates AI CEO, Platform, and Referral wallets and updates module parameters

set -e

CHAIN_ID="nuahchain"
KEYRING_BACKEND="test"
NODE="tcp://localhost:26657"
NUAHD_BINARY="./build/nuahd"

echo "🚀 Setting up usertoken module special wallets..."

# Check if nuahd binary exists
if [ ! -f "$NUAHD_BINARY" ]; then
    echo "❌ Error: nuahd binary not found at $NUAHD_BINARY"
    echo "Please build the binary first: make build"
    exit 1
fi

# Function to create wallet if it doesn't exist
create_wallet_if_not_exists() {
    local wallet_name=$1
    local wallet_description=$2

    echo "📝 Checking wallet: $wallet_name ($wallet_description)"

    # Check if wallet already exists
    if $NUAHD_BINARY keys show $wallet_name --keyring-backend $KEYRING_BACKEND >/dev/null 2>&1; then
        echo "✅ Wallet $wallet_name already exists"
        local address=$($NUAHD_BINARY keys show -a $wallet_name --keyring-backend $KEYRING_BACKEND)
        echo "   Address: $address"
        echo "$address"
    else
        echo "🔑 Creating new wallet: $wallet_name"
        # Create wallet and capture output
        local wallet_output=$($NUAHD_BINARY keys add $wallet_name --keyring-backend $KEYRING_BACKEND --output json 2>/dev/null)
        local address=$(echo "$wallet_output" | jq -r '.address')
        echo "✅ Created wallet $wallet_name"
        echo "   Address: $address"
        echo "$address"
    fi
}

# Create special wallets
echo ""
echo "🏦 Creating special wallets..."
AI_CEO_WALLET=$(create_wallet_if_not_exists "ai-ceo-wallet" "AI CEO Wallet")
PLATFORM_WALLET=$(create_wallet_if_not_exists "platform-wallet" "Platform Fee Wallet")
REFERRAL_WALLET=$(create_wallet_if_not_exists "referral-wallet" "Referral Wallet")

echo ""
echo "📋 Wallet Summary:"
echo "   AI CEO Wallet:    $AI_CEO_WALLET"
echo "   Platform Wallet:  $PLATFORM_WALLET"
echo "   Referral Wallet:  $REFERRAL_WALLET"

# Get the authority address (usually the governance module)
echo ""
echo "🔍 Getting authority address for parameter update..."
AUTHORITY_ADDRESS=$($NUAHD_BINARY keys show -a validator --keyring-backend $KEYRING_BACKEND)
echo "   Authority: $AUTHORITY_ADDRESS"

# Create update params message using direct command instead of JSON
echo ""
echo "🔐 Updating module parameters..."
$NUAHD_BINARY tx usertoken update-params \
    --ai-ceo-wallet "$AI_CEO_WALLET" \
    --platform-fee-wallet "$PLATFORM_WALLET" \
    --referral-wallet "$REFERRAL_WALLET" \
    --founder-tranche-price "1000000" \
    --founder-tranche-amount "1000000" \
    --bonding-curve-start-price "1000000" \
    --bonding-curve-end-price "10000000" \
    --bonding-curve-max-supply "30000000" \
    --min-creator-purchase "1000000" \
    --from validator \
    --keyring-backend $KEYRING_BACKEND \
    --chain-id $CHAIN_ID \
    --node $NODE \
    --gas 200000 \
    --fees 5000unuah \
    --yes

# Verify updated parameters
echo ""
echo "🔍 Verifying updated parameters..."
sleep 2
UPDATED_PARAMS=$($NUAHD_BINARY query usertoken params --node $NODE --output json)
echo "✅ Updated parameters:"
echo "$UPDATED_PARAMS" | jq '.'

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "📋 Summary:"
echo "   ✅ AI CEO Wallet created:    $AI_CEO_WALLET"
echo "   ✅ Platform Wallet created:  $PLATFORM_WALLET"
echo "   ✅ Referral Wallet created:  $REFERRAL_WALLET"
echo "   ✅ Module parameters updated"
echo ""
echo "🚀 You can now create user tokens and they will be distributed correctly to the special wallets!"

# Cleanup temporary files
echo "✨ Script completed!"
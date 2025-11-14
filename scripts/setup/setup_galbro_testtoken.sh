#!/bin/bash

# Setup GALBRO test token for Exchange module testing
# GALBRO can be exchanged for unuah at fixed 1:1 USD rate

set -e

CHAIN_ID="nuahchain"
KEYRING_BACKEND="test"
VALIDATOR_KEY="validator"
GALBRO_AMOUNT="1000000000000" # 1,000,000 GALBRO (with 6 decimals)
TEST_USER_AMOUNT="50000000000"  # 50,000 GALBRO per user

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║          🪙 SETUP GALBRO TEST TOKEN                              ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# Get validator address
VALIDATOR_ADDR=$(./build/nuahd keys show $VALIDATOR_KEY -a --keyring-backend $KEYRING_BACKEND)
echo "Validator address: $VALIDATOR_ADDR"

# Define GALBRO denom
GALBRO_DENOM="factory/${VALIDATOR_ADDR}/galbro"
echo "GALBRO denom: $GALBRO_DENOM"
echo ""

# Step 1: Create denom (if not exists)
echo "🔄 Step 1: Creating GALBRO denom..."
./build/nuahd tx tokenfactory create-denom galbro \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING_BACKEND \
  --gas 2000000 \
  --fees 20000unuah \
  -y > /dev/null 2>&1 || echo "  ℹ️  Denom might already exist"

sleep 3
echo "✅ Denom created"
echo ""

# Step 2: Mint tokens
echo "🔄 Step 2: Minting 1,000,000 GALBRO..."
./build/nuahd tx tokenfactory mint "${GALBRO_AMOUNT}${GALBRO_DENOM}" "$VALIDATOR_ADDR" \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING_BACKEND \
  --gas 300000 \
  --fees 3000unuah \
  -y > /dev/null 2>&1

sleep 3
echo "✅ Tokens minted"
echo ""

# Step 3: Distribute to test users
echo "🔄 Step 3: Distributing to test users..."

# Alice
ALICE_ADDR=$(./build/nuahd keys show alice -a --keyring-backend $KEYRING_BACKEND 2>/dev/null || echo "")
if [ -n "$ALICE_ADDR" ]; then
  echo "  Sending 50,000 GALBRO to Alice ($ALICE_ADDR)..."
  ./build/nuahd tx bank send $VALIDATOR_KEY "$ALICE_ADDR" "${TEST_USER_AMOUNT}${GALBRO_DENOM}" \
    --chain-id $CHAIN_ID \
    --keyring-backend $KEYRING_BACKEND \
    --gas 200000 \
    --fees 2000unuah \
    -y > /dev/null 2>&1
  sleep 2
fi

# Garold (test user from server)
GAROLD_ADDR="nuah10us33fwsvajr57pgjxw638xzqjsfntqxk6yw56"
echo "  Sending 50,000 GALBRO to Garold ($GAROLD_ADDR)..."
./build/nuahd tx bank send $VALIDATOR_KEY "$GAROLD_ADDR" "${TEST_USER_AMOUNT}${GALBRO_DENOM}" \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING_BACKEND \
  --gas 200000 \
  --fees 2000unuah \
  -y > /dev/null 2>&1

sleep 3
echo "✅ Distribution complete"
echo ""

# Step 4: Check balances
echo "📊 Token balances:"
echo "  Validator: $(./build/nuahd query bank balance $VALIDATOR_ADDR $GALBRO_DENOM --output json 2>/dev/null | jq -r '.balance.amount') GALBRO (micro-units)"
if [ -n "$ALICE_ADDR" ]; then
  echo "  Alice:     $(./build/nuahd query bank balance $ALICE_ADDR $GALBRO_DENOM --output json 2>/dev/null | jq -r '.balance.amount') GALBRO (micro-units)"
fi
echo "  Garold:    $(./build/nuahd query bank balance $GAROLD_ADDR $GALBRO_DENOM --output json 2>/dev/null | jq -r '.balance.amount') GALBRO (micro-units)"
echo ""

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║                  ✅ GALBRO SETUP COMPLETE!                       ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""
echo "📝 Token Details:"
echo "  Name:     GALBRO Test Token"
echo "  Symbol:   GALBRO"
echo "  Denom:    $GALBRO_DENOM"
echo "  Decimals: 6"
echo "  Supply:   1,000,000 GALBRO"
echo ""
echo "⚠️  TO USE WITH EXCHANGE MODULE:"
echo "  1. Register in x/usdoracle (add to supported_tokens)"
echo "  2. Set fixed price: 1.00 USD"
echo "  3. Add to x/exchange supported_tokens list"
echo ""
echo "🧪 Test users can now exchange GALBRO for unuah!"
echo ""


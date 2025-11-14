#!/bin/bash

# Set GALBRO price to 1.00 USD in x/usdoracle
# This is needed for Exchange module to work

set -e

CHAIN_ID="${CHAIN_ID:-nuahchain}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
VALIDATOR_KEY="validator"

echo "Setting GALBRO price to 1.00 USD..."

# Get validator address
VALIDATOR_ADDR=$(./build/nuahd keys show $VALIDATOR_KEY -a --keyring-backend $KEYRING_BACKEND)
GALBRO_DENOM="factory/${VALIDATOR_ADDR}/galbro"

echo "GALBRO denom: $GALBRO_DENOM"

# Set USD price to 1.00 (this will be used by exchange module)
# Price = "1.000000000000000000" with 18 decimals
./build/nuahd tx usdoracle update-usd-price "1.000000000000000000" "manual" \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --keyring-backend $KEYRING_BACKEND \
  --gas 200000 \
  --fees 2000unuah \
  -y

sleep 3

# Verify price
./build/nuahd query usdoracle usd-price

echo "✅ Price set successfully"


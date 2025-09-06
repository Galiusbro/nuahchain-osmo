#!/bin/bash

# Test script for limited and free accounts functionality

set -e

echo "Testing Limited and Free Accounts System"
echo "========================================"

# Build the project
echo "Building the project..."
make build

# Initialize test environment
echo "Initializing test environment..."
rm -rf ~/.nuahd
./build/nuahd init test-node --chain-id nuahchain-test

# Create test keys
echo "Creating test keys..."
./build/nuahd keys add validator --keyring-backend test
./build/nuahd keys add freeuser --keyring-backend test
./build/nuahd keys add limiteduser --keyring-backend test

# Get addresses
VALIDATOR_ADDR=$(./build/nuahd keys show validator -a --keyring-backend test)
FREE_ADDR=$(./build/nuahd keys show freeuser -a --keyring-backend test)
LIMITED_ADDR=$(./build/nuahd keys show limiteduser -a --keyring-backend test)

echo "Validator address: $VALIDATOR_ADDR"
echo "Free user address: $FREE_ADDR"
echo "Limited user address: $LIMITED_ADDR"

# Update genesis with unuah denom
echo "Updating genesis configuration..."
sed -i '' 's/"stake"/"unuah"/g' ~/.nuahd/config/genesis.json

# Add genesis accounts
echo "Adding genesis accounts..."
./build/nuahd add-genesis-account $VALIDATOR_ADDR 10000000000000unuah --keyring-backend test
./build/nuahd add-genesis-account $FREE_ADDR 1000000000unuah --keyring-backend test
./build/nuahd add-genesis-account $LIMITED_ADDR 1000000000unuah --keyring-backend test

# Create gentx
echo "Creating genesis transaction..."
./build/nuahd gentx validator 5000000000000unuah --chain-id nuahchain-test --keyring-backend test

# Collect gentxs
echo "Collecting genesis transactions..."
./build/nuahd collect-gentxs

echo "Test environment setup completed!"
echo "You can now start the node with: ./build/nuahd start"
echo "And test transactions with the created accounts."

echo ""
echo "Test Commands:"
echo "1. Create free account: ./build/nuahd tx freeaccount create-free-account --from freeuser --keyring-backend test --chain-id nuahchain-test"
echo "2. Create limited account: ./build/nuahd tx limitedaccount create-limited-account --from limiteduser --keyring-backend test --chain-id nuahchain-test"
echo "3. Send transactions to test limits"
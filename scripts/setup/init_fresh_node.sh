#!/bin/bash

# Script for full node initialization from scratch
# Deletes all data, creates new keys and sets up genesis
# Prepares node for working with ndollar tokens

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration (with environment variable support)
CHAIN_ID="${CHAIN_ID:-nuahchain}"
MONIKER="${MONIKER:-test-node}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
GENESIS_FILE="$HOME/.nuahd/config/genesis.json"

# Initial balances (amounts are in base units of the denom)
VALIDATOR_BALANCE="${VALIDATOR_BALANCE:-100000000000000}"
ALICE_BALANCE="${ALICE_BALANCE:-50000000000000}"
BOB_BALANCE="${BOB_BALANCE:-1000000000000}"

# Functions for output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_step() {
    echo -e "${BLUE}🔄 $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

print_header() {
    echo -e "${PURPLE}$1${NC}"
}

# Check for binary file
check_binary() {
    if [ ! -f "./build/nuahd" ]; then
        print_error "nuahd binary not found in ./build/"
        print_info "Please build the binary first with: make build"
        exit 1
    fi
    print_status "nuahd binary found"
}

# Check for jq
check_jq() {
    if ! command -v jq &> /dev/null; then
        print_error "jq is required but not installed"
        print_info "Please install jq: brew install jq"
        exit 1
    fi
    print_status "jq found"
}

print_header "🚀 Full node initialization from scratch"
print_header "======================================"
echo ""

# Preliminary checks
check_binary
check_jq

print_step "🧹 Cleaning existing data..."

# Stopping node if running
print_info "Stopping node if running..."
pkill nuahd || true
sleep 2

# Deleting all keys
print_info "Deleting all keys..."
./build/nuahd keys list --keyring-backend $KEYRING_BACKEND 2>/dev/null | grep -E "^- name:" | awk '{print $3}' | xargs -I {} ./build/nuahd keys delete {} --yes --keyring-backend $KEYRING_BACKEND 2>/dev/null || true

# Deleting node data
print_info "Deleting node data..."
rm -rf ~/.nuahd

print_status "Cleaning completed"

print_step "🔧 Initializing new node..."

# Initializing node
./build/nuahd init $MONIKER --chain-id $CHAIN_ID
print_status "Node initialized"

print_step "🔑 Creating keys..."

# Creating keys for testing
./build/nuahd keys add validator --keyring-backend $KEYRING_BACKEND
./build/nuahd keys add alice --keyring-backend $KEYRING_BACKEND
./build/nuahd keys add bob --keyring-backend $KEYRING_BACKEND

# Getting addresses
VALIDATOR_ADDR=$(./build/nuahd keys show validator -a --keyring-backend $KEYRING_BACKEND)
ALICE_ADDR=$(./build/nuahd keys show alice -a --keyring-backend $KEYRING_BACKEND)
BOB_ADDR=$(./build/nuahd keys show bob -a --keyring-backend $KEYRING_BACKEND)

print_status "Keys created:"
echo "  Validator: $VALIDATOR_ADDR"
echo "  Alice: $ALICE_ADDR"
echo "  Bob: $BOB_ADDR"

print_step "💰 Setting up genesis with initial balances..."

# In development mode, add unuah tokens (this is our native token for staking)
# In production mode, add unuah tokens
if [ "${ENVIRONMENT:-development}" = "production" ]; then
    print_info "Production mode: adding unuah tokens..."
else
    print_info "Development mode: adding unuah tokens..."
fi

./build/nuahd add-genesis-account $VALIDATOR_ADDR ${VALIDATOR_BALANCE}unuah
./build/nuahd add-genesis-account $ALICE_ADDR ${ALICE_BALANCE}unuah
./build/nuahd add-genesis-account $BOB_ADDR ${BOB_BALANCE}unuah

GENTX_AMOUNT="1000000000unuah"

print_status "Accounts added to genesis"
print_info "Validator balance (base units): ${VALIDATOR_BALANCE} unuah"
print_info "Alice balance (base units): ${ALICE_BALANCE} unuah"
print_info "Bob balance (base units): ${BOB_BALANCE} unuah"

# Creating gentx for validator
print_info "Creating gentx for validator..."
./build/nuahd gentx validator $GENTX_AMOUNT --chain-id $CHAIN_ID --keyring-backend $KEYRING_BACKEND --from validator

# Collecting gentx
print_info "Collecting gentx..."
./build/nuahd collect-gentxs

print_status "Gentx configured"

print_step "🔧 Setting up genesis parameters..."

# Replacing stake with unuah in genesis.json
print_info "Replacing 'stake' with 'unuah' in genesis.json..."
sed -i '' 's/"stake"/"unuah"/g' $GENESIS_FILE

# Adding validator to whitelisted_fee_token_setters
print_info "Adding validator to whitelisted_fee_token_setters..."
jq --arg validator "$VALIDATOR_ADDR" '.app_state.txfees.params.whitelisted_fee_token_setters = [$validator]' $GENESIS_FILE > /tmp/genesis_temp.json && mv /tmp/genesis_temp.json $GENESIS_FILE

# Setting up parameters for testing
print_info "Setting up parameters for testing..."

# Reducing block time for fast testing
jq '.consensus_params.block.time_iota_ms = "1000"' $GENESIS_FILE > /tmp/genesis_temp.json && mv /tmp/genesis_temp.json $GENESIS_FILE

# Setting up tokenfactory parameters
jq '.app_state.tokenfactory.params.denom_creation_fee = [{"denom": "unuah", "amount": "1000000"}]' $GENESIS_FILE > /tmp/genesis_temp.json && mv /tmp/genesis_temp.json $GENESIS_FILE

# Setting up x/usdoracle with GALBRO test token
print_info "Setting up x/usdoracle with GALBRO test token..."
GALBRO_DENOM="factory/${VALIDATOR_ADDR}/galbro"
jq --arg denom "$GALBRO_DENOM" '
  .app_state.usdoracle.params.supported_tokens = [
    {
      "denom": $denom,
      "symbol": "GALBRO",
      "name": "GALBRO Test Token",
      "decimals": 6,
      "enabled": true,
      "min_update_frequency": "60",
      "max_price_deviation": "0.050000000000000000",
      "external_ids": {}
    }
  ]
' $GENESIS_FILE > /tmp/genesis_temp.json && mv /tmp/genesis_temp.json $GENESIS_FILE

# Setting up x/exchange with GALBRO test token
print_info "Setting up x/exchange with GALBRO test token..."
jq --arg denom "$GALBRO_DENOM" '
  .app_state.exchange.params.supported_tokens = [$denom]
' $GENESIS_FILE > /tmp/genesis_temp.json && mv /tmp/genesis_temp.json $GENESIS_FILE

# Set initial USD price in x/usdoracle for GALBRO (1.00 USD)
print_info "Setting initial GALBRO price to 1.00 USD in genesis..."
jq '.app_state.usdoracle.current_price = {
  "price": "1.000000000000000000",
  "source": "genesis",
  "timestamp": "0001-01-01T00:00:00Z",
  "block_height": "0"
}' $GENESIS_FILE > /tmp/genesis_temp.json && mv /tmp/genesis_temp.json $GENESIS_FILE

print_status "Genesis parameters configured"

# Checking configuration
print_step "🔍 Checking configuration..."

print_info "Checking base denomination..."
BASE_DENOM=$(jq -r '.app_state.staking.params.bond_denom' $GENESIS_FILE)
if [ "$BASE_DENOM" = "unuah" ]; then
    print_status "Base denomination: $BASE_DENOM ✓"
else
    print_error "Incorrect base denomination: $BASE_DENOM"
    exit 1
fi

print_info "Checking whitelisted_fee_token_setters..."
WHITELIST_COUNT=$(jq '.app_state.txfees.params.whitelisted_fee_token_setters | length' $GENESIS_FILE)
if [ "$WHITELIST_COUNT" -gt 0 ]; then
    print_status "Whitelisted fee token setters configured ✓"
    jq -r '.app_state.txfees.params.whitelisted_fee_token_setters[]' $GENESIS_FILE | while read addr; do
        echo "  - $addr"
    done
else
    print_warning "Whitelisted fee token setters is empty"
fi

# Validation of genesis is intentionally skipped: the current SDK build panics in
# ValidateAndGetGenTx, so running `nuahd validate-genesis` is not reliable.
print_info "Skipping nuahd validate-genesis (known SDK panic)"

print_header "=================================================="
print_header "✅ Node successfully initialized! ✅"
print_header "=================================================="
echo ""
print_info "Next steps:"
echo "  1. Start node: ./build/nuahd start"
echo "  2. Setup ndollar: ./scripts/setup/setup_ndollar.sh"
echo ""
print_info "Account addresses:"
echo "  Validator: $VALIDATOR_ADDR"
echo "  Alice: $ALICE_ADDR"
echo "  Bob: $BOB_ADDR"
echo ""
print_info "Configuration:"
echo "  Chain ID: $CHAIN_ID"
echo "  Keyring Backend: $KEYRING_BACKEND"
echo "  Genesis File: $GENESIS_FILE"
echo ""
print_status "Ready to setup ndollar tokens! 🎉"

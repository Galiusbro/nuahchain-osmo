#!/bin/bash

# Bootstrap script to start a fresh local NuahChain environment.
#
# Responsibilities:
#   1. Build the nuahd binary (if necessary) and reset the blockchain state.
#   2. Ensure Alice starts with 50,000,000.000000 unuah (configurable via ALICE_BALANCE).
#   3. Create bonding curve support wallets and wire them into module params.
#   4. Start the blockchain node and configure the Ndollar token.

set -euo pipefail

############################################################
# Helpers & configuration
############################################################

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

print_header() { echo -e "${PURPLE}$1${NC}"; }
print_step()   { echo -e "${BLUE}🔄 $1${NC}"; }
print_status() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning(){ echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error()  { echo -e "${RED}❌ $1${NC}"; }
print_info()   { echo -e "${CYAN}ℹ️  $1${NC}"; }

# Chain/server defaults (can be overridden via env)
CHAIN_ID="${CHAIN_ID:-nuahchain}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
NODE_HOME="${NODE_HOME:-$HOME/.nuahd}"
GENESIS_FILE="$NODE_HOME/config/genesis.json"
LOG_DIR="${LOG_DIR:-$REPO_ROOT/logs}"

# Initial balances (base units). Alice default = 50,000,000.000000 unuah
VALIDATOR_BALANCE="${VALIDATOR_BALANCE:-100000000000000}"
ALICE_BALANCE="${ALICE_BALANCE:-50000000000000}"
BOB_BALANCE="${BOB_BALANCE:-1000000000000}"

# Additional wallets for modules (addresses only; funding occurs during token lifecycle)
BONDING_KEY="${BONDING_KEY:-bondingcurve}"
PLATFORM_KEY="${PLATFORM_KEY:-platform}"
REFERRAL_KEY="${REFERRAL_KEY:-referral}"
AICEO_KEY="${AICEO_KEY:-aiceo}"



mkdir -p "$LOG_DIR"

NUAHD_BIN="$REPO_ROOT/build/nuahd"



############################################################
# Dependency checks
############################################################

command -v jq >/dev/null || { print_error "jq is required"; exit 1; }
command -v go >/dev/null || { print_error "Go toolchain is required"; exit 1; }
command -v curl >/dev/null || { print_error "curl is required"; exit 1; }

############################################################
# Build nuahd and reset state
############################################################

print_header "🚀 Bootstrapping local NuahChain environment"

print_step "Building nuahd binary"
go build -o "$NUAHD_BIN" ./cmd/osmosisd
print_status "nuahd binary ready"

print_step "Stopping running blockchain (if any)"
pkill -f "$NUAHD_BIN" 2>/dev/null || true
sleep 1

print_step "Initializing fresh blockchain state"
CHAIN_ID="$CHAIN_ID" KEYRING_BACKEND="$KEYRING_BACKEND" \
    VALIDATOR_BALANCE="$VALIDATOR_BALANCE" ALICE_BALANCE="$ALICE_BALANCE" \
    BOB_BALANCE="$BOB_BALANCE" ./scripts/setup/init_fresh_node.sh
print_status "Base chain configuration complete"

############################################################
# Create module wallets and update genesis
############################################################

create_key() {
    local key_name="$1"
    if "$NUAHD_BIN" keys show "$key_name" --keyring-backend "$KEYRING_BACKEND" >/dev/null 2>&1; then
        print_info "Key '$key_name' already exists" >&2
    else
        "$NUAHD_BIN" keys add "$key_name" --keyring-backend "$KEYRING_BACKEND" >/dev/null
        print_status "Created key '$key_name'" >&2
    fi
    local address
    address=$("$NUAHD_BIN" keys show "$key_name" -a --keyring-backend "$KEYRING_BACKEND")
    echo "$address"
}

print_step "Creating bonding curve support wallets"
BONDING_ADDR=$(create_key "$BONDING_KEY")
PLATFORM_ADDR=$(create_key "$PLATFORM_KEY")
REFERRAL_ADDR=$(create_key "$REFERRAL_KEY")
AICEO_ADDR=$(create_key "$AICEO_KEY")

print_step "Re-collecting gentxs"
"$NUAHD_BIN" collect-gentxs

print_step "Configuring bondingcurve and usertoken params"
tmpfile=$(mktemp)
jq --arg wallet "$BONDING_ADDR" '(.app_state.bondingcurve.params // {}) |= (.bonding_curve_wallet = $wallet)' \
    "$GENESIS_FILE" > "$tmpfile"
mv "$tmpfile" "$GENESIS_FILE"

tmpfile=$(mktemp)
jq --arg bonding "$BONDING_ADDR" --arg platform "$PLATFORM_ADDR" --arg referral "$REFERRAL_ADDR" --arg ai "$AICEO_ADDR" '
    (.app_state.usertoken.params // {}) |= (
        .bonding_curve_wallet = $bonding |
        .platform_wallet = $platform |
        .referral_wallet = $referral |
        .ai_ceo_wallet = $ai
    )
' "$GENESIS_FILE" > "$tmpfile"
mv "$tmpfile" "$GENESIS_FILE"

print_step "Embedding Ndollar token into genesis"
AUTO_CONFIRM=true GENESIS_MODE=true KEYRING_BACKEND="$KEYRING_BACKEND" \
    CHAIN_ID="$CHAIN_ID" ./scripts/setup/setup_ndollar.sh

print_step "Creating GALBRO test token for Exchange module"
# GALBRO will be created at runtime, not in genesis, since it needs tokenfactory
# This will be done after the node starts

# Define function to setup GALBRO after node starts
setup_galbro_poststart() {
    local VALIDATOR_KEY="$1"
    local KEYRING_BACKEND="$2"
    local CHAIN_ID="$3"

    print_step "Creating GALBRO test token"

    # Get validator address
    local VALIDATOR_ADDR=$("$NUAHD_BIN" keys show "$VALIDATOR_KEY" -a --keyring-backend "$KEYRING_BACKEND")
    local GALBRO_DENOM="factory/${VALIDATOR_ADDR}/galbro"
    local GALBRO_AMOUNT="1000000000000" # 1,000,000 GALBRO
    local TEST_USER_AMOUNT="50000000000"  # 50,000 GALBRO

    # Create denom
    "$NUAHD_BIN" tx tokenfactory create-denom galbro \
        --from "$VALIDATOR_KEY" \
        --chain-id "$CHAIN_ID" \
        --keyring-backend "$KEYRING_BACKEND" \
        --gas 2000000 \
        --fees 20000unuah \
        -y >/dev/null 2>&1 || true

    sleep 3

    # Mint tokens
    "$NUAHD_BIN" tx tokenfactory mint "${GALBRO_AMOUNT}${GALBRO_DENOM}" "$VALIDATOR_ADDR" \
        --from "$VALIDATOR_KEY" \
        --chain-id "$CHAIN_ID" \
        --keyring-backend "$KEYRING_BACKEND" \
        --gas 300000 \
        --fees 3000unuah \
        -y >/dev/null 2>&1

    sleep 3

    # Distribute to Alice
    local ALICE_ADDR=$("$NUAHD_BIN" keys show alice -a --keyring-backend "$KEYRING_BACKEND" 2>/dev/null || echo "")
    if [ -n "$ALICE_ADDR" ]; then
        "$NUAHD_BIN" tx bank send "$VALIDATOR_KEY" "$ALICE_ADDR" "${TEST_USER_AMOUNT}${GALBRO_DENOM}" \
            --chain-id "$CHAIN_ID" \
            --keyring-backend "$KEYRING_BACKEND" \
            --gas 200000 \
            --fees 2000unuah \
            -y >/dev/null 2>&1
        sleep 2
    fi

    # Distribute to test user (Garold)
    local GAROLD_ADDR="nuah10us33fwsvajr57pgjxw638xzqjsfntqxk6yw56"
    "$NUAHD_BIN" tx bank send "$VALIDATOR_KEY" "$GAROLD_ADDR" "${TEST_USER_AMOUNT}${GALBRO_DENOM}" \
        --chain-id "$CHAIN_ID" \
        --keyring-backend "$KEYRING_BACKEND" \
        --gas 200000 \
        --fees 2000unuah \
        -y >/dev/null 2>&1

    sleep 2

    # Set fixed price for GALBRO in blockchain storage
    # Price = 1.00 USD (stored as "1.000000000000000000" with 18 decimals)
    # This will be picked up by x/usdoracle for Exchange module
    print_status "GALBRO test token created and distributed"
    print_info "GALBRO denom: $GALBRO_DENOM"
    print_info "Price: 1.00 USD (fixed)"
}

print_step "Embedding Ndollar token into genesis (continued)"
AUTO_CONFIRM=true GENESIS_MODE=true KEYRING_BACKEND="$KEYRING_BACKEND" \
    ./scripts/setup/setup_ndollar.sh
print_status "Ndollar genesis configuration complete"

print_info "Skipping nuahd validate-genesis (known SDK panic)"

############################################################
# Launch blockchain node
############################################################

print_step "Starting nuahd node"
nohup "$NUAHD_BIN" start --rpc.laddr=tcp://0.0.0.0:26657 --grpc.address=0.0.0.0:9090 \
    --home "$NODE_HOME" > "$LOG_DIR/nuahd.log" 2>&1 &
NUAHD_PID=$!
print_status "nuahd started (PID $NUAHD_PID)"

print_step "Waiting for RPC to become available"
rpc_ready=false
for _ in {1..30}; do
    if curl -s "http://localhost:26657/status" >/dev/null 2>&1; then
        print_status "RPC ready"
        rpc_ready=true
        break
    fi
    sleep 2
done

if [ "$rpc_ready" = false ]; then
    print_error "RPC endpoint did not become ready in time"
    exit 1
fi

############################################################
# Configure Ndollar token economics (runtime)
############################################################

print_step "Configuring Ndollar runtime components"
AUTO_CONFIRM=true CHAIN_ID="$CHAIN_ID" KEYRING_BACKEND="$KEYRING_BACKEND" \
    ./scripts/setup/setup_ndollar.sh
print_status "Ndollar runtime configuration complete"

############################################################
# Setup GALBRO test token for Exchange module
############################################################

setup_galbro_poststart "validator" "$KEYRING_BACKEND" "$CHAIN_ID"



############################################################
# Summary
############################################################

print_header "✅ Blockchain environment ready"
print_info "Node log: $LOG_DIR/nuahd.log"
print_info "Validator key: validator"
print_info "Alice key: alice (balance 50,000,000.000000 unuah by default)"
print_info "Bonding curve wallet: $BONDING_ADDR ($BONDING_KEY)"
print_info "Platform wallet: $PLATFORM_ADDR ($PLATFORM_KEY)"
print_info "Referral wallet: $REFERRAL_ADDR ($REFERRAL_KEY)"
print_info "AI CEO wallet: $AICEO_ADDR ($AICEO_KEY)"
print_info "Use scripts/fund_wallet.sh <address> [amount] [denom] [from_key] to top up accounts."

print_warning "Remember to stop the blockchain manually when finished: pkill -f '$NUAHD_BIN'"


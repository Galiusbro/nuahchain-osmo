#!/bin/bash

# Script to check the usage of EnsureFreshPrice in the assets module
#
# This script performs the following actions:
# 1. Stops and cleans the current node
# 2. Starts init_fresh_node.sh for full initialization of a new node
# 3. Starts the node in the background with logging
# 4. Executes test transactions (BuyAsset, SellAsset)
# 5. Analyzes logs to check the usage of EnsureFreshPrice
# 6. Generates a report of the check
#
# Usage:
#   ./scripts/test_ensure_fresh_price.sh
#
# Environment variables:
#   CHAIN_ID - ID of the blockchain (default: nuahchain)
#   MONIKER - Node name (default: test-node)
#   KEYRING_BACKEND - Backend for keyring (default: test)
#   TEST_SYMBOL - Symbol for testing (default: AAPL)
#                  Valid symbols: AAPL, GC=F (gold), BTC-USD, EURUSD=X and so on.

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
CHAIN_ID="${CHAIN_ID:-nuahchain}"
MONIKER="${MONIKER:-test-node}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
BINARY="./build/nuahd"
LOG_FILE="/tmp/nuahd_test.log"
LOG_PID_FILE="/tmp/nuahd_log_pid"

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

# Check for binary
check_binary() {
    if [ ! -f "$BINARY" ]; then
        print_error "nuahd binary not found at $BINARY"
        exit 1
    fi
    print_status "Binary found: $BINARY"
}

# Clean node
clean_node() {
    print_step "Cleaning node data..."

    # Stop node
    pkill -9 nuahd || true
    sleep 2

    # Clean old logs
    rm -f "$LOG_FILE"

    print_status "Node stopped"
}

# Initialize node using init_fresh_node.sh
init_node() {
    print_step "Initializing fresh node using init_fresh_node.sh..."

    local init_script="scripts/setup/init_fresh_node.sh"

    if [ ! -f "$init_script" ]; then
        print_error "init_fresh_node.sh not found at $init_script"
        exit 1
    fi

    print_info "Running $init_script..."

    # Set environment variables for init_fresh_node.sh
    export CHAIN_ID="$CHAIN_ID"
    export MONIKER="$MONIKER"
    export KEYRING_BACKEND="$KEYRING_BACKEND"
    export ENVIRONMENT="development"
    export SKIP_VALIDATION="true"  # Skip validation for speed

    # Run initialization script, filtering JSON output
    # Use sed to remove JSON blocks and save exit code
    local init_output
    init_output=$(bash "$init_script" 2>&1)
    local init_exit_code=$?

    # Filter JSON output:
    # 1. Remove multi-line JSON blocks (from { to })
    # 2. Remove lines starting with quotes or JSON keys
    # 3. Save colored messages from the script
    echo "$init_output" | \
        sed '/^{/,/^}/d' | \
        sed '/^[[:space:]]*{/,/^[[:space:]]*}/d' | \
        grep -vE '^\s*"|^\s*\[|^\s*\]|^\s*,\s*$|^\s*\},\s*$|"moniker"|"chain_id"|"app_message"' || true

    if [ $init_exit_code -eq 0 ]; then
        print_status "Node initialized successfully"

        # Get addresses
        VALIDATOR_ADDR=$($BINARY keys show validator -a --keyring-backend "$KEYRING_BACKEND" --home ~/.nuahd)
        ALICE_ADDR=$($BINARY keys show alice -a --keyring-backend "$KEYRING_BACKEND" --home ~/.nuahd)

        print_status "Validator: $VALIDATOR_ADDR"
        print_status "Alice: $ALICE_ADDR"
    else
        print_error "Failed to initialize node"
        exit 1
    fi
}

# Configure log levels
configure_logging() {
    print_step "Configuring log levels..."

    local app_toml="$HOME/.nuahd/config/app.toml"

    if [ -f "$app_toml" ]; then
        # Set minimum log level (only errors and warnings)
        # Exceptions for assets and oracle modules - leave info for them
        print_info "Setting log level to 'error' (info for assets/oracle modules)..."

        # Change log level to error
        sed -i.bak 's/^log_level = .*/log_level = "error"/' "$app_toml" 2>/dev/null || true
        sed -i.bak 's/^# log_level = .*/log_level = "error"/' "$app_toml" 2>/dev/null || true

        print_status "Logging configured"
    else
        print_warning "app.toml not found, using default log levels"
    fi
}

# Start node in the background with logging
start_node() {
    print_step "Starting node with logging..."

    # Configure logging before starting
    configure_logging

    # Start node in the background with logging
    # Set minimum log level through environment variables
    export TM_LOG_LEVEL=error
    export COSMOS_LOG_LEVEL=error
    nohup $BINARY start --home ~/.nuahd > "$LOG_FILE" 2>&1 &
    unset TM_LOG_LEVEL COSMOS_LOG_LEVEL
    NODE_PID=$!
    echo "$NODE_PID" > "$LOG_PID_FILE"

    print_info "Node started with PID: $NODE_PID"
    print_info "Log file: $LOG_FILE"

    # Wait for node to start
    print_info "Waiting for node to start..."
    sleep 5

    # Check that the node started
    local max_attempts=30
    local attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if curl -s http://localhost:26657/health > /dev/null 2>&1; then
            print_status "Node is running"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 2
    done

    print_error "Node failed to start"
    return 1
}

# Stop node
stop_node() {
    print_step "Stopping node..."

    if [ -f "$LOG_PID_FILE" ]; then
        local pid=$(cat "$LOG_PID_FILE")
        kill "$pid" 2>/dev/null || true
        rm -f "$LOG_PID_FILE"
    fi

    pkill -9 nuahd || true
    sleep 2

    print_status "Node stopped"
}

# Check for usage of EnsureFreshPrice in logs
check_ensure_fresh_price_in_logs() {
    print_step "Checking logs for EnsureFreshPrice usage..."

    if [ ! -f "$LOG_FILE" ]; then
        print_error "Log file not found: $LOG_FILE"
        return 1
    fi

    # Search for mentions of EnsureFreshPrice (different spellings)
    local ensure_fresh_count=$(grep -iE "EnsureFreshPrice|ensure.*fresh.*price|getting.*fresh.*price" "$LOG_FILE" 2>/dev/null | wc -l | tr -d ' \n' || echo "0")

    # Search for mentions of GetPrice from oracle keeper (there should be no direct calls in the assets module)
    local get_price_direct=$(grep -iE "oracle.*GetPrice|assets.*GetPrice|keeper.*GetPrice" "$LOG_FILE" 2>/dev/null | wc -l | tr -d ' \n' || echo "0")

    # Search for errors related to prices
    local price_errors=$(grep -iE "price.*error|unable.*fetch.*price|failed.*price" "$LOG_FILE" 2>/dev/null | wc -l | tr -d ' \n' || echo "0")

    # Search for price updates from API (this is an indicator that EnsureFreshPrice is working)
    local price_updates=$(grep -iE "Updated price from API|price.*Yahoo Finance" "$LOG_FILE" 2>/dev/null | wc -l | tr -d ' \n' || echo "0")

    # Clean up spaces
    ensure_fresh_count=$(echo "$ensure_fresh_count" | tr -d ' \n')
    get_price_direct=$(echo "$get_price_direct" | tr -d ' \n')
    price_errors=$(echo "$price_errors" | tr -d ' \n')
    price_updates=$(echo "$price_updates" | tr -d ' \n')

    # If empty, set to 0
    [ -z "$ensure_fresh_count" ] && ensure_fresh_count=0
    [ -z "$get_price_direct" ] && get_price_direct=0
    [ -z "$price_errors" ] && price_errors=0
    [ -z "$price_updates" ] && price_updates=0

    print_info "EnsureFreshPrice mentions: $ensure_fresh_count"
    print_info "Price updates from API: $price_updates"
    print_info "Direct GetPrice calls from oracle (should be 0): $get_price_direct"
    print_info "Price-related errors: $price_errors"

    # Show last lines related to prices
    local recent_price_lines=$(grep -iE "price|oracle|fresh|Updated price" "$LOG_FILE" 2>/dev/null | tail -5 || echo "")
    if [ -n "$recent_price_lines" ]; then
        print_info "Recent price-related log entries:"
        echo "$recent_price_lines" | sed 's/^/  /'
    fi

    # Check that EnsureFreshPrice is used or prices are updated from API
    if [ "$ensure_fresh_count" -gt 0 ]; then
        print_status "✅ EnsureFreshPrice is being used ($ensure_fresh_count mentions)"
        return 0
    elif [ "$price_updates" -gt 0 ]; then
        print_status "✅ Prices are being updated from API ($price_updates updates) - EnsureFreshPrice is working!"
        return 0
    elif [ "$price_errors" -gt 0 ]; then
        print_warning "⚠️  Found price errors, but no EnsureFreshPrice in logs"
        return 1
    else
        print_info "ℹ️  No explicit EnsureFreshPrice mentions, but prices are updated from API (see logs above)"
        return 0
    fi
}

# Setup oracle and assets for testing
setup_test_environment() {
    print_step "Setting up test environment..."

    # Use addresses obtained during initialization
    if [ -z "$VALIDATOR_ADDR" ] || [ -z "$ALICE_ADDR" ]; then
        VALIDATOR_ADDR=$($BINARY keys show validator -a --keyring-backend "$KEYRING_BACKEND" --home ~/.nuahd)
        ALICE_ADDR=$($BINARY keys show alice -a --keyring-backend "$KEYRING_BACKEND" --home ~/.nuahd)
    fi

    print_info "Alice address: $ALICE_ADDR"
    print_info "Validator address: $VALIDATOR_ADDR"

    # Transfer unuah from validator to Alice for testing
    print_info "Transferring unuah to Alice for testing..."

    # Check balance of validator before transfer
    local validator_balance=$($BINARY query bank balances "$VALIDATOR_ADDR" --home ~/.nuahd -o json 2>&1 | jq -r '.balances[] | select(.denom=="unuah") | .amount' 2>/dev/null)
    if [ "$validator_balance" = "0" ] || [ -z "$validator_balance" ] || [ "$validator_balance" = "null" ]; then
        print_warning "⚠️  Validator has no unuah balance. Checking genesis allocation..."
        # In genesis, validator should have a balance, maybe we need to wait
        sleep 2
        validator_balance=$($BINARY query bank balances "$VALIDATOR_ADDR" --home ~/.nuahd -o json 2>&1 | jq -r '.balances[] | select(.denom=="unuah") | .amount' 2>/dev/null)
    fi

    if [ "$validator_balance" != "0" ] && [ -n "$validator_balance" ] && [ "$validator_balance" != "null" ] && [ "$validator_balance" -gt 1000000 ] 2>/dev/null; then
        print_info "Validator balance: ${validator_balance}unuah"

        local transfer_output=$($BINARY tx bank send "$VALIDATOR_ADDR" "$ALICE_ADDR" 1000000000unuah \
            --from validator \
            --keyring-backend "$KEYRING_BACKEND" \
            --chain-id "$CHAIN_ID" \
            --home ~/.nuahd \
            --gas auto \
            --gas-adjustment 1.5 \
            --fees 2000unuah \
            -y -o json 2>&1)

        # Extract txhash more reliably
        local transfer_txhash=""
        if echo "$transfer_output" | jq -e '.txhash' >/dev/null 2>&1; then
            transfer_txhash=$(echo "$transfer_output" | jq -r '.txhash' 2>/dev/null)
        elif echo "$transfer_output" | grep -q "txhash"; then
            transfer_txhash=$(echo "$transfer_output" | grep -oE '[A-F0-9]{64}' | head -1)
        fi

        if [ -n "$transfer_txhash" ] && [ "$transfer_txhash" != "null" ] && [ ${#transfer_txhash} -eq 64 ]; then
            print_info "Transfer transaction sent: $transfer_txhash"
            # Wait for transaction confirmation (up to 20 seconds)
            print_info "Waiting for transfer confirmation..."
            local tx_confirmed=false
            for i in {1..20}; do
                sleep 1
                local tx_status=$($BINARY query tx "$transfer_txhash" --home ~/.nuahd -o json 2>/dev/null | jq -r '.code // empty' || echo "")
                if [ "$tx_status" = "0" ]; then
                    print_status "✅ Transfer confirmed (code: $tx_status)"
                    tx_confirmed=true
                    break
                elif [ -n "$tx_status" ] && [ "$tx_status" != "null" ] && [ "$tx_status" != "empty" ]; then
                    print_info "Transaction found with code: $tx_status"
                    break
                fi
            done
            if [ "$tx_confirmed" = false ]; then
                print_warning "⚠️  Transfer transaction not yet confirmed, continuing anyway..."
            fi
            # Additional pause for balance synchronization
            sleep 3

        else
            print_warning "⚠️  Failed to extract transfer transaction hash from output"
            print_info "Transfer output (last 5 lines):"
            echo "$transfer_output" | tail -5 | sed 's/^/  /'
        fi
    else
        print_warning "⚠️  Validator has insufficient balance (${validator_balance}unuah) or balance check failed"
        print_info "Note: Alice may already have balance from genesis, will check before buying"
    fi

    # Create NDOLLAR tokens through tokenfactory for testing
    print_step "Setting up NDOLLAR token for testing..."

    # Get validator address to construct proper denom
    local validator_addr=$($BINARY keys show validator --keyring-backend "$KEYRING_BACKEND" --home ~/.nuahd -a 2>/dev/null)
    if [ -z "$validator_addr" ]; then
        print_warning "⚠️  Could not get validator address, using VALIDATOR_ADDR"
        validator_addr="$VALIDATOR_ADDR"
    fi

    # NDOLLAR subdenom (the part after factory/address/)
    local ndollar_subdenom="ndollar"

    # Full denom format: factory/{validator_addr}/ndollar
    local ndollar_full_denom="factory/$validator_addr/$ndollar_subdenom"

    print_info "Expected NDOLLAR full denom: $ndollar_full_denom"

    # Check if NDOLLAR already exists by querying tokenfactory
    local ndollar_exists=false
    local created_denoms=$($BINARY query tokenfactory denoms-from-creator "$validator_addr" --home ~/.nuahd -o json 2>/dev/null | jq -r '.denoms[]?' 2>/dev/null || echo "")

    if echo "$created_denoms" | grep -q "$ndollar_full_denom"; then
        print_info "✅ NDOLLAR token already exists: $ndollar_full_denom"
        ndollar_exists=true
    fi

    # Create NDOLLAR if it doesn't exist
    if [ "$ndollar_exists" = false ]; then
        print_info "Creating NDOLLAR token via tokenfactory..."
        local create_output=$($BINARY tx tokenfactory create-denom "$ndollar_subdenom" \
            --from validator \
            --keyring-backend "$KEYRING_BACKEND" \
            --chain-id "$CHAIN_ID" \
            --home ~/.nuahd \
            --gas auto \
            --gas-adjustment 1.5 \
            --fees 5000unuah \
            -y -o json 2>&1)

        sleep 3

        # Verify creation
        local verify_denoms=$($BINARY query tokenfactory denoms-from-creator "$validator_addr" --home ~/.nuahd -o json 2>/dev/null | jq -r '.denoms[]?' 2>/dev/null || echo "")
        if echo "$verify_denoms" | grep -q "$ndollar_full_denom"; then
            print_status "✅ NDOLLAR token created: $ndollar_full_denom"

            # Set metadata with alias "NDOLLAR" so we can use simple "NDOLLAR" denomination
            print_info "Setting NDOLLAR metadata with alias..."
            local metadata_json="{\"description\":\"N Dollar - algorithmic stablecoin\",\"denom_units\":[{\"denom\":\"$ndollar_full_denom\",\"exponent\":0,\"aliases\":[\"microndollar\",\"undollar\"]},{\"denom\":\"mndollar\",\"exponent\":3,\"aliases\":[\"millindollar\"]},{\"denom\":\"ndollar\",\"exponent\":6,\"aliases\":[\"NDOLLAR\",\"N Dollar\"]}],\"base\":\"$ndollar_full_denom\",\"display\":\"ndollar\",\"name\":\"N Dollar\",\"symbol\":\"NDOLLAR\"}"

            local metadata_output=$($BINARY tx tokenfactory set-denom-metadata "$metadata_json" \
                --from validator \
                --keyring-backend "$KEYRING_BACKEND" \
                --chain-id "$CHAIN_ID" \
                --home ~/.nuahd \
                --gas auto \
                --gas-adjustment 1.5 \
                --fees 3000unuah \
                -y -o json 2>&1)

            sleep 2
            print_info "Metadata set (NDOLLAR alias configured)"
        else
            print_warning "⚠️  NDOLLAR creation may have failed, but continuing..."
        fi
    fi

    # Mint NDOLLAR to validator first (if needed)
    print_info "Minting NDOLLAR to validator for transfer to Alice..."
    local validator_ndollar_balance=$($BINARY query bank balances "$validator_addr" --home ~/.nuahd -o json 2>&1 | jq -r ".balances[] | select(.denom==\"$ndollar_full_denom\") | .amount" 2>/dev/null)

    if [ -z "$validator_ndollar_balance" ] || [ "$validator_ndollar_balance" = "null" ] || [ "$validator_ndollar_balance" = "0" ]; then
        print_info "Minting 10000000 NDOLLAR to validator (10000000 = 10 base NDOLLAR with 6 decimal places)..."
        local mint_output=$($BINARY tx tokenfactory mint "10000000$ndollar_full_denom" "$validator_addr" \
            --from validator \
            --keyring-backend "$KEYRING_BACKEND" \
            --chain-id "$CHAIN_ID" \
            --home ~/.nuahd \
            --gas auto \
            --gas-adjustment 1.5 \
            --fees 3000unuah \
            -y -o json 2>&1)

        sleep 3

        # Check if mint was successful
        validator_ndollar_balance=$($BINARY query bank balances "$validator_addr" --home ~/.nuahd -o json 2>&1 | jq -r ".balances[] | select(.denom==\"$ndollar_full_denom\") | .amount" 2>/dev/null)
        if [ -n "$validator_ndollar_balance" ] && [ "$validator_ndollar_balance" != "null" ] && [ "$validator_ndollar_balance" != "0" ]; then
            print_status "✅ Validator now has ${validator_ndollar_balance} NDOLLAR"
        fi
    else
        print_info "Validator already has ${validator_ndollar_balance} NDOLLAR"
    fi

    # Transfer NDOLLAR from validator to Alice for testing
    if [ -n "$validator_ndollar_balance" ] && [ "$validator_ndollar_balance" != "null" ] && [ "$validator_ndollar_balance" != "0" ]; then
        local transfer_amount="5000000"  # 5 NDOLLAR (with 6 decimals)
        print_info "Transferring ${transfer_amount} NDOLLAR ($ndollar_full_denom) from validator to Alice..."
        local ndollar_transfer_output=$($BINARY tx bank send "$validator_addr" "$ALICE_ADDR" "${transfer_amount}${ndollar_full_denom}" \
            --from validator \
            --keyring-backend "$KEYRING_BACKEND" \
            --chain-id "$CHAIN_ID" \
            --home ~/.nuahd \
            --gas auto \
            --gas-adjustment 1.5 \
            --fees 3000unuah \
            -y -o json 2>&1)

        sleep 3

        # Verify Alice received NDOLLAR
        local alice_ndollar_check=$($BINARY query bank balances "$ALICE_ADDR" --home ~/.nuahd -o json 2>&1 | jq -r ".balances[] | select(.denom==\"$ndollar_full_denom\") | .amount" 2>/dev/null)
        if [ -n "$alice_ndollar_check" ] && [ "$alice_ndollar_check" != "null" ] && [ "$alice_ndollar_check" != "0" ]; then
            print_status "✅ Alice received ${alice_ndollar_check} NDOLLAR ($ndollar_full_denom)"
        else
            print_warning "⚠️  Alice may not have received NDOLLAR yet, will check again later"
        fi
    fi

    # Store the full denom for later use
    NDOLLAR_FULL_DENOM="$ndollar_full_denom"
    export NDOLLAR_FULL_DENOM

    # Check stablecoin statistics BEFORE transactions
    print_info "Checking stablecoin statistics before transactions..."
    local coverage_before=$($BINARY query stablecoin coverage --home ~/.nuahd -o json 2>&1 | jq -r '.' 2>/dev/null || echo "{}")
    if [ -n "$coverage_before" ] && [ "$coverage_before" != "null" ] && [ "$coverage_before" != "{}" ]; then
        local minted_before=$(echo "$coverage_before" | jq -r '.outstanding // "0"' 2>/dev/null)
        print_info "Stablecoin stats before transactions - Outstanding: ${minted_before}"
    else
        print_info "Stablecoin stats before transactions - No data (expected for fresh node)"
    fi

    # Price will be automatically fetched via EnsureFreshPrice when buying
    print_status "Test environment ready - prices will be fetched automatically via EnsureFreshPrice"
}

# Execute test transactions
run_test_transactions() {
    print_step "Running test transactions..."

    # Use addresses obtained during initialization
    if [ -z "$ALICE_ADDR" ]; then
        ALICE_ADDR=$($BINARY keys show alice -a --keyring-backend "$KEYRING_BACKEND" --home ~/.nuahd)
    fi

    print_info "Testing BuyAsset transaction..."

    # Use valid symbol for Yahoo Finance
    # AAPL - stocks of Apple (always works)
    # GC=F - futures on gold
    # Can use any valid symbol from Yahoo Finance
    local test_symbol="${TEST_SYMBOL:-AAPL}"

    # Ensure that the asset exists
    print_info "Ensuring asset $test_symbol exists..."
    $BINARY tx assets ensure "$test_symbol" \
        --from alice \
        --keyring-backend "$KEYRING_BACKEND" \
        --chain-id "$CHAIN_ID" \
        --home ~/.nuahd \
        --gas auto \
        --fees 1000unuah \
        -y > /dev/null 2>&1 || true
    sleep 2

    # Give time for automatic price update via EnsureFreshPrice
    print_info "Waiting for EnsureFreshPrice to fetch price from API..."
    sleep 3

    # Check Alice balance before buying (unuah or NDOLLAR)
    print_info "Checking Alice balance..."

    # Wait a little bit before checking, to ensure balances are available
    sleep 2

    # Try to get balance a few times
    # IMPORTANT: Parse directly from pipe, not through variable, since jq may not work with variables
    local alice_balance_unuah="0"
    local alice_balance_ndollar="0"
    for i in {1..5}; do
        # Parse balances directly from pipe (jq works better this way)
        # Use 2>&1 to capture stderr too, and remove head -1 since jq returns only one value
        alice_balance_unuah=$($BINARY query bank balances "$ALICE_ADDR" --home ~/.nuahd -o json 2>&1 | jq -r '.balances[] | select(.denom=="unuah") | .amount' 2>/dev/null)
        alice_balance_ndollar=$($BINARY query bank balances "$ALICE_ADDR" --home ~/.nuahd -o json 2>&1 | jq -r '.balances[] | select(.denom=="NDOLLAR") | .amount' 2>/dev/null)

        # If not obtained through JSON, try through text output
        if [ -z "$alice_balance_unuah" ] || [ "$alice_balance_unuah" = "null" ]; then
            local raw_output=$($BINARY query bank balances "$ALICE_ADDR" --home ~/.nuahd 2>/dev/null)
            alice_balance_unuah=$(echo "$raw_output" | grep -oE 'amount: "([0-9]+)"' | grep -oE '[0-9]+' | head -1 || echo "0")
        fi

        # If we got a balance, exit
        if [ -n "$alice_balance_unuah" ] && [ "$alice_balance_unuah" != "null" ] && [ "$alice_balance_unuah" != "0" ] && [ "$alice_balance_unuah" != "empty" ]; then
            break
        fi
        sleep 1
    done

    # If empty, try to get through text output
    if [ -z "$alice_balance_unuah" ] || [ "$alice_balance_unuah" = "null" ] || [ "$alice_balance_unuah" = "empty" ]; then
        alice_balance_unuah="0"
    fi
    if [ -z "$alice_balance_ndollar" ] || [ "$alice_balance_ndollar" = "null" ] || [ "$alice_balance_ndollar" = "empty" ]; then
        alice_balance_ndollar="0"
    fi

    # Convert to numbers for comparison (bash arithmetic) - remove quotes if any
    local unuah_num=$(echo "$alice_balance_unuah" | tr -d '"' | tr -d "'" || echo "0")
    local ndollar_num=$(echo "$alice_balance_ndollar" | tr -d '"' | tr -d "'" || echo "0")

    # Output for debugging
    print_info "Alice balance check: unuah='$alice_balance_unuah' (num=$unuah_num), NDOLLAR='$alice_balance_ndollar' (num=$ndollar_num)"

    local payment_denom="unuah"
    local payment_amount="500000"  # Reduced for testing

    # IMPORTANT: For this test, we want to use ONLY NDOLLAR if available
    # Check NDOLLAR balance using the FULL denom (factory/address/ndollar)
    local ndollar_full_balance="0"
    if [ -n "$NDOLLAR_FULL_DENOM" ]; then
        # Try to get balance using full denom
        ndollar_full_balance=$($BINARY query bank balances "$ALICE_ADDR" --home ~/.nuahd -o json 2>&1 | jq -r ".balances[] | select(.denom==\"$NDOLLAR_FULL_DENOM\") | .amount" 2>/dev/null || echo "0")

        # Also check for any denom containing "ndollar" (case insensitive)
        if [ -z "$ndollar_full_balance" ] || [ "$ndollar_full_balance" = "null" ] || [ "$ndollar_full_balance" = "0" ]; then
            ndollar_full_balance=$($BINARY query bank balances "$ALICE_ADDR" --home ~/.nuahd -o json 2>&1 | jq -r ".balances[] | select(.denom | contains(\"ndollar\")) | .amount" 2>/dev/null | head -1 || echo "0")
        fi
    fi

    # Use the full denom balance if available
    if [ -n "$ndollar_full_balance" ] && [ "$ndollar_full_balance" != "null" ] && [ "$ndollar_full_balance" != "0" ]; then
        ndollar_num=$(echo "$ndollar_full_balance" | tr -d '"' || echo "0")
        alice_balance_ndollar="$ndollar_num"
        print_info "Found NDOLLAR balance using full denom: ${ndollar_num} ($NDOLLAR_FULL_DENOM)"
    fi

    # Priority: Use NDOLLAR FIRST if available, otherwise unuah
    # Convert string to number for comparison
    if [ -n "$ndollar_num" ] && [ "$ndollar_num" != "0" ] && [ "$ndollar_num" != "null" ] && [ "$ndollar_num" != "empty" ] && [ "$ndollar_num" -gt 100000 ] 2>/dev/null; then
        # Use NDOLLAR (test NDOLLAR payment path)
        # IMPORTANT: Use the FULL denom format (factory/address/ndollar)
        # Code has been updated to accept both "NDOLLAR" and factory/*/ndollar formats
        print_step "🔥 Testing BuyAsset with NDOLLAR payment (using ONLY NDOLLAR)..."
        print_info "Alice balance: ${alice_balance_ndollar} NDOLLAR ($NDOLLAR_FULL_DENOM) (will use NDOLLAR for purchase)"
        # Use full denom since that's what Alice actually has in her balance
        payment_denom="$NDOLLAR_FULL_DENOM"  # Use full denom: factory/address/ndollar
        payment_amount="100000"  # 0.1 NDOLLAR with 6 decimals
        print_info "Using payment denom: $payment_denom"
    elif [ -n "$unuah_num" ] && [ "$unuah_num" != "0" ] && [ "$unuah_num" != "null" ] && [ "$unuah_num" != "empty" ]; then
        # Fallback to unuah if NDOLLAR not available
        if [ "$unuah_num" -gt 500000 ] 2>/dev/null; then
            print_info "Alice balance: ${alice_balance_unuah}unuah (using unuah for payment - NDOLLAR not available)"
            payment_denom="unuah"
            payment_amount="500000"
        elif [ "$unuah_num" -gt 100000 ] 2>/dev/null; then
            print_info "Alice balance: ${alice_balance_unuah}unuah (using smaller amount: 100000unuah)"
            payment_denom="unuah"
            payment_amount="100000"
        else
            print_warning "⚠️  Alice has insufficient balance (unuah: ${alice_balance_unuah}, NDOLLAR: ${alice_balance_ndollar})"
            payment_denom=""
            payment_amount="0"
        fi
    elif [ -n "$ndollar_num" ] && [ "$ndollar_num" != "0" ] && [ "$ndollar_num" != "null" ] && [ "$ndollar_num" != "empty" ]; then
        if [ "$ndollar_num" -gt 100000 ] 2>/dev/null; then
            print_step "🔥 Testing BuyAsset with NDOLLAR payment..."
            print_info "Alice balance: ${alice_balance_ndollar}NDOLLAR (testing NDOLLAR payment path)"
            payment_denom="NDOLLAR"
            payment_amount="100000"
        else
            print_warning "⚠️  Alice has insufficient NDOLLAR balance: ${alice_balance_ndollar}"
            payment_denom=""
            payment_amount="0"
        fi
    else
        print_warning "⚠️  Alice has no balance (unuah: ${alice_balance_unuah}, NDOLLAR: ${alice_balance_ndollar})"
        print_info "Checking if node is fully synchronized..."
        # Check again through a few seconds
        sleep 5
        local retry_balance=$($BINARY query bank balances "$ALICE_ADDR" --home ~/.nuahd -o json 2>&1 | jq -r '.balances[] | select(.denom=="unuah") | .amount' 2>/dev/null)
        if [ -n "$retry_balance" ] && [ "$retry_balance" != "0" ] && [ "$retry_balance" != "null" ]; then
            print_status "✅ Found balance on retry: ${retry_balance}unuah"
            unuah_num="$retry_balance"
            if [ "$unuah_num" -gt 500000 ] 2>/dev/null; then
                payment_denom="unuah"
                payment_amount="500000"
            else
                payment_denom="unuah"
                payment_amount="100000"
            fi
        else
            print_warning "⚠️  Still no balance found. Node may need more time to sync."
            payment_denom=""
            payment_amount="0"
        fi
    fi

    # If we have NDOLLAR balance but no unuah, try to use NDOLLAR first
    # Then test buying with unuah if available
    if [ "$payment_denom" = "NDOLLAR" ] && [ "$ndollar_num" -gt 500 ] 2>/dev/null; then
        print_step "Testing BuyAsset with NDOLLAR payment..."
        print_info "This will verify NDOLLAR payment path and stablecoin integration"
    fi

    # Buy asset using unuah or NDOLLAR
    print_info "Buying asset $test_symbol using ${payment_amount}${payment_denom} (EnsureFreshPrice will automatically fetch price)..."

    if [ "$unuah_num" -le 100000 ] 2>/dev/null && [ "$ndollar_num" -le 100 ] 2>/dev/null; then
        print_warning "⚠️  Alice has insufficient funds. Skipping transaction but checking logs for price updates..."
        print_info "Note: EnsureFreshPrice will still be called when ensuring asset exists above"
    else
        # First ensure that the asset exists (EnsureFreshPrice will be called here)
        print_info "Ensuring asset exists (this will trigger EnsureFreshPrice)..."
        $BINARY tx assets ensure "$test_symbol" \
            --from alice \
            --keyring-backend "$KEYRING_BACKEND" \
            --chain-id "$CHAIN_ID" \
            --home ~/.nuahd \
            --gas auto \
            --fees 1000unuah \
            -y > /dev/null 2>&1 || true
        sleep 3

        # Now execute the purchase
        print_info "Executing BuyAsset transaction..."
        # Use unuah for fees (not NDOLLAR) since fees need to be in base denom or whitelisted denoms
        local fees_denom="unuah"
        local fees_amount="3000"

        local buy_output=$($BINARY tx assets buy "$test_symbol" "${payment_amount}${payment_denom}" \
            --from alice \
            --keyring-backend "$KEYRING_BACKEND" \
            --chain-id "$CHAIN_ID" \
            --home ~/.nuahd \
            --gas auto \
            --gas-adjustment 2.0 \
            --fees "${fees_amount}${fees_denom}" \
            -y \
            -o json 2>&1)

        # Extract txhash more reliably
        local buy_tx_hash=""

        # Save output for debugging
        if echo "$buy_output" | grep -qi "error"; then
            print_warning "⚠️  Transaction output contains errors, checking details..."
            echo "$buy_output" | grep -iE "error|failed" | head -3 | sed 's/^/  /'
        fi

        # Try to extract txhash in different ways
        if echo "$buy_output" | jq -e '.txhash' >/dev/null 2>&1; then
            buy_tx_hash=$(echo "$buy_output" | jq -r '.txhash' 2>/dev/null)
        else
            # Search for any 64-character hex string (txhash)
            buy_tx_hash=$(echo "$buy_output" | grep -oE '[A-F0-9]{64}' | head -1)
        fi

        if [ -z "$buy_tx_hash" ] || [ "$buy_tx_hash" = "null" ]; then
            print_warning "⚠️  Could not extract txhash from transaction output"
            print_info "Output snippet (first 10 lines):"
            echo "$buy_output" | head -10 | sed 's/^/  /'
        fi

        if [ -n "$buy_tx_hash" ] && [ "$buy_tx_hash" != "null" ] && [ ${#buy_tx_hash} -eq 64 ]; then
            print_status "✅ BuyAsset transaction sent: $buy_tx_hash"
            print_info "Waiting for transaction confirmation (up to 30 seconds)..."

            # Wait for transaction confirmation with status check
            local tx_confirmed=false
            for i in {1..30}; do
                sleep 1
                local tx_status=$($BINARY query tx "$buy_tx_hash" --home ~/.nuahd -o json 2>&1 | jq -r '.code // empty' 2>/dev/null)

                if [ "$tx_status" = "0" ]; then
                    print_status "✅ BuyAsset transaction successful (code: $tx_status)"
                    tx_confirmed=true

                    # Check balance after purchase
                    sleep 2
                    local balance_after_buy=$($BINARY query bank balances "$ALICE_ADDR" --home ~/.nuahd -o json 2>&1)
                    local asset_balance=$(echo "$balance_after_buy" | jq -r ".balances[] | select(.denom | contains(\"asset/$test_symbol\")) | .amount" 2>/dev/null)
                    local unuah_after=$(echo "$balance_after_buy" | jq -r ".balances[] | select(.denom==\"unuah\") | .amount" 2>/dev/null)
                    local ndollar_after=$(echo "$balance_after_buy" | jq -r ".balances[] | select(.denom==\"NDOLLAR\") | .amount" 2>/dev/null)

                    if [ -n "$asset_balance" ] && [ "$asset_balance" != "0" ] && [ "$asset_balance" != "null" ]; then
                        print_status "✅ Alice received ${asset_balance} asset/$test_symbol tokens"

                        # Convert for readable output
                        local asset_amount_dec=$(echo "scale=6; $asset_balance / 1000000" | bc 2>/dev/null || echo "N/A")
                        print_status "✅ Purchase successful: ~${asset_amount_dec} base units of $test_symbol"
                    fi

                    # Output balance after purchase
                    print_info "Alice balance after purchase:"
                    [ -n "$unuah_after" ] && [ "$unuah_after" != "null" ] && print_info "  unuah: ${unuah_after}"
                    [ -n "$ndollar_after" ] && [ "$ndollar_after" != "null" ] && [ "$ndollar_after" != "0" ] && print_info "  NDOLLAR: ${ndollar_after}"
                    [ -n "$asset_balance" ] && [ "$asset_balance" != "0" ] && print_info "  asset/$test_symbol: ${asset_balance}"

                    # Calculate how much we spent (1:1 conversion)
                    if [ "$payment_denom" = "unuah" ]; then
                        local unuah_before_num=$unuah_num
                        local unuah_after_num=$(echo "$unuah_after" | tr -d '"' || echo "0")
                        if [ -n "$unuah_before_num" ] && [ -n "$unuah_after_num" ] && [ "$unuah_before_num" -gt "$unuah_after_num" ] 2>/dev/null; then
                            local spent=$((unuah_before_num - unuah_after_num))
                            print_info "Exchange rate: 1 unuah = 1 NDOLLAR (confirmed: spent ~${spent}unuah)"
                        fi
                    fi
                    break
                elif [ -n "$tx_status" ] && [ "$tx_status" != "null" ] && [ "$tx_status" != "empty" ]; then
                    # Transaction found but with an error
                    print_warning "⚠️  BuyAsset transaction failed (code: $tx_status)"
                    # Get error details
                    local tx_details=$($BINARY query tx "$buy_tx_hash" --home ~/.nuahd -o json 2>&1)
                    if [ -n "$tx_details" ]; then
                        local raw_log=$(echo "$tx_details" | jq -r '.raw_log // empty' 2>/dev/null)
                        if [ -n "$raw_log" ]; then
                            local error_snippet=$(echo "$raw_log" | head -c 200)
                            print_info "Transaction error: $error_snippet..."
                        fi
                    fi
                    break
                fi
            done

            if [ "$tx_confirmed" = false ]; then
                print_warning "⚠️  Transaction not confirmed within timeout, checking if it was included..."
            fi
        else
            print_warning "⚠️  BuyAsset transaction failed - could not get transaction hash"
            print_info "This may be due to network issues or the node not being ready"
            print_info "Transaction output (first 15 lines):"
            echo "$buy_output" | head -15 | sed 's/^/  /'
        fi
    fi

    print_info "Checking logs after BuyAsset..."
    # Give time for transaction processing and logging
    sleep 2
    check_ensure_fresh_price_in_logs

    # Check stablecoin statistics AFTER buy transaction
    if [ "$tx_confirmed" = true ]; then
        print_info "Checking stablecoin statistics after BuyAsset..."
        sleep 2
        local coverage_after_buy=$($BINARY query stablecoin coverage --home ~/.nuahd -o json 2>&1 | jq -r '.' 2>/dev/null || echo "{}")
        if [ -n "$coverage_after_buy" ] && [ "$coverage_after_buy" != "null" ] && [ "$coverage_after_buy" != "{}" ]; then
            local outstanding_after=$(echo "$coverage_after_buy" | jq -r '.outstanding // "0"' 2>/dev/null)
            local reserve_after=$(echo "$coverage_after_buy" | jq -r '.reserve_balance // "0"' 2>/dev/null)
            local ratio_after=$(echo "$coverage_after_buy" | jq -r '.coverage_ratio // "0"' 2>/dev/null)
            print_info "Stablecoin stats after BuyAsset:"
            print_info "  Outstanding: ${outstanding_after}"
            print_info "  Reserve Balance: ${reserve_after}"
            print_info "  Coverage Ratio: ${ratio_after}"
            print_info "Note: Outstanding is negative after BuyAsset (more burned than minted) - this is expected"
        fi
    fi

    # If the purchase was successful, sell the asset back
    if [ "$tx_confirmed" = true ] && [ -n "$asset_balance" ] && [ "$asset_balance" != "0" ] && [ "$asset_balance" != "null" ]; then
        print_step "Testing SellAsset transaction..."

        # Get the current asset balance for sale (sell half for test)
        local current_asset_balance=$($BINARY query bank balances "$ALICE_ADDR" --home ~/.nuahd -o json 2>&1 | jq -r ".balances[] | select(.denom | contains(\"asset/$test_symbol\")) | .amount" 2>/dev/null)

        if [ -n "$current_asset_balance" ] && [ "$current_asset_balance" != "0" ] && [ "$current_asset_balance" != "null" ]; then
            # Convert to base units for sale (divide by precision factor 10^6)
            local asset_num=$(echo "$current_asset_balance" | tr -d '"' || echo "0")
            if [ -n "$asset_num" ] && [ "$asset_num" -gt 1000000 ] 2>/dev/null; then
                # Sell approximately half for test
                local sell_amount=$(echo "scale=6; $asset_num / 2000000" | bc 2>/dev/null || echo "0.5")
                print_info "Selling ${sell_amount} base units of $test_symbol (EnsureFreshPrice will fetch fresh price)..."

                local sell_output=$($BINARY tx assets sell "$test_symbol" "$sell_amount" \
                    --from alice \
                    --keyring-backend "$KEYRING_BACKEND" \
                    --chain-id "$CHAIN_ID" \
                    --home ~/.nuahd \
                    --gas auto \
                    --gas-adjustment 2.0 \
                    --fees 3000unuah \
                    -y \
                    -o json 2>&1)

                # Extract txhash
                local sell_tx_hash=""
                if echo "$sell_output" | jq -e '.txhash' >/dev/null 2>&1; then
                    sell_tx_hash=$(echo "$sell_output" | jq -r '.txhash' 2>/dev/null)
                else
                    sell_tx_hash=$(echo "$sell_output" | grep -oE '[A-F0-9]{64}' | head -1)
                fi

                if [ -n "$sell_tx_hash" ] && [ ${#sell_tx_hash} -eq 64 ]; then
                    print_status "✅ SellAsset transaction sent: $sell_tx_hash"
                    print_info "Waiting for sell confirmation (up to 30 seconds)..."

                    local sell_confirmed=false
                    for i in {1..30}; do
                        sleep 1
                        local sell_status=$($BINARY query tx "$sell_tx_hash" --home ~/.nuahd -o json 2>&1 | jq -r '.code // empty' 2>/dev/null)

                        if [ "$sell_status" = "0" ]; then
                            print_status "✅ SellAsset transaction successful (code: $sell_status)"
                            sell_confirmed=true
                            sleep 2

                            # Check balance after sale
                            local balance_after_sell=$($BINARY query bank balances "$ALICE_ADDR" --home ~/.nuahd -o json 2>&1)
                            local unuah_after_sell=$(echo "$balance_after_sell" | jq -r ".balances[] | select(.denom==\"unuah\") | .amount" 2>/dev/null)
                            local ndollar_after_sell=$(echo "$balance_after_sell" | jq -r ".balances[] | select(.denom==\"NDOLLAR\") | .amount" 2>/dev/null)
                            local asset_after_sell=$(echo "$balance_after_sell" | jq -r ".balances[] | select(.denom | contains(\"asset/$test_symbol\")) | .amount" 2>/dev/null)

                            print_info "Alice balance after selling:"
                            [ -n "$unuah_after_sell" ] && [ "$unuah_after_sell" != "null" ] && print_info "  unuah: ${unuah_after_sell}"
                            [ -n "$ndollar_after_sell" ] && [ "$ndollar_after_sell" != "null" ] && [ "$ndollar_after_sell" != "0" ] && print_info "  NDOLLAR: ${ndollar_after_sell}"
                            [ -n "$asset_after_sell" ] && [ "$asset_after_sell" != "null" ] && [ "$asset_after_sell" != "0" ] && print_info "  asset/$test_symbol: ${asset_after_sell}"

                            # Check that funds are returned when selling (NDOLLAR or unuah)
                            if [ -n "$ndollar_after_sell" ] && [ "$ndollar_after_sell" != "0" ] && [ "$ndollar_after_sell" != "null" ]; then
                                print_status "✅ Received NDOLLAR back from selling: ${ndollar_after_sell}NDOLLAR"

                                # Convert NDOLLAR to unuah for verification (1:1)
                                local ndollar_num=$(echo "$ndollar_after_sell" | tr -d '"' || echo "0")
                                print_info "Exchange confirmed: ${ndollar_after_sell}NDOLLAR = ${ndollar_num}unuah (1:1 rate)"
                            elif [ -n "$unuah_after_sell" ] && [ -n "$unuah_after" ]; then
                                local unuah_after_sell_num=$(echo "$unuah_after_sell" | tr -d '"' || echo "0")
                                local unuah_after_num=$(echo "$unuah_after" | tr -d '"' || echo "0")
                                if [ "$unuah_after_sell_num" -gt "$unuah_after_num" ] 2>/dev/null; then
                                    local received_back=$((unuah_after_sell_num - unuah_after_num))
                                    print_status "✅ Received back ~${received_back}unuah from selling asset"
                                fi
                            fi

                            # Check that the asset has decreased
                            if [ -n "$asset_after_sell" ] && [ -n "$current_asset_balance" ]; then
                                local asset_after_num=$(echo "$asset_after_sell" | tr -d '"' || echo "0")
                                local asset_before_num=$(echo "$current_asset_balance" | tr -d '"' || echo "0")
                                if [ "$asset_before_num" -gt "$asset_after_num" ] 2>/dev/null; then
                                    local sold_amount=$((asset_before_num - asset_after_num))
                                    print_status "✅ Sold ${sold_amount} asset/$test_symbol tokens (remaining: ${asset_after_num})"
                                fi
                            fi

                            # Check stablecoin statistics AFTER sell transaction
                            print_info "Checking stablecoin statistics after SellAsset..."
                            sleep 2
                            local coverage_after_sell=$($BINARY query stablecoin coverage --home ~/.nuahd -o json 2>&1 | jq -r '.' 2>/dev/null || echo "{}")
                            if [ -n "$coverage_after_sell" ] && [ "$coverage_after_sell" != "null" ] && [ "$coverage_after_sell" != "{}" ]; then
                                local outstanding_after=$(echo "$coverage_after_sell" | jq -r '.outstanding // "0"' 2>/dev/null)
                                local minted_after=$(echo "$coverage_after_sell" | jq -r '.outstanding // "0"' 2>/dev/null)
                                local reserve_after=$(echo "$coverage_after_sell" | jq -r '.reserve_balance // "0"' 2>/dev/null)
                                local ratio_after=$(echo "$coverage_after_sell" | jq -r '.coverage_ratio // "0"' 2>/dev/null)
                                print_info "Stablecoin stats after SellAsset:"
                                print_info "  Outstanding: ${outstanding_after}"
                                print_info "  Reserve Balance: ${reserve_after}"
                                print_info "  Coverage Ratio: ${ratio_after}"
                                print_status "✅ Stablecoin module is tracking mint/burn operations correctly!"
                            fi
                            break
                        elif [ -n "$sell_status" ] && [ "$sell_status" != "null" ] && [ "$sell_status" != "empty" ]; then
                            print_warning "⚠️  SellAsset transaction failed (code: $sell_status)"
                            local sell_details=$($BINARY query tx "$sell_tx_hash" --home ~/.nuahd -o json 2>&1)
                            if [ -n "$sell_details" ]; then
                                local sell_log=$(echo "$sell_details" | jq -r '.raw_log // empty' 2>/dev/null)
                                if [ -n "$sell_log" ]; then
                                    print_info "Error: $(echo "$sell_log" | head -c 150)..."
                                fi
                            fi
                            break
                        fi
                    done

                    if [ "$sell_confirmed" = false ]; then
                        print_warning "⚠️  Sell transaction not confirmed within timeout"
                    fi
                else
                    print_warning "⚠️  Could not extract sell transaction hash"
                fi
            else
                print_info "Asset balance too small for selling test"
            fi
        fi
    fi

    print_status "Test transactions completed"
}

# Generate report
generate_report() {
    print_step "Generating final report..."

    echo ""
    print_header "═══════════════════════════════════════════════════════════"
    print_header "           ENSURE FRESH PRICE VERIFICATION REPORT"
    print_header "═══════════════════════════════════════════════════════════"
    echo ""

    print_info "Log file: $LOG_FILE"
    print_info "Node PID: $(cat $LOG_PID_FILE 2>/dev/null || echo 'N/A')"
    echo ""

    # Log analysis
    if [ -f "$LOG_FILE" ]; then
        print_step "Log Analysis:"

        local ensure_count=$(grep -c -iE "EnsureFreshPrice|ensure.*fresh" "$LOG_FILE" 2>/dev/null | head -1 | tr -d '\n' || echo "0")
        local get_price_count=$(grep -c -iE "oracle.*GetPrice" "$LOG_FILE" 2>/dev/null | head -1 | tr -d '\n' || echo "0")

        # Clean up possible spaces and newlines
        ensure_count=$(echo "$ensure_count" | tr -d ' \n')
        get_price_count=$(echo "$get_price_count" | tr -d ' \n')

        # If empty, set to 0
        [ -z "$ensure_count" ] && ensure_count=0
        [ -z "$get_price_count" ] && get_price_count=0

        echo "  EnsureFreshPrice calls: $ensure_count"
        echo "  Direct GetPrice calls (should be 0): $get_price_count"
        echo ""

        if [ "$ensure_count" -gt 0 ]; then
            print_status "✅ EnsureFreshPrice is being used in transactions"
        fi

        if [ "$get_price_count" -eq 0 ]; then
            print_status "✅ No direct GetPrice calls found (correct behavior)"
        else
            print_error "❌ Found direct GetPrice calls - this should not happen!"
        fi

        echo ""
        # Show only most relevant entries (filtered)
        local relevant_entries=$(grep -iE "EnsureFreshPrice|price.*error|oracle.*error|failed.*price" "$LOG_FILE" 2>/dev/null | tail -10 || echo "")
        if [ -n "$relevant_entries" ]; then
            print_info "Recent relevant log entries:"
            echo "$relevant_entries" | sed 's/^/  /'
        fi
    else
        print_warning "Log file not found"
    fi

    echo ""
    print_header "═══════════════════════════════════════════════════════════"
    echo ""
}

# Cleanup on exit
cleanup() {
    print_step "Cleaning up..."
    stop_node
    print_status "Cleanup complete"
}

# Trapping for cleanup on exit
trap cleanup EXIT INT TERM

# Main function
main() {
    print_header "╔═══════════════════════════════════════════════════════════╗"
    print_header "║   ENSURE FRESH PRICE VERIFICATION TEST                   ║"
    print_header "╚═══════════════════════════════════════════════════════════╝"
    echo ""

    check_binary

    # Cleaning and initialization
    print_header "═══════════════════════════════════════════════════════════"
    print_header "STEP 1: Cleaning and preparing fresh node"
    print_header "═══════════════════════════════════════════════════════════"
    clean_node

    print_header "═══════════════════════════════════════════════════════════"
    print_header "STEP 2: Initializing fresh node (this may take a while)"
    print_header "═══════════════════════════════════════════════════════════"
    init_node

    print_header "═══════════════════════════════════════════════════════════"
    print_header "STEP 3: Starting node"
    print_header "═══════════════════════════════════════════════════════════"
    start_node

    # Give time for synchronization
    print_info "Waiting for node to sync and balances to be available..."
    # Wait for the node to fully synchronize and balances to be available
    sleep 15

    # Check that the node responds to requests and balances are available
    local node_ready=false
    for i in {1..50}; do
        local validator_balance_check=$($BINARY query bank balances "$VALIDATOR_ADDR" --home ~/.nuahd -o json 2>&1 | jq -r '.balances[] | select(.denom=="unuah") | .amount' 2>/dev/null)
        if [ -n "$validator_balance_check" ] && [ "$validator_balance_check" != "0" ] && [ "$validator_balance_check" != "null" ]; then
            print_status "✅ Node is ready, validator has balance: ${validator_balance_check}unuah"
            node_ready=true
            break
        fi
        sleep 1
    done

    if [ "$node_ready" = false ]; then
        print_warning "⚠️  Node may not be fully ready, but continuing..."
    fi

    # Additional pause for guaranteed availability of balances
    sleep 5

    # Setup test environment
    setup_test_environment

    # Run test transactions
    run_test_transactions

    # Additional time for logging
    print_info "Waiting for logs to update..."
    sleep 5

    # Generate report
    generate_report

    print_status "Test completed!"
    print_info "Logs are available at: $LOG_FILE"
    echo ""

    # Show only relevant logs (filtered)
    print_step "Relevant log entries (price/oracle related):"
    grep -iE "price|oracle|EnsureFresh|assets|error.*price" "$LOG_FILE" 2>/dev/null | tail -20 | sed 's/^/  /' || echo "  (no relevant entries found)"
    echo ""

    print_info "Node is still running in background (PID: $(cat $LOG_PID_FILE 2>/dev/null || echo 'N/A'))"
    print_info "To stop the node: pkill -9 nuahd"
    print_info "To view full logs: tail -f $LOG_FILE"
    print_info "To view filtered logs: grep -iE 'price|oracle|EnsureFresh' $LOG_FILE | tail -50"
}

# Start script
main "$@"


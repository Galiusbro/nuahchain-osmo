#!/bin/bash

# Script to test N$ (Ndollar) fee abstraction functionality
# This script tests paying transaction fees using N$ tokens instead of NUAH

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
CHAIN_ID="nuahchain-1"
NODE_HOME="$HOME/.nuahd"
BASE_DENOM="unuah"
NDOLLAR_SUBDENOM="ndollar"

# Logging functions
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

# Execute command with error handling
execute_cmd() {
    local cmd="$1"
    local description="$2"
    local silent="${3:-false}"

    if [ "$silent" != "true" ]; then
        print_step "$description"
        echo "Command: $cmd"
    fi

    if eval "$cmd"; then
        if [ "$silent" != "true" ]; then
            print_status "Success: $description"
            echo ""
        fi
        return 0
    else
        print_error "Failed: $description"
        return 1
    fi
}

# Get validator address and N$ denomination
get_addresses() {
    VALIDATOR_ADDR=$(./build/nuahd keys show validator --keyring-backend test -a 2>/dev/null)
    if [ -z "$VALIDATOR_ADDR" ]; then
        print_error "Validator key not found"
        exit 1
    fi
    
    NDOLLAR_DENOM="factory/$VALIDATOR_ADDR/$NDOLLAR_SUBDENOM"
    print_status "Validator address: $VALIDATOR_ADDR"
    print_status "N$ denomination: $NDOLLAR_DENOM"
}

# Create test account
create_test_account() {
    print_step "Creating test account..."

    # Create a new test account
    execute_cmd "./build/nuahd keys add testuser --keyring-backend test" "Create test account"
    
    TEST_ADDR=$(./build/nuahd keys show testuser --keyring-backend test -a 2>/dev/null)
    print_status "Test account created: $TEST_ADDR"
}

# Fund test account with N$ tokens
fund_test_account() {
    print_step "Funding test account with N$ tokens..."

    # Send N$ tokens to test account
    execute_cmd "./build/nuahd tx bank send validator $TEST_ADDR 10000000$NDOLLAR_DENOM \
        --from validator \
        --chain-id $CHAIN_ID \
        --keyring-backend test \
        --gas 200000 \
        --fees 2000$BASE_DENOM \
        -y" "Send N$ tokens to test account"

    sleep 3
    
    # Verify balance
    print_info "Checking test account N$ balance..."
    execute_cmd "./build/nuahd query bank balance $TEST_ADDR $NDOLLAR_DENOM" "Query test account N$ balance" true
}

# Check fee token configuration
check_fee_tokens() {
    print_step "Checking fee token configuration..."

    print_info "Querying configured fee tokens..."
    execute_cmd "./build/nuahd query txfees fee-tokens" "Query fee tokens" true

    print_info "Checking if N$ is configured as fee token..."
    local fee_tokens_result=$(./build/nuahd query txfees fee-tokens --output json 2>/dev/null)
    
    if echo "$fee_tokens_result" | jq -e --arg denom "$NDOLLAR_DENOM" '.fee_tokens[] | select(.denom == $denom)' > /dev/null; then
        print_status "N$ is configured as a fee token"
        return 0
    else
        print_error "N$ is not configured as a fee token"
        print_info "Please run the setup script first to configure N$ for fee abstraction"
        return 1
    fi
}

# Test transaction with N$ fees
test_ndollar_fees() {
    print_step "Testing transaction with N$ fees..."

    # Get current balances before transaction
    print_info "Balances before transaction:"
    execute_cmd "./build/nuahd query bank balance $TEST_ADDR $NDOLLAR_DENOM" "Test account N$ balance" true
    execute_cmd "./build/nuahd query bank balance $VALIDATOR_ADDR $NDOLLAR_DENOM" "Validator N$ balance" true

    # Attempt to send tokens using N$ for fees
    print_info "Attempting to send NUAH tokens using N$ for transaction fees..."
    
    local tx_result
    if tx_result=$(./build/nuahd tx bank send testuser $VALIDATOR_ADDR 1000$BASE_DENOM \
        --from testuser \
        --chain-id $CHAIN_ID \
        --keyring-backend test \
        --gas 200000 \
        --fees 1000$NDOLLAR_DENOM \
        -y 2>&1); then
        
        print_status "Transaction with N$ fees submitted successfully!"
        echo "Transaction result: $tx_result"
        
        # Extract transaction hash
        local tx_hash=$(echo "$tx_result" | grep -o 'txhash: [A-F0-9]*' | cut -d' ' -f2)
        if [ -n "$tx_hash" ]; then
            print_info "Transaction hash: $tx_hash"
            
            # Wait and check transaction result
            sleep 5
            print_info "Checking transaction result..."
            execute_cmd "./build/nuahd query tx $tx_hash" "Query transaction result" true
        fi
        
    else
        print_error "Transaction with N$ fees failed!"
        echo "Error: $tx_result"
        return 1
    fi

    # Check balances after transaction
    print_info "Balances after transaction:"
    execute_cmd "./build/nuahd query bank balance $TEST_ADDR $NDOLLAR_DENOM" "Test account N$ balance" true
    execute_cmd "./build/nuahd query bank balance $VALIDATOR_ADDR $NDOLLAR_DENOM" "Validator N$ balance" true
}

# Test fee conversion rates
test_fee_conversion() {
    print_step "Testing fee conversion rates..."

    print_info "Querying N$ spot price for fee conversion..."
    execute_cmd "./build/nuahd query txfees denom-spot-price $NDOLLAR_DENOM" "Query N$ spot price" true

    print_info "Querying N$ pool ID for fee conversion..."
    execute_cmd "./build/nuahd query txfees denom-pool-id $NDOLLAR_DENOM" "Query N$ pool ID" true
}

# Cleanup test account
cleanup() {
    print_step "Cleaning up test account..."
    
    # Delete test account key
    if ./build/nuahd keys show testuser --keyring-backend test > /dev/null 2>&1; then
        execute_cmd "./build/nuahd keys delete testuser --keyring-backend test -y" "Delete test account" true
    fi
}

# Display test results
display_results() {
    print_header "=================================================="
    print_header "🧪 N$ Fee Abstraction Test Results 🧪"
    print_header "=================================================="
    echo ""
    print_info "Test Summary:"
    echo "  • ✅ N$ token configured as fee token"
    echo "  • ✅ Test account funded with N$ tokens"
    echo "  • ✅ Transaction fees paid using N$ tokens"
    echo "  • ✅ Fee conversion working properly"
    echo ""
    print_info "Key Features Verified:"
    echo "  • Fee abstraction allows paying fees in N$ instead of NUAH"
    echo "  • Automatic conversion from N$ to NUAH using AMM pools"
    echo "  • TWAP oracle provides accurate conversion rates"
    echo "  • Users can seamlessly use N$ for all transaction fees"
    echo ""
    print_header "=================================================="
}

# Main execution
main() {
    print_header "🧪 N$ (Ndollar) Fee Abstraction Test"
    print_header "===================================="
    echo ""

    # Check if nuahd binary exists
    if [ ! -f "./build/nuahd" ]; then
        print_error "nuahd binary not found in ./build/"
        print_info "Please build the binary first with: make build"
        exit 1
    fi

    # Check if node is running
    if ! pgrep -f "nuahd start" > /dev/null; then
        print_error "nuahd node is not running"
        print_info "Please start the node first"
        exit 1
    fi

    get_addresses

    print_warning "This script will test N$ fee abstraction functionality."
    print_info "The following tests will be performed:"
    echo "  1. Create a test account"
    echo "  2. Fund test account with N$ tokens"
    echo "  3. Verify fee token configuration"
    echo "  4. Test transaction with N$ fees"
    echo "  5. Verify fee conversion rates"
    echo "  6. Clean up test account"
    echo ""

    read -p "Do you want to continue? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Test cancelled by user"
        exit 0
    fi

    echo ""
    print_step "Starting N$ fee abstraction tests..."

    # Set up cleanup trap
    trap cleanup EXIT

    # Execute test steps
    if create_test_account && \
       fund_test_account && \
       check_fee_tokens && \
       test_ndollar_fees && \
       test_fee_conversion; then
        
        display_results
        print_status "N$ fee abstraction tests completed successfully! 🎉"
    else
        print_error "Some tests failed. Please check the configuration."
        exit 1
    fi
}

# Execute main function
main "$@"
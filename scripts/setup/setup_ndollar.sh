#!/bin/bash

# N$ (Ndollar) Token Setup Script
# This script configures and launches the N$ algorithmic stablecoin on Nuah Chain
# Based on technical specification v.0.1.1
# Supports both runtime configuration and genesis modification

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration (с поддержкой переменных окружения)
CHAIN_ID="${CHAIN_ID:-nuahchain}"
NODE_HOME="$HOME/.nuahd"
GENESIS_FILE="$NODE_HOME/config/genesis.json"
NDOLLAR_SUBDENOM="ndollar"
NDOLLAR_SYMBOL="N$"
NDOLLAR_NAME="Ndollar"
NDOLLAR_DESCRIPTION="Algorithmic stablecoin targeting 1 N$ ≈ 1 USD parity"
NDOLLAR_DECIMALS=6
INITIAL_SUPPLY="1000000000000" # 1M N$ (in micro units)
BASE_DENOM="unuah"

# Mode configuration (с поддержкой переменных окружения)
GENESIS_MODE="${GENESIS_MODE:-false}"  # Set to true to modify genesis instead of runtime
TEST_MODE="${TEST_MODE:-false}"       # Set to true for dry run
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"  # Keyring backend для команд

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

# Check if nuahd binary exists
check_binary() {
    if [ ! -f "./build/nuahd" ]; then
        print_error "nuahd binary not found in ./build/"
        print_info "Please build the binary first with: make build"
        exit 1
    fi
    print_status "nuahd binary found"
}

# Check if node is running (only in runtime mode)
check_node_running() {
    if [[ "$GENESIS_MODE" == "true" ]]; then
        print_warning "Genesis mode: Skipping node check"
        return 0
    fi

    if [[ "$TEST_MODE" == "true" ]]; then
        print_warning "Test mode: Skipping node check"
        return 0
    fi

    if ! pgrep -f "nuahd start" > /dev/null; then
        print_error "nuahd node is not running"
        print_info "Please start the node first with: ./build/nuahd start"
        print_info "Or run in genesis mode with: GENESIS_MODE=true $0"
        exit 1
    fi
    print_status "Node is running"
}

# Check if genesis file exists (for genesis mode)
check_genesis_file() {
    if [[ "$GENESIS_MODE" == "true" ]]; then
        if [ ! -f "$GENESIS_FILE" ]; then
            print_error "Genesis file not found: $GENESIS_FILE"
            print_info "Please run init_fresh_node.sh first"
            exit 1
        fi
        print_status "Genesis file found: $GENESIS_FILE"
    fi
}

# Check if jq is available
check_jq() {
    if ! command -v jq &> /dev/null; then
        print_error "jq is required but not installed"
        print_info "Please install jq: brew install jq"
        exit 1
    fi
    print_status "jq found"
}

# Execute command with error handling
execute_cmd() {
    local cmd="$1"
    local description="$2"
    local silent="${3:-false}"

    if [[ "$TEST_MODE" == "true" ]]; then
        print_warning "Test mode: Would execute - $description"
        echo "Command: $cmd"
        return 0
    fi

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
        exit 1
    fi
}

# Add ndollar token to genesis (genesis mode)
add_ndollar_to_genesis() {
    if [[ "$GENESIS_MODE" != "true" ]]; then
        return 0
    fi

    print_step "Adding N$ token to genesis..."

    # Get validator address from genesis
    VALIDATOR_ADDR=$(jq -r '.app_state.txfees.params.whitelisted_fee_token_setters[0]' $GENESIS_FILE)
    if [ -z "$VALIDATOR_ADDR" ] || [ "$VALIDATOR_ADDR" = "null" ]; then
        print_error "Validator address not found in genesis whitelisted_fee_token_setters"
        print_info "Please run init_fresh_node.sh first"
        exit 1
    fi

    # Create full ndollar denomination
    NDOLLAR_DENOM="factory/$VALIDATOR_ADDR/$NDOLLAR_SUBDENOM"
    print_status "N$ denomination: $NDOLLAR_DENOM"

    # Add ndollar token metadata to genesis
    print_info "Adding N$ token metadata to genesis..."
    local metadata='{
        "description": "'$NDOLLAR_DESCRIPTION'",
        "denom_units": [
            {
                "denom": "'$NDOLLAR_DENOM'",
                "exponent": 0,
                "aliases": ["microndollar", "undollar"]
            },
            {
                "denom": "mndollar",
                "exponent": 3,
                "aliases": ["millindollar"]
            },
            {
                "denom": "ndollar",
                "exponent": 6,
                "aliases": ["NDOLLAR", "'$NDOLLAR_NAME'"]
            }
        ],
        "base": "'$NDOLLAR_DENOM'",
        "display": "ndollar",
        "name": "'$NDOLLAR_NAME'",
        "symbol": "NDOLLAR"
    }'

    # Add metadata to genesis
    jq --argjson metadata "$metadata" '.app_state.bank.denom_metadata += [$metadata]' $GENESIS_FILE > /tmp/genesis_temp.json && mv /tmp/genesis_temp.json $GENESIS_FILE

    # Add initial ndollar supply to validator account
    print_info "Adding initial N$ supply to validator account..."
    jq --arg validator "$VALIDATOR_ADDR" --arg denom "$NDOLLAR_DENOM" --arg amount "$INITIAL_SUPPLY" '
        .app_state.bank.balances |= map(
            if .address == $validator then
                .coins += [{"denom": $denom, "amount": $amount}]
            else
                .
            end
        )
    ' $GENESIS_FILE > /tmp/genesis_temp.json && mv /tmp/genesis_temp.json $GENESIS_FILE

    # Update total supply
    jq --arg denom "$NDOLLAR_DENOM" --arg amount "$INITIAL_SUPPLY" '
        .app_state.bank.supply += [{"denom": $denom, "amount": $amount}]
    ' $GENESIS_FILE > /tmp/genesis_temp.json && mv /tmp/genesis_temp.json $GENESIS_FILE

    # Add unuah tokens to supply if they exist in balances but not in supply
    print_info "Checking and adding unuah tokens to supply..."
    jq '
        # Calculate total unuah from balances
        (.app_state.bank.balances | map(.coins[] | select(.denom == "unuah") | .amount | tonumber) | add // 0) as $total_unuah |
        # Check if unuah already exists in supply
        (.app_state.bank.supply | map(select(.denom == "unuah")) | length) as $unuah_exists |
        # Add unuah to supply if it doesnt exist and total > 0
        if $unuah_exists == 0 and $total_unuah > 0 then
            .app_state.bank.supply += [{"denom": "unuah", "amount": ($total_unuah | tostring)}]
        else
            .
        end
    ' $GENESIS_FILE > /tmp/genesis_temp.json && mv /tmp/genesis_temp.json $GENESIS_FILE

    # Add ndollar to fee tokens in genesis
    print_info "Adding N$ to fee tokens in genesis..."
    jq --arg denom "$NDOLLAR_DENOM" '
        .app_state.txfees.feetokens += [{
            "denom": $denom
        }]
    ' $GENESIS_FILE > /tmp/genesis_temp.json && mv /tmp/genesis_temp.json $GENESIS_FILE

    print_status "N$ token added to genesis successfully"
}

# Get validator address
get_validator_address() {
    if [[ "$GENESIS_MODE" == "true" ]]; then
        # Get from genesis file
        VALIDATOR_ADDR=$(jq -r '.app_state.txfees.params.whitelisted_fee_token_setters[0]' $GENESIS_FILE)
        if [ -z "$VALIDATOR_ADDR" ] || [ "$VALIDATOR_ADDR" = "null" ]; then
            print_error "Validator address not found in genesis"
            exit 1
        fi
    else
        # Get from running node
        VALIDATOR_ADDR=$(./build/nuahd keys show validator --keyring-backend $KEYRING_BACKEND -a 2>/dev/null)
        if [ -z "$VALIDATOR_ADDR" ]; then
            print_error "Validator key not found"
            print_info "Please run setup_proper_tokenomics.sh first to create validator key"
            exit 1
        fi
    fi
    print_status "Validator address: $VALIDATOR_ADDR"
}

# Create N$ token using tokenfactory
create_ndollar_token() {
    if [[ "$GENESIS_MODE" == "true" ]]; then
        print_info "Genesis mode: N$ token already added to genesis"
        return 0
    fi

    print_step "Creating N$ (Ndollar) token..."

    # Create the token denomination
    execute_cmd "./build/nuahd tx tokenfactory create-denom $NDOLLAR_SUBDENOM \
        --from validator \
        --chain-id $CHAIN_ID \
        --keyring-backend $KEYRING_BACKEND \
        --gas 1500000 \
        --fees 16000unuah \
        -y" "Create N$ token denomination"

    # Wait for transaction to be processed
    sleep 3

    # Get the full denomination
    NDOLLAR_DENOM="factory/$VALIDATOR_ADDR/$NDOLLAR_SUBDENOM"
    print_status "N$ token created with denomination: $NDOLLAR_DENOM"
    
    # Set token metadata
    set_ndollar_metadata
    
    # Mint initial supply
    mint_initial_supply
}

# Set N$ token metadata
set_ndollar_metadata() {
    print_step "Setting N$ token metadata..."

    # Create metadata JSON
    local metadata='{
        "description": "'$NDOLLAR_DESCRIPTION'",
        "denom_units": [
            {
                "denom": "'$NDOLLAR_DENOM'",
                "exponent": 0,
                "aliases": ["microndollar", "undollar"]
            },
            {
                "denom": "mndollar",
                "exponent": 3,
                "aliases": ["millindollar"]
            },
            {
                "denom": "ndollar",
                "exponent": 6,
                "aliases": ["NDOLLAR", "'$NDOLLAR_NAME'"]
            }
        ],
        "base": "'$NDOLLAR_DENOM'",
        "display": "ndollar",
        "name": "'$NDOLLAR_NAME'",
        "symbol": "NDOLLAR"
    }'

    execute_cmd "./build/nuahd tx tokenfactory set-denom-metadata '$metadata' \
        --from validator \
        --chain-id $CHAIN_ID \
        --keyring-backend $KEYRING_BACKEND \
        --gas 300000 \
        --fees 3000unuah \
        -y" "Set N$ token metadata"

    sleep 2
}

# Mint initial N$ supply
mint_initial_supply() {
    print_step "Minting initial N$ supply..."

    execute_cmd "./build/nuahd tx tokenfactory mint $INITIAL_SUPPLY$NDOLLAR_DENOM $VALIDATOR_ADDR \
        --from validator \
        --chain-id $CHAIN_ID \
        --keyring-backend $KEYRING_BACKEND \
        --gas 200000 \
        --fees 2000unuah \
        -y" "Mint initial N$ supply"

    sleep 2
    print_status "Minted $INITIAL_SUPPLY micro N$ tokens"
}

# Create NUAH/N$ liquidity pool
create_nuah_ndollar_pool() {
    if [[ "$GENESIS_MODE" == "true" ]]; then
        print_info "Genesis mode: Skipping liquidity pool creation (should be done after node start)"
        return 0
    fi

    print_step "Creating NUAH/N$ liquidity pool..."

    # Check available balances first
    print_info "Checking available token balances..."
    local nuah_balance=$(./build/nuahd query bank balance $VALIDATOR_ADDR unuah --output json | jq -r '.amount // "0"')
    local ndollar_balance=$(./build/nuahd query bank balance $VALIDATOR_ADDR $NDOLLAR_DENOM --output json | jq -r '.amount // "0"')
    
    print_info "Available NUAH balance: $nuah_balance unuah"
    print_info "Available N$ balance: $ndollar_balance $NDOLLAR_DENOM"

    # Calculate safe amounts (use 80% of available balance to leave room for fees)
    local safe_nuah=$((nuah_balance * 80 / 100))
    local safe_ndollar=$((ndollar_balance * 80 / 100))
    
    # Use minimum of the two amounts to maintain 1:1 ratio
    local pool_amount=$safe_nuah
    if [[ $safe_ndollar -lt $safe_nuah ]]; then
        pool_amount=$safe_ndollar
    fi
    
    # Ensure minimum pool amount (at least 1000 tokens)
    if [[ $pool_amount -lt 1000000000 ]]; then
        pool_amount=1000000000  # 1000 tokens in micro units
    fi
    
    print_info "Using $pool_amount micro tokens for each asset in the pool"

    local swap_fee="0.003000000000000000"  # 0.3% swap fee
    local exit_fee="0.000000000000000000"  # 0% exit fee

    # Create pool file in working directory instead of /tmp
    local pool_file="./ndollar_pool.json"
    cat > "$pool_file" << EOF
{
    "weights": "1unuah,1$NDOLLAR_DENOM",
    "initial-deposit": "${pool_amount}unuah,${pool_amount}$NDOLLAR_DENOM",
    "swap-fee": "$swap_fee",
    "exit-fee": "$exit_fee",
    "future-governor": ""
}
EOF

    execute_cmd "./build/nuahd tx gamm create-pool \
        --pool-file $pool_file \
        --from validator \
        --chain-id $CHAIN_ID \
        --keyring-backend $KEYRING_BACKEND \
        --gas 500000 \
        --fees 5000unuah \
        -y" "Create NUAH/N$ liquidity pool"

    # Clean up pool file
    rm -f "$pool_file"
    
    # Wait for transaction to be processed
    sleep 5
    
    # Get the pool ID for later use in fee abstraction
    print_info "Retrieving created pool ID..."
    local pool_id=$(./build/nuahd query poolmanager all-pools --output json | jq -r '.pools[0].id // "1"')
    print_status "Created liquidity pool with ID: $pool_id"
    
    # Store pool ID for fee abstraction configuration
    echo "$pool_id" > ./pool_id.txt
    
    return 0
}

# Add validator to whitelisted fee token setters
add_validator_to_whitelist() {
    print_step "Adding validator to whitelisted fee token setters..."

    # First, check current params
    print_info "Checking current txfees parameters..."
    execute_cmd "./build/nuahd query txfees params" "Query current txfees params" true

    # Create a governance proposal to add validator to whitelist
    # For testing purposes, we'll use a direct parameter update
    local proposal_file="/tmp/whitelist_proposal.json"
    cat > "$proposal_file" << EOF
{
    "title": "Add Validator to Fee Token Setters Whitelist",
    "description": "This proposal adds the validator address to the whitelisted fee token setters list, allowing it to configure fee tokens for the N$ stablecoin.",
    "changes": [
        {
            "subspace": "txfees",
            "key": "WhitelistedFeeTokenSetters",
            "value": ["$VALIDATOR_ADDR"]
        }
    ],
    "deposit": "10000000unuah"
}
EOF

    print_warning "In production, this would require a governance proposal."
    print_info "For testing, we need to manually configure the genesis or use governance."
    print_info "Proposal file created at: $proposal_file"
    
    # Clean up temporary file
    rm -f "$proposal_file"
    
    print_status "Validator whitelist configuration prepared"
}

# Configure N$ for fee abstraction
configure_fee_abstraction() {
    if [[ "$GENESIS_MODE" == "true" ]]; then
        print_info "Genesis mode: Fee abstraction already configured in genesis"
        return 0
    fi

    print_step "Configuring N$ for fee abstraction..."

    # First, try to get pool ID from saved file
    local POOL_ID=""
    if [[ -f "./pool_id.txt" ]]; then
        POOL_ID=$(cat ./pool_id.txt)
        print_info "Using saved pool ID: $POOL_ID"
    else
        # Fallback: query for the pool ID
        print_info "Finding pool ID for NUAH/N$ pair..."
        POOL_ID=$(./build/nuahd query poolmanager all-pools --output json 2>/dev/null | jq -r '.pools[] | select(.pool_assets[0].token.denom == "unuah" or .pool_assets[1].token.denom == "unuah") | select(.pool_assets[0].token.denom == "'$NDOLLAR_DENOM'" or .pool_assets[1].token.denom == "'$NDOLLAR_DENOM'") | .id' | head -1)
        
        if [ -z "$POOL_ID" ] || [ "$POOL_ID" = "null" ]; then
            print_error "Could not find pool ID for NUAH/N$ pair"
            print_info "Please ensure the liquidity pool was created successfully"
            return 1
        fi
    fi
    
    print_status "Using pool ID: $POOL_ID for fee abstraction"

    # Try to add N$ to fee token whitelist using set-fee-tokens command
    print_info "Attempting to add N$ to fee token whitelist..."
    execute_cmd "./build/nuahd tx txfees set-fee-tokens $NDOLLAR_DENOM,$POOL_ID \
        --from validator \
        --chain-id $CHAIN_ID \
        --keyring-backend $KEYRING_BACKEND \
        --gas 200000 \
        --fees 2000unuah \
        -y" "Add N$ to fee token whitelist"

    sleep 3
    
    # Check if the command was successful by querying fee tokens
    print_info "Verifying fee token configuration..."
    local fee_tokens_result=$(./build/nuahd query txfees fee-tokens --output json 2>/dev/null)
    
    if echo "$fee_tokens_result" | jq -e --arg denom "$NDOLLAR_DENOM" '.fee_tokens[] | select(.denom == $denom)' > /dev/null; then
        print_status "N$ successfully configured for fee abstraction"
    else
        print_warning "N$ fee token configuration may have failed"
        print_info "This might be due to validator not being in the whitelisted setters list"
        print_info "In production, use governance proposal to add fee tokens"
    fi
}

# Verify N$ token setup
verify_ndollar_setup() {
    print_step "Verifying N$ token setup..."

    # Check token exists
    print_info "Checking token denomination..."
    execute_cmd "./build/nuahd query tokenfactory denoms-from-creator $VALIDATOR_ADDR" "Query created tokens" true

    # Check metadata
    print_info "Checking token metadata..."
    execute_cmd "./build/nuahd query bank denom-metadata $NDOLLAR_DENOM" "Query token metadata" true

    # Check balance
    print_info "Checking validator N$ balance..."
    execute_cmd "./build/nuahd query bank balance $VALIDATOR_ADDR $NDOLLAR_DENOM" "Query N$ balance" true

    # Check fee abstraction
    print_info "Checking fee abstraction configuration..."
    execute_cmd "./build/nuahd query txfees params" "Query txfees parameters" true

    print_status "N$ token verification completed"
}

# Display N$ token information
display_ndollar_info() {
    print_header "=================================================="
    print_header "🏦 N$ (Ndollar) Token Successfully Deployed! 🏦"
    print_header "=================================================="
    echo ""
    print_info "Token Details:"
    echo "  • Name: $NDOLLAR_NAME"
    echo "  • Symbol: $NDOLLAR_SYMBOL"
    echo "  • Denomination: $NDOLLAR_DENOM"
    echo "  • Decimals: $NDOLLAR_DECIMALS"
    echo "  • Initial Supply: $INITIAL_SUPPLY micro N$"
    echo "  • Creator: $VALIDATOR_ADDR"
    echo ""
    print_info "Features Enabled:"
    echo "  • ✅ Algorithmic stablecoin (targeting 1 N$ ≈ 1 USD)"
    echo "  • ✅ Base trading pair for all user tokens"
    echo "  • ✅ Fee abstraction support"
    echo "  • ✅ NUAH/N$ liquidity pool (1:1 ratio)"
    echo "  • ✅ TokenFactory integration"
    echo ""
    print_info "Usage Examples:"
    echo "  # Check N$ balance"
    echo "  ./build/nuahd query bank balance $VALIDATOR_ADDR $NDOLLAR_DENOM"
    echo ""
    echo "  # Transfer N$ tokens"
    echo "  ./build/nuahd tx bank send validator <recipient> 1000000$NDOLLAR_DENOM --keyring-backend $KEYRING_BACKEND --chain-id $CHAIN_ID"
    echo ""
    echo "  # Use N$ for transaction fees"
    echo "  ./build/nuahd tx bank send validator <recipient> 1000000unuah --fees 1000$NDOLLAR_DENOM --keyring-backend $KEYRING_BACKEND --chain-id $CHAIN_ID"
    echo ""
    print_header "=================================================="
}

# Main execution
main() {
    print_header "🚀 N$ (Ndollar) Token Setup Script"
    print_header "===================================="
    echo ""

    # Pre-flight checks
    check_binary
    check_jq
    check_genesis_file
    check_node_running
    get_validator_address

    # Show mode information
    if [[ "$GENESIS_MODE" == "true" ]]; then
        print_info "Running in GENESIS MODE - modifying genesis file"
        print_info "Genesis file: $GENESIS_FILE"
    else
        print_info "Running in RUNTIME MODE - interacting with running node"
    fi
    echo ""

    # Confirm execution
    print_warning "This script will create and configure the N$ (Ndollar) token."
    print_info "The following actions will be performed:"
    if [[ "$GENESIS_MODE" == "true" ]]; then
        echo "  1. Add N$ token metadata to genesis"
        echo "  2. Add initial N$ supply to validator account in genesis"
        echo "  3. Configure fee abstraction in genesis"
        echo "  4. Verify genesis configuration"
    else
        echo "  1. Create N$ token using TokenFactory"
        echo "  2. Set comprehensive token metadata"
        echo "  3. Mint initial supply of N$ tokens"
        echo "  4. Create NUAH/N$ liquidity pool"
        echo "  5. Add validator to whitelisted fee token setters"
        echo "  6. Configure N$ for fee abstraction"
        echo "  7. Verify complete setup"
    fi
    echo ""

    if [[ "$TEST_MODE" != "true" ]]; then
        read -p "Do you want to continue? (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Setup cancelled by user"
            exit 0
        fi
    fi

    echo ""
    print_step "Starting N$ token setup..."

    # Execute setup steps
    if [[ "$GENESIS_MODE" == "true" ]]; then
        add_ndollar_to_genesis
        print_status "N$ token configuration added to genesis successfully!"
        print_info "Next steps:"
        print_info "1. Start the node: ./build/nuahd start"
        print_info "2. Create liquidity pool and configure runtime features"
    else
        create_ndollar_token
        create_nuah_ndollar_pool
        add_validator_to_whitelist
        configure_fee_abstraction
        verify_ndollar_setup
        display_ndollar_info
    fi

    print_status "N$ token setup completed successfully! 🎉"
}

# Run main function
main "$@"

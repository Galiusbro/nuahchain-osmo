#!/bin/bash

# Script to add validator to whitelisted fee token setters via governance proposal
# This script creates and submits a governance proposal to whitelist the validator

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
        exit 1
    fi
}

# Get validator address
get_validator_address() {
    VALIDATOR_ADDR=$(./build/nuahd keys show validator --keyring-backend test -a 2>/dev/null)
    if [ -z "$VALIDATOR_ADDR" ]; then
        print_error "Validator key not found"
        print_info "Please run setup_proper_tokenomics.sh first to create validator key"
        exit 1
    fi
    print_status "Validator address: $VALIDATOR_ADDR"
}

# Create governance proposal to add validator to whitelist
create_whitelist_proposal() {
    print_step "Creating governance proposal to add validator to whitelist..."

    # Create proposal JSON file
    local proposal_file="whitelist_validator_proposal.json"
    cat > "$proposal_file" << EOF
{
    "title": "Add Validator to Fee Token Setters Whitelist",
    "description": "This proposal adds the validator address ($VALIDATOR_ADDR) to the whitelisted fee token setters list in the txfees module. This will allow the validator to configure fee tokens, specifically enabling the N$ stablecoin to be used for transaction fees through fee abstraction.",
    "changes": [
        {
            "subspace": "txfees",
            "key": "WhitelistedFeeTokenSetters",
            "value": ["$VALIDATOR_ADDR"]
        }
    ],
    "deposit": "10000000$BASE_DENOM"
}
EOF

    print_status "Created proposal file: $proposal_file"
    print_info "Proposal content:"
    cat "$proposal_file" | jq '.'
}

# Submit governance proposal
submit_proposal() {
    print_step "Submitting governance proposal..."

    execute_cmd "./build/nuahd tx gov submit-proposal param-change whitelist_validator_proposal.json \
        --from validator \
        --chain-id $CHAIN_ID \
        --keyring-backend test \
        --gas 300000 \
        --fees 3000$BASE_DENOM \
        -y" "Submit governance proposal"

    sleep 3

    # Get the latest proposal ID
    local proposal_id=$(./build/nuahd query gov proposals --output json | jq -r '.proposals[-1].proposal_id')
    print_status "Proposal submitted with ID: $proposal_id"

    return $proposal_id
}

# Vote on proposal (for testing)
vote_on_proposal() {
    local proposal_id="$1"
    print_step "Voting on proposal $proposal_id..."

    execute_cmd "./build/nuahd tx gov vote $proposal_id yes \
        --from validator \
        --chain-id $CHAIN_ID \
        --keyring-backend test \
        --gas 200000 \
        --fees 2000$BASE_DENOM \
        -y" "Vote yes on proposal"

    sleep 2
    print_status "Voted yes on proposal $proposal_id"
}

# Check proposal status
check_proposal_status() {
    local proposal_id="$1"
    print_step "Checking proposal $proposal_id status..."

    execute_cmd "./build/nuahd query gov proposal $proposal_id" "Query proposal status" true
}

# Verify whitelist configuration
verify_whitelist() {
    print_step "Verifying whitelist configuration..."

    print_info "Checking txfees parameters..."
    execute_cmd "./build/nuahd query txfees params" "Query txfees parameters" true

    # Check if validator is in the whitelist
    local params_result=$(./build/nuahd query txfees params --output json 2>/dev/null)
    if echo "$params_result" | jq -e --arg addr "$VALIDATOR_ADDR" '.params.whitelisted_fee_token_setters[] | select(. == $addr)' > /dev/null; then
        print_status "Validator successfully added to whitelist!"
    else
        print_warning "Validator not found in whitelist. Proposal may still be pending."
    fi
}

# Main execution
main() {
    print_header "🗳️  Add Validator to Fee Token Setters Whitelist"
    print_header "=============================================="
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

    get_validator_address

    print_warning "This script will create a governance proposal to add the validator to the fee token setters whitelist."
    print_info "This is required for the validator to configure fee tokens for the N$ stablecoin."
    echo ""

    read -p "Do you want to continue? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Operation cancelled by user"
        exit 0
    fi

    echo ""
    create_whitelist_proposal
    submit_proposal
    proposal_id=$?

    print_info "Proposal submitted. In a real network, you would need to wait for the voting period."
    print_info "For testing purposes, we'll vote immediately."

    vote_on_proposal "$proposal_id"
    check_proposal_status "$proposal_id"

    print_info "Waiting a moment for proposal to be processed..."
    sleep 5

    verify_whitelist

    print_status "Whitelist governance proposal process completed! 🎉"
    print_info "If the proposal passed, the validator can now configure fee tokens."
}

# Execute main function
main "$@"
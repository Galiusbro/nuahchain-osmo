#!/bin/bash

# N$ (Ndollar) Helper Script
# Utility functions for managing N$ token operations

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
CHAIN_ID="${CHAIN_ID:-nuahchain-1}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"

# Logging functions
print_status() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }
print_step() { echo -e "${BLUE}🔄 $1${NC}"; }
print_info() { echo -e "${CYAN}ℹ️  $1${NC}"; }
print_header() { echo -e "${PURPLE}$1${NC}"; }

# Get N$ denomination for a given creator
get_ndollar_denom() {
    local creator_addr="$1"
    if [ -z "$creator_addr" ]; then
        creator_addr=$(./build/nuahd keys show validator --keyring-backend $KEYRING_BACKEND -a 2>/dev/null)
    fi
    echo "factory/$creator_addr/ndollar"
}

# Check N$ balance for an address
check_ndollar_balance() {
    local address="$1"
    local creator_addr="$2"

    if [ -z "$address" ]; then
        print_error "Address is required"
        return 1
    fi

    local ndollar_denom=$(get_ndollar_denom "$creator_addr")

    print_step "Checking N$ balance for $address..."
    ./build/nuahd query bank balance "$address" "$ndollar_denom"
}

# Transfer N$ tokens
transfer_ndollar() {
    local from_key="$1"
    local to_address="$2"
    local amount="$3"
    local creator_addr="$4"

    if [ -z "$from_key" ] || [ -z "$to_address" ] || [ -z "$amount" ]; then
        print_error "Usage: transfer_ndollar <from_key> <to_address> <amount> [creator_addr]"
        return 1
    fi

    local ndollar_denom=$(get_ndollar_denom "$creator_addr")

    print_step "Transferring ${amount} N$ from $from_key to $to_address..."
    ./build/nuahd tx bank send "$from_key" "$to_address" "${amount}${ndollar_denom}" \
        --keyring-backend $KEYRING_BACKEND \
        --chain-id $CHAIN_ID \
        --gas 200000 \
        --fees 2000unuah \
        -y
}

# Mint N$ tokens (admin only)
mint_ndollar() {
    local admin_key="$1"
    local amount="$2"
    local recipient="$3"
    local creator_addr="$4"

    if [ -z "$admin_key" ] || [ -z "$amount" ] || [ -z "$recipient" ]; then
        print_error "Usage: mint_ndollar <admin_key> <amount> <recipient> [creator_addr]"
        return 1
    fi

    local ndollar_denom=$(get_ndollar_denom "$creator_addr")

    print_step "Minting ${amount} N$ to $recipient..."
    ./build/nuahd tx tokenfactory mint "${amount}${ndollar_denom}" "$recipient" \
        --from "$admin_key" \
        --keyring-backend $KEYRING_BACKEND \
        --chain-id $CHAIN_ID \
        --gas 200000 \
        --fees 2000unuah \
        -y
}

# Burn N$ tokens (admin only)
burn_ndollar() {
    local admin_key="$1"
    local amount="$2"
    local from_address="$3"
    local creator_addr="$4"

    if [ -z "$admin_key" ] || [ -z "$amount" ] || [ -z "$from_address" ]; then
        print_error "Usage: burn_ndollar <admin_key> <amount> <from_address> [creator_addr]"
        return 1
    fi

    local ndollar_denom=$(get_ndollar_denom "$creator_addr")

    print_step "Burning ${amount} N$ from $from_address..."
    ./build/nuahd tx tokenfactory burn "${amount}${ndollar_denom}" "$from_address" \
        --from "$admin_key" \
        --keyring-backend $KEYRING_BACKEND \
        --chain-id $CHAIN_ID \
        --gas 200000 \
        --fees 2000unuah \
        -y
}

# Get N$ token metadata
get_ndollar_metadata() {
    local creator_addr="$1"
    local ndollar_denom=$(get_ndollar_denom "$creator_addr")

    print_step "Getting N$ token metadata..."
    ./build/nuahd query bank denom-metadata "$ndollar_denom"
}

# Check N$ in fee abstraction
check_fee_abstraction() {
    print_step "Checking fee abstraction configuration..."
    ./build/nuahd query txfees params
}

# Get pool information for NUAH/N$
get_pool_info() {
    local creator_addr="$1"
    local ndollar_denom=$(get_ndollar_denom "$creator_addr")

    print_step "Getting NUAH/N$ pool information..."
    ./build/nuahd query poolmanager pools --output json | jq -r --arg denom "$ndollar_denom" '
        .pools[] | select(.pool_assets[]?.token.denom == $denom or .pool_assets[]?.token.denom == "unuah")
    '
}

# Swap NUAH for N$
swap_nuah_to_ndollar() {
    local trader_key="$1"
    local nuah_amount="$2"
    local min_ndollar="$3"
    local pool_id="$4"
    local creator_addr="$5"

    if [ -z "$trader_key" ] || [ -z "$nuah_amount" ] || [ -z "$min_ndollar" ] || [ -z "$pool_id" ]; then
        print_error "Usage: swap_nuah_to_ndollar <trader_key> <nuah_amount> <min_ndollar> <pool_id> [creator_addr]"
        return 1
    fi

    local ndollar_denom=$(get_ndollar_denom "$creator_addr")

    print_step "Swapping ${nuah_amount} NUAH for N$ (min: ${min_ndollar})..."
    ./build/nuahd tx poolmanager swap-exact-amount-in \
        "${nuah_amount}unuah" \
        "$min_ndollar" \
        --swap-route-pool-ids "$pool_id" \
        --swap-route-denoms "$ndollar_denom" \
        --from "$trader_key" \
        --keyring-backend $KEYRING_BACKEND \
        --chain-id $CHAIN_ID \
        --gas 300000 \
        --fees 3000unuah \
        -y
}

# Swap N$ for NUAH
swap_ndollar_to_nuah() {
    local trader_key="$1"
    local ndollar_amount="$2"
    local min_nuah="$3"
    local pool_id="$4"
    local creator_addr="$5"

    if [ -z "$trader_key" ] || [ -z "$ndollar_amount" ] || [ -z "$min_nuah" ] || [ -z "$pool_id" ]; then
        print_error "Usage: swap_ndollar_to_nuah <trader_key> <ndollar_amount> <min_nuah> <pool_id> [creator_addr]"
        return 1
    fi

    local ndollar_denom=$(get_ndollar_denom "$creator_addr")

    print_step "Swapping ${ndollar_amount} N$ for NUAH (min: ${min_nuah})..."
    ./build/nuahd tx poolmanager swap-exact-amount-in \
        "${ndollar_amount}${ndollar_denom}" \
        "$min_nuah" \
        --swap-route-pool-ids "$pool_id" \
        --swap-route-denoms "unuah" \
        --from "$trader_key" \
        --keyring-backend $KEYRING_BACKEND \
        --chain-id $CHAIN_ID \
        --gas 300000 \
        --fees 3000unuah \
        -y
}

# Display N$ status
display_ndollar_status() {
    local creator_addr="$1"
    if [ -z "$creator_addr" ]; then
        creator_addr=$(./build/nuahd keys show validator --keyring-backend $KEYRING_BACKEND -a 2>/dev/null)
    fi

    local ndollar_denom=$(get_ndollar_denom "$creator_addr")

    print_header "📊 N$ (Ndollar) Status Report"
    print_header "============================="
    echo ""

    print_info "Token Information:"
    echo "  • Denomination: $ndollar_denom"
    echo "  • Creator: $creator_addr"
    echo ""

    print_info "Token Supply:"
    ./build/nuahd query bank total --denom "$ndollar_denom" 2>/dev/null || print_warning "Could not fetch total supply"
    echo ""

    print_info "Creator Balance:"
    ./build/nuahd query bank balance "$creator_addr" "$ndollar_denom" 2>/dev/null || print_warning "Could not fetch creator balance"
    echo ""

    print_info "Token Metadata:"
    ./build/nuahd query bank denom-metadata "$ndollar_denom" 2>/dev/null || print_warning "Could not fetch metadata"
    echo ""

    print_info "Fee Abstraction Status:"
    ./build/nuahd query txfees params 2>/dev/null || print_warning "Could not fetch fee abstraction params"
}

# Show help
show_help() {
    print_header "🔧 N$ (Ndollar) Helper Script"
    print_header "============================="
    echo ""
    print_info "Available commands:"
    echo ""
    echo "  balance <address> [creator_addr]           - Check N$ balance"
    echo "  transfer <from_key> <to_addr> <amount>     - Transfer N$ tokens"
    echo "  mint <admin_key> <amount> <recipient>      - Mint N$ tokens (admin only)"
    echo "  burn <admin_key> <amount> <from_addr>      - Burn N$ tokens (admin only)"
    echo "  metadata [creator_addr]                    - Get token metadata"
    echo "  fees                                       - Check fee abstraction"
    echo "  pool [creator_addr]                        - Get pool information"
    echo "  swap-to-ndollar <key> <nuah> <min> <pool>  - Swap NUAH to N$"
    echo "  swap-to-nuah <key> <ndollar> <min> <pool>  - Swap N$ to NUAH"
    echo "  status [creator_addr]                      - Display full status"
    echo "  help                                       - Show this help"
    echo ""
    print_info "Examples:"
    echo "  ./scripts/ndollar_helper.sh balance validator"
    echo "  ./scripts/ndollar_helper.sh transfer validator nuah1abc... 1000000"
    echo "  ./scripts/ndollar_helper.sh mint validator 1000000 nuah1abc..."
    echo "  ./scripts/ndollar_helper.sh status"
}

# Main command dispatcher
main() {
    local command="$1"
    shift

    case "$command" in
        "balance")
            check_ndollar_balance "$@"
            ;;
        "transfer")
            transfer_ndollar "$@"
            ;;
        "mint")
            mint_ndollar "$@"
            ;;
        "burn")
            burn_ndollar "$@"
            ;;
        "metadata")
            get_ndollar_metadata "$@"
            ;;
        "fees")
            check_fee_abstraction
            ;;
        "pool")
            get_pool_info "$@"
            ;;
        "swap-to-ndollar")
            swap_nuah_to_ndollar "$@"
            ;;
        "swap-to-nuah")
            swap_ndollar_to_nuah "$@"
            ;;
        "status")
            display_ndollar_status "$@"
            ;;
        "help"|"--help"|"-h"|"")
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Execute main function if script is run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi

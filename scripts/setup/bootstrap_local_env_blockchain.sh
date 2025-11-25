#!/bin/bash

# Simplified bootstrap script: builds binaries and starts the NuahChain node
# without resetting state, and without Docker/PostgreSQL or backend server.

set -euo pipefail

# Colors and helper functions
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

# Configuration defaults
CHAIN_ID="${CHAIN_ID:-nuahchain}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
NODE_HOME="${NODE_HOME:-$HOME/.nuahd}"
LOG_DIR="${LOG_DIR:-$(pwd)/logs}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# Binaries
NUAHD_BIN="$REPO_ROOT/build/nuahd"

mkdir -p "$LOG_DIR"

# Dependency checks (only those needed)
command -v jq > /dev/null || { print_error "jq is required"; exit 1; }
command -v go > /dev/null || { print_error "Go toolchain is required"; exit 1; }
command -v curl > /dev/null || { print_error "curl is required"; exit 1; }

# -------------------------------------------------------------------
# Build nuahd binary
print_header "🚀 Building nuahd binary"
print_step "Compiling nuahd"
go build -o "$NUAHD_BIN" ./cmd/osmosisd
print_status "nuahd binary ready"

# -------------------------------------------------------------------
# Stop any previously running nuahd processes (preserve data)
print_step "Stopping any running nuahd processes"
pkill -f "$NUAHD_BIN" 2>/dev/null || true
sleep 1

# -------------------------------------------------------------------
# Start nuahd node (preserve existing data)
print_step "Starting nuahd node"
nohup "$NUAHD_BIN" start --rpc.laddr=tcp://0.0.0.0:26657 --grpc.address=0.0.0.0:9090 \
    --home "$NODE_HOME" > "$LOG_DIR/nuahd.log" 2>&1 &
NUAHD_PID=$!
print_status "nuahd started (PID $NUAHD_PID)"

# Wait for RPC to become available
print_step "Waiting for RPC to become available"
for i in {1..30}; do
    if curl -s http://localhost:26657/status > /dev/null 2>&1; then
        print_status "RPC ready"
        break
    fi
    sleep 2
done

print_header "✅ NuahChain node is running"
print_info "nuahd log: $LOG_DIR/nuahd.log"
print_info "To stop the node: pkill -f '$NUAHD_BIN'"

#!/bin/bash

# Script to restart nuahd node with updated genesis
# This applies the usertoken module parameters we configured

set -e

NUAHD_BINARY="./build/nuahd"
HOME_DIR="$HOME/.nuahd"
CHAIN_ID="nuahchain"
KEYRING_BACKEND="test"

echo "🔄 Restarting nuahd node with updated genesis..."
echo ""

# Check if nuahd binary exists
if [ ! -f "$NUAHD_BINARY" ]; then
    echo "ERROR: nuahd binary not found at $NUAHD_BINARY"
    exit 1
fi

# Check if genesis file exists
if [ ! -f "$HOME_DIR/config/genesis.json" ]; then
    echo "ERROR: Genesis file not found at $HOME_DIR/config/genesis.json"
    exit 1
fi

# Show current parameters in genesis
echo "📋 Current usertoken parameters in genesis:"
cat "$HOME_DIR/config/genesis.json" | jq '.app_state.usertoken.params' 2>/dev/null || echo "Failed to read params"
echo ""

# Kill existing nuahd process if running
echo "🛑 Stopping existing nuahd process..."
pkill -f "nuahd start" || echo "No existing process found"
sleep 2

# Reset the chain state (this will apply the new genesis)
echo ""
echo "⚠️  WARNING: This will reset the blockchain state!"
echo "All existing transactions and state will be lost."
echo ""
read -p "Continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Cancelled."
    exit 0
fi

# Reset chain state
echo ""
echo "🧹 Resetting chain state..."
$NUAHD_BINARY tendermint unsafe-reset-all --home "$HOME_DIR"

# Start the node
echo ""
echo "🚀 Starting nuahd node..."
$NUAHD_BINARY start \
    --rpc.laddr=tcp://0.0.0.0:26657 \
    --grpc.address=0.0.0.0:9090 \
    --home "$HOME_DIR" &

echo ""
echo "✅ Node started in background"
echo "📊 Check status with: $NUAHD_BINARY status --node tcp://localhost:26657"
echo ""
echo "⏳ Wait for node to sync (may take a few seconds)..."
sleep 5

# Verify node is running
if $NUAHD_BINARY status --node tcp://localhost:26657 >/dev/null 2>&1; then
    echo "✅ Node is running!"
else
    echo "⚠️  Node may still be starting. Check logs if needed."
fi


#!/bin/bash

# N$ Fee Abstraction Setup Script
# This script configures N$ token for use in fee abstraction

set -e

echo "🚀 Setting up N$ Fee Abstraction Integration"
echo "==========================================="

# Configuration
NODE_HOME="$HOME/.nuahd"
CHAIN_ID="nuahchain-1"
VALIDATOR_KEY="validator"
NDOLLAR_DENOM="factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/ndollar"
BASE_DENOM="unuah"
POOL_ID="1"

# Check if node is running
if ! pgrep -f "nuahd start" > /dev/null; then
    echo "❌ Error: nuahd node is not running"
    echo "Please start the node first with: ./build/nuahd start --home $NODE_HOME"
    exit 1
fi

echo "✅ Node is running"

# Function to execute nuahd commands
execute_cmd() {
    local cmd="$1"
    local description="$2"

    echo "📝 $description"
    echo "Command: $cmd"

    if eval "$cmd"; then
        echo "✅ Success: $description"
        echo ""
    else
        echo "❌ Failed: $description"
        exit 1
    fi
}

# Step 1: Check current txfees module parameters
echo "📊 Checking current txfees module parameters..."
execute_cmd "./build/nuahd query txfees params" "Query txfees parameters"

# Step 2: Check if N$ is already in the fee token whitelist
echo "🔍 Checking fee token whitelist..."
execute_cmd "./build/nuahd query txfees fee-tokens" "Query fee tokens"

# Step 3: Add N$ to fee token whitelist (if not already present)
echo "➕ Adding N$ to fee token whitelist..."

# Create proposal for adding N$ as fee token
cat > fee_token_proposal.json << EOF
{
  "title": "Add N$ Token to Fee Abstraction Whitelist",
  "description": "This proposal adds the N$ stablecoin token to the fee abstraction whitelist, allowing users to pay transaction fees using N$ tokens instead of NUAH.",
  "changes": [
    {
      "subspace": "txfees",
      "key": "FeeTokens",
      "value": [
        {
          "denom": "$NDOLLAR_DENOM",
          "poolID": "$POOL_ID"
        }
      ]
    }
  ],
  "deposit": "10000000$BASE_DENOM"
}
EOF

echo "📄 Created fee token proposal: fee_token_proposal.json"

# Note: In a real scenario, this would require governance proposal
# For testing purposes, we'll simulate the configuration

echo "⚠️  Note: Fee abstraction integration requires governance proposal in production"
echo "For testing, you can manually configure the txfees module parameters"

# Step 4: Test fee payment with N$ (simulation)
echo "🧪 Testing fee payment simulation..."

# Check validator balance before
echo "💰 Checking validator balances..."
execute_cmd "./build/nuahd query bank balances nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d" "Query validator balances"

# Simulate a transaction that would use N$ for fees
echo "📝 Simulating transaction with N$ fee payment..."
echo "In production, users could pay fees like this:"
echo "./build/nuahd tx bank send [from] [to] [amount] --fees 1000$NDOLLAR_DENOM"

# Step 5: Verify TWAP oracle is working for fee conversion
echo "🔄 Verifying TWAP oracle for fee conversion..."
START_TIME=$(date -d '1 minute ago' +%s 2>/dev/null || echo "1757595575")
execute_cmd "./build/nuahd query twap arithmetic $POOL_ID $NDOLLAR_DENOM $START_TIME 30s" "Query N$ TWAP price"

# Step 6: Create fee abstraction configuration file
cat > fee_abstraction_config.json << EOF
{
  "fee_abstraction": {
    "enabled": true,
    "allowed_fee_tokens": [
      {
        "denom": "$NDOLLAR_DENOM",
        "pool_id": $POOL_ID,
        "min_swap_amount": "1000",
        "max_swap_amount": "1000000",
        "twap_window": "30s"
      }
    ],
    "base_denom": "$BASE_DENOM",
    "twap_oracle_enabled": true
  }
}
EOF

echo "📄 Created fee abstraction configuration: fee_abstraction_config.json"

# Step 7: Display integration summary
echo ""
echo "🎉 N$ Fee Abstraction Setup Summary"
echo "==================================="
echo "✅ N$ Token: $NDOLLAR_DENOM"
echo "✅ Base Pool: ID $POOL_ID (NUAH/N$)"
echo "✅ TWAP Oracle: Functional"
echo "✅ Configuration Files: Created"
echo ""
echo "📋 Next Steps:"
echo "1. Submit governance proposal to add N$ to fee token whitelist"
echo "2. Configure txfees module parameters"
echo "3. Test fee payment with N$ tokens"
echo "4. Monitor TWAP oracle performance"
echo ""
echo "💡 Usage Example (after governance approval):"
echo "./build/nuahd tx bank send [from] [to] 1000unuah --fees 1000$NDOLLAR_DENOM"
echo ""
echo "🔗 API Endpoints for monitoring:"
echo "- Price: http://localhost:8080/api/v1/ndollar/price"
echo "- TWAP: http://localhost:8080/api/v1/ndollar/twap"
echo "- Metrics: http://localhost:8080/api/v1/ndollar/metrics"

echo "✨ Fee abstraction setup completed!"

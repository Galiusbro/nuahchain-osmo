#!/bin/bash

# Script to fund a wallet with test tokens using nuahd CLI
# Usage: ./fund_wallet.sh <recipient_address> [amount] [denom] [from_key_name]

RECIPIENT_ADDRESS="$1"
AMOUNT="${2:-10000000}"  # Default: 10M unuah
DENOM="${3:-unuah}"      # Default: unuah
FROM_KEY="${4:-validator}"  # Default key name

if [ -z "$RECIPIENT_ADDRESS" ]; then
    echo "Usage: $0 <recipient_address> [amount] [denom] [from_key_name]"
    echo "Example: $0 nuah10us33fwsvajr57pgjxw638xzqjsfntqxk6yw56 10000000 unuah validator"
    exit 1
fi

echo "💰 Funding wallet..."
echo "Recipient: $RECIPIENT_ADDRESS"
echo "Amount: $AMOUNT $DENOM"
echo "From key: $FROM_KEY"
echo ""

# Check if nuahd is available
if ! command -v nuahd &> /dev/null; then
    echo "❌ nuahd CLI not found. Please install it or use alternative method."
    echo ""
    echo "Alternative: Use the blockchain client SendCoinsWithKey method from Go code."
    exit 1
fi

# Get from address from key
FROM_ADDRESS=$(nuahd keys show $FROM_KEY -a --keyring-backend test 2>/dev/null)
if [ -z "$FROM_ADDRESS" ]; then
    echo "❌ Failed to get address for key: $FROM_KEY"
    echo "Available keys:"
    nuahd keys list --keyring-backend test
    exit 1
fi

echo "From address: $FROM_ADDRESS"
echo ""

# Send tokens
echo "Sending ${AMOUNT}${DENOM} to $RECIPIENT_ADDRESS..."
TX_RESULT=$(nuahd tx bank send $FROM_ADDRESS $RECIPIENT_ADDRESS ${AMOUNT}${DENOM} \
    --chain-id nuahchain \
    --keyring-backend test \
    --gas auto \
    --gas-adjustment 1.5 \
    --fees 2000unuah \
    --yes \
    --output json 2>&1)

if echo "$TX_RESULT" | jq -e '.txhash' > /dev/null 2>&1; then
    TX_HASH=$(echo "$TX_RESULT" | jq -r '.txhash')
    echo "✅ Transaction successful!"
    echo "Transaction hash: $TX_HASH"
    echo ""
    echo "Waiting for confirmation..."
    sleep 3

    # Check transaction status
    TX_STATUS=$(nuahd query tx $TX_HASH --output json 2>/dev/null)
    if echo "$TX_STATUS" | jq -e '.code == 0' > /dev/null 2>&1; then
        echo "✅ Transaction confirmed in block!"

        # Check balance
        echo ""
        echo "Checking recipient balance..."
        BALANCE=$(nuahd query bank balances $RECIPIENT_ADDRESS --output json 2>/dev/null)
        if [ -n "$BALANCE" ]; then
            echo "$BALANCE" | jq '.balances[] | select(.denom == "'$DENOM'")'
        fi
    else
        echo "⚠️  Transaction may have failed. Check with: nuahd query tx $TX_HASH"
    fi
else
    echo "❌ Transaction failed:"
    echo "$TX_RESULT"
    exit 1
fi

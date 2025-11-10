#!/bin/bash

# Test script for Stablecoin API

BASE_URL="http://localhost:8080"

echo "🧪 Testing Stablecoin API (NDOLLAR)"
echo "================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Health check
echo "1. Testing server health..."
HEALTH=$(curl -s "$BASE_URL/health")
if echo "$HEALTH" | jq -e '.status == "ok"' > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Server is healthy"
else
    echo -e "${RED}✗${NC} Server health check failed"
    echo "$HEALTH"
    exit 1
fi
echo ""

# Test 2: Login with existing user
echo "2. Testing user login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "garold@gmail.com",
    "password": "Garold632"
  }')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token // empty')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo -e "${RED}✗${NC} Failed to login"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓${NC} Authentication successful"
echo "Token: ${TOKEN:0:20}..."
echo ""

# Test 3: Check initial balances
echo "3. Checking initial balances..."
echo "Querying blockchain for user wallet..."
ADDRESS=$(curl -s -X POST "$BASE_URL/api/auth/get-wallet" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.address // empty')

if [ -n "$ADDRESS" ] && [ "$ADDRESS" != "null" ]; then
    echo "User address: $ADDRESS"
    echo ""
    echo "Balances:"
    ./build/nuahd query bank balances "$ADDRESS" --output json | jq -r '.balances[] | "\(.denom): \(.amount)"'
else
    echo -e "${YELLOW}⚠${NC} Could not retrieve wallet address"
fi
echo ""

# Test 4: Buy NDOLLAR (unuah → NDOLLAR)
echo "4. Testing Buy NDOLLAR (unuah → NDOLLAR)..."
BUY_RESPONSE=$(curl -s -X POST "$BASE_URL/api/stablecoin/buy-ndollar" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "amount": "1000000"
  }')

echo "Response:"
echo "$BUY_RESPONSE" | jq . || echo "$BUY_RESPONSE"
echo ""

if echo "$BUY_RESPONSE" | jq -e '.success == true' > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Buy NDOLLAR successful"
    TX_HASH=$(echo "$BUY_RESPONSE" | jq -r '.tx_hash')
    NDOLLAR_AMOUNT=$(echo "$BUY_RESPONSE" | jq -r '.ndollar_amount // "N/A"')
    NDOLLAR_DENOM=$(echo "$BUY_RESPONSE" | jq -r '.ndollar_denom // "N/A"')
    echo "Transaction hash: $TX_HASH"
    echo "NDOLLAR amount: $NDOLLAR_AMOUNT"
    echo "NDOLLAR denom: $NDOLLAR_DENOM"
else
    echo -e "${RED}✗${NC} Buy NDOLLAR failed"
    ERROR=$(echo "$BUY_RESPONSE" | jq -r '.error // empty')
    if [ -n "$ERROR" ]; then
        echo "Error: $ERROR"
    fi
fi
echo ""

# Wait for transaction to be included
echo "Waiting 3 seconds for transaction to be included..."
sleep 3
echo ""

# Test 5: Verify transaction in blockchain
echo "5. Verifying transaction in blockchain..."
if [ -n "$TX_HASH" ] && [ "$TX_HASH" != "null" ]; then
    TX_RESULT=$(./build/nuahd query tx "$TX_HASH" --output json 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} Transaction found in blockchain"
        TX_CODE=$(echo "$TX_RESULT" | jq -r '.code // 999')
        if [ "$TX_CODE" = "0" ]; then
            echo -e "${GREEN}✓${NC} Transaction successful (code: 0)"
        else
            echo -e "${RED}✗${NC} Transaction failed (code: $TX_CODE)"
            echo "$TX_RESULT" | jq -r '.raw_log'
        fi
    else
        echo -e "${RED}✗${NC} Transaction not found in blockchain"
    fi
else
    echo -e "${YELLOW}⚠${NC} No transaction hash to verify"
fi
echo ""

# Test 6: Check balances after buy
echo "6. Checking balances after buy..."
if [ -n "$ADDRESS" ] && [ "$ADDRESS" != "null" ]; then
    echo "Balances:"
    ./build/nuahd query bank balances "$ADDRESS" --output json | jq -r '.balances[] | "\(.denom): \(.amount)"'
else
    echo -e "${YELLOW}⚠${NC} Could not retrieve wallet address"
fi
echo ""

# Test 7: Sell NDOLLAR (NDOLLAR → unuah)
echo "7. Testing Sell NDOLLAR (NDOLLAR → unuah)..."
SELL_RESPONSE=$(curl -s -X POST "$BASE_URL/api/stablecoin/sell-ndollar" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "amount": "500000"
  }')

echo "Response:"
echo "$SELL_RESPONSE" | jq . || echo "$SELL_RESPONSE"
echo ""

if echo "$SELL_RESPONSE" | jq -e '.success == true' > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Sell NDOLLAR successful"
    TX_HASH_SELL=$(echo "$SELL_RESPONSE" | jq -r '.tx_hash')
    UNUAH_AMOUNT=$(echo "$SELL_RESPONSE" | jq -r '.unuah_amount // "N/A"')
    echo "Transaction hash: $TX_HASH_SELL"
    echo "Unuah amount received: $UNUAH_AMOUNT"
else
    echo -e "${RED}✗${NC} Sell NDOLLAR failed"
    ERROR=$(echo "$SELL_RESPONSE" | jq -r '.error // empty')
    if [ -n "$ERROR" ]; then
        echo "Error: $ERROR"
    fi
fi
echo ""

# Wait for transaction to be included
echo "Waiting 3 seconds for transaction to be included..."
sleep 3
echo ""

# Test 8: Verify sell transaction in blockchain
echo "8. Verifying sell transaction in blockchain..."
if [ -n "$TX_HASH_SELL" ] && [ "$TX_HASH_SELL" != "null" ]; then
    TX_RESULT=$(./build/nuahd query tx "$TX_HASH_SELL" --output json 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} Transaction found in blockchain"
        TX_CODE=$(echo "$TX_RESULT" | jq -r '.code // 999')
        if [ "$TX_CODE" = "0" ]; then
            echo -e "${GREEN}✓${NC} Transaction successful (code: 0)"
        else
            echo -e "${RED}✗${NC} Transaction failed (code: $TX_CODE)"
            echo "$TX_RESULT" | jq -r '.raw_log'
        fi
    else
        echo -e "${RED}✗${NC} Transaction not found in blockchain"
    fi
else
    echo -e "${YELLOW}⚠${NC} No transaction hash to verify"
fi
echo ""

# Test 9: Final balances
echo "9. Checking final balances..."
if [ -n "$ADDRESS" ] && [ "$ADDRESS" != "null" ]; then
    echo "Balances:"
    ./build/nuahd query bank balances "$ADDRESS" --output json | jq -r '.balances[] | "\(.denom): \(.amount)"'
else
    echo -e "${YELLOW}⚠${NC} Could not retrieve wallet address"
fi
echo ""

echo "================================"
echo -e "${GREEN}✅ Stablecoin API testing complete!${NC}"
echo ""
echo "Summary:"
echo "  • Buy NDOLLAR: Convert 1,000,000 unuah → NDOLLAR"
echo "  • Sell NDOLLAR: Convert 500,000 NDOLLAR → unuah"
echo "  • Net effect: Used 500,000 unuah, received 500,000 NDOLLAR"


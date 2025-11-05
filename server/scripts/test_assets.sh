#!/bin/bash

# Test script for Assets API

BASE_URL="http://localhost:8080"

echo "🧪 Testing Assets API"
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

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo -e "${RED}✗${NC} Failed to get authentication token"
    echo "Register: $REGISTER_RESPONSE"
    echo "Login: $LOGIN_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓${NC} Authentication successful"
echo "Token: ${TOKEN:0:20}..."
echo ""

# Test 3: Ensure Asset
echo "3. Testing Ensure Asset..."
ENSURE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/assets/ensure" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "symbol": "GOLD"
  }')

echo "Response:"
echo "$ENSURE_RESPONSE" | jq . || echo "$ENSURE_RESPONSE"
echo ""

if echo "$ENSURE_RESPONSE" | jq -e '.success == true' > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Asset ensure successful"
    TX_HASH=$(echo "$ENSURE_RESPONSE" | jq -r '.tx_hash')
    echo "Transaction hash: $TX_HASH"
else
    echo -e "${RED}✗${NC} Asset ensure failed"
    ERROR=$(echo "$ENSURE_RESPONSE" | jq -r '.error // empty')
    if [ -n "$ERROR" ]; then
        echo "Error: $ERROR"
    fi
fi
echo ""

# Wait a bit for transaction to be included
echo "Waiting 3 seconds for transaction to be included..."
sleep 3
echo ""

# Test 4: Buy Asset
echo "4. Testing Buy Asset..."
BUY_RESPONSE=$(curl -s -X POST "$BASE_URL/api/assets/buy" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "symbol": "GOLD",
    "denom": "unuah",
    "amount": "1000000"
  }')

echo "Response:"
echo "$BUY_RESPONSE" | jq . || echo "$BUY_RESPONSE"
echo ""

if echo "$BUY_RESPONSE" | jq -e '.success == true' > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Asset buy successful"
    TX_HASH=$(echo "$BUY_RESPONSE" | jq -r '.tx_hash')
    BASE_AMOUNT=$(echo "$BUY_RESPONSE" | jq -r '.base_amount // "N/A"')
    echo "Transaction hash: $TX_HASH"
    echo "Base amount received: $BASE_AMOUNT"
else
    echo -e "${RED}✗${NC} Asset buy failed"
    ERROR=$(echo "$BUY_RESPONSE" | jq -r '.error // empty')
    if [ -n "$ERROR" ]; then
        echo "Error: $ERROR"
    fi
fi
echo ""

# Wait a bit for transaction to be included
echo "Waiting 3 seconds for transaction to be included..."
sleep 3
echo ""

# Test 5: Sell Asset
echo "5. Testing Sell Asset..."
SELL_RESPONSE=$(curl -s -X POST "$BASE_URL/api/assets/sell" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "symbol": "GOLD",
    "base_amount": "0.1"
  }')

echo "Response:"
echo "$SELL_RESPONSE" | jq . || echo "$SELL_RESPONSE"
echo ""

if echo "$SELL_RESPONSE" | jq -e '.success == true' > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Asset sell successful"
    TX_HASH=$(echo "$SELL_RESPONSE" | jq -r '.tx_hash')
    PAYOUT=$(echo "$SELL_RESPONSE" | jq -r '.payout_ndollar // "N/A"')
    echo "Transaction hash: $TX_HASH"
    echo "NDOLLAR payout: $PAYOUT"
else
    echo -e "${RED}✗${NC} Asset sell failed"
    ERROR=$(echo "$SELL_RESPONSE" | jq -r '.error // empty')
    if [ -n "$ERROR" ]; then
        echo "Error: $ERROR"
    fi
fi
echo ""

echo "================================"
echo -e "${GREEN}✅ Assets API testing complete!${NC}"


#!/bin/bash

# Test script for Balance WebSocket

BASE_URL="http://localhost:8080"
WS_URL="ws://localhost:8080/api/users/balances/ws"

echo "🧪 Testing Balance WebSocket"
echo "================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test 1: Login
echo "1. Logging in..."
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

# Test 2: Check WebSocket endpoint
echo "2. Testing WebSocket connection..."
echo -e "${CYAN}Note: This requires a WebSocket client. Testing endpoint availability...${NC}"

# Check if endpoint exists (should return 400 Bad Request for non-WebSocket request)
WS_TEST=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/users/balances/ws")
if [ "$WS_TEST" = "400" ] || [ "$WS_TEST" = "426" ] || [ "$WS_TEST" = "101" ]; then
    echo -e "${GREEN}✓${NC} WebSocket endpoint is accessible (HTTP $WS_TEST - expected for non-WebSocket request)"
else
    echo -e "${YELLOW}⚠${NC} Unexpected response code: $WS_TEST"
fi
echo ""

# Test 3: Get initial balances
echo "3. Getting initial balances from DB..."
BALANCES_DB=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/users/balances-db")
DB_COUNT=$(echo "$BALANCES_DB" | jq -r '.count // 0')
echo -e "${GREEN}✓${NC} Found $DB_COUNT balances in database"
echo ""

# Test 4: Instructions for WebSocket testing
echo "4. WebSocket Testing Instructions:"
echo -e "${CYAN}To test WebSocket connection, use a WebSocket client:${NC}"
echo ""
echo "URL: $WS_URL"
echo "Headers:"
echo "  Authorization: Bearer $TOKEN"
echo ""
echo "Example with wscat (install: npm install -g wscat):"
echo "  wscat -c \"$WS_URL\" -H \"Authorization: Bearer $TOKEN\""
echo ""
echo "Expected messages:"
echo "  1. {\"type\":\"connected\",\"user_id\":<id>,\"message\":\"Connected to balance updates\"}"
echo "  2. {\"type\":\"balance_update\",...} (when balance changes)"
echo ""
echo "You can subscribe to specific denoms:"
echo "  {\"action\":\"subscribe\",\"denoms\":[\"unuah\",\"asset/GOLD\"]}"
echo ""
echo "Or unsubscribe:"
echo "  {\"action\":\"unsubscribe\"}"
echo ""

# Test 5: Check if sync service is running
echo "5. Checking sync service status..."
if pgrep -f "balance.*sync" > /dev/null 2>&1 || grep -q "Balance sync service started" /tmp/server.log 2>/dev/null; then
    echo -e "${GREEN}✓${NC} Balance sync service appears to be running"
else
    echo -e "${YELLOW}⚠${NC} Could not confirm sync service is running (check server logs)"
fi
echo ""

echo "================================"
echo -e "${GREEN}✅ Balance WebSocket test setup complete!${NC}"
echo ""
echo -e "${CYAN}Next steps:${NC}"
echo "1. Connect to WebSocket using the URL and token above"
echo "2. Perform a transaction (buy/sell asset) to trigger balance update"
echo "3. Verify you receive balance_update messages in real-time"
echo ""


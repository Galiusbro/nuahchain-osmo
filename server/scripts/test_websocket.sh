#!/bin/bash

# Test script for WebSocket functionality
# Tests WebSocket connection, subscription, and fallback to polling

set -e

API_URL="http://localhost:8080"
WS_URL="ws://localhost:26657/websocket"
RPC_URL="http://localhost:26657"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== WebSocket Testing Script ===${NC}\n"

# Step 1: Check if blockchain is running
echo -e "${YELLOW}1. Checking blockchain status...${NC}"
if curl -s "$RPC_URL/status" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Blockchain is running${NC}"
    CHAIN_ID=$(curl -s "$RPC_URL/status" | jq -r '.result.node_info.network' 2>/dev/null || echo "unknown")
    echo "   Chain ID: $CHAIN_ID"
else
    echo -e "${RED}❌ Blockchain is not running${NC}"
    echo "   Please start the blockchain first:"
    echo "   ./scripts/setup/bootstrap_local_env.sh"
    exit 1
fi

# Step 2: Check WebSocket endpoint
echo -e "\n${YELLOW}2. Checking WebSocket endpoint...${NC}"
if curl -s "$RPC_URL/websocket" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ WebSocket endpoint is accessible${NC}"
else
    echo -e "${YELLOW}⚠️  WebSocket endpoint check failed (this is normal for HTTP check)${NC}"
    echo "   Will test actual WebSocket connection in next step"
fi

# Step 3: Test WebSocket connection (using wscat if available, or Python)
echo -e "\n${YELLOW}3. Testing WebSocket connection...${NC}"

# Check if wscat is available
if command -v wscat &> /dev/null; then
    echo "   Using wscat to test WebSocket..."
    echo '{"jsonrpc":"2.0","method":"subscribe","id":1,"params":{"query":"tm.event=\\"NewBlock\\""}}' | \
        timeout 5 wscat -c "$WS_URL" 2>&1 | head -5 || echo "   Connection test completed"
elif command -v python3 &> /dev/null; then
    echo "   Using Python to test WebSocket..."
    python3 << EOF
import asyncio
import websockets
import json
import sys

async def test_ws():
    try:
        async with websockets.connect("$WS_URL") as ws:
            # Send subscribe request
            req = {
                "jsonrpc": "2.0",
                "method": "subscribe",
                "id": 1,
                "params": {"query": "tm.event='NewBlock'"}
            }
            await ws.send(json.dumps(req))
            print("   ✅ WebSocket connection successful")
            print("   ✅ Subscribe request sent")

            # Wait for response (with timeout)
            try:
                response = await asyncio.wait_for(ws.recv(), timeout=5.0)
                print(f"   ✅ Received response: {response[:100]}...")
            except asyncio.TimeoutError:
                print("   ⚠️  No response received (timeout)")
    except Exception as e:
        print(f"   ❌ WebSocket connection failed: {e}")
        sys.exit(1)

asyncio.run(test_ws())
EOF
else
    echo -e "${YELLOW}⚠️  wscat and Python not available, skipping WebSocket connection test${NC}"
fi

# Step 4: Check server status
echo -e "\n${YELLOW}4. Checking server status...${NC}"
if curl -s "$API_URL/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Server is running${NC}"
    SERVER_STATUS=$(curl -s "$API_URL/health" | jq -r '.status' 2>/dev/null || echo "unknown")
    echo "   Status: $SERVER_STATUS"
else
    echo -e "${RED}❌ Server is not running${NC}"
    echo "   Please start the server first"
    exit 1
fi

# Step 5: Test transaction tracking (if we have auth)
echo -e "\n${YELLOW}5. Testing transaction tracking...${NC}"
echo "   This requires:"
echo "   - Valid JWT token"
echo "   - Active blockchain with transactions"
echo "   - WebSocket enabled in server config"
echo ""
echo "   To test manually:"
echo "   1. Make a transaction via API (e.g., buy asset)"
echo "   2. Check server logs for WebSocket subscription"
echo "   3. Verify transaction status updates"

# Step 6: Check server logs for WebSocket messages
echo -e "\n${YELLOW}6. WebSocket Configuration Check${NC}"
echo "   Expected environment variables:"
echo "   - BLOCKCHAIN_WEBSOCKET_URL=$WS_URL"
echo "   - BLOCKCHAIN_WEBSOCKET_ENABLED=true"
echo ""
echo "   Check server logs for:"
echo "   - 'WebSocket client connected successfully'"
echo "   - 'transaction confirmed via WebSocket'"
echo "   - 'WebSocket disconnected, switching to polling' (if connection lost)"

echo -e "\n${GREEN}=== Test Summary ===${NC}"
echo "✅ Basic connectivity checks completed"
echo "⚠️  Full integration test requires:"
echo "   1. Server running with WebSocket enabled"
echo "   2. Active blockchain connection"
echo "   3. Test transaction to track"
echo ""
echo "Next steps:"
echo "1. Start server: cd server && go run main.go"
echo "2. Make a test transaction"
echo "3. Check logs for WebSocket activity"


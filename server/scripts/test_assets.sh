#!/bin/bash

# Test script for Assets API with WebSocket tracking

BASE_URL="http://localhost:8080"
WS_URL="ws://localhost:26657/websocket"
SERVER_LOG="/tmp/server.log"

echo "🧪 Testing Assets API with WebSocket Tracking"
echo "================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to check WebSocket status
check_websocket_status() {
    echo -e "${CYAN}📡 Checking WebSocket status...${NC}"

    # Check if server log exists and contains WebSocket messages
    if [ -f "$SERVER_LOG" ]; then
        if grep -q "WebSocket client connected successfully" "$SERVER_LOG"; then
            echo -e "${GREEN}✓${NC} WebSocket client connected (from server logs)"
        else
            echo -e "${YELLOW}⚠${NC} WebSocket connection not found in logs"
        fi

        # Check for WebSocket transaction confirmations
        WS_CONFIRMS=$(grep -c "transaction confirmed via WebSocket" "$SERVER_LOG" 2>/dev/null | head -1 || echo "0")
        POLLING_CONFIRMS=$(grep -c "transaction confirmed successfully via polling" "$SERVER_LOG" 2>/dev/null | head -1 || echo "0")

        WS_COUNT=$(echo "$WS_CONFIRMS" | tr -d '\n' | awk '{print $1}')
        POLLING_COUNT=$(echo "$POLLING_CONFIRMS" | tr -d '\n' | awk '{print $1}')

        if [ "${WS_COUNT:-0}" -gt 0 ]; then
            echo -e "${GREEN}✓${NC} WebSocket confirmations: $WS_COUNT"
        fi
        if [ "${POLLING_COUNT:-0}" -gt 0 ]; then
            echo -e "${YELLOW}⚠${NC} Polling fallback used: $POLLING_COUNT times"
        fi
    else
        echo -e "${YELLOW}⚠${NC} Server log not found at $SERVER_LOG"
    fi

    # Check WebSocket endpoint directly
    if curl -s http://localhost:26657/status > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Blockchain RPC endpoint is accessible"
    else
        echo -e "${RED}✗${NC} Blockchain RPC endpoint not accessible"
    fi
    echo ""
}

# Function to monitor transaction tracking
monitor_transaction() {
    local tx_hash=$1
    local operation=$2

    echo -e "${BLUE}🔍 Monitoring transaction: ${tx_hash:0:16}...${NC}"

    # Wait a bit for transaction to be tracked
    sleep 2

    # Check server logs for this transaction
    if [ -f "$SERVER_LOG" ]; then
        # Look for WebSocket confirmation
        if grep "$tx_hash" "$SERVER_LOG" 2>/dev/null | grep -q "via WebSocket"; then
            echo -e "${GREEN}✓${NC} Transaction tracked via WebSocket"
            return 0
        # Look for polling confirmation
        elif grep "$tx_hash" "$SERVER_LOG" 2>/dev/null | grep -q "via polling"; then
            echo -e "${YELLOW}⚠${NC} Transaction tracked via polling (WebSocket fallback)"
            return 0
        else
            echo -e "${YELLOW}⚠${NC} Transaction tracking status not found in logs yet"
        fi
    fi

    # Also check transaction status via API
    sleep 1
    TX_STATUS=$(curl -s "$BASE_URL/api/tx/$tx_hash" 2>/dev/null)
    if echo "$TX_STATUS" | jq -e '.found == true' > /dev/null 2>&1; then
        SUCCESS=$(echo "$TX_STATUS" | jq -r '.success // false')
        if [ "$SUCCESS" = "true" ]; then
            echo -e "${GREEN}✓${NC} Transaction confirmed on blockchain"
        else
            echo -e "${RED}✗${NC} Transaction failed on blockchain"
            ERROR=$(echo "$TX_STATUS" | jq -r '.error // "Unknown error"')
            echo "   Error: $ERROR"
        fi
    else
        echo -e "${YELLOW}⚠${NC} Transaction not yet found on blockchain"
    fi
    echo ""
}

# Test 0: WebSocket Status Check
echo "0. Checking WebSocket setup..."
check_websocket_status

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

# Test 2.5: Check initial balances
echo "2.5. Checking initial balances..."
INITIAL_BALANCES=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/users/balances-db")
INITIAL_COUNT=$(echo "$INITIAL_BALANCES" | jq -r '.count // 0')
echo "Initial balances in DB: $INITIAL_COUNT"
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

TX_HASH=$(echo "$ENSURE_RESPONSE" | jq -r '.tx_hash // empty')
STATUS=$(echo "$ENSURE_RESPONSE" | jq -r '.status // "unknown"')

if [ -n "$TX_HASH" ] && [ "$TX_HASH" != "null" ]; then
    echo -e "${GREEN}✓${NC} Asset ensure transaction broadcast"
    echo "Transaction hash: $TX_HASH"
    echo "Status: $STATUS"

    # Monitor transaction tracking
    monitor_transaction "$TX_HASH" "ensure"

    # Wait for balance indexer to process
    echo -e "${CYAN}⏳ Waiting for balance indexer to process transaction...${NC}"
    sleep 5

    # Check if balances were updated
    UPDATED_BALANCES=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/users/balances-db")
    UPDATED_COUNT=$(echo "$UPDATED_BALANCES" | jq -r '.count // 0')
    if [ "$UPDATED_COUNT" -ge "$INITIAL_COUNT" ]; then
        echo -e "${GREEN}✓${NC} Balance indexer updated balances (count: $INITIAL_COUNT -> $UPDATED_COUNT)"
    else
        echo -e "${YELLOW}⚠${NC} Balance indexer may not have processed yet"
    fi
else
    echo -e "${RED}✗${NC} Asset ensure failed"
    ERROR=$(echo "$ENSURE_RESPONSE" | jq -r '.error_msg // .error // empty')
    if [ -n "$ERROR" ]; then
        echo "Error: $ERROR"
    fi
fi
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

TX_HASH=$(echo "$BUY_RESPONSE" | jq -r '.tx_hash // empty')
STATUS=$(echo "$BUY_RESPONSE" | jq -r '.status // "unknown"')

if [ -n "$TX_HASH" ] && [ "$TX_HASH" != "null" ]; then
    echo -e "${GREEN}✓${NC} Asset buy transaction broadcast"
    echo "Transaction hash: $TX_HASH"
    echo "Status: $STATUS"
    BASE_AMOUNT=$(echo "$BUY_RESPONSE" | jq -r '.base_amount // "N/A"')
    echo "Base amount: $BASE_AMOUNT"

    # Monitor transaction tracking
    monitor_transaction "$TX_HASH" "buy"
else
    echo -e "${RED}✗${NC} Asset buy failed"
    ERROR=$(echo "$BUY_RESPONSE" | jq -r '.error_msg // .error // empty')
    if [ -n "$ERROR" ]; then
        echo "Error: $ERROR"
    fi
fi
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

TX_HASH=$(echo "$SELL_RESPONSE" | jq -r '.tx_hash // empty')
STATUS=$(echo "$SELL_RESPONSE" | jq -r '.status // "unknown"')

if [ -n "$TX_HASH" ] && [ "$TX_HASH" != "null" ]; then
    echo -e "${GREEN}✓${NC} Asset sell transaction broadcast"
    echo "Transaction hash: $TX_HASH"
    echo "Status: $STATUS"
    PAYOUT=$(echo "$SELL_RESPONSE" | jq -r '.payout_ndollar // "N/A"')
    echo "NDOLLAR payout: $PAYOUT"

    # Monitor transaction tracking
    monitor_transaction "$TX_HASH" "sell"
else
    echo -e "${RED}✗${NC} Asset sell failed"
    ERROR=$(echo "$SELL_RESPONSE" | jq -r '.error_msg // .error // empty')
    if [ -n "$ERROR" ]; then
        echo "Error: $ERROR"
    fi
fi
echo ""

echo "Waiting 3 seconds before margin operations..."
sleep 3
echo ""

echo "6. Testing Margin Open (long)..."
MARGIN_OPEN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/assets/margin/open" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "symbol": "GOLD",
    "side": "long",
    "quote_amount": "500",
    "leverage": "2"
  }')

echo "Response:"
echo "$MARGIN_OPEN_RESPONSE" | jq . || echo "$MARGIN_OPEN_RESPONSE"
echo ""

TX_HASH=$(echo "$MARGIN_OPEN_RESPONSE" | jq -r '.tx_hash // empty')
STATUS=$(echo "$MARGIN_OPEN_RESPONSE" | jq -r '.status // "unknown"')
POSITION_ID=$(echo "$MARGIN_OPEN_RESPONSE" | jq -r '.position_id // empty')

if [ -n "$TX_HASH" ] && [ "$TX_HASH" != "null" ]; then
    echo -e "${GREEN}✓${NC} Margin open transaction broadcast"
    echo "Transaction hash: $TX_HASH"
    echo "Status: $STATUS"
    if [ -n "$POSITION_ID" ] && [ "$POSITION_ID" != "null" ] && [ "$POSITION_ID" != "" ]; then
        echo "Position ID: $POSITION_ID"
    else
        echo -e "${YELLOW}ℹ${NC} Position ID will be available after block inclusion"
    fi

    # Monitor transaction tracking
    monitor_transaction "$TX_HASH" "margin_open"
else
    echo -e "${RED}✗${NC} Margin open failed"
    ERROR=$(echo "$MARGIN_OPEN_RESPONSE" | jq -r '.error_msg // .error // empty')
    if [ -n "$ERROR" ]; then
        echo "Error: $ERROR"
    fi
fi
echo ""

if [ -n "$POSITION_ID" ] && [ "$POSITION_ID" != "null" ] && [ "$POSITION_ID" != "" ]; then
    echo "Waiting 2 seconds before closing margin position..."
    sleep 2
    echo ""

    echo "7. Testing Margin Close..."
    MARGIN_CLOSE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/assets/margin/close" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d "{\"position_id\": \"$POSITION_ID\"}")

    echo "Response:"
    echo "$MARGIN_CLOSE_RESPONSE" | jq . || echo "$MARGIN_CLOSE_RESPONSE"
    echo ""

    TX_HASH_CLOSE=$(echo "$MARGIN_CLOSE_RESPONSE" | jq -r '.tx_hash // empty')
    STATUS_CLOSE=$(echo "$MARGIN_CLOSE_RESPONSE" | jq -r '.status // "unknown"')

    if [ -n "$TX_HASH_CLOSE" ] && [ "$TX_HASH_CLOSE" != "null" ]; then
        echo -e "${GREEN}✓${NC} Margin close transaction broadcast"
        echo "Transaction hash: $TX_HASH_CLOSE"
        echo "Status: $STATUS_CLOSE"
        PNL=$(echo "$MARGIN_CLOSE_RESPONSE" | jq -r '.pnl // "N/A"')
        echo "PnL: $PNL"

        # Monitor transaction tracking
        monitor_transaction "$TX_HASH_CLOSE" "margin_close"
    else
        echo -e "${RED}✗${NC} Margin close failed"
        ERROR=$(echo "$MARGIN_CLOSE_RESPONSE" | jq -r '.error_msg // .error // empty')
        if [ -n "$ERROR" ]; then
            echo "Error: $ERROR"
        fi
    fi
    echo ""
else
    echo -e "${YELLOW}!${NC} Could not find position ID, skipping margin close"
    echo "Note: Position may have been closed already or transaction is still pending"
    echo ""
fi

# Test 8: Check Balances
echo "8. Testing Balance Endpoints..."
echo ""

echo "8.1. Getting balances from database..."
BALANCES_DB=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/users/balances-db")
echo "$BALANCES_DB" | jq . || echo "$BALANCES_DB"
DB_COUNT=$(echo "$BALANCES_DB" | jq -r '.count // 0')
if [ "$DB_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓${NC} Found $DB_COUNT balances in database"
else
    echo -e "${YELLOW}⚠${NC} No balances in database (may need sync)"
fi
echo ""

echo "8.2. Getting balances from blockchain..."
BALANCES_BC=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/users/balances")
echo "$BALANCES_BC" | jq . || echo "$BALANCES_BC"
BC_COUNT=$(echo "$BALANCES_BC" | jq -r '.count // 0')
if [ "$BC_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓${NC} Found $BC_COUNT balances from blockchain"
else
    echo -e "${RED}✗${NC} No balances found from blockchain"
fi
echo ""

echo "8.3. Syncing balances..."
SYNC_RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/users/balances/sync")
echo "$SYNC_RESPONSE" | jq . || echo "$SYNC_RESPONSE"
if echo "$SYNC_RESPONSE" | jq -e '.message' > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Balance sync completed"
else
    echo -e "${YELLOW}⚠${NC} Balance sync response unclear"
fi
echo ""

echo "8.4. Getting balance history..."
HISTORY=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/users/balances/history?limit=10")
echo "$HISTORY" | jq . || echo "$HISTORY"
HISTORY_COUNT=$(echo "$HISTORY" | jq -r '.count // 0')
if [ "$HISTORY_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓${NC} Found $HISTORY_COUNT balance history entries"
else
    echo -e "${YELLOW}⚠${NC} No balance history yet"
fi
echo ""

echo "8.5. Testing balance filter (specific denom)..."
if [ "$BC_COUNT" -gt 0 ]; then
    FIRST_DENOM=$(echo "$BALANCES_BC" | jq -r '.balances[0].denom // empty')
    if [ -n "$FIRST_DENOM" ] && [ "$FIRST_DENOM" != "null" ]; then
        FILTERED=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/users/balances-db?denom=$FIRST_DENOM")
        FILTERED_COUNT=$(echo "$FILTERED" | jq -r '.count // 0')
        if [ "$FILTERED_COUNT" -eq 1 ]; then
            echo -e "${GREEN}✓${NC} Balance filter works (found 1 balance for $FIRST_DENOM)"
        else
            echo -e "${YELLOW}⚠${NC} Balance filter returned $FILTERED_COUNT results (expected 1)"
        fi
    fi
fi
echo ""

echo "8.6. Testing Balance Indexer (real-time updates)..."
echo -e "${CYAN}📊 Checking if balance indexer is processing transactions...${NC}"
if [ -f "$SERVER_LOG" ]; then
    INDEXER_STARTED=$(grep -c "Balance indexer started" "$SERVER_LOG" 2>/dev/null | head -1 | tr -d '\n' | awk '{print $1}' || echo "0")
    INDEXER_PROCESSING=$(grep -c "Balance indexer processEvents started" "$SERVER_LOG" 2>/dev/null | head -1 | tr -d '\n' | awk '{print $1}' || echo "0")
    BALANCE_UPDATES=$(grep -c "balance_update" "$SERVER_LOG" 2>/dev/null | head -1 | tr -d '\n' | awk '{print $1}' || echo "0")

    INDEXER_STARTED_COUNT=${INDEXER_STARTED:-0}
    INDEXER_PROCESSING_COUNT=${INDEXER_PROCESSING:-0}
    BALANCE_UPDATES_COUNT=${BALANCE_UPDATES:-0}

    if [ "$INDEXER_STARTED_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓${NC} Balance indexer is running"
    else
        echo -e "${YELLOW}⚠${NC} Balance indexer may not be running"
    fi

    if [ "$INDEXER_PROCESSING_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓${NC} Balance indexer is processing events"
    fi

    if [ "$BALANCE_UPDATES_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓${NC} Balance updates detected in logs ($BALANCE_UPDATES_COUNT)"
    else
        echo -e "${YELLOW}ℹ${NC} No balance updates in logs yet (may need transactions)"
    fi
else
    echo -e "${YELLOW}⚠${NC} Server log not available"
fi
echo ""

echo "8.7. Testing Periodic Sync Service..."
echo -e "${CYAN}📊 Checking if balance sync service is running...${NC}"
if [ -f "$SERVER_LOG" ]; then
    SYNC_STARTED=$(grep -c "Balance sync service started" "$SERVER_LOG" 2>/dev/null | head -1 | tr -d '\n' | awk '{print $1}' || echo "0")
    SYNC_BATCH=$(grep -c "Starting balance sync batch" "$SERVER_LOG" 2>/dev/null | head -1 | tr -d '\n' | awk '{print $1}' || echo "0")
    SYNC_COMPLETED=$(grep -c "Balance sync batch completed" "$SERVER_LOG" 2>/dev/null | head -1 | tr -d '\n' | awk '{print $1}' || echo "0")

    SYNC_STARTED_COUNT=${SYNC_STARTED:-0}
    SYNC_BATCH_COUNT=${SYNC_BATCH:-0}
    SYNC_COMPLETED_COUNT=${SYNC_COMPLETED:-0}

    if [ "$SYNC_STARTED_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓${NC} Balance sync service is running"
    else
        echo -e "${YELLOW}⚠${NC} Balance sync service may not be running"
    fi

    if [ "$SYNC_BATCH_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓${NC} Sync batches executed: $SYNC_BATCH_COUNT"
    fi

    if [ "$SYNC_COMPLETED_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓${NC} Sync batches completed: $SYNC_COMPLETED_COUNT"
    fi
else
    echo -e "${YELLOW}⚠${NC} Server log not available"
fi
echo ""

echo "8.8. Testing WebSocket Endpoint..."
echo -e "${CYAN}📡 Checking WebSocket endpoint availability...${NC}"
WS_TEST=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/users/balances/ws" 2>/dev/null)
if [ "$WS_TEST" = "400" ] || [ "$WS_TEST" = "426" ] || [ "$WS_TEST" = "101" ]; then
    echo -e "${GREEN}✓${NC} WebSocket endpoint is accessible (HTTP $WS_TEST - expected for non-WebSocket request)"
    echo -e "${CYAN}ℹ${NC} To test WebSocket connection, use:"
    echo "   wscat -c \"ws://localhost:8080/api/users/balances/ws\" -H \"Authorization: Bearer $TOKEN\""
else
    echo -e "${YELLOW}⚠${NC} Unexpected response code: $WS_TEST"
fi
echo ""

echo "================================"
echo -e "${GREEN}✅ Assets API testing complete!${NC}"
echo ""

# Final WebSocket statistics
echo -e "${CYAN}📊 WebSocket Statistics:${NC}"
if [ -f "$SERVER_LOG" ]; then
    WS_TOTAL=$(grep -c "transaction confirmed via WebSocket" "$SERVER_LOG" 2>/dev/null | head -1 | tr -d '\n' | awk '{print $1}' || echo "0")
    POLLING_TOTAL=$(grep -c "transaction confirmed successfully via polling" "$SERVER_LOG" 2>/dev/null | head -1 | tr -d '\n' | awk '{print $1}' || echo "0")
    WS_FALLBACK=$(grep -c "WebSocket failed, falling back to polling" "$SERVER_LOG" 2>/dev/null | head -1 | tr -d '\n' | awk '{print $1}' || echo "0")
    WS_RESTORED=$(grep -c "WebSocket restored, switching back from polling" "$SERVER_LOG" 2>/dev/null | head -1 | tr -d '\n' | awk '{print $1}' || echo "0")

    WS_COUNT=${WS_TOTAL:-0}
    POLLING_COUNT=${POLLING_TOTAL:-0}
    FALLBACK_COUNT=${WS_FALLBACK:-0}
    RESTORED_COUNT=${WS_RESTORED:-0}

    echo "   WebSocket confirmations: $WS_COUNT"
    echo "   Polling fallback uses: $POLLING_COUNT"
    if [ "$FALLBACK_COUNT" -gt 0 ]; then
        echo -e "   ${YELLOW}⚠ Fallback events: $FALLBACK_COUNT${NC}"
    fi
    if [ "$RESTORED_COUNT" -gt 0 ]; then
        echo -e "   ${GREEN}✓ Restored events: $RESTORED_COUNT${NC}"
    fi

    if [ "$WS_COUNT" -gt 0 ]; then
        echo -e "   ${GREEN}✅ WebSocket is working!${NC}"
    elif [ "$POLLING_COUNT" -gt 0 ]; then
        echo -e "   ${YELLOW}⚠ Using polling fallback${NC}"
    else
        echo -e "   ${YELLOW}ℹ No transactions tracked yet${NC}"
    fi
else
    echo "   Server log not available"
fi
echo ""


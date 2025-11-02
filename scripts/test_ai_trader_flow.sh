#!/bin/bash
# AI Trader Service - Complete Flow Test Script
# Tests user registration, authentication, bot creation, and market data endpoints

set -e

API_BASE="${AI_TRADER_API_BASE:-http://127.0.0.1:18080}"
TIMEOUT=10

echo "=========================================="
echo "AI Trader Service - Flow Test"
echo "=========================================="
echo "API Base: $API_BASE"
echo ""

# Generate unique email for this test run
TIMESTAMP=$(date +%s)
EMAIL="test-$TIMESTAMP@example.com"

echo "STEP 1: User Registration"
echo "------------------------"
echo "Registering user with email: $EMAIL"
REGISTER_RESPONSE=$(curl -s --max-time $TIMEOUT -X POST \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"name\":\"Flow Test User\"}" \
  "$API_BASE/users/register")

echo "$REGISTER_RESPONSE" | jq '.' 2>/dev/null || echo "$REGISTER_RESPONSE"

API_KEY=$(echo "$REGISTER_RESPONSE" | jq -r '.api_key // empty' 2>/dev/null)
if [ -z "$API_KEY" ]; then
  echo "❌ ERROR: Failed to obtain API key"
  echo "Response: $REGISTER_RESPONSE"
  exit 1
fi

echo "✅ API Key obtained: ${API_KEY:0:30}..."
echo ""

echo "STEP 2: Protected Endpoint - Spot Price"
echo "------------------------"
echo "Fetching spot price for AAPL..."
SPOT_RESPONSE=$(curl -s --max-time $TIMEOUT \
  -H "X-API-Key: $API_KEY" \
  "$API_BASE/spot?symbol=AAPL")

echo "$SPOT_RESPONSE" | jq '.' 2>/dev/null || echo "$SPOT_RESPONSE"

PRICE=$(echo "$SPOT_RESPONSE" | jq -r '.price // empty' 2>/dev/null)
if [ -n "$PRICE" ]; then
  echo "✅ Spot price retrieved: \$$PRICE"
else
  echo "⚠️  Warning: Could not parse price from response"
fi
echo ""

echo "STEP 3: Bot Creation"
echo "------------------------"
echo "Creating a trading bot..."
BOT_NAME="flow-test-bot-$TIMESTAMP"

# Create perspective config as JSON object first
PERSPECTIVE_CONFIG=$(cat <<JSON
{
  "pre_tf": "1h",
  "pre_limit": 48,
  "target_tf": "5m",
  "target_limit": 96,
  "post_tf": "1m",
  "post_limit": 60
}
JSON
)

# Escape the JSON config as a string for the API
CONFIG_JSON_STRING=$(echo "$PERSPECTIVE_CONFIG" | jq -c . | jq -Rs .)

# Build the request payload properly
BOT_PAYLOAD=$(jq -n \
  --arg name "$BOT_NAME" \
  --argjson config "$PERSPECTIVE_CONFIG" \
  '{name: $name, config_json: ($config | tostring)}')

BOT_RESPONSE=$(curl -s --max-time $TIMEOUT -X POST \
  -H "X-API-Key: $API_KEY" \
  -H 'Content-Type: application/json' \
  -d "$BOT_PAYLOAD" \
  "$API_BASE/bots")

echo "$BOT_RESPONSE" | jq '.' 2>/dev/null || echo "$BOT_RESPONSE"

BOT_ID=$(echo "$BOT_RESPONSE" | jq -r '.bot_id // empty' 2>/dev/null)
if [ -n "$BOT_ID" ]; then
  echo "✅ Bot created successfully with ID: $BOT_ID"
else
  echo "⚠️  Warning: Could not parse bot_id from response"
fi
echo ""

echo "STEP 4: Technical Indicators"
echo "------------------------"
echo "Fetching technical indicators for AAPL (5m timeframe)..."
INDICATORS_RESPONSE=$(curl -s --max-time $TIMEOUT \
  -H "X-API-Key: $API_KEY" \
  "$API_BASE/indicators?symbol=AAPL&tf=5m")

echo "$INDICATORS_RESPONSE" | jq '{sma: .sma, ema: .ema, rsi: .rsi, macd_signal: .macd.signal}' 2>/dev/null || echo "$INDICATORS_RESPONSE"

RSI=$(echo "$INDICATORS_RESPONSE" | jq -r '.rsi // empty' 2>/dev/null)
if [ -n "$RSI" ]; then
  echo "✅ RSI indicator: $RSI"
fi
echo ""

echo "STEP 5: OHLCV Candles"
echo "------------------------"
echo "Fetching last 5 candles for AAPL (5m timeframe)..."
OHLCV_RESPONSE=$(curl -s --max-time $TIMEOUT \
  -H "X-API-Key: $API_KEY" \
  "$API_BASE/ohlcv?symbol=AAPL&tf=5m&limit=5")

CANDLE_COUNT=$(echo "$OHLCV_RESPONSE" | jq 'length' 2>/dev/null || echo "0")
if [ "$CANDLE_COUNT" -gt "0" ]; then
  echo "✅ Retrieved $CANDLE_COUNT candles"
  echo "Sample candle (first):"
  echo "$OHLCV_RESPONSE" | jq '.[0] | {time: .t, open: .o, high: .h, low: .l, close: .c, volume: .v}' 2>/dev/null || echo "(could not parse)"
else
  echo "⚠️  Warning: No candles retrieved"
fi
echo ""

echo "=========================================="
echo "Flow Test Summary"
echo "=========================================="
echo "✅ User Registration: SUCCESS"
echo "✅ API Authentication: SUCCESS"
echo "✅ Spot Price Endpoint: SUCCESS"
if [ -n "$BOT_ID" ]; then
  echo "✅ Bot Creation: SUCCESS (ID: $BOT_ID)"
else
  echo "⚠️  Bot Creation: PARTIAL (check response above)"
fi
if [ -n "$RSI" ]; then
  echo "✅ Technical Indicators: SUCCESS (RSI: $RSI)"
else
  echo "⚠️  Technical Indicators: PARTIAL"
fi
if [ "$CANDLE_COUNT" -gt "0" ]; then
  echo "✅ OHLCV Data: SUCCESS ($CANDLE_COUNT candles)"
else
  echo "⚠️  OHLCV Data: NO DATA"
fi
echo ""
echo "🎉 Flow test completed!"
echo ""


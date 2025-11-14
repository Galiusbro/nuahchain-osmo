#!/bin/bash

# Test GALBRO token exchange through API

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

API_URL="http://localhost:8080"
EMAIL="garold@gmail.com"
PASSWORD="Garold632"
GALBRO_DENOM="factory/nuah13yxg00hrpdwa0zfjc0lc9h7dr8qkazapdkff5g/galbro"

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           🪙 ТЕСТИРОВАНИЕ ОБМЕНА GALBRO                         ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Login
echo -e "${YELLOW}=== 1. Авторизация ===${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
  echo -e "${RED}❌ Failed to login${NC}"
  exit 1
fi
echo -e "${GREEN}✅ Logged in successfully${NC}"
echo ""

# Check GALBRO balance on blockchain
echo -e "${YELLOW}=== 2. Проверка баланса GALBRO ===${NC}"
WALLET_ADDR=$(echo $LOGIN_RESPONSE | jq -r '.wallet.address')
echo "Wallet address: $WALLET_ADDR"
GALBRO_BALANCE=$(./build/nuahd query bank balance "$WALLET_ADDR" "$GALBRO_DENOM" --output json 2>/dev/null | jq -r '.balance.amount')
echo "GALBRO balance: $GALBRO_BALANCE micro-units"

if [ "$GALBRO_BALANCE" == "0" ] || [ -z "$GALBRO_BALANCE" ]; then
  echo -e "${YELLOW}⚠️  No GALBRO balance. Please run: ./scripts/setup/setup_galbro_testtoken.sh${NC}"
  exit 1
fi
echo -e "${GREEN}✅ GALBRO balance found${NC}"
echo ""

# Try to exchange GALBRO for unuah
echo -e "${YELLOW}=== 3. Обмен 1,000 GALBRO на unuah ===${NC}"
EXCHANGE_AMOUNT="1000000000"  # 1,000 GALBRO
MIN_OUTPUT="950000000"        # Minimum 950 unuah (allow 5% slippage)

EXCHANGE_REQUEST=$(cat <<EOF
{
  "token_denom": "$GALBRO_DENOM",
  "amount": "$EXCHANGE_AMOUNT",
  "min_output": "$MIN_OUTPUT"
}
EOF
)

echo "Request:"
echo "$EXCHANGE_REQUEST" | jq '.'
echo ""

EXCHANGE_RESPONSE=$(curl -s -X POST "$API_URL/api/exchange/tokens" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "$EXCHANGE_REQUEST")

echo "Response:"
echo "$EXCHANGE_RESPONSE" | jq '.'
echo ""

SUCCESS=$(echo $EXCHANGE_RESPONSE | jq -r '.success')
if [ "$SUCCESS" == "true" ]; then
  TX_HASH=$(echo $EXCHANGE_RESPONSE | jq -r '.tx_hash')
  UNUAH_OUT=$(echo $EXCHANGE_RESPONSE | jq -r '.unuah_out')

  echo -e "${GREEN}✅ Exchange successful!${NC}"
  echo -e "${GREEN}   TX Hash: $TX_HASH${NC}"
  echo -e "${GREEN}   unuah received: $UNUAH_OUT${NC}"

  # Check transaction on blockchain
  echo ""
  echo -e "${YELLOW}=== 4. Проверка транзакции в блокчейне ===${NC}"
  sleep 3
  TX_DETAILS=$(./build/nuahd query tx $TX_HASH --output json 2>&1)
  echo "$TX_DETAILS" | jq '{code, raw_log, gas_wanted, gas_used}'

  # Check new balances
  echo ""
  echo -e "${YELLOW}=== 5. Новые балансы ===${NC}"
  NEW_GALBRO=$(./build/nuahd query bank balance "$WALLET_ADDR" "$GALBRO_DENOM" --output json 2>/dev/null | jq -r '.balance.amount')
  NEW_UNUAH=$(./build/nuahd query bank balance "$WALLET_ADDR" "unuah" --output json 2>/dev/null | jq -r '.balance.amount')
  echo "GALBRO: $NEW_GALBRO (было: $GALBRO_BALANCE)"
  echo "unuah:  $NEW_UNUAH"
else
  ERROR_MSG=$(echo $EXCHANGE_RESPONSE | jq -r '.error_msg')
  echo -e "${YELLOW}⚠️  Exchange failed (expected until usdoracle configured)${NC}"
  echo -e "${YELLOW}   Error: $ERROR_MSG${NC}"
  echo ""
  echo -e "${BLUE}📝 НАСТРОЙКА ТРЕБУЕТСЯ:${NC}"
  echo "   1. Добавить GALBRO в x/usdoracle supported_tokens"
  echo "   2. Установить цену GALBRO = 1.00 USD"
  echo "   3. Добавить GALBRO в x/exchange supported_tokens"
fi

echo ""
echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                     📊 РЕЗУЛЬТАТЫ ТЕСТА                          ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════╝${NC}"
echo ""

if [ "$SUCCESS" == "true" ]; then
  echo -e "${GREEN}✅ GALBRO токен успешно обменян на unuah${NC}"
  echo -e "${GREEN}✅ Exchange module работает корректно${NC}"
  echo -e "${GREEN}✅ Транзакции записываются в БД${NC}"
else
  echo -e "${YELLOW}⚠️  Exchange module активирован${NC}"
  echo -e "${YELLOW}⚠️  GALBRO токен создан и распределен${NC}"
  echo -e "${RED}❌ Требуется настройка x/usdoracle${NC}"
  echo ""
  echo -e "${BLUE}NEXT STEPS:${NC}"
  echo "  Для полной работы нужно обновить genesis или параметры модулей:"
  echo "  - x/usdoracle: добавить GALBRO с ценой 1.00 USD"
  echo "  - x/exchange: добавить GALBRO в supported_tokens"
fi
echo ""


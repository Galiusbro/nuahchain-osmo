#!/bin/bash

# Скрипт для получения курсов валют, золота и ценных бумаг через TwelveData API
# API ключ: 01124156c61844b386a6f017d1836e0c

API_KEY="01124156c61844b386a6f017d1836e0c"
BASE_URL="https://api.twelvedata.com"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
OUTPUT_FILE="market_data_${TIMESTAMP}.json"

echo "🚀 Получение данных с TwelveData API"
echo "=================================="

# Функция для получения данных
get_data() {
    local symbol=$1
    local name=$2

    echo "📊 Получение $name ($symbol)..."

    response=$(curl -s "${BASE_URL}/time_series?symbol=${symbol}&interval=1day&apikey=${API_KEY}&outputsize=1")

    if echo "$response" | grep -q '"status":"ok"'; then
        echo "  ✅ $symbol: $(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(f\"{data['values'][0]['close']} (время: {data['values'][0]['datetime']})\")" 2>/dev/null || echo "Данные получены")"
        echo "$response" >> temp_data.json
    else
        echo "  ❌ $symbol: Ошибка получения данных"
        echo "  Ответ: $response"
    fi
}

# Очищаем временный файл
> temp_data.json

echo "💱 Курсы валют:"
echo "--------------"
get_data "EUR/USD" "Евро/Доллар"
get_data "GBP/USD" "Фунт/Доллар"
get_data "USD/JPY" "Доллар/Йена"
get_data "USD/CHF" "Доллар/Франк"
get_data "AUD/USD" "Австралийский доллар/Доллар"
get_data "USD/CAD" "Доллар/Канадский доллар"
get_data "USD/RUB" "Доллар/Рубль"
get_data "USD/CNY" "Доллар/Юань"
get_data "USD/INR" "Доллар/Рупия"
get_data "USD/BRL" "Доллар/Реал"
get_data "USD/MXN" "Доллар/Песо"

echo ""
echo "🥇 Цены драгоценных металлов:"
echo "----------------------------"
get_data "XAU/USD" "Золото"
get_data "XAG/USD" "Серебро"
get_data "XPT/USD" "Платина"
get_data "XPD/USD" "Палладий"

echo ""
echo "📈 Цены акций:"
echo "-------------"
get_data "AAPL" "Apple"
get_data "GOOGL" "Google"
get_data "MSFT" "Microsoft"
get_data "AMZN" "Amazon"
get_data "TSLA" "Tesla"
get_data "META" "Meta"
get_data "NVDA" "NVIDIA"
get_data "NFLX" "Netflix"
get_data "AMD" "AMD"
get_data "INTC" "Intel"

echo ""
echo "₿ Цены криптовалют:"
echo "------------------"
get_data "BTC/USD" "Bitcoin"
get_data "ETH/USD" "Ethereum"
get_data "BNB/USD" "Binance Coin"
get_data "ADA/USD" "Cardano"
get_data "SOL/USD" "Solana"
get_data "XRP/USD" "Ripple"
get_data "DOT/USD" "Polkadot"
get_data "DOGE/USD" "Dogecoin"
get_data "AVAX/USD" "Avalanche"
get_data "MATIC/USD" "Polygon"

echo ""
echo "🏛️ Индексы:"
echo "----------"
get_data "SPX" "S&P 500"
get_data "DJI" "Dow Jones"
get_data "IXIC" "NASDAQ"
get_data "VIX" "VIX (Индекс волатильности)"

# Создаем итоговый JSON файл
echo "💾 Сохранение результатов в $OUTPUT_FILE..."

# Создаем JSON структуру
cat > "$OUTPUT_FILE" << EOF
{
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "data_source": "TwelveData API",
  "api_key": "$API_KEY",
  "market_data": [
EOF

# Добавляем данные из временного файла
if [ -s temp_data.json ]; then
    # Убираем последнюю запятую и закрываем JSON
    sed '$ s/,$//' temp_data.json >> "$OUTPUT_FILE"
fi

cat >> "$OUTPUT_FILE" << EOF
  ]
}
EOF

# Удаляем временный файл
rm -f temp_data.json

echo "✅ Данные сохранены в файл: $OUTPUT_FILE"
echo ""
echo "📊 Статистика:"
echo "  - Время выполнения: $(date)"
echo "  - Файл результатов: $OUTPUT_FILE"
echo "  - Размер файла: $(wc -c < "$OUTPUT_FILE") байт"

# Показываем краткую сводку
echo ""
echo "📋 Краткая сводка:"
echo "=================="

# Показываем несколько ключевых курсов
echo "💱 Ключевые валютные пары:"
curl -s "${BASE_URL}/time_series?symbol=EUR/USD&interval=1day&apikey=${API_KEY}&outputsize=1" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if 'values' in data and data['values']:
        print(f\"  EUR/USD: {data['values'][0]['close']}\")
except:
    pass
" 2>/dev/null

curl -s "${BASE_URL}/time_series?symbol=USD/RUB&interval=1day&apikey=${API_KEY}&outputsize=1" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if 'values' in data and data['values']:
        print(f\"  USD/RUB: {data['values'][0]['close']}\")
except:
    pass
" 2>/dev/null

echo ""
echo "🥇 Золото:"
curl -s "${BASE_URL}/time_series?symbol=XAU/USD&interval=1day&apikey=${API_KEY}&outputsize=1" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if 'values' in data and data['values']:
        print(f\"  XAU/USD: {data['values'][0]['close']}\")
except:
    pass
" 2>/dev/null

echo ""
echo "📈 Ключевые акции:"
curl -s "${BASE_URL}/time_series?symbol=AAPL&interval=1day&apikey=${API_KEY}&outputsize=1" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if 'values' in data and data['values']:
        print(f\"  AAPL: \${data['values'][0]['close']}\")
except:
    pass
" 2>/dev/null

curl -s "${BASE_URL}/time_series?symbol=TSLA&interval=1day&apikey=${API_KEY}&outputsize=1" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if 'values' in data and data['values']:
        print(f\"  TSLA: \${data['values'][0]['close']}\")
except:
    pass
" 2>/dev/null

echo ""
echo "₿ Криптовалюты:"
curl -s "${BASE_URL}/time_series?symbol=BTC/USD&interval=1day&apikey=${API_KEY}&outputsize=1" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if 'values' in data and data['values']:
        print(f\"  BTC/USD: \${data['values'][0]['close']}\")
except:
    pass
" 2>/dev/null

curl -s "${BASE_URL}/time_series?symbol=ETH/USD&interval=1day&apikey=${API_KEY}&outputsize=1" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if 'values' in data and data['values']:
        print(f\"  ETH/USD: \${data['values'][0]['close']}\")
except:
    pass
" 2>/dev/null

echo ""
echo "🎉 Готово! Все данные получены и сохранены."

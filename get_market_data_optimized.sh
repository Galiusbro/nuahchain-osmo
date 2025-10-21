#!/bin/bash

# Оптимизированный скрипт для получения курсов валют, золота и ценных бумаг через TwelveData API
# С задержкой между запросами для соблюдения лимитов API

API_KEY="01124156c61844b386a6f017d1836e0c"
BASE_URL="https://api.twelvedata.com"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
OUTPUT_FILE="market_data_optimized_${TIMESTAMP}.json"

echo "🚀 Получение данных с TwelveData API (оптимизированная версия)"
echo "============================================================="

# Функция для получения данных с задержкой
get_data_with_delay() {
    local symbol=$1
    local name=$2
    local delay=${3:-10}  # Задержка по умолчанию 10 секунд

    echo "📊 Получение $name ($symbol)..."

    response=$(curl -s "${BASE_URL}/time_series?symbol=${symbol}&interval=1day&apikey=${API_KEY}&outputsize=1")

    if echo "$response" | grep -q '"status":"ok"'; then
        close_price=$(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data['values'][0]['close'])" 2>/dev/null)
        datetime=$(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data['values'][0]['datetime'])" 2>/dev/null)
        echo "  ✅ $symbol: $close_price (время: $datetime)"
        echo "$response" >> temp_data.json
        return 0
    else
        echo "  ❌ $symbol: Ошибка получения данных"
        if echo "$response" | grep -q "429"; then
            echo "  ⏳ Превышен лимит запросов. Ожидание $delay секунд..."
            sleep $delay
            return 1
        else
            echo "  Ответ: $response"
            return 1
        fi
    fi
}

# Очищаем временный файл
> temp_data.json

echo "💱 Курсы валют (основные):"
echo "-------------------------"

# Получаем основные валютные пары
get_data_with_delay "EUR/USD" "Евро/Доллар" 10
get_data_with_delay "GBP/USD" "Фунт/Доллар" 10
get_data_with_delay "USD/JPY" "Доллар/Йена" 10
get_data_with_delay "USD/RUB" "Доллар/Рубль" 10
get_data_with_delay "USD/CNY" "Доллар/Юань" 10

echo ""
echo "🥇 Цены драгоценных металлов:"
echo "----------------------------"
get_data_with_delay "XAU/USD" "Золото" 10
get_data_with_delay "XAG/USD" "Серебро" 10

echo ""
echo "📈 Цены акций (топ-5):"
echo "---------------------"
get_data_with_delay "AAPL" "Apple" 10
get_data_with_delay "GOOGL" "Google" 10
get_data_with_delay "MSFT" "Microsoft" 10
get_data_with_delay "TSLA" "Tesla" 10
get_data_with_delay "NVDA" "NVIDIA" 10

echo ""
echo "₿ Цены криптовалют (топ-5):"
echo "--------------------------"
get_data_with_delay "BTC/USD" "Bitcoin" 10
get_data_with_delay "ETH/USD" "Ethereum" 10
get_data_with_delay "BNB/USD" "Binance Coin" 10
get_data_with_delay "ADA/USD" "Cardano" 10
get_data_with_delay "SOL/USD" "Solana" 10

echo ""
echo "🏛️ Индексы:"
echo "----------"
get_data_with_delay "SPX" "S&P 500" 10
get_data_with_delay "VIX" "VIX (Индекс волатильности)" 10

# Создаем итоговый JSON файл
echo ""
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
echo "📋 Краткая сводка полученных данных:"
echo "===================================="

# Показываем все полученные курсы
if [ -s "$OUTPUT_FILE" ]; then
    echo "💱 Валютные пары:"
    python3 -c "
import json
try:
    with open('$OUTPUT_FILE', 'r') as f:
        data = json.load(f)

    for item in data['market_data']:
        if 'values' in item and item['values']:
            symbol = item['meta']['symbol']
            close = item['values'][0]['close']
            print(f\"  {symbol}: {close}\")
except Exception as e:
    print(f\"Ошибка чтения файла: {e}\")
" 2>/dev/null
fi

echo ""
echo "🎉 Готово! Все данные получены и сохранены."
echo ""
echo "💡 Рекомендации:"
echo "  - Для получения большего количества данных рассмотрите платный план"
echo "  - Текущий лимит: 8 запросов в минуту"
echo "  - Для массового получения данных используйте задержки между запросами"

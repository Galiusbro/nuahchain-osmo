#!/bin/bash

# Простой скрипт для проверки цены одного символа
# Использование: ./check_price.sh SYMBOL
# Пример: ./check_price.sh EUR/USD

API_KEY="01124156c61844b386a6f017d1836e0c"
BASE_URL="https://api.twelvedata.com"

if [ $# -eq 0 ]; then
    echo "❌ Ошибка: Укажите символ для проверки"
    echo "Использование: $0 SYMBOL"
    echo "Примеры:"
    echo "  $0 EUR/USD"
    echo "  $0 BTC/USD"
    echo "  $0 AAPL"
    echo "  $0 XAU/USD"
    exit 1
fi

SYMBOL=$1

echo "🔍 Проверка цены для $SYMBOL..."
echo "================================"

response=$(curl -s "${BASE_URL}/time_series?symbol=${SYMBOL}&interval=1day&apikey=${API_KEY}&outputsize=1")

if echo "$response" | grep -q '"status":"ok"'; then
    echo "✅ Данные получены успешно!"
    echo ""

    # Парсим и показываем данные
    python3 -c "
import json
import sys

try:
    data = json.loads('$response')

    print('📊 Информация о символе:')
    print(f\"  Символ: {data['meta']['symbol']}\")
    print(f\"  Тип: {data['meta']['type']}\")
    if 'currency_base' in data['meta']:
        print(f\"  Базовая валюта: {data['meta']['currency_base']}\")
    if 'currency_quote' in data['meta']:
        print(f\"  Котируемая валюта: {data['meta']['currency_quote']}\")

    print('')
    print('💰 Ценовые данные:')
    if 'values' in data and data['values']:
        latest = data['values'][0]
        print(f\"  Дата: {latest['datetime']}\")
        print(f\"  Открытие: {latest['open']}\")
        print(f\"  Максимум: {latest['high']}\")
        print(f\"  Минимум: {latest['low']}\")
        print(f\"  Закрытие: {latest['close']}\")
        if 'volume' in latest:
            print(f\"  Объем: {latest['volume']}\")

except Exception as e:
    print(f\"Ошибка парсинга: {e}\")
    print('Сырой ответ:')
    print('$response')
" 2>/dev/null

else
    echo "❌ Ошибка получения данных:"
    echo "$response"

    if echo "$response" | grep -q "429"; then
        echo ""
        echo "💡 Превышен лимит запросов. Попробуйте позже."
    elif echo "$response" | grep -q "404"; then
        echo ""
        echo "💡 Символ не найден. Проверьте правильность написания."
    fi
fi

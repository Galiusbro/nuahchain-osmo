#!/bin/bash

# Скрипт для получения всех данных одним запросом через TwelveData API
# Это намного эффективнее, чем отдельные запросы!

API_KEY="01124156c61844b386a6f017d1836e0c"
BASE_URL="https://api.twelvedata.com"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
OUTPUT_FILE="all_market_data_${TIMESTAMP}.json"

echo "🚀 Получение ВСЕХ данных одним запросом через TwelveData API"
echo "=========================================================="

# Определяем все символы для получения
SYMBOLS="EUR/USD,GBP/USD,USD/JPY,USD/CHF,AUD/USD,USD/CAD,USD/RUB,USD/CNY,USD/INR,USD/BRL,USD/MXN,XAU/USD,XAG/USD,XPT/USD,XPD/USD,AAPL,GOOGL,MSFT,AMZN,TSLA,META,NVDA,NFLX,AMD,INTC,BTC/USD,ETH/USD,BNB/USD,ADA/USD,SOL/USD,XRP/USD,DOT/USD,DOGE/USD,AVAX/USD,MATIC/USD"

echo "📊 Получение данных для всех символов..."
echo "Символы: $SYMBOLS"
echo ""

# Делаем один запрос для всех символов
response=$(curl -s "${BASE_URL}/time_series?symbol=${SYMBOLS}&interval=1day&apikey=${API_KEY}&outputsize=1")

# Проверяем ответ
if echo "$response" | grep -q '"status":"ok"'; then
    echo "✅ Данные получены успешно одним запросом!"
    echo ""
    
    # Сохраняем в файл
    echo "$response" | python3 -m json.tool > "$OUTPUT_FILE"
    
    echo "💾 Данные сохранены в файл: $OUTPUT_FILE"
    echo ""
    
    # Показываем краткую сводку
    echo "📋 Краткая сводка полученных данных:"
    echo "===================================="
    
    # Парсим и показываем данные
    python3 -c "
import json
import sys

try:
    with open('$OUTPUT_FILE', 'r') as f:
        data = json.load(f)
    
    # Группируем по типам
    currencies = []
    metals = []
    stocks = []
    crypto = []
    
    for symbol, info in data.items():
        if 'values' in info and info['values']:
            close_price = info['values'][0]['close']
            symbol_type = info['meta'].get('type', 'Unknown')
            
            if 'Currency' in symbol_type:
                if 'Physical' in symbol_type:
                    currencies.append((symbol, close_price))
                elif 'Digital' in symbol_type:
                    crypto.append((symbol, close_price))
            elif 'Metal' in symbol_type:
                metals.append((symbol, close_price))
            elif 'Stock' in symbol_type:
                stocks.append((symbol, close_price))
    
    print('💱 Валютные пары:')
    for symbol, price in currencies:
        print(f'  {symbol}: {price}')
    
    print('')
    print('🥇 Драгоценные металлы:')
    for symbol, price in metals:
        print(f'  {symbol}: \${price}')
    
    print('')
    print('📈 Акции:')
    for symbol, price in stocks:
        print(f'  {symbol}: \${price}')
    
    print('')
    print('₿ Криптовалюты:')
    for symbol, price in crypto:
        print(f'  {symbol}: \${price}')
    
    print('')
    print(f'📊 Всего получено данных: {len(data)} символов')
    
except Exception as e:
    print(f'Ошибка обработки данных: {e}')
    print('Сырой ответ:')
    print('$response')
" 2>/dev/null

else
    echo "❌ Ошибка получения данных:"
    echo "$response"
    
    if echo "$response" | grep -q "429"; then
        echo ""
        echo "💡 Превышен лимит запросов. Попробуйте позже."
    elif echo "$response" | grep -q "400"; then
        echo ""
        echo "💡 Ошибка в запросе. Возможно, слишком много символов."
    fi
fi

echo ""
echo "🎉 Готово! Все данные получены одним запросом."
echo ""
echo "💡 Преимущества этого подхода:"
echo "  - Один запрос вместо множества"
echo "  - Экономия API лимитов"
echo "  - Быстрее выполнение"
echo "  - Атомарность данных (все на один момент времени)"

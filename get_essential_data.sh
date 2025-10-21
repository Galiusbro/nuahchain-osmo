#!/bin/bash

# Скрипт для получения 8 самых важных символов одним запросом
# Оптимизирован для бесплатного плана TwelveData (8 запросов/минуту)

API_KEY="01124156c61844b386a6f017d1836e0c"
BASE_URL="https://api.twelvedata.com"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
OUTPUT_FILE="essential_market_data_${TIMESTAMP}.json"

echo "🚀 Получение 8 самых важных символов одним запросом"
echo "================================================="

# 8 самых важных символов (в пределах лимита бесплатного плана)
SYMBOLS="EUR/USD,GBP/USD,USD/JPY,USD/RUB,AAPL,GOOGL,BTC/USD,XAU/USD"

echo "📊 Символы для получения:"
echo "  💱 EUR/USD - Евро/Доллар"
echo "  💱 GBP/USD - Фунт/Доллар" 
echo "  💱 USD/JPY - Доллар/Йена"
echo "  💱 USD/RUB - Доллар/Рубль"
echo "  📈 AAPL - Apple"
echo "  📈 GOOGL - Google"
echo "  ₿ BTC/USD - Bitcoin"
echo "  🥇 XAU/USD - Золото"
echo ""

echo "🔄 Выполнение запроса..."

# Делаем один запрос для 8 символов
response=$(curl -s "${BASE_URL}/time_series?symbol=${SYMBOLS}&interval=1day&apikey=${API_KEY}&outputsize=1")

# Проверяем ответ
if echo "$response" | grep -q '"status":"ok"'; then
    echo "✅ Данные получены успешно одним запросом!"
    echo ""
    
    # Сохраняем в файл
    echo "$response" | python3 -m json.tool > "$OUTPUT_FILE"
    
    echo "💾 Данные сохранены в файл: $OUTPUT_FILE"
    echo ""
    
    # Показываем результаты
    echo "📋 Полученные данные:"
    echo "===================="
    
    python3 -c "
import json
import sys

try:
    with open('$OUTPUT_FILE', 'r') as f:
        data = json.load(f)
    
    print('💱 Валютные пары:')
    for symbol in ['EUR/USD', 'GBP/USD', 'USD/JPY', 'USD/RUB']:
        if symbol in data and 'values' in data[symbol] and data[symbol]['values']:
            close_price = data[symbol]['values'][0]['close']
            datetime = data[symbol]['values'][0]['datetime']
            print(f'  {symbol}: {close_price} (время: {datetime})')
    
    print('')
    print('📈 Акции:')
    for symbol in ['AAPL', 'GOOGL']:
        if symbol in data and 'values' in data[symbol] and data[symbol]['values']:
            close_price = data[symbol]['values'][0]['close']
            datetime = data[symbol]['values'][0]['datetime']
            volume = data[symbol]['values'][0].get('volume', 'N/A')
            print(f'  {symbol}: \${close_price} (объем: {volume}, время: {datetime})')
    
    print('')
    print('₿ Криптовалюты:')
    if 'BTC/USD' in data and 'values' in data['BTC/USD'] and data['BTC/USD']['values']:
        close_price = data['BTC/USD']['values'][0]['close']
        datetime = data['BTC/USD']['values'][0]['datetime']
        print(f'  BTC/USD: \${close_price} (время: {datetime})')
    
    print('')
    print('🥇 Драгоценные металлы:')
    if 'XAU/USD' in data and 'values' in data['XAU/USD'] and data['XAU/USD']['values']:
        close_price = data['XAU/USD']['values'][0]['close']
        datetime = data['XAU/USD']['values'][0]['datetime']
        print(f'  XAU/USD: \${close_price} (время: {datetime})')
    
    print('')
    print(f'📊 Всего получено: {len(data)} символов')
    print(f'⏱️  Время выполнения: $(date)')
    
except Exception as e:
    print(f'Ошибка обработки данных: {e}')
" 2>/dev/null

else
    echo "❌ Ошибка получения данных:"
    echo "$response"
    
    if echo "$response" | grep -q "429"; then
        echo ""
        echo "💡 Превышен лимит запросов. Подождите минуту и попробуйте снова."
    elif echo "$response" | grep -q "400"; then
        echo ""
        echo "💡 Ошибка в запросе. Проверьте символы."
    fi
fi

echo ""
echo "🎉 Готово! 8 самых важных символов получены одним запросом."
echo ""
echo "💡 Преимущества этого подхода:"
echo "  ✅ Один запрос вместо 8 отдельных"
echo "  ✅ Экономия API лимитов (используется только 1 из 8 запросов/минуту)"
echo "  ✅ Быстрое выполнение"
echo "  ✅ Атомарность данных (все на один момент времени)"
echo "  ✅ Покрывает основные категории: валюты, акции, крипто, золото"
echo ""
echo "🔄 Для получения большего количества данных:"
echo "  - Подождите минуту и запустите скрипт снова"
echo "  - Или рассмотрите платный план TwelveData"

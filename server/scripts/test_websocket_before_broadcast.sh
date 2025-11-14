#!/bin/bash

# Тест WebSocket для транзакций, которые ещё не включены в блок
# Стратегия: подписаться ДО broadcast, затем отправить транзакцию

echo "=== Тест WebSocket для транзакций ДО включения в блок ==="
echo ""

# 1. Получаем токен
echo "1. Получение токена..."
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"garold@gmail.com","password":"Garold632"}' \
  | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "   ❌ Не удалось получить токен"
  exit 1
fi
echo "   ✅ Токен получен"

# 2. Очищаем старые логи
echo ""
echo "2. Очистка старых WebSocket логов..."
# Сохраняем только последние 50 строк
tail -50 /tmp/server.log > /tmp/server.log.backup 2>/dev/null || true

# 3. Отправляем транзакцию и сразу смотрим логи
echo ""
echo "3. Отправка транзакции (подписка должна произойти ДО broadcast)..."
echo "   Внимание: смотрим логи в реальном времени"

# Отправляем транзакцию в фоне и сразу мониторим логи
(
  curl -s -X POST "http://localhost:8080/api/assets/ensure" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{"symbol":"GOLD"}' > /dev/null
) &

TX_PID=$!

# Мониторим логи в реальном времени
echo "   Мониторинг логов (10 секунд)..."
for i in {1..10}; do
  sleep 1
  # Ищем новые записи о подписке и событиях
  tail -100 /tmp/server.log | grep -E "\[WebSocket\].*Subscribing|\[WebSocket\].*tm_event|\[WebSocket\].*Parsed tx event" | tail -3
done

wait $TX_PID

# 4. Получаем хеш транзакции из ответа API
echo ""
echo "4. Получение хеша транзакции..."
TX_HASH=$(curl -s -X POST "http://localhost:8080/api/assets/ensure" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"symbol":"GOLD"}' \
  | grep -o '"tx_hash":"[^"]*' | cut -d'"' -f4)

if [ -z "$TX_HASH" ]; then
  echo "   ❌ Не удалось получить хеш транзакции"
  exit 1
fi

echo "   ✅ Tx Hash: $TX_HASH"

# 5. Анализируем логи
echo ""
echo "5. Анализ логов для этой транзакции..."
echo "   Ищем подписку и события:"

# Ищем все записи, связанные с этим хешем
tail -500 /tmp/server.log | grep -E "\[WebSocket\].*$TX_HASH|Subscribing.*$TX_HASH|tm_event|Parsed tx event.*$TX_HASH" | tail -20

echo ""
echo "6. Проверка временной последовательности:"
echo "   Когда произошла подписка относительно broadcast?"

# Извлекаем временные метки
SUBS_TIME=$(tail -500 /tmp/server.log | grep "Subscribing.*$TX_HASH" | tail -1 | grep -o "\[.*\]" | head -1)
EVENT_TIME=$(tail -500 /tmp/server.log | grep "Parsed tx event.*$TX_HASH" | tail -1 | grep -o "\[.*\]" | head -1)

if [ -n "$SUBS_TIME" ]; then
  echo "   Подписка: $SUBS_TIME"
else
  echo "   Подписка: не найдена"
fi

if [ -n "$EVENT_TIME" ]; then
  echo "   Событие: $EVENT_TIME"
  echo "   ✅ Событие получено через WebSocket!"
else
  echo "   Событие: не найдено"
  echo "   ⚠️  Событие не пришло через WebSocket (возможно, транзакция уже в блоке)"
fi

echo ""
echo "=== Выводы ==="
echo "Если событие получено - WebSocket работает для будущих транзакций"
echo "Если событие не получено - транзакция была включена слишком быстро"


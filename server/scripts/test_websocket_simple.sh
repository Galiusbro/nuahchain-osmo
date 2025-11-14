#!/bin/bash

# Простой тест WebSocket подписки
# Цель: понять, работает ли вообще WebSocket и приходят ли события

echo "=== Тест WebSocket подписки ==="
echo ""

# 1. Проверяем, что WebSocket endpoint доступен
echo "1. Проверка WebSocket endpoint..."
curl -s "http://localhost:26657/status" > /dev/null && echo "   ✅ RPC доступен" || echo "   ❌ RPC недоступен"

# 2. Смотрим текущие логи сервера
echo ""
echo "2. Текущее состояние WebSocket в логах:"
tail -5 /tmp/server.log | grep -i "websocket" || echo "   (нет записей)"

# 3. Отправляем транзакцию и сразу подписываемся
echo ""
echo "3. Отправка транзакции и подписка..."
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"garold@gmail.com","password":"Garold632"}' \
  | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "   ❌ Не удалось получить токен"
  exit 1
fi

echo "   ✅ Токен получен"

# Отправляем транзакцию
TX_HASH=$(curl -s -X POST "http://localhost:8080/api/assets/ensure" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"symbol":"GOLD"}' \
  | grep -o '"tx_hash":"[^"]*' | cut -d'"' -f4)

if [ -z "$TX_HASH" ]; then
  echo "   ❌ Не удалось отправить транзакцию"
  exit 1
fi

echo "   ✅ Транзакция отправлена: $TX_HASH"

# 4. Ждём и смотрим логи
echo ""
echo "4. Ожидание событий (10 секунд)..."
sleep 10

echo ""
echo "5. Логи WebSocket за последние 20 секунд:"
tail -200 /tmp/server.log | grep -E "\[WebSocket\]|tm_event|subscribe" | tail -20

echo ""
echo "6. Проверка статуса транзакции на блокчейне:"
./build/nuahd query tx "$TX_HASH" --chain-id nuahchain 2>&1 | head -5

echo ""
echo "=== Ключевые вопросы ==="
echo "1. Пришло ли событие tm_event?"
echo "2. Правильно ли мы подписываемся?"
echo "3. Может быть транзакция уже в блоке до подписки?"


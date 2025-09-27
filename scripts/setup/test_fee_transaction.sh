#!/bin/bash

# Скрипт для тестирования транзакций с комиссией в ndollar

set -e

CHAIN_ID="${CHAIN_ID:-nuahchain}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"

echo "🧪 Тестирование транзакций с комиссией в ndollar..."

# Проверка что узел запущен
if ! pgrep -f "nuahd start" > /dev/null; then
    echo "❌ Узел не запущен. Запустите узел командой: ./build/nuahd start"
    exit 1
fi

# Получение адресов
ALICE_ADDR=$(./build/nuahd keys show alice -a --keyring-backend $KEYRING_BACKEND 2>/dev/null || echo "")
BOB_ADDR=$(./build/nuahd keys show bob -a --keyring-backend $KEYRING_BACKEND 2>/dev/null || echo "")

if [ -z "$ALICE_ADDR" ] || [ -z "$BOB_ADDR" ]; then
    echo "❌ Аккаунты alice или bob не найдены. Запустите сначала init_fresh_node.sh"
    exit 1
fi

echo "Alice address: $ALICE_ADDR"
echo "Bob address: $BOB_ADDR"

# Проверка балансов
echo ""
echo "💰 Проверка балансов..."
echo "Alice balance:"
./build/nuahd query bank balances $ALICE_ADDR

echo ""
echo "Bob balance:"
./build/nuahd query bank balances $BOB_ADDR

# Проверка допустимых токенов комиссии
echo ""
echo "🔍 Проверка допустимых токенов комиссии..."
./build/nuahd query txfees fee-tokens

# Тест транзакции с комиссией в ndollar
echo ""
echo "🚀 Тестирование транзакции с комиссией 0.01 ndollar..."

# Получаем полное имя токена ndollar
NDOLLAR_DENOM=$(./build/nuahd query txfees fee-tokens | grep "denom:" | awk '{print $2}')

if [ -z "$NDOLLAR_DENOM" ]; then
    echo "❌ Токен ndollar не найден в списке допустимых токенов комиссии"
    exit 1
fi

echo "Используем токен комиссии: $NDOLLAR_DENOM"

# Выполняем dry-run транзакции
echo "Выполняем dry-run транзакции..."
./build/nuahd tx bank send $ALICE_ADDR $BOB_ADDR 1unuah \
    --fees 0.01$NDOLLAR_DENOM \
    --chain-id $CHAIN_ID \
    --keyring-backend $KEYRING_BACKEND \
    --dry-run

echo ""
echo "✅ Dry-run успешно выполнен!"

# Выполняем реальную транзакцию
echo ""
echo "Выполняем реальную транзакцию..."
./build/nuahd tx bank send $ALICE_ADDR $BOB_ADDR 1unuah \
    --fees 0.01$NDOLLAR_DENOM \
    --chain-id $CHAIN_ID \
    --keyring-backend $KEYRING_BACKEND \
    --yes

echo ""
echo "✅ Транзакция отправлена!"

# Ждем несколько секунд для обработки
echo "Ожидание обработки транзакции..."
sleep 5

# Проверяем балансы после транзакции
echo ""
echo "💰 Проверка балансов после транзакции..."
echo "Alice balance:"
./build/nuahd query bank balances $ALICE_ADDR

echo ""
echo "Bob balance:"
./build/nuahd query bank balances $BOB_ADDR

echo ""
echo "🎉 Тест завершен!"

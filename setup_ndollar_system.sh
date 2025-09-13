#!/bin/bash

# Скрипт автоматической настройки системы N$ токена
# Этот скрипт создает N$ токен и настраивает параметры pegkeeper и usdoracle модулей

set -e

echo "🚀 Начинаем настройку системы N$ токена..."

# Конфигурация
CHAIN_ID="nuahchain-1"
KEYRING_BACKEND="test"
VALIDATOR_KEY="validator"
GAS="auto"
GAS_ADJUSTMENT="1.5"
FEES="5000unuah"

# Проверяем, что узел запущен
echo "📡 Проверяем статус узла..."
if ! ./build/nuahd status &>/dev/null; then
    echo "❌ Ошибка: узел не запущен. Запустите узел командой: ./build/nuahd start"
    exit 1
fi

echo "✅ Узел работает"

# Получаем адрес validator ключа
VALIDATOR_ADDRESS=$(./build/nuahd keys show $VALIDATOR_KEY --keyring-backend $KEYRING_BACKEND --address)
echo "🔑 Адрес validator: $VALIDATOR_ADDRESS"

# Шаг 1: Создание N$ токена через tokenfactory
echo "\n💰 Шаг 1: Создание N$ токена..."
NDOLLAR_DENOM="factory/$VALIDATOR_ADDRESS/ndollar"

echo "Создаем токен N$ с деноминацией: $NDOLLAR_DENOM"
./build/nuahd tx tokenfactory create-denom ndollar \
    --from $VALIDATOR_KEY \
    --keyring-backend $KEYRING_BACKEND \
    --chain-id $CHAIN_ID \
    --gas $GAS \
    --gas-adjustment $GAS_ADJUSTMENT \
    --fees $FEES \
    --yes

echo "⏳ Ждем подтверждения транзакции..."
sleep 6

# Проверяем создание токена
echo "🔍 Проверяем создание токена..."
if ./build/nuahd query tokenfactory denoms-from-creator $VALIDATOR_ADDRESS | grep -q "ndollar"; then
    echo "✅ Токен N$ успешно создан: $NDOLLAR_DENOM"
else
    echo "❌ Ошибка: токен N$ не был создан"
    exit 1
fi

# Шаг 2: Настройка параметров USDOracle модуля
echo "\n🔮 Шаг 2: Настройка параметров USDOracle модуля..."
./build/nuahd tx usdoracle update-params \
    "true" \
    "$VALIDATOR_ADDRESS" \
    "3600" \
    "0.05" \
    --from $VALIDATOR_KEY \
    --keyring-backend $KEYRING_BACKEND \
    --chain-id $CHAIN_ID \
    --gas $GAS \
    --gas-adjustment $GAS_ADJUSTMENT \
    --fees $FEES \
    --yes

echo "⏳ Ждем подтверждения транзакции..."
sleep 6

echo "✅ Параметры USDOracle обновлены"

# Шаг 3: Настройка параметров PegKeeper модуля
echo "\n⚖️ Шаг 3: Настройка параметров PegKeeper модуля..."
./build/nuahd tx pegkeeper update-params \
    "$NDOLLAR_DENOM" \
    "usd" \
    "0.05" \
    "0.1" \
    "3600" \
    "0.02" \
    "usdoracle" \
    "true" \
    "1.0" \
    --from $VALIDATOR_KEY \
    --keyring-backend $KEYRING_BACKEND \
    --chain-id $CHAIN_ID \
    --gas $GAS \
    --gas-adjustment $GAS_ADJUSTMENT \
    --fees $FEES \
    --yes

echo "⏳ Ждем подтверждения транзакции..."
sleep 6

echo "✅ Параметры PegKeeper обновлены"

# Шаг 4: Проверка конфигурации
echo "\n🔍 Шаг 4: Проверка конфигурации системы..."

echo "📊 Параметры USDOracle:"
./build/nuahd query usdoracle params

echo "\n📊 Параметры PegKeeper:"
./build/nuahd query pegkeeper params

echo "\n📊 Информация о токене N$:"
./build/nuahd query tokenfactory denoms-from-creator $VALIDATOR_ADDRESS

echo "\n🎉 Настройка системы N$ токена завершена успешно!"
echo "\n📝 Сводка:"
echo "   • Токен N$: $NDOLLAR_DENOM"
echo "   • USDOracle: включен, интервал обновления 1 час, порог отклонения 5%"
echo "   • PegKeeper: включен, целевая цена 1.0 USD, максимальное отклонение 5%"
echo "   • Validator адрес: $VALIDATOR_ADDRESS"
echo "\n✨ Система готова к использованию!"

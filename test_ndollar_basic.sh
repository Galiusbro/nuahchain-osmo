#!/bin/bash

echo "🧪 Тестирование базовой функциональности N$ токена..."

# Получаем адрес validator
VALIDATOR_ADDR=$(./build/nuahd keys show validator --keyring-backend test -a)
echo "🔑 Адрес validator: $VALIDATOR_ADDR"

# Определяем N$ токен
NDOLLAR_DENOM="factory/$VALIDATOR_ADDR/ndollar"
echo "💰 N$ токен: $NDOLLAR_DENOM"

echo "\n🔍 Тест 1: Проверка баланса N$ токена..."
NDOLLAR_BALANCE=$(./build/nuahd query bank balance $VALIDATOR_ADDR $NDOLLAR_DENOM --output json 2>/dev/null | jq -r '.amount // "0"')
echo "💰 Текущий баланс N$: $NDOLLAR_BALANCE"

echo "\n🔍 Тест 2: Проверка метаданных токена..."
./build/nuahd query tokenfactory denom-metadata $NDOLLAR_DENOM --output json 2>/dev/null || echo "❌ Метаданные не найдены"

echo "\n🔍 Тест 3: Попытка mint N$ токенов..."
echo "Пытаемся создать 1000 N$ токенов..."
MINT_RESULT=$(./build/nuahd tx tokenfactory mint "1000$NDOLLAR_DENOM" $VALIDATOR_ADDR --from validator --keyring-backend test --chain-id nuahchain --yes --gas auto --gas-adjustment 1.5 2>&1)
if echo "$MINT_RESULT" | grep -q "txhash"; then
    echo "✅ Mint транзакция отправлена успешно"
    TXHASH=$(echo "$MINT_RESULT" | grep "txhash" | cut -d':' -f2 | tr -d ' ')
    echo "📝 TX Hash: $TXHASH"

    echo "⏳ Ждем подтверждения транзакции..."
    sleep 6

    echo "\n💰 Новый баланс N$ токена:"
    NEW_BALANCE=$(./build/nuahd query bank balance $VALIDATOR_ADDR $NDOLLAR_DENOM --output json 2>/dev/null | jq -r '.amount // "0"')
    echo "💰 Баланс после mint: $NEW_BALANCE"

    if [ "$NEW_BALANCE" -gt "$NDOLLAR_BALANCE" ]; then
        echo "✅ Mint операция прошла успешно!"
        MINTED_AMOUNT=$((NEW_BALANCE - NDOLLAR_BALANCE))
        echo "📈 Создано токенов: $MINTED_AMOUNT"
    else
        echo "❌ Mint операция не изменила баланс"
    fi
else
    echo "❌ Ошибка при mint операции:"
    echo "$MINT_RESULT"
fi

echo "\n🔍 Тест 4: Проверка параметров USDOracle..."
echo "📊 Текущие параметры USDOracle:"
./build/nuahd query usdoracle params --output json 2>/dev/null || echo "❌ Не удалось получить параметры USDOracle"

echo "\n🔍 Тест 5: Проверка параметров PegKeeper..."
echo "📊 Текущие параметры PegKeeper:"
./build/nuahd query pegkeeper params --output json 2>/dev/null || echo "❌ Не удалось получить параметры PegKeeper"

echo "\n🔍 Тест 6: Проверка доступных команд pegkeeper..."
echo "📋 Доступные команды pegkeeper:"
./build/nuahd query pegkeeper --help 2>/dev/null | grep -A 20 "Available Commands:" || echo "❌ Команды pegkeeper недоступны"

echo "\n✅ Тестирование завершено!"

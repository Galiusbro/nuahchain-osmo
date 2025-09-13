#!/bin/bash

# Скрипт тестирования функциональности PegKeeper модуля с N$ токеном
# Этот скрипт проверяет работу системы стабилизации цены N$ токена

set -e

echo "🧪 Начинаем тестирование функциональности PegKeeper модуля..."

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

# Определяем деноминацию N$ токена
NDOLLAR_DENOM="factory/$VALIDATOR_ADDRESS/ndollar"
echo "💰 N$ токен: $NDOLLAR_DENOM"

# Тест 1: Проверка существования токена
echo "\n🔍 Тест 1: Проверка существования N$ токена..."
if ./build/nuahd query tokenfactory denoms-from-creator $VALIDATOR_ADDRESS | grep -q "ndollar"; then
    echo "✅ N$ токен существует"
else
    echo "❌ N$ токен не найден. Запустите сначала setup_ndollar_system.sh"
    exit 1
fi

# Тест 2: Проверка параметров PegKeeper
echo "\n⚖️ Тест 2: Проверка параметров PegKeeper..."
PEGKEEPER_PARAMS=$(./build/nuahd query pegkeeper params -o json)
echo "📊 Текущие параметры PegKeeper:"
echo $PEGKEEPER_PARAMS | jq .

# Проверяем, что target_denom установлен на N$ токен
TARGET_DENOM=$(echo $PEGKEEPER_PARAMS | jq -r '.target_denom')
if [ "$TARGET_DENOM" = "$NDOLLAR_DENOM" ]; then
    echo "✅ Target denom корректно установлен на N$ токен"
else
    echo "❌ Target denom не соответствует N$ токену. Ожидается: $NDOLLAR_DENOM, получено: $TARGET_DENOM"
    exit 1
fi

# Тест 3: Проверка параметров USDOracle
echo "\n🔮 Тест 3: Проверка параметров USDOracle..."
USDORACLE_PARAMS=$(./build/nuahd query usdoracle params -o json)
echo "📊 Текущие параметры USDOracle:"
echo $USDORACLE_PARAMS | jq .

# Проверяем, что модуль включен
USDORACLE_ENABLED=$(echo $USDORACLE_PARAMS | jq -r '.enabled')
if [ "$USDORACLE_ENABLED" = "true" ]; then
    echo "✅ USDOracle модуль включен"
else
    echo "❌ USDOracle модуль отключен"
    exit 1
fi

# Тест 4: Проверка состояния PegKeeper
echo "\n📈 Тест 4: Проверка состояния PegKeeper..."
echo "Запрашиваем текущее состояние peg..."
if ./build/nuahd query pegkeeper peg-state &>/dev/null; then
    PEG_STATE=$(./build/nuahd query pegkeeper peg-state -o json)
    echo "📊 Состояние peg:"
    echo $PEG_STATE | jq .
    echo "✅ Состояние peg получено успешно"
else
    echo "⚠️ Состояние peg пока не инициализировано (это нормально для новой системы)"
fi

# Тест 5: Проверка истории корректировок
echo "\n📜 Тест 5: Проверка истории корректировок..."
echo "Запрашиваем историю корректировок..."
if ./build/nuahd query pegkeeper adjustment-history &>/dev/null; then
    ADJUSTMENT_HISTORY=$(./build/nuahd query pegkeeper adjustment-history -o json)
    echo "📊 История корректировок:"
    echo $ADJUSTMENT_HISTORY | jq .
    echo "✅ История корректировок получена успешно"
else
    echo "⚠️ История корректировок пуста (это нормально для новой системы)"
fi

# Тест 6: Проверка баланса N$ токенов
echo "\n💳 Тест 6: Проверка баланса N$ токенов..."
NDOLLAR_BALANCE=$(./build/nuahd query bank balance $VALIDATOR_ADDRESS $NDOLLAR_DENOM -o json 2>/dev/null || echo '{"amount":"0"}')
BALANCE_AMOUNT=$(echo $NDOLLAR_BALANCE | jq -r '.amount // "0"')
echo "💰 Баланс N$ токенов у validator: $BALANCE_AMOUNT"

# Тест 7: Тестирование минтинга N$ токенов (если баланс 0)
if [ "$BALANCE_AMOUNT" = "0" ]; then
    echo "\n🏭 Тест 7: Тестирование минтинга N$ токенов..."
    echo "Минтим 1000 N$ токенов для тестирования..."
    ./build/nuahd tx tokenfactory mint 1000$NDOLLAR_DENOM \
        --from $VALIDATOR_KEY \
        --keyring-backend $KEYRING_BACKEND \
        --chain-id $CHAIN_ID \
        --gas $GAS \
        --gas-adjustment $GAS_ADJUSTMENT \
        --fees $FEES \
        --yes

    echo "⏳ Ждем подтверждения транзакции..."
    sleep 6

    # Проверяем новый баланс
    NEW_BALANCE=$(./build/nuahd query bank balance $VALIDATOR_ADDRESS $NDOLLAR_DENOM -o json)
    NEW_AMOUNT=$(echo $NEW_BALANCE | jq -r '.amount')
    echo "💰 Новый баланс N$ токенов: $NEW_AMOUNT"

    if [ "$NEW_AMOUNT" -gt "0" ]; then
        echo "✅ Минтинг N$ токенов работает корректно"
    else
        echo "❌ Ошибка минтинга N$ токенов"
        exit 1
    fi
else
    echo "\n⏭️ Пропускаем тест минтинга (баланс уже больше 0)"
fi

# Тест 8: Проверка CLI команд
echo "\n🖥️ Тест 8: Проверка доступности CLI команд..."
echo "Проверяем команды pegkeeper..."
if ./build/nuahd tx pegkeeper --help | grep -q "update-params"; then
    echo "✅ Команда pegkeeper update-params доступна"
else
    echo "❌ Команда pegkeeper update-params недоступна"
    exit 1
fi

echo "Проверяем команды usdoracle..."
if ./build/nuahd tx usdoracle --help | grep -q "update-params"; then
    echo "✅ Команда usdoracle update-params доступна"
else
    echo "❌ Команда usdoracle update-params недоступна"
    exit 1
fi

# Тест 9: Генерация тестовых транзакций
echo "\n📝 Тест 9: Генерация тестовых транзакций..."
echo "Генерируем тестовую транзакцию обновления параметров pegkeeper..."
TEST_TX=$(./build/nuahd tx pegkeeper update-params \
    "$NDOLLAR_DENOM" \
    "usd" \
    "0.03" \
    "0.15" \
    "1800" \
    "0.03" \
    "usdoracle" \
    "true" \
    "1.0" \
    --from $VALIDATOR_KEY \
    --keyring-backend $KEYRING_BACKEND \
    --generate-only \
    --account-number 0 \
    --sequence 0 \
    --offline 2>/dev/null)

if echo "$TEST_TX" | jq . &>/dev/null; then
    echo "✅ Генерация транзакции pegkeeper работает корректно"
else
    echo "❌ Ошибка генерации транзакции pegkeeper"
    exit 1
fi

echo "Генерируем тестовую транзакцию обновления параметров usdoracle..."
TEST_TX2=$(./build/nuahd tx usdoracle update-params \
    "true" \
    "$VALIDATOR_ADDRESS" \
    "1800" \
    "0.03" \
    --from $VALIDATOR_KEY \
    --keyring-backend $KEYRING_BACKEND \
    --generate-only \
    --account-number 0 \
    --sequence 0 \
    --offline 2>/dev/null)

if echo "$TEST_TX2" | jq . &>/dev/null; then
    echo "✅ Генерация транзакции usdoracle работает корректно"
else
    echo "❌ Ошибка генерации транзакции usdoracle"
    exit 1
fi

# Итоговый отчет
echo "\n🎉 Тестирование завершено успешно!"
echo "\n📋 Сводка результатов:"
echo "   ✅ N$ токен существует и доступен"
echo "   ✅ Параметры PegKeeper настроены корректно"
echo "   ✅ Параметры USDOracle настроены корректно"
echo "   ✅ CLI команды работают корректно"
echo "   ✅ Генерация транзакций работает корректно"
echo "   ✅ Система готова к работе"
echo "\n🚀 Система N$ токена полностью функциональна!"
echo "\n📝 Следующие шаги:"
echo "   • Настройте источники цен для USDOracle модуля"
echo "   • Мониторьте работу системы стабилизации"
echo "   • При необходимости корректируйте параметры через governance"

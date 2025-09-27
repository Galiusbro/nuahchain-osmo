#!/bin/bash

# Скрипт для полной инициализации узла с нуля
# Удаляет все данные, создает новые ключи и настраивает генезис
# Подготавливает ноду для работы с ndollar токенами

set -e

# Цветовые коды для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Конфигурация (с поддержкой переменных окружения)
CHAIN_ID="${CHAIN_ID:-nuahchain}"
MONIKER="${MONIKER:-test-node}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
GENESIS_FILE="$HOME/.nuahd/config/genesis.json"

# Функции для вывода
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_step() {
    echo -e "${BLUE}🔄 $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

print_header() {
    echo -e "${PURPLE}$1${NC}"
}

# Проверка наличия бинарного файла
check_binary() {
    if [ ! -f "./build/nuahd" ]; then
        print_error "nuahd binary not found in ./build/"
        print_info "Please build the binary first with: make build"
        exit 1
    fi
    print_status "nuahd binary found"
}

# Проверка наличия jq
check_jq() {
    if ! command -v jq &> /dev/null; then
        print_error "jq is required but not installed"
        print_info "Please install jq: brew install jq"
        exit 1
    fi
    print_status "jq found"
}

print_header "🚀 Инициализация свежей ноды Nuah Chain"
print_header "======================================"
echo ""

# Предварительные проверки
check_binary
check_jq

print_step "🧹 Очистка существующих данных..."

# Остановка узла если запущен
print_info "Остановка узла если запущен..."
pkill nuahd || true
sleep 2

# Удаление всех ключей
print_info "Удаление всех ключей..."
./build/nuahd keys list --keyring-backend $KEYRING_BACKEND 2>/dev/null | grep -E "^- name:" | awk '{print $3}' | xargs -I {} ./build/nuahd keys delete {} --yes --keyring-backend $KEYRING_BACKEND 2>/dev/null || true

# Удаление данных узла
print_info "Удаление данных узла..."
rm -rf ~/.nuahd

print_status "Очистка завершена"

print_step "🔧 Инициализация нового узла..."

# Инициализация узла
./build/nuahd init $MONIKER --chain-id $CHAIN_ID
print_status "Узел инициализирован"

print_step "🔑 Создание ключей..."

# Создание ключей для тестирования
./build/nuahd keys add validator --keyring-backend $KEYRING_BACKEND
./build/nuahd keys add alice --keyring-backend $KEYRING_BACKEND
./build/nuahd keys add bob --keyring-backend $KEYRING_BACKEND

# Получение адресов
VALIDATOR_ADDR=$(./build/nuahd keys show validator -a --keyring-backend $KEYRING_BACKEND)
ALICE_ADDR=$(./build/nuahd keys show alice -a --keyring-backend $KEYRING_BACKEND)
BOB_ADDR=$(./build/nuahd keys show bob -a --keyring-backend $KEYRING_BACKEND)

print_status "Ключи созданы:"
echo "  Validator: $VALIDATOR_ADDR"
echo "  Alice: $ALICE_ADDR"
echo "  Bob: $BOB_ADDR"

print_step "💰 Настройка генезиса с начальными балансами..."

# В development режиме добавляем unuah токены (это наш нативный токен для стейкинга)
# В production режиме добавляем unuah токены
if [ "${ENVIRONMENT:-development}" = "production" ]; then
    print_info "Production режим: добавляем unuah токены..."
    ./build/nuahd add-genesis-account $VALIDATOR_ADDR 10000000000unuah
    ./build/nuahd add-genesis-account $ALICE_ADDR 1000000000unuah
    ./build/nuahd add-genesis-account $BOB_ADDR 1000000000unuah
else
    print_info "Development режим: добавляем unuah токены..."
    ./build/nuahd add-genesis-account $VALIDATOR_ADDR 10000000000unuah
    ./build/nuahd add-genesis-account $ALICE_ADDR 1000000000unuah  
    ./build/nuahd add-genesis-account $BOB_ADDR 1000000000unuah
fi

GENTX_AMOUNT="1000000000unuah"

print_status "Аккаунты добавлены в генезис"

# Создание gentx для валидатора
print_info "Создание gentx для валидатора..."
./build/nuahd gentx validator $GENTX_AMOUNT --chain-id $CHAIN_ID --keyring-backend $KEYRING_BACKEND --from validator

# Сбор gentx
print_info "Сбор gentx..."
./build/nuahd collect-gentxs

print_status "Gentx настроен"

print_step "🔧 Настройка параметров генезиса..."

# Замена stake на unuah в genesis.json
print_info "Замена 'stake' на 'unuah' в genesis.json..."
sed -i '' 's/"stake"/"unuah"/g' $GENESIS_FILE

# Добавление валидатора в whitelisted_fee_token_setters
print_info "Добавление валидатора в whitelisted_fee_token_setters..."
jq --arg validator "$VALIDATOR_ADDR" '.app_state.txfees.params.whitelisted_fee_token_setters = [$validator]' $GENESIS_FILE > /tmp/genesis_temp.json && mv /tmp/genesis_temp.json $GENESIS_FILE

# Настройка параметров для тестирования
print_info "Настройка параметров для тестирования..."

# Уменьшение времени блока для быстрого тестирования
jq '.consensus_params.block.time_iota_ms = "1000"' $GENESIS_FILE > /tmp/genesis_temp.json && mv /tmp/genesis_temp.json $GENESIS_FILE

# Настройка параметров токенфабрики
jq '.app_state.tokenfactory.params.denom_creation_fee = [{"denom": "unuah", "amount": "1000000"}]' $GENESIS_FILE > /tmp/genesis_temp.json && mv /tmp/genesis_temp.json $GENESIS_FILE

print_status "Параметры генезиса настроены"

# Проверка конфигурации
print_step "🔍 Проверка конфигурации..."

print_info "Проверка базовой деноминации..."
BASE_DENOM=$(jq -r '.app_state.staking.params.bond_denom' $GENESIS_FILE)
if [ "$BASE_DENOM" = "unuah" ]; then
    print_status "Базовая деноминация: $BASE_DENOM ✓"
else
    print_error "Неправильная базовая деноминация: $BASE_DENOM"
    exit 1
fi

print_info "Проверка whitelisted_fee_token_setters..."
WHITELIST_COUNT=$(jq '.app_state.txfees.params.whitelisted_fee_token_setters | length' $GENESIS_FILE)
if [ "$WHITELIST_COUNT" -gt 0 ]; then
    print_status "Whitelisted fee token setters настроены ✓"
    jq -r '.app_state.txfees.params.whitelisted_fee_token_setters[]' $GENESIS_FILE | while read addr; do
        echo "  - $addr"
    done
else
    print_warning "Whitelisted fee token setters пусты"
fi

# Валидация генезиса (опционально)
if [ "${SKIP_VALIDATION:-false}" != "true" ]; then
    print_info "Валидация генезиса..."
    if ./build/nuahd validate-genesis $GENESIS_FILE; then
        print_status "Генезис валиден ✓"
    else
        print_warning "Валидация генезиса не прошла, но продолжаем..."
    fi
fi

print_header "=================================================="
print_header "✅ Узел успешно инициализирован! ✅"
print_header "=================================================="
echo ""
print_info "Следующие шаги:"
echo "  1. Запустите узел: ./build/nuahd start"
echo "  2. Настройте ndollar: ./scripts/setup/setup_ndollar.sh"
echo ""
print_info "Адреса аккаунтов:"
echo "  Validator: $VALIDATOR_ADDR"
echo "  Alice: $ALICE_ADDR"
echo "  Bob: $BOB_ADDR"
echo ""
print_info "Конфигурация:"
echo "  Chain ID: $CHAIN_ID"
echo "  Keyring Backend: $KEYRING_BACKEND"
echo "  Genesis File: $GENESIS_FILE"
echo ""
print_status "Готово к настройке ndollar токенов! 🎉"

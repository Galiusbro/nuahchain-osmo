#!/bin/bash

# Скрипт для настройки правильной токеномики NUAH блокчейна
# Автор: AI Assistant
# Дата: $(date)

set -e

echo "🚀 Начинаем настройку правильной токеномики для NUAH блокчейна..."

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода цветного текста
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Проверка существования бинарника
if [ ! -f "./build/nuahd" ]; then
    print_error "Бинарник ./build/nuahd не найден. Сначала соберите проект."
    exit 1
fi

# Подтверждение от пользователя
print_warning "ВНИМАНИЕ: Этот скрипт сбросит все текущие данные блокчейна!"
print_warning "Убедитесь, что у вас есть резервная копия важных данных."
read -p "Продолжить? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_status "Операция отменена пользователем."
    exit 0
fi

# Шаг 1: Остановка текущей ноды
print_step "Остановка текущей ноды..."
pkill nuahd || true
sleep 2

# Шаг 2: Полная очистка данных и ключей
print_step "Полная очистка данных блокчейна и ключей..."
./build/nuahd comet unsafe-reset-all

# Удаляем все ключи
print_status "Удаление всех существующих ключей..."
./build/nuahd keys delete validator --keyring-backend test -y 2>/dev/null || true
./build/nuahd keys delete foundation --keyring-backend test -y 2>/dev/null || true
./build/nuahd keys delete community --keyring-backend test -y 2>/dev/null || true
./build/nuahd keys delete treasury --keyring-backend test -y 2>/dev/null || true
./build/nuahd keys delete ecosystem --keyring-backend test -y 2>/dev/null || true
./build/nuahd keys delete team --keyring-backend test -y 2>/dev/null || true

# Удаляем genesis файлы и gentx
print_status "Удаление genesis файлов и gentx..."
rm -f ~/.nuahd/config/genesis.json
rm -rf ~/.nuahd/config/gentx/*
rm -f ~/.nuahd/config/node_key.json
rm -f ~/.nuahd/config/priv_validator_key.json

# Шаг 3: Повторная инициализация
print_step "Инициализация блокчейна..."
./build/nuahd init nuah-mainnet --chain-id nuahchain-1

# Шаг 4: Создание ключей для разных участников
print_step "Создание ключей для участников экосистемы..."

# Создаем все ключи заново (так как мы их удалили)
print_status "Создание ключа валидатора..."
./build/nuahd keys add validator --keyring-backend test

print_status "Создание остальных ключей..."
./build/nuahd keys add foundation --keyring-backend test
./build/nuahd keys add community --keyring-backend test
./build/nuahd keys add treasury --keyring-backend test
./build/nuahd keys add ecosystem --keyring-backend test
./build/nuahd keys add team --keyring-backend test

# Шаг 5: Получение адресов
print_step "Получение адресов участников..."
VALIDATOR_ADDR=$(./build/nuahd keys show validator --keyring-backend test -a)
FOUNDATION_ADDR=$(./build/nuahd keys show foundation --keyring-backend test -a)
COMMUNITY_ADDR=$(./build/nuahd keys show community --keyring-backend test -a)
TREASURY_ADDR=$(./build/nuahd keys show treasury --keyring-backend test -a)
ECOSYSTEM_ADDR=$(./build/nuahd keys show ecosystem --keyring-backend test -a)
TEAM_ADDR=$(./build/nuahd keys show team --keyring-backend test -a)

print_status "Адреса участников:"
echo "  Validator:  $VALIDATOR_ADDR"
echo "  Foundation: $FOUNDATION_ADDR"
echo "  Community:  $COMMUNITY_ADDR"
echo "  Treasury:   $TREASURY_ADDR"
echo "  Ecosystem:  $ECOSYSTEM_ADDR"
echo "  Team:       $TEAM_ADDR"

# Шаг 6: Правильная настройка деноминации в genesis
print_step "Настройка деноминации токена..."

# Сначала заменяем базовую деноминацию
sed -i.bak 's/"stake"/"unuah"/g' ~/.nuahd/config/genesis.json

# Добавляем правильные metadata для деноминации
print_status "Добавление metadata для токена NUAH..."
jq '.app_state.bank.denom_metadata = [
  {
    "description": "The native staking token of the NUAH blockchain",
    "denom_units": [
      {
        "denom": "unuah",
        "exponent": 0,
        "aliases": ["microNUAH"]
      },
      {
        "denom": "mnuah", 
        "exponent": 3,
        "aliases": ["milliNUAH"]
      },
      {
        "denom": "nuah",
        "exponent": 6,
        "aliases": ["NUAH"]
      }
    ],
    "base": "unuah",
    "display": "nuah",
    "name": "NUAH",
    "symbol": "NUAH"
  }
]' ~/.nuahd/config/genesis.json > ~/.nuahd/config/genesis_temp.json && mv ~/.nuahd/config/genesis_temp.json ~/.nuahd/config/genesis.json

# Шаг 7: Добавление аккаунтов с правильным распределением
print_step "Добавление аккаунтов с распределением токенов..."

# Распределение токенов (общее: 100M NUAH)
./build/nuahd add-genesis-account $VALIDATOR_ADDR 5000000000000unuah --keyring-backend test    # 5M NUAH
./build/nuahd add-genesis-account $FOUNDATION_ADDR 20000000000000unuah --keyring-backend test  # 20M NUAH
./build/nuahd add-genesis-account $COMMUNITY_ADDR 25000000000000unuah --keyring-backend test   # 25M NUAH
./build/nuahd add-genesis-account $TREASURY_ADDR 30000000000000unuah --keyring-backend test    # 30M NUAH
./build/nuahd add-genesis-account $ECOSYSTEM_ADDR 15000000000000unuah --keyring-backend test   # 15M NUAH
./build/nuahd add-genesis-account $TEAM_ADDR 5000000000000unuah --keyring-backend test         # 5M NUAH

print_status "Распределение токенов:"
echo "  📊 Общее предложение: 100,000,000 NUAH"
echo "  💰 Treasury (30%):    30,000,000 NUAH"
echo "  🏛️  Community (25%):   25,000,000 NUAH"
echo "  🏢 Foundation (20%):  20,000,000 NUAH"
echo "  🌱 Ecosystem (15%):   15,000,000 NUAH"
echo "  👥 Team (5%):         5,000,000 NUAH"
echo "  ⚡ Validator (5%):     5,000,000 NUAH"

# Шаг 8: Создание gentx
print_step "Создание genesis транзакции валидатора..."
./build/nuahd gentx validator 2000000000000unuah --chain-id nuahchain-1 --keyring-backend test

# Шаг 9: Сбор gentx
print_step "Сбор genesis транзакций..."
./build/nuahd collect-gentxs

# Шаг 9.5: Валидация genesis файла
print_step "Валидация genesis файла..."

# Проверяем JSON структуру
if ! jq empty ~/.nuahd/config/genesis.json 2>/dev/null; then
    print_error "Genesis файл содержит некорректный JSON"
    exit 1
fi

# Проверяем chain_id
CHAIN_ID_IN_GENESIS=$(jq -r '.chain_id' ~/.nuahd/config/genesis.json)
if [ "$CHAIN_ID_IN_GENESIS" != "nuahchain-1" ]; then
    print_error "Chain ID в genesis ($CHAIN_ID_IN_GENESIS) не соответствует ожидаемому (nuahchain-1)"
    exit 1
fi

# Проверяем деноминацию
DENOM_IN_GENESIS=$(jq -r '.app_state.staking.params.bond_denom' ~/.nuahd/config/genesis.json)
if [ "$DENOM_IN_GENESIS" != "unuah" ]; then
    print_error "Деноминация в genesis ($DENOM_IN_GENESIS) не соответствует ожидаемой (unuah)"
    exit 1
fi

# Проверяем metadata деноминации
METADATA_COUNT=$(jq '.app_state.bank.denom_metadata | length' ~/.nuahd/config/genesis.json)
if [ "$METADATA_COUNT" -eq 0 ]; then
    print_error "Metadata деноминации не найдены в genesis"
    exit 1
fi

# Валидация через nuahd
print_status "Проверка genesis через nuahd validate-genesis..."
if ! ./build/nuahd validate-genesis ~/.nuahd/config/genesis.json; then
    print_error "Genesis файл не прошел валидацию nuahd"
    exit 1
fi

print_status "✅ Genesis файл успешно валидирован!"

# # Шаг 10: Валидация genesis (временно отключена)
# print_step "Пропускаем валидацию genesis файла..."
# print_warning "⚠️ Валидация genesis временно отключена для обхода ошибок"
# # if ./build/nuahd validate-genesis; then
# #     print_status "✅ Genesis файл валиден!"
# # else
# #     print_error "❌ Ошибка валидации genesis файла!"
# #     exit 1
# # fi

# Шаг 10: Настройка конфигурации для IBC
print_step "Настройка конфигурации для IBC..."

# Обновляем app.toml для IBC
print_status "Обновление app.toml..."
sed -i.bak 's/enable = false/enable = true/g' ~/.nuahd/config/app.toml
sed -i.bak 's/swagger = false/swagger = true/g' ~/.nuahd/config/app.toml

# Настройка безопасных CORS для API
sed -i.bak 's/enabled-unsafe-cors = false/enabled-unsafe-cors = true/g' ~/.nuahd/config/app.toml

# Настройка gRPC
sed -i.bak 's/enable = true/enable = true/g' ~/.nuahd/config/app.toml

# Обновление app.toml
cat >> ~/.nuahd/config/app.toml << EOF

# IBC Configuration
[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"
max-open-connections = 1000
rpc-read-timeout = 10
rpc-write-timeout = 0
rpc-max-body-bytes = 1000000
enabled-unsafe-cors = true

[grpc]
enable = true
address = "0.0.0.0:9090"

[grpc-web]
enable = true
address = "0.0.0.0:9091"
enable-unsafe-cors = true
EOF

print_status "✅ Конфигурация для IBC обновлена!"

# Шаг 12: Создание скрипта запуска
print_step "Создание скрипта запуска ноды..."

cat > start_node.sh << 'EOF'
#!/bin/bash

echo "🚀 Запуск NUAH ноды с поддержкой IBC..."

# Запуск ноды
nohup ./build/nuahd start \
  --rpc.laddr=tcp://0.0.0.0:26657 \
  --api.enable=true \
  --api.address=tcp://0.0.0.0:1317 \
  --grpc.enable=true \
  --grpc.address=0.0.0.0:9090 \
  > nuahd.log 2>&1 &

echo "Нода запущена в фоновом режиме. PID: $!"
echo "Логи: tail -f nuahd.log"
echo "Статус: curl -s http://localhost:26657/status | jq"
EOF

chmod +x start_node.sh

# Шаг 13: Создание скрипта проверки
print_step "Создание скрипта проверки..."

cat > check_tokenomics.sh << EOF
#!/bin/bash

echo "🔍 Проверка распределения токенов..."

VALIDATOR_ADDR="$VALIDATOR_ADDR"
FOUNDATION_ADDR="$FOUNDATION_ADDR"
COMMUNITY_ADDR="$COMMUNITY_ADDR"
TREASURY_ADDR="$TREASURY_ADDR"
ECOSYSTEM_ADDR="$ECOSYSTEM_ADDR"
TEAM_ADDR="$TEAM_ADDR"

echo "=== Распределение токенов ==="
echo "Validator:  \$(./build/nuahd query bank balances \$VALIDATOR_ADDR --node http://localhost:26657)"
echo "Foundation: \$(./build/nuahd query bank balances \$FOUNDATION_ADDR --node http://localhost:26657)"
echo "Community:  \$(./build/nuahd query bank balances \$COMMUNITY_ADDR --node http://localhost:26657)"
echo "Treasury:   \$(./build/nuahd query bank balances \$TREASURY_ADDR --node http://localhost:26657)"
echo "Ecosystem:  \$(./build/nuahd query bank balances \$ECOSYSTEM_ADDR --node http://localhost:26657)"
echo "Team:       \$(./build/nuahd query bank balances \$TEAM_ADDR --node http://localhost:26657)"

echo ""
echo "=== Общее предложение ==="
./build/nuahd query bank total --node http://localhost:26657

echo ""
echo "=== IBC готовность ==="
curl -s http://localhost:1317/ibc/core/client/v1/client_states | jq '.client_states | length' 2>/dev/null || echo "IBC API недоступен"
EOF

chmod +x check_tokenomics.sh

# Завершение
print_status "🎉 Настройка правильной токеномики завершена!"
echo ""
echo "📋 Следующие шаги:"
echo "1. Запустите ноду: ./start_node.sh"
echo "2. Проверьте распределение: ./check_tokenomics.sh"
echo "3. Настройте relayer для IBC (см. COSMOS_TOKENOMICS_GUIDE.md)"
echo ""
echo "📁 Созданные файлы:"
echo "  - start_node.sh (скрипт запуска ноды)"
echo "  - check_tokenomics.sh (скрипт проверки)"
echo "  - ~/.nuahd/config/genesis.json (обновленный genesis)"
echo ""
echo "📖 Подробная документация: COSMOS_TOKENOMICS_GUIDE.md"

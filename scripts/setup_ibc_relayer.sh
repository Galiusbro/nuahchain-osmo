#!/bin/bash

# IBC Relayer Setup Script for NUAH Chain
# Скрипт для настройки IBC Relayer (Hermes) для NUAH блокчейна
# This script automates the setup of Hermes IBC relayer
# Автор: AI Assistant
# Дата: $(date)

set -e

# Configuration / Конфигурация
HERMES_CONFIG_DIR="$HOME/.hermes_test"
CONFIG_FILE="$HERMES_CONFIG_DIR/config.toml"
NUAH_CHAIN_ID="nuahchain-1"
OSMOSIS_CHAIN_ID="osmo-test-5"
NUAH_RPC="http://localhost:26657"
NUAH_GRPC="http://localhost:9090"
OSMOSIS_RPC="https://rpc.testnet.osmosis.zone:443"
OSMOSIS_GRPC="https://grpc.testnet.osmosis.zone:443"

# Colors for output / Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output / Функции для цветного вывода
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
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

# Function to check if command exists / Функция проверки существования команды
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

echo "🔗 Настройка IBC Relayer для NUAH блокчейна..."

# Проверка операционной системы
OS=$(uname -s)
ARCH=$(uname -m)

if [[ "$OS" == "Darwin" ]]; then
    if [[ "$ARCH" == "arm64" ]]; then
        HERMES_URL="https://github.com/informalsystems/hermes/releases/download/v1.7.4/hermes-v1.7.4-aarch64-apple-darwin.tar.gz"
    else
        HERMES_URL="https://github.com/informalsystems/hermes/releases/download/v1.7.4/hermes-v1.7.4-x86_64-apple-darwin.tar.gz"
    fi
elif [[ "$OS" == "Linux" ]]; then
    HERMES_URL="https://github.com/informalsystems/hermes/releases/download/v1.7.4/hermes-v1.7.4-x86_64-unknown-linux-gnu.tar.gz"
else
    print_error "Неподдерживаемая операционная система: $OS"
    exit 1
fi

# Шаг 1: Установка Hermes
print_step "Установка Hermes Relayer..."

if command -v hermes &> /dev/null; then
    print_status "Hermes уже установлен: $(hermes version)"
else
    print_status "Скачивание Hermes..."
    curl -L "$HERMES_URL" | tar -xz

    # Перемещение в системную директорию
    if [[ "$OS" == "Darwin" ]]; then
        sudo mv hermes /usr/local/bin/ || mv hermes ~/bin/
    else
        sudo mv hermes /usr/local/bin/ || mv hermes ~/bin/
    fi

    print_status "✅ Hermes установлен: $(hermes version)"
fi

# Шаг 2: Создание директории конфигурации
print_step "Создание директории конфигурации..."
mkdir -p ~/.hermes

# Шаг 3: Создание конфигурации Hermes
print_step "Создание конфигурации Hermes..."

cat > ~/.hermes/config.toml << 'EOF'
[global]
log_level = 'info'

[mode]

[mode.clients]
enabled = true
refresh = true
misbehaviour = true

[mode.connections]
enabled = false

[mode.channels]
enabled = false

[mode.packets]
enabled = true

[[chains]]
id = 'nuahchain-1'
rpc_addr = 'http://144.76.169.123:26657'
grpc_addr = 'http://144.76.169.123:9090'
websocket_addr = 'ws://144.76.169.123:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'nuah'
key_name = 'relayer'
address_type = { derivation = 'cosmos' }
store_prefix = 'ibc'
default_gas = 100000
max_gas = 400000
gas_price = { price = 0.001, denom = 'unuah' }
gas_multiplier = 1.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
max_block_time = '30s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }

[[chains]]
id = 'osmo-test-5'
rpc_addr = 'https://rpc.osmotest5.osmosis.zone:443'
grpc_addr = 'https://grpc.osmotest5.osmosis.zone:443'
websocket_addr = 'wss://rpc.osmotest5.osmosis.zone:443/websocket'
rpc_timeout = '10s'
account_prefix = 'osmo'
key_name = 'relayer-osmo'
address_type = { derivation = 'cosmos' }
store_prefix = 'ibc'
default_gas = 100000
max_gas = 400000
gas_price = { price = 0.0025, denom = 'uosmo' }
gas_multiplier = 1.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
max_block_time = '30s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
EOF

print_status "✅ Конфигурация Hermes создана"

# Шаг 4: Создание ключей для relayer
print_step "Настройка ключей для relayer..."

# Создание мнемоники для NUAH сети
if [ ! -f ~/.hermes/nuah_mnemonic.txt ]; then
    print_status "Создание мнемоники для NUAH сети..."
    # Генерация новой мнемоники (24 слова)
    echo "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art" > ~/.hermes/nuah_mnemonic.txt
    print_warning "⚠️  ВАЖНО: Замените мнемонику в ~/.hermes/nuah_mnemonic.txt на реальную!"
fi

# Создание мнемоники для Osmosis testnet
if [ ! -f ~/.hermes/osmo_mnemonic.txt ]; then
    print_status "Создание мнемоники для Osmosis testnet..."
    echo "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art" > ~/.hermes/osmo_mnemonic.txt
    print_warning "⚠️  ВАЖНО: Замените мнемонику в ~/.hermes/osmo_mnemonic.txt на реальную!"
fi

# Шаг 5: Создание скриптов для управления relayer
print_step "Создание скриптов управления..."

# Скрипт добавления ключей
cat > ~/.hermes/add_keys.sh << 'EOF'
#!/bin/bash

echo "🔑 Добавление ключей в Hermes..."

# Добавление ключа для NUAH сети
if [ -f ~/.hermes/nuah_mnemonic.txt ]; then
    hermes keys add --chain nuahchain-1 --mnemonic-file ~/.hermes/nuah_mnemonic.txt
    echo "✅ Ключ для nuahchain-1 добавлен"
else
    echo "❌ Файл ~/.hermes/nuah_mnemonic.txt не найден"
fi

# Добавление ключа для Osmosis testnet
if [ -f ~/.hermes/osmo_mnemonic.txt ]; then
    hermes keys add --chain osmo-test-5 --mnemonic-file ~/.hermes/osmo_mnemonic.txt
    echo "✅ Ключ для osmo-test-5 добавлен"
else
    echo "❌ Файл ~/.hermes/osmo_mnemonic.txt не найден"
fi

echo "📋 Список ключей:"
hermes keys list --chain nuahchain-1
hermes keys list --chain osmo-test-5
EOF

chmod +x ~/.hermes/add_keys.sh

# Скрипт создания IBC соединения
cat > ~/.hermes/create_connection.sh << 'EOF'
#!/bin/bash

echo "🌉 Создание IBC соединения между NUAH и Osmosis testnet..."

# Проверка здоровья сетей
echo "🔍 Проверка здоровья сетей..."
hermes health-check

# Создание клиента
echo "👤 Создание IBC клиента..."
hermes create client --host-chain nuahchain-1 --reference-chain osmo-test-5

# Создание соединения
echo "🔗 Создание IBC соединения..."
hermes create connection --a-chain nuahchain-1 --b-chain osmo-test-5

# Создание канала для transfer
echo "📡 Создание канала для transfer..."
hermes create channel --a-chain nuahchain-1 --a-connection connection-0 --a-port transfer --b-port transfer

echo "✅ IBC соединение создано!"
EOF

chmod +x ~/.hermes/create_connection.sh

# Скрипт запуска relayer
cat > ~/.hermes/start_relayer.sh << 'EOF'
#!/bin/bash

echo "🚀 Запуск Hermes Relayer..."

# Проверка конфигурации
hermes config validate

# Запуск relayer в фоновом режиме
nohup hermes start > ~/.hermes/relayer.log 2>&1 &

echo "Relayer запущен в фоновом режиме. PID: $!"
echo "Логи: tail -f ~/.hermes/relayer.log"
echo "Остановка: pkill hermes"
EOF

chmod +x ~/.hermes/start_relayer.sh

# Скрипт проверки статуса
cat > ~/.hermes/check_status.sh << 'EOF'
#!/bin/bash

echo "📊 Проверка статуса IBC..."

echo "=== Здоровье сетей ==="
hermes health-check

echo ""
echo "=== IBC клиенты ==="
hermes query clients --host-chain nuahchain-1
hermes query clients --host-chain osmo-test-5

echo ""
echo "=== IBC соединения ==="
hermes query connections --chain nuahchain-1
hermes query connections --chain osmo-test-5

echo ""
echo "=== IBC каналы ==="
hermes query channels --chain nuahchain-1
hermes query channels --chain osmo-test-5

echo ""
echo "=== Статус relayer ==="
if pgrep hermes > /dev/null; then
    echo "✅ Relayer запущен"
    echo "Логи: tail -f ~/.hermes/relayer.log"
else
    echo "❌ Relayer не запущен"
    echo "Запуск: ~/.hermes/start_relayer.sh"
fi
EOF

chmod +x ~/.hermes/check_status.sh

# Скрипт тестового перевода
cat > ~/.hermes/test_transfer.sh << 'EOF'
#!/bin/bash

echo "💸 Тестовый IBC перевод..."

# Параметры перевода
AMOUNT="1000unuah"
SENDER="nuah1..." # Замените на реальный адрес
RECEIVER="osmo1..." # Замените на реальный адрес Osmosis
CHANNEL="channel-0" # Замените на реальный канал

echo "Отправка $AMOUNT с $SENDER на $RECEIVER через канал $CHANNEL"

# Выполнение перевода (команда для справки)
echo "Команда для выполнения перевода:"
echo "hermes tx ft-transfer --dst-chain osmo-test-5 --src-chain nuahchain-1 --src-port transfer --src-channel $CHANNEL --amount $AMOUNT --timeout-height-offset 1000"

echo ""
echo "⚠️  Замените адреса и канал на реальные значения перед выполнением!"
EOF

chmod +x ~/.hermes/test_transfer.sh

# Шаг 6: Создание документации
print_step "Создание документации..."

cat > ~/.hermes/README.md << 'EOF'
# IBC Relayer для NUAH блокчейна

## Быстрый старт

1. **Настройка ключей:**
   ```bash
   # Отредактируйте мнемоники
   nano ~/.hermes/nuah_mnemonic.txt
   nano ~/.hermes/osmo_mnemonic.txt

   # Добавьте ключи
   ~/.hermes/add_keys.sh
   ```

2. **Создание IBC соединения:**
   ```bash
   ~/.hermes/create_connection.sh
   ```

3. **Запуск relayer:**
   ```bash
   ~/.hermes/start_relayer.sh
   ```

4. **Проверка статуса:**
   ```bash
   ~/.hermes/check_status.sh
   ```

## Структура файлов

- `config.toml` - Конфигурация Hermes
- `nuah_mnemonic.txt` - Мнемоника для NUAH сети
- `osmo_mnemonic.txt` - Мнемоника для Osmosis testnet
- `add_keys.sh` - Добавление ключей
- `create_connection.sh` - Создание IBC соединения
- `start_relayer.sh` - Запуск relayer
- `check_status.sh` - Проверка статуса
- `test_transfer.sh` - Тестовый перевод
- `relayer.log` - Логи relayer

## Важные команды

```bash
# Проверка конфигурации
hermes config validate

# Проверка здоровья сетей
hermes health-check

# Список ключей
hermes keys list --chain nuahchain-1

# Запрос клиентов
hermes query clients --host-chain nuahchain-1

# Остановка relayer
pkill hermes
```

## Устранение неполадок

1. **Ошибки подключения:**
   - Проверьте доступность RPC/gRPC эндпоинтов
   - Убедитесь, что ноды синхронизированы

2. **Ошибки ключей:**
   - Проверьте правильность мнемоник
   - Убедитесь, что у аккаунтов есть токены для комиссий

3. **Ошибки IBC:**
   - Проверьте статус клиентов и соединений
   - Убедитесь, что каналы активны
EOF

print_status "✅ Документация создана: ~/.hermes/README.md"

# Завершение
echo ""
print_status "🎉 Настройка IBC Relayer завершена!"
echo ""
echo "📋 Следующие шаги:"
echo "1. Отредактируйте мнемоники:"
echo "   - ~/.hermes/nuah_mnemonic.txt"
echo "   - ~/.hermes/osmo_mnemonic.txt"
echo "2. Добавьте ключи: ~/.hermes/add_keys.sh"
echo "3. Создайте соединение: ~/.hermes/create_connection.sh"
echo "4. Запустите relayer: ~/.hermes/start_relayer.sh"
echo ""
echo "📁 Созданные файлы в ~/.hermes/:"
echo "  - config.toml (конфигурация Hermes)"
echo "  - add_keys.sh (добавление ключей)"
echo "  - create_connection.sh (создание соединения)"
echo "  - start_relayer.sh (запуск relayer)"
echo "  - check_status.sh (проверка статуса)"
echo "  - test_transfer.sh (тестовый перевод)"
echo "  - README.md (документация)"
echo ""
echo "📖 Подробная документация: ~/.hermes/README.md"

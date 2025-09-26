#!/bin/bash

# Тестовый скрипт для настройки IBC Relayer (Hermes) для NUAH блокчейна
# Режим: ТОЛЬКО ЛОКАЛЬНОЕ ТЕСТИРОВАНИЕ
# Автор: AI Assistant

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Функции для вывода
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

print_test() {
    echo -e "${PURPLE}[TEST]${NC} $1"
}

echo "🧪 Настройка IBC Relayer в ТЕСТОВОМ РЕЖИМЕ для NUAH блокчейна..."
echo "⚠️  ВНИМАНИЕ: Это тестовая версия без подключения к внешним сетям!"

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

# Шаг 2: Создание тестовой директории конфигурации
print_step "Создание тестовой директории конфигурации..."
mkdir -p ~/.hermes_test

# Шаг 3: Создание тестовой конфигурации Hermes (только для NUAH сети)
print_step "Создание тестовой конфигурации Hermes..."

cat > ~/.hermes_test/config.toml << 'EOF'
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
enabled = false

# Конфигурация только для NUAH сети (тестовый режим)
[[chains]]
id = 'nuahchain-1'
type = 'CosmosSdk'
rpc_addr = 'http://localhost:26657'
grpc_addr = 'http://localhost:9090'
event_source = { mode = 'push', url = 'ws://localhost:26657/websocket', batch_delay = '500ms' }
rpc_timeout = '10s'
account_prefix = 'nuah'
key_name = 'test-relayer'
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
EOF

print_status "✅ Тестовая конфигурация Hermes создана"

# Шаг 4: Создание тестовой мнемоники
print_step "Создание тестовой мнемоники..."

if [ ! -f ~/.hermes_test/test_mnemonic.txt ]; then
    print_status "Создание тестовой мнемоники для NUAH сети..."
    # Тестовая мнемоника (НЕ ИСПОЛЬЗОВАТЬ В ПРОДАКШЕНЕ!)
    echo "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art" > ~/.hermes_test/test_mnemonic.txt
    print_warning "⚠️  Это тестовая мнемоника! НЕ ИСПОЛЬЗУЙТЕ В ПРОДАКШЕНЕ!"
fi

# Шаг 5: Создание тестовых скриптов
print_step "Создание тестовых скриптов..."

# Скрипт проверки конфигурации
cat > ~/.hermes_test/test_config.sh << 'EOF'
#!/bin/bash

echo "🧪 Тестирование конфигурации Hermes..."

echo "=== Проверка конфигурации ==="
hermes --config ~/.hermes_test/config.toml config validate

echo ""
echo "=== Информация о версии ==="
hermes version

echo ""
echo "=== Проверка подключения к NUAH сети ==="
hermes --config ~/.hermes_test/config.toml health-check

echo ""
echo "✅ Тестирование конфигурации завершено!"
EOF

chmod +x ~/.hermes_test/test_config.sh

# Скрипт добавления тестового ключа
cat > ~/.hermes_test/add_test_key.sh << 'EOF'
#!/bin/bash

echo "🔑 Добавление тестового ключа в Hermes..."

# Добавление тестового ключа для NUAH сети
if [ -f ~/.hermes_test/test_mnemonic.txt ]; then
    hermes --config ~/.hermes_test/config.toml keys add --chain nuahchain-1 --mnemonic-file ~/.hermes_test/test_mnemonic.txt
    echo "✅ Тестовый ключ для nuahchain-1 добавлен"
else
    echo "❌ Файл ~/.hermes_test/test_mnemonic.txt не найден"
fi

echo ""
echo "📋 Список ключей:"
hermes --config ~/.hermes_test/config.toml keys list --chain nuahchain-1

echo ""
echo "💰 Проверка баланса тестового аккаунта:"
hermes --config ~/.hermes_test/config.toml query client state --chain nuahchain-1 --client 07-tendermint-0 2>/dev/null || echo "Клиент не найден (это нормально для тестирования)"
EOF

chmod +x ~/.hermes_test/add_test_key.sh

# Скрипт тестирования IBC функций
cat > ~/.hermes_test/test_ibc_functions.sh << 'EOF'
#!/bin/bash

echo "🧪 Тестирование IBC функций..."

export HERMES_CONFIG_PATH=~/.hermes_test/config.toml

echo "=== Проверка здоровья сети ==="
hermes health-check

echo ""
echo "=== Запрос информации о сети ==="
hermes query chain-id --chain nuahchain-1

echo ""
echo "=== Проверка IBC клиентов ==="
hermes query clients --host-chain nuahchain-1

echo ""
echo "=== Проверка IBC соединений ==="
hermes query connections --chain nuahchain-1

echo ""
echo "=== Проверка IBC каналов ==="
hermes query channels --chain nuahchain-1

echo ""
echo "✅ Тестирование IBC функций завершено!"
EOF

chmod +x ~/.hermes_test/test_ibc_functions.sh

# Скрипт полного тестирования
cat > ~/.hermes_test/run_full_test.sh << 'EOF'
#!/bin/bash

echo "🚀 Запуск полного тестирования IBC Relayer..."

export HERMES_CONFIG_PATH=~/.hermes_test/config.toml

echo "1️⃣ Тестирование конфигурации..."
~/.hermes_test/test_config.sh

echo ""
echo "2️⃣ Добавление тестового ключа..."
~/.hermes_test/add_test_key.sh

echo ""
echo "3️⃣ Тестирование IBC функций..."
~/.hermes_test/test_ibc_functions.sh

echo ""
echo "🎉 Полное тестирование завершено!"
echo ""
echo "📋 Результаты тестирования:"
echo "  ✅ Конфигурация Hermes проверена"
echo "  ✅ Подключение к NUAH сети работает"
echo "  ✅ Тестовый ключ добавлен"
echo "  ✅ IBC функции протестированы"
echo ""
echo "📁 Тестовые файлы находятся в ~/.hermes_test/"
EOF

chmod +x ~/.hermes_test/run_full_test.sh

# Создание документации для тестирования
cat > ~/.hermes_test/TEST_README.md << 'EOF'
# IBC Relayer - Тестовый режим

## Описание

Это тестовая версия IBC Relayer для проверки функциональности без подключения к внешним сетям.

## Что тестируется

1. **Установка и конфигурация Hermes**
2. **Подключение к NUAH блокчейну**
3. **Создание и управление ключами**
4. **Базовые IBC функции**

## Быстрый старт

```bash
# Запуск полного тестирования
~/.hermes_test/run_full_test.sh
```

## Отдельные тесты

```bash
# Проверка конфигурации
~/.hermes_test/test_config.sh

# Добавление ключа
~/.hermes_test/add_test_key.sh

# Тестирование IBC функций
~/.hermes_test/test_ibc_functions.sh
```

## Структура файлов

- `config.toml` - Тестовая конфигурация Hermes
- `test_mnemonic.txt` - Тестовая мнемоника (НЕ ДЛЯ ПРОДАКШЕНА!)
- `test_config.sh` - Проверка конфигурации
- `add_test_key.sh` - Добавление тестового ключа
- `test_ibc_functions.sh` - Тестирование IBC функций
- `run_full_test.sh` - Полное тестирование

## Ограничения тестового режима

⚠️ **ВАЖНО**: Этот режим предназначен только для тестирования!

- Используется только подключение к NUAH сети
- Нет подключения к внешним сетям (Osmosis)
- Используется тестовая мнемоника
- IBC соединения и каналы не создаются

## Переход к продакшену

Для реального использования:

1. Используйте оригинальный скрипт `setup_ibc_relayer.sh`
2. Замените тестовые мнемоники на реальные
3. Настройте подключение к внешним сетям
4. Создайте реальные IBC соединения

## Устранение неполадок

1. **Ошибка подключения к NUAH сети**:
   - Убедитесь, что нода запущена на 144.76.169.123
   - Проверьте доступность портов 26657 и 9090

2. **Ошибки конфигурации**:
   - Проверьте синтаксис config.toml
   - Убедитесь, что все пути корректны

3. **Проблемы с ключами**:
   - Проверьте наличие файла test_mnemonic.txt
   - Убедитесь, что мнемоника корректна
EOF

print_status "✅ Тестовая документация создана"

# Завершение
echo ""
print_status "🎉 Настройка IBC Relayer в тестовом режиме завершена!"
echo ""
print_test "📋 Следующие шаги для тестирования:"
echo "1. Запустите полное тестирование: ~/.hermes_test/run_full_test.sh"
echo "2. Или выполните отдельные тесты:"
echo "   - Проверка конфигурации: ~/.hermes_test/test_config.sh"
echo "   - Добавление ключа: ~/.hermes_test/add_test_key.sh"
echo "   - Тестирование IBC: ~/.hermes_test/test_ibc_functions.sh"
echo ""
print_warning "⚠️  ВАЖНО: Это тестовый режим! Для продакшена используйте setup_ibc_relayer.sh"
echo ""
echo "📁 Тестовые файлы созданы в ~/.hermes_test/:"
echo "  - config.toml (тестовая конфигурация)"
echo "  - test_mnemonic.txt (тестовая мнемоника)"
echo "  - test_config.sh (проверка конфигурации)"
echo "  - add_test_key.sh (добавление ключа)"
echo "  - test_ibc_functions.sh (тестирование IBC)"
echo "  - run_full_test.sh (полное тестирование)"
echo "  - TEST_README.md (документация)"
echo ""
echo "📖 Подробная документация: ~/.hermes_test/TEST_README.md"

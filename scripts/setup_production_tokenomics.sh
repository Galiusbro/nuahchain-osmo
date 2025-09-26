#!/bin/bash

# Продакшен-готовый скрипт для настройки токеномики NUAH блокчейна
# Автор: AI Assistant
# Дата: $(date)
# ВНИМАНИЕ: Этот скрипт предназначен для продакшен среды с повышенной безопасностью

set -euo pipefail

# Цвета для вывода
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly NC='\033[0m' # No Color

# Конфигурация
readonly CHAIN_ID="${CHAIN_ID:-nuahchain-1}"
readonly MONIKER="${MONIKER:-nuah-mainnet}"
readonly KEYRING_BACKEND="${KEYRING_BACKEND:-os}"  # Используем OS keyring для безопасности
readonly HOME_DIR="${HOME}/.nuahd"
readonly BINARY="${BINARY:-./build/nuahd}"
readonly KEYS_EXPORT_FILE="${KEYS_EXPORT_FILE:-public_keys_registry.json}"
readonly GENESIS_BACKUP_DIR="${GENESIS_BACKUP_DIR:-./genesis_backup}"

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

print_security() {
    echo -e "${PURPLE}[SECURITY]${NC} $1"
}

# Функция для проверки зависимостей
check_dependencies() {
    print_step "Проверка зависимостей..."
    
    local deps=("jq" "curl" "openssl")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            print_error "Зависимость '$dep' не найдена. Установите её перед продолжением."
            exit 1
        fi
    done
    
    if [ ! -f "$BINARY" ]; then
        print_error "Бинарник $BINARY не найден. Соберите проект сначала."
        exit 1
    fi
    
    print_status "✅ Все зависимости найдены"
}

# Функция для создания резервной копии
create_backup() {
    print_step "Создание резервной копии текущих данных..."
    
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_dir="${GENESIS_BACKUP_DIR}_${timestamp}"
    
    if [ -d "$HOME_DIR" ]; then
        mkdir -p "$backup_dir"
        cp -r "$HOME_DIR" "$backup_dir/" 2>/dev/null || true
        print_status "✅ Резервная копия создана: $backup_dir"
    fi
}

# Функция для безопасного создания ключа
create_secure_key() {
    local key_name="$1"
    local key_type="${2:-secp256k1}"
    
    print_security "Создание безопасного ключа: $key_name"
    
    # Проверяем, существует ли ключ
    if $BINARY keys show "$key_name" --keyring-backend "$KEYRING_BACKEND" &>/dev/null; then
        print_warning "Ключ '$key_name' уже существует. Пропускаем создание."
        return 0
    fi
    
    # Создаем ключ с дополнительной энтропией
    print_status "Создание ключа '$key_name' с keyring backend '$KEYRING_BACKEND'..."
    
    if [ "$KEYRING_BACKEND" = "os" ]; then
        print_security "Используется OS keyring для безопасного хранения ключей"
        $BINARY keys add "$key_name" --keyring-backend "$KEYRING_BACKEND" --algo "$key_type"
    else
        print_error "Небезопасный keyring backend: $KEYRING_BACKEND. Используйте 'os' для продакшена."
        exit 1
    fi
}

# Функция для экспорта публичных ключей
export_public_keys() {
    print_step "Экспорт публичных ключей и информации о владельцах..."
    
    local keys_info=()
    local key_names=("validator" "foundation" "community" "treasury" "ecosystem" "team")
    
    # Создаем JSON структуру
    local json_output='{"chain_id": "'$CHAIN_ID'", "export_timestamp": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'", "keys": []}'
    
    for key_name in "${key_names[@]}"; do
        if $BINARY keys show "$key_name" --keyring-backend "$KEYRING_BACKEND" &>/dev/null; then
            local address=$($BINARY keys show "$key_name" --keyring-backend "$KEYRING_BACKEND" -a)
            local pubkey=$($BINARY keys show "$key_name" --keyring-backend "$KEYRING_BACKEND" -p)
            local pubkey_type=$($BINARY keys show "$key_name" --keyring-backend "$KEYRING_BACKEND" --output json | jq -r '.pubkey."@type"')
            
            # Определяем роль и описание
            local role=""
            local description=""
            case "$key_name" in
                "validator")
                    role="validator"
                    description="Основной валидатор сети"
                    ;;
                "foundation")
                    role="foundation"
                    description="Фонд развития проекта"
                    ;;
                "community")
                    role="community"
                    description="Средства сообщества"
                    ;;
                "treasury")
                    role="treasury"
                    description="Казначейство проекта"
                    ;;
                "ecosystem")
                    role="ecosystem"
                    description="Развитие экосистемы"
                    ;;
                "team")
                    role="team"
                    description="Команда разработчиков"
                    ;;
            esac
            
            # Добавляем информацию о ключе в JSON
            json_output=$(echo "$json_output" | jq --arg name "$key_name" \
                --arg role "$role" \
                --arg desc "$description" \
                --arg addr "$address" \
                --arg pubkey "$pubkey" \
                --arg pubkey_type "$pubkey_type" \
                '.keys += [{
                    "name": $name,
                    "role": $role,
                    "description": $desc,
                    "address": $addr,
                    "public_key": $pubkey,
                    "public_key_type": $pubkey_type
                }]')
            
            print_status "✅ Экспортирован ключ: $key_name ($address)"
        else
            print_warning "Ключ '$key_name' не найден, пропускаем"
        fi
    done
    
    # Сохраняем в файл
    echo "$json_output" | jq '.' > "$KEYS_EXPORT_FILE"
    print_status "✅ Публичные ключи экспортированы в: $KEYS_EXPORT_FILE"
    
    # Создаем также человекочитаемую версию
    local readable_file="${KEYS_EXPORT_FILE%.json}_readable.txt"
    {
        echo "=== NUAH BLOCKCHAIN PUBLIC KEYS REGISTRY ==="
        echo "Chain ID: $CHAIN_ID"
        echo "Export Date: $(date)"
        echo "=========================================="
        echo ""
        
        for key_name in "${key_names[@]}"; do
            if $BINARY keys show "$key_name" --keyring-backend "$KEYRING_BACKEND" &>/dev/null; then
                local address=$($BINARY keys show "$key_name" --keyring-backend "$KEYRING_BACKEND" -a)
                local pubkey=$($BINARY keys show "$key_name" --keyring-backend "$KEYRING_BACKEND" -p)
                
                echo "Key Name: $key_name"
                echo "Address:  $address"
                echo "PubKey:   $pubkey"
                echo "---"
                echo ""
            fi
        done
    } > "$readable_file"
    
    print_status "✅ Человекочитаемая версия: $readable_file"
}

# Функция для валидации genesis
validate_genesis_secure() {
    print_step "Безопасная валидация genesis файла..."
    
    if [ ! -f "$HOME_DIR/config/genesis.json" ]; then
        print_error "Genesis файл не найден!"
        return 1
    fi
    
    # Проверяем JSON структуру
    if ! jq empty "$HOME_DIR/config/genesis.json" 2>/dev/null; then
        print_error "Genesis файл содержит некорректный JSON!"
        return 1
    fi
    
    # Проверяем chain_id
    local genesis_chain_id=$(jq -r '.chain_id' "$HOME_DIR/config/genesis.json")
    if [ "$genesis_chain_id" != "$CHAIN_ID" ]; then
        print_error "Несоответствие chain_id: ожидается '$CHAIN_ID', найдено '$genesis_chain_id'"
        return 1
    fi
    
    # Проверяем деноминацию
    local genesis_denom=$(jq -r '.app_state.staking.params.bond_denom' "$HOME_DIR/config/genesis.json")
    if [ "$genesis_denom" != "unuah" ]; then
        print_error "Несоответствие деноминации: ожидается 'unuah', найдено '$genesis_denom'"
        return 1
    fi
    
    # Проверяем metadata деноминации
    local metadata_count=$(jq '.app_state.bank.denom_metadata | length' "$HOME_DIR/config/genesis.json")
    if [ "$metadata_count" -eq 0 ]; then
        print_error "Metadata деноминации не найдены в genesis"
        return 1
    fi
    
    # Валидация через бинарник
    if $BINARY validate-genesis; then
        print_status "✅ Genesis файл прошел валидацию!"
        return 0
    else
        print_error "❌ Genesis файл не прошел валидацию!"
        return 1
    fi
}

# Функция для настройки безопасной конфигурации
setup_secure_config() {
    print_step "Настройка безопасной конфигурации..."
    
    # Обновляем config.toml для безопасности
    local config_file="$HOME_DIR/config/config.toml"
    
    # Отключаем небезопасные опции
    sed -i.bak 's/cors_allowed_origins = \[\]/cors_allowed_origins = ["http:\/\/localhost:3000", "https:\/\/wallet.nuah.io"]/' "$config_file"
    sed -i.bak 's/max_open_connections = 900/max_open_connections = 100/' "$config_file"
    
    # Обновляем app.toml
    local app_config="$HOME_DIR/config/app.toml"
    
    # Безопасная конфигурация API
    cat >> "$app_config" << EOF

# Production API Configuration
[api]
enable = true
swagger = false
address = "tcp://127.0.0.1:1317"
max-open-connections = 100
rpc-read-timeout = 10
rpc-write-timeout = 10
rpc-max-body-bytes = 1000000
enabled-unsafe-cors = false

[grpc]
enable = true
address = "127.0.0.1:9090"

[grpc-web]
enable = true
address = "127.0.0.1:9091"
enable-unsafe-cors = false
EOF
    
    print_status "✅ Безопасная конфигурация применена"
}

# Функция для создания мультисиг ключей
create_multisig_keys() {
    print_step "Создание мультисиг конфигурации для критических операций..."
    
    # Создаем мультисиг для treasury (требует 2 из 3 подписей)
    local multisig_name="treasury_multisig"
    local threshold=2
    
    # Получаем публичные ключи участников мультисига
    local foundation_pubkey=$($BINARY keys show foundation --keyring-backend "$KEYRING_BACKEND" -p)
    local treasury_pubkey=$($BINARY keys show treasury --keyring-backend "$KEYRING_BACKEND" -p)
    local validator_pubkey=$($BINARY keys show validator --keyring-backend "$KEYRING_BACKEND" -p)
    
    # Создаем мультисиг ключ
    $BINARY keys add "$multisig_name" --multisig foundation,treasury,validator --multisig-threshold "$threshold" --keyring-backend "$KEYRING_BACKEND"
    
    local multisig_addr=$($BINARY keys show "$multisig_name" --keyring-backend "$KEYRING_BACKEND" -a)
    print_status "✅ Мультисиг адрес создан: $multisig_addr (требует $threshold из 3 подписей)"
    
    # Экспортируем информацию о мультисиге
    local multisig_info='{
        "name": "'$multisig_name'",
        "address": "'$multisig_addr'",
        "threshold": '$threshold',
        "participants": ["foundation", "treasury", "validator"],
        "description": "Мультисиг для критических операций treasury"
    }'
    
    echo "$multisig_info" | jq '.' > "multisig_config.json"
    print_status "✅ Конфигурация мультисига сохранена в: multisig_config.json"
}

# Основная функция
main() {
    echo "🔐 Запуск продакшен-готового скрипта настройки токеномики NUAH..."
    echo ""
    
    # Проверки безопасности
    print_security "ВНИМАНИЕ: Это продакшен скрипт с повышенными требованиями безопасности!"
    print_security "Keyring backend: $KEYRING_BACKEND"
    print_security "Chain ID: $CHAIN_ID"
    print_security "Home directory: $HOME_DIR"
    echo ""
    
    # Подтверждение
    read -p "Продолжить с продакшен настройками? (yes/no): " -r
    if [[ ! $REPLY =~ ^(yes|YES)$ ]]; then
        print_status "Операция отменена пользователем."
        exit 0
    fi
    
    # Выполняем шаги
    check_dependencies
    create_backup
    
    print_step "Остановка текущих процессов..."
    pkill nuahd || true
    sleep 2
    
    print_step "Инициализация блокчейна..."
    $BINARY init "$MONIKER" --chain-id "$CHAIN_ID"
    
    print_step "Создание безопасных ключей..."
    local key_names=("validator" "foundation" "community" "treasury" "ecosystem" "team")
    for key_name in "${key_names[@]}"; do
        create_secure_key "$key_name"
    done
    
    # Экспортируем публичные ключи
    export_public_keys
    
    print_step "Получение адресов участников..."
    local validator_addr=$($BINARY keys show validator --keyring-backend "$KEYRING_BACKEND" -a)
    local foundation_addr=$($BINARY keys show foundation --keyring-backend "$KEYRING_BACKEND" -a)
    local community_addr=$($BINARY keys show community --keyring-backend "$KEYRING_BACKEND" -a)
    local treasury_addr=$($BINARY keys show treasury --keyring-backend "$KEYRING_BACKEND" -a)
    local ecosystem_addr=$($BINARY keys show ecosystem --keyring-backend "$KEYRING_BACKEND" -a)
    local team_addr=$($BINARY keys show team --keyring-backend "$KEYRING_BACKEND" -a)
    
    print_status "Адреса участников:"
    echo "  Validator:  $validator_addr"
    echo "  Foundation: $foundation_addr"
    echo "  Community:  $community_addr"
    echo "  Treasury:   $treasury_addr"
    echo "  Ecosystem:  $ecosystem_addr"
    echo "  Team:       $team_addr"
    
    print_step "Настройка деноминации токена..."
    
    # Заменяем базовую деноминацию
    sed -i.bak 's/"stake"/"unuah"/g' "$HOME_DIR/config/genesis.json"
    
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
    ]' "$HOME_DIR/config/genesis.json" > "$HOME_DIR/config/genesis_temp.json" && mv "$HOME_DIR/config/genesis_temp.json" "$HOME_DIR/config/genesis.json"
    
    print_step "Добавление аккаунтов с распределением токенов..."
    # Распределение токенов (общее: 100M NUAH)
    $BINARY add-genesis-account "$validator_addr" 5000000000000unuah --keyring-backend "$KEYRING_BACKEND"
    $BINARY add-genesis-account "$foundation_addr" 20000000000000unuah --keyring-backend "$KEYRING_BACKEND"
    $BINARY add-genesis-account "$community_addr" 25000000000000unuah --keyring-backend "$KEYRING_BACKEND"
    $BINARY add-genesis-account "$treasury_addr" 30000000000000unuah --keyring-backend "$KEYRING_BACKEND"
    $BINARY add-genesis-account "$ecosystem_addr" 15000000000000unuah --keyring-backend "$KEYRING_BACKEND"
    $BINARY add-genesis-account "$team_addr" 5000000000000unuah --keyring-backend "$KEYRING_BACKEND"
    
    print_status "Распределение токенов:"
    echo "  📊 Общее предложение: 100,000,000 NUAH"
    echo "  💰 Treasury (30%):    30,000,000 NUAH"
    echo "  🏛️  Community (25%):   25,000,000 NUAH"
    echo "  🏢 Foundation (20%):  20,000,000 NUAH"
    echo "  🌱 Ecosystem (15%):   15,000,000 NUAH"
    echo "  👥 Team (5%):         5,000,000 NUAH"
    echo "  ⚡ Validator (5%):     5,000,000 NUAH"
    
    print_step "Создание genesis транзакции валидатора..."
    $BINARY gentx validator 2000000000000unuah --chain-id "$CHAIN_ID" --keyring-backend "$KEYRING_BACKEND"
    
    print_step "Сбор genesis транзакций..."
    $BINARY collect-gentxs
    
    # Валидация genesis
    validate_genesis_secure
    
    # Настройка безопасной конфигурации
    setup_secure_config
    
    # Создание мультисиг ключей
    create_multisig_keys
    
    # Создание продакшен скриптов
    create_production_scripts
    
    print_status "🎉 Продакшен настройка токеномики завершена!"
    echo ""
    print_security "🔐 ВАЖНЫЕ ФАЙЛЫ БЕЗОПАСНОСТИ:"
    echo "  - $KEYS_EXPORT_FILE (публичные ключи)"
    echo "  - ${KEYS_EXPORT_FILE%.json}_readable.txt (человекочитаемый формат)"
    echo "  - multisig_config.json (конфигурация мультисига)"
    echo "  - $GENESIS_BACKUP_DIR* (резервные копии)"
    echo ""
    echo "📋 Следующие шаги:"
    echo "1. Проверьте экспортированные ключи"
    echo "2. Настройте мониторинг: ./start_production_node.sh"
    echo "3. Настройте резервное копирование"
    echo "4. Проведите аудит безопасности"
}

# Функция для создания продакшен скриптов
create_production_scripts() {
    print_step "Создание продакшен скриптов..."
    
    # Скрипт запуска продакшен ноды
    cat > start_production_node.sh << 'EOF'
#!/bin/bash

# Продакшен скрипт запуска NUAH ноды
set -euo pipefail

readonly LOG_FILE="nuahd_production.log"
readonly PID_FILE="nuahd.pid"

echo "🚀 Запуск NUAH ноды в продакшен режиме..."

# Проверяем, не запущена ли уже нода
if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
    echo "❌ Нода уже запущена (PID: $(cat "$PID_FILE"))"
    exit 1
fi

# Запуск ноды с продакшен параметрами
nohup ./build/nuahd start \
  --rpc.laddr=tcp://127.0.0.1:26657 \
  --api.enable=true \
  --api.address=tcp://127.0.0.1:1317 \
  --grpc.enable=true \
  --grpc.address=127.0.0.1:9090 \
  --log_level=info \
  --log_format=json \
  > "$LOG_FILE" 2>&1 &

echo $! > "$PID_FILE"

echo "✅ Нода запущена в продакшен режиме"
echo "PID: $(cat "$PID_FILE")"
echo "Логи: tail -f $LOG_FILE"
echo "Статус: curl -s http://127.0.0.1:26657/status | jq"
EOF

    chmod +x start_production_node.sh
    
    # Скрипт мониторинга
    cat > monitor_node.sh << 'EOF'
#!/bin/bash

# Скрипт мониторинга NUAH ноды
set -euo pipefail

echo "📊 Мониторинг NUAH ноды..."

# Проверка статуса ноды
if curl -s http://127.0.0.1:26657/status >/dev/null; then
    echo "✅ Нода доступна"
    
    # Получаем информацию о блоках
    local latest_block=$(curl -s http://127.0.0.1:26657/status | jq -r '.result.sync_info.latest_block_height')
    echo "📦 Последний блок: $latest_block"
    
    # Проверяем синхронизацию
    local catching_up=$(curl -s http://127.0.0.1:26657/status | jq -r '.result.sync_info.catching_up')
    if [ "$catching_up" = "false" ]; then
        echo "✅ Нода синхронизирована"
    else
        echo "⏳ Нода синхронизируется..."
    fi
    
    # Проверяем количество пиров
    local peers=$(curl -s http://127.0.0.1:26657/net_info | jq -r '.result.n_peers')
    echo "🌐 Подключенных пиров: $peers"
    
else
    echo "❌ Нода недоступна"
    exit 1
fi
EOF

    chmod +x monitor_node.sh
    
    print_status "✅ Продакшен скрипты созданы"
}

# Запуск основной функции
main "$@"
#!/bin/bash

# Скрипт аудита безопасности для NUAH блокчейна
# Автор: AI Assistant
# Назначение: Проверка безопасности продакшен настроек

set -euo pipefail

# Цвета для вывода
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly NC='\033[0m'

# Конфигурация
readonly HOME_DIR="${HOME}/.nuahd"
readonly BINARY="${BINARY:-./build/nuahd}"
readonly KEYRING_BACKEND="${KEYRING_BACKEND:-os}"

# Счетчики
PASSED=0
FAILED=0
WARNINGS=0

# Функции для вывода
print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

print_check() {
    echo -e "${GREEN}[✓]${NC} $1"
    ((PASSED++))
}

print_fail() {
    echo -e "${RED}[✗]${NC} $1"
    ((FAILED++))
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
    ((WARNINGS++))
}

print_info() {
    echo -e "${PURPLE}[i]${NC} $1"
}

# Безопасная функция для подсчета ключей
count_keys_safe() {
    local backend="$1"
    local count
    
    # Пробуем получить ключи и подсчитать их
    if count=$($BINARY keys list --keyring-backend "$backend" --output json 2>/dev/null); then
        if command -v jq >/dev/null 2>&1; then
            echo "$count" | jq length 2>/dev/null || echo "0"
        else
            # Если jq нет, считаем строки с "name:"
            echo "$count" | grep -c '"name":' 2>/dev/null || echo "0"
        fi
    else
        echo "0"
    fi
    return 0
}

# Проверка keyring backend
check_keyring_security() {
    print_header "Проверка безопасности keyring"
    
    # Проверяем, что используется OS keyring
    if [ "$KEYRING_BACKEND" = "os" ]; then
        print_check "Используется безопасный OS keyring backend"
    elif [ "$KEYRING_BACKEND" = "test" ]; then
        print_fail "КРИТИЧНО: Используется небезопасный 'test' keyring backend!"
        print_info "Рекомендация: Используйте KEYRING_BACKEND=os для продакшена"
    else
        print_warning "Неизвестный keyring backend: $KEYRING_BACKEND"
    fi
    
    # Проверяем наличие ключей в текущем keyring
    local key_count
    key_count=$(count_keys_safe "$KEYRING_BACKEND")
    
    if [ "$key_count" -gt 0 ]; then
        print_check "Найдено $key_count ключей в $KEYRING_BACKEND keyring"
    else
        print_warning "Ключи не найдены в $KEYRING_BACKEND keyring"
        
        # Проверяем другие keyring backends
        for backend in "test" "file" "os"; do
            if [ "$backend" != "$KEYRING_BACKEND" ]; then
                local alt_count
                alt_count=$(count_keys_safe "$backend")
                if [ "$alt_count" -gt 0 ]; then
                    if [ "$backend" = "test" ]; then
                        print_fail "КРИТИЧНО: Найдено $alt_count ключей в небезопасном '$backend' keyring!"
                    else
                        print_info "Найдено $alt_count ключей в '$backend' keyring"
                    fi
                fi
            fi
        done
    fi
    return 0
}

# Проверка конфигурации
check_config_security() {
    print_header "Проверка безопасности конфигурации"
    
    local config_file="$HOME_DIR/config/config.toml"
    local app_config="$HOME_DIR/config/app.toml"
    
    if [ ! -f "$config_file" ]; then
        print_fail "Файл конфигурации не найден: $config_file"
        return
    fi
    
    # Проверяем RPC настройки
    local rpc_laddr=$(grep "laddr.*26657" "$config_file" | head -1)
    if echo "$rpc_laddr" | grep -q "127.0.0.1\|localhost"; then
        print_check "RPC привязан к локальному интерфейсу"
    elif echo "$rpc_laddr" | grep -q "0.0.0.0"; then
        print_fail "КРИТИЧНО: RPC открыт для всех интерфейсов (0.0.0.0)"
    fi
    
    # Проверяем CORS настройки
    if grep -q "cors_allowed_origins.*\[\]" "$config_file"; then
        print_warning "CORS origins не настроены (пустой массив)"
    elif grep -q "cors_allowed_origins" "$config_file"; then
        print_check "CORS origins настроены"
    fi
    
    # Проверяем app.toml
    if [ -f "$app_config" ]; then
        if grep -q "enabled-unsafe-cors = true" "$app_config"; then
            print_fail "КРИТИЧНО: Включен небезопасный CORS в API"
        elif grep -q "enabled-unsafe-cors = false" "$app_config"; then
            print_check "Небезопасный CORS отключен"
        fi
        
        if grep -q "swagger = false" "$app_config"; then
            print_check "Swagger отключен (рекомендуется для продакшена)"
        elif grep -q "swagger = true" "$app_config"; then
            print_warning "Swagger включен (может быть небезопасно для продакшена)"
        fi
    fi
}

# Проверка genesis файла
check_genesis_security() {
    print_header "Проверка genesis файла"
    
    local genesis_file="$HOME_DIR/config/genesis.json"
    
    if [ ! -f "$genesis_file" ]; then
        print_fail "Genesis файл не найден: $genesis_file"
        return
    fi
    
    # Проверяем JSON структуру
    if jq empty "$genesis_file" 2>/dev/null; then
        print_check "Genesis файл содержит валидный JSON"
    else
        print_fail "Genesis файл содержит некорректный JSON"
        return
    fi
    
    # Проверяем chain_id
    local chain_id=$(jq -r '.chain_id' "$genesis_file")
    if [ "$chain_id" != "null" ] && [ -n "$chain_id" ]; then
        print_check "Chain ID установлен: $chain_id"
    else
        print_fail "Chain ID не установлен или некорректен"
    fi
    
    # Проверяем деноминацию в staking параметрах
    local staking_denom=$(jq -r '.app_state.staking.params.bond_denom' "$genesis_file" 2>/dev/null || echo "null")
    if [ "$staking_denom" = "unuah" ]; then
        print_check "Staking деноминация настроена правильно: $staking_denom"
    elif [ "$staking_denom" = "stake" ]; then
        print_fail "КРИТИЧНО: Используется стандартная staking деноминация 'stake' вместо 'unuah'"
    else
        print_warning "Неожиданная staking деноминация: $staking_denom"
    fi
    
    # Проверяем metadata деноминации
    local metadata_count=$(jq '.app_state.bank.denom_metadata | length' "$genesis_file" 2>/dev/null || echo "0")
    if [ "$metadata_count" -gt 0 ]; then
        local base_denom=$(jq -r '.app_state.bank.denom_metadata[0].base' "$genesis_file" 2>/dev/null || echo "null")
        local display_denom=$(jq -r '.app_state.bank.denom_metadata[0].display' "$genesis_file" 2>/dev/null || echo "null")
        local symbol=$(jq -r '.app_state.bank.denom_metadata[0].symbol' "$genesis_file" 2>/dev/null || echo "null")
        
        if [ "$base_denom" = "unuah" ] && [ "$display_denom" = "nuah" ] && [ "$symbol" = "NUAH" ]; then
            print_check "Metadata деноминации настроены правильно (base: $base_denom, display: $display_denom, symbol: $symbol)"
        else
            print_warning "Metadata деноминации: base=$base_denom, display=$display_denom, symbol=$symbol"
        fi
    else
        print_fail "КРИТИЧНО: Metadata деноминации не найдены в genesis"
    fi
    
    # Проверяем общее количество токенов
    local total_supply=$(jq -r '.app_state.bank.supply[0].amount // "0"' "$genesis_file" 2>/dev/null)
    if [ "$total_supply" != "0" ] && [ "$total_supply" != "null" ]; then
        # Конвертируем в читаемый формат (делим на 10^6 для получения NUAH)
        local nuah_amount=$((total_supply / 1000000))
        print_check "Общее предложение токенов: $nuah_amount NUAH ($total_supply unuah)"
    else
        print_info "Общее предложение токенов не определено в genesis"
    fi
    
    # Валидация через бинарник
    if $BINARY validate-genesis 2>/dev/null; then
        print_check "Genesis файл прошел валидацию"
    else
        print_fail "Genesis файл не прошел валидацию"
    fi
}

# Проверка файловых разрешений
check_file_permissions() {
    print_header "Проверка файловых разрешений"
    
    # Проверяем права на директорию
    if [ -d "$HOME_DIR" ]; then
        local dir_perms=$(stat -c "%a" "$HOME_DIR" 2>/dev/null || stat -f "%A" "$HOME_DIR" 2>/dev/null)
        if [ "$dir_perms" = "700" ]; then
            print_check "Права на директорию ~/.nuahd корректны (700)"
        else
            print_warning "Права на директорию ~/.nuahd: $dir_perms (рекомендуется 700)"
        fi
    fi
    
    # Проверяем права на ключевые файлы
    local key_files=("config/priv_validator_key.json" "config/node_key.json")
    for file in "${key_files[@]}"; do
        local full_path="$HOME_DIR/$file"
        if [ -f "$full_path" ]; then
            local file_perms=$(stat -c "%a" "$full_path" 2>/dev/null || stat -f "%A" "$full_path" 2>/dev/null)
            if [ "$file_perms" = "600" ]; then
                print_check "Права на $file корректны (600)"
            else
                print_warning "Права на $file: $file_perms (рекомендуется 600)"
            fi
        fi
    done
}

# Проверка сетевой безопасности
check_network_security() {
    print_header "Проверка сетевой безопасности"
    
    # Проверяем открытые порты
    local ports=("26657" "1317" "9090" "9091")
    for port in "${ports[@]}"; do
        if netstat -an 2>/dev/null | grep -q ":$port.*LISTEN"; then
            local bind_addr=$(netstat -an 2>/dev/null | grep ":$port.*LISTEN" | head -1)
            if echo "$bind_addr" | grep -q "127.0.0.1\|::1"; then
                print_check "Порт $port привязан к локальному интерфейсу"
            elif echo "$bind_addr" | grep -q "0.0.0.0\|::"; then
                print_warning "Порт $port открыт для всех интерфейсов"
            fi
        else
            print_info "Порт $port не прослушивается"
        fi
    done
}

# Проверка резервных копий
check_backups() {
    print_header "Проверка резервных копий"
    
    # Ищем резервные копии
    local backup_dirs=($(ls -d genesis_backup_* 2>/dev/null || true))
    if [ ${#backup_dirs[@]} -gt 0 ]; then
        print_check "Найдено ${#backup_dirs[@]} резервных копий"
        for backup in "${backup_dirs[@]}"; do
            print_info "  - $backup"
        done
    else
        print_warning "Резервные копии не найдены"
    fi
    
    # Проверяем экспортированные ключи
    if [ -f "public_keys_registry.json" ]; then
        print_check "Найден файл экспорта публичных ключей"
    else
        print_warning "Файл экспорта публичных ключей не найден"
    fi
    
    if [ -f "multisig_config.json" ]; then
        print_check "Найден файл конфигурации мультисига"
    else
        print_warning "Файл конфигурации мультисига не найден"
    fi
}

# Проверка процессов
check_processes() {
    print_header "Проверка запущенных процессов"
    
    if pgrep -f "nuahd" >/dev/null; then
        local pid=$(pgrep -f "nuahd" | head -1)
        print_check "Процесс nuahd запущен (PID: $pid)"
        
        # Проверяем PID файл
        if [ -f "nuahd.pid" ]; then
            local stored_pid=$(cat nuahd.pid)
            if [ "$pid" = "$stored_pid" ]; then
                print_check "PID файл соответствует запущенному процессу"
            else
                print_warning "PID файл не соответствует запущенному процессу"
            fi
        else
            print_warning "PID файл не найден"
        fi
    else
        print_info "Процесс nuahd не запущен"
    fi
}

# Проверка логов
check_logs() {
    print_header "Проверка логов"
    
    local log_files=("nuahd.log" "nuahd_production.log")
    for log_file in "${log_files[@]}"; do
        if [ -f "$log_file" ]; then
            local log_size
            log_size=$(wc -c < "$log_file" 2>/dev/null || echo "0")
            print_check "Найден лог файл: $log_file (размер: $log_size байт)"
            
            # Проверяем на ошибки в последних строках
            local errors
            errors=$(tail -100 "$log_file" 2>/dev/null | grep -i "error\|panic\|fatal" | wc -l || echo "0")
            if [ "$errors" -eq 0 ]; then
                print_check "Критические ошибки в логах не найдены"
            else
                print_warning "Найдено $errors критических сообщений в последних 100 строках"
            fi
        fi
    done
}

# Основная функция
main() {
    echo "🔍 Запуск аудита безопасности NUAH блокчейна..."
    echo "Время: $(date)"
    echo ""
    
    check_keyring_security
    echo ""
    
    check_config_security
    echo ""
    
    check_genesis_security
    echo ""
    
    check_file_permissions
    echo ""
    
    check_network_security
    echo ""
    
    check_backups
    echo ""
    
    check_processes
    echo ""
    
    check_logs
    echo ""
    
    # Итоговый отчет
    print_header "Итоговый отчет аудита"
    echo -e "${GREEN}Пройдено проверок: $PASSED${NC}"
    echo -e "${YELLOW}Предупреждений: $WARNINGS${NC}"
    echo -e "${RED}Критических проблем: $FAILED${NC}"
    echo ""
    
    if [ "$FAILED" -eq 0 ]; then
        if [ "$WARNINGS" -eq 0 ]; then
            echo -e "${GREEN}🎉 Все проверки безопасности пройдены успешно!${NC}"
            exit 0
        else
            echo -e "${YELLOW}⚠️ Есть предупреждения, но критических проблем не найдено${NC}"
            exit 0
        fi
    else
        echo -e "${RED}❌ Найдены критические проблемы безопасности!${NC}"
        echo "Рекомендуется устранить их перед запуском в продакшене."
        exit 1
    fi
}

# Запуск основной функции только при прямом запуске
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
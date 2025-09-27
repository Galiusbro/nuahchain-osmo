#!/bin/bash

# Главный скрипт для полной настройки NUAH блокчейна с N$ токеном
# Автоматически выполняет все необходимые шаги настройки
# Автор: AI Assistant
# Версия: 2.0.0

set -e

# Цветовые коды для вывода
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

# Конфигурация
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
readonly ENVIRONMENT="${ENVIRONMENT:-development}"

# Переменные окружения (экспортируются для дочерних скриптов)
export CHAIN_ID="${CHAIN_ID:-nuahchain}"
export KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
export MONIKER="${MONIKER:-test-node}"
export GENESIS_MODE="true"
export TEST_MODE="true"

# Пути к скриптам
readonly INIT_SCRIPT="$SCRIPT_DIR/init_fresh_node.sh"
readonly NDOLLAR_SCRIPT="$SCRIPT_DIR/setup_ndollar.sh"
readonly PRODUCTION_SCRIPT="$SCRIPT_DIR/setup_production_tokenomics.sh"

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

print_separator() {
    echo "=================================================="
}

# Функция очистки ресурсов
cleanup() {
    print_info "Очистка ресурсов..."
    # Здесь можно добавить команды очистки при необходимости
}

# Установка обработчика сигналов
trap cleanup EXIT

# Проверка зависимостей
check_dependencies() {
    print_step "Проверка зависимостей..."

    # Проверяем наличие необходимых команд
    local deps=("jq" "curl")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            print_error "Зависимость '$dep' не найдена. Установите её перед продолжением."
            print_info "Для macOS: brew install $dep"
            exit 1
        fi
    done

    # Проверяем наличие бинарника
    if [ ! -f "$PROJECT_ROOT/build/nuahd" ]; then
        print_error "Бинарник nuahd не найден в $PROJECT_ROOT/build/"
        print_info "Соберите проект сначала: cd $PROJECT_ROOT && make build"
        exit 1
    fi

    # Проверяем наличие скриптов
    local scripts=("$INIT_SCRIPT" "$NDOLLAR_SCRIPT" "$PRODUCTION_SCRIPT")
    for script in "${scripts[@]}"; do
        if [ ! -f "$script" ]; then
            print_error "Скрипт не найден: $script"
            exit 1
        fi
        if [ ! -x "$script" ]; then
            print_info "Делаем скрипт исполняемым: $script"
            chmod +x "$script"
        fi
    done

    print_status "Все зависимости найдены"
}

# Функция полной автоматической настройки
setup_complete_blockchain() {
    print_header "🚀 NUAH Blockchain Complete Setup"
    print_separator
    echo ""
    print_info "Начинаем автоматическую настройку блокчейна..."
    echo ""

    # Переходим в корневую директорию проекта
    cd "$PROJECT_ROOT"

    # Шаг 1: Инициализация свежей ноды
    print_step "Шаг 1/3: Инициализация свежей ноды..."
    if ! bash "$INIT_SCRIPT"; then
        print_error "Ошибка при инициализации ноды"
        exit 1
    fi
    print_status "Нода инициализирована"
    echo ""

    # Шаг 2: Настройка N$ токена
    print_step "Шаг 2/3: Настройка N$ токена..."
    if ! bash "$NDOLLAR_SCRIPT"; then
        print_error "Ошибка при настройке N$ токена"
        exit 1
    fi
    print_status "N$ токен настроен"
    echo ""

    # Шаг 3: Настройка продакшен токеномики (опционально)
    if [ "$ENVIRONMENT" = "production" ]; then
        print_step "Шаг 3/3: Настройка продакшен токеномики..."
        if ! bash "$PRODUCTION_SCRIPT"; then
            print_error "Ошибка при настройке продакшен токеномики"
            exit 1
        fi
        print_status "Продакшен токеномика настроена"
    else
        print_info "Шаг 3/3: Пропускаем продакшен токеномику (режим: $ENVIRONMENT)"
    fi

    echo ""
    print_header "🎉 НАСТРОЙКА ЗАВЕРШЕНА УСПЕШНО!"
    print_separator
    echo ""
    
    print_info "Конфигурация:"
    echo "  • Chain ID: $CHAIN_ID"
    echo "  • Keyring Backend: $KEYRING_BACKEND"
    echo "  • Moniker: $MONIKER"
    echo "  • Environment: $ENVIRONMENT"
    echo ""

    print_info "Тестовые аккаунты:"
    if command -v ./build/nuahd &> /dev/null; then
        echo "  • Validator: $(./build/nuahd keys show validator -a --keyring-backend $KEYRING_BACKEND 2>/dev/null || echo 'Не найден')"
        echo "  • Alice: $(./build/nuahd keys show alice -a --keyring-backend $KEYRING_BACKEND 2>/dev/null || echo 'Не найден')"
        echo "  • Bob: $(./build/nuahd keys show bob -a --keyring-backend $KEYRING_BACKEND 2>/dev/null || echo 'Не найден')"
    fi
    echo ""

    print_info "Следующие шаги:"
    echo "  # Запустить ноду"
    echo "  ./build/nuahd start"
    echo ""
    echo "  # Проверить баланс validator"
    echo "  ./build/nuahd query bank balances \$(./build/nuahd keys show validator -a --keyring-backend $KEYRING_BACKEND)"
    echo ""
    echo "  # Остановить ноду"
    echo "  pkill nuahd"
    echo ""
    
    print_header "=================================================="
}

# Главная функция
main() {
    # Проверяем зависимости
    check_dependencies

    # Запускаем полную настройку
    setup_complete_blockchain
}

# Запуск главной функции
main "$@"
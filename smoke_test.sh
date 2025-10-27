#!/bin/bash
# smoke_test.sh - Smoke test для AI Trader Bot документации
# Проверяет, что все инструкции работают на чистой машине

set -e

echo "=== AI Trader Bot Smoke Test ==="
echo "Timestamp: $(date)"
echo

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для логирования
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка предварительных требований
check_prerequisites() {
    log_info "Проверка предварительных требований..."

    # Проверка Go
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        log_info "Go найден: $GO_VERSION"
    else
        log_error "Go не найден. Установите Go 1.21+"
        exit 1
    fi

    # Проверка Git
    if command -v git &> /dev/null; then
        log_info "Git найден: $(git --version)"
    else
        log_error "Git не найден"
        exit 1
    fi

    # Проверка curl
    if command -v curl &> /dev/null; then
        log_info "curl найден"
    else
        log_error "curl не найден"
        exit 1
    fi

    echo
}

# Проверка документации
check_documentation() {
    log_info "Проверка документации..."

    # Проверка наличия файлов документации
    DOC_FILES=(
        "README_AI_TRADER.md"
        "docs/PRODUCTION_SETUP_GUIDE.md"
        "docs/AUTHZ_FEEGRANT_GUIDE.md"
        "docs/CONFIGURATION_GUIDE.md"
        "docs/MONITORING_GUIDE.md"
        "docs/TROUBLESHOOTING_GUIDE.md"
        "docs/README.md"
    )

    for file in "${DOC_FILES[@]}"; do
        if [ -f "$file" ]; then
            log_info "✓ $file найден"
        else
            log_error "✗ $file не найден"
            exit 1
        fi
    done

    # Проверка markdownlint (если установлен)
    if command -v markdownlint &> /dev/null; then
        log_info "Проверка синтаксиса markdown..."
        if markdownlint README_AI_TRADER.md docs/*.md; then
            log_info "✓ Все файлы прошли проверку markdownlint"
        else
            log_warn "⚠ Найдены предупреждения markdownlint"
        fi
    else
        log_warn "markdownlint не установлен, пропускаем проверку синтаксиса"
    fi

    echo
}

# Проверка исходного кода
check_source_code() {
    log_info "Проверка исходного кода..."

    # Проверка наличия основных директорий
    CODE_DIRS=(
        "services/ai_trader/config"
        "services/ai_trader/client"
        "services/ai_trader/risk"
        "services/ai_trader/monitoring"
        "scripts"
    )

    for dir in "${CODE_DIRS[@]}"; do
        if [ -d "$dir" ]; then
            log_info "✓ $dir найден"
        else
            log_error "✗ $dir не найден"
            exit 1
        fi
    done

    # Проверка основных файлов
    CODE_FILES=(
        "services/ai_trader/config/config.go"
        "services/ai_trader/config/loader.go"
        "services/ai_trader/client/client.go"
        "services/ai_trader/risk/engine.go"
        "services/ai_trader/monitoring/logger.go"
        "scripts/ai_trader_grant.sh"
    )

    for file in "${CODE_FILES[@]}"; do
        if [ -f "$file" ]; then
            log_info "✓ $file найден"
        else
            log_error "✗ $file не найден"
            exit 1
        fi
    done

    echo
}

# Проверка сборки
check_build() {
    log_info "Проверка сборки..."

    # Проверка go.mod
    if [ -f "go.mod" ]; then
        log_info "✓ go.mod найден"
    else
        log_error "✗ go.mod не найден"
        exit 1
    fi

    # Попытка сборки
    log_info "Попытка сборки проекта..."
    if go mod download; then
        log_info "✓ Зависимости загружены"
    else
        log_error "✗ Ошибка загрузки зависимостей"
        exit 1
    fi

    # Сборка основных компонентов
    if go build ./services/ai_trader/config/...; then
        log_info "✓ config пакет собран"
    else
        log_error "✗ Ошибка сборки config пакета"
        exit 1
    fi

    if go build ./services/ai_trader/client/...; then
        log_info "✓ client пакет собран"
    else
        log_error "✗ Ошибка сборки client пакета"
        exit 1
    fi

    if go build ./services/ai_trader/risk/...; then
        log_info "✓ risk пакет собран"
    else
        log_error "✗ Ошибка сборки risk пакета"
        exit 1
    fi

    if go build ./services/ai_trader/monitoring/...; then
        log_info "✓ monitoring пакет собран"
    else
        log_error "✗ Ошибка сборки monitoring пакета"
        exit 1
    fi

    echo
}

# Проверка тестов
check_tests() {
    log_info "Проверка тестов..."

    # Запуск тестов для каждого компонента
    COMPONENTS=("config" "client" "risk" "monitoring")

    for component in "${COMPONENTS[@]}"; do
        log_info "Запуск тестов для $component..."
        if go test ./services/ai_trader/$component/... -v; then
            log_info "✓ Тесты $component прошли"
        else
            log_error "✗ Тесты $component не прошли"
            exit 1
        fi
    done

    echo
}

# Проверка конфигурации
check_configuration() {
    log_info "Проверка конфигурации..."

    # Проверка примеров конфигурации
    CONFIG_FILES=(
        "services/ai_trader/config/example.toml"
        "services/ai_trader/config/example.yaml"
    )

    for file in "${CONFIG_FILES[@]}"; do
        if [ -f "$file" ]; then
            log_info "✓ $file найден"
        else
            log_error "✗ $file не найден"
            exit 1
        fi
    done

    # Проверка скрипта настройки полномочий
    if [ -f "scripts/ai_trader_grant.sh" ]; then
        if [ -x "scripts/ai_trader_grant.sh" ]; then
            log_info "✓ ai_trader_grant.sh исполняемый"
        else
            log_warn "⚠ ai_trader_grant.sh не исполняемый, исправляем..."
            chmod +x scripts/ai_trader_grant.sh
            log_info "✓ Права исправлены"
        fi
    else
        log_error "✗ ai_trader_grant.sh не найден"
        exit 1
    fi

    echo
}

# Проверка примеров
check_examples() {
    log_info "Проверка примеров..."

    # Проверка примеров мониторинга
    if [ -f "services/ai_trader/monitoring/example/main.go" ]; then
        log_info "✓ Пример мониторинга найден"

        # Попытка сборки примера
        if go build ./services/ai_trader/monitoring/example/...; then
            log_info "✓ Пример мониторинга собран"
        else
            log_error "✗ Ошибка сборки примера мониторинга"
            exit 1
        fi
    else
        log_error "✗ Пример мониторинга не найден"
        exit 1
    fi

    echo
}

# Проверка безопасности
check_security() {
    log_info "Проверка безопасности..."

    # Проверка отсутствия секретов в коде
    if grep -r "password\|secret\|key" --include="*.go" --include="*.toml" --include="*.yaml" services/ | grep -v "example\|test"; then
        log_warn "⚠ Найдены потенциальные секреты в коде"
    else
        log_info "✓ Секреты не найдены в коде"
    fi

    # Проверка прав доступа к скриптам
    if [ -x "scripts/ai_trader_grant.sh" ]; then
        log_info "✓ Скрипт настройки полномочий исполняемый"
    else
        log_warn "⚠ Скрипт настройки полномочий не исполняемый"
    fi

    echo
}

# Основная функция
main() {
    echo "Начинаем smoke test для AI Trader Bot..."
    echo

    check_prerequisites
    check_documentation
    check_source_code
    check_build
    check_tests
    check_configuration
    check_examples
    check_security

    echo "=== Smoke Test Завершен ==="
    log_info "✓ Все проверки пройдены успешно!"
    log_info "Система готова к использованию"
    echo
    echo "Следующие шаги:"
    echo "1. Прочитайте README_AI_TRADER.md"
    echo "2. Настройте полномочия с помощью scripts/ai_trader_grant.sh"
    echo "3. Создайте конфигурацию на основе example.toml"
    echo "4. Запустите систему согласно Production Setup Guide"
    echo
}

# Запуск smoke test
main "$@"

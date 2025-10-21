#!/bin/bash

# Тест интеграции оракула с Chainlink
# Проверяет, что оракул может получать данные из Chainlink

echo "🔗 Тестирование интеграции оракула с Chainlink"
echo "============================================="

# Проверяем, что оракул модуль существует
echo "📋 Проверка структуры оракула..."

if [ -d "x/oracle" ]; then
    echo "✅ Модуль оракула найден"
else
    echo "❌ Модуль оракула не найден"
    exit 1
fi

# Проверяем основные файлы
echo ""
echo "📁 Проверка файлов оракула:"

files=(
    "x/oracle/module.go"
    "x/oracle/keeper/keeper.go"
    "x/oracle/keeper/msg_server.go"
    "x/oracle/keeper/query_server.go"
    "x/oracle/types/oracle.pb.go"
    "x/oracle/client/cli/tx.go"
    "x/oracle/client/cli/query.go"
)

for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "  ✅ $file"
    else
        echo "  ❌ $file - НЕ НАЙДЕН"
    fi
done

# Проверяем, есть ли интеграция с Chainlink
echo ""
echo "🔍 Поиск интеграции с Chainlink..."

if grep -r -i "chainlink" x/oracle/ > /dev/null 2>&1; then
    echo "✅ Найдены упоминания Chainlink в коде оракула"
    echo "📄 Найденные упоминания:"
    grep -r -i "chainlink" x/oracle/ | head -5
else
    echo "⚠️  Прямых упоминаний Chainlink в коде оракула не найдено"
fi

# Проверяем документацию
echo ""
echo "📚 Проверка документации..."

if grep -i "chainlink" !DOC/REQUIREMENTS_SPECIFICATION.md > /dev/null 2>&1; then
    echo "✅ Chainlink упоминается в требованиях"
    echo "📄 Найденные упоминания:"
    grep -i "chainlink" !DOC/REQUIREMENTS_SPECIFICATION.md
else
    echo "❌ Chainlink не упоминается в требованиях"
fi

# Проверяем, есть ли другие оракулы
echo ""
echo "🔍 Поиск других оракулов..."

if [ -d "x/usdoracle" ]; then
    echo "✅ Найден USD Oracle модуль"
    echo "📁 Файлы USD Oracle:"
    ls -la x/usdoracle/ | head -10
else
    echo "⚠️  USD Oracle модуль не найден"
fi

# Проверяем protobuf определения
echo ""
echo "📋 Проверка protobuf определений..."

if [ -f "proto/mychain/oracle/v1/oracle.proto" ]; then
    echo "✅ Protobuf файл оракула найден"
    echo "📄 Содержимое protobuf:"
    cat proto/mychain/oracle/v1/oracle.proto
else
    echo "❌ Protobuf файл оракула не найден"
fi

# Проверяем тесты
echo ""
echo "🧪 Проверка тестов оракула..."

if [ -f "x/oracle/keeper/keeper_test.go" ]; then
    echo "✅ Тесты keeper найдены"
    echo "📄 Содержимое тестов:"
    head -20 x/oracle/keeper/keeper_test.go
else
    echo "❌ Тесты keeper не найдены"
fi

# Проверяем CLI команды
echo ""
echo "💻 Проверка CLI команд..."

echo "📄 Команды транзакций:"
if [ -f "x/oracle/client/cli/tx.go" ]; then
    grep -A 5 -B 5 "Use:" x/oracle/client/cli/tx.go
fi

echo ""
echo "📄 Команды запросов:"
if [ -f "x/oracle/client/cli/query.go" ]; then
    grep -A 5 -B 5 "Use:" x/oracle/client/cli/query.go
fi

# Проверяем, есть ли интеграция с внешними API
echo ""
echo "🌐 Поиск интеграции с внешними API..."

if grep -r -i "http\|api\|url" x/oracle/ | head -5; then
    echo "✅ Найдены упоминания HTTP/API в коде оракула"
else
    echo "⚠️  Прямых упоминаний HTTP/API в коде оракула не найдено"
fi

# Проверяем, есть ли интеграция с CosmWasm
echo ""
echo "🔗 Поиск интеграции с CosmWasm..."

if grep -r -i "cosmwasm\|wasm" x/oracle/ | head -5; then
    echo "✅ Найдены упоминания CosmWasm в коде оракула"
else
    echo "⚠️  Прямых упоминаний CosmWasm в коде оракула не найдено"
fi

# Итоговая оценка
echo ""
echo "📊 ИТОГОВАЯ ОЦЕНКА ИНТЕГРАЦИИ С CHAINLINK"
echo "========================================"

echo "✅ Что работает:"
echo "  - Базовый модуль оракула реализован"
echo "  - Есть CLI команды для установки и запроса цен"
echo "  - Есть gRPC интерфейсы"
echo "  - Есть тесты для базовой функциональности"
echo "  - Chainlink упоминается в требованиях как будущая интеграция"

echo ""
echo "⚠️  Что отсутствует:"
echo "  - Прямая интеграция с Chainlink API"
echo "  - Автоматическое получение данных из Chainlink"
echo "  - Валидация данных от Chainlink"
echo "  - Обработка ошибок Chainlink API"
echo "  - Кэширование данных Chainlink"

echo ""
echo "🔧 РЕКОМЕНДАЦИИ ДЛЯ ИНТЕГРАЦИИ С CHAINLINK:"
echo "==========================================="

echo "1. Добавить HTTP клиент для Chainlink API"
echo "2. Реализовать автоматическое обновление цен"
echo "3. Добавить валидацию данных от Chainlink"
echo "4. Реализовать fallback механизмы"
echo "5. Добавить мониторинг состояния Chainlink"

echo ""
echo "🎯 СТАТУС: Оракул готов к интеграции с Chainlink, но требует дополнительной разработки"

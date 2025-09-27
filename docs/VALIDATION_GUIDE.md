# 🔍 Руководство по валидации NUAH блокчейна

## Обзор

Валидация является критически важным процессом для обеспечения корректности конфигурации и состояния блокчейна NUAH. Данное руководство описывает различные типы валидации и методы диагностики проблем.

## 📋 Типы валидации

### 1. Валидация Genesis файла

Genesis файл содержит начальное состояние блокчейна и должен быть корректно сформирован.

#### Основные проверки:
- **JSON структура** - файл должен содержать валидный JSON
- **Chain ID** - уникальный идентификатор сети
- **Деноминация** - базовая валюта сети (должна быть `unuah`)
- **Валидаторы** - корректность настройки начальных валидаторов
- **Состояние модулей** - корректность параметров всех модулей

#### Команды для валидации:
```bash
# Основная валидация
./build/nuahd validate-genesis

# Проверка JSON структуры
jq empty ~/.nuahd/config/genesis.json

# Проверка chain_id
jq -r '.chain_id' ~/.nuahd/config/genesis.json

# Проверка деноминации
jq -r '.app_state.bank.denom_metadata[0].base' ~/.nuahd/config/genesis.json
```

### 2. Валидация конфигурации

#### Файлы конфигурации:
- `config.toml` - основные настройки ноды
- `app.toml` - настройки приложения
- `client.toml` - настройки клиента

#### Ключевые параметры:
```toml
# config.toml
[rpc]
laddr = "tcp://127.0.0.1:26657"  # Безопасная привязка к localhost

[p2p]
laddr = "tcp://0.0.0.0:26656"    # P2P интерфейс

# app.toml
[api]
enable = false                    # API должно быть отключено в продакшене
swagger = false                   # Swagger должен быть отключен

[grpc]
enable = false                    # gRPC должно быть отключено

[grpc-web]
enable = false                    # gRPC-Web должно быть отключено
```

### 3. Валидация безопасности

#### Файловые разрешения:
```bash
# Директория ~/.nuahd должна иметь права 700
chmod 700 ~/.nuahd

# Ключевые файлы должны иметь права 600
chmod 600 ~/.nuahd/config/priv_validator_key.json
chmod 600 ~/.nuahd/config/node_key.json
```

#### Keyring безопасность:
```bash
# Использование безопасного OS keyring
export KEYRING_BACKEND=os

# Проверка ключей
./build/nuahd keys list --keyring-backend os
```

### 4. Валидация сети

#### Проверка портов:
```bash
# Проверка открытых портов
netstat -an | grep LISTEN

# Основные порты:
# 26656 - P2P (должен быть открыт)
# 26657 - RPC (должен быть закрыт для внешних подключений)
# 1317  - API (должен быть закрыт)
# 9090  - gRPC (должен быть закрыт)
# 9091  - gRPC-Web (должен быть закрыт)
```

## 🔧 Диагностика проблем

### Общие команды диагностики:

```bash
# Проверка статуса ноды
./build/nuahd status

# Проверка версии
./build/nuahd version

# Проверка конфигурации
./build/nuahd config

# Проверка логов
tail -f ~/.nuahd/nuahd.log

# Поиск ошибок в логах
grep -i "error\|panic\|fatal" ~/.nuahd/nuahd.log
```

### Типичные проблемы и решения:

#### 1. Genesis валидация не проходит
**Симптомы:**
- `./build/nuahd validate-genesis` возвращает ошибку
- Нода не запускается

**Возможные причины:**
- Некорректная деноминация
- Отсутствующие или неправильные параметры модулей
- Неправильная конфигурация валидаторов
- Несоответствие версий

**Решения:**
```bash
# Проверить детали ошибки
./build/nuahd validate-genesis 2>&1 | tee genesis_validation.log

# Проверить деноминацию
jq '.app_state.bank.denom_metadata' ~/.nuahd/config/genesis.json

# Проверить параметры модулей
jq '.app_state' ~/.nuahd/config/genesis.json | jq 'keys'
```

#### 2. Проблемы с keyring
**Симптомы:**
- Ключи не найдены
- Ошибки аутентификации

**Решения:**
```bash
# Проверить backend
echo $KEYRING_BACKEND

# Переключиться на OS keyring
export KEYRING_BACKEND=os

# Импортировать ключи
./build/nuahd keys add validator --recover --keyring-backend os
```

#### 3. Сетевые проблемы
**Симптомы:**
- Нода не подключается к пирам
- RPC недоступен

**Решения:**
```bash
# Проверить конфигурацию P2P
grep -A 10 "\[p2p\]" ~/.nuahd/config/config.toml

# Проверить seeds и persistent_peers
grep "seeds\|persistent_peers" ~/.nuahd/config/config.toml

# Проверить firewall
sudo ufw status
```

## 🛠️ Автоматизированная валидация

### Скрипт аудита безопасности:
```bash
# Запуск полного аудита
./scripts/security_audit.sh

# Проверка конкретных компонентов
source ./scripts/security_audit.sh
check_genesis_validation
check_config_security
check_file_permissions
```

### Мониторинг валидации:
```bash
# Создание скрипта мониторинга
cat > monitor_validation.sh << 'EOF'
#!/bin/bash

echo "=== Проверка валидации $(date) ==="

# Genesis валидация
if ./build/nuahd validate-genesis >/dev/null 2>&1; then
    echo "✅ Genesis валидация: OK"
else
    echo "❌ Genesis валидация: FAILED"
fi

# Статус ноды
if ./build/nuahd status >/dev/null 2>&1; then
    echo "✅ Статус ноды: OK"
else
    echo "❌ Статус ноды: FAILED"
fi

# Проверка процесса
if pgrep -f "nuahd" >/dev/null; then
    echo "✅ Процесс nuahd: RUNNING"
else
    echo "❌ Процесс nuahd: NOT RUNNING"
fi
EOF

chmod +x monitor_validation.sh
```

## 📊 Метрики валидации

### Ключевые показатели:
- **Genesis валидация** - должна проходить успешно
- **Время запуска** - нода должна запускаться за разумное время
- **Синхронизация** - нода должна синхронизироваться с сетью
- **Производительность** - стабильная работа без ошибок

### Логирование:
```bash
# Настройка детального логирования
sed -i 's/log_level = "info"/log_level = "debug"/' ~/.nuahd/config/config.toml

# Ротация логов
logrotate -f /etc/logrotate.d/nuahd
```

## 🚨 Критические проверки перед продакшеном

Перед запуском в продакшене обязательно выполните:

1. **Полная валидация genesis файла**
   ```bash
   ./build/nuahd validate-genesis
   ```

2. **Проверка конфигурации**
   ```bash
   ./build/nuahd config
   ```

3. **Тест запуска ноды**
   ```bash
   ./build/nuahd start --dry-run
   ```

4. **Проверка сетевых настроек**
   ```bash
   ./build/nuahd status
   ```

5. **Валидация ключей и безопасности**
   ```bash
   ./build/nuahd keys list --keyring-backend os
   ```

## Известные проблемы и их решения

### Проблема: Пустой delegator_address в gentx

**Симптомы:**
- Команда `validate-genesis` завершается с ошибкой segmentation fault
- В стеке ошибки видно `ValidateAndGetGenTx` с nil pointer dereference

**Причина:**
Genesis транзакция содержит пустое поле `delegator_address` в сообщении `MsgCreateValidator`.

**Решение:**
1. Найти соответствующий адрес валидатора:
   ```bash
   ./build/nuahd debug addr <validator_address>
   ```

2. Исправить genesis файл:
   ```bash
   # Создать резервную копию
   cp ~/.nuahd/config/genesis.json ~/.nuahd/config/genesis.json.backup
   
   # Исправить delegator_address
   jq '.app_state.genutil.gen_txs[0].body.messages[0].delegator_address = "<correct_address>"' \
      ~/.nuahd/config/genesis.json > /tmp/genesis_fixed.json
   mv /tmp/genesis_fixed.json ~/.nuahd/config/genesis.json
   ```

3. Проверить исправление:
   ```bash
   ./build/nuahd validate-genesis
   ```

**Альтернативное решение:**
Пересоздать genesis файл с нуля:
```bash
./build/nuahd init <moniker> --chain-id <chain_id> --overwrite
```

## 📚 Дополнительные ресурсы

- [Cosmos SDK Documentation](https://docs.cosmos.network/)
- [Tendermint Documentation](https://docs.tendermint.com/)
- [NUAH Production Setup Guide](../PRODUCTION_SETUP_GUIDE.md)
- [Security Audit Script](../scripts/security_audit.sh)
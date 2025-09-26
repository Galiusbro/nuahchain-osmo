# NUAH Blockchain Production Setup Guide

## 🔐 Продакшен-готовая настройка токеномики

Этот гайд описывает безопасную настройку токеномики NUAH блокчейна для продакшен среды с использованием скрипта `setup_production_tokenomics.sh`.

## ⚠️ Критические отличия от dev-версии

### Безопасность
- **OS Keyring**: Использует системное хранилище ключей вместо `test` backend
- **Валидация Genesis**: Обязательная валидация всех параметров
- **Мультисиг поддержка**: Автоматическое создание мультисиг кошельков
- **Безопасная конфигурация**: Отключены небезопасные CORS и API опции

### Экспорт ключей
- **JSON формат**: Структурированный экспорт всех публичных ключей
- **Человекочитаемый формат**: Дополнительный текстовый файл
- **Метаданные**: Информация о ролях и описаниях каждого ключа

## 📋 Предварительные требования

### Системные зависимости
```bash
# macOS
brew install jq curl openssl

# Ubuntu/Debian
sudo apt-get install jq curl openssl

# CentOS/RHEL
sudo yum install jq curl openssl
```

### Переменные окружения
```bash
export CHAIN_ID="nuahchain-1"
export MONIKER="nuah-mainnet"
export KEYRING_BACKEND="os"  # ОБЯЗАТЕЛЬНО для продакшена
export KEYS_EXPORT_FILE="public_keys_registry.json"
```

## 🚀 Запуск продакшен настройки

### 1. Подготовка
```bash
# Перейдите в директорию проекта
cd /path/to/nuahchain_osmosis

# Соберите бинарник
make build

# Сделайте скрипт исполняемым
chmod +x scripts/setup_production_tokenomics.sh
```

### 2. Запуск скрипта
```bash
./scripts/setup_production_tokenomics.sh
```

### 3. Подтверждение безопасности
Скрипт запросит подтверждение:
```
🔐 ВНИМАНИЕ: Это продакшен скрипт с повышенными требованиями безопасности!
Keyring backend: os
Chain ID: nuahchain-1
Home directory: /Users/username/.nuahd

Продолжить с продакшен настройками? (yes/no):
```

**ВАЖНО**: Введите `yes` только если вы понимаете последствия!

## 📁 Создаваемые файлы

### Файлы безопасности
- `public_keys_registry.json` - Структурированный реестр публичных ключей
- `public_keys_registry_readable.txt` - Человекочитаемая версия
- `multisig_config.json` - Конфигурация мультисиг кошельков
- `genesis_backup_YYYYMMDD_HHMMSS/` - Резервные копии

### Продакшен скрипты
- `start_production_node.sh` - Безопасный запуск ноды
- `monitor_node.sh` - Мониторинг состояния ноды

## 🔑 Управление ключами

### Структура ключей
```json
{
  "chain_id": "nuahchain-1",
  "export_timestamp": "2024-01-15T10:30:00Z",
  "keys": [
    {
      "name": "validator",
      "role": "validator",
      "description": "Основной валидатор сети",
      "address": "nuah1...",
      "public_key": "...",
      "public_key_type": "/cosmos.crypto.secp256k1.PubKey"
    }
  ]
}
```

### Роли и распределение токенов
| Роль | Количество | Процент | Описание |
|------|------------|---------|----------|
| Treasury | 30M NUAH | 30% | Казначейство проекта |
| Community | 25M NUAH | 25% | Средства сообщества |
| Foundation | 20M NUAH | 20% | Фонд развития |
| Ecosystem | 15M NUAH | 15% | Развитие экосистемы |
| Validator | 5M NUAH | 5% | Валидатор сети |
| Team | 5M NUAH | 5% | Команда разработчиков |

### Мультисиг конфигурация
- **Treasury Multisig**: Требует 2 из 3 подписей (foundation, treasury, validator)
- **Критические операции**: Все крупные транзакции должны проходить через мультисиг

## 🛡️ Безопасные практики

### 1. Хранение ключей
```bash
# Ключи хранятся в OS keyring
# macOS: Keychain Access
# Linux: gnome-keyring или kde-wallet
# Windows: Windows Credential Store

# Проверка ключей
./build/nuahd keys list --keyring-backend os
```

### 2. Резервное копирование
```bash
# Автоматические резервные копии создаются в:
ls -la genesis_backup_*/

# Дополнительное резервное копирование
cp -r ~/.nuahd ~/nuahd_backup_$(date +%Y%m%d)
```

### 3. Валидация
```bash
# Проверка genesis файла
./build/nuahd validate-genesis

# Проверка конфигурации
./build/nuahd config validate
```

## 🚀 Запуск продакшен ноды

### 1. Запуск ноды
```bash
./start_production_node.sh
```

### 2. Мониторинг
```bash
# Проверка статуса
./monitor_node.sh

# Просмотр логов
tail -f nuahd_production.log

# Проверка синхронизации
curl -s http://127.0.0.1:26657/status | jq '.result.sync_info'
```

### 3. Остановка ноды
```bash
# Найти PID
cat nuahd.pid

# Остановить ноду
kill $(cat nuahd.pid)
```

## 🔍 Проверка токеномики

### Проверка балансов
```bash
# Проверка всех балансов
for role in validator foundation community treasury ecosystem team; do
  addr=$(./build/nuahd keys show $role --keyring-backend os -a)
  echo "$role: $(./build/nuahd query bank balances $addr --node http://127.0.0.1:26657)"
done
```

### Проверка общего предложения
```bash
./build/nuahd query bank total --node http://127.0.0.1:26657
```

### Проверка мультисига
```bash
# Информация о мультисиг адресе
./build/nuahd keys show treasury_multisig --keyring-backend os

# Проверка баланса мультисига
multisig_addr=$(./build/nuahd keys show treasury_multisig --keyring-backend os -a)
./build/nuahd query bank balances $multisig_addr --node http://127.0.0.1:26657
```

## 🔧 Конфигурация

### Сетевые настройки
```toml
# config.toml
[rpc]
laddr = "tcp://127.0.0.1:26657"  # Только локальный доступ
cors_allowed_origins = ["https://wallet.nuah.io"]  # Только доверенные домены

[p2p]
max_num_inbound_peers = 40
max_num_outbound_peers = 10
```

### API настройки
```toml
# app.toml
[api]
enable = true
swagger = false  # Отключено для безопасности
address = "tcp://127.0.0.1:1317"
enabled-unsafe-cors = false  # Безопасный CORS
```

## 🚨 Аварийные процедуры

### Восстановление из резервной копии
```bash
# Остановить ноду
kill $(cat nuahd.pid)

# Восстановить из резервной копии
rm -rf ~/.nuahd
cp -r genesis_backup_YYYYMMDD_HHMMSS/.nuahd ~/

# Перезапустить ноду
./start_production_node.sh
```

### Восстановление ключей
```bash
# Импорт ключа из мнемоники
./build/nuahd keys add validator --recover --keyring-backend os

# Проверка восстановленного ключа
./build/nuahd keys show validator --keyring-backend os
```

## 📊 Мониторинг и алерты

### Основные метрики
- Высота блока
- Статус синхронизации
- Количество пиров
- Использование ресурсов

### Настройка алертов
```bash
# Пример скрипта мониторинга
#!/bin/bash
HEIGHT=$(curl -s http://127.0.0.1:26657/status | jq -r '.result.sync_info.latest_block_height')
if [ "$HEIGHT" -lt 100 ]; then
  echo "ALERT: Block height too low: $HEIGHT"
fi
```

## 🔍 Мониторинг и аудит

### Аудит безопасности
```bash
# Запуск полного аудита безопасности
./scripts/security_audit.sh

# Аудит проверяет:
# - Безопасность keyring backend
# - Конфигурацию безопасности
# - Валидность genesis файла
# - Файловые разрешения
# - Сетевую безопасность
# - Наличие резервных копий
# - Статус процессов
# - Анализ логов
```

### Мониторинг ноды
```bash
# Запуск мониторинга
./monitor_node.sh

# Проверка статуса
curl -s http://localhost:26657/status | jq .

# Проверка синхронизации
curl -s http://localhost:26657/status | jq .result.sync_info
```

### Проверка токеномики
```bash
# Запуск проверки
./check_tokenomics.sh

# Проверка балансов
./build/nuahd query bank balances $(./build/nuahd keys show validator --address --keyring-backend os)
```

## 🔐 Аудит безопасности

### Чеклист безопасности
- [ ] Используется OS keyring backend
- [ ] Genesis файл прошел валидацию
- [ ] Отключены небезопасные CORS настройки
- [ ] Созданы мультисиг кошельки
- [ ] Настроено резервное копирование
- [ ] Ограничен сетевой доступ
- [ ] Настроен мониторинг

### Регулярные проверки
```bash
# Еженедельная проверка
./scripts/security_audit.sh

# Проверка целостности ключей
./build/nuahd keys list --keyring-backend os

# Проверка конфигурации
grep -E "(cors|unsafe)" ~/.nuahd/config/*.toml
```

## 📞 Поддержка

При возникновении проблем:
1. Проверьте логи: `tail -f nuahd_production.log`
2. Проверьте статус: `./monitor_node.sh`
3. Проверьте конфигурацию: `./build/nuahd config validate`
4. Обратитесь к команде разработчиков

## ⚖️ Лицензия

Этот гайд и скрипты распространяются под той же лицензией, что и основной проект NUAH.

# Authz & Feegrant Setup Guide

Руководство по настройке делегирования полномочий и управления комиссиями для AI Trader Bot.

## 🎯 Обзор

Модули `authz` и `feegrant` позволяют пользователям делегировать торговые полномочия AI-агенту и предоставлять ему возможность оплачивать комиссии за транзакции.

## 🔐 Принципы работы

### Authz (Authorization)
- Пользователь выдает разрешения AI-агенту на выполнение определенных операций
- AI-агент может выполнять операции от имени пользователя
- Полномочия можно ограничить по времени, количеству операций или сумме

### Feegrant (Fee Grant)
- Пользователь выделяет бюджет для оплаты комиссий
- AI-агент может использовать этот бюджет для оплаты комиссий
- Бюджет можно ограничить по сумме и времени

## 🚀 Быстрая настройка

### Использование интерактивного скрипта

```bash
# Запуск интерактивного скрипта
./scripts/ai_trader_grant.sh
```

Скрипт проведет вас через процесс:
1. Ввод адресов пользователя и AI-агента
2. Настройка торговых лимитов
3. Настройка бюджета на комиссии
4. Генерация команд для выполнения

### Ручная настройка

#### Шаг 1: Получение адресов

```bash
# Адрес пользователя
USER_ADDRESS=$(nuahd keys show <user-key-name> -a)

# Адрес AI-агента
AI_AGENT_ADDRESS=$(nuahd keys show <ai-agent-key-name> -a)

echo "User address: $USER_ADDRESS"
echo "AI Agent address: $AI_AGENT_ADDRESS"
```

#### Шаг 2: Выдача торговых полномочий

```bash
# Разрешение на покупку активов
nuahd tx authz grant $AI_AGENT_ADDRESS generic \
  --msg-type="/nuah.assets.MsgBuyAsset" \
  --from <user-key-name> \
  --chain-id nuah-testnet \
  --gas auto \
  --gas-adjustment 1.5 \
  --fees 1000factory/test/ndollar

# Разрешение на продажу активов
nuahd tx authz grant $AI_AGENT_ADDRESS generic \
  --msg-type="/nuah.assets.MsgSellAsset" \
  --from <user-key-name> \
  --chain-id nuah-testnet \
  --gas auto \
  --gas-adjustment 1.5 \
  --fees 1000factory/test/ndollar
```

#### Шаг 3: Выдача полномочий на оплату комиссий

```bash
# Создание fee grant
nuahd tx feegrant grant $AI_AGENT_ADDRESS $USER_ADDRESS \
  --spend-limit 10000factory/test/ndollar \
  --expiration 2025-12-31T23:59:59Z \
  --from <user-key-name> \
  --chain-id nuah-testnet \
  --gas auto \
  --gas-adjustment 1.5 \
  --fees 1000factory/test/ndollar
```

## ⚙️ Расширенная настройка

### Ограниченные полномочия

#### Временные ограничения

```bash
# Полномочия с истечением через 30 дней
nuahd tx authz grant $AI_AGENT_ADDRESS generic \
  --msg-type="/nuah.assets.MsgBuyAsset" \
  --expiration $(date -d "+30 days" -u +"%Y-%m-%dT%H:%M:%SZ") \
  --from <user-key-name> \
  --chain-id nuah-testnet
```

#### Количественные ограничения

```bash
# Полномочия с ограничением по количеству операций
nuahd tx authz grant $AI_AGENT_ADDRESS generic \
  --msg-type="/nuah.assets.MsgBuyAsset" \
  --max-msgs 100 \
  --from <user-key-name> \
  --chain-id nuah-testnet
```

### Условные полномочия

#### Ограничение по сумме транзакции

```bash
# Создание условного разрешения
nuahd tx authz grant $AI_AGENT_ADDRESS generic \
  --msg-type="/nuah.assets.MsgBuyAsset" \
  --allow-list "nuah1..." \
  --from <user-key-name> \
  --chain-id nuah-testnet
```

### Типы Fee Grant

#### Basic Fee Grant

```bash
# Простой fee grant с лимитом
nuahd tx feegrant grant $AI_AGENT_ADDRESS $USER_ADDRESS \
  --spend-limit 5000factory/test/ndollar \
  --from <user-key-name> \
  --chain-id nuah-testnet
```

#### Periodic Fee Grant

```bash
# Периодический fee grant (ежедневно)
nuahd tx feegrant grant $AI_AGENT_ADDRESS $USER_ADDRESS \
  --spend-limit 1000factory/test/ndollar \
  --period 86400 \
  --period-spend-limit 100factory/test/ndollar \
  --from <user-key-name> \
  --chain-id nuah-testnet
```

#### Allowed Messages Fee Grant

```bash
# Fee grant только для определенных типов сообщений
nuahd tx feegrant grant $AI_AGENT_ADDRESS $USER_ADDRESS \
  --spend-limit 2000factory/test/ndollar \
  --allowed-messages "/nuah.assets.MsgBuyAsset,/nuah.assets.MsgSellAsset" \
  --from <user-key-name> \
  --chain-id nuah-testnet
```

## 🔍 Проверка и мониторинг

### Проверка выданных полномочий

```bash
# Список всех полномочий пользователя
nuahd query authz grants $USER_ADDRESS

# Полномочия конкретного грантера
nuahd query authz grants $USER_ADDRESS $AI_AGENT_ADDRESS

# Полномочия по типу сообщения
nuahd query authz grants $USER_ADDRESS $AI_AGENT_ADDRESS --msg-type="/nuah.assets.MsgBuyAsset"
```

### Проверка Fee Grants

```bash
# Список всех fee grants пользователя
nuahd query feegrant grants $USER_ADDRESS

# Fee grants от конкретного грантера
nuahd query feegrant grants $USER_ADDRESS $AI_AGENT_ADDRESS

# Детали конкретного fee grant
nuahd query feegrant grant $AI_AGENT_ADDRESS $USER_ADDRESS
```

### Мониторинг использования

```bash
# Проверка оставшегося лимита
nuahd query feegrant grant $AI_AGENT_ADDRESS $USER_ADDRESS --output json | jq '.allowance.spend_limit'

# История транзакций AI-агента
nuahd query txs --events "message.sender=$AI_AGENT_ADDRESS" --limit 10
```

## 🔄 Управление полномочиями

### Обновление полномочий

```bash
# Отзыв старых полномочий
nuahd tx authz revoke $AI_AGENT_ADDRESS /nuah.assets.MsgBuyAsset \
  --from <user-key-name> \
  --chain-id nuah-testnet

# Выдача новых полномочий
nuahd tx authz grant $AI_AGENT_ADDRESS generic \
  --msg-type="/nuah.assets.MsgBuyAsset" \
  --spend-limit 2000factory/test/ndollar \
  --from <user-key-name> \
  --chain-id nuah-testnet
```

### Обновление Fee Grant

```bash
# Отзыв старого fee grant
nuahd tx feegrant revoke $AI_AGENT_ADDRESS $USER_ADDRESS \
  --from <user-key-name> \
  --chain-id nuah-testnet

# Создание нового fee grant
nuahd tx feegrant grant $AI_AGENT_ADDRESS $USER_ADDRESS \
  --spend-limit 15000factory/test/ndollar \
  --from <user-key-name> \
  --chain-id nuah-testnet
```

## 🛡️ Безопасность

### Рекомендации по безопасности

1. **Ограничение полномочий**:
   - Используйте минимально необходимые полномочия
   - Устанавливайте временные ограничения
   - Ограничивайте суммы транзакций

2. **Мониторинг**:
   - Регулярно проверяйте использование полномочий
   - Настройте алерты на подозрительную активность
   - Ведите журнал всех операций

3. **Резервные планы**:
   - Подготовьте команды для быстрого отзыва полномочий
   - Имейте план действий в случае компрометации
   - Регулярно тестируйте процедуры восстановления

### Команды экстренного отзыва

```bash
# Экстренный отзыв всех полномочий
nuahd tx authz revoke $AI_AGENT_ADDRESS /nuah.assets.MsgBuyAsset \
  --from <user-key-name> \
  --chain-id nuah-testnet

nuahd tx authz revoke $AI_AGENT_ADDRESS /nuah.assets.MsgSellAsset \
  --from <user-key-name> \
  --chain-id nuah-testnet

# Отзыв fee grant
nuahd tx feegrant revoke $AI_AGENT_ADDRESS $USER_ADDRESS \
  --from <user-key-name> \
  --chain-id nuah-testnet
```

## 📊 Примеры конфигураций

### Консервативная настройка

```bash
# Ограниченные полномочия с низкими лимитами
nuahd tx authz grant $AI_AGENT_ADDRESS generic \
  --msg-type="/nuah.assets.MsgBuyAsset" \
  --spend-limit 1000factory/test/ndollar \
  --max-msgs 10 \
  --expiration $(date -d "+7 days" -u +"%Y-%m-%dT%H:%M:%SZ") \
  --from <user-key-name>

# Небольшой fee grant
nuahd tx feegrant grant $AI_AGENT_ADDRESS $USER_ADDRESS \
  --spend-limit 500factory/test/ndollar \
  --from <user-key-name>
```

### Агрессивная настройка

```bash
# Широкие полномочия с высокими лимитами
nuahd tx authz grant $AI_AGENT_ADDRESS generic \
  --msg-type="/nuah.assets.MsgBuyAsset" \
  --spend-limit 50000factory/test/ndollar \
  --max-msgs 1000 \
  --expiration $(date -d "+90 days" -u +"%Y-%m-%dT%H:%M:%SZ") \
  --from <user-key-name>

# Большой fee grant с периодическим пополнением
nuahd tx feegrant grant $AI_AGENT_ADDRESS $USER_ADDRESS \
  --spend-limit 10000factory/test/ndollar \
  --period 86400 \
  --period-spend-limit 1000factory/test/ndollar \
  --from <user-key-name>
```

### Настройка для тестирования

```bash
# Полномочия только для тестовой сети
nuahd tx authz grant $AI_AGENT_ADDRESS generic \
  --msg-type="/nuah.assets.MsgBuyAsset" \
  --spend-limit 100factory/test/ndollar \
  --max-msgs 5 \
  --expiration $(date -d "+1 day" -u +"%Y-%m-%dT%H:%M:%SZ") \
  --from <user-key-name> \
  --chain-id nuah-testnet

# Минимальный fee grant для тестов
nuahd tx feegrant grant $AI_AGENT_ADDRESS $USER_ADDRESS \
  --spend-limit 50factory/test/ndollar \
  --from <user-key-name> \
  --chain-id nuah-testnet
```

## 🧪 Тестирование

### Проверка полномочий

```bash
# Тестовая транзакция от имени AI-агента
nuahd tx assets buy-asset \
  --from $AI_AGENT_ADDRESS \
  --granter $USER_ADDRESS \
  --amount 100factory/test/ndollar \
  --chain-id nuah-testnet
```

### Проверка Fee Grant

```bash
# Проверка использования fee grant
nuahd query feegrant grant $AI_AGENT_ADDRESS $USER_ADDRESS --output json | jq '.allowance'
```

## 🆘 Устранение неполадок

### Частые проблемы

1. **"insufficient funds"**:
   - Проверьте баланс пользователя
   - Убедитесь, что fee grant активен
   - Проверьте лимиты fee grant

2. **"unauthorized"**:
   - Проверьте, что полномочия выданы
   - Убедитесь, что полномочия не истекли
   - Проверьте тип сообщения

3. **"fee grant not found"**:
   - Проверьте, что fee grant создан
   - Убедитесь в правильности адресов
   - Проверьте, что fee grant не истек

### Диагностические команды

```bash
# Проверка статуса аккаунта
nuahd query account $USER_ADDRESS
nuahd query account $AI_AGENT_ADDRESS

# Проверка баланса
nuahd query bank balances $USER_ADDRESS
nuahd query bank balances $AI_AGENT_ADDRESS

# Проверка полномочий
nuahd query authz grants $USER_ADDRESS $AI_AGENT_ADDRESS

# Проверка fee grants
nuahd query feegrant grants $USER_ADDRESS
```

---

**⚠️ Важно**: Всегда тестируйте настройки на тестовой сети перед применением в продакшене. Регулярно проверяйте и обновляйте полномочия в соответствии с изменяющимися потребностями.


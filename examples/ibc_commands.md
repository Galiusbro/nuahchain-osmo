# IBC Commands Reference / Справочник команд IBC

This document contains common IBC commands for Hermes relayer.
Этот документ содержит основные команды IBC для релейера Hermes.

## Configuration Commands / Команды конфигурации

### Health Check / Проверка состояния
```bash
# Check configuration and connectivity
# Проверка конфигурации и подключения
hermes --config ~/.hermes_test/config.toml health-check

# Check specific chain
# Проверка конкретной сети
hermes --config ~/.hermes_test/config.toml health-check --chain nuahchain-1
```

### Key Management / Управление ключами

```bash
# Add key from mnemonic
# Добавление ключа из мнемонической фразы
echo "your mnemonic phrase here" | hermes --config ~/.hermes_test/config.toml keys add --chain nuahchain-1 --mnemonic-file /dev/stdin

# List keys for a chain
# Список ключей для сети
hermes --config ~/.hermes_test/config.toml keys list --chain nuahchain-1

# Delete key
# Удаление ключа
hermes --config ~/.hermes_test/config.toml keys delete --chain nuahchain-1 --key-name test-relayer
```

## Client Commands / Команды клиентов

### Create Clients / Создание клиентов

```bash
# Create client for NUAH on Osmosis
# Создание клиента для NUAH на Osmosis
hermes --config ~/.hermes_test/config.toml create client --host-chain osmo-test-5 --reference-chain nuahchain-1

# Create client for Osmosis on NUAH
# Создание клиента для Osmosis на NUAH
hermes --config ~/.hermes_test/config.toml create client --host-chain nuahchain-1 --reference-chain osmo-test-5
```

### Update Clients / Обновление клиентов

```bash
# Update client
# Обновление клиента
hermes --config ~/.hermes_test/config.toml update client --host-chain osmo-test-5 --client-id 07-tendermint-0

# Update all clients
# Обновление всех клиентов
hermes --config ~/.hermes_test/config.toml update clients --host-chain osmo-test-5
```

### Query Clients / Запрос клиентов

```bash
# List all clients on a chain
# Список всех клиентов в сети
hermes --config ~/.hermes_test/config.toml query clients --host-chain osmo-test-5

# Query specific client
# Запрос конкретного клиента
hermes --config ~/.hermes_test/config.toml query client state --chain osmo-test-5 --client-id 07-tendermint-0
```

## Connection Commands / Команды соединений

### Create Connection / Создание соединения

```bash
# Create connection between chains
# Создание соединения между сетями
hermes --config ~/.hermes_test/config.toml create connection --a-chain nuahchain-1 --b-chain osmo-test-5

# Create connection with specific client IDs
# Создание соединения с конкретными ID клиентов
hermes --config ~/.hermes_test/config.toml create connection --a-chain nuahchain-1 --a-client 07-tendermint-0 --b-client 07-tendermint-0
```

### Query Connections / Запрос соединений

```bash
# List all connections on a chain
# Список всех соединений в сети
hermes --config ~/.hermes_test/config.toml query connections --chain nuahchain-1

# Query specific connection
# Запрос конкретного соединения
hermes --config ~/.hermes_test/config.toml query connection end --chain nuahchain-1 --connection-id connection-0
```

## Channel Commands / Команды каналов

### Create Channel / Создание канала

```bash
# Create transfer channel
# Создание канала для трансферов
hermes --config ~/.hermes_test/config.toml create channel --a-chain nuahchain-1 --a-connection connection-0 --a-port transfer --b-port transfer

# Create channel with specific version
# Создание канала с конкретной версией
hermes --config ~/.hermes_test/config.toml create channel --a-chain nuahchain-1 --a-connection connection-0 --a-port transfer --b-port transfer --channel-version ics20-1
```

### Query Channels / Запрос каналов

```bash
# List all channels on a chain
# Список всех каналов в сети
hermes --config ~/.hermes_test/config.toml query channels --chain nuahchain-1

# Query specific channel
# Запрос конкретного канала
hermes --config ~/.hermes_test/config.toml query channel end --chain nuahchain-1 --port transfer --channel-id channel-0
```

## Packet Commands / Команды пакетов

### Clear Packets / Очистка пакетов

```bash
# Clear pending packets on a channel
# Очистка ожидающих пакетов в канале
hermes --config ~/.hermes_test/config.toml clear packets --chain nuahchain-1 --port transfer --channel-id channel-0

# Clear packets in both directions
# Очистка пакетов в обоих направлениях
hermes --config ~/.hermes_test/config.toml clear packets --chain nuahchain-1 --port transfer --channel-id channel-0 --counterparty-chain osmo-test-5
```

### Query Packets / Запрос пакетов

```bash
# Query pending packets
# Запрос ожидающих пакетов
hermes --config ~/.hermes_test/config.toml query packet pending --chain nuahchain-1 --port transfer --channel-id channel-0

# Query packet commitments
# Запрос обязательств пакетов
hermes --config ~/.hermes_test/config.toml query packet commitments --chain nuahchain-1 --port transfer --channel-id channel-0
```

## Transfer Commands / Команды трансферов

### IBC Transfer / IBC трансфер

```bash
# Transfer tokens from NUAH to Osmosis
# Трансфер токенов с NUAH на Osmosis
hermes --config ~/.hermes_test/config.toml tx ft-transfer --dst-chain osmo-test-5 --src-chain nuahchain-1 --src-port transfer --src-channel channel-0 --amount 1000 --denom unuah --receiver osmo19rl4cm2hmr8afy4kldpxz3fka4jguq0a5m7df8

# Transfer with timeout
# Трансфер с таймаутом
hermes --config ~/.hermes_test/config.toml tx ft-transfer --dst-chain osmo-test-5 --src-chain nuahchain-1 --src-port transfer --src-channel channel-0 --amount 1000 --denom unuah --receiver osmo19rl4cm2hmr8afy4kldpxz3fka4jguq0a5m7df8 --timeout-height-offset 1000
```

## Relayer Operations / Операции релейера

### Start Relayer / Запуск релейера

```bash
# Start relaying for all configured chains
# Запуск релейинга для всех настроенных сетей
hermes --config ~/.hermes_test/config.toml start

# Start relaying for specific chains
# Запуск релейинга для конкретных сетей
hermes --config ~/.hermes_test/config.toml start --chains nuahchain-1,osmo-test-5
```

### Listen Mode / Режим прослушивания

```bash
# Listen for events on a specific chain
# Прослушивание событий в конкретной сети
hermes --config ~/.hermes_test/config.toml listen --chain nuahchain-1

# Listen for events on all chains
# Прослушивание событий во всех сетях
hermes --config ~/.hermes_test/config.toml listen
```

## Monitoring Commands / Команды мониторинга

### Query Chain Status / Запрос статуса сети

```bash
# Query chain status
# Запрос статуса сети
hermes --config ~/.hermes_test/config.toml query chain status --chain nuahchain-1

# Query balance
# Запрос баланса
hermes --config ~/.hermes_test/config.toml query balance --chain nuahchain-1
```

### Evidence Commands / Команды доказательств

```bash
# Submit misbehaviour evidence
# Отправка доказательств неправильного поведения
hermes --config ~/.hermes_test/config.toml misbehaviour --chain nuahchain-1 --client-id 07-tendermint-0
```

## Troubleshooting Commands / Команды для устранения неполадок

### Debug Information / Отладочная информация

```bash
# Get version information
# Получение информации о версии
hermes version

# Validate configuration
# Проверка конфигурации
hermes --config ~/.hermes_test/config.toml config validate

# Test connectivity
# Тестирование подключения
hermes --config ~/.hermes_test/config.toml health-check --verbose
```

### Reset Commands / Команды сброса

```bash
# Clear all packets (use with caution)
# Очистка всех пакетов (используйте осторожно)
hermes --config ~/.hermes_test/config.toml clear packets --chain nuahchain-1 --port transfer --channel-id channel-0

# Upgrade client
# Обновление клиента
hermes --config ~/.hermes_test/config.toml upgrade client --host-chain osmo-test-5 --client-id 07-tendermint-0 --upgrade-height 1000
```

## Common Flags / Общие флаги

- `--config`: Path to configuration file / Путь к файлу конфигурации
- `--verbose`: Enable verbose output / Включить подробный вывод
- `--json`: Output in JSON format / Вывод в формате JSON
- `--help`: Show help information / Показать справочную информацию

## Environment Variables / Переменные окружения

```bash
# Set log level
# Установка уровня логирования
export RUST_LOG=info

# Set configuration path
# Установка пути к конфигурации
export HERMES_CONFIG_PATH=~/.hermes_test/config.toml
```

## Tips / Советы

1. Always check chain connectivity before creating clients
   Всегда проверяйте подключение к сетям перед созданием клиентов

2. Ensure sufficient balance for gas fees
   Убедитесь в достаточном балансе для оплаты газа

3. Monitor logs for errors and warnings
   Отслеживайте логи на предмет ошибок и предупреждений

4. Use `--dry-run` flag to test commands without execution
   Используйте флаг `--dry-run` для тестирования команд без выполнения

5. Keep clients updated to prevent expiration
   Поддерживайте клиенты обновленными, чтобы предотвратить истечение срока действия

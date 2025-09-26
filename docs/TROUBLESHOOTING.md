# IBC Relayer Troubleshooting Guide / Руководство по устранению неполадок IBC релейера

This guide helps resolve common issues when setting up and running Hermes IBC relayer.
Это руководство поможет решить распространенные проблемы при настройке и запуске IBC релейера Hermes.

## Table of Contents / Содержание

1. [Installation Issues / Проблемы установки](#installation-issues--проблемы-установки)
2. [Configuration Issues / Проблемы конфигурации](#configuration-issues--проблемы-конфигурации)
3. [Connection Issues / Проблемы подключения](#connection-issues--проблемы-подключения)
4. [Key Management Issues / Проблемы управления ключами](#key-management-issues--проблемы-управления-ключами)
5. [Client Creation Issues / Проблемы создания клиентов](#client-creation-issues--проблемы-создания-клиентов)
6. [Channel and Connection Issues / Проблемы каналов и соединений](#channel-and-connection-issues--проблемы-каналов-и-соединений)
7. [Transfer Issues / Проблемы трансферов](#transfer-issues--проблемы-трансферов)
8. [Performance Issues / Проблемы производительности](#performance-issues--проблемы-производительности)
9. [Monitoring and Debugging / Мониторинг и отладка](#monitoring-and-debugging--мониторинг-и-отладка)

---

## Installation Issues / Проблемы установки

### Problem: Hermes installation fails / Проблема: Не удается установить Hermes

**Symptoms / Симптомы:**
```bash
error: failed to compile `ibc-relayer-cli`
```

**Solutions / Решения:**

1. **Update Rust / Обновите Rust:**
   ```bash
   rustup update
   rustc --version  # Should be 1.70.0 or later / Должна быть 1.70.0 или новее
   ```

2. **Install with specific features / Установка с конкретными функциями:**
   ```bash
   cargo install ibc-relayer-cli --bin hermes --locked --features=telemetry
   ```

3. **Use pre-compiled binary / Используйте предварительно скомпилированный бинарник:**
   ```bash
   # For macOS ARM64
   wget https://github.com/informalsystems/hermes/releases/download/v1.7.4/hermes-v1.7.4-aarch64-apple-darwin.tar.gz
   tar -xzf hermes-v1.7.4-aarch64-apple-darwin.tar.gz
   sudo mv hermes /usr/local/bin/
   ```

### Problem: Permission denied / Проблема: Отказано в доступе

**Symptoms / Симптомы:**
```bash
Permission denied (os error 13)
```

**Solutions / Решения:**
```bash
# Make binary executable / Сделайте бинарник исполняемым
chmod +x /usr/local/bin/hermes

# Or install to user directory / Или установите в пользовательскую директорию
cargo install ibc-relayer-cli --bin hermes --locked --root ~/.local
export PATH="$HOME/.local/bin:$PATH"
```

---

## Configuration Issues / Проблемы конфигурации

### Problem: Invalid configuration / Проблема: Неверная конфигурация

**Symptoms / Симптомы:**
```bash
error: invalid configuration file
```

**Solutions / Решения:**

1. **Validate configuration / Проверьте конфигурацию:**
   ```bash
   hermes --config ~/.hermes_test/config.toml config validate
   ```

2. **Check TOML syntax / Проверьте синтаксис TOML:**
   ```bash
   # Use online TOML validator or
   # Используйте онлайн валидатор TOML или
   python3 -c "import toml; toml.load('~/.hermes_test/config.toml')"
   ```

3. **Common configuration errors / Распространенные ошибки конфигурации:**
   - Missing quotes around strings / Отсутствующие кавычки вокруг строк
   - Incorrect indentation / Неправильные отступы
   - Invalid URLs / Неверные URL
   - Wrong chain IDs / Неправильные ID сетей

### Problem: Chain configuration not found / Проблема: Конфигурация сети не найдена

**Symptoms / Симптомы:**
```bash
error: chain 'nuahchain-1' not found in configuration
```

**Solutions / Решения:**

1. **Check chain ID in config / Проверьте ID сети в конфиге:**
   ```bash
   grep -n "id = " ~/.hermes_test/config.toml
   ```

2. **Verify chain is running / Убедитесь, что сеть запущена:**
   ```bash
   curl -s http://localhost:26657/status | jq '.result.node_info.network'
   ```

---

## Connection Issues / Проблемы подключения

### Problem: RPC connection failed / Проблема: Не удается подключиться к RPC

**Symptoms / Симптомы:**
```bash
error: RPC error: connection refused
```

**Solutions / Решения:**

1. **Check if node is running / Проверьте, запущен ли узел:**
   ```bash
   # For local node / Для локального узла
   ps aux | grep nuahd

   # Check port availability / Проверьте доступность порта
   netstat -an | grep 26657
   ```

2. **Test RPC connectivity / Тестируйте RPC подключение:**
   ```bash
   curl -s http://localhost:26657/health
   curl -s http://localhost:26657/status
   ```

3. **Check firewall settings / Проверьте настройки файрвола:**
   ```bash
   # macOS
   sudo pfctl -sr | grep 26657

   # Linux
   sudo ufw status | grep 26657
   ```

### Problem: gRPC connection failed / Проблема: Не удается подключиться к gRPC

**Symptoms / Симптомы:**
```bash
error: gRPC error: connection refused
```

**Solutions / Решения:**

1. **Check gRPC endpoint / Проверьте gRPC эндпоинт:**
   ```bash
   # Test gRPC connectivity / Тестируйте gRPC подключение
   grpcurl -plaintext localhost:9090 list
   ```

2. **Verify gRPC is enabled in node config / Убедитесь, что gRPC включен в конфиге узла:**
   ```bash
   grep -A 5 "\[grpc\]" ~/.nuahd/config/app.toml
   ```

### Problem: WebSocket connection failed / Проблема: Не удается подключиться к WebSocket

**Symptoms / Симптомы:**
```bash
error: WebSocket connection failed
```

**Solutions / Решения:**

1. **Test WebSocket connection / Тестируйте WebSocket подключение:**
   ```bash
   # Using websocat (install with: cargo install websocat)
   echo '{"jsonrpc":"2.0","method":"subscribe","params":["tm.event='\''NewBlock'\''"],"id":1}' | websocat ws://localhost:26657/websocket
   ```

2. **Check Tendermint config / Проверьте конфиг Tendermint:**
   ```bash
   grep -A 5 "\[rpc\]" ~/.nuahd/config/config.toml
   ```

---

## Key Management Issues / Проблемы управления ключами

### Problem: Key already exists / Проблема: Ключ уже существует

**Symptoms / Симптомы:**
```bash
error: key with name 'test-relayer' already exists
```

**Solutions / Решения:**

1. **List existing keys / Список существующих ключей:**
   ```bash
   hermes --config ~/.hermes_test/config.toml keys list --chain nuahchain-1
   ```

2. **Delete existing key / Удалите существующий ключ:**
   ```bash
   hermes --config ~/.hermes_test/config.toml keys delete --chain nuahchain-1 --key-name test-relayer
   ```

3. **Use different key name / Используйте другое имя ключа:**
   ```bash
   # Update config.toml with new key name
   # Обновите config.toml с новым именем ключа
   key_name = 'relayer-v2'
   ```

### Problem: Invalid mnemonic / Проблема: Неверная мнемоническая фраза

**Symptoms / Симптомы:**
```bash
error: invalid mnemonic phrase
```

**Solutions / Решения:**

1. **Verify mnemonic format / Проверьте формат мнемоники:**
   - Should be 12 or 24 words / Должно быть 12 или 24 слова
   - Words separated by spaces / Слова разделены пробелами
   - No extra characters / Никаких лишних символов

2. **Test with standard mnemonic / Тестируйте со стандартной мнемоникой:**
   ```bash
   echo "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about" | hermes --config ~/.hermes_test/config.toml keys add --chain nuahchain-1 --mnemonic-file /dev/stdin
   ```

### Problem: Insufficient funds / Проблема: Недостаточно средств

**Symptoms / Симптомы:**
```bash
error: insufficient funds for gas
```

**Solutions / Решения:**

1. **Check relayer balance / Проверьте баланс релейера:**
   ```bash
   # Get relayer address / Получите адрес релейера
   RELAYER_ADDR=$(hermes --config ~/.hermes_test/config.toml keys list --chain nuahchain-1 | grep -o 'nuah[a-z0-9]*')

   # Check balance / Проверьте баланс
   ./build/nuahd query bank balances $RELAYER_ADDR
   ```

2. **Fund relayer account / Пополните счет релейера:**
   ```bash
   # For local testnet / Для локального тестнета
   ./build/nuahd tx bank send validator $RELAYER_ADDR 1000000unuah --keyring-backend test --chain-id nuahchain-1 --yes

   # For Osmosis testnet, use faucet / Для тестнета Osmosis используйте faucet
   # https://faucet.testnet.osmosis.zone/
   ```

---

## Client Creation Issues / Проблемы создания клиентов

### Problem: Client creation fails / Проблема: Не удается создать клиент

**Symptoms / Симптомы:**
```bash
error: failed to create client
```

**Solutions / Решения:**

1. **Check chain synchronization / Проверьте синхронизацию сетей:**
   ```bash
   # Check if chains are synced / Проверьте, синхронизированы ли сети
   hermes --config ~/.hermes_test/config.toml query chain status --chain nuahchain-1
   hermes --config ~/.hermes_test/config.toml query chain status --chain osmo-test-5
   ```

2. **Verify client parameters / Проверьте параметры клиента:**
   ```bash
   # Check trusting period vs unbonding period
   # Проверьте период доверия против периода разблокировки
   hermes --config ~/.hermes_test/config.toml query staking params --chain nuahchain-1
   ```

3. **Increase gas limits / Увеличьте лимиты газа:**
   ```toml
   # In config.toml
   default_gas = 200000
   max_gas = 800000
   ```

### Problem: Client already exists / Проблема: Клиент уже существует

**Symptoms / Симптомы:**
```bash
error: client already exists
```

**Solutions / Решения:**

1. **List existing clients / Список существующих клиентов:**
   ```bash
   hermes --config ~/.hermes_test/config.toml query clients --host-chain osmo-test-5
   ```

2. **Use existing client / Используйте существующий клиент:**
   ```bash
   # Note the client ID and use it for connections
   # Запомните ID клиента и используйте его для соединений
   ```

---

## Channel and Connection Issues / Проблемы каналов и соединений

### Problem: Connection handshake fails / Проблема: Не удается выполнить рукопожатие соединения

**Symptoms / Симптомы:**
```bash
error: connection handshake failed
```

**Solutions / Решения:**

1. **Check client states / Проверьте состояния клиентов:**
   ```bash
   hermes --config ~/.hermes_test/config.toml query client state --chain nuahchain-1 --client-id 07-tendermint-0
   ```

2. **Update clients if needed / Обновите клиенты при необходимости:**
   ```bash
   hermes --config ~/.hermes_test/config.toml update client --host-chain nuahchain-1 --client-id 07-tendermint-0
   ```

3. **Check for misbehaviour / Проверьте на неправильное поведение:**
   ```bash
   hermes --config ~/.hermes_test/config.toml query client consensus --chain nuahchain-1 --client-id 07-tendermint-0
   ```

### Problem: Channel creation fails / Проблема: Не удается создать канал

**Symptoms / Симптомы:**
```bash
error: channel creation failed
```

**Solutions / Решения:**

1. **Verify connection exists / Убедитесь, что соединение существует:**
   ```bash
   hermes --config ~/.hermes_test/config.toml query connections --chain nuahchain-1
   ```

2. **Check port availability / Проверьте доступность порта:**
   ```bash
   hermes --config ~/.hermes_test/config.toml query channels --chain nuahchain-1 --port transfer
   ```

---

## Transfer Issues / Проблемы трансферов

### Problem: Transfer timeout / Проблема: Таймаут трансфера

**Symptoms / Симптомы:**
```bash
error: packet timeout
```

**Solutions / Решения:**

1. **Increase timeout height / Увеличьте высоту таймаута:**
   ```bash
   hermes --config ~/.hermes_test/config.toml tx ft-transfer \
     --dst-chain osmo-test-5 \
     --src-chain nuahchain-1 \
     --src-port transfer \
     --src-channel channel-0 \
     --amount 1000 \
     --denom unuah \
     --receiver osmo19rl4cm2hmr8afy4kldpxz3fka4jguq0a5m7df8 \
     --timeout-height-offset 2000
   ```

2. **Check packet relay / Проверьте релей пакетов:**
   ```bash
   hermes --config ~/.hermes_test/config.toml query packet pending --chain nuahchain-1 --port transfer --channel-id channel-0
   ```

### Problem: Invalid denomination / Проблема: Неверная деноминация

**Symptoms / Симптомы:**
```bash
error: invalid denomination
```

**Solutions / Решения:**

1. **Check available denominations / Проверьте доступные деноминации:**
   ```bash
   ./build/nuahd query bank balances $RELAYER_ADDR
   ```

2. **Use correct IBC denomination format / Используйте правильный формат IBC деноминации:**
   ```bash
   # For IBC tokens, use the full IBC path
   # Для IBC токенов используйте полный IBC путь
   --denom ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
   ```

---

## Performance Issues / Проблемы производительности

### Problem: Slow packet relay / Проблема: Медленный релей пакетов

**Solutions / Решения:**

1. **Optimize configuration / Оптимизируйте конфигурацию:**
   ```toml
   # In config.toml
   max_msg_num = 50
   batch_delay = '200ms'
   ```

2. **Enable packet batching / Включите пакетирование:**
   ```toml
   [mode.packets]
   enabled = true
   clear_interval = 100
   clear_on_start = true
   ```

### Problem: High resource usage / Проблема: Высокое потребление ресурсов

**Solutions / Решения:**

1. **Reduce log level / Снизьте уровень логирования:**
   ```toml
   [global]
   log_level = 'warn'
   ```

2. **Optimize event source / Оптимизируйте источник событий:**
   ```toml
   event_source = { mode = 'pull', interval = '1s' }
   ```

---

## Monitoring and Debugging / Мониторинг и отладка

### Enable Debug Logging / Включение отладочного логирования

```bash
# Set environment variable / Установите переменную окружения
export RUST_LOG=debug

# Or use verbose flag / Или используйте флаг verbose
hermes --config ~/.hermes_test/config.toml start --verbose
```

### Health Check Commands / Команды проверки состояния

```bash
# Comprehensive health check / Комплексная проверка состояния
hermes --config ~/.hermes_test/config.toml health-check

# Check specific chain / Проверка конкретной сети
hermes --config ~/.hermes_test/config.toml health-check --chain nuahchain-1

# Check with verbose output / Проверка с подробным выводом
hermes --config ~/.hermes_test/config.toml health-check --verbose
```

### Log Analysis / Анализ логов

```bash
# Filter error logs / Фильтрация логов ошибок
hermes --config ~/.hermes_test/config.toml start 2>&1 | grep ERROR

# Monitor specific events / Мониторинг конкретных событий
hermes --config ~/.hermes_test/config.toml listen --chain nuahchain-1 | grep -i "packet\|client\|connection"
```

### Common Log Messages / Распространенные сообщения логов

1. **INFO messages / Информационные сообщения:**
   - `client update successful` - Client updated / Клиент обновлен
   - `packet relay successful` - Packet relayed / Пакет передан

2. **WARN messages / Предупреждения:**
   - `client expiry warning` - Client near expiry / Клиент близок к истечению
   - `low balance warning` - Low relayer balance / Низкий баланс релейера

3. **ERROR messages / Сообщения об ошибках:**
   - `insufficient funds` - Need to fund relayer / Нужно пополнить релейер
   - `client expired` - Client needs update / Клиент нужно обновить

---

## Emergency Procedures / Экстренные процедуры

### Reset Relayer State / Сброс состояния релейера

```bash
# Stop relayer / Остановите релейер
pkill hermes

# Clear pending packets / Очистите ожидающие пакеты
hermes --config ~/.hermes_test/config.toml clear packets --chain nuahchain-1 --port transfer --channel-id channel-0

# Update all clients / Обновите всех клиентов
hermes --config ~/.hermes_test/config.toml update clients --host-chain nuahchain-1
hermes --config ~/.hermes_test/config.toml update clients --host-chain osmo-test-5
```

### Backup and Restore / Резервное копирование и восстановление

```bash
# Backup configuration / Резервное копирование конфигурации
cp -r ~/.hermes_test ~/.hermes_test_backup

# Restore configuration / Восстановление конфигурации
cp -r ~/.hermes_test_backup ~/.hermes_test
```

---

## Getting Help / Получение помощи

### Community Resources / Ресурсы сообщества

- **Hermes Documentation:** https://hermes.informal.systems/
- **IBC Protocol:** https://ibc.cosmos.network/
- **Cosmos SDK:** https://docs.cosmos.network/
- **GitHub Issues:** https://github.com/informalsystems/hermes/issues

### Reporting Issues / Сообщение о проблемах

When reporting issues, include:
При сообщении о проблемах включите:

1. Hermes version: `hermes version`
2. Configuration file (sanitized)
3. Full error message and logs
4. Steps to reproduce
5. System information (OS, architecture)

### Log Collection / Сбор логов

```bash
# Collect comprehensive logs / Сбор комплексных логов
hermes --config ~/.hermes_test/config.toml start --verbose > hermes.log 2>&1

# System information / Информация о системе
uname -a > system_info.txt
hermes version >> system_info.txt
```

This troubleshooting guide should help resolve most common issues with Hermes IBC relayer setup and operation.
Это руководство по устранению неполадок должно помочь решить большинство распространенных проблем с настройкой и работой IBC релейера Hermes.

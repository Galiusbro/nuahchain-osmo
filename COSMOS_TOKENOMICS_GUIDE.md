# Правильная настройка токеномики и IBC в Cosmos SDK

## 🎯 Проблемы текущей настройки

Ваша текущая настройка имеет несколько проблем:

1. **Централизация токенов** - все токены находятся у валидатора
2. **Отсутствие IBC** - токен не виден в других сетях Cosmos
3. **Нет распределения** между разными участниками экосистемы
4. **Отсутствие relayer** для межблокчейнового взаимодействия

## 🏗 Правильная структура токеномики

### 1. Создание дополнительных аккаунтов

```bash
# Создание аккаунтов для разных целей
./build/nuahd keys add foundation --keyring-backend test
./build/nuahd keys add community --keyring-backend test
./build/nuahd keys add treasury --keyring-backend test
./build/nuahd keys add ecosystem --keyring-backend test
./build/nuahd keys add team --keyring-backend test

# Получение адресов
VALIDATOR_ADDR=$(./build/nuahd keys show validator --keyring-backend test -a)
FOUNDATION_ADDR=$(./build/nuahd keys show foundation --keyring-backend test -a)
COMMUNITY_ADDR=$(./build/nuahd keys show community --keyring-backend test -a)
TREASURY_ADDR=$(./build/nuahd keys show treasury --keyring-backend test -a)
ECOSYSTEM_ADDR=$(./build/nuahd keys show ecosystem --keyring-backend test -a)
TEAM_ADDR=$(./build/nuahd keys show team --keyring-backend test -a)
```

### 2. Правильное распределение токенов в genesis

```bash
# Сброс текущего genesis (ОСТОРОЖНО!)
./build/nuahd unsafe-reset-all

# Повторная инициализация
./build/nuahd init nuah-mainnet --chain-id nuahchain-1

# Изменение деноминации
sed -i 's/"stake"/"unuah"/g' ~/.nuahd/config/genesis.json

# Добавление аккаунтов с правильным распределением
./build/nuahd add-genesis-account $VALIDATOR_ADDR 5000000000000unuah --keyring-backend test    # 5M NUAH - валидатор
./build/nuahd add-genesis-account $FOUNDATION_ADDR 20000000000000unuah --keyring-backend test  # 20M NUAH - фонд
./build/nuahd add-genesis-account $COMMUNITY_ADDR 25000000000000unuah --keyring-backend test   # 25M NUAH - сообщество
./build/nuahd add-genesis-account $TREASURY_ADDR 30000000000000unuah --keyring-backend test    # 30M NUAH - казначейство
./build/nuahd add-genesis-account $ECOSYSTEM_ADDR 15000000000000unuah --keyring-backend test   # 15M NUAH - экосистема
./build/nuahd add-genesis-account $TEAM_ADDR 5000000000000unuah --keyring-backend test         # 5M NUAH - команда

# Создание gentx только с частью токенов валидатора
./build/nuahd gentx validator 2000000000000unuah --chain-id nuahchain-1 --keyring-backend test

# Сбор gentx
./build/nuahd collect-gentxs
```

### 3. Типичное распределение токенов в Cosmos

```
Общее предложение: 100,000,000 NUAH (100M)

📊 Распределение:
├── 30% (30M) - Казначейство/Treasury
├── 25% (25M) - Сообщество/Community Pool
├── 20% (20M) - Фонд/Foundation
├── 15% (15M) - Экосистема/Ecosystem Development
├── 5% (5M)   - Команда/Team
└── 5% (5M)   - Валидаторы/Initial Validators
```

## 🌐 Настройка IBC для межблокчейнового взаимодействия

### 1. Проверка IBC модулей в app.go

Убедитесь, что в вашем `app/app.go` включены IBC модули:

```go
import (
    "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
    transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
    ibc "github.com/cosmos/ibc-go/v7/modules/core"
    ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
    ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
    ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
    ibchost "github.com/cosmos/ibc-go/v7/modules/core/24-host"
    ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
)

// В NewApp функции
app.IBCKeeper = ibckeeper.NewKeeper(
    appCodec, keys[ibchost.StoreKey], app.GetSubspace(ibchost.ModuleName),
    app.StakingKeeper, app.UpgradeKeeper, scopedIBCKeeper,
)

app.TransferKeeper = ibctransferkeeper.NewKeeper(
    appCodec, keys[ibctransfertypes.StoreKey], app.GetSubspace(ibctransfertypes.ModuleName),
    app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
    app.AccountKeeper, app.BankKeeper, scopedTransferKeeper,
)
```

### 2. Настройка конфигурации для IBC

В `~/.nuahd/config/app.toml`:

```toml
[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"

[grpc]
enable = true
address = "0.0.0.0:9090"

[grpc-web]
enable = true
address = "0.0.0.0:9091"
```

### 3. Регистрация IBC каналов

```bash
# После запуска ноды, создание IBC канала (пример с Osmosis testnet)
# Это делается через relayer, но сначала нужно настроить relayer
```

## 🔗 Настройка Relayer для подключения к другим сетям

### 1. Установка Hermes Relayer

```bash
# Установка Hermes
curl -L https://github.com/informalsystems/hermes/releases/download/v1.7.4/hermes-v1.7.4-x86_64-unknown-linux-gnu.tar.gz | tar -xz
sudo mv hermes /usr/local/bin/

# Проверка установки
hermes version
```

### 2. Конфигурация Hermes

Создайте файл `~/.hermes/config.toml`:

```toml
[global]
log_level = 'info'

[mode]

[mode.clients]
enabled = true
refresh = true
misbehaviour = true

[mode.connections]
enabled = false

[mode.channels]
enabled = false

[mode.packets]
enabled = true

[[chains]]
id = 'nuahchain-1'
rpc_addr = 'http://144.76.169.123:26657'
grpc_addr = 'http://144.76.169.123:9090'
websocket_addr = 'ws://144.76.169.123:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'nuah'
key_name = 'relayer'
address_type = { derivation = 'cosmos' }
store_prefix = 'ibc'
default_gas = 100000
max_gas = 400000
gas_price = { price = 0.001, denom = 'unuah' }
gas_multiplier = 1.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
max_block_time = '30s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }

[[chains]]
id = 'osmo-test-5'  # Osmosis testnet
rpc_addr = 'https://rpc.osmotest5.osmosis.zone:443'
grpc_addr = 'https://grpc.osmotest5.osmosis.zone:443'
websocket_addr = 'wss://rpc.osmotest5.osmosis.zone:443/websocket'
rpc_timeout = '10s'
account_prefix = 'osmo'
key_name = 'relayer-osmo'
address_type = { derivation = 'cosmos' }
store_prefix = 'ibc'
default_gas = 100000
max_gas = 400000
gas_price = { price = 0.0025, denom = 'uosmo' }
gas_multiplier = 1.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
max_block_time = '30s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
```

### 3. Создание ключей для relayer

```bash
# Создание ключа для вашей сети
hermes keys add --chain nuahchain-1 --mnemonic-file ~/.nuahd/relayer_mnemonic.txt

# Создание ключа для Osmosis testnet
hermes keys add --chain osmo-test-5 --mnemonic-file ~/.hermes/osmo_mnemonic.txt
```

### 4. Создание IBC соединения

```bash
# Создание клиента
hermes create client --host-chain nuahchain-1 --reference-chain osmo-test-5

# Создание соединения
hermes create connection --a-chain nuahchain-1 --b-chain osmo-test-5

# Создание канала для transfer
hermes create channel --a-chain nuahchain-1 --a-connection connection-0 --a-port transfer --b-port transfer
```

## 🚀 Пошаговая инструкция по правильной настройке

### Шаг 1: Пересоздание genesis с правильным распределением

```bash
#!/bin/bash

# Остановка текущей ноды
pkill nuahd

# Сброс данных
./build/nuahd unsafe-reset-all

# Повторная инициализация
./build/nuahd init nuah-mainnet --chain-id nuahchain-1

# Создание всех необходимых ключей
./build/nuahd keys add validator --keyring-backend test --recover # Используйте существующую мнемонику
./build/nuahd keys add foundation --keyring-backend test
./build/nuahd keys add community --keyring-backend test
./build/nuahd keys add treasury --keyring-backend test
./build/nuahd keys add ecosystem --keyring-backend test
./build/nuahd keys add team --keyring-backend test

# Получение адресов
VALIDATOR_ADDR=$(./build/nuahd keys show validator --keyring-backend test -a)
FOUNDATION_ADDR=$(./build/nuahd keys show foundation --keyring-backend test -a)
COMMUNITY_ADDR=$(./build/nuahd keys show community --keyring-backend test -a)
TREASURY_ADDR=$(./build/nuahd keys show treasury --keyring-backend test -a)
ECOSYSTEM_ADDR=$(./build/nuahd keys show ecosystem --keyring-backend test -a)
TEAM_ADDR=$(./build/nuahd keys show team --keyring-backend test -a)

# Изменение деноминации
sed -i 's/"stake"/"unuah"/g' ~/.nuahd/config/genesis.json

# Добавление аккаунтов с правильным распределением
./build/nuahd add-genesis-account $VALIDATOR_ADDR 5000000000000unuah --keyring-backend test
./build/nuahd add-genesis-account $FOUNDATION_ADDR 20000000000000unuah --keyring-backend test
./build/nuahd add-genesis-account $COMMUNITY_ADDR 25000000000000unuah --keyring-backend test
./build/nuahd add-genesis-account $TREASURY_ADDR 30000000000000unuah --keyring-backend test
./build/nuahd add-genesis-account $ECOSYSTEM_ADDR 15000000000000unuah --keyring-backend test
./build/nuahd add-genesis-account $TEAM_ADDR 5000000000000unuah --keyring-backend test

# Создание gentx
./build/nuahd gentx validator 2000000000000unuah --chain-id nuahchain-1 --keyring-backend test

# Сбор gentx
./build/nuahd collect-gentxs

# Валидация genesis
./build/nuahd validate-genesis
```

### Шаг 2: Обновление конфигурации для IBC

```bash
# Обновление app.toml для включения всех API
cat >> ~/.nuahd/config/app.toml << EOF

[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"
max-open-connections = 1000
rpc-read-timeout = 10
rpc-write-timeout = 0
rpc-max-body-bytes = 1000000
enabled-unsafe-cors = true

[grpc]
enable = true
address = "0.0.0.0:9090"

[grpc-web]
enable = true
address = "0.0.0.0:9091"
enable-unsafe-cors = true
EOF
```

### Шаг 3: Запуск обновленной ноды

```bash
# Запуск ноды с поддержкой IBC
nohup ./build/nuahd start \
  --rpc.laddr=tcp://0.0.0.0:26657 \
  --api.enable=true \
  --api.address=tcp://0.0.0.0:1317 \
  --grpc.enable=true \
  --grpc.address=0.0.0.0:9090 \
  > nuahd.log 2>&1 &
```

## 🔍 Проверка правильности настройки

### 1. Проверка распределения токенов

```bash
# Проверка балансов всех аккаунтов
echo "=== Token Distribution Check ==="
echo "Validator: $(./build/nuahd query bank balances $VALIDATOR_ADDR)"
echo "Foundation: $(./build/nuahd query bank balances $FOUNDATION_ADDR)"
echo "Community: $(./build/nuahd query bank balances $COMMUNITY_ADDR)"
echo "Treasury: $(./build/nuahd query bank balances $TREASURY_ADDR)"
echo "Ecosystem: $(./build/nuahd query bank balances $ECOSYSTEM_ADDR)"
echo "Team: $(./build/nuahd query bank balances $TEAM_ADDR)"

# Проверка общего предложения
./build/nuahd query bank total
```

### 2. Проверка IBC готовности

```bash
# Проверка IBC клиентов
curl -s http://144.76.169.123:1317/ibc/core/client/v1/client_states | jq

# Проверка IBC каналов
curl -s http://144.76.169.123:1317/ibc/core/channel/v1/channels | jq

# Проверка поддержки transfer модуля
curl -s http://144.76.169.123:1317/ibc/apps/transfer/v1/params | jq
```

## 🌟 Преимущества правильной настройки

### 1. Децентрализация
- Токены распределены между разными участниками
- Снижение риска централизации власти
- Лучшее управление экосистемой

### 2. IBC совместимость
- Токен виден в других сетях Cosmos
- Возможность торговли на DEX (например, Osmosis)
- Межблокчейновые переводы

### 3. Экосистемное развитие
- Фонды для развития сообщества
- Ресурсы для экосистемных проектов
- Устойчивое финансирование

## 📋 Чек-лист для продакшн запуска

- [ ] Правильное распределение токенов между участниками
- [ ] IBC модули включены и настроены
- [ ] Relayer настроен и работает
- [ ] API эндпоинты доступны (REST, gRPC)
- [ ] Безопасность: приватные ключи защищены
- [ ] Мониторинг: логи и метрики настроены
- [ ] Документация: инструкции для пользователей
- [ ] Тестирование: IBC переводы работают

## 🔗 Полезные ссылки

- [Cosmos SDK Documentation](https://docs.cosmos.network/)
- [IBC Protocol](https://ibc.cosmos.network/)
- [Hermes Relayer](https://hermes.informal.systems/)
- [Osmosis DEX](https://osmosis.zone/)

---

*После выполнения этих шагов ваш токен NUAH будет правильно распределен и готов для межблокчейнового взаимодействия в экосистеме Cosmos!*

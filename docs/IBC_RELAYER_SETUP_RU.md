# Руководство по настройке IBC релейера (Русский)

## Содержание
1. [Обзор](#обзор)
2. [Предварительные требования](#предварительные-требования)
3. [Установка](#установка)
4. [Конфигурация](#конфигурация)
5. [Настройка ключей](#настройка-ключей)
6. [Создание IBC соединений](#создание-ibc-соединений)
7. [Создание IBC каналов](#создание-ibc-каналов)
8. [Тестирование IBC трансферов](#тестирование-ibc-трансферов)
9. [Устранение неполадок](#устранение-неполадок)
10. [Лучшие практики](#лучшие-практики)

## Обзор

Данное руководство предоставляет пошаговые инструкции по настройке IBC (Inter-Blockchain Communication) релейера с использованием Hermes для подключения вашего блокчейна NUAH к внешним сетям, таким как Osmosis testnet.

IBC релейеры являются важными компонентами, которые обеспечивают межсетевое взаимодействие путем:
- Создания и поддержания IBC клиентов
- Установления соединений между блокчейнами
- Создания каналов для конкретных протоколов (например, трансферы токенов)
- Передачи пакетов между подключенными сетями

## Предварительные требования

Перед началом убедитесь, что у вас есть:

1. **Работающая нода NUAH** с включенным IBC
2. **Установленный Hermes релейер** (версия 1.7.4 или новее)
3. **Доступ к целевой сети** (например, Osmosis testnet)
4. **Достаточные средства** в обеих сетях для оплаты комиссий
5. **Базовое понимание** концепций IBC

### Системные требования
- Операционная система: macOS, Linux или Windows (WSL)
- ОЗУ: Минимум 4ГБ, Рекомендуется 8ГБ+
- Хранилище: Минимум 10ГБ свободного места
- Сеть: Стабильное интернет-соединение

## Установка

### Установка Hermes

#### Вариант 1: Использование Cargo (Рекомендуется)
```bash
# Установите Rust, если еще не установлен
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source ~/.cargo/env

# Установите Hermes
cargo install ibc-relayer-cli --bin hermes --locked
```

#### Вариант 2: Скачивание готового бинарника
```bash
# Скачайте последний релиз
wget https://github.com/informalsystems/hermes/releases/download/v1.7.4/hermes-v1.7.4-x86_64-unknown-linux-gnu.tar.gz

# Распакуйте и установите
tar -xzf hermes-v1.7.4-x86_64-unknown-linux-gnu.tar.gz
sudo mv hermes /usr/local/bin/
```

### Проверка установки
```bash
hermes --version
# Должно вывести: hermes 1.7.4+ab73266
```

## Конфигурация

### 1. Создание директории конфигурации
```bash
mkdir -p ~/.hermes_test
```

### 2. Создание файла конфигурации
Создайте файл `~/.hermes_test/config.toml` со следующим содержимым:

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
enabled = false

# Конфигурация сети NUAH
[[chains]]
id = 'nuahchain-1'
type = 'CosmosSdk'
rpc_addr = 'http://localhost:26657'
grpc_addr = 'http://localhost:9090'
event_source = { mode = 'push', url = 'ws://localhost:26657/websocket', batch_delay = '500ms' }
rpc_timeout = '10s'
account_prefix = 'nuah'
key_name = 'test-relayer'
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

# Конфигурация Osmosis Testnet
[[chains]]
id = 'osmo-test-5'
type = 'CosmosSdk'
rpc_addr = 'https://rpc.testnet.osmosis.zone:443'
grpc_addr = 'https://grpc.testnet.osmosis.zone:443'
event_source = { mode = 'push', url = 'wss://rpc.testnet.osmosis.zone:443/websocket', batch_delay = '500ms' }
rpc_timeout = '10s'
account_prefix = 'osmo'
key_name = 'osmosis-relayer'
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

### 3. Проверка конфигурации
```bash
hermes --config ~/.hermes_test/config.toml health-check
```

## Настройка ключей

### 1. Создание ключей релейера

Для сети NUAH:
```bash
# Сгенерируйте или импортируйте мнемонику для NUAH релейера
hermes --config ~/.hermes_test/config.toml keys add \
  --chain nuahchain-1 \
  --mnemonic-file <(echo "ваша мнемоническая фраза здесь")
```

Для Osmosis testnet:
```bash
# Сгенерируйте или импортируйте мнемонику для Osmosis релейера
hermes --config ~/.hermes_test/config.toml keys add \
  --chain osmo-test-5 \
  --mnemonic-file <(echo "ваша мнемоническая фраза здесь")
```

### 2. Проверка ключей
```bash
# Список ключей для сети NUAH
hermes --config ~/.hermes_test/config.toml keys list --chain nuahchain-1

# Список ключей для Osmosis testnet
hermes --config ~/.hermes_test/config.toml keys list --chain osmo-test-5
```

### 3. Пополнение счетов релейера

**Сеть NUAH:**
```bash
# Проверка баланса
./build/nuahd query bank balances [адрес-релейера]

# Отправка средств при необходимости
./build/nuahd tx bank send [отправитель] [адрес-релейера] 1000000unuah --fees 1000unuah
```

**Osmosis Testnet:**
- Используйте кран Osmosis testnet: https://faucet.testnet.osmosis.zone/
- Запросите токены для адреса вашего релейера

## Создание IBC соединений

### 1. Создание IBC клиентов

Создание клиента для сети NUAH на Osmosis:
```bash
hermes --config ~/.hermes_test/config.toml create client \
  --host-chain osmo-test-5 \
  --reference-chain nuahchain-1
```

Создание клиента для Osmosis на сети NUAH:
```bash
hermes --config ~/.hermes_test/config.toml create client \
  --host-chain nuahchain-1 \
  --reference-chain osmo-test-5
```

### 2. Создание соединения
```bash
hermes --config ~/.hermes_test/config.toml create connection \
  --a-chain nuahchain-1 \
  --b-chain osmo-test-5
```

### 3. Проверка соединения
```bash
# Проверка соединений в сети NUAH
./build/nuahd query ibc connection connections

# Проверка соединений в Osmosis (если есть доступ)
osmosisd query ibc connection connections
```

## Создание IBC каналов

### 1. Создание канала для трансферов
```bash
hermes --config ~/.hermes_test/config.toml create channel \
  --a-chain nuahchain-1 \
  --a-connection connection-0 \
  --a-port transfer \
  --b-port transfer \
  --channel-version ics20-1 \
  --order unordered
```

### 2. Проверка канала
```bash
# Проверка каналов в сети NUAH
./build/nuahd query ibc channel channels

# Список всех каналов через Hermes
hermes --config ~/.hermes_test/config.toml query channels --chain nuahchain-1
```

## Тестирование IBC трансферов

### 1. Трансфер из NUAH в Osmosis
```bash
./build/nuahd tx ibc-transfer transfer \
  transfer \
  channel-0 \
  [адрес-получателя-osmosis] \
  1000unuah \
  --from [отправитель-nuah] \
  --fees 1000unuah \
  --timeout-height 0-0 \
  --timeout-timestamp 0
```

### 2. Трансфер из Osmosis в NUAH
```bash
osmosisd tx ibc-transfer transfer \
  transfer \
  channel-X \
  [адрес-получателя-nuah] \
  1000uosmo \
  --from [отправитель-osmosis] \
  --fees 1000uosmo \
  --chain-id osmo-test-5 \
  --node https://rpc.testnet.osmosis.zone:443
```

### 3. Запуск релейера (для передачи пакетов)
```bash
hermes --config ~/.hermes_test/config.toml start
```

## Устранение неполадок

### Распространенные проблемы

#### 1. Ошибка "Insufficient funds" (Недостаточно средств)
**Проблема:** У аккаунта релейера недостаточно токенов для оплаты комиссий.
**Решение:**
- Пополните аккаунт релейера через кран или прямой перевод
- Проверьте минимальный требуемый баланс для сети

#### 2. Ошибка "Client state type not supported" (Тип состояния клиента не поддерживается)
**Проблема:** Несовместимые типы клиентов между сетями.
**Решение:**
- Убедитесь, что обе сети поддерживают одинаковую версию IBC
- Проверьте совместимость Hermes с версиями сетей

#### 3. Ошибка "Connection not found" (Соединение не найдено)
**Проблема:** Попытка создать канал с несуществующим соединением.
**Решение:**
- Проверьте существование соединения: `hermes query connections --chain [chain-id]`
- Сначала создайте соединение, если оно не существует

#### 4. Предупреждения Health Check
**Проблема:** Предупреждения о совместимости версий SDK.
**Решение:**
- Обычно это некритичные предупреждения
- Убедитесь, что используете совместимую версию Hermes
- Обновите программное обеспечение сети при необходимости

### Команды для отладки

```bash
# Проверка логов Hermes
hermes --config ~/.hermes_test/config.toml health-check

# Запрос конкретного клиента
hermes --config ~/.hermes_test/config.toml query client state --chain [chain-id] --client [client-id]

# Запрос деталей соединения
hermes --config ~/.hermes_test/config.toml query connection end --chain [chain-id] --connection [connection-id]

# Запрос деталей канала
hermes --config ~/.hermes_test/config.toml query channel end --chain [chain-id] --port [port-id] --channel [channel-id]
```

## Лучшие практики

### Безопасность
1. **Используйте выделенные аккаунты релейера** - Не используйте валидаторские или основные аккаунты
2. **Безопасное хранение ключей** - Храните мнемоники безопасно, рассмотрите аппаратные кошельки для продакшена
3. **Мониторинг балансов** - Настройте уведомления о низких балансах
4. **Регулярные обновления** - Поддерживайте Hermes и программное обеспечение сетей в актуальном состоянии

### Производительность
1. **Оптимизация настроек газа** - Настройте цены на газ в зависимости от условий сети
2. **Мониторинг задержек пакетов** - Установите подходящие значения таймаутов
3. **Использование нескольких релейеров** - Для высоконагруженных каналов рассмотрите несколько экземпляров релейера
4. **Распределение ресурсов** - Обеспечьте достаточные CPU и память для операций релейера

### Мониторинг
1. **Настройка логирования** - Настройте подходящие уровни логов для мониторинга
2. **Проверки состояния** - Регулярный мониторинг проверок состояния
3. **Сбор метрик** - Используйте Prometheus/Grafana для расширенного мониторинга
4. **Системы уведомлений** - Настройте уведомления о сбоях релейера или застрявших пакетах

### Обслуживание
1. **Регулярные обновления клиентов** - Обновляйте IBC клиенты до их истечения
2. **Мониторинг соединений** - Отслеживайте состояние и статус соединений
3. **Обслуживание каналов** - Мониторинг состояния каналов и потока пакетов
4. **Процедуры резервного копирования** - Регулярное резервное копирование конфигурации и ключей релейера

---

Для получения более подробной информации обратитесь к:
- [Документация Hermes](https://hermes.informal.systems/)
- [Спецификация протокола IBC](https://github.com/cosmos/ibc)
- [IBC модуль Cosmos SDK](https://ibc.cosmos.network/)

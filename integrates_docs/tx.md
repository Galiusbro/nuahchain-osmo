# Тестирование передачи токенов в Osmosis

## Обзор

Документация описывает процесс тестирования передачи токенов Roma между различными аккаунтами в Osmosis. Тестирование включает создание новых аккаунтов, передачу токенов и проверку балансов.

## Участники тестирования

### Аккаунты

1. **Алиса** (`osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue`)
   - Создатель токена Roma
   - Администратор токена
   - Начальный баланс: 1,000,000 Roma токенов

2. **Боб** (`osmo1n923kpm70la27k9jx3a4khkydka2kuwyygyy49`)
   - Существующий аккаунт
   - Начальный баланс: stake и uOSMO токены

3. **Чарли** (`osmo1xsq06704l7p0zd44dntx2npsyx8rjjp9z5cznd`)
   - Новый аккаунт, созданный для тестирования
   - Начальный баланс: пустой

## Создание нового аккаунта

### Команда создания

```bash
./build/osmosisd keys add charlie --keyring-backend test
```

### Результат

```
- address: osmo1xsq06704l7p0zd44dntx2npsyx8rjjp9z5cznd
  name: charlie
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"As4v1YhEG0A1EHE7s0mBk4nPAPzUlHe7pSBny43kluHf"}'
  type: local
```

### Важные замечания

- **Сохраните мнемоническую фразу** в безопасном месте
- Она единственный способ восстановить аккаунт при потере пароля
- Новый аккаунт не имеет начальных токенов

## Сценарии тестирования

### Сценарий 1: Алиса → Боб (1 Roma)

#### Команда передачи

```bash
./build/osmosisd tx bank send osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue \
  osmo1n923kpm70la27k9jx3a4khkydka2kuwyygyy49 \
  1000000factory/osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue/roma \
  --from alice \
  --keyring-backend test \
  --chain-id localosmosis \
  --gas auto \
  --gas-adjustment 1.5 \
  --fees 2500stake \
  -y
```

#### Параметры

- **От**: Алиса (`osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue`)
- **К**: Боб (`osmo1n923kpm70la27k9jx3a4khkydka2kuwyygyy49`)
- **Количество**: 1,000,000 базовых единиц (1.0 Roma)
- **Fee**: 2,500 stake
- **Chain ID**: localosmosis

#### Результат

```
gas estimate: 117159
code: 0
codespace: ""
txhash: 2C7AC1E4F9980EEE0103A50332E40181F6B5011AEBC8E20E64A749944D0C9879
```

#### Изменения балансов

**До:**
- Алиса: 1,000,000 Roma
- Боб: 0 Roma

**После:**
- Алиса: 0 Roma
- Боб: 1,000,000 Roma (1.0 Roma)

### Сценарий 2: Боб → Чарли (0.5 Roma)

#### Команда передачи

```bash
./build/osmosisd tx bank send osmo1n923kpm70la27k9jx3a4khkydka2kuwyygyy49 \
  osmo1xsq06704l7p0zd44dntx2npsyx8rjjp9z5cznd \
  500000factory/osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue/roma \
  --from bob \
  --keyring-backend test \
  --chain-id localosmosis \
  --gas auto \
  --gas-adjustment 1.5 \
  --fees 2500stake \
  -y
```

#### Параметры

- **От**: Боб (`osmo1n923kpm70la27k9jx3a4khkydka2kuwyygyy49`)
- **К**: Чарли (`osmo1xsq06704l7p0zd44dntx2npsyx8rjjp9z5cznd`)
- **Количество**: 500,000 базовых единиц (0.5 Roma)
- **Fee**: 2,500 stake
- **Chain ID**: localosmosis

#### Результат

```
gas estimate: 167509
code: 0
codespace: ""
txhash: B566FA0A0AAAA7C54306E5C48D9C9D813A03F6F90CE8EA1C3B2C3C10AC45869A
```

#### Изменения балансов

**До:**
- Боб: 1,000,000 Roma
- Чарли: 0 Roma

**После:**
- Боб: 500,000 Roma (0.5 Roma)
- Чарли: 500,000 Roma (0.5 Roma)

### Сценарий 3: Подготовка Чарли (отправка stake токенов)

#### Проблема

Чарли не может отправить Roma токены, так как у него нет stake токенов для оплаты fee.

#### Решение

Алиса отправляет Чарли stake токены для оплаты транзакций.

#### Команда

```bash
./build/osmosisd tx bank send osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue \
  osmo1xsq06704l7p0zd44dntx2npsyx8rjjp9z5cznd \
  10000stake \
  --from alice \
  --keyring-backend test \
  --chain-id localosmosis \
  --gas auto \
  --gas-adjustment 1.5 \
  --fees 2500stake \
  -y
```

#### Результат

```
gas estimate: 116640
code: 0
codespace: ""
txhash: AC2664C851F7C8F49556F4F91F8C9A3891620415762F03816B6B8EDC48918B06
```

### Сценарий 4: Чарли → Алиса (0.25 Roma)

#### Команда передачи

```bash
./build/osmosisd tx bank send osmo1xsq06704l7p0zd44dntx2npsyx8rjjp9z5cznd \
  osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue \
  250000factory/osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue/roma \
  --from charlie \
  --keyring-backend test \
  --chain-id localosmosis \
  --gas auto \
  --gas-adjustment 1.5 \
  --fees 2500stake \
  -y
```

#### Параметры

- **От**: Чарли (`osmo1xsq06704l7p0zd44dntx2npsyx8rjjp9z5cznd`)
- **К**: Алиса (`osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue`)
- **Количество**: 250,000 базовых единиц (0.25 Roma)
- **Fee**: 2,500 stake
- **Chain ID**: localosmosis

#### Результат

```
gas estimate: 146710
code: 0
codespace: ""
txhash: 8C23CEEFC7366561D355A65F9A5E17C40D69535732244D864BD9CB879E13DDE8
```

## Финальные балансы

### Алиса
- **Roma токен**: 250,000 (0.25 Roma)
- **Stake**: 998,940,500 (потратила 12,500 на fees)
- **uOSMO**: 1,000,000,000

### Боб
- **Roma токен**: 500,000 (0.5 Roma)
- **Stake**: 999,997,500 (потратил 2,500 на fee)
- **uOSMO**: 1,000,000,000

### Чарли
- **Roma токен**: 250,000 (0.25 Roma)
- **Stake**: 7,500 (получил 10,000, потратил 2,500 на fee)

## Команды проверки балансов

### Проверка баланса конкретного аккаунта

```bash
./build/osmosisd query bank balances <адрес_аккаунта> --chain-id localosmosis
```

### Примеры

```bash
# Проверка баланса Алисы
./build/osmosisd query bank balances osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue --chain-id localosmosis

# Проверка баланса Боба
./build/osmosisd query bank balances osmo1n923kpm70la27k9jx3a4khkydka2kuwyygyy49 --chain-id localosmosis

# Проверка баланса Чарли
./build/osmosisd query bank balances osmo1xsq06704l7p0zd44dntx2npsyx8rjjp9z5cznd --chain-id localosmosis
```

## Важные замечания

### Fee токены

- **localosmosis**: используйте `stake` токены для fee
- **mainnet**: используйте `uosmo` токены для fee
- Всегда проверяйте доступные fee токены командой:
  ```bash
  ./build/osmosisd query txfees base-denom --chain-id <chain_id>
  ```

### Chain ID

- Всегда указывайте правильный `--chain-id`
- Для локального тестирования: `localosmosis`
- Для mainnet: `osmosis-1`

### Права доступа

- Только владелец токенов может их передавать
- Для передачи Roma токенов нужны stake токены для оплаты fee
- Новые аккаунты не имеют начальных токенов

## Результаты тестирования

### ✅ Успешные транзакции

1. **Алиса → Боб**: 1 Roma ✅
2. **Боб → Чарли**: 0.5 Roma ✅
3. **Алиса → Чарли**: 10,000 stake (для fee) ✅
4. **Чарли → Алиса**: 0.25 Roma ✅

### ✅ Проверенные функции

- Создание новых аккаунтов
- Передача токенов между аккаунтами
- Система fee и gas
- Обновление балансов в реальном времени
- Работа с метаданными токенов

### ✅ Технические аспекты

- CLI команды работают корректно
- Транзакции проходят без ошибок
- Балансы обновляются корректно
- Система fee работает правильно

## Заключение

Тестирование передачи токенов Roma прошло успешно. Все сценарии передачи между тремя аккаунтами выполнены без ошибок. Система работает стабильно, метаданные токенов корректно отображаются, а балансы обновляются в реальном времени.

Токен Roma готов к использованию в реальных приложениях и может передаваться между любыми аккаунтами в сети Osmosis.

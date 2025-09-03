# Создание Stableswap пула в Osmosis

## Обзор

Документация описывает успешный процесс создания Stableswap пула с токенами ROMA и Stake в Osmosis blockchain.

## Предварительные требования

### 1. Настройка окружения
```bash
export CHAIN_ID=localnuah
```

### 2. Проверка доступных ключей
```bash
./build/nuahd keys list --keyring-backend=test
```

### 3. Убедиться, что у Алисы достаточно токенов
```bash
./build/nuahd q bank balances osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue --chain-id $CHAIN_ID
```

## Создание пула

### 1. Подготовка JSON файла конфигурации

Создайте файл `pool_stableswap_roma_stake.json`:

```json
{
  "initial-deposit": "1000000factory/osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue/roma,1000000stake",
  "swap-fee": "0.01",
  "exit-fee": "0.01",
  "future-governor": "168h",
  "scaling-factors": "1,1"
}
```

**Важные моменты:**
- Токены в `initial-deposit` должны быть отсортированы по алфавиту
- `scaling-factors` обязателен для Stableswap пулов
- `future-governor` определяет время блокировки управления пулом

### 2. Команда создания пула

```bash
./build/nuahd tx gamm create-pool \
  --pool-type stableswap \
  --pool-file pool_stableswap_roma_stake.json \
  --from alice \
  --chain-id $CHAIN_ID \
  --keyring-backend=test \
  --fees 5000stake \
  --gas 500000 \
  -y
```

**Ключевые параметры:**
- `--pool-type stableswap` - тип пула
- `--fees 5000stake` - комиссия за транзакцию
- `--gas 500000` - лимит газа (важно для успешного выполнения)

### 3. Проверка создания пула

```bash
# Подождать несколько секунд для обработки
sleep 5

# Проверить список пулов
./build/nuahd q gamm pools --chain-id $CHAIN_ID
```

## Результат

### Созданный пул
- **ID пула**: 1
- **Тип**: Stableswap
- **Адрес**: osmo19e2mf7cywkv7zaug6nk5f87d07fxrdgrladvymh2gwv5crvm3vnsuewhh7
- **Ликвидность**: 1,000,000 ROMA + 1,000,000 Stake
- **Swap fee**: 1%
- **Exit fee**: 0%
- **LP токены**: 100,000,000,000,000,000,000 gamm/pool/1

### Баланс Алисы после создания
- **ROMA**: 99,000,000 (было 100,000,000)
- **Stake**: 997,890,500 (было 998,895,500)
- **LP токены**: 100,000,000,000,000,000,000 gamm/pool/1

## Типичные ошибки и их решения

### 1. "insufficient fees"
**Ошибка**: `insufficient fees; got: 2500stake which converts to 2500stake. required: 5000stake`
**Решение**: Увеличить fee пропорционально gas limit

### 2. "out of gas"
**Ошибка**: `out of gas in location: Has; gasWanted: 200000, gasUsed: 200940: out of gas`
**Решение**: Увеличить `--gas` до 500000 или больше

### 3. "unsorted initial pool liquidity"
**Ошибка**: Токены в `initial-deposit` не отсортированы по алфавиту
**Решение**: Отсортировать токены: `stake` должен идти перед `factory/...`

### 4. "decimal string cannot be empty"
**Ошибка**: Проблема парсинга JSON файла
**Решение**: Проверить синтаксис JSON и убедиться, что все поля заполнены

## Команды для управления пулом

### Проверка баланса
```bash
./build/nuahd q bank balances osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue --chain-id $CHAIN_ID
```

### Информация о пуле
```bash
./build/nuahd q gamm pool 1 --chain-id $CHAIN_ID
```

### Выполнение swap
```bash
./build/nuahd tx gamm swap-exact-amount-in \
  --pool-id 1 \
  --token-in 1000stake \
  --token-out-min-amount 900 \
  --from alice \
  --chain-id $CHAIN_ID \
  --keyring-backend=test \
  --fees 2500stake \
  -y
```

## Заключение

Успешное создание Stableswap пула требует:
1. Правильной конфигурации JSON файла
2. Достаточного количества токенов
3. Корректных параметров gas и fees
4. Соблюдения порядка токенов в initial-deposit

Stableswap пулы идеально подходят для токенов с коррелированными ценами и обеспечивают низкий slippage при торговле.

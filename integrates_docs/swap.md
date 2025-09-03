# Выполнение Swap в Stableswap пуле

## Обзор

Документация описывает успешное выполнение swap операции в Stableswap пуле ID 1 (ROMA + Stake) в Osmosis blockchain.

## Предварительные требования

### 1. Настройка окружения
```bash
export CHAIN_ID=localnuah
```

### 2. Проверка состояния пула
```bash
./build/nuahd q gamm pool 1 --chain-id $CHAIN_ID
```

### 3. Проверка доступных ключей
```bash
./build/nuahd keys list --keyring-backend=test
```

## Выполнение Swap

### 1. Проверка текущей цены

Перед выполнением swap рекомендуется проверить текущую spot price:

```bash
./build/nuahd q gamm spot-price 1 stake factory/osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue/roma true --chain-id $CHAIN_ID
```

**Результат:**
```
spot_price: "1.000000000000000000"
```

Это означает, что **1 stake = 1.000000000000000000 ROMA** (без учёта swap fee).

### 2. Команда Swap

#### Swap Exact Amount In
```bash
./build/nuahd tx gamm swap-exact-amount-in \
  <token-in> \
  <token-out-min-amount> \
  --swap-route-pool-ids=<pool-id> \
  --swap-route-denoms=<token-out-denom> \
  --from <key-name> \
  --chain-id $CHAIN_ID \
  --keyring-backend=test \
  --fees <fee-amount> \
  -y
```

#### Пример успешного swap
```bash
./build/nuahd tx gamm swap-exact-amount-in 1000stake 900 \
  --swap-route-pool-ids=1 \
  --swap-route-denoms=factory/osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue/roma \
  --from alice \
  --chain-id $CHAIN_ID \
  --keyring-backend=test \
  --fees 2500stake \
  -y
```

**Ключевые параметры:**
- `1000stake` - количество токенов для обмена
- `900` - минимальное количество выходных токенов (защита от slippage)
- `--swap-route-pool-ids=1` - ID пула для swap
- `--swap-route-denoms=factory/osmo137m5vcnusv4mmps0yna9hg5kypwayj6e2tjjue/roma` - **только выходной токен**

## Результаты Swap

### Успешная транзакция
```
code: 0
codespace: ""
txhash: 7ED429C68516B875462E67EDE1929E143945CA660248F4A6CBAF1F07F24D705C
```

### Изменения баланса

**До swap:**
- ROMA: 99,000,000
- Stake: 997,888,000

**После swap:**
- ROMA: 99,000,989 (+989)
- Stake: 997,884,500 (-3,500)

**Детали:**
- **Обменено:** 1,000 stake
- **Получено:** 989 ROMA
- **Комиссия за транзакцию:** 2,500 stake
- **Итого потрачено:** 3,500 stake

### Изменения баланса пула

**До swap:**
- ROMA: 1,000,000
- Stake: 1,000,000

**После swap:**
- ROMA: 999,011 (-989)
- Stake: 1,001,000 (+1,000)

## Типичные ошибки и их решения

### 1. "swap route pool ids and denoms mismatch"
**Ошибка**: Неправильный порядок или количество denoms
**Решение**: Указывать в `--swap-route-denoms` **только выходной токен**

### 2. "token is lesser than min amount"
**Ошибка**: `tokenOutMinAmount` слишком высокий
**Решение**: Уменьшить `tokenOutMinAmount` с учётом slippage и swap fee

### 3. "insufficient fees"
**Ошибка**: Недостаточно fee для транзакции
**Решение**: Увеличить `--fees` пропорционально gas limit

## Важные моменты

### 1. Синтаксис команды
- `--swap-route-denoms` принимает **только выходной токен**, не оба
- `--swap-route-pool-ids` указывает ID пула для swap

### 2. Защита от slippage
- Всегда устанавливайте реалистичный `tokenOutMinAmount`
- Учитывайте swap fee пула (в нашем случае 1%)
- Учитывайте возможное проскальзывание цены

### 3. Комиссии
- `--fees` - комиссия за транзакцию
- Swap fee пула - комиссия самого пула (1% в нашем случае)

## Команды для анализа

### Проверка баланса пользователя
```bash
./build/nuahd q bank balances <address> --chain-id $CHAIN_ID
```

### Проверка состояния пула
```bash
./build/nuahd q gamm pool <pool-id> --chain-id $CHAIN_ID
```

### Проверка spot price
```bash
./build/nuahd q gamm spot-price <pool-id> <base-denom> <quote-denom> <with-swap-fee> --chain-id $CHAIN_ID
```

### Детали транзакции
```bash
./build/nuahd q tx <txhash> --chain-id $CHAIN_ID
```

## Заключение

Успешное выполнение swap требует:
1. Правильного синтаксиса команды
2. Корректного указания `--swap-route-denoms` (только выходной токен)
3. Реалистичного `tokenOutMinAmount` для защиты от slippage
4. Достаточного количества fee для транзакции

Stableswap пулы обеспечивают эффективный обмен токенов с предсказуемыми комиссиями и минимальным slippage.
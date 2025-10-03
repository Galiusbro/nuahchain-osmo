# Leverage Module

Модуль `leverage` предоставляет функциональность маржинальной торговли с плечом до 100x для токенов, созданных через модуль `usertoken`. Модуль поддерживает как LONG, так и SHORT позиции с реальным заимствованием токенов.

## 🚀 Основные возможности

- **Маржинальная торговля** с плечом от 1.1x до 100x
- **LONG позиции** - покупка токенов с плечом
- **SHORT позиции** - продажа токенов с плечом через реальное заимствование
- **Система кредитования** с автоматическим созданием пулов ликвидности
- **Ликвидация позиций** при достижении критического уровня
- **Управление коллатералом** - добавление/удаление залога
- **Slippage protection** - защита от проскальзывания цены

## 📋 Архитектура модуля

### Основные компоненты

1. **LeverageKeeper** - основной keeper модуля
2. **LendingKeeper** - внутренний keeper для управления кредитованием
3. **Position Management** - управление позициями
4. **Risk Management** - система управления рисками

### Взаимодействие с другими модулями

- **usertoken** - получение цен токенов, покупка/продажа через bonding curve
- **bank** - управление балансами и переводами
- **account** - управление аккаунтами

## 🔧 Установка и настройка

### 1. Создание аккаунтов

Для полноценного тестирования модуля создайте следующие аккаунты:

```bash
# Создание ключей
nuahd keys add trader --keyring-backend test
nuahd keys add provider --keyring-backend test
nuahd keys add liquidator --keyring-backend test

# Получение адресов
TRADER_ADDR=$(nuahd keys show trader -a --keyring-backend test)
PROVIDER_ADDR=$(nuahd keys show provider -a --keyring-backend test)
LIQUIDATOR_ADDR=$(nuahd keys show liquidator -a --keyring-backend test)
```

### 2. Финансирование аккаунтов

```bash
# Финансирование трейдера (для открытия позиций)
nuahd tx bank send alice $TRADER_ADDR 1000000000unuah --from alice --chain-id nuahchain --fees 1000unuah --keyring-backend test -y

# Финансирование провайдера ликвидности
nuahd tx bank send alice $PROVIDER_ADDR 1000000000unuah --from alice --chain-id nuahchain --fees 1000unuah --keyring-backend test -y

# Финансирование ликвидатора
nuahd tx bank send alice $LIQUIDATOR_ADDR 1000000000unuah --from alice --chain-id nuahchain --fees 1000unuah --keyring-backend test -y
```

### 3. Создание пользовательского токена

```bash
# Создание токена для торговли
nuahd tx usertoken create-user-token testcoin "Test Coin" "TEST" 6 --from alice --chain-id nuahchain --fees 1000unuah --gas 2000000 --keyring-backend test -y

# Получение denom созданного токена
TOKEN_DENOM="factory/nuah1f7u8xmej7mwa5xvuew7xk3ls3m4f5zemmzyx2x/testcoin"
```

### 4. Покупка токенов для установления цены

```bash
# Покупка токенов через bonding curve для установления цены
nuahd tx usertoken buy-tokens $TOKEN_DENOM 1000000unuah 1000000 --from alice --chain-id nuahchain --fees 1000unuah --keyring-backend test -y
```

## 💼 Использование модуля

### Параметры модуля

```bash
# Просмотр текущих параметров
nuahd query leverage params --chain-id nuahchain
```

**Дефолтные параметры:**
- **MaxLeverage**: 100x (максимальное плечо)
- **MaintenanceMargin**: 1% (минимальная маржа)
- **LiquidationFee**: 0.5% (комиссия за ликвидацию)
- **TradingFee**: 0.1% (торговая комиссия)
- **MaxPositionSize**: 1B токенов (максимальный размер позиции)
- **MinCollateralAmount**: 1 unuah (минимальный коллатерал)
- **BaseInterestRate**: 2% (базовая процентная ставка)
- **MaxInterestRate**: 50% (максимальная процентная ставка)
- **MaxBorrowRatio**: 80% (максимальный коэффициент заимствования)

### Открытие позиций

#### LONG позиция (покупка с плечом)

```bash
# Открытие LONG позиции
nuahd tx leverage open-position \
  $TOKEN_DENOM \
  1000000 \
  unuah \
  2.0 \
  long \
  0.0 \
  1000000.0 \
  --from trader \
  --chain-id nuahchain \
  --fees 1000unuah \
  --gas 500000 \
  --keyring-backend test \
  -y
```

**Параметры:**
- `token-denom`: Деноминация токена для торговли
- `collateral-amount`: Размер коллатерала (в микро-единицах)
- `collateral-denom`: Деноминация коллатерала (обычно "unuah")
- `leverage`: Плечо (от 1.1 до 100)
- `side`: Сторона позиции ("long" или "short")
- `min-price`: Минимальная цена (для SHORT позиций)
- `max-price`: Максимальная цена (для LONG позиций)

#### SHORT позиция (продажа с плечом)

```bash
# Сначала создание пула ликвидности для SHORT позиций
nuahd tx leverage provide-liquidity \
  $TOKEN_DENOM \
  1000000000 \
  --from provider \
  --chain-id nuahchain \
  --fees 1000unuah \
  --gas 500000 \
  --keyring-backend test \
  -y

# Открытие SHORT позиции
nuahd tx leverage open-position \
  $TOKEN_DENOM \
  1000000 \
  unuah \
  2.0 \
  short \
  0.0 \
  1000000.0 \
  --from trader \
  --chain-id nuahchain \
  --fees 1000unuah \
  --gas 500000 \
  --keyring-backend test \
  -y
```

### Управление позициями

#### Просмотр позиций

```bash
# Все позиции трейдера
nuahd query leverage positions-by-trader $TRADER_ADDR --chain-id nuahchain

# Конкретная позиция
nuahd query leverage position 1 --chain-id nuahchain

# Все позиции по токену
nuahd query leverage positions-by-token $TOKEN_DENOM --chain-id nuahchain
```

#### Добавление коллатерала

```bash
# Добавление коллатерала к позиции
nuahd tx leverage add-collateral \
  1 \
  500000 \
  unuah \
  --from trader \
  --chain-id nuahchain \
  --fees 1000unuah \
  --gas 200000 \
  --keyring-backend test \
  -y
```

#### Удаление коллатерала

```bash
# Удаление коллатерала из позиции
nuahd tx leverage remove-collateral \
  1 \
  200000 \
  unuah \
  --from trader \
  --chain-id nuahchain \
  --fees 1000unuah \
  --gas 200000 \
  --keyring-backend test \
  -y
```

#### Закрытие позиции

```bash
# Закрытие позиции
nuahd tx leverage close-position \
  1 \
  0.0 \
  1000000.0 \
  --from trader \
  --chain-id nuahchain \
  --fees 1000unuah \
  --gas 500000 \
  --keyring-backend test \
  -y
```

### Система кредитования

#### Предоставление ликвидности

```bash
# Предоставление ликвидности в пул
nuahd tx leverage provide-liquidity \
  $TOKEN_DENOM \
  1000000000 \
  --from provider \
  --chain-id nuahchain \
  --fees 1000unuah \
  --gas 500000 \
  --keyring-backend test \
  -y
```

#### Просмотр пулов ликвидности

```bash
# Все пулы ликвидности
nuahd query leverage lending-pools --chain-id nuahchain

# Конкретный пул
nuahd query leverage lending-pool $TOKEN_DENOM --chain-id nuahchain

# Позиции заимствования
nuahd query leverage borrow-positions-by-borrower $TRADER_ADDR --chain-id nuahchain
```

### Ликвидация позиций

```bash
# Ликвидация позиции
nuahd tx leverage liquidate-position \
  1 \
  --from liquidator \
  --chain-id nuahchain \
  --fees 1000unuah \
  --gas 500000 \
  --keyring-backend test \
  -y
```

## 🔍 Мониторинг и аналитика

### Просмотр статистики

```bash
# Общая статистика модуля
nuahd query leverage stats --chain-id nuahchain

# Параметры модуля
nuahd query leverage params --chain-id nuahchain

# Все позиции
nuahd query leverage all-positions --chain-id nuahchain
```

### Проверка ликвидности

```bash
# Проверка, можно ли ликвидировать позицию
nuahd query leverage can-liquidate 1 --chain-id nuahchain
```

## ⚙️ Технические детали

### Расчет цены ликвидации

Для LONG позиций:
```
liquidation_price = entry_price * (1 - maintenance_margin) / leverage
```

Для SHORT позиций:
```
liquidation_price = entry_price * (1 + maintenance_margin) / leverage
```

### Расчет PnL

Для LONG позиций:
```
pnl = (current_price - entry_price) * position_size
```

Для SHORT позиций:
```
pnl = (entry_price - current_price) * position_size
```

### Процентные ставки

Процентная ставка рассчитывается динамически на основе коэффициента использования пула:

```
interest_rate = base_rate + (utilization_rate * interest_multiplier)
```

Где:
- `utilization_rate = total_borrowed / total_liquidity`
- Максимальная ставка ограничена параметром `max_interest_rate`

## 🛡️ Управление рисками

### Автоматическая ликвидация

Позиции автоматически ликвидируются когда:
1. Цена достигает цены ликвидации
2. Коллатерал становится меньше минимального требования

### Slippage Protection

- **LONG позиции**: Цена покупки не должна превышать `max_price`
- **SHORT позиции**: Цена продажи не должна быть ниже `min_price`

### Ограничения

- Максимальное плечо: 100x
- Минимальный коллатерал: 1 unuah
- Максимальный размер позиции: 1B токенов
- Максимальный коэффициент заимствования: 80%

## 🔧 Разработка и тестирование

### Запуск тестов

```bash
# Запуск всех тестов модуля
go test ./x/leverage/... -v

# Запуск конкретного теста
go test ./x/leverage/keeper -run TestLeverageMathematics -v
```

### Тестирование на запущенной ноде

1. Запустите ноду: `nuahd start --home ~/.nuahd`
2. Создайте аккаунты и профинансируйте их
3. Создайте пользовательский токен
4. Протестируйте все функции модуля

## 📊 Примеры использования

### Полный цикл торговли

```bash
# 1. Создание и финансирование аккаунтов
nuahd keys add trader --keyring-backend test
TRADER_ADDR=$(nuahd keys show trader -a --keyring-backend test)
nuahd tx bank send alice $TRADER_ADDR 1000000000unuah --from alice --chain-id nuahchain --fees 1000unuah --keyring-backend test -y

# 2. Создание токена
nuahd tx usertoken create-user-token testcoin "Test Coin" "TEST" 6 --from alice --chain-id nuahchain --fees 1000unuah --gas 2000000 --keyring-backend test -y
TOKEN_DENOM="factory/nuah1f7u8xmej7mwa5xvuew7xk3ls3m4f5zemmzyx2x/testcoin"

# 3. Установление цены через покупку
nuahd tx usertoken buy-tokens $TOKEN_DENOM 1000000unuah 1000000 --from alice --chain-id nuahchain --fees 1000unuah --keyring-backend test -y

# 4. Создание пула ликвидности для SHORT
nuahd keys add provider --keyring-backend test
PROVIDER_ADDR=$(nuahd keys show provider -a --keyring-backend test)
nuahd tx bank send alice $PROVIDER_ADDR 1000000000unuah --from alice --chain-id nuahchain --fees 1000unuah --keyring-backend test -y
nuahd tx leverage provide-liquidity $TOKEN_DENOM 1000000000 --from provider --chain-id nuahchain --fees 1000unuah --gas 500000 --keyring-backend test -y

# 5. Открытие LONG позиции
nuahd tx leverage open-position $TOKEN_DENOM 1000000 unuah 2.0 long 0.0 1000000.0 --from trader --chain-id nuahchain --fees 1000unuah --gas 500000 --keyring-backend test -y

# 6. Открытие SHORT позиции
nuahd tx leverage open-position $TOKEN_DENOM 1000000 unuah 2.0 short 0.0 1000000.0 --from trader --chain-id nuahchain --fees 1000unuah --gas 500000 --keyring-backend test -y

# 7. Просмотр позиций
nuahd query leverage positions-by-trader $TRADER_ADDR --chain-id nuahchain

# 8. Закрытие позиций
nuahd tx leverage close-position 1 0.0 1000000.0 --from trader --chain-id nuahchain --fees 1000unuah --gas 500000 --keyring-backend test -y
nuahd tx leverage close-position 2 0.0 1000000.0 --from trader --chain-id nuahchain --fees 1000unuah --gas 500000 --keyring-backend test -y
```

## 🚨 Важные замечания

1. **Единицы измерения**: Все суммы указываются в микро-единицах (1 unuah = 1,000,000 микро-unuah)
2. **Газ**: Для сложных операций (открытие/закрытие позиций) используйте высокий лимит газа (500,000+)
3. **Цены**: Цены рассчитываются через bonding curve модуля usertoken
4. **Ликвидность**: SHORT позиции требуют предварительного создания пула ликвидности
5. **Безопасность**: Всегда проверяйте slippage protection для защиты от неблагоприятных движений цены

## 📞 Поддержка

При возникновении проблем:
1. Проверьте логи ноды: `nuahd logs`
2. Убедитесь в достаточном количестве газа
3. Проверьте балансы аккаунтов
4. Убедитесь в корректности параметров транзакций

---

**Модуль leverage предоставляет мощный инструментарий для маржинальной торговли с полной интеграцией в экосистему NUAH Chain.**

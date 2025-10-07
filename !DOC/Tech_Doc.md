# Техническое задание: Модуль создания пользовательских токенов с Bonding Curve для NUAH-сети

## 1. Общая архитектура

Система состоит из двух основных модулей на базе Cosmos SDK:

### 1.1 Базовые модули
- **x/tokenfactory** - создание и управление токенами (на основе Osmosis TokenFactory)
- **x/bondingcurve** - управление bonding curve и торговлей с плечом
- **x/leveragedex** - торговля с плечом на DEX

### 1.2 Зависимости
- **x/bank** - управление балансами и переводы токенов
- **x/auth** - аутентификация аккаунтов
- **x/gov** - управление параметрами модуля
- Existing **TokenFactory** (Osmosis) - базовый функционал создания токенов[1][2][3]

## 2. Структуры данных

### 2.1 Token Metadata
```proto
message Token {
  string creator = 1;
  string denom = 2;          // factory/{creator}/{subdenom}
  string name = 3;           // уникальное имя
  string symbol = 4;         // уникальный символ
  string image = 5;          // IPFS/HTTPS URL
  string description = 6;    // описание токена
  uint64 created_at = 7;     // timestamp создания
  TokenDistribution distribution = 8;
  TokenState state = 9;
}

message TokenDistribution {
  string total_supply = 1;           // 100,000,000.000000
  string bonding_curve_supply = 2;   // 30,000,000
  string platform_wallet = 3;       // 10,000,000
  string referral_wallet = 4;       // 10,000,000
  string ai_ceo_wallet = 5;         // 40,000,000
  string founder_reserved = 6;      // 10,000,000
  bool founder_claimed = 7;         // статус claim
  uint64 founder_claim_deadline = 8; // deadline для claim
}

message TokenState {
  string bonding_curve_sold = 1;    // проданные токены на кривой
  string current_price = 2;         // текущая цена
  bool curve_completed = 3;         // кривая завершена
  bool dex_trading_enabled = 4;     // торговля на DEX включена
  bool soft_lock_enabled = 5;       // soft lock активен
}
```

### 2.2 Bonding Curve State
```proto
message BondingCurvePool {
  string denom = 1;
  string reserve_nuah = 2;         // резерв NUAH
  string reserve_ndollar = 3;      // резерв NDOLLAR
  string tokens_sold = 4;          // проданные токены
  string current_price_nuah = 5;   // текущая цена в NUAH
  string current_price_ndollar = 6; // текущая цена в NDOLLAR
}
```

### 2.3 Margin Position
```proto
message MarginPosition {
  uint64 id = 1;
  string trader = 2;
  string denom = 3;
  string collateral_denom = 4;     // NUAH или NDOLLAR
  string collateral_amount = 5;    // залог
  string position_size = 6;        // размер позиции
  string entry_price = 7;          // цена входа
  uint32 leverage = 8;             // плечо (1-100)
  PositionType type = 9;           // LONG/SHORT
  uint64 created_at = 10;
  string liquidation_price = 11;   // цена ликвидации
}

enum PositionType {
  LONG = 0;
  SHORT = 1;
}
```

## 3. Основные сообщения (Messages)

### 3.1 Token Management
```proto
// Создание нового токена
message MsgCreateToken {
  string creator = 1;
  string name = 2;
  string symbol = 3;
  string image = 4;
  string description = 5;
}

// Выкуп зарезервированных токенов основателем
message MsgFounderClaim {
  string founder = 1;
  string denom = 2;
}
```

### 3.2 Bonding Curve Trading
```proto
message MsgBuyFromCurve {
  string trader = 1;
  string denom = 2;
  string payment_denom = 3;     // NUAH или NDOLLAR
  string payment_amount = 4;
  string min_tokens_out = 5;    // slippage protection
}

message MsgSellToCurve {
  string trader = 1;
  string denom = 2;
  string token_amount = 3;
  string payment_denom = 4;     // NUAH или NDOLLAR
  string min_payment_out = 5;   // slippage protection
}
```

### 3.3 Margin Trading
```proto
message MsgOpenMarginPosition {
  string trader = 1;
  string denom = 2;
  string collateral_denom = 3;
  string collateral_amount = 4;
  uint32 leverage = 5;          // 1-100
  PositionType type = 6;        // LONG/SHORT
  string min_position_size = 7; // slippage protection
}

message MsgCloseMarginPosition {
  string trader = 1;
  uint64 position_id = 2;
  string min_payout = 3;        // slippage protection
}
```

## 4. Ключевые функции Keeper

### 4.1 Token Creation
```go
func (k Keeper) CreateToken(ctx sdk.Context, msg *MsgCreateToken) error {
    // 1. Проверка уникальности имени и символа
    // 2. Создание токена через TokenFactory
    // 3. Инициализация bonding curve pool
    // 4. Распределение токенов по кошелькам
    // 5. Активация founder claim timer (1 час)
}
```

### 4.2 Bonding Curve Pricing
```go
func (k Keeper) CalculateBuyPrice(tokensSold sdk.Dec) sdk.Dec {
    // Price = 0.0002 + (tokensSold / 30,000,000) * (1.0 - 0.0002)
    ratio := tokensSold.Quo(sdk.NewDec(30_000_000))
    return sdk.NewDecWithPrec(2, 4).Add(ratio.Mul(sdk.NewDecWithPrec(9998, 4)))
}

func (k Keeper) CalculateTokensOut(paymentAmount sdk.Dec, currentSupply sdk.Dec) sdk.Dec {
    // Интегральное вычисление для точного расчета токенов
    // Учитывает изменение цены в процессе покупки
}
```

### 4.3 Margin Trading Logic
```go
func (k Keeper) OpenMarginPosition(ctx sdk.Context, msg *MsgOpenMarginPosition) error {
    // 1. Проверка залога
    // 2. Расчет размера позиции с плечом
    // 3. Расчет цены ликвидации
    // 4. Создание позиции
    // 5. Блокировка залога
}

func (k Keeper) CalculateLiquidationPrice(entryPrice sdk.Dec, leverage uint32, posType PositionType) sdk.Dec {
    // Для LONG: liquidationPrice = entryPrice * (1 - 1/leverage * 0.9)
    // Для SHORT: liquidationPrice = entryPrice * (1 + 1/leverage * 0.9)
    // 0.9 - коэффициент безопасности (10% буфер)
}
```

## 5. End-Block обработчик

```go
func (k Keeper) EndBlock(ctx sdk.Context) {
    // 1. Проверка истечения founder claim (1 час)
    k.ProcessFounderClaimDeadlines(ctx)

    // 2. Проверка позиций на ликвидацию
    k.ProcessLiquidations(ctx)

    // 3. Активация DEX торговли для завершенных кривых
    k.ProcessCurveCompletions(ctx)
}
```

## 6. Критические нюансы и риски

### 6.1 Безопасность
- **Reentrancy защита** для всех функций торговли[4][5]
- **Rate limiting** для предотвращения flash-loan атак[5][6]
- **Slippage protection** обязателен для всех торговых операций
- **Oracle manipulation защита** через Time-Weighted Average Price (TWAP)[7]

### 6.2 Ликвидность и экономика
- **MEV защита** через batch-обработку транзакций[8]
- **Front-running защита** через commit-reveal схему или временные задержки
- **Арбитражная защита** между bonding curve и DEX ценами

### 6.3 Техническая реализация
- **Точность вычислений**: использование sdk.Dec с высокой точностью для избежания ошибок округления
- **Газ-оптимизация**: кэширование часто используемых расчетов
- **Параметризация**: все константы должны быть настраиваемыми через governance

### 6.4 Compliance и управление
- **Pause механизм** для экстренного отключения торговли
- **Upgrade path** для обновления логики bonding curve
- **Monitoring** всех ключевых метрик и событий

### 6.5 Интеграция с существующими модулями
- **IBC compatibility** для кросс-чейн переводов[6][9]
- **CosmWasm интеграция** для расширенной логики[10][11]
- **Совместимость с Osmosis DEX** для плавного перехода

### 6.6 Специфичные риски margin trading
- **Liquidation cascades** - массовые ликвидации при волатильности
- **Insufficient liquidity** при больших позициях
- **Price manipulation** через concentrated trading

## 7. Параметры модуля

```proto
message Params {
  string token_creation_fee = 1;      // Комиссия за создание (NUAH)
  uint64 founder_claim_period = 2;    // Период для claim (секунды)
  string min_collateral_ratio = 3;    // Мин. коэффициент залога (0.1 = 10%)
  uint32 max_leverage = 4;            // Максимальное плечо (100)
  string liquidation_penalty = 5;     // Штраф за ликвидацию (0.05 = 5%)
  string protocol_fee_rate = 6;       // Комиссия протокола (0.003 = 0.3%)
}
```

Это техническое задание покрывает все основные аспекты создания комплекса модулей для пользовательских токенов с bonding curve и маржинальной торговлей в экосистеме Cosmos SDK, с учетом критических нюансов безопасности и производительности.

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

<!-- [1](https://docs.injective.network/defi/tokens/token-factory)
[2](https://github.com/strangelove-ventures/tokenfactory)
[3](https://docs.osmosis.zone/osmosis-core/modules/tokenfactory/)
[4](https://www.range.org/blog/terra-ibc-hooks-exploit-analysis)
[5](https://blog.cosmos.network/ibc-rate-limits-elevating-cross-chain-security-to-the-next-level-15ce193ea7a3)
[6](https://www.range.org/blog/ibc-rate-limits-introduction-and-state-of-the-art-1-3)
[7](https://everstake.one/blog/what-is-osmosis-and-how-it-works)
[8](https://github.com/tendermint/liquidity)
[9](https://ibcprotocol.dev/blog/trust-assumptions-in-interoperability)
[10](https://docs.osmosis.zone/osmosis-core/modules/cosmwasmpool/)
[11](https://cosmwasm.cosmos.network)
[12](https://docs.osmosis.zone/overview/educate/osmosis)
[13](https://forum.cosmos.network/t/chips-discussion-phase-intent-centric-automation-cosmos-sdk-module-on-cosmos-hub/12283)
[14](https://tutorials.ignite.com/how-to-create-the-tokenfactory-module/)
[15](https://osom.finance/insights-sub/learn/terms/bonding-curve)
[16](https://webisoft.com/articles/cosmos-sdk/)
[17](https://hackmd.io/@tendermint-devx/Hkgm5-6Nr6)
[18](https://www.cryptohopper.com/currencies/detail?currency=OSMO)
[19](https://www.sotatek.com/blogs/cosmos-101-interconnected-blockchain-explained/)
[20](https://www.zeeve.io/blog/why-cosmos-sdk-stands-among-the-top-choices-to-build-depin-today/)
[21](https://www.binance.com/en/square/post/321378)
[22](https://www.rapidinnovation.io/post/build-private-blockchains-with-cosmos-sdk)
[23](https://docs.terra.money/develop/examples/token-factory)
[24](https://www.reddit.com/r/thegraph/comments/u9sik2/bonding_curve/)
[25](http://www.diva-portal.org/smash/get/diva2:1987380/FULLTEXT01.pdf)
[26](https://hackmd.io/fW6OLYVkTcS-TM3Y_MS6Zw)
[27](https://onramp.money/partners/osmosis/)
[28](https://blog.cosmos.network/defi-in-cosmos-meet-these-4-cross-chains-dexs-1df23dd413b0)
[29](https://www.kraken.com/features/margin-trading/cosmos)
[30](https://www.youtube.com/watch?v=pIpVDJLbKFQ)
[31](https://docs.cosmos.network/main/build/modules/distribution)
[32](https://blog.oqtacore.com/mastering-cosmos-sdk/)
[33](https://supra.com/academy/cosmos-zones/)
[34](https://beincrypto.com/learn/osmosis-crypto/)
[35](https://docs.cosmos.network/v0.45/modules/)
[36](https://www.coinbase.com/learn/advanced-trading/what-is-an-automated-market-maker-amm)
[37](https://aofiee.dev/1-work-from-home-application-blockchain-cosmos-network/)
[38](https://www.gocrypto.com/blog/what-is-an-automated-market-maker-amm)
[39](https://tutorials.ignite.com)
[40](https://www.dydx.xyz/crypto-learning/what-is-cosmos)
[41](https://hackmd.io/@gPsqrfHCRuG5HWBqx0WFeQ/Bk0Yl9CCN)
[42](https://github.com/ixoworld/bonds)
[43](https://www.dydx.xyz/crypto-learning/bonding-curve)
[44](https://blog.cosmos.network/elys-network-the-universal-liquidity-layer-secured-by-atom-32b31f05f0bb)
[45](https://www.antiersolutions.com/blogs/how-to-build-a-memecoin-launchpad-with-an-integrated-bonding-curve/)
[46](https://leastauthority.com/static/publications/LeastAuthority_Tendermint_Cosmos_SDK_Liquidity_Module_Final_Audit_Report.pdf)
[47](https://procarems.co.za/index.php/2025/07/23/why-ibc-transfers-airdrops-and-wallet-security-matter-more-than-ever-for-cosmos-users/)
[48](https://www.antiersolutions.com/blogs/how-to-build-a-bonding-curve-decentralized-crypto-exchange-software/)
[49](https://github.com/BlockScience/Risk-Adjusted-Bonding-Curves)
[50](https://cosmos.network)
[51](https://wasteconcern.org/why-ibc-transfers-and-secret-network-staking-are-game-changers-for-cosmos-users/)
[52](https://hackmd.io/@tendermint-devx/rJKM1TvST) -->



------

steps

# Этапный план реализации модулей для пользовательских токенов с Bonding Curve и маржинальной торговлей

## Этап 1. Инициализация и подготовка окружения
1. Настроить локальную сеть на основе Cosmos SDK (v0.50+) и Osmosis TokenFactory.
2. Добавить зависимые модули: x/bank, x/auth, x/gov.
3. Подключить репозиторий с исходниками TokenFactory и убедиться, что базовые MsgCreateDenom и MsgMint работают.
4. Написать smoke-тест для создания кастомного токена через существующий TokenFactory.

**Критерий проверки:**
– Успешное создание нового denom и mint начального запаса;
– Автоматический перевод токенов создателю.

***

## Этап 2. Расширение TokenFactory: MsgCreateToken
1. Реализовать собственное сообщение MsgCreateToken, инкапсулирующее параметры name/symbol/image/description.
2. В Keeper.CreateToken:
   - Проверка уникальности name и symbol;
   - Вызов TokenFactory для создания denom;
   - Запись metadata в on-chain хранилище.
3. Запрограммировать распределение 100 M токенов по кошелькам: bonding_curve, platform, referral, ai_ceo, founder.
4. Реализовать таймер founder_claim_deadline (1 час) через EndBlock.

**Критерий проверки:**
– После MsgCreateToken в state хранится корректный Token и распределение;
– Founder_deadline устанавливается и хранится;
– Smoke-тест создания двух токенов с разными именами.

***

## Этап 3. Bonding Curve: первичная стадия торговли
1. Создать x/bondingcurve модуль с Keeper, Store и MsgBuyFromCurve/MsgSellToCurve.
2. Реализовать функцию CalculateBuyPrice и точное интегральное вычисление токенов по формуле
   $$Price = 0.0002 + \frac{\text{sold}}{30\,000\,000}\times(1 - 0.0002)$$.
3. При покупке:
   - Проверка max_supply_curve = 30 M;
   - Перевод NUAH/NDOLLAR в модульный аккаунт;
   - Mint и распределение токенов.
4. При продаже: обратная логика с burn и возвратом оплаты.
5. Тест-кейсы на граничные условия (покупка всех 30 M, частичные покупки).

**Критерий проверки:**
– Корректная цена при разных sold;
– Баланс module-account меняется ожидаемо;
– Юнит-тесты на покупку/продажу с учётом slippage.

***

## Этап 4. Автоматическая активация DEX
1. В EndBlock проверять, достигнута ли кривая (sold == 30 M).
2. При достижении:
   - Установить TokenState.curve_completed = true;
   - Создать пул на DEX с равными резервами (30 M TOKEN : 30 M NUAH);
   - Открыть торговлю на DEX (x/liquidity).
3. Soft-lock: до completion запрещать MsgBuyFromDex/MsgSellToDex модулю x/leveragedex.

**Критерий проверки:**
– После выкупа 30 M активируется пул DEX;
– Запросы на DEX-торговлю до этого шага отклоняются, после — проходят.

***

## Этап 5. Реализация маржинальной торговли
1. Создать x/leveragedex модуль с MsgOpenMarginPosition и MsgCloseMarginPosition.
2. При открытии позиции:
   - Проверить collateral и рассчитать максимальный размер позиции = collateral × leverage;
   - Вычислить цену ликвидации:
     – для LONG: $$P_{\rm liq} = P_{\rm entry}\times(1 - \tfrac{1}{L}\times0.9)$$;
     – для SHORT: $$P_{\rm liq} = P_{\rm entry}\times(1 + \tfrac{1}{L}\times0.9)$$.
   - Заблокировать collateral в модульном аккаунте.
3. При закрытии: вычислить PnL и вернуть collateral с учётом прибыли/убытка.
4. EndBlock: ProcessLiquidations — ликвидировать позиции с ценой текущего рынка ≤ P_​liq.

**Критерий проверки:**
– Открытие/закрытие позиций с разными плечами и типами;
– Ликвидация позиций при достижении цены;
– Тесты на рассчёт P_​liq и расчёт PnL.

***

## Этап 6. Админ-функции и governance
1. Добавить параметры Params: token_creation_fee, max_leverage, protocol_fee_rate и др.
2. Реализовать MsgUpdateParams (x/gov) для изменения настроек через proposal.
3. Добавить MsgFreezeToken/MsgUnfreezeToken и MsgPauseTrading/MsgResumeTrading.
4. Обработчики в Keeper, проверка paused flag перед всеми торговыми операциями.

**Критерий проверки:**
– Изменение параметров через governance;
– Freeze/Unfreeze действительно блокирует торговлю и переводы;
– Юнит-тесты на защиту paused-состояния.

***

## Этап 7. Безопасность, мониторинг и подготовка к деплою
1. Провести аудит кода на reentrancy, overflow, front-running.
2. Внедрить rate-limit и TWAP-oracle для защиты от манипуляций.
3. Настроить мониторинг ключевых метрик: объем кривой, открытые позиции, ликвидации.
4. Обновить документацию и написать примеры CLI/REST.
5. Развернуть тестовую сеть и провести end-to-end тестирование со сценариями создания токена, торговли и маржи.

**Критерий завершения:**
– Успешное прохождение всех тестов в тестнете;
– Подготовка релизных артефактов (Docker, upgrade handler).

***

Каждый этап завершается набором автоматических юнит- и интеграционных тестов, а также smoke-тестами в локальном и тестовом окружении, прежде чем перейти к следующему шагу.

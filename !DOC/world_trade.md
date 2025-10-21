🎯 Цель проекта
Создать систему для моего блокчейна на Cosmos SDK, которая позволяет пользователям торговать любыми активами— фиатными валютами, сырьевыми ресурсами (золото, нефть, серебро), акциями компаний и криптовалютами — внутри одного блокчейна, без подключения к внешним биржам.
Торговля осуществляется на доверии и без внешней ликвидности.Всё обращение активов происходит через внутренние токены, чьи цены отражают реальные рыночные значения.

💰 Принцип работы
1. Пользователь выбирает актив, например "EUR", "GOLD", "BTC", "TSLA".
2. Если этого актива ещё нет в системе — он создаётся автоматически.
    * Создаётся токен (asset/EUR, asset/GOLD, и т.д.).
    * Вносятся базовые параметры (symbol, decimals, type).
3. Модуль oracle получает текущую цену из внешних API.
4. Пользователь покупает актив за внутренний стейблкоин (NDOLLAR) или нативный токен (NUAH).
5. При покупке:
    * Минтится актив-токен по текущей цене.
    * Сжигается эквивалентное количество NDOLLAR.
6. При продаже:
    * Актив-токен сжигается.
    * Эмитируется NDOLLAR по текущей цене актива.
    * Прибыль пользователя возникает за счёт контролируемой эмиссии NDOLLAR, а убыток возвращает токены в систему.
7. Все токеномические балансы поддерживаются через резервный пул, комиссии и спреды.

💎 Токеномика и баланс системы
Базовые токены:
* NUAH — нативный токен сети (для комиссий, управления, обеспечения).
* NDOLLAR — внутренний стейблкоин (эквивалент USD), используемый для торгов и расчётов.
Механика:
* При покупке актива:→ Минтится новый актив-токен.→ Сжигается NDOLLAR.
* При продаже:→ Сжигается актив-токен.→ Эмитируется NDOLLAR по текущей цене.
* Разница между ценой покупки и продажи создаёт прибыль/убыток пользователя.
* Эмиссия и сжигание NDOLLAR контролируются модулями fees и stablecoin.
Баланс доверия:
* Комиссии, спреды и ликвидации формируют резервный пул — он компенсирует прибыль пользователей и предотвращает гиперинфляцию NDOLLAR.
* Торговля с плечом, маржинальные ставки и ликвидации создают дополнительный приток NDOLLAR в пул.
* Все участники видят прозрачный баланс резервов на блокчейне.
Доход разработчика (владельца сети):
* Комиссии за сделки (buy/sell).
* Комиссии за торговлю с плечом.
* Небольшой торговый спред (например, ±0.1–0.5%).
* Потенциальная эмиссия части нативных токенов в “фонд разработчика”.

🧩 Модули системы
1. assets — управление активами
* Автоматическое создание активов при первом обращении.
* Минтинг и сжигание токенов при торговых операциях.
* Хранение метаданных (symbol, type, price_source, decimals).
* Взаимодействие с oracle и trade.
2. oracle — получение цен
* Получение рыночных цен через API (CoinGecko, Forex, Metals API, Yahoo Finance и т.п.).
* Автоматическое добавление новых активов в ценовую таблицу.
* Поддержка многоканальных источников и усреднения данных.
* Обновление цен через валидаторов или оффчейн-ноды.
3. trade — торговля активами
* Основная логика покупки и продажи.
* Проверка существования актива → автоматическое создание.
* Получение цены из oracle.
* Расчёт комиссий, спреда и итоговой суммы сделки.
* Минтинг/сжигание активов и NDOLLAR.
4. leverage — торговля с плечом
* Торговля с плечом до ×100.
* Хранение позиций и расчёт PnL.
* Ликвидации при нарушении маржинальных требований.
* Интеграция с collateral и risk.
5. collateral — залоги
* Управление обеспечением позиций пользователей.
* Заморозка токенов в залоге.
* Автоматические ликвидации при нехватке обеспечения.
* Возврат залога при закрытии позиции.
6. stablecoin — эмиссия NDOLLAR
* Контроль эмиссии и сжигания NDOLLAR.
* Балансировка системы (дефицит/профицит NDOLLAR).
* Поддержание “мягкого” курса к USD (soft peg).
* Работа с резервным пулом и комиссиями.
7. fees — комиссии и резервный пул
* Хранение всех торговых комиссий.
* Сжигание части комиссий или отправка в резервный пул.
* Распределение дохода между сетью и разработчиком.
8. risk — управление рисками
* Настройка лимитов плеча.
* Мониторинг устойчивости пула.
* Автоматическое повышение комиссий при перегреве рынка.

⚙️ Технические детали
* Платформа: Cosmos SDK + CometBFT
* Язык: Golang
* Архитектура: модульная (каждый модуль — отдельный keeper и msg handler)
* Хранение: key-value store (KVStore)
* API: REST + gRPC
* Минтинг: только в момент покупки пользователем
* Сжигание: при продаже или закрытии позиции
* Резервный пул: отдельный модуль или субаккаунт сети
* Комиссии и баланс: динамически регулируются на уровне risk

💡 Ключевые особенности
Особенность	Описание
Автоматическое создание активов	Любой актив создаётся при первом запросе пользователя
Без внешней ликвидности	Вся торговля внутри сети, на доверии
Динамическая эмиссия	NDOLLAR создаётся и сжигается при сделках
Контроль прибыли	Резервный пул компенсирует прибыль и удерживает баланс
Плечо до ×100	Система поддерживает маржинальную торговлю
Гибкая архитектура	Можно добавлять новые активы, источники цен и типы торговли
📈 Пример сценария торговли
1. Пользователь хочет купить 1 GOLD.
    * Цена 2000 USD.
    * Сжигается 2000 NDOLLAR, минтится 1 asset/GOLD.
2. Через время цена растёт до 2200 USD.
    * Пользователь продаёт 1 GOLD.
    * Сжигается GOLD, эмитируется 2200 NDOLLAR.
    * Прибыль пользователя: +200 NDOLLAR.
    * Система удерживает 1% комиссии (22 NDOLLAR).
3. 22 NDOLLAR поступают в резервный пул → поддержание баланса эмиссии.

🧠 Цель системы
Создать универсальную доверительную торговую инфраструктуру,в которой пользователи могут свободно покупать и продавать любые активы,а баланс системы и прибыльность обеспечиваются внутренней экономикой токенов,без зависимости от внешних бирж или ликвидности.

—

🔹 Промпт 1. Базовый Cosmos SDK App + модуль assets
Цель: Создать скелет модуля assets и базовой proto-структурой.
Промпт:
Создай минимальный модуль x/assets.Требования:
* Каталоги:
    * proto/mychain/assets/v1/
    * x/assets/ (keeper, types, module)
    * app/, cmd/
* Создай proto/mychain/assets/v1/genesis.proto с:syntax = "proto3";
* package mychain.assets.v1;
* import "cosmos/base/v1beta1/coin.proto";
* option go_package = "x/assets/types";
* message GenesisState {
*   repeated Asset assets = 1;
* }
* message Asset {
*   string symbol = 1;
*   string name = 2;
*   string type = 3;
*   uint32 decimals = 4;
*   string status = 5;
* }
*
* Реализуй модуль инициализации/экспорта генезиса.
* Зарегистрируй модуль в app.go.

🔹 Промпт 2. Query сервис для assets
Цель: Добавить gRPC-запросы и Keeper API.
Промпт:
В proto/mychain/assets/v1/query.proto создай:

syntax = "proto3";
package mychain.assets.v1;
import "google/api/annotations.proto";
import "cosmos/base/query/v1beta1/pagiNUAHion.proto";
option go_package = "x/assets/types";

service Query {
  rpc Asset(QueryAssetRequest) returns (QueryAssetResponse) {
    option (google.api.http).get = "/mychain/assets/v1/assets/{symbol}";
  }
  rpc Assets(QueryAssetsRequest) returns (QueryAssetsResponse) {
    option (google.api.http).get = "/mychain/assets/v1/assets";
  }
}

message QueryAssetRequest { string symbol = 1; }
message QueryAssetResponse { Asset asset = 1; }
message QueryAssetsRequest { cosmos.base.query.v1beta1.PageRequest pagiNUAHion = 1; }
message QueryAssetsResponse { repeated Asset assets = 1; cosmos.base.query.v1beta1.PageResponse pagiNUAHion = 2; }
Реализуй Keeper и gRPC-сервер.Добавь CLI-команды:
* mychaind q assets asset EUR
* mychaind q assets assets
Тест:Создай запись в хранилище и запроси через CLI → возвращает JSON с Asset.

🔹 Промпт 3. MsgEnsureAsset — автоматическое создание актива
Цель: Добавить сообщение для автосоздания актива.
Промпт:
Создай proto/mychain/assets/v1/tx.proto:

syntax = "proto3";
package mychain.assets.v1;
import "cosmos/msg/v1/msg.proto";
import "google/api/annotations.proto";
option go_package = "x/assets/types";

service Msg {
  rpc EnsureAsset(MsgEnsureAsset) returns (MsgEnsureAssetResponse) {
    option (google.api.http).post = "/mychain/assets/v1/ensure";
  }
}

message MsgEnsureAsset {
  option (cosmos.msg.v1.signer) = "creator";
  string creator = 1;
  string symbol = 2;
}
message MsgEnsureAssetResponse { Asset asset = 1; }
Keeper должен:
* Проверить наличие актива;
* Если нет — создать Asset{symbol, name=symbol, type="unknown", decimals=2, status="active"};
* Событие assets.asset_created.
Тест:tx assets ensure EUR --from alice → q assets asset EUR возвращает актив.

🔹 Промпт 4. Модуль oracle (цены)
Цель: добавить proto и базовую функциональность оракула.
Промпт:
Создай proto/mychain/oracle/v1/oracle.proto:

syntax = "proto3";
package mychain.oracle.v1;
import "google/api/annotations.proto";
option go_package = "x/oracle/types";

message Price {
  string symbol = 1;
  string value = 2;
}

message QueryPriceRequest { string symbol = 1; }
message QueryPriceResponse { Price price = 1; }

service Query {
  rpc Price(QueryPriceRequest) returns (QueryPriceResponse) {
    option (google.api.http).get = "/mychain/oracle/v1/price/{symbol}";
  }
}

service Msg {
  rpc SetPrice(MsgSetPrice) returns (MsgSetPriceResponse) {
    option (google.api.http).post = "/mychain/oracle/v1/set-price";
  }
}

message MsgSetPrice {
  option (cosmos.msg.v1.signer) = "authority";
  string authority = 1;
  string symbol = 2;
  string value = 3;
}
message MsgSetPriceResponse {}
Реализуй Keeper с GetPrice/SetPrice.
Тест:tx oracle set-price GOLD 2000 → q oracle price GOLD → 2000.

🔹 Промпт 5. MsgBuyAsset — покупка актива за NDOLLAR
Цель: добавить логику покупки и mint.
Промпт:
Добавь в proto/mychain/assets/v1/tx.proto новое сообщение:

rpc BuyAsset(MsgBuyAsset) returns (MsgBuyAssetResponse) {
  option (google.api.http).post = "/mychain/assets/v1/buy";
}

message MsgBuyAsset {
  option (cosmos.msg.v1.signer) = "buyer";
  string buyer = 1;
  string symbol = 2;
  string amount_NDOLLAR = 3; // in uNDOLLAR
}
message MsgBuyAssetResponse {
  string base_amount = 1;
}
Keeper логика:
* EnsureAsset(Symbol)
* Получить цену из oracle
* Рассчитать количество актива
* Списать NDOLLAR у покупателя, минтнуть asset/SYMBOL
* Записать событие assets.asset_bought
Тест:
1. tx oracle set-price GOLD 2000
2. tx assets buy GOLD 1000NDOLLAR→ Баланс: 0.5 asset/GOLD.

🔹 Промпт 6. MsgSellAsset — продажа и сжигание
Цель: завершить торговый цикл.
Промпт:
Добавь в тот же tx.proto:

rpc SellAsset(MsgSellAsset) returns (MsgSellAssetResponse) {
  option (google.api.http).post = "/mychain/assets/v1/sell";
}

message MsgSellAsset {
  option (cosmos.msg.v1.signer) = "seller";
  string seller = 1;
  string symbol = 2;
  string base_amount = 3;
}
message MsgSellAssetResponse {
  string payout_NDOLLAR = 1;
}
Keeper:
* Проверить баланс asset/SYMBOL;
* Получить цену из oracle;
* Сжечь актив, эмитировать NDOLLAR пользователю;
* Событие assets.asset_sold.
Тест:Купить GOLD, поменять цену в oracle, продать → получить больше NDOLLAR.

🔹 Промпт 7. Модуль fees (комиссии)
Промпт:
Создай proto/mychain/fees/v1/fees.proto:

syntax = "proto3";
package mychain.fees.v1;
option go_package = "x/fees/types";

message Params {
  string trade_fee_rate = 1; // e.g. "0.003"
}
message QueryParamsRequest {}
message QueryParamsResponse { Params params = 1; }
Реализуй Params storage и интеграцию: при BuyAsset и SellAsset удерживай комиссию в NDOLLAR и отправляй на модульный аккаунт fees.
Тест:Комиссия попадает в fees-баланс.

🔹 Промпт 8. Модуль stablecoin (NDOLLAR эмиссия/сжигание)
Промпт:
Создай proto/osmosis/stablecoin/v1/stablecoin.proto:

syntax = "proto3";
package mychain.stablecoin.v1;
option go_package = "x/stablecoin/types";

message Stats {
  string total_minted = 1;
  string total_burned = 2;
  string outstanding = 3;
}

message QueryStatsRequest {}
message QueryStatsResponse { Stats stats = 1; }

service Query {
  rpc Stats(QueryStatsRequest) returns (QueryStatsResponse);
}
Реализуй Keeper с подсчётом эмиссии/сжигания, API для assets модуля (разрешённый минт/берн).
Тест:После BuyAsset → TotalBurned растёт, после SellAsset → TotalMinted растёт.

🔹 Промпт 9. Модуль risk (лимиты плеча)
Промпт:
Создай proto/mychain/risk/v1/risk.proto:

syntax = "proto3";
package mychain.risk.v1;
option go_package = "x/risk/types";

message RiskParams {
  string symbol = 1;
  string max_leverage = 2;
  string maintenance_margin = 3;
  string initial_margin = 4;
}
message MsgSetRiskParams {
  option (cosmos.msg.v1.signer) = "authority";
  string authority = 1;
  RiskParams params = 2;
}
message MsgSetRiskParamsResponse {}
service Msg {
  rpc SetRiskParams(MsgSetRiskParams) returns (MsgSetRiskParamsResponse);
}
service Query {
  rpc RiskParams(QueryRiskParamsRequest) returns (QueryRiskParamsResponse);
}
message QueryRiskParamsRequest { string symbol = 1; }
message QueryRiskParamsResponse { RiskParams params = 1; }
Тест:Установить параметры → q risk risk-params BTC.

🔹 Промпт 10. Модуль collateral (залог)
Промпт:
proto/mychain/collateral/v1/collateral.proto:

syntax = "proto3";
package mychain.collateral.v1;
option go_package = "x/collateral/types";

message Position {
  string owner = 1;
  string denom = 2;
  string amount = 3;
}

message MsgDeposit {
  option (cosmos.msg.v1.signer) = "depositor";
  string depositor = 1;
  string amount = 2;
}
message MsgDepositResponse {}

message MsgWithdraw {
  option (cosmos.msg.v1.signer) = "owner";
  string owner = 1;
  string amount = 2;
}
message MsgWithdrawResponse {}

service Msg {
  rpc Deposit(MsgDeposit) returns (MsgDepositResponse);
  rpc Withdraw(MsgWithdraw) returns (MsgWithdrawResponse);
}
service Query {
  rpc Collateral(QueryCollateralRequest) returns (QueryCollateralResponse);
}
message QueryCollateralRequest { string owner = 1; }
message QueryCollateralResponse { repeated Position positions = 1; }

🔹 Промпт 11. Модуль leverage (открытие позиции)
Промпт:
proto/mychain/leverage/v1/leverage.proto:

syntax = "proto3";
package mychain.leverage.v1;
option go_package = "x/leverage/types";

enum Side { SIDE_UNSPECIFIED = 0; SIDE_LONG = 1; SIDE_SHORT = 2; }

message Position {
  uint64 id = 1;
  string owner = 2;
  string symbol = 3;
  Side side = 4;
  string base_qty = 5;
  string entry_price = 6;
  string leverage = 7;
}

message MsgOpenPosition {
  option (cosmos.msg.v1.signer) = "owner";
  string owner = 1;
  string symbol = 2;
  Side side = 3;
  string quote_NDOLLAR = 4;
  string leverage = 5;
}
message MsgOpenPositionResponse { Position position = 1; }

message MsgClosePosition {
  option (cosmos.msg.v1.signer) = "owner";
  uint64 id = 1;
}
message MsgClosePositionResponse { string pnl = 1; }

service Msg {
  rpc OpenPosition(MsgOpenPosition) returns (MsgOpenPositionResponse);
  rpc ClosePosition(MsgClosePosition) returns (MsgClosePositionResponse);
}

service Query {
  rpc Position(QueryPositionRequest) returns (QueryPositionResponse);
}
message QueryPositionRequest { uint64 id = 1; }
message QueryPositionResponse { Position position = 1; }

🔹 Промпт 12. Модуль fees + stablecoin интеграция и отчёт устойчивости
Промпт:
Добавь proto/mychain/stablecoin/v1/query.proto:

service Query {
  rpc Coverage(QueryCoverageRequest) returns (QueryCoverageResponse);
}
message QueryCoverageRequest {}
message QueryCoverageResponse {
  string outstanding = 1;
  string reserve_balance = 2;
  string coverage_ratio = 3;
}
Keeper: coverage_ratio = reserve_balance / outstanding.
Тест:Сделай пару сделок, проверь изменение покрытия.

🔹 Промпт 13. Governance параметры
Промпт:
Для fees, risk, stablecoin добавь params.proto с:

message Params {
  string param_name = 1;
}
Зарегистрируй ParamKeyTable и доступ к изменению через gov.

🔹 Промпт 14. Сценарии e2e тестирования (через proto CLI)
Промпт:
Напиши e2e-тесты на основе CLI и gRPC-запросов:
1. Автосоздание нового актива;
2. Покупка и продажа;
3. Изменение цены через оракул;
4. Проверка комиссий и stablecoin статистики;
5. Позиции с плечом и PnL.

🔹 Промпт 15. Финал — документация API
Промпт:
Сгенерируй swagger (OpenAPI) файлы из .proto через buf generate и добавь их в /docs/api/.Добавь README с описанием всех gRPC сервисов (assets, oracle, fees, leverage, и т.д.).

⚙️ Результат по итогу всех шагов:
Ты получаешь полностью модульное Cosmos SDK-приложение:
* Все модули определены через Proto;
* Каждый Msg и Query с gRPC и REST endpoint’ами;
* Полная цепочка:oracle → assets → stablecoin → fees → risk → collateral → leverage;
* Всё тестируется по шагам и расширяется постепенно.

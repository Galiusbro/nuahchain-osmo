## Модуль Оракула

Этот документ описывает модуль `x/oracle`: назначение, модель данных, сообщения, запросы и точки интеграции.

### Обзор
- Оракул хранит цены активов и при необходимости поддерживает историю цен.
- Цены могут устанавливаться уполномоченным адресом через сообщение или автоматически подтягиваться из внешних API (Yahoo Finance через `APIIKeeper`).
- Модуль предоставляет gRPC/REST/CLI‑запросы для получения текущей цены и исторических записей.

### Ключевые компоненты
- `AppModule` (`x/oracle/module.go`)
  - Регистрирует Msg/Query серверы и gRPC‑Gateway маршруты.
  - Импорт/экспорт цен в генезисе.
  - Периодически обновляет цены в `BeginBlock` (каждые 300 блоков ~ 5 минут) с помощью `APIIKeeper` и сохраняет историю.
- `Keeper` (`x/oracle/keeper/keeper.go`)
  - Доступ к состоянию: `SetPrice`, `GetPrice`, `GetAllPrices`.
  - История цен: `SetPriceHistory`, `GetPriceHistory`, `GetLatestPriceHistory`.
- `APIIKeeper` (`x/oracle/keeper/sources.go`)
  - Расширяет `Keeper` возможностями внешних источников.
  - Получает цены и историю из Yahoo Finance (`sources.APIClient`).
  - Предоставляет категории и списки символов по умолчанию, вспомогательные методы и статистику.
- `Query Server` (`x/oracle/keeper/query_server.go`)
  - gRPC‑обработчики для запросов цен и истории с валидацией и разумными ограничениями.

### Модель данных
- Прототипы (`proto/osmosis/oracle/v1/oracle.proto`):
  - `Price { symbol, value, source, timestamp, confidence }`
  - `PriceHistoryEntry { symbol, value, source, timestamp, confidence, block_height }`
- Ключи хранилища (`x/oracle/types/keys.go`):
  - `PriceKeyPrefix = 0x01` → `PriceKey(symbol)`
  - `PriceHistoryKeyPrefix = 0x02` → `PriceHistoryKey(symbol, timestamp)`; сортировка по времени за счёт big‑endian; префикс‑хелпер для итерации.

### Сообщения
- `Msg SetPrice` (только для `authority`):
  - `MsgSetPrice { authority, symbol, value } → MsgSetPriceResponse`
  - Маршрут: `POST /osmosis/oracle/v1/set-price`
  - Генерирует событие `oracle.price_updated` с атрибутами `symbol`, `value`, `authority`.

### Запросы
- Сервис: `Query` (через proto в `x/oracle/types`).
  - `Price(QueryPriceRequest{symbol}) → QueryPriceResponse{price}`
    - REST: `GET /osmosis/oracle/v1/price/{symbol}`
    - Пример: `curl http://localhost:1317/osmosis/oracle/v1/price/BTC-USD`
  - `PriceHistory(QueryPriceHistoryRequest{symbol, start_time?, end_time?, limit?}) → QueryPriceHistoryResponse{entries}`
    - REST: `GET /osmosis/oracle/v1/price-history/{symbol}`
    - Query‑параметры: `start_time` (Unix timestamp), `end_time` (Unix timestamp), `limit` (по умолчанию: 100, макс: 1000)
    - Пример: `curl http://localhost:1317/osmosis/oracle/v1/price-history/BTC-USD?start_time=1700000000&end_time=1700086400&limit=100`
    - Без диапазона времени возвращаются последние записи; `limit` по умолчанию 100, максимум 1000.

### CLI
- Корень: `osmosisd q oracle`
- Текущая цена:
  - `osmosisd q oracle price [symbol]`
- История (флаги `--start-time`, `--end-time`, `--limit`):
  - `osmosisd q oracle price-history [symbol] --limit 200`

### Генезис
- Импорт: `InitGenesis` сохраняет переданные `prices` после валидации.
- Экспорт: `ExportGenesis` выгружает все сохранённые цены.

### Автообновления и история
- `BeginBlock` каждые 300 блоков:
  - Получает категории и списки символов через `APIIKeeper.GetDefaultSymbols()`.
  - Обновляет небольшой поднабор в каждой категории (до 3 символов) для соблюдения лимитов API.
  - При успехе сохраняет `PriceHistoryEntry` с текущим `block_height`.

### Внешний источник данных
- Yahoo Finance через `sources.APIClient`:
  - `UpdatePriceFromAPI(ctx, symbol)` записывает `Price` c полями `source`, `timestamp`, `confidence`.
  - `GetHistoricalData(ctx, symbol, period)` — опционально для интеграций.
  - `SearchSymbols(ctx, query)` — опционально.
  - По умолчанию доступен набор категорий и символов `GetDefaultSymbols()`.

### Валидация и лимиты
- Символ нормализуется `EnsureSymbol(strings.TrimSpace)`.
- Сервер запросов валидирует входные данные и ограничивает `limit` (по умолчанию 100, максимум 1000).

### События
- `oracle.price_updated` с атрибутами: `symbol`, `value`, `authority`.

### Рекомендации по интеграции
- Логика обновления в `BeginBlock` — по принципу «best‑effort»: ошибки логируются и пропускаются.
- Для внешних обновлений используйте методы `APIIKeeper` или `GetPriceWithFallback`, чтобы при отсутствии цены выполнить загрузку и запись.
- В продакшене учитывайте лимиты API и пакетную обработку; корректируйте списки символов.

### Тестирование
- См. `x/oracle/keeper/*_test.go` и `test_api_keeper.go` для примеров получения реальных данных и обновления по категориям.

### Безопасность
- Сообщение `MsgSetPrice` доступно только адресам‑владельцам полномочий (`authority`).
- Внешние данные помечаются `source` и `confidence` для последующей фильтрации потребителями.
- При смене источников учитывайте секреты доступа и ограничение частоты запросов.



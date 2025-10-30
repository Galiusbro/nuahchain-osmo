Perfect! Now I have comprehensive information. Let me create a complete documentation guide.

## Полная документация Yahoo Finance API v8

### Обзор версий API

Yahoo Finance не предоставляет официального API, однако имеются несколько неофициальных версий эндпоинтов, которые используются сообществом разработчиков. Актуально по состоянию на октябрь 2025 года работают следующие версии:[1][2][3]

- **v1** - поиск и базовые запросы
- **v6** - устарел (прекращен в 2023 году)
- **v7** - исторические данные (частично ограничен)
- **v8** - **современный основной эндпоинт** для исторических данных
- **v10** - сводная информация по компаниям

### Основные хосты для запросов

Существуют два хоста для запросов:[2]

| Хост | HTTP версия | Использование |
|------|-----------|---------|
| `query1.finance.yahoo.com` | HTTP/1.0 | Одноразовые запросы, без persistent connections |
| `query2.finance.yahoo.com` | HTTP/1.1 | Persistent connections, proxy, bulk запросы |

### Структура эндпоинтов v8

#### 1. **Chart Endpoint** (Исторические данные о ценах)

**URL:**
```
https://query1.finance.yahoo.com/v8/finance/chart/{ticker}
```

**Метод:** GET

**Полный набор параметров запроса:**

| Параметр | Тип | Значения | Описание |
|----------|-----|---------|---------|
| **range** | string | `1d`, `5d`, `1mo`, `3mo`, `6mo`, `1y`, `2y`, `5y`, `10y`, `ytd`, `max` | Предустановленный диапазон (альтернатива period1/period2) |
| **period1** | integer | Unix timestamp (секунды) | Начальная дата (вместо range) |
| **period2** | integer | Unix timestamp (секунды) | Конечная дата (вместо range) |
| **interval** | string | `1m`, `2m`, `5m`, `15m`, `30m`, `60m`, `90m`, `1h`, `1d`, `5d`, `1wk`, `1mo`, `3mo` | Интервал между данными (default: `1d`) |
| **events** | string | `history`, `div`, `split`, `earn` | Типы событий (разделяются через `\|`, например: `div\|split`) |
| **includePrePost** | boolean | `true`, `false` | Включать pre-market и post-market данные (default: `false`) |
| **includeAdjustedClose** | boolean | `true`, `false` | Включить скорректированную цену закрытия (default: `false`) |
| **region** | string | `US`, `GB`, `FR`, и т.д. | Регион (влияет на форматирование) |
| **lang** | string | `en-US`, `ru-RU`, и т.д. | Язык ответа |
| **corsDomain** | string | `finance.yahoo.com` | CORS домен для браузерных запросов |
| **useYfid** | boolean | `true`, `false` | Использовать yfid идентификаторы |

**Примеры запросов:**

```
# Дневные данные за 1 год
https://query1.finance.yahoo.com/v8/finance/chart/AAPL?range=1y&interval=1d

# 5-минутные данные за день с pre/post-market
https://query1.finance.yahoo.com/v8/finance/chart/AAPL?range=1d&interval=5m&includePrePost=true

# С событиями (дивиденды и сплиты)
https://query1.finance.yahoo.com/v8/finance/chart/AAPL?range=1y&interval=1d&events=div%7Csplit&includeAdjustedClose=true

# Используя Unix timestamps
https://query1.finance.yahoo.com/v8/finance/chart/AAPL?period1=1609459200&period2=1640995200&interval=1d
```

**Структура JSON ответа:**

```json
{
  "chart": {
    "result": [
      {
        "meta": {
          "currency": "USD",
          "symbol": "AAPL",
          "exchangeName": "NMS",
          "instrumentType": "EQUITY",
          "firstTradeDate": 345479400,
          "regularMarketTime": 1698768000,
          "gmtoffset": -18000,
          "timezone": "EST",
          "exchangeTimezoneName": "America/New_York",
          "regularMarketPrice": 189.95,
          "chartPreviousClose": 188.84,
          "priceHint": 2,
          "currentTradingPeriod": {
            "pre": {
              "timezone": "EST",
              "start": 1698768000,
              "end": 1698776400,
              "gmtoffset": -18000
            },
            "regular": {
              "timezone": "EST",
              "start": 1698776400,
              "end": 1698797200,
              "gmtoffset": -18000
            },
            "post": {
              "timezone": "EST",
              "start": 1698797200,
              "end": 1698811600,
              "gmtoffset": -18000
            }
          },
          "dataGranularity": "1d",
          "range": "1y",
          "validRanges": ["1d", "5d", "1mo", "3mo", "6mo", "1y", "2y", "5y", "10y", "ytd", "max"]
        },
        "timestamp": [1698148800, 1698235200, 1698321600],
        "indicators": {
          "quote": [
            {
              "open": [189.95, 190.24, 191.45],
              "high": [190.75, 191.85, 192.10],
              "low": [188.90, 189.50, 190.80],
              "close": [189.50, 190.10, 191.00],
              "volume": [52456789, 45123456, 38976543]
            }
          ],
          "adjclose": [
            {
              "adjclose": [189.50, 190.10, 191.00]
            }
          ]
        }
      }
    ],
    "error": null
  }
}
```

**JSONPath для извлечения данных:**[4]

```
$.chart.result[0].meta.symbol                              # Символ тикера
$.chart.result[0].timestamp[*]                            # Все временные метки
$.chart.result[0].indicators.quote[0].close               # Все цены закрытия
$.chart.result[0].indicators.quote[0].close[0]            # Первая цена закрытия
$.chart.result[0].indicators.adjclose[0].adjclose         # Скорректированные цены
```

***

### Дополнительные эндпоинты v8 и соседние версии

#### 2. **Quote Endpoint** (Текущие котировки) - v7

**Статус:** Частично ограничен в 2023+, требует User-Agent заголовок

**URL:**
```
https://query1.finance.yahoo.com/v7/finance/quote?symbols=AAPL,GOOG
```

**Параметры:**

| Параметр | Описание |
|----------|---------|
| **symbols** | Коды тикеров через запятую |
| **fields** | Специфические поля (EBITDA, shortRatio, priceToSales, и т.д.) |

***

#### 3. **Search Endpoint** (Поиск) - v1

**URL:**
```
https://query1.finance.yahoo.com/v1/finance/search?q=Apple
```

**Параметры:**

| Параметр | Описание |
|----------|---------|
| **q** | Поисковый запрос (название компании или символ) |

**Ответ содержит:**
- Список котировок (quotes)
- Новости (news)
- Лучшие совпадения (best match)

***

#### 4. **Options Endpoint** - v7

**URL:**
```
https://query1.finance.yahoo.com/v7/finance/options/{ticker}
```

**Параметры:**

| Параметр | Описание |
|----------|---------|
| **date** | Unix timestamp даты истечения опциона (опционально) |

**Ответ содержит:**
- Underlying symbol
- Expiration dates
- Call/Put options для каждой страйк-цены

***

#### 5. **Quote Summary Endpoint** - v10

**URL:**
```
https://query1.finance.yahoo.com/v10/finance/quoteSummary/{ticker}?modules=...
```

**Доступные модули:**[3][1]

```
assetProfile, summaryProfile, summaryDetail, esgScores, price,
incomeStatementHistory, incomeStatementHistoryQuarterly,
balanceSheetHistory, balanceSheetHistoryQuarterly,
cashflowStatementHistory, cashflowStatementHistoryQuarterly,
defaultKeyStatistics, financialData, calendarEvents, secFilings,
recommendationTrend, upgradeDowngradeHistory,
institutionOwnership, fundOwnership, majorDirectHolders,
majorHoldersBreakdown, insiderTransactions, insiderHolders,
netSharePurchaseActivity, earnings, earningsHistory,
earningsTrend, industryTrend, indexTrend, sectorTrend
```

**Пример:**
```
https://query1.finance.yahoo.com/v10/finance/quoteSummary/AAPL?modules=price,financialData,defaultKeyStatistics
```

***

#### 6. **Download Endpoint** - v7

**URL:**
```
https://query1.finance.yahoo.com/v7/finance/download/{ticker}?period1=...&period2=...
```

**Использование:** Получение исторических данных в CSV формате

**Параметры:**

| Параметр | Значения |
|----------|----------|
| **period1** | Unix timestamp начало |
| **period2** | Unix timestamp конец |
| **interval** | `1d`, `1wk`, `1mo` |
| **events** | `history`, `div`, `split` |

***

#### 7. **Spark Endpoint** - v7

**URL:**
```
https://query1.finance.yahoo.com/v7/finance/spark
```

**Описание:** Быстрый запрос для получения цен с графиками

---

### Обработка аутентификации и ограничений

**Проблемы, которые могут возникнуть:**[2][4]

1. **"Invalid Crumb"** - требуется извлечение crumb из браузера
2. **"Invalid Cookie"** - v7 endpoints нужны валидные cookies/crumbs
3. **HTTP 401 Unauthorized** - некоторые endpoints требуют User-Agent заголовок

**Решения:**

```python
# Добавить User-Agent заголовок
headers = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
}
response = requests.get(url, headers=headers)

# Для v8 обычно не требуется authentication
# Но рекомендуется добавить задержку между запросами
import time
time.sleep(1)  # 1 секунда между запросами
```

***

### Особенности v8 эндпоинта

**Преимущества:**[1][4]

- ✅ **Работает без аутентификации** для исторических данных
- ✅ **Разработка по состоянию на 2024-2025** актуально
- ✅ **Поддержка гранулярности до 1 минуты** для intraday данных
- ✅ **Возвращает структурированный JSON** с метаданными
- ✅ **Latency ~50 миллисекунд** для real-time котировок (60% улучшение vs 2023)

**Ограничения:**

- ❌ Нет официальной документации от Yahoo
- ❌ Может работать медленнее при массовых запросах
- ❌ Требует правильной обработки rate limiting (рекомендуется exponential backoff)
- ❌ Для некоторых криптовалют нужен суффикс `-USD`

***

### Примеры кода

**Python - получение исторических данных:**

```python
import requests
import json
from datetime import datetime, timedelta

def get_yahoo_chart_data(ticker, range_period="1mo", interval="1d"):
    """Получить исторические данные из v8 эндпоинта"""
    url = f"https://query1.finance.yahoo.com/v8/finance/chart/{ticker}"

    params = {
        "range": range_period,
        "interval": interval,
        "includeAdjustedClose": "true",
        "events": "div|split"
    }

    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
    }

    try:
        response = requests.get(url, params=params, headers=headers)
        response.raise_for_status()
        data = response.json()

        if data['chart']['result']:
            result = data['chart']['result'][0]
            timestamps = result['timestamp']
            closes = result['indicators']['quote'][0]['close']

            dates = [datetime.fromtimestamp(ts) for ts in timestamps]

            return list(zip(dates, closes))
        else:
            print(f"Ошибка: {data['chart']['error']}")
            return None

    except requests.exceptions.RequestException as e:
        print(f"Ошибка запроса: {e}")
        return None

# Использование
data = get_yahoo_chart_data("AAPL", range_period="6mo", interval="1d")
if data:
    for date, close_price in data[:5]:
        print(f"{date.date()}: ${close_price}")
```

***

### Рекомендации и best practices

1. **Используйте v8 для исторических данных** - самый стабильный вариант
2. **Добавляйте User-Agent заголовок** для всех запросов
3. **Используйте exponential backoff** для обработки rate limiting:

```python
import time
import random

def make_request_with_backoff(url, max_retries=3):
    for attempt in range(max_retries):
        try:
            response = requests.get(url, timeout=10)
            if response.status_code == 429:  # Rate limited
                wait_time = (2 ** attempt) + random.uniform(0, 1)
                time.sleep(wait_time)
                continue
            response.raise_for_status()
            return response
        except requests.exceptions.RequestException:
            if attempt == max_retries - 1:
                raise
            wait_time = (2 ** attempt) + random.uniform(0, 1)
            time.sleep(wait_time)
```

4. **Кэшируйте результаты** когда возможно
5. **Обработайте ошибки gracefully** - API может периодически меняться
6. **Для crypto** используйте формат: `BTC-USD`, `ETH-USD`

***

### Сравнение версий

| Версия | Статус | Использование | Требует аутентификации |
|--------|--------|---------------|----------------------|
| v1 | ✅ Активен | Поиск, новости | ❌ Нет |
| v6 | ❌ Убран | Котировки | ❌ Нет |
| v7 | ⚠️ Ограничен | Опции, загрузка | ⚠️ Иногда |
| v8 | ✅ **Основной** | **Исторические данные** | ❌ Нет |
| v10 | ✅ Активен | Финансовые данные | ⚠️ Иногда |

Версия v8 остается **самой надежной и актуальной** для получения исторических данных о ценах на финансовые инструменты по состоянию на октябрь 2025 года.[3][4][1][2]

[1](https://www.marketcalls.in/intraday/exploring-yahoo-finance-realtime-quotes-and-historical-data-feed-api.html)
[2](https://stackoverflow.com/questions/44030983/yahoo-finance-url-not-working)
[3](https://blog.coupler.io/yahoo-finance-to-excel/)
[4](https://stackoverflow.com/questions/76059562/yahoo-finance-api-get-quotes-returns-invalid-cookie)
[5](https://stackoverflow.com/questions/39676160/query-to-retrieve-stock-quotes-variation-from-a-single-day)
[6](https://virtual-finance-api.readthedocs.io/_/downloads/en/latest/pdf/)
[7](https://community.fabric.microsoft.com/t5/Power-Query/Yahoo-Finance-Power-Query-Link-Not-Working/m-p/4142239)
[8](https://www.reddit.com/r/algotrading/comments/1ivlhoz/yahoo_finance_api/)
[9](https://python-yahoofinance.readthedocs.io/en/latest/api.html)
[10](https://query1.finance.yahoo.com/v8/finance/chart/)
[11](https://github.com/ranaroussi/yfinance/issues/2474)
[12](https://hostagenda.com/blog/yahoo-finance-api-get-news)
[13](https://github.com/herval/yahoo-finance/issues/51)
[14](https://github.com/ranaroussi/yfinance/discussions/1755)
[15](https://www.freepublicapis.com/yahoo-finance)
[16](https://www.tiingo.com/blog/yahoo-finance-api/)
[17](https://help.portfolio-performance.info/en/how-to/downloading-historical-prices/json/)
[18](https://ui-patterns.com/patterns/Autocomplete/examples/1847)
[19](https://scrapfly.io/blog/posts/guide-to-yahoo-finance-api)
[20](https://github.com/gadicc/node-yahoo-finance2/issues/8)
[21](https://apidojo.net/documentations/yahoo)
[22](https://www.reddit.com/r/GoogleAppsScript/comments/1ad1b2y/pulling_info_from_yahoo_finance_not_totally/)
[23](https://blog.coupler.io/googlefinance-function-advanced-tutorial/)
[24](https://www.youtube.com/watch?v=yVB71LL0LnE)
[25](https://github.com/finance-quote/finance-quote/issues/369)
[26](https://stackoverflow.com/questions/59799567/how-does-yahoo-finance-calculate-adjusted-close-stock-prices)
[27](https://algotrading101.com/learn/yahoo-finance-api-guide/)
[28](https://www.reddit.com/r/sheets/comments/12snqft/broken_yahoo_finance_api_url/)
[29](http://help.yahoo.com/kb/SLN28256.html)
[30](https://stackoverflow.com/questions/68318610/dividend-history-from-yahoo-finance)

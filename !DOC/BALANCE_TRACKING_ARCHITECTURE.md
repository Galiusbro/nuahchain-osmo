# Архитектура отслеживания балансов пользователей

## Текущая ситуация

### Что есть сейчас:
1. **Мониторинг транзакций** через WebSocket (`BlockchainMonitor`)
   - Подписка на все транзакции: `tm.event='Tx'`
   - Сохранение в `blockchain_transactions` таблицу
   - События содержат информацию о `transfer`, `coin_received`, `coin_spent`

2. **Получение балансов из блокчейна**
   - Метод `GetAllBalances()` в `blockchain.Client`
   - Использует gRPC `bank.AllBalances`
   - Текущий эндпоинт `/api/users/balances-db` **на самом деле берет из блокчейна** (не из БД!)

3. **НЕТ таблицы для балансов в БД**
   - Балансы не кешируются
   - Каждый запрос идет в блокчейн
   - Нет истории изменений балансов

## Проблемы текущего подхода

1. **Производительность**: Каждый запрос балансов = запрос в блокчейн (медленно)
2. **Нагрузка на ноду**: Много запросов к gRPC
3. **Нет истории**: Нельзя посмотреть историю изменений балансов
4. **Нет real-time**: Нет WebSocket для обновлений балансов
5. **Нет кеширования**: Нет быстрого доступа к актуальным балансам

## Предлагаемая архитектура

### 1. Таблицы для балансов в БД

#### 1.1. Текущие балансы (`user_balances`)

```sql
CREATE TABLE user_balances (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address VARCHAR(255) NOT NULL,  -- Cosmos address
    denom VARCHAR(255) NOT NULL,    -- Token denom
    amount NUMERIC NOT NULL,        -- Balance amount (as string for precision)
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by_tx_hash VARCHAR(64),  -- Last transaction that changed this balance
    updated_by_height BIGINT,       -- Block height of last update

    UNIQUE(user_id, denom),
    UNIQUE(address, denom)
);

CREATE INDEX idx_user_balances_user_id ON user_balances(user_id);
CREATE INDEX idx_user_balances_address ON user_balances(address);
CREATE INDEX idx_user_balances_denom ON user_balances(denom);
CREATE INDEX idx_user_balances_updated ON user_balances(updated_at);
CREATE INDEX idx_user_balances_height ON user_balances(updated_by_height);
```

#### 1.2. История изменений (`balance_history`)

```sql
CREATE TABLE balance_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address VARCHAR(255) NOT NULL,
    denom VARCHAR(255) NOT NULL,
    amount_before NUMERIC,          -- Balance before change (NULL for first entry)
    amount_after NUMERIC NOT NULL,  -- Balance after change
    amount_delta NUMERIC NOT NULL,  -- Change amount (positive = increase, negative = decrease)
    tx_hash VARCHAR(64) NOT NULL,   -- Transaction that caused the change
    height BIGINT NOT NULL,         -- Block height
    event_type VARCHAR(50),         -- 'transfer', 'coin_received', 'coin_spent', 'sync'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_balance_history_user_id ON balance_history(user_id);
CREATE INDEX idx_balance_history_address ON balance_history(address);
CREATE INDEX idx_balance_history_denom ON balance_history(denom);
CREATE INDEX idx_balance_history_tx_hash ON balance_history(tx_hash);
CREATE INDEX idx_balance_history_height ON balance_history(height);
CREATE INDEX idx_balance_history_created ON balance_history(created_at);
CREATE INDEX idx_balance_history_user_denom ON balance_history(user_id, denom, created_at DESC);
```

**Примечание**: История хранится для аудита и аналитики. Можно настроить автоматическую очистку старых записей (например, старше 1 года) для оптимизации.

### 2. Индексатор балансов (Balance Indexer)

**Принцип работы:**
- Слушает события транзакций из `BlockchainMonitor`
- Парсит события `transfer`, `coin_received`, `coin_spent`
- Обновляет балансы в БД при каждом изменении
- Работает асинхронно, не блокирует обработку транзакций

**События Cosmos SDK для отслеживания:**

Cosmos SDK генерирует следующие события при переводах:

1. **`transfer`** - основное событие перевода:
   ```
   transfer.amount          # "1000unuah" (может быть несколько через запятую)
   transfer.recipient       # "cosmos1..." (получатель)
   transfer.sender          # "cosmos1..." (отправитель)
   ```

2. **`coin_received`** - монеты получены:
   ```
   coin_received.amount     # "1000unuah"
   coin_received.receiver   # "cosmos1..."
   ```

3. **`coin_spent`** - монеты потрачены:
   ```
   coin_spent.amount        # "1000unuah"
   coin_spent.spender       # "cosmos1..."
   ```

**Важно**: События могут содержать несколько деномов в одном `amount` (например, `"1000unuah,5000factory/.../token"`). Нужно парсить каждый деном отдельно.

**Алгоритм индексации:**

1. Получаем событие транзакции из `BlockchainMonitor`
2. Проверяем, что транзакция успешна (`Success == true`)
3. Извлекаем все события типа `transfer`, `coin_received`, `coin_spent`
4. Для каждого события:
   - Парсим `amount` (может быть несколько деномов через запятую)
   - Для каждого денома:
     - Определяем адрес (sender/receiver/spender)
     - Определяем изменение баланса (положительное для получателя, отрицательное для отправителя)
     - Проверяем, есть ли адрес в таблице `wallets`
     - Если есть:
       - Получаем текущий баланс из `user_balances` (или 0, если нет)
       - Вычисляем новый баланс
       - Обновляем `user_balances` (UPSERT)
       - Сохраняем запись в `balance_history`
5. Отправляем событие в WebSocket канал для real-time обновлений

**Обработка ошибок:**
- Retry с exponential backoff (3 попытки: 1s, 2s, 4s)
- Dead letter queue для неудачных обработок
- Логирование всех ошибок
- Алерты при критических ошибках (например, >10 ошибок подряд)

### 3. Эндпоинты

#### 3.1. `/api/users/balances-db` (из БД - быстрый)
- **Источник**: Таблица `user_balances`
- **Скорость**: Быстро (запрос к БД)
- **Актуальность**: Может быть немного устаревшим (зависит от скорости индексатора)
- **Использование**: Для UI, списков, быстрого отображения

#### 3.2. `/api/users/balances` (из блокчейна - свежий)
- **Источник**: gRPC `bank.AllBalances`
- **Скорость**: Медленнее (запрос в блокчейн)
- **Актуальность**: Всегда актуальный (прямо из блокчейна)
- **Использование**: Для критичных операций, проверки перед транзакциями

#### 3.3. `/api/users/balances/ws` (WebSocket - real-time)
- **Протокол**: WebSocket
- **Функционал**:
  - Подписка на обновления балансов пользователя
  - Отправка обновлений при изменении балансов
  - Поддержка фильтрации по denom
- **Использование**: Real-time UI обновления

### 4. WebSocket для балансов

**Архитектура:**
```
Client <--WS--> Server <--Events--> Balance Indexer
```

**Протокол:**
```json
// Подписка
{
  "action": "subscribe",
  "user_id": 123,
  "denoms": ["unuah", "factory/.../token"] // опционально
}

// Отписка
{
  "action": "unsubscribe"
}

// Обновление баланса (от сервера)
{
  "type": "balance_update",
  "user_id": 123,
  "denom": "unuah",
  "amount": "1000000",
  "tx_hash": "abc123...",
  "timestamp": "2025-11-15T10:00:00Z"
}
```

**Реализация:**
- Использовать существующий WebSocket инфраструктуру
- Balance Indexer отправляет события в канал
- WebSocket handler подписывается на канал и отправляет клиентам

### 5. Синхронизация и восстановление

**Проблема**: Что делать, если индексатор пропустил транзакции или был сбой?

**Решение (Best Practices):**

#### 5.1. Периодическая синхронизация активных пользователей

**Алгоритм:**
- Фоновый процесс каждые **5 минут**
- Проверяет только **активных пользователей** (заходили за последние 7 дней)
- Для каждого активного пользователя:
  - Получает балансы из блокчейна
  - Сравнивает с БД
  - Если есть расхождения:
    - Обновляет `user_balances`
    - Создает запись в `balance_history` с `event_type='sync'`
    - Логирует расхождение для анализа

**Оптимизация:**
- Использует batch запросы к блокчейну
- Ограничение: максимум 100 пользователей за раз
- Пропускает пользователей, которые недавно синхронизировались

#### 5.2. Полная синхронизация (реже)

- Раз в **1 час** проверяет всех пользователей (включая неактивных)
- Используется для восстановления после длительных сбоев
- Можно запускать в ночное время для снижения нагрузки

#### 5.3. Ручная синхронизация

- Эндпоинт `POST /api/users/balances/sync` для принудительной синхронизации
- Может синхронизировать:
  - Текущего пользователя (если авторизован)
  - Конкретного пользователя (только для админов)
  - Всех пользователей (только для админов)

#### 5.4. Backfill (одноразово)

- Скрипт для заполнения исторических балансов
- Парсит `blockchain_transactions` в обратном порядке (от новых к старым)
- Восстанавливает балансы для всех пользователей
- Полезно при первом запуске системы или после миграции

#### 5.5. Восстановление после сбоя

**Механизм:**
1. Индексатор сохраняет последний обработанный `height` в БД
2. При запуске проверяет, нет ли пропущенных блоков
3. Если есть пропуски:
   - Запрашивает транзакции из пропущенных блоков
   - Обрабатывает их в порядке возрастания `height`
   - Обновляет балансы

**Таблица для отслеживания:**
```sql
CREATE TABLE balance_indexer_state (
    id INT PRIMARY KEY DEFAULT 1,
    last_processed_height BIGINT NOT NULL,
    last_processed_at TIMESTAMPTZ NOT NULL,
    last_error TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT single_row CHECK (id = 1)
);
```

## План реализации (Production-Ready)

### Этап 1: Базовая инфраструктура (MVP)
1. ✅ Создать таблицы `user_balances` и `balance_history`
2. ✅ Создать таблицу `balance_indexer_state` для отслеживания состояния
3. ✅ Создать таблицу `failed_balance_updates` для DLQ
4. ✅ Реализовать Balance Indexer с парсингом событий
5. ✅ Обновить эндпоинт `/api/users/balances-db` (читать из БД)
6. ✅ Создать эндпоинт `/api/users/balances` (читать из блокчейна)

### Этап 2: Синхронизация и восстановление
7. ✅ Реализовать периодическую синхронизацию активных пользователей
8. ✅ Реализовать полную синхронизацию (раз в час)
9. ✅ Реализовать восстановление после сбоя (проверка пропущенных блоков)
10. ✅ Создать эндпоинт `/api/users/balances/sync`

### Этап 3: Real-time обновления
11. ✅ Реализовать WebSocket для балансов
12. ✅ Интеграция Balance Indexer с WebSocket каналом
13. ✅ Поддержка подписок/отписок

### Этап 4: Обработка ошибок и мониторинг
14. ✅ Реализовать retry с exponential backoff
15. ✅ Реализовать Dead Letter Queue
16. ✅ Добавить логирование и метрики
17. ✅ Настроить алерты

### Этап 5: Оптимизация и тестирование
18. ✅ Оптимизация запросов (batch обновления)
19. ✅ Тестирование под нагрузкой
20. ✅ Backfill скрипт для исторических данных
21. ✅ Документация и мониторинг

## Рекомендации по реализации

### Приоритеты:
1. **Высокий**: Таблицы БД, Balance Indexer, два эндпоинта
2. **Средний**: Синхронизация, восстановление после сбоя
3. **Средний**: WebSocket для real-time
4. **Низкий**: Оптимизация, метрики, алерты

### Подход:
- Реализовывать поэтапно, тестируя каждый этап
- Начать с MVP (Этап 1), затем добавить остальное
- Использовать существующую инфраструктуру (BlockchainMonitor, WebSocket)
- Фокус на правильности данных и надежности

## Решения (Best Practices)

### 1. Частота обновления балансов

**Решение: В реальном времени при каждой транзакции**

**Обоснование:**
- Это стандарт для production систем (как делают Coinbase, Binance, и другие)
- Пользователи ожидают актуальные данные
- WebSocket события приходят в реальном времени, нет смысла задерживать
- Нагрузка на БД минимальна (UPSERT операции быстрые)
- Можно использовать batch обновления для оптимизации (обновлять несколько балансов за один запрос)

**Реализация:**
- Индексатор обрабатывает каждое событие транзакции немедленно
- Используется транзакция БД для атомарности (обновление `user_balances` + запись в `balance_history`)
- Batch обновления для нескольких балансов в одной транзакции

### 2. История изменений балансов

**Решение: Да, нужна таблица `balance_history`**

**Обоснование:**
- Аудит и прозрачность (важно для финансовых систем)
- Аналитика и отчетность
- Восстановление данных при ошибках
- Отслеживание изменений для compliance

**Оптимизация:**
- Хранить только последние N записей (например, 10,000) на пользователя
- Старые записи можно архивировать или удалять
- Можно использовать партиционирование по дате для производительности

### 3. Обработка ошибок индексатора

**Решение: Retry с exponential backoff + Dead Letter Queue + Алерты**

**Механизм:**
1. **Retry с exponential backoff:**
   - 1-я попытка: сразу
   - 2-я попытка: через 1 секунду
   - 3-я попытка: через 2 секунды
   - 4-я попытка: через 4 секунды
   - Максимум 4 попытки

2. **Dead Letter Queue (DLQ):**
   - Если все попытки неудачны, событие сохраняется в `failed_balance_updates`
   - Администратор может просмотреть и обработать вручную
   - Можно настроить автоматический retry для DLQ раз в час

3. **Алерты:**
   - Если >10 ошибок подряд - критический алерт
   - Если >50 ошибок за час - предупреждение
   - Если индексатор не обрабатывает события >5 минут - критический алерт

**Таблица для DLQ:**
```sql
CREATE TABLE failed_balance_updates (
    id BIGSERIAL PRIMARY KEY,
    tx_hash VARCHAR(64) NOT NULL,
    height BIGINT NOT NULL,
    address VARCHAR(255) NOT NULL,
    denom VARCHAR(255) NOT NULL,
    amount_delta NUMERIC NOT NULL,
    error_message TEXT NOT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_retry_at TIMESTAMPTZ
);
```

### 4. Синхронизация пользователей

**Решение: Активные пользователи часто, все пользователи реже**

**Стратегия:**
- **Активные пользователи** (заходили за последние 7 дней):
  - Синхронизация каждые 5 минут
  - Приоритетная обработка событий
- **Неактивные пользователи**:
  - Синхронизация раз в 1 час
  - Можно пропускать, если нет изменений

**Оптимизация:**
- Кеширование списка активных пользователей
- Batch обработка для снижения нагрузки
- Пропуск пользователей без изменений (проверка по `updated_at`)

### 5. Обработка большого количества деномов

**Решение: Без ограничений, но с оптимизацией**

**Подход:**
- Нет жесткого ограничения на количество деномов
- Пагинация в эндпоинтах (по умолчанию 100, максимум 1000)
- Индексы для быстрого поиска
- Кеширование популярных деномов

**Эндпоинты:**
- `/api/users/balances-db?limit=100&offset=0` - пагинация
- `/api/users/balances-db?denoms=unuah,factory/.../token` - фильтрация по конкретным деномам

## Следующие шаги

1. ✅ Обсудить архитектуру
2. ⏳ Создать таблицу `user_balances`
3. ⏳ Реализовать Balance Indexer
4. ⏳ Обновить эндпоинт `/api/users/balances-db` (читать из БД)
5. ⏳ Создать эндпоинт `/api/users/balances` (читать из блокчейна)
6. ⏳ Добавить WebSocket для real-time обновлений
7. ⏳ Реализовать синхронизацию
8. ⏳ Тестирование и оптимизация


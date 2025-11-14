# Анализ данных транзакций из блокчейна

## 📊 Что мы получаем через WebSocket (текущая реализация)

### Структура `TxEvent` (из WebSocket событий)

```go
type TxEvent struct {
    TxHash  string    // Хеш транзакции
    Height  int64     // Высота блока
    Code    uint32    // Код результата (0 = успех)
    Log     string    // Лог выполнения
    Events  []Event   // События от модулей
    Success bool      // Успешность (Code == 0)
}
```

### Что приходит в WebSocket событии:

**Из `TxResult`:**
- `height` - высота блока (string или int64)
- `tx` - base64-encoded транзакция (полные байты транзакции, но мы их не парсим)
- `result.code` - код результата
- `result.log` - лог выполнения
- `result.events` - массив событий с атрибутами

**Ограничения WebSocket событий:**
- ❌ Нет информации о gas (gas_wanted, gas_used)
- ❌ Нет информации о fee (комиссии)
- ❌ Нет информации о signers (подписантах)
- ❌ Нет информации о messages (типах сообщений)
- ❌ Нет timestamp
- ❌ Нет codespace
- ❌ Нет info
- ❌ Нет msg_responses (ответы от модулей)

**Важно:** В WebSocket событии есть поле `tx` (base64-encoded транзакция), но мы его **не парсим**. Это полные байты транзакции, из которых можно извлечь все данные, но это требует дополнительной обработки.

---

## 🔍 Что можно получить через GetTx (полная информация)

### Структура `TxResponse` (из Cosmos SDK)

```go
type TxResponse struct {
    Height      int64          // Высота блока
    TxHash      string         // Хеш транзакции
    Codespace   string         // Пространство кодов ошибок
    Code        uint32         // Код результата
    Data        string         // Данные результата (base64)
    RawLog      string         // Сырой лог
    Logs        []ABCIMessageLog // Типизированные логи
    Info        string         // Дополнительная информация
    GasWanted   int64          // Запрошенный gas
    GasUsed     int64          // Использованный gas
    Tx          *Any           // Полная транзакция (сообщения, подписи, fee)
    Timestamp   string         // Время блока
    Events      []Event        // Все события
    // ... и другие поля
}
```

### Полная транзакция (`Tx`) содержит:

**1. Messages (сообщения):**
- Тип сообщения (MsgBuyAsset, MsgSellAsset и т.д.)
- Параметры сообщения (amount, denom, sender и т.д.)
- Все сообщения в транзакции (может быть несколько)

**2. Signers (подписанты):**
- Адреса всех подписантов
- Публичные ключи
- Последовательность (sequence) для каждого аккаунта

**3. Fee (комиссия):**
- Amount (сумма комиссии)
- Denom (деноминация)
- Gas limit

**4. Memo:**
- Текстовая заметка к транзакции

**5. Timeout Height:**
- Высота блока, после которой транзакция недействительна

---

## 📋 Сравнительная таблица

| Данные | WebSocket | GetTx | Важность для мониторинга |
|--------|-----------|-------|--------------------------|
| **TxHash** | ✅ | ✅ | 🔴 Критично |
| **Height** | ✅ | ✅ | 🔴 Критично |
| **Code** | ✅ | ✅ | 🔴 Критично |
| **Success** | ✅ (вычисляется) | ✅ | 🔴 Критично |
| **Log** | ✅ | ✅ | 🟡 Важно |
| **Events** | ✅ | ✅ | 🟡 Важно |
| **GasWanted** | ❌ | ✅ | 🟡 Важно (аналитика) |
| **GasUsed** | ❌ | ✅ | 🟡 Важно (аналитика) |
| **Codespace** | ❌ | ✅ | 🟢 Полезно |
| **Info** | ❌ | ✅ | 🟢 Полезно |
| **Timestamp** | ❌ | ✅ | 🟡 Важно |
| **Messages** | ❌ (есть в tx, но не парсим) | ✅ | 🔴 Критично (тип операции) |
| **Signers** | ❌ (есть в tx, но не парсим) | ✅ | 🔴 Критично (отправитель) |
| **Fee** | ❌ (есть в tx, но не парсим) | ✅ | 🟡 Важно (комиссии) |
| **Memo** | ❌ (есть в tx, но не парсим) | ✅ | 🟢 Опционально |
| **MsgResponses** | ❌ | ✅ | 🟡 Важно (результаты модулей) |

---

## 🎯 Что можно извлечь из Events (без GetTx)

События содержат много полезной информации, которую мы **уже получаем**:

**Пример события:**
```json
{
  "type": "message",
  "attributes": [
    {"key": "action", "value": "/osmosis.assets.v1.MsgBuyAsset"},
    {"key": "sender", "value": "cosmos1abc..."},
    {"key": "module", "value": "assets"}
  ]
}
```

**Можно извлечь из Events:**
- ✅ Тип операции (из `action` или `message.action`)
- ✅ Отправитель (из `sender` или `message.sender`)
- ✅ Модуль (из `module` или специфичных событий)
- ✅ Параметры операции (из событий модуля, например `transfer.amount`, `transfer.recipient`)

**Примеры событий для разных операций:**

**BuyAsset:**
```json
{
  "type": "message",
  "attributes": [
    {"key": "action", "value": "/osmosis.assets.v1.MsgBuyAsset"},
    {"key": "sender", "value": "cosmos1..."}
  ]
},
{
  "type": "transfer",
  "attributes": [
    {"key": "amount", "value": "1000000factory/.../ndollar"},
    {"key": "recipient", "value": "cosmos1..."}
  ]
}
```

**Margin Open:**
```json
{
  "type": "message",
  "attributes": [
    {"key": "action", "value": "/osmosis.leverage.v1.MsgOpenPosition"},
    {"key": "sender", "value": "cosmos1..."}
  ]
},
{
  "type": "leverage.position_opened",
  "attributes": [
    {"key": "position_id", "value": "123"},
    {"key": "base_denom", "value": "GOLD"}
  ]
}
```

---

## 💡 Рекомендации по сохранению данных

### Вариант 1: Минимальный набор (только WebSocket, без парсинга tx)

```sql
CREATE TABLE blockchain_transactions_minimal (
    tx_hash VARCHAR(64) PRIMARY KEY,
    height BIGINT NOT NULL,
    code INT NOT NULL,
    success BOOLEAN NOT NULL,
    log TEXT,
    events JSONB,  -- Все события с атрибутами
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Плюсы:**
- ✅ Быстро (нет дополнительных запросов)
- ✅ Real-time
- ✅ Минимальная нагрузка
- ✅ Можно извлечь тип операции и отправителя из events

**Минусы:**
- ❌ Нет информации о gas/fee
- ❌ Нет timestamp блока
- ❌ Нет codespace

---

### Вариант 2: Обогащенный набор (WebSocket + парсинг tx из события)

**Можно парсить `tx` из WebSocket события:**
- В событии есть поле `tx` (base64-encoded транзакция)
- Можно декодировать и извлечь messages, signers, fee
- Не требует дополнительного запроса к блокчейну

```sql
CREATE TABLE blockchain_transactions_enriched (
    tx_hash VARCHAR(64) PRIMARY KEY,
    height BIGINT NOT NULL,
    code INT NOT NULL,
    codespace VARCHAR(50),
    success BOOLEAN NOT NULL,
    log TEXT,
    gas_wanted BIGINT,
    gas_used BIGINT,
    fee_amount JSONB,  -- {denom: "unuah", amount: "1000"}
    messages JSONB,   -- [{type: "MsgBuyAsset", value: {...}}]
    signers TEXT[],   -- ["cosmos1...", "cosmos1..."]
    memo TEXT,
    events JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Плюсы:**
- ✅ Все данные без дополнительных запросов
- ✅ Real-time
- ✅ Полная информация

**Минусы:**
- ⚠️ Требует парсинга protobuf транзакции
- ⚠️ Нет timestamp (его нужно получать отдельно)

---

### Вариант 3: Гибридный подход (рекомендуется)

**1. Сохранять из WebSocket сразу:**
- tx_hash
- height
- code
- success
- log
- events (полные, с атрибутами)
- created_at (время получения события)

**2. Извлекать из Events:**
- Тип операции (из `message.action`)
- Отправитель (из `message.sender`)
- Параметры операции (из специфичных событий)

**3. Обогащать через GetTx асинхронно (опционально):**
- Для транзакций с ошибками (code != 0) - для анализа
- Для транзакций определенных типов (по events)
- По запросу через API (lazy loading)
- Получать: gas_wanted, gas_used, timestamp, codespace, msg_responses

**4. Индексировать:**
- По height (для быстрого поиска по блоку)
- По events (для фильтрации по типу операции)
- По sender (извлеченный из events)
- По created_at (для временных запросов)

---

## 🔧 Пример структуры для сохранения

```sql
CREATE TABLE blockchain_transactions (
    tx_hash VARCHAR(64) PRIMARY KEY,
    height BIGINT NOT NULL,
    code INT NOT NULL,
    success BOOLEAN NOT NULL,
    log TEXT,
    events JSONB NOT NULL,  -- Все события с атрибутами

    -- Извлеченные из events (для быстрого поиска)
    operation_type VARCHAR(100),  -- Из message.action
    sender TEXT,                   -- Из message.sender
    module_name VARCHAR(50),       -- Из events

    -- Обогащенные данные (заполняются опционально через GetTx)
    codespace VARCHAR(50),
    gas_wanted BIGINT,
    gas_used BIGINT,
    fee_amount JSONB,
    timestamp TIMESTAMPTZ,
    messages JSONB,
    signers TEXT[],
    memo TEXT,
    msg_responses JSONB,

    -- Метаданные
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    enriched_at TIMESTAMPTZ,  -- Когда обогатили через GetTx
    enriched BOOLEAN DEFAULT FALSE
);

-- Индексы
CREATE INDEX idx_blockchain_tx_height ON blockchain_transactions(height);
CREATE INDEX idx_blockchain_tx_sender ON blockchain_transactions(sender);
CREATE INDEX idx_blockchain_tx_operation ON blockchain_transactions(operation_type);
CREATE INDEX idx_blockchain_tx_created ON blockchain_transactions(created_at);
CREATE INDEX idx_blockchain_tx_enriched ON blockchain_transactions(enriched) WHERE enriched = FALSE;
```

---

## 📊 Итоговая рекомендация

**Для админ-панели мониторинга:**

1. **Сохранять из WebSocket сразу:**
   - Все базовые поля (hash, height, code, success, log, events)
   - Извлекать из events: operation_type, sender, module_name
   - Это дает полную картину всех транзакций в реальном времени

2. **Обогащать через GetTx (опционально, асинхронно):**
   - Для транзакций с ошибками (для детального анализа)
   - По запросу пользователя (lazy loading)
   - Получать: gas, fee, timestamp, codespace, msg_responses

**Это даст:**
- ✅ Полный мониторинг всех транзакций в реальном времени
- ✅ Минимальную нагрузку на блокчейн (только WebSocket)
- ✅ Возможность детального анализа при необходимости
- ✅ Быстрый доступ к истории с фильтрацией
- ✅ Информацию об отправителе и типе операции без дополнительных запросов

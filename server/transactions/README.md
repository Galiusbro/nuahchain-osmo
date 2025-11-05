# Transactions Module

Модуль для записи всех операций пользователей в базу данных.

## Назначение

Этот модуль записывает **все операции** пользователей с токенами и активами в БД для:
- 📊 Истории операций пользователя
- 📈 Аналитики и статистики
- 🔍 Быстрого доступа к транзакциям (без запросов к блокчейну)
- 📝 Аудита операций

## Структура

```
server/transactions/
├── models.go      # Модели данных и типы операций
├── repository.go  # Работа с базой данных
└── README.md      # Эта документация
```

## Типы операций

### Операции с токенами
- `TOKEN_CREATE` - создание токена
- `TOKEN_BUY` - покупка токена
- `TOKEN_SELL` - продажа токена

### Операции с активами
- `ASSET_ENSURE` - создание/обеспечение актива
- `ASSET_BUY` - покупка актива
- `ASSET_SELL` - продажа актива

## Статусы транзакций

- `PENDING` - транзакция отправлена, ожидает подтверждения в блокчейне
- `SUCCESS` - транзакция успешно выполнена
- `FAILED` - транзакция не удалась

## Структура таблицы

```sql
CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,           -- ID пользователя
    operation_type VARCHAR(50) NOT NULL, -- Тип операции
    tx_hash VARCHAR(255) NOT NULL,      -- Хеш транзакции в блокчейне
    status VARCHAR(20) NOT NULL,        -- Статус (PENDING, SUCCESS, FAILED)
    operation_data JSONB,               -- JSON данные операции
    error_message TEXT,                 -- Сообщение об ошибке (если есть)
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

## Использование

### Создание записи о транзакции

```go
transactionsRepo.CreateTransaction(transactions.CreateTransactionRequest{
    UserID:        userID,
    OperationType: transactions.OperationTypeTokenCreate,
    TxHash:        txHash,
    Status:        transactions.StatusPending,
    OperationData: transactions.TokenCreateData(denom, name, symbol, image, description),
})
```

### Получение транзакций пользователя

```go
// Все транзакции пользователя
txs, err := transactionsRepo.GetUserTransactions(userID, 10, 0)

// Транзакции определенного типа
txs, err := transactionsRepo.GetUserTransactionsByType(userID, transactions.OperationTypeTokenBuy, 10, 0)
```

### Обновление статуса транзакции

```go
transactionsRepo.UpdateTransactionByTxHash(
    txHash,
    transactions.StatusSuccess,
    updatedData,
    nil,
)
```

## Данные операций (operation_data)

Данные хранятся в формате JSON для гибкости. Каждый тип операции имеет свои поля:

### TOKEN_CREATE
```json
{
  "denom": "factory/...",
  "name": "My Token",
  "symbol": "MTK",
  "image": "...",
  "description": "..."
}
```

### TOKEN_BUY
```json
{
  "denom": "factory/...",
  "payment_denom": "unuah",
  "payment_amount": "1000000",
  "tokens_out": "500",
  "price_paid": "2000"
}
```

### TOKEN_SELL
```json
{
  "denom": "factory/...",
  "token_amount": "500",
  "payment_denom": "unuah",
  "payment_out": "950000",
  "price_received": "1900"
}
```

### ASSET_ENSURE
```json
{
  "symbol": "GOLD"
}
```

### ASSET_BUY
```json
{
  "symbol": "GOLD",
  "denom": "unuah",
  "amount": "1000000",
  "base_amount": "0.5"
}
```

### ASSET_SELL
```json
{
  "symbol": "GOLD",
  "base_amount": "0.5",
  "payout_ndollar": "1000000"
}
```

## Индексы

Таблица имеет индексы для быстрого поиска:
- По `user_id` - все транзакции пользователя
- По `operation_type` - транзакции определенного типа
- По `tx_hash` - поиск по хешу транзакции
- По `status` - фильтрация по статусу
- По `created_at` - сортировка по времени
- Комбинированные индексы для частых запросов

## Автоматическая запись

Все операции автоматически записываются в БД при вызове методов:
- `assets.EnsureAsset()` → запись с типом `ASSET_ENSURE`
- `assets.BuyAsset()` → запись с типом `ASSET_BUY`
- `assets.SellAsset()` → запись с типом `ASSET_SELL`
- `usertokens.CreateToken()` → запись с типом `TOKEN_CREATE`
- `usertokens.BuyToken()` → запись с типом `TOKEN_BUY`
- `usertokens.SellToken()` → запись с типом `TOKEN_SELL`

Если запись в БД не удалась, операция **не прерывается** - это гарантирует, что проблемы с БД не блокируют блокчейн операции.

## Примеры запросов

### Получить последние 10 транзакций пользователя
```sql
SELECT * FROM transactions
WHERE user_id = 1
ORDER BY created_at DESC
LIMIT 10;
```

### Получить все покупки токенов пользователя
```sql
SELECT * FROM transactions
WHERE user_id = 1
  AND operation_type = 'TOKEN_BUY'
ORDER BY created_at DESC;
```

### Найти транзакцию по хешу
```sql
SELECT * FROM transactions
WHERE tx_hash = 'abc123...';
```

### Получить статистику операций пользователя
```sql
SELECT
    operation_type,
    status,
    COUNT(*) as count
FROM transactions
WHERE user_id = 1
GROUP BY operation_type, status;
```


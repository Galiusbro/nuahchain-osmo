# User Tokens Module

Модуль для работы с пользовательскими токенами на блокчейне.

## Структура

```
usertokens/
├── models.go      # Модели запросов и ответов (DTO)
├── service.go     # Основной сервис с общей логикой
├── handlers.go    # HTTP обработчики запросов
├── create.go      # Операция создания токена
├── buy.go         # Операция покупки токена из bonding curve
└── sell.go        # Операция продажи токена в bonding curve
```

## Архитектура

### Models (`models.go`)
Содержит структуры данных для запросов и ответов API:
- `CreateTokenRequest` - запрос на создание токена
- `CreateTokenResponse` - ответ при создании токена (включает `Status`, `TxHash`, `Message`)
- `BuyTokenRequest` - запрос на покупку токена из bonding curve
- `BuyTokenResponse` - ответ при покупке токена (возвращает `Status`, `TxHash`, `TokensOut`, `Message`)
- `SellTokenRequest` - запрос на продажу токена в bonding curve
- `SellTokenResponse` - ответ при продаже токена (`Status`, `TxHash`, `PaymentOut`, `Message`)
- `TokenInfo` - информация о токене

### Service (`service.go`)
Основной сервис, предоставляющий:
- `NewService()` - создание экземпляра сервиса (с репозиторием транзакций и трекером)
- `GetUserWallet()` - получение и расшифровка кошелька пользователя
- Методы `CreateToken/BuyToken/SellToken`, которые:
  - Шифруют/расшифровывают приватный ключ
  - Отправляют транзакцию через blockchain client
  - Сохраняют запись в таблице `transactions` со статусом `PENDING`
  - Передают `tx_hash` в фонового трекера для последующего обновления

### Handlers (`handlers.go`)
HTTP обработчики для маршрутов:
- `HandleCreateToken` - POST `/api/tokens/create`
- `HandleBuyToken` - POST `/api/tokens/buy`
- `HandleSellToken` - POST `/api/tokens/sell`

### Operations (отдельные файлы)
Каждая операция с токенами вынесена в отдельный файл:
- `create.go` - создание токена (`CreateToken()`)
- `buy.go` - покупка токена из bonding curve (`BuyToken()`)
- `sell.go` - продажа токена в bonding curve (`SellToken()`)

## Добавление новой операции

Для добавления новой операции с токенами:

1. **Добавьте модели** в `models.go`:
```go
type NewOperationRequest struct {
    Field1 string `json:"field1"`
    Field2 string `json:"field2"`
}

type NewOperationResponse struct {
    Status string `json:"status"`
    TxHash string `json:"tx_hash"`
    Message string `json:"message,omitempty"`
}
```

2. **Создайте файл операции** (например, `new_operation.go`):
```go
package usertokens

type NewOperationResponse struct {
    Status string `json:"status"`
    TxHash string `json:"tx_hash"`
    Message string `json:"message,omitempty"`
}

func (s *Service) NewOperation(ctx context.Context, userID int64, req NewOperationRequest) (*NewOperationResponse, error) {
    // Получаем кошелек пользователя
    wallet, privKeyBytes, err := s.GetUserWallet(ctx, userID)
    if err != nil {
        return nil, err
    }

    _ = wallet // используем кошелек при формировании сообщения

    // Выполняем операцию через blockchain client
    txHash, err := s.blockchainCli.SomeOperation(ctx, fmt.Sprintf("%x", privKeyBytes), req)
    if err != nil {
        return &NewOperationResponse{Status: transactions.StatusFailed, TxHash: txHash, Message: err.Error()}, err
    }

    // Записываем транзакцию и ставим в очередь трекера
    if err := s.recordPendingTransaction(userID, "CUSTOM_OPERATION", txHash, map[string]interface{}{"payload": req}); err != nil {
        return &NewOperationResponse{Status: transactions.StatusFailed, TxHash: txHash, Message: err.Error()}, err
    }

    return &NewOperationResponse{
        Status:  string(transactions.StatusPending),
        TxHash:  txHash,
        Message: "Operation broadcast, awaiting confirmation",
    }, nil
}
```

3. **Добавьте обработчик** в `handlers.go`:
```go
func HandleNewOperation(w http.ResponseWriter, r *http.Request) {
    // Аутентификация
    user, err := authenticateRequest(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    // Парсинг запроса
    var req NewOperationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Вызов сервиса
    resp, err := tokenService.NewOperation(r.Context(), user.ID, req)
    // ...
}
```

4. **Зарегистрируйте маршрут** в `server/api/router.go`:
```go
mux.HandleFunc("/api/tokens/new-operation", usertokens.HandleNewOperation)
```

## Зависимости

- `server/auth` - для аутентификации и работы с кошельками
- `server/blockchain` - для взаимодействия с блокчейном
- `server/transactions` - для записи и последующего обновления статусов (`PENDING` → `SUCCESS/FAILED`)
- `server/transactions/tracker` - фоновый трекер, который дергает `GetTx` и обновляет строки в БД

Финальные статусы и события можно получать через `GET /api/tx/<tx_hash>` или соответствующие сервисные методы.


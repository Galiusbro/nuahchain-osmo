# План реализации WebSocket для трекера транзакций

## Обзор

Реализация WebSocket-подписок для отслеживания транзакций в реальном времени с fallback на polling при недоступности WebSocket.

**Текущее состояние:**
- Трекер использует polling каждые 3 секунды через `GetTxStatus`
- Задержка обновления: до 3 секунд
- Нагрузка на ноду: постоянные HTTP запросы

**Целевое состояние:**
- WebSocket как основной метод (мгновенные уведомления)
- Polling как fallback (при недоступности WebSocket)
- Автоматическое переключение между методами

---

## Этап 1: Исследование и сбор информации

### 1.1 Изучение CometBFT WebSocket API

**Задачи:**
- [ ] Изучить официальную документацию CometBFT WebSocket API
- [ ] Проверить формат JSON-RPC сообщений для подписок
- [ ] Изучить типы событий: `tm.event`, `tx.hash`, custom events
- [ ] Проверить примеры запросов/ответов

**Ресурсы:**
- CometBFT RPC API: https://docs.cometbft.com/v0.38/rpc/
- WebSocket endpoint: `ws://localhost:26657/websocket`
- JSON-RPC 2.0 протокол

**Формат подписки:**
```json
{
  "jsonrpc": "2.0",
  "method": "subscribe",
  "id": 1,
  "params": {
    "query": "tx.hash='ABC123...'"
  }
}
```

**Ожидаемый ответ:**
```json
{
  "jsonrpc": "2.0",
  "method": "tm_event",
  "params": {
    "result": {
      "query": "tx.hash='ABC123...'",
      "data": {
        "type": "tendermint/event/Tx",
        "value": {
          "TxResult": {
            "height": 12345,
            "tx": "...",
            "result": {
              "code": 0,
              "log": "...",
              "events": [...]
            }
          }
        }
      }
    }
  }
}
```

### 1.2 Выбор Go библиотеки для WebSocket

**Варианты:**
1. **gorilla/websocket** (рекомендуется)
   - Стандарт де-факто для Go
   - Хорошая документация
   - Поддержка ping/pong для keepalive

2. **nhooyr.io/websocket**
   - Современная альтернатива
   - Лучшая производительность
   - Меньше зависимостей

**Решение:** `gorilla/websocket` (более стабильная, больше примеров)

**Установка:**
```bash
go get github.com/gorilla/websocket
```

### 1.3 Проверка текущей архитектуры

**Файлы для изучения:**
- `server/blockchain/client.go` - текущий gRPC/REST клиент
- `server/transactions/tracker/tracker.go` - логика трекера
- `server/config/config.go` - конфигурация

**Текущие endpoints:**
- gRPC: `localhost:9090`
- REST: `localhost:26657`
- WebSocket: `ws://localhost:26657/websocket` (нужно добавить)

---

## Этап 2: Подготовка конфигурации

### 2.1 Расширение config.go

**Добавить в `BlockchainConfig`:**
```go
type BlockchainConfig struct {
    NodeURL string `yaml:"node_url"`        // gRPC (localhost:9090)
    ChainID string `yaml:"chain_id"`

    // Новые поля для WebSocket
    RPCURL          string        `yaml:"rpc_url"`           // REST/WebSocket (localhost:26657)
    WebSocketURL    string        `yaml:"websocket_url"`     // ws://localhost:26657/websocket
    WebSocketEnabled bool         `yaml:"websocket_enabled"`  // Флаг включения
    ReconnectInterval time.Duration `yaml:"reconnect_interval"` // Интервал переподключения
    WebSocketTimeout  time.Duration `yaml:"websocket_timeout"`  // Timeout для операций
}
```

**Значения по умолчанию:**
```go
Blockchain: BlockchainConfig{
    NodeURL:          getEnv("BLOCKCHAIN_NODE_URL", "localhost:9090"),
    ChainID:          getEnv("BLOCKCHAIN_CHAIN_ID", "nuahchain"),
    RPCURL:           getEnv("BLOCKCHAIN_RPC_URL", "localhost:26657"),
    WebSocketURL:     getEnv("BLOCKCHAIN_WEBSOCKET_URL", "ws://localhost:26657/websocket"),
    WebSocketEnabled:  getEnvBool("BLOCKCHAIN_WEBSOCKET_ENABLED", true),
    ReconnectInterval: 5 * time.Second,
    WebSocketTimeout:  30 * time.Second,
}
```

### 2.2 Переменные окружения

**Добавить в `.env`:**
```bash
# WebSocket configuration
BLOCKCHAIN_RPC_URL=localhost:26657
BLOCKCHAIN_WEBSOCKET_URL=ws://localhost:26657/websocket
BLOCKCHAIN_WEBSOCKET_ENABLED=true
BLOCKCHAIN_WEBSOCKET_RECONNECT_INTERVAL=5s
BLOCKCHAIN_WEBSOCKET_TIMEOUT=30s
```

---

## Этап 3: Создание WebSocket клиента

### 3.1 Структура файла `server/blockchain/websocket.go`

**Основные компоненты:**

```go
package blockchain

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

// WebSocketClient управляет WebSocket соединением с CometBFT
type WebSocketClient struct {
    url            string
    conn           *websocket.Conn
    connMu         sync.RWMutex

    subscriptions  map[string]*Subscription  // txHash -> Subscription
    subsMu         sync.RWMutex

    reconnectInterval time.Duration
    timeout          time.Duration

    ctx              context.Context
    cancel           context.CancelFunc
    wg               sync.WaitGroup

    // Каналы для сообщений
    sendCh           chan []byte
    recvCh           chan *JSONRPCResponse

    // Состояние
    connected        bool
    reconnectAttempt int
}

// Subscription представляет подписку на транзакцию
type Subscription struct {
    ID      string
    TxHash  string
    Query   string
    Events  chan *TxEvent
    Done    chan struct{}
}

// JSONRPCRequest/Response для протокола
type JSONRPCRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    Method  string      `json:"method"`
    ID      int         `json:"id"`
    Params  interface{} `json:"params"`
}

type JSONRPCResponse struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      int             `json:"id,omitempty"`
    Method  string          `json:"method,omitempty"`
    Result  json.RawMessage  `json:"result,omitempty"`
    Params  *EventParams    `json:"params,omitempty"`
    Error   *JSONRPCError   `json:"error,omitempty"`
}

// TxEvent представляет событие транзакции
type TxEvent struct {
    TxHash  string
    Height  int64
    Code    uint32
    Log     string
    Events  []Event
    Success bool
}
```

### 3.2 Методы WebSocket клиента

**Основные методы:**

1. **NewWebSocketClient** - создание клиента
2. **Connect** - установка соединения
3. **Subscribe** - подписка на транзакцию
4. **Unsubscribe** - отписка
5. **Reconnect** - переподключение
6. **Close** - закрытие соединения
7. **readLoop** - чтение сообщений (goroutine)
8. **writeLoop** - отправка сообщений (goroutine)
9. **keepAlive** - ping/pong для keepalive (goroutine)

---

## Этап 4: Реализация JSON-RPC протокола

### 4.1 Метод Subscribe

```go
func (ws *WebSocketClient) Subscribe(ctx context.Context, txHash string) (*Subscription, error) {
    query := fmt.Sprintf("tx.hash='%s'", txHash)

    sub := &Subscription{
        ID:     generateSubscriptionID(),
        TxHash: txHash,
        Query:  query,
        Events: make(chan *TxEvent, 1),
        Done:   make(chan struct{}),
    }

    // Сохраняем подписку
    ws.subsMu.Lock()
    ws.subscriptions[txHash] = sub
    ws.subsMu.Unlock()

    // Отправляем JSON-RPC запрос
    req := JSONRPCRequest{
        JSONRPC: "2.0",
        Method:  "subscribe",
        ID:      ws.getNextID(),
        Params: map[string]string{
            "query": query,
        },
    }

    if err := ws.sendRequest(req); err != nil {
        return nil, fmt.Errorf("failed to subscribe: %w", err)
    }

    return sub, nil
}
```

### 4.2 Обработка событий

```go
func (ws *WebSocketClient) handleEvent(msg *JSONRPCResponse) {
    if msg.Method != "tm_event" || msg.Params == nil {
        return
    }

    // Парсим событие транзакции
    txEvent := parseTxEvent(msg.Params)

    // Находим подписку
    ws.subsMu.RLock()
    sub, exists := ws.subscriptions[txEvent.TxHash]
    ws.subsMu.RUnlock()

    if !exists {
        return
    }

    // Отправляем событие в канал подписки
    select {
    case sub.Events <- txEvent:
    case <-sub.Done:
    case <-time.After(5 * time.Second):
        // Timeout - подписка не обрабатывается
    }
}
```

### 4.3 Парсинг событий

```go
func parseTxEvent(params *EventParams) *TxEvent {
    // Извлекаем TxResult из params.result.data.value
    // Парсим code, log, events
    // Возвращаем структурированное событие
}
```

---

## Этап 5: Механизм переподключения

### 5.1 Автоматическое переподключение

```go
func (ws *WebSocketClient) reconnectLoop() {
    defer ws.wg.Done()

    for {
        select {
        case <-ws.ctx.Done():
            return
        default:
        }

        if !ws.isConnected() {
            ws.reconnectAttempt++
            backoff := calculateBackoff(ws.reconnectAttempt)

            time.Sleep(backoff)

            if err := ws.Connect(ws.ctx); err != nil {
                ws.logger.WithError(err).Warn("WebSocket reconnect failed")
                continue
            }

            // Восстанавливаем все подписки
            ws.restoreSubscriptions()
            ws.reconnectAttempt = 0
        }

        time.Sleep(ws.reconnectInterval)
    }
}
```

### 5.2 Exponential Backoff

```go
func calculateBackoff(attempt int) time.Duration {
    base := 1 * time.Second
    max := 60 * time.Second

    backoff := time.Duration(1<<uint(attempt)) * base
    if backoff > max {
        backoff = max
    }

    return backoff
}
```

### 5.3 Восстановление подписок

```go
func (ws *WebSocketClient) restoreSubscriptions() {
    ws.subsMu.RLock()
    defer ws.subsMu.RUnlock()

    for txHash, sub := range ws.subscriptions {
        // Переподписываемся на каждую транзакцию
        req := JSONRPCRequest{
            JSONRPC: "2.0",
            Method:  "subscribe",
            ID:      ws.getNextID(),
            Params: map[string]string{
                "query": sub.Query,
            },
        }
        ws.sendRequest(req)
    }
}
```

---

## Этап 6: Интеграция с трекером

### 6.1 Модификация tracker.go

**Добавить в структуру Tracker:**
```go
type Tracker struct {
    // ... существующие поля

    wsClient        *blockchain.WebSocketClient
    useWebSocket    bool
    fallbackToPoll  bool  // Флаг для переключения на polling
}
```

**Модифицировать handleTx:**
```go
func (t *Tracker) handleTx(txHash string) {
    defer t.markDone(txHash)

    // Пробуем WebSocket если доступен
    if t.useWebSocket && !t.fallbackToPoll {
        if err := t.handleTxWebSocket(txHash); err == nil {
            return  // Успешно обработано через WebSocket
        }
        // Ошибка WebSocket - переключаемся на polling
        t.fallbackToPoll = true
        t.logger.Warn("WebSocket failed, falling back to polling")
    }

    // Fallback на polling
    t.handleTxPolling(txHash)
}
```

### 6.2 Новый метод handleTxWebSocket

```go
func (t *Tracker) handleTxWebSocket(txHash string) error {
    // Подписываемся на транзакцию
    sub, err := t.wsClient.Subscribe(t.ctx, txHash)
    if err != nil {
        return fmt.Errorf("failed to subscribe: %w", err)
    }
    defer t.wsClient.Unsubscribe(txHash)

    // Ждём события с таймаутом
    timeout := time.After(5 * time.Minute) // Максимальное время ожидания

    select {
    case event := <-sub.Events:
        // Обновляем БД
        status := transactions.StatusSuccess
        if !event.Success {
            status = transactions.StatusFailed
        }

        return t.repo.UpdateTransactionByTxHash(
            txHash,
            status,
            nil,
            getErrorMessage(event),
        )

    case <-timeout:
        return fmt.Errorf("transaction timeout")

    case <-t.ctx.Done():
        return t.ctx.Err()
    }
}
```

### 6.3 Переименовать текущий handleTx в handleTxPolling

```go
func (t *Tracker) handleTxPolling(txHash string) {
    // Текущая логика из handleTx
    // ...
}
```

---

## Этап 7: Логика Fallback

### 7.1 Определение доступности WebSocket

```go
func (t *Tracker) isWebSocketAvailable() bool {
    if !t.useWebSocket {
        return false
    }

    if t.wsClient == nil {
        return false
    }

    return t.wsClient.IsConnected()
}
```

### 7.2 Автоматическое переключение

```go
func (t *Tracker) checkWebSocketHealth() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-t.ctx.Done():
            return
        case <-ticker.C:
            if t.fallbackToPoll && t.isWebSocketAvailable() {
                // WebSocket восстановлен - возвращаемся к нему
                t.fallbackToPoll = false
                t.logger.Info("WebSocket restored, switching back from polling")
            }
        }
    }
}
```

### 7.3 Инициализация в main.go

```go
// Создаём WebSocket клиент
wsClient, err := blockchain.NewWebSocketClient(
    cfg.Blockchain.WebSocketURL,
    blockchain.WebSocketConfig{
        ReconnectInterval: cfg.Blockchain.ReconnectInterval,
        Timeout:          cfg.Blockchain.WebSocketTimeout,
    },
)
if err != nil {
    appLogger.WithError(err).Warn("Failed to create WebSocket client, using polling only")
    wsClient = nil
}

// Подключаемся
if wsClient != nil && cfg.Blockchain.WebSocketEnabled {
    if err := wsClient.Connect(context.Background()); err != nil {
        appLogger.WithError(err).Warn("Failed to connect WebSocket, using polling only")
        wsClient = nil
    }
}

// Создаём трекер с WebSocket клиентом
trackerCfg := transactionstracker.Config{
    PollInterval:     3 * time.Second,
    MaxAttempts:      30,
    InitialBatchSize: 0,
    UseWebSocket:     wsClient != nil,
    WebSocketClient:  wsClient,
}
```

---

## Этап 8: Тестирование

### 8.1 Unit тесты

**Файл: `server/blockchain/websocket_test.go`**

```go
func TestWebSocketClient_Subscribe(t *testing.T) {
    // Mock WebSocket соединение
    // Тест подписки
    // Проверка отправки JSON-RPC запроса
}

func TestWebSocketClient_Reconnect(t *testing.T) {
    // Тест переподключения
    // Проверка восстановления подписок
}

func TestWebSocketClient_ParseEvent(t *testing.T) {
    // Тест парсинга событий CometBFT
    // Проверка извлечения данных транзакции
}
```

### 8.2 Интеграционные тесты

**Файл: `server/transactions/tracker/tracker_websocket_test.go`**

```go
func TestTracker_WebSocketIntegration(t *testing.T) {
    // Тест с реальным CometBFT нодой (если доступен)
    // Или mock WebSocket сервер
    // Проверка обработки событий
}

func TestTracker_FallbackToPolling(t *testing.T) {
    // Симулируем обрыв WebSocket
    // Проверяем переключение на polling
    // Проверяем возврат к WebSocket после восстановления
}
```

### 8.3 Тесты производительности

```go
func BenchmarkWebSocket_vs_Polling(b *testing.B) {
    // Сравнение задержки WebSocket vs Polling
    // Измерение нагрузки на ноду
}
```

---

## Этап 9: Мониторинг и логирование

### 9.1 Метрики

**Добавить в Tracker:**
```go
type Metrics struct {
    WebSocketConnected    bool
    WebSocketSubscriptions int
    WebSocketEventsReceived int64
    PollingFallbacks      int64
    WebSocketReconnects   int64
}
```

### 9.2 Логирование

```go
// При подключении
t.logger.WithField("url", ws.url).Info("WebSocket connected")

// При подписке
t.logger.WithField("tx_hash", txHash).Debug("Subscribed to transaction")

// При получении события
t.logger.WithField("tx_hash", event.TxHash).
    WithField("height", event.Height).
    Info("Transaction event received via WebSocket")

// При переподключении
t.logger.WithField("attempt", attempt).
    WithField("backoff", backoff).
    Warn("WebSocket reconnecting")

// При fallback
t.logger.Warn("WebSocket unavailable, using polling fallback")
```

### 9.3 Health check

**Добавить endpoint:**
```go
// GET /health/websocket
func HandleWebSocketHealth(w http.ResponseWriter, r *http.Request) {
    status := map[string]interface{}{
        "connected": tracker.IsWebSocketConnected(),
        "subscriptions": tracker.GetSubscriptionCount(),
        "fallback_active": tracker.IsFallbackActive(),
    }
    json.NewEncoder(w).Encode(status)
}
```

---

## Этап 10: Документация

### 10.1 Обновить README трекера

**Файл: `server/transactions/tracker/README.md`**

Добавить разделы:
- Архитектура WebSocket
- Конфигурация
- Fallback механизм
- Troubleshooting

### 10.2 API документация

Обновить `server/API_DOCUMENTATION.md`:
- Добавить описание WebSocket health endpoint
- Объяснить улучшения производительности

### 10.3 Troubleshooting guide

**Проблемы и решения:**

1. **WebSocket не подключается**
   - Проверить доступность `ws://localhost:26657/websocket`
   - Проверить firewall
   - Проверить логи CometBFT

2. **Подписки не работают**
   - Проверить формат query
   - Проверить логи WebSocket клиента
   - Убедиться что транзакция существует

3. **Частые переподключения**
   - Проверить стабильность сети
   - Увеличить timeout
   - Проверить нагрузку на ноду

---

## Порядок реализации

1. ✅ **Этап 1**: Исследование (1-2 часа)
2. ✅ **Этап 2**: Конфигурация (30 мин)
3. ✅ **Этап 3**: WebSocket клиент (4-6 часов)
4. ✅ **Этап 4**: JSON-RPC протокол (2-3 часа)
5. ✅ **Этап 5**: Переподключение (2 часа)
6. ✅ **Этап 6**: Интеграция с трекером (2-3 часа)
7. ✅ **Этап 7**: Fallback логика (1-2 часа)
8. ✅ **Этап 8**: Тестирование (3-4 часа)
9. ✅ **Этап 9**: Мониторинг (1-2 часа)
10. ✅ **Этап 10**: Документация (1-2 часа)

**Общее время:** ~20-30 часов

---

## Критерии успеха

- ✅ WebSocket успешно подключается к CometBFT
- ✅ Подписки на транзакции работают
- ✅ События обрабатываются мгновенно (< 1 секунда)
- ✅ Автоматическое переподключение работает
- ✅ Fallback на polling при недоступности WebSocket
- ✅ Все тесты проходят
- ✅ Документация обновлена

---

## Риски и митигация

| Риск | Вероятность | Митигация |
|------|-------------|-----------|
| WebSocket API изменился | Низкая | Использовать стабильную версию CometBFT |
| Проблемы с переподключением | Средняя | Тщательное тестирование, fallback на polling |
| Утечки памяти | Средняя | Правильное закрытие каналов, context cancellation |
| Производительность | Низкая | Мониторинг метрик, оптимизация при необходимости |

---

## Следующие шаги

После завершения можно рассмотреть:
- WebSocket для real-time уведомлений клиентам
- Индексер для аналитики
- Подписки на события модулей (assets, leverage, etc.)


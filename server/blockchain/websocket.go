package blockchain

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/gorilla/websocket"
)

// WebSocketClient manages WebSocket connection to CometBFT
type WebSocketClient struct {
	url    string
	conn   *websocket.Conn
	connMu sync.RWMutex
	dialer *websocket.Dialer

	subscriptions map[string]*Subscription // txHash -> Subscription
	subsMu        sync.RWMutex
	subIDCounter  int64

	reconnectInterval time.Duration
	timeout           time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Channels for message handling
	sendCh chan *JSONRPCRequest
	recvCh chan *JSONRPCResponse

	// State
	connected        atomic.Bool
	reconnectAttempt int64
	lastError        error
	errorMu          sync.RWMutex
}

// Config for WebSocket client
type WebSocketConfig struct {
	ReconnectInterval time.Duration
	Timeout           time.Duration
}

// Subscription represents a subscription to a transaction
type Subscription struct {
	ID     int64
	TxHash string
	Query  string
	Events chan *TxEvent
	Done   chan struct{}
	mu     sync.RWMutex
	active bool
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	ID      int64       `json:"id"`
	Params  interface{} `json:"params"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Params  *EventParams    `json:"params,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// EventParams represents event parameters from CometBFT
type EventParams struct {
	Result *EventResult `json:"result"`
}

// EventResult contains the actual event data
type EventResult struct {
	Query string     `json:"query"`
	Data  *EventData `json:"data"`
}

// EventData contains the transaction event
type EventData struct {
	Type  string       `json:"type"`
	Value *TxEventData `json:"value"`
}

// TxEventData contains transaction result
type TxEventData struct {
	TxResult *TxResult `json:"TxResult"`
}

// TxResult contains the transaction result
type TxResult struct {
	Height interface{} `json:"height"` // Can be string or int64
	Tx     string      `json:"tx"`
	Result *Result     `json:"result"`
}

// Result contains transaction execution result
type Result struct {
	Code   uint32  `json:"code"`
	Log    string  `json:"log"`
	Events []Event `json:"events"`
}

// Event represents a blockchain event
type Event struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

// Attribute represents an event attribute
type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TxEvent represents a parsed transaction event
type TxEvent struct {
	TxHash  string
	Height  int64
	Code    uint32
	Log     string
	Events  []Event
	Success bool
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(url string, cfg WebSocketConfig) *WebSocketClient {
	if cfg.ReconnectInterval <= 0 {
		cfg.ReconnectInterval = 5 * time.Second
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WebSocketClient{
		url:               url,
		dialer:            websocket.DefaultDialer,
		subscriptions:     make(map[string]*Subscription),
		reconnectInterval: cfg.ReconnectInterval,
		timeout:           cfg.Timeout,
		ctx:               ctx,
		cancel:            cancel,
		sendCh:            make(chan *JSONRPCRequest, 100),
		recvCh:            make(chan *JSONRPCResponse, 100),
	}
}

// Connect establishes WebSocket connection
func (ws *WebSocketClient) Connect(ctx context.Context) error {
	ws.connMu.Lock()
	defer ws.connMu.Unlock()

	if ws.conn != nil {
		ws.conn.Close()
	}

	conn, _, err := ws.dialer.DialContext(ctx, ws.url, nil)
	if err != nil {
		ws.setError(err)
		ws.connected.Store(false)
		return fmt.Errorf("failed to connect: %w", err)
	}

	ws.conn = conn
	ws.connected.Store(true)
	ws.setError(nil)
	atomic.StoreInt64(&ws.reconnectAttempt, 0)

	// Start goroutines
	ws.wg.Add(4)
	go ws.readLoop()
	go ws.writeLoop()
	go ws.keepAlive()
	go ws.reconnectLoop()

	return nil
}

// IsConnected returns whether WebSocket is connected
func (ws *WebSocketClient) IsConnected() bool {
	return ws.connected.Load()
}

// Subscribe subscribes to a transaction by hash
func (ws *WebSocketClient) Subscribe(ctx context.Context, txHash string) (*Subscription, error) {
	if !ws.IsConnected() {
		return nil, fmt.Errorf("websocket not connected")
	}

	query := fmt.Sprintf("tx.hash='%s'", txHash)
	subID := atomic.AddInt64(&ws.subIDCounter, 1)

	sub := &Subscription{
		ID:     subID,
		TxHash: txHash,
		Query:  query,
		Events: make(chan *TxEvent, 1),
		Done:   make(chan struct{}),
		active: true,
	}

	// Store subscription
	ws.subsMu.Lock()
	ws.subscriptions[txHash] = sub
	ws.subsMu.Unlock()

	// Send subscribe request
	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "subscribe",
		ID:      subID,
		Params: map[string]string{
			"query": query,
		},
	}

	select {
	case ws.sendCh <- req:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(ws.timeout):
		return nil, fmt.Errorf("subscribe timeout")
	}

	return sub, nil
}

// Unsubscribe unsubscribes from a transaction
func (ws *WebSocketClient) Unsubscribe(txHash string) error {
	ws.subsMu.Lock()
	sub, exists := ws.subscriptions[txHash]
	if exists {
		delete(ws.subscriptions, txHash)
		sub.mu.Lock()
		sub.active = false
		close(sub.Done)
		sub.mu.Unlock()
	}
	ws.subsMu.Unlock()

	if !exists {
		return nil
	}

	// Send unsubscribe request
	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "unsubscribe",
		ID:      atomic.AddInt64(&ws.subIDCounter, 1),
		Params: map[string]string{
			"query": sub.Query,
		},
	}

	select {
	case ws.sendCh <- req:
	case <-time.After(ws.timeout):
		return fmt.Errorf("unsubscribe timeout")
	}

	return nil
}

// Close closes the WebSocket connection
func (ws *WebSocketClient) Close() error {
	ws.cancel()

	ws.connMu.Lock()
	if ws.conn != nil {
		ws.conn.Close()
		ws.conn = nil
	}
	ws.connMu.Unlock()

	ws.connected.Store(false)

	// Close all subscriptions
	ws.subsMu.Lock()
	for txHash, sub := range ws.subscriptions {
		sub.mu.Lock()
		sub.active = false
		close(sub.Done)
		sub.mu.Unlock()
		delete(ws.subscriptions, txHash)
	}
	ws.subsMu.Unlock()

	close(ws.sendCh)
	close(ws.recvCh)

	ws.wg.Wait()
	return nil
}

// readLoop reads messages from WebSocket
func (ws *WebSocketClient) readLoop() {
	defer ws.wg.Done()

	for {
		select {
		case <-ws.ctx.Done():
			return
		default:
		}

		ws.connMu.RLock()
		conn := ws.conn
		ws.connMu.RUnlock()

		if conn == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Set a longer read deadline (5 minutes) to avoid frequent timeouts
		// WebSocket can be idle for long periods
		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
		_, message, err := conn.ReadMessage()
		if err != nil {
			// Check if it's a timeout - this is normal for idle connections
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) ||
				err.Error() == "i/o timeout" {
				// Normal timeout, just continue reading
				continue
			}
			// Only log non-timeout errors
			log.Printf("[WebSocket] Read error: %v", err)
			ws.handleReadError(err)
			continue
		}

		// Log raw message for debugging (first 200 chars)
		msgStr := string(message)
		if len(msgStr) > 200 {
			msgStr = msgStr[:200] + "..."
		}
		log.Printf("[WebSocket] Raw message: %s", msgStr)

		var resp JSONRPCResponse
		if err := json.Unmarshal(message, &resp); err != nil {
			log.Printf("[WebSocket] Failed to unmarshal: %v", err)
			continue
		}

		// Log all messages for debugging monitor
		log.Printf("[WebSocket] Parsed: method='%s', id=%d, hasParams=%v, hasResult=%v, resultLen=%d",
			resp.Method, resp.ID, resp.Params != nil, len(resp.Result) > 0, len(resp.Result))

		ws.handleResponse(&resp)
	}
}

// writeLoop writes messages to WebSocket
func (ws *WebSocketClient) writeLoop() {
	defer ws.wg.Done()

	for {
		select {
		case <-ws.ctx.Done():
			return
		case req, ok := <-ws.sendCh:
			if !ok {
				return
			}

			ws.connMu.RLock()
			conn := ws.conn
			ws.connMu.RUnlock()

			if conn == nil {
				continue
			}

			conn.SetWriteDeadline(time.Now().Add(ws.timeout))
			if err := conn.WriteJSON(req); err != nil {
				ws.handleWriteError(err)
			}
		}
	}
}

// keepAlive sends ping messages to keep connection alive
func (ws *WebSocketClient) keepAlive() {
	defer ws.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ws.ctx.Done():
			return
		case <-ticker.C:
			ws.connMu.RLock()
			conn := ws.conn
			ws.connMu.RUnlock()

			if conn == nil {
				continue
			}

			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				ws.handleWriteError(err)
			}
		}
	}
}

// handleResponse processes incoming JSON-RPC responses
func (ws *WebSocketClient) handleResponse(resp *JSONRPCResponse) {
	if resp.Error != nil {
		log.Printf("[WebSocket] Error in response: %+v", resp.Error)
	}

	// Handle events from params (notifications with method="tm_event")
	if resp.Method == "tm_event" {
		log.Printf("[WebSocket] Processing tm_event notification")
		ws.handleEvent(resp)
		return
	}

	// Handle events from result (responses to subscribe with event data)
	if len(resp.Result) > 0 && resp.ID > 0 {
		// Check if result contains event data (not just empty {})
		var resultData map[string]interface{}
		if err := json.Unmarshal(resp.Result, &resultData); err == nil {
			// Check if it's an event (has "query" and "data" fields)
			if query, hasQuery := resultData["query"].(string); hasQuery {
				if _, hasData := resultData["data"].(map[string]interface{}); hasData {
					log.Printf("[WebSocket] Processing event from result, query: %s", query)
					ws.handleEventFromResult(resp.Result)
					return
				}
			}
		}
	}
}

// handleEvent processes transaction events
func (ws *WebSocketClient) handleEvent(resp *JSONRPCResponse) {
	log.Printf("[WebSocket] handleEvent called, Params=%v", resp.Params != nil)
	if resp.Params == nil {
		log.Printf("[WebSocket] handleEvent: Params is nil")
		return
	}
	if resp.Params.Result == nil {
		log.Printf("[WebSocket] handleEvent: Result is nil")
		return
	}

	log.Printf("[WebSocket] Parsing tx event, query: %s", resp.Params.Result.Query)
	txEvent := ws.parseTxEvent(resp.Params.Result)
	if txEvent == nil {
		log.Printf("[WebSocket] parseTxEvent returned nil")
		return
	}

	log.Printf("[WebSocket] Received tx event: hash=%s, height=%d, success=%v", txEvent.TxHash, txEvent.Height, txEvent.Success)

	// Find subscription - check both specific hash and "__all__" subscription
	ws.subsMu.RLock()
	sub, exists := ws.subscriptions[txEvent.TxHash]
	if !exists {
		// Check for global subscription
		sub, exists = ws.subscriptions["__all__"]
	}
	ws.subsMu.RUnlock()

	if !exists {
		return
	}

	// Check if subscription is still active
	sub.mu.RLock()
	active := sub.active
	sub.mu.RUnlock()

	if !active {
		return
	}

	ws.dispatchEvent(txEvent)
}

// handleEventFromResult processes transaction events from result field
func (ws *WebSocketClient) handleEventFromResult(result json.RawMessage) {
	// Parse result as EventResult
	var resultData struct {
		Query string     `json:"query"`
		Data  *EventData `json:"data"`
	}

	if err := json.Unmarshal(result, &resultData); err != nil {
		log.Printf("[WebSocket] Failed to unmarshal result: %v", err)
		return
	}

	if resultData.Data == nil {
		return
	}

	// Convert to EventResult format
	eventResult := &EventResult{
		Query: resultData.Query,
		Data:  resultData.Data,
	}

	txEvent := ws.parseTxEvent(eventResult)
	if txEvent == nil {
		log.Printf("[WebSocket] parseTxEvent from result returned nil")
		return
	}

	ws.dispatchEvent(txEvent)
}

// dispatchEvent dispatches transaction event to appropriate subscription
func (ws *WebSocketClient) dispatchEvent(txEvent *TxEvent) {
	log.Printf("[WebSocket] Received tx event: hash=%s, height=%d, success=%v", txEvent.TxHash, txEvent.Height, txEvent.Success)

	// Find subscription - check both specific hash and "__all__" subscription
	ws.subsMu.RLock()

	// Log all subscriptions for debugging
	allKeys := make([]string, 0, len(ws.subscriptions))
	for k := range ws.subscriptions {
		allKeys = append(allKeys, k)
	}
	log.Printf("[WebSocket] Available subscriptions: %v", allKeys)

	// Find both specific and global subscriptions
	specificSub, hasSpecific := ws.subscriptions[txEvent.TxHash]
	globalSub, hasGlobal := ws.subscriptions["__all__"]
	ws.subsMu.RUnlock()

	// Send to both subscriptions if they exist
	subsToNotify := []*Subscription{}
	if hasSpecific {
		subsToNotify = append(subsToNotify, specificSub)
		log.Printf("[WebSocket] Found specific subscription for tx %s, subID=%d", txEvent.TxHash, specificSub.ID)
	}
	if hasGlobal {
		subsToNotify = append(subsToNotify, globalSub)
		log.Printf("[WebSocket] Found global subscription __all__, subID=%d", globalSub.ID)
	}

	if len(subsToNotify) == 0 {
		log.Printf("[WebSocket] No subscription found for tx %s", txEvent.TxHash)
		return
	}

	// Send event to all matching subscriptions
	for _, sub := range subsToNotify {
		// Check if subscription is still active
		sub.mu.RLock()
		active := sub.active
		sub.mu.RUnlock()

		if !active {
			continue
		}

		// Send event to subscription channel
		select {
		case sub.Events <- txEvent:
			log.Printf("[WebSocket] Event dispatched to subscription (tx=%s, subID=%d)", txEvent.TxHash, sub.ID)
		case <-sub.Done:
			log.Printf("[WebSocket] Subscription done channel closed, cannot send event")
		case <-time.After(5 * time.Second):
			log.Printf("[WebSocket] Timeout sending event to subscription (tx=%s, subID=%d)", txEvent.TxHash, sub.ID)
		}
	}
}

// parseTxEvent parses transaction event from CometBFT response
func (ws *WebSocketClient) parseTxEvent(result *EventResult) *TxEvent {
	if result == nil || result.Data == nil || result.Data.Value == nil {
		return nil
	}

	txResult := result.Data.Value.TxResult
	if txResult == nil || txResult.Result == nil {
		return nil
	}

	// Extract tx hash - try from query first, then from transaction bytes
	txHash := extractTxHashFromQuery(result.Query)
	if txHash == "" && txResult.Tx != "" {
		// For "all transactions" subscription, hash is not in query
		// We need to compute it from the transaction bytes
		// txResult.Tx is base64 encoded transaction
		txHash = computeTxHashFromBase64(txResult.Tx)
	}
	if txHash == "" {
		return nil
	}

	// Convert height to int64 (can be string or number)
	var height int64
	switch v := txResult.Height.(type) {
	case string:
		if h, err := strconv.ParseInt(v, 10, 64); err == nil {
			height = h
		}
	case int64:
		height = v
	case int:
		height = int64(v)
	case float64:
		height = int64(v)
	}

	success := txResult.Result.Code == 0

	return &TxEvent{
		TxHash:  txHash,
		Height:  height,
		Code:    txResult.Result.Code,
		Log:     txResult.Result.Log,
		Events:  txResult.Result.Events,
		Success: success,
	}
}

// computeTxHashFromBase64 computes transaction hash from base64 encoded transaction
func computeTxHashFromBase64(txBase64 string) string {
	// Decode base64
	txBytes, err := base64.StdEncoding.DecodeString(txBase64)
	if err != nil {
		return ""
	}
	// Compute hash using tmhash (same as in client.go)
	hashBytes := tmhash.Sum(txBytes)
	return strings.ToLower(hex.EncodeToString(hashBytes))
}

// extractTxHashFromQuery extracts transaction hash from query string
func extractTxHashFromQuery(query string) string {
	// Query format: "tx.hash='ABC123...'"
	prefix := "tx.hash='"
	suffix := "'"

	start := len(prefix)
	if len(query) < start {
		return ""
	}

	if !strings.HasPrefix(query, prefix) {
		return ""
	}

	end := strings.Index(query[start:], suffix)
	if end == -1 {
		return ""
	}

	return query[start : start+end]
}

// handleReadError handles read errors and triggers reconnection
func (ws *WebSocketClient) handleReadError(err error) {
	ws.setError(err)
	ws.connected.Store(false)

	ws.connMu.Lock()
	if ws.conn != nil {
		ws.conn.Close()
		ws.conn = nil
	}
	ws.connMu.Unlock()
}

// handleWriteError handles write errors
func (ws *WebSocketClient) handleWriteError(err error) {
	ws.setError(err)
	ws.connected.Store(false)
}

// setError sets the last error
func (ws *WebSocketClient) setError(err error) {
	ws.errorMu.Lock()
	ws.lastError = err
	ws.errorMu.Unlock()
}

// GetLastError returns the last error
func (ws *WebSocketClient) GetLastError() error {
	ws.errorMu.RLock()
	defer ws.errorMu.RUnlock()
	return ws.lastError
}

// GetSubscriptionCount returns the number of active subscriptions
func (ws *WebSocketClient) GetSubscriptionCount() int {
	ws.subsMu.RLock()
	defer ws.subsMu.RUnlock()
	return len(ws.subscriptions)
}

// reconnectLoop handles automatic reconnection
func (ws *WebSocketClient) reconnectLoop() {
	defer ws.wg.Done()

	ticker := time.NewTicker(ws.reconnectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ws.ctx.Done():
			return
		case <-ticker.C:
			if !ws.IsConnected() {
				attempt := atomic.AddInt64(&ws.reconnectAttempt, 1)
				backoff := calculateBackoff(int(attempt))

				time.Sleep(backoff)

				// Try to reconnect
				ws.connMu.Lock()
				conn, _, err := ws.dialer.DialContext(ws.ctx, ws.url, nil)
				if err != nil {
					ws.connMu.Unlock()
					ws.setError(err)
					continue
				}

				// Close old connection if exists
				if ws.conn != nil {
					ws.conn.Close()
				}

				ws.conn = conn
				ws.connected.Store(true)
				ws.setError(nil)
				atomic.StoreInt64(&ws.reconnectAttempt, 0)
				ws.connMu.Unlock()

				// Restore all subscriptions
				ws.restoreSubscriptions()

				// Restart read/write loops if needed
				// They will automatically restart on next iteration
			}
		}
	}
}

// calculateBackoff calculates exponential backoff delay
func calculateBackoff(attempt int) time.Duration {
	base := 1 * time.Second
	max := 60 * time.Second

	backoff := time.Duration(1<<uint(attempt)) * base
	if backoff > max {
		backoff = max
	}

	return backoff
}

// restoreSubscriptions restores all active subscriptions after reconnection
func (ws *WebSocketClient) restoreSubscriptions() {
	ws.subsMu.RLock()
	defer ws.subsMu.RUnlock()

	for _, sub := range ws.subscriptions {
		sub.mu.RLock()
		active := sub.active
		sub.mu.RUnlock()

		if !active {
			continue
		}

		// Resubscribe
		req := &JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "subscribe",
			ID:      sub.ID,
			Params: map[string]string{
				"query": sub.Query,
			},
		}

		select {
		case ws.sendCh <- req:
		case <-time.After(ws.timeout):
			// Timeout - skip this subscription
		}
	}
}

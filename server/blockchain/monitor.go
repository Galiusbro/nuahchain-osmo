package blockchain

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"sync/atomic"
)

// BlockchainMonitor monitors all transactions on the blockchain
type BlockchainMonitor struct {
	wsClient *WebSocketClient
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup

	// Channel for broadcasting all transactions
	TxEvents chan *TxEvent
	sub      *Subscription

	mu sync.RWMutex
}

// NewBlockchainMonitor creates a new blockchain monitor
func NewBlockchainMonitor(wsClient *WebSocketClient) *BlockchainMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &BlockchainMonitor{
		wsClient: wsClient,
		ctx:      ctx,
		cancel:   cancel,
		TxEvents: make(chan *TxEvent, 1000), // Large buffer for high throughput
	}
}

// SubscribeAll subscribes to all transactions on the blockchain
func (ws *WebSocketClient) SubscribeAll(ctx context.Context) (*Subscription, error) {
	if !ws.IsConnected() {
		return nil, fmt.Errorf("websocket not connected")
	}

	// Subscribe to all transaction events
	query := "tm.event='Tx'"
	subID := atomic.AddInt64(&ws.subIDCounter, 1)

	sub := &Subscription{
		ID:     subID,
		TxHash: "__all__", // Special marker for "all transactions"
		Query:  query,
		Events: make(chan *TxEvent, 100),
		Done:   make(chan struct{}),
		active: true,
	}

	// Store subscription
	ws.subsMu.Lock()
	ws.subscriptions["__all__"] = sub
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
		log.Printf("[Monitor] Subscribed to all blockchain transactions")
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(ws.timeout):
		return nil, fmt.Errorf("subscribe timeout")
	}

	return sub, nil
}

// Start starts monitoring all transactions on the blockchain
func (m *BlockchainMonitor) Start() error {
	if m.wsClient == nil || !m.wsClient.IsConnected() {
		return fmt.Errorf("websocket not connected")
	}

	// Subscribe to all transactions
	sub, err := m.wsClient.SubscribeAll(m.ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe to all transactions: %w", err)
	}

	log.Printf("[Monitor] SubscribeAll returned sub ID=%d, Events channel=%p", sub.ID, sub.Events)

	m.mu.Lock()
	m.sub = sub
	m.mu.Unlock()

	log.Printf("[Monitor] Stored subscription, m.sub.ID=%d, m.sub.Events=%p", m.sub.ID, m.sub.Events)

	// Start processing events
	m.wg.Add(1)
	go m.processEvents()

	return nil
}

// processEvents processes incoming transaction events
func (m *BlockchainMonitor) processEvents() {
	defer m.wg.Done()
	log.Printf("[Monitor] processEvents started, waiting for events from subscription")

	for {
		select {
		case <-m.ctx.Done():
			log.Printf("[Monitor] processEvents context cancelled")
			return
		case event := <-m.sub.Events:
			if event == nil {
				log.Printf("[Monitor] Received nil event")
				continue
			}
			log.Printf("[Monitor] Received event from subscription: tx=%s, height=%d", event.TxHash, event.Height)

			// Broadcast to all listeners
			select {
			case m.TxEvents <- event:
				log.Printf("[Monitor] Event sent to TxEvents channel: tx=%s, height=%d", event.TxHash, event.Height)
			case <-m.ctx.Done():
				return
			default:
				// Channel full, skip this event
				log.Printf("[Monitor] TxEvents channel full, dropping event for tx %s", event.TxHash)
			}
		case <-m.sub.Done:
			log.Printf("[Monitor] Subscription closed")
			return
		}
	}
}

// Stop stops the monitor
func (m *BlockchainMonitor) Stop() {
	m.cancel()

	// Unsubscribe
	if m.wsClient != nil {
		m.wsClient.Unsubscribe("__all__")
	}

	m.wg.Wait()
	close(m.TxEvents)
}

// GetTxEvents returns the channel for receiving transaction events
func (m *BlockchainMonitor) GetTxEvents() <-chan *TxEvent {
	return m.TxEvents
}


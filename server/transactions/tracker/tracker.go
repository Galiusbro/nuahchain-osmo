package tracker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/logger"
	"github.com/osmosis-labs/osmosis/v30/server/transactions"
)

// Config задаёт параметры трекера транзакций
type Config struct {
	PollInterval     time.Duration
	MaxAttempts      int
	InitialBatchSize int
	UseWebSocket     bool
	WebSocketClient  *blockchain.WebSocketClient
}

// Tracker отслеживает статус on-chain транзакций и обновляет записи в БД
type Tracker struct {
	log           *logger.Logger
	repo          *transactions.Repository
	blockchainCli *blockchain.Client

	cfg Config

	queue  chan string
	active map[string]struct{}
	mu     sync.Mutex

	// WebSocket state
	useWebSocket   bool
	fallbackToPoll bool
	fallbackMu     sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New создаёт новый экземпляр трекера
func New(log *logger.Logger, repo *transactions.Repository, blockchainCli *blockchain.Client, cfg Config) *Tracker {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 3 * time.Second
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 30
	}
	if cfg.InitialBatchSize < 0 {
		cfg.InitialBatchSize = 0
	}

	tracker := &Tracker{
		log:            log,
		repo:           repo,
		blockchainCli:  blockchainCli,
		cfg:            cfg,
		queue:          make(chan string, 1024),
		active:         make(map[string]struct{}),
		useWebSocket:   cfg.UseWebSocket && cfg.WebSocketClient != nil,
		fallbackToPoll: false,
	}

	// Start WebSocket health check if WebSocket is enabled
	if tracker.useWebSocket && tracker.cfg.WebSocketClient != nil {
		tracker.wg.Add(1)
		go tracker.checkWebSocketHealth()
	}

	return tracker
}

// Start запускает фоновых воркеров и подхватывает незавершённые транзакции
func (t *Tracker) Start(parent context.Context) error {
	if parent == nil {
		parent = context.Background()
	}
	t.ctx, t.cancel = context.WithCancel(parent)

	// Подхватываем уже существующие PENDING транзакции
	pending, err := t.repo.ListPendingTransactions(t.cfg.InitialBatchSize)
	if err != nil {
		t.log.WithError(err).Error("transaction tracker failed to load pending transactions")
	} else {
		for _, tx := range pending {
			if tx.TxHash != "" {
				t.enqueue(tx.TxHash)
			}
		}
		t.log.WithField("count", len(pending)).Info("transaction tracker picked up pending transactions")
	}

	t.wg.Add(1)
	go t.worker()

	return nil
}

// Stop останавливает трекер и ожидает завершения фоновых задач
func (t *Tracker) Stop(ctx context.Context) {
	if t.cancel != nil {
		t.cancel()
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		t.wg.Wait()
	}()

	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-ctx.Done():
	case <-done:
	}
}

// Track добавляет транзакцию к отслеживанию
func (t *Tracker) Track(txHash string) {
	if txHash == "" {
		return
	}
	t.enqueue(txHash)
}

func (t *Tracker) enqueue(txHash string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.active[txHash]; exists {
		return
	}
	t.active[txHash] = struct{}{}

	select {
	case t.queue <- txHash:
	default:
		// Если очередь заполнена, логируем и добавляем блокирующе
		t.log.WithField("tx_hash", txHash).Warn("transaction tracker queue is full; blocking enqueue")
		t.queue <- txHash
	}
}

func (t *Tracker) worker() {
	defer t.wg.Done()

	for {
		select {
		case <-t.ctx.Done():
			return
		case txHash := <-t.queue:
			t.handleTx(txHash)
		}
	}
}

func (t *Tracker) handleTx(txHash string) {
	defer t.markDone(txHash)

	// Try WebSocket if available and not in fallback mode
	if t.useWebSocket && !t.isFallbackActive() {
		if err := t.handleTxWebSocket(txHash); err == nil {
			return // Successfully handled via WebSocket
		}
		// WebSocket failed - switch to fallback
		t.setFallback(true)
		t.log.WithField("tx_hash", txHash).Warn("WebSocket failed, falling back to polling")
	}

	// Fallback to polling
	t.handleTxPolling(txHash)
}

// handleTxWebSocket handles transaction tracking via WebSocket
func (t *Tracker) handleTxWebSocket(txHash string) error {
	if t.cfg.WebSocketClient == nil || !t.cfg.WebSocketClient.IsConnected() {
		return fmt.Errorf("websocket not available")
	}

	// IMPORTANT: Subscribe FIRST, then check if already in block
	// This ensures we catch the event if transaction is included while we're subscribing
	sub, err := t.cfg.WebSocketClient.Subscribe(t.ctx, txHash)
	if err != nil {
		t.log.WithError(err).WithField("tx_hash", txHash).Warn("WebSocket subscribe failed")
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	defer t.cfg.WebSocketClient.Unsubscribe(txHash)

	// Subscription created, will wait for event or use quickCheck

	// Quick check if transaction is already in a block (race condition handling)
	// But don't return immediately - still wait for WebSocket event in case it comes
	quickCheck, err := t.blockchainCli.GetTxStatus(t.ctx, txHash)
	// Wait for event with timeout (shorter timeout if already found)
	timeout := 5 * time.Minute
	if err == nil && quickCheck != nil && quickCheck.Found {
		// Transaction already found - use shorter timeout (10 seconds) to see if WebSocket event arrives
		// If not, we'll use the quickCheck result
		timeout = 10 * time.Second
		t.log.WithField("tx_hash", txHash).Debug("transaction already found, waiting briefly for WebSocket event")
	} else if err == nil && quickCheck != nil && !quickCheck.Found {
		// Transaction not found yet - will wait for WebSocket event
	}

	timeoutCh := time.After(timeout)

	select {
	case event := <-sub.Events:
		// Update database
		status := transactions.StatusSuccess
		var errorMsg *string
		if !event.Success {
			status = transactions.StatusFailed
			msg := event.Log
			if msg == "" {
				msg = fmt.Sprintf("transaction failed with code %d", event.Code)
			}
			errorMsg = &msg
		}

		if err := t.repo.UpdateTransactionByTxHash(txHash, status, nil, errorMsg); err != nil {
			t.log.WithError(err).WithField("tx_hash", txHash).Error("failed to update transaction status")
			return err
		}

		t.log.WithField("tx_hash", txHash).
			WithField("status", status).
			WithField("height", event.Height).
			Info("transaction confirmed via WebSocket")
		return nil

	case <-timeoutCh:
		// Timeout - check if we have quickCheck result
		if err == nil && quickCheck != nil && quickCheck.Found {
			// Use quickCheck result
			status := transactions.StatusSuccess
			var errorMsg *string
			if !quickCheck.Success {
				status = transactions.StatusFailed
				msg := quickCheck.Error
				if msg == "" {
					msg = quickCheck.Log
				}
				if msg == "" {
					msg = fmt.Sprintf("transaction failed with code %d", quickCheck.Code)
				}
				errorMsg = &msg
			}

			if err := t.repo.UpdateTransactionByTxHash(txHash, status, nil, errorMsg); err != nil {
				t.log.WithError(err).WithField("tx_hash", txHash).Error("failed to update transaction status")
				return err
			}

			t.log.WithField("tx_hash", txHash).
				WithField("status", status).
				Info("transaction confirmed via quickCheck (WebSocket timeout)")
			return nil
		}

		t.log.WithField("tx_hash", txHash).Warn("WebSocket subscription timeout, no transaction found")
		return fmt.Errorf("transaction timeout")

	case <-t.ctx.Done():
		return t.ctx.Err()

	case <-sub.Done:
		t.log.WithField("tx_hash", txHash).Warn("WebSocket subscription closed")
		return fmt.Errorf("subscription closed")
	}
}

// handleTxPolling handles transaction tracking via polling (original implementation)
func (t *Tracker) handleTxPolling(txHash string) {
	attempts := 0
	for {
		if err := t.ctx.Err(); err != nil {
			return
		}

		attempts++
		resp, err := t.blockchainCli.GetTxStatus(t.ctx, txHash)
		if err != nil {
			t.log.WithError(err).WithField("tx_hash", txHash).Warn("transaction tracker: get tx status failed")
		} else if resp != nil && resp.Found {
			if resp.Success {
				if err := t.repo.UpdateTransactionByTxHash(txHash, transactions.StatusSuccess, nil, nil); err != nil {
					t.log.WithError(err).WithField("tx_hash", txHash).Error("failed to update transaction status to SUCCESS")
				} else {
					t.log.WithField("tx_hash", txHash).Info("transaction confirmed successfully via polling")
				}
			} else {
				errMsg := resp.Error
				if errMsg == "" {
					errMsg = resp.Log
				}
				if errMsg == "" {
					errMsg = fmt.Sprintf("transaction failed with code %d", resp.Code)
				}
				if updateErr := t.repo.UpdateTransactionByTxHash(txHash, transactions.StatusFailed, nil, &errMsg); updateErr != nil {
					t.log.WithError(updateErr).WithField("tx_hash", txHash).Error("failed to update transaction status to FAILED")
				} else {
					t.log.WithField("tx_hash", txHash).WithField("error", errMsg).Info("transaction failed on-chain")
				}
			}
			return
		}

		if attempts >= t.cfg.MaxAttempts {
			errMsg := fmt.Sprintf("transaction not confirmed after %d attempts", attempts)
			if updateErr := t.repo.UpdateTransactionByTxHash(txHash, transactions.StatusFailed, nil, &errMsg); updateErr != nil {
				t.log.WithError(updateErr).WithField("tx_hash", txHash).Error("failed to update transaction status after timeout")
			} else {
				t.log.WithField("tx_hash", txHash).Warn("transaction confirmation timed out")
			}
			return
		}

		select {
		case <-t.ctx.Done():
			return
		case <-time.After(t.cfg.PollInterval):
		}
	}
}

// checkWebSocketHealth periodically checks WebSocket health and manages fallback
func (t *Tracker) checkWebSocketHealth() {
	defer t.wg.Done()

	// Wait for context to be initialized
	if t.ctx == nil {
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			if t.cfg.WebSocketClient == nil {
				continue
			}

			if t.isFallbackActive() {
				// Check if WebSocket is restored
				if t.cfg.WebSocketClient.IsConnected() {
					t.setFallback(false)
					t.log.Info("WebSocket restored, switching back from polling")
				}
			} else if t.useWebSocket {
				// Check if WebSocket is still connected
				if !t.cfg.WebSocketClient.IsConnected() {
					t.setFallback(true)
					t.log.Warn("WebSocket disconnected, switching to polling")
				}
			}
		}
	}
}

// isFallbackActive returns whether fallback to polling is active
func (t *Tracker) isFallbackActive() bool {
	t.fallbackMu.RLock()
	defer t.fallbackMu.RUnlock()
	return t.fallbackToPoll
}

// setFallback sets the fallback state
func (t *Tracker) setFallback(active bool) {
	t.fallbackMu.Lock()
	defer t.fallbackMu.Unlock()
	t.fallbackToPoll = active
}

// IsWebSocketConnected returns whether WebSocket is connected
func (t *Tracker) IsWebSocketConnected() bool {
	return t.useWebSocket && t.cfg.WebSocketClient != nil && t.cfg.WebSocketClient.IsConnected()
}

// IsFallbackActive returns whether fallback to polling is active
func (t *Tracker) IsFallbackActive() bool {
	return t.isFallbackActive()
}

func (t *Tracker) markDone(txHash string) {
	t.mu.Lock()
	delete(t.active, txHash)
	t.mu.Unlock()
}

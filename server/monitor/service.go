package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/logger"
)

// Service provides blockchain monitoring functionality for admin panel
type Service struct {
	monitor       *blockchain.BlockchainMonitor
	blockchainCli *blockchain.Client
	repo          *Repository
	log           *logger.Logger
	txHistory     []*TxInfo // Keep in-memory for quick access
	mu            sync.RWMutex
	maxHistory    int
}

// TxInfo represents transaction information for admin panel
type TxInfo struct {
	TxHash  string    `json:"tx_hash"`
	Height  int64     `json:"height"`
	Success bool      `json:"success"`
	Code    uint32    `json:"code"`
	Log     string    `json:"log,omitempty"`
	Time    time.Time `json:"time"`
	Events  []string  `json:"events,omitempty"` // Event types
}

// NewService creates a new monitoring service
func NewService(monitor *blockchain.BlockchainMonitor, blockchainCli *blockchain.Client, repo *Repository, log *logger.Logger) *Service {
	return &Service{
		monitor:       monitor,
		blockchainCli: blockchainCli,
		repo:          repo,
		log:           log,
		txHistory:     make([]*TxInfo, 0),
		maxHistory:    1000, // Keep last 1000 transactions in memory
	}
}

// Start starts the monitoring service
func (s *Service) Start(ctx context.Context) error {
	if err := s.monitor.Start(); err != nil {
		return err
	}

	// Start processing events
	go s.processEvents(ctx)

	return nil
}

// processEvents processes incoming transaction events
func (s *Service) processEvents(ctx context.Context) {
	s.log.Info("Monitor processEvents started")
	for {
		select {
		case <-ctx.Done():
			s.log.Info("Monitor processEvents stopped")
			return
		case event := <-s.monitor.GetTxEvents():
			if event == nil {
				s.log.Warn("Received nil event")
				continue
			}

			s.log.WithField("tx_hash", event.TxHash).
				WithField("height", event.Height).
				Info("Processing transaction event")

			// Extract data from events
			operationType, sender, moduleName := s.extractFromEvents(event.Events)

			// Convert events to map for JSON storage
			eventsMap := s.eventsToMap(event.Events)

			// Save to database immediately with WebSocket data
			createReq := CreateBlockchainTransactionRequest{
				TxHash:        event.TxHash,
				Height:        event.Height,
				Code:          int(event.Code),
				Success:       event.Success,
				Log:           stringPtr(event.Log),
				Events:        eventsMap,
				OperationType: stringPtr(operationType),
				Sender:        stringPtr(sender),
				ModuleName:    stringPtr(moduleName),
			}

			_, err := s.repo.CreateTransaction(createReq)
			if err != nil {
				s.log.WithError(err).WithField("tx_hash", event.TxHash).
					Error("Failed to save transaction to database")
			} else {
				s.log.WithField("tx_hash", event.TxHash).
					Debug("Transaction saved to database")
			}

			// Enrich with GetTx data asynchronously (get full details like nuahd query tx)
			// DISABLED: Enrichment is disabled to avoid extra blockchain queries
			// go s.enrichTransaction(ctx, event.TxHash)

			// Also keep in memory for quick access
			txInfo := &TxInfo{
				TxHash:  event.TxHash,
				Height:  event.Height,
				Success: event.Success,
				Code:    event.Code,
				Log:     event.Log,
				Time:    time.Now(),
			}

			// Extract event types
			for _, evt := range event.Events {
				txInfo.Events = append(txInfo.Events, evt.Type)
			}

			// Add to history
			s.mu.Lock()
			s.txHistory = append(s.txHistory, txInfo)
			// Keep only last maxHistory transactions
			if len(s.txHistory) > s.maxHistory {
				s.txHistory = s.txHistory[len(s.txHistory)-s.maxHistory:]
			}
			s.mu.Unlock()

			s.log.WithField("tx_hash", event.TxHash).
				WithField("height", event.Height).
				WithField("success", event.Success).
				Info("Blockchain transaction monitored and added to history")
		}
	}
}

// enrichTransaction enriches transaction with full data from GetTx (like nuahd query tx)
func (s *Service) enrichTransaction(ctx context.Context, txHash string) {
	// Wait a bit for transaction to be indexed
	time.Sleep(2 * time.Second)

	// Get full transaction details via WaitForTx (uses gRPC GetTx, same as nuahd query tx)
	txResp, err := s.blockchainCli.WaitForTx(ctx, txHash, 3, 2*time.Second)
	if err != nil {
		s.log.WithError(err).WithField("tx_hash", txHash).
			Warn("Failed to enrich transaction via WaitForTx")
		return
	}

	if txResp == nil {
		s.log.WithField("tx_hash", txHash).
			Warn("Transaction response is nil")
		return
	}

	// ЛОГИРУЕМ СЫРЫЕ ДАННЫЕ для анализа типов
	s.log.WithField("tx_hash", txHash).
		WithField("txResp.Height", txResp.Height).
		WithField("txResp.Code", txResp.Code).
		WithField("txResp.Codespace", txResp.Codespace).
		WithField("txResp.GasWanted", txResp.GasWanted).
		WithField("txResp.GasUsed", txResp.GasUsed).
		WithField("txResp.Data", txResp.Data).
		WithField("txResp.RawLog", txResp.RawLog).
		WithField("txResp.Timestamp", txResp.Timestamp).
		WithField("txResp.Tx != nil", txResp.Tx != nil).
		Info("RAW TxResponse data received")

	// Decode transaction to extract messages, signers, fee
	var messages []map[string]interface{}
	var signers []string
	var feeAmount map[string]interface{}
	var memo *string

	if txResp.Tx != nil {
		s.log.WithField("tx_hash", txHash).
			WithField("txResp.Tx.TypeUrl", txResp.Tx.TypeUrl).
			WithField("txResp.Tx.Value length", len(txResp.Tx.Value)).
			Info("RAW Tx Any data")

		decodedTx, err := s.blockchainCli.DecodeTxFromResponse(txResp.Tx)
		if err != nil {
			s.log.WithError(err).WithField("tx_hash", txHash).
				Warn("Failed to decode transaction from TxResponse")
		} else {
			messages = decodedTx.MessagesData
			signers = decodedTx.Signers
			feeAmount = decodedTx.Fee
			if decodedTx.Memo != "" {
				memo = &decodedTx.Memo
			}

			// ЛОГИРУЕМ ДЕКОДИРОВАННЫЕ ДАННЫЕ
			s.log.WithField("tx_hash", txHash).
				WithField("decoded_messages_count", len(messages)).
				WithField("decoded_signers_count", len(signers)).
				WithField("decoded_fee", feeAmount).
				WithField("decoded_memo", memo).
				Info("DECODED transaction data")
		}
	}

	// Parse timestamp if available
	var blockTimestamp *time.Time
	if txResp.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, txResp.Timestamp); err == nil {
			blockTimestamp = &t
		}
	}

	// Update transaction with enriched data
	updateReq := UpdateBlockchainTransactionRequest{
		Codespace:      stringPtr(txResp.Codespace),
		RawLog:         stringPtr(txResp.RawLog),
		GasWanted:      int64Ptr(txResp.GasWanted),
		GasUsed:        int64Ptr(txResp.GasUsed),
		Data:           stringPtr(txResp.Data),
		Messages:       messages,
		Signers:        signers,
		FeeAmount:      feeAmount,
		Memo:           memo,
		BlockTimestamp: blockTimestamp,
		Enriched:       true,
		EnrichedAt:     timePtr(time.Now()),
	}

	// ЛОГИРУЕМ ВСЕ ДАННЫЕ ПЕРЕД ОБНОВЛЕНИЕМ
	s.log.WithField("tx_hash", txHash).
		WithField("updateReq.Codespace", updateReq.Codespace).
		WithField("updateReq.RawLog", updateReq.RawLog != nil).
		WithField("updateReq.GasWanted", updateReq.GasWanted).
		WithField("updateReq.GasUsed", updateReq.GasUsed).
		WithField("updateReq.Data", updateReq.Data != nil).
		WithField("updateReq.Messages", messages != nil).
		WithField("updateReq.Messages_len", len(messages)).
		WithField("updateReq.Signers", signers != nil).
		WithField("updateReq.Signers_len", len(signers)).
		WithField("updateReq.FeeAmount", feeAmount != nil).
		WithField("updateReq.Memo", memo != nil).
		WithField("updateReq.BlockTimestamp", blockTimestamp != nil).
		WithField("updateReq.Enriched", updateReq.Enriched).
		WithField("updateReq.EnrichedAt", updateReq.EnrichedAt != nil).
		Info("ALL updateReq data before DB update")

	if err := s.repo.UpdateTransaction(txHash, updateReq); err != nil {
		s.log.WithError(err).WithField("tx_hash", txHash).
			Error("Failed to update transaction with enriched data")
	} else {
		s.log.WithField("tx_hash", txHash).
			WithField("gas_used", txResp.GasUsed).
			WithField("gas_wanted", txResp.GasWanted).
			WithField("messages_count", len(messages)).
			Info("Transaction enriched with full data (like nuahd query tx)")
	}
}

// extractFromEvents extracts operation type, sender, and module from events
func (s *Service) extractFromEvents(events []blockchain.Event) (operationType, sender, moduleName string) {
	for _, evt := range events {
		if evt.Type == "message" {
			for _, attr := range evt.Attributes {
				if attr.Key == "action" {
					operationType = attr.Value
				}
				if attr.Key == "sender" {
					sender = attr.Value
				}
				if attr.Key == "module" {
					moduleName = attr.Value
				}
			}
		}
		// Also check for module-specific events
		if moduleName == "" {
			// Try to extract from event type (e.g., "assets.buy", "leverage.open_position")
			if len(evt.Type) > 0 {
				parts := splitEventType(evt.Type)
				if len(parts) > 0 {
					moduleName = parts[0]
				}
			}
		}
	}
	return
}

// eventsToMap converts events to map for JSON storage
func (s *Service) eventsToMap(events []blockchain.Event) map[string]interface{} {
	result := make(map[string]interface{})
	eventsList := make([]map[string]interface{}, 0, len(events))

	for _, evt := range events {
		eventMap := map[string]interface{}{
			"type": evt.Type,
		}
		attrs := make([]map[string]string, 0, len(evt.Attributes))
		for _, attr := range evt.Attributes {
			attrs = append(attrs, map[string]string{
				"key":   attr.Key,
				"value": attr.Value,
			})
		}
		eventMap["attributes"] = attrs
		eventsList = append(eventsList, eventMap)
	}

	result["events"] = eventsList
	return result
}

// Helper functions
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func int64Ptr(i int64) *int64 {
	// Always return pointer, even for 0 (gas can be 0)
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func splitEventType(eventType string) []string {
	// Split by "." to get module and action
	// e.g., "assets.buy" -> ["assets", "buy"]
	parts := []string{}
	current := ""
	for _, r := range eventType {
		if r == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// GetRecentTransactions returns recent transactions
func (s *Service) GetRecentTransactions(limit int) []*TxInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	start := len(s.txHistory) - limit
	if start < 0 {
		start = 0
	}

	// Return in reverse order (newest first)
	result := make([]*TxInfo, 0, limit)
	for i := len(s.txHistory) - 1; i >= start; i-- {
		result = append(result, s.txHistory[i])
	}

	return result
}

// GetTransactionByHash returns transaction info by hash
func (s *Service) GetTransactionByHash(txHash string) *TxInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := len(s.txHistory) - 1; i >= 0; i-- {
		if s.txHistory[i].TxHash == txHash {
			return s.txHistory[i]
		}
	}

	return nil
}

// GetStats returns monitoring statistics
func (s *Service) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	successCount := 0
	failedCount := 0

	for _, tx := range s.txHistory {
		if tx.Success {
			successCount++
		} else {
			failedCount++
		}
	}

	return map[string]interface{}{
		"total_transactions": len(s.txHistory),
		"successful":         successCount,
		"failed":             failedCount,
		"max_history":        s.maxHistory,
	}
}

// GetTxEvents returns the channel for receiving transaction events (exposed for handlers)
func (s *Service) GetTxEvents() <-chan *blockchain.TxEvent {
	return s.monitor.GetTxEvents()
}

// Stop stops the monitoring service
func (s *Service) Stop() {
	s.monitor.Stop()
}

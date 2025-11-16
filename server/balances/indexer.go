package balances

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/logger"
)

// Indexer indexes balance changes from blockchain events
type Indexer struct {
	monitor       *blockchain.BlockchainMonitor
	repo          *Repository
	log           *logger.Logger
	updateChannel chan *BalanceUpdate // Channel for WebSocket updates
	mu            sync.RWMutex
	errorCount    int64 // Track consecutive errors for alerts
}

// BalanceUpdate represents a balance update event for WebSocket
type BalanceUpdate struct {
	UserID    int64     `json:"user_id"`
	Address   string    `json:"address"`
	Denom     string    `json:"denom"`
	Amount    string    `json:"amount"`
	TxHash    string    `json:"tx_hash"`
	Height    int64     `json:"height"`
	Timestamp time.Time `json:"timestamp"`
}

// NewIndexer creates a new balance indexer
func NewIndexer(monitor *blockchain.BlockchainMonitor, repo *Repository, log *logger.Logger) *Indexer {
	return &Indexer{
		monitor:       monitor,
		repo:          repo,
		log:           log,
		updateChannel: make(chan *BalanceUpdate, 1000),
	}
}

// GetUpdateChannel returns the channel for balance updates (for WebSocket)
func (idx *Indexer) GetUpdateChannel() <-chan *BalanceUpdate {
	return idx.updateChannel
}

// Start starts the balance indexer
func (idx *Indexer) Start(ctx context.Context) error {
	if idx.monitor == nil {
		return fmt.Errorf("blockchain monitor not configured")
	}

	// Start processing events
	go idx.processEvents(ctx)

	// Start recovery process (check for missed blocks)
	go idx.recoveryProcess(ctx)

	idx.log.Info("Balance indexer started")
	return nil
}

// processEvents processes incoming transaction events
func (idx *Indexer) processEvents(ctx context.Context) {
	idx.log.Info("Balance indexer processEvents started")

	for {
		select {
		case <-ctx.Done():
			idx.log.Info("Balance indexer processEvents stopped")
			return
		case event := <-idx.monitor.GetTxEvents():
			if event == nil {
				continue
			}

			// Only process successful transactions
			if !event.Success {
				continue
			}

			// Process event with retry
			processErr := idx.processEventWithRetry(ctx, event)
			if processErr != nil {
				idx.log.WithError(processErr).
					WithField("tx_hash", event.TxHash).
					WithField("height", event.Height).
					Error("Failed to process balance event after retries")
			}

			// Update indexer state
			var errMsg *string
			if processErr != nil {
				errStr := processErr.Error()
				errMsg = &errStr
			}
			if err := idx.repo.UpdateIndexerState(event.Height, errMsg); err != nil {
				idx.log.WithError(err).Warn("Failed to update indexer state")
			}
		}
	}
}

// processEventWithRetry processes an event with exponential backoff retry
func (idx *Indexer) processEventWithRetry(ctx context.Context, event *blockchain.TxEvent) error {
	maxRetries := 4
	delays := []time.Duration{0, 1 * time.Second, 2 * time.Second, 4 * time.Second}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delays[attempt]):
			}
		}

		err := idx.processEvent(event)
		if err == nil {
			// Success - reset error count
			idx.mu.Lock()
			idx.errorCount = 0
			idx.mu.Unlock()
			return nil
		}

		lastErr = err
		idx.log.WithError(err).
			WithField("tx_hash", event.TxHash).
			WithField("attempt", attempt+1).
			Warn("Failed to process balance event, retrying")
	}

	// All retries failed - increment error count
	idx.mu.Lock()
	idx.errorCount++
	errorCount := idx.errorCount
	idx.mu.Unlock()

	// Alert if too many consecutive errors
	if errorCount >= 10 {
		idx.log.WithField("error_count", errorCount).
			Error("CRITICAL: Too many consecutive balance indexer errors")
	}

	// Save to DLQ
	idx.saveToDLQ(event, lastErr)

	return lastErr
}

// processEvent processes a single transaction event
func (idx *Indexer) processEvent(event *blockchain.TxEvent) error {
	// Parse balance changes from events
	balanceChanges, err := idx.parseBalanceChanges(event)
	if err != nil {
		return fmt.Errorf("failed to parse balance changes: %w", err)
	}

	if len(balanceChanges) == 0 {
		return nil // No balance changes in this transaction
	}

	// Process each balance change
	for _, change := range balanceChanges {
		if err := idx.processBalanceChange(event, change); err != nil {
			return fmt.Errorf("failed to process balance change: %w", err)
		}
	}

	return nil
}

// BalanceChange represents a parsed balance change from events
type BalanceChange struct {
	Address string
	Denom   string
	Delta   string // Positive for increase, negative for decrease
}

// parseBalanceChanges parses balance changes from transaction events
func (idx *Indexer) parseBalanceChanges(event *blockchain.TxEvent) ([]BalanceChange, error) {
	var changes []BalanceChange

	for _, evt := range event.Events {
		switch evt.Type {
		case "transfer":
			// Parse transfer events
			transferChanges, err := idx.parseTransferEvent(evt)
			if err != nil {
				idx.log.WithError(err).Warn("Failed to parse transfer event")
				continue
			}
			changes = append(changes, transferChanges...)

		case "coin_received":
			// Parse coin_received events
			receivedChanges, err := idx.parseCoinReceivedEvent(evt)
			if err != nil {
				idx.log.WithError(err).Warn("Failed to parse coin_received event")
				continue
			}
			changes = append(changes, receivedChanges...)

		case "coin_spent":
			// Parse coin_spent events
			spentChanges, err := idx.parseCoinSpentEvent(evt)
			if err != nil {
				idx.log.WithError(err).Warn("Failed to parse coin_spent event")
				continue
			}
			changes = append(changes, spentChanges...)
		}
	}

	return changes, nil
}

// parseTransferEvent parses a transfer event
func (idx *Indexer) parseTransferEvent(evt blockchain.Event) ([]BalanceChange, error) {
	var sender, recipient, amount string

	for _, attr := range evt.Attributes {
		switch attr.Key {
		case "sender":
			sender = attr.Value
		case "recipient":
			recipient = attr.Value
		case "amount":
			amount = attr.Value
		}
	}

	if sender == "" || recipient == "" || amount == "" {
		return nil, nil // Incomplete event
	}

	var changes []BalanceChange

	// Parse amount (can contain multiple denoms: "1000unuah,5000factory/.../token")
	amounts := idx.parseAmounts(amount)
	for _, amt := range amounts {
		// Decrease for sender
		changes = append(changes, BalanceChange{
			Address: sender,
			Denom:   amt.Denom,
			Delta:   "-" + amt.Amount, // Negative delta
		})

		// Increase for recipient
		changes = append(changes, BalanceChange{
			Address: recipient,
			Denom:   amt.Denom,
			Delta:   amt.Amount, // Positive delta
		})
	}

	return changes, nil
}

// parseCoinReceivedEvent parses a coin_received event
func (idx *Indexer) parseCoinReceivedEvent(evt blockchain.Event) ([]BalanceChange, error) {
	var receiver, amount string

	for _, attr := range evt.Attributes {
		switch attr.Key {
		case "receiver":
			receiver = attr.Value
		case "amount":
			amount = attr.Value
		}
	}

	if receiver == "" || amount == "" {
		return nil, nil
	}

	var changes []BalanceChange
	amounts := idx.parseAmounts(amount)
	for _, amt := range amounts {
		changes = append(changes, BalanceChange{
			Address: receiver,
			Denom:   amt.Denom,
			Delta:   amt.Amount, // Positive delta
		})
	}

	return changes, nil
}

// parseCoinSpentEvent parses a coin_spent event
func (idx *Indexer) parseCoinSpentEvent(evt blockchain.Event) ([]BalanceChange, error) {
	var spender, amount string

	for _, attr := range evt.Attributes {
		switch attr.Key {
		case "spender":
			spender = attr.Value
		case "amount":
			amount = attr.Value
		}
	}

	if spender == "" || amount == "" {
		return nil, nil
	}

	var changes []BalanceChange
	amounts := idx.parseAmounts(amount)
	for _, amt := range amounts {
		changes = append(changes, BalanceChange{
			Address: spender,
			Denom:   amt.Denom,
			Delta:   "-" + amt.Amount, // Negative delta
		})
	}

	return changes, nil
}

// Amount represents a parsed amount with denom
type Amount struct {
	Denom  string
	Amount string
}

// parseAmounts parses amount string (can be "1000unuah" or "1000unuah,5000factory/.../token")
func (idx *Indexer) parseAmounts(amountStr string) []Amount {
	var amounts []Amount

	// Split by comma for multiple denoms
	parts := strings.Split(amountStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Parse amount and denom (format: "1000unuah")
		// Find the first non-digit character
		var denomStart int
		for i, r := range part {
			if (r < '0' || r > '9') && r != '.' {
				denomStart = i
				break
			}
		}

		if denomStart == 0 {
			continue // Invalid format
		}

		amount := part[:denomStart]
		denom := part[denomStart:]

		amounts = append(amounts, Amount{
			Denom:  denom,
			Amount: amount,
		})
	}

	return amounts
}

// processBalanceChange processes a single balance change
func (idx *Indexer) processBalanceChange(event *blockchain.TxEvent, change BalanceChange) error {
	// Check if address belongs to a user
	userID, err := idx.repo.GetUserIDByAddress(change.Address)
	if err != nil {
		return fmt.Errorf("failed to get user ID for address %s: %w", change.Address, err)
	}

	if userID == nil {
		// Address doesn't belong to any user, skip
		return nil
	}

	// Get current balance
	currentBalance, err := idx.repo.GetBalanceByAddress(change.Address, change.Denom)
	if err != nil {
		return fmt.Errorf("failed to get current balance: %w", err)
	}

	// Calculate new balance
	var newAmount string
	if currentBalance == nil {
		// First balance entry
		if strings.HasPrefix(change.Delta, "-") {
			// Can't have negative balance from start
			return nil
		}
		newAmount = change.Delta
	} else {
		// Calculate new balance
		currentBig, ok := new(big.Int).SetString(currentBalance.Amount, 10)
		if !ok {
			return fmt.Errorf("invalid current balance: %s", currentBalance.Amount)
		}

		deltaBig, ok := new(big.Int).SetString(strings.TrimPrefix(change.Delta, "-"), 10)
		if !ok {
			return fmt.Errorf("invalid delta: %s", change.Delta)
		}

		if strings.HasPrefix(change.Delta, "-") {
			newAmount = new(big.Int).Sub(currentBig, deltaBig).String()
		} else {
			newAmount = new(big.Int).Add(currentBig, deltaBig).String()
		}

		// Don't update if balance would be negative (shouldn't happen, but safety check)
		if strings.HasPrefix(newAmount, "-") {
			idx.log.WithField("address", change.Address).
				WithField("denom", change.Denom).
				WithField("current", currentBalance.Amount).
				WithField("delta", change.Delta).
				Warn("Balance would be negative, skipping update")
			return nil
		}
	}

	// Determine event type
	eventType := "transfer"
	if strings.HasPrefix(change.Delta, "-") {
		eventType = "coin_spent"
	} else {
		eventType = "coin_received"
	}

	// Get amount before for history
	var amountBefore *string
	if currentBalance != nil {
		amountBefore = &currentBalance.Amount
	}

	// Upsert balance
	req := UpdateBalanceRequest{
		UserID:       *userID,
		Address:      change.Address,
		Denom:        change.Denom,
		Amount:       newAmount,
		TxHash:       event.TxHash,
		Height:       event.Height,
		EventType:    eventType,
		AmountBefore: amountBefore,
	}

	if err := idx.repo.UpsertBalance(req); err != nil {
		return fmt.Errorf("failed to upsert balance: %w", err)
	}

	// Send update to WebSocket channel
	select {
	case idx.updateChannel <- &BalanceUpdate{
		UserID:    *userID,
		Address:   change.Address,
		Denom:     change.Denom,
		Amount:    newAmount,
		TxHash:    event.TxHash,
		Height:    event.Height,
		Timestamp: time.Now(),
	}:
	default:
		// Channel full, log but don't fail
		idx.log.Warn("Balance update channel full, dropping update")
	}

	return nil
}

// saveToDLQ saves a failed update to dead letter queue
func (idx *Indexer) saveToDLQ(event *blockchain.TxEvent, err error) {
	// Try to extract address and denom from error or event
	// For now, save with minimal info
	address := "unknown"
	denom := "unknown"
	amountDelta := "0"

	// Try to parse from events
	for _, evt := range event.Events {
		if evt.Type == "transfer" {
			for _, attr := range evt.Attributes {
				if attr.Key == "sender" || attr.Key == "recipient" {
					address = attr.Value
				}
				if attr.Key == "amount" {
					amounts := idx.parseAmounts(attr.Value)
					if len(amounts) > 0 {
						denom = amounts[0].Denom
						amountDelta = amounts[0].Amount
					}
				}
			}
		}
	}

	if err := idx.repo.SaveFailedUpdate(
		event.TxHash,
		event.Height,
		address,
		denom,
		amountDelta,
		err.Error(),
	); err != nil {
		idx.log.WithError(err).Error("Failed to save to DLQ")
	}
}

// recoveryProcess checks for missed blocks and processes them
func (idx *Indexer) recoveryProcess(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check if indexer is behind
			lastHeight, lastTime, err := idx.repo.GetIndexerState()
			if err != nil {
				idx.log.WithError(err).Warn("Failed to get indexer state")
				continue
			}

			// If last processed was more than 5 minutes ago, log warning
			if time.Since(lastTime) > 5*time.Minute {
				idx.log.WithField("last_height", lastHeight).
					WithField("last_time", lastTime).
					Warn("Balance indexer appears to be stuck")
			}
		}
	}
}

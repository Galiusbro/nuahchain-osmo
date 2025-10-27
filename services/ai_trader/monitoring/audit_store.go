package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	osmomath "github.com/osmosis-labs/osmosis/osmomath"
)

// AuditStore provides storage and retrieval of audit events
type AuditStore struct {
	events    map[string]AuditEvent
	alerts    map[string]Alert
	mu        sync.RWMutex
	logger    *Logger
	maxEvents int
	maxAlerts int
}

// NewAuditStore creates a new audit store
func NewAuditStore(logger *Logger, maxEvents, maxAlerts int) *AuditStore {
	return &AuditStore{
		events:    make(map[string]AuditEvent),
		alerts:    make(map[string]Alert),
		logger:    logger,
		maxEvents: maxEvents,
		maxAlerts: maxAlerts,
	}
}

// StoreEvent stores an audit event
func (s *AuditStore) StoreEvent(event AuditEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store event
	s.events[event.ID] = event

	// Maintain max events limit
	if len(s.events) > s.maxEvents {
		// Remove oldest events (simple implementation)
		count := 0
		for id := range s.events {
			if count >= len(s.events)-s.maxEvents {
				break
			}
			delete(s.events, id)
			count++
		}
	}
}

// StoreAlert stores an alert
func (s *AuditStore) StoreAlert(alert Alert) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store alert
	s.alerts[alert.ID] = alert

	// Maintain max alerts limit
	if len(s.alerts) > s.maxAlerts {
		// Remove oldest alerts (simple implementation)
		count := 0
		for id := range s.alerts {
			if count >= len(s.alerts)-s.maxAlerts {
				break
			}
			delete(s.alerts, id)
			count++
		}
	}
}

// GetEvents returns events matching the given criteria
func (s *AuditStore) GetEvents(ctx context.Context, req EventQueryRequest) ([]AuditEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []AuditEvent

	for _, event := range s.events {
		if s.matchesEvent(event, req) {
			result = append(result, event)
		}
	}

	// Sort by timestamp (newest first)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Timestamp.Before(result[j].Timestamp) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	// Apply pagination
	start := req.Offset
	if start >= len(result) {
		return []AuditEvent{}, nil
	}

	end := start + req.Limit
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], nil
}

// GetAlerts returns alerts matching the given criteria
func (s *AuditStore) GetAlerts(ctx context.Context, req AlertQueryRequest) ([]Alert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Alert

	for _, alert := range s.alerts {
		if s.matchesAlert(alert, req) {
			result = append(result, alert)
		}
	}

	// Sort by timestamp (newest first)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Timestamp.Before(result[j].Timestamp) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	// Apply pagination
	start := req.Offset
	if start >= len(result) {
		return []Alert{}, nil
	}

	end := start + req.Limit
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], nil
}

// GetTradeHistory returns trading history for a specific trader and symbol
func (s *AuditStore) GetTradeHistory(ctx context.Context, trader, symbol string) (*TradeHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := &TradeHistory{
		Trader:      trader,
		Symbol:      symbol,
		RiskMetrics: make(map[string]interface{}),
	}

	var trades []AuditEvent
	var totalVolume sdk.Coin
	var totalProfit sdk.Coin
	var totalLoss sdk.Coin
	var priceSum float64
	var priceCount int

	for _, event := range s.events {
		if event.Trader == trader &&
			(symbol == "" || event.Symbol == symbol) &&
			(event.EventType == EventTypeTradeExecuted || event.EventType == EventTypeTradeRequest) {
			trades = append(trades, event)

			// Update counters
			if event.EventType == EventTypeTradeExecuted {
				if event.Success {
					history.SuccessfulTrades++
				} else {
					history.FailedTrades++
				}

				// Update volume (simplified - assumes all trades are buy orders)
				if event.Metadata != nil {
					if amountStr, ok := event.Metadata["amount"].(string); ok {
						if coin, err := sdk.ParseCoinNormalized(amountStr); err == nil {
							if totalVolume.Denom == "" {
								totalVolume = coin
							} else {
								totalVolume = totalVolume.Add(coin)
							}
						}
					}
				}

				// Update price average
				if event.Price != "" {
					if price, err := osmomath.NewDecFromStr(event.Price); err == nil {
						priceSum += price.MustFloat64()
						priceCount++
					}
				}
			}
		}
	}

	// Count total trades (both executed and requested)
	history.TotalTrades = len(trades)
	history.TotalVolume = totalVolume
	history.TotalProfit = totalProfit
	history.TotalLoss = totalLoss

	if priceCount > 0 {
		avgPrice := priceSum / float64(priceCount)
		if !isNaN(avgPrice) {
			history.AveragePrice = fmt.Sprintf("%.2f", avgPrice)
		} else {
			history.AveragePrice = "0.00"
		}
	}

	if len(trades) > 0 {
		history.FirstTradeAt = trades[len(trades)-1].Timestamp // Oldest
		history.LastTradeAt = trades[0].Timestamp              // Newest
	}

	// Calculate risk metrics
	history.RiskMetrics["total_violations"] = s.countViolations(trader, symbol)
	history.RiskMetrics["emergency_stops"] = s.countEmergencyStops(trader)
	if history.TotalTrades > 0 {
		successRate := float64(history.SuccessfulTrades) / float64(history.TotalTrades)
		if !isNaN(successRate) {
			history.RiskMetrics["success_rate"] = successRate
		} else {
			history.RiskMetrics["success_rate"] = 0.0
		}
	} else {
		history.RiskMetrics["success_rate"] = 0.0
	}

	return history, nil
}

// GetSystemMetrics returns system-wide metrics
func (s *AuditStore) GetSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := &SystemMetrics{
		Timestamp: time.Now(),
		Metrics:   make(map[string]interface{}),
	}

	traders := make(map[string]bool)
	var totalVolume sdk.Coin

	for _, event := range s.events {
		traders[event.Trader] = true

		if event.EventType == EventTypeTradeExecuted && event.Success {
			metrics.TotalTrades++

			// Update volume (simplified)
			if event.Metadata != nil {
				if amountStr, ok := event.Metadata["amount"].(string); ok {
					if coin, err := sdk.ParseCoinNormalized(amountStr); err == nil {
						if totalVolume.Denom == "" {
							totalVolume = coin
						} else {
							totalVolume = totalVolume.Add(coin)
						}
					}
				}
			}
		}

		if event.EventType == EventTypePolicyViolation {
			metrics.PolicyViolations++
		}

		if event.EventType == EventTypeEmergencyStop {
			metrics.EmergencyStops++
		}
	}

	metrics.ActiveTraders = len(traders)
	metrics.TotalVolume = totalVolume
	metrics.ActiveAlerts = len(s.alerts)

	// Calculate additional metrics
	metrics.Metrics["events_per_hour"] = s.calculateEventsPerHour()
	metrics.Metrics["average_trade_size"] = s.calculateAverageTradeSize()
	if metrics.TotalTrades > 0 {
		violationRate := float64(metrics.PolicyViolations) / float64(metrics.TotalTrades)
		if !isNaN(violationRate) {
			metrics.Metrics["violation_rate"] = violationRate
		} else {
			metrics.Metrics["violation_rate"] = 0.0
		}
	} else {
		metrics.Metrics["violation_rate"] = 0.0
	}

	return metrics, nil
}

// AcknowledgeAlert acknowledges an alert
func (s *AuditStore) AcknowledgeAlert(ctx context.Context, alertID, acknowledgedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	alert, exists := s.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	now := time.Now()
	alert.Acknowledged = true
	alert.AcknowledgedAt = &now
	alert.AcknowledgedBy = acknowledgedBy

	s.alerts[alertID] = alert
	return nil
}

// Helper methods

func (s *AuditStore) matchesEvent(event AuditEvent, req EventQueryRequest) bool {
	if req.Trader != "" && event.Trader != req.Trader {
		return false
	}
	if req.Symbol != "" && event.Symbol != req.Symbol {
		return false
	}
	if req.EventType != "" && event.EventType != EventType(req.EventType) {
		return false
	}
	if req.RiskLevel != "" && event.RiskLevel != RiskLevel(req.RiskLevel) {
		return false
	}
	if req.Success != nil && event.Success != *req.Success {
		return false
	}
	if !req.StartTime.IsZero() && event.Timestamp.Before(req.StartTime) {
		return false
	}
	if !req.EndTime.IsZero() && event.Timestamp.After(req.EndTime) {
		return false
	}
	return true
}

func (s *AuditStore) matchesAlert(alert Alert, req AlertQueryRequest) bool {
	if req.Trader != "" && alert.Trader != req.Trader {
		return false
	}
	if req.Symbol != "" && alert.Symbol != req.Symbol {
		return false
	}
	if req.AlertType != "" && alert.AlertType != AlertType(req.AlertType) {
		return false
	}
	if req.Severity != "" && alert.Severity != AlertSeverity(req.Severity) {
		return false
	}
	if req.Acknowledged != nil && alert.Acknowledged != *req.Acknowledged {
		return false
	}
	if !req.StartTime.IsZero() && alert.Timestamp.Before(req.StartTime) {
		return false
	}
	if !req.EndTime.IsZero() && alert.Timestamp.After(req.EndTime) {
		return false
	}
	return true
}

func (s *AuditStore) countViolations(trader, symbol string) int {
	count := 0
	for _, event := range s.events {
		if event.Trader == trader && event.Symbol == symbol &&
			event.EventType == EventTypePolicyViolation {
			count++
		}
	}
	return count
}

func (s *AuditStore) countEmergencyStops(trader string) int {
	count := 0
	for _, event := range s.events {
		if event.Trader == trader && event.EventType == EventTypeEmergencyStop {
			count++
		}
	}
	return count
}

func (s *AuditStore) calculateEventsPerHour() float64 {
	if len(s.events) == 0 {
		return 0
	}

	// Find time range
	var earliest, latest time.Time
	for _, event := range s.events {
		if earliest.IsZero() || event.Timestamp.Before(earliest) {
			earliest = event.Timestamp
		}
		if latest.IsZero() || event.Timestamp.After(latest) {
			latest = event.Timestamp
		}
	}

	duration := latest.Sub(earliest)
	if duration == 0 {
		return 0
	}

	return float64(len(s.events)) / duration.Hours()
}

func (s *AuditStore) calculateAverageTradeSize() float64 {
	var total float64
	var count int

	for _, event := range s.events {
		if event.EventType == EventTypeTradeExecuted && event.Success {
			if event.Metadata != nil {
				if amountStr, ok := event.Metadata["amount"].(string); ok {
					if coin, err := sdk.ParseCoinNormalized(amountStr); err == nil {
						total += float64(coin.Amount.Int64())
						count++
					}
				}
			}
		}
	}

	if count == 0 {
		return 0
	}

	return total / float64(count)
}

// Helper function to check for NaN
func isNaN(f float64) bool {
	return f != f
}

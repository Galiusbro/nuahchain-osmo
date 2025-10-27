package monitoring

import (
	"fmt"
	"sync"
	"time"
)

// AlertManager manages alerts and notifications
type AlertManager struct {
	auditStore *AuditStore
	logger     *Logger
	notifiers  []Notifier
	mu         sync.RWMutex
	rules      []AlertRule
}

// Notifier interface for sending alerts
type Notifier interface {
	SendAlert(alert Alert) error
	Name() string
}

// AlertRule defines when to trigger an alert
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Condition   AlertCondition         `json:"condition"`
	Severity    AlertSeverity          `json:"severity"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AlertCondition defines the condition for triggering an alert
type AlertCondition struct {
	Type       string                 `json:"type"` // event_count, volume_threshold, consecutive_losses, etc.
	Threshold  float64                `json:"threshold"`
	TimeWindow time.Duration          `json:"time_window"`
	Trader     string                 `json:"trader,omitempty"`
	Symbol     string                 `json:"symbol,omitempty"`
	EventType  string                 `json:"event_type,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// NewAlertManager creates a new alert manager
func NewAlertManager(auditStore *AuditStore, logger *Logger) *AlertManager {
	return &AlertManager{
		auditStore: auditStore,
		logger:     logger,
		notifiers:  make([]Notifier, 0),
		rules:      make([]AlertRule, 0),
	}
}

// AddNotifier adds a notifier to the alert manager
func (am *AlertManager) AddNotifier(notifier Notifier) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.notifiers = append(am.notifiers, notifier)
}

// AddRule adds an alert rule
func (am *AlertManager) AddRule(rule AlertRule) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.rules = append(am.rules, rule)
}

// CheckRules checks all alert rules and triggers alerts if conditions are met
func (am *AlertManager) CheckRules() error {
	am.mu.RLock()
	rules := make([]AlertRule, len(am.rules))
	copy(rules, am.rules)
	am.mu.RUnlock()

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		shouldTrigger, err := am.evaluateRule(rule)
		if err != nil {
			am.logger.logger.WithError(err).WithField("rule_id", rule.ID).Error("Failed to evaluate alert rule")
			continue
		}

		if shouldTrigger {
			alert := Alert{
				ID:        generateEventID(),
				Timestamp: time.Now(),
				AlertType: AlertType(rule.Condition.Type),
				Severity:  rule.Severity,
				Title:     rule.Name,
				Message:   rule.Description,
				Trader:    rule.Condition.Trader,
				Symbol:    rule.Condition.Symbol,
				Metadata:  rule.Metadata,
			}

			if err := am.triggerAlert(alert); err != nil {
				am.logger.logger.WithError(err).WithField("alert_id", alert.ID).Error("Failed to trigger alert")
			}
		}
	}

	return nil
}

// evaluateRule evaluates a single alert rule
func (am *AlertManager) evaluateRule(rule AlertRule) (bool, error) {
	switch rule.Condition.Type {
	case "event_count":
		return am.evaluateEventCountRule(rule)
	case "volume_threshold":
		return am.evaluateVolumeThresholdRule(rule)
	case "consecutive_losses":
		return am.evaluateConsecutiveLossesRule(rule)
	case "policy_violations":
		return am.evaluatePolicyViolationsRule(rule)
	case "emergency_stop":
		return am.evaluateEmergencyStopRule(rule)
	default:
		return false, fmt.Errorf("unknown rule type: %s", rule.Condition.Type)
	}
}

// evaluateEventCountRule evaluates an event count rule
func (am *AlertManager) evaluateEventCountRule(rule AlertRule) (bool, error) {
	req := EventQueryRequest{
		Trader:    rule.Condition.Trader,
		Symbol:    rule.Condition.Symbol,
		EventType: rule.Condition.EventType,
		StartTime: time.Now().Add(-rule.Condition.TimeWindow),
		EndTime:   time.Now(),
		Limit:     10000, // Large limit to count all events
	}

	events, err := am.auditStore.GetEvents(nil, req)
	if err != nil {
		return false, err
	}

	return float64(len(events)) >= rule.Condition.Threshold, nil
}

// evaluateVolumeThresholdRule evaluates a volume threshold rule
func (am *AlertManager) evaluateVolumeThresholdRule(rule AlertRule) (bool, error) {
	req := EventQueryRequest{
		Trader:    rule.Condition.Trader,
		Symbol:    rule.Condition.Symbol,
		EventType: "trade_executed",
		StartTime: time.Now().Add(-rule.Condition.TimeWindow),
		EndTime:   time.Now(),
		Limit:     10000,
	}

	events, err := am.auditStore.GetEvents(nil, req)
	if err != nil {
		return false, err
	}

	var totalVolume float64
	for _, event := range events {
		if event.Metadata != nil {
			if amountStr, ok := event.Metadata["amount"].(string); ok {
				// Parse amount and add to total (simplified)
				// In real implementation, would parse sdk.Coin properly
				if amount, err := parseAmount(amountStr); err == nil {
					totalVolume += amount
				}
			}
		}
	}

	return totalVolume >= rule.Condition.Threshold, nil
}

// evaluateConsecutiveLossesRule evaluates a consecutive losses rule
func (am *AlertManager) evaluateConsecutiveLossesRule(rule AlertRule) (bool, error) {
	req := EventQueryRequest{
		Trader:    rule.Condition.Trader,
		Symbol:    rule.Condition.Symbol,
		EventType: "trade_executed",
		StartTime: time.Now().Add(-rule.Condition.TimeWindow),
		EndTime:   time.Now(),
		Limit:     10000,
	}

	events, err := am.auditStore.GetEvents(nil, req)
	if err != nil {
		return false, err
	}

	consecutiveLosses := 0
	maxConsecutiveLosses := 0

	for _, event := range events {
		if !event.Success {
			consecutiveLosses++
			if consecutiveLosses > maxConsecutiveLosses {
				maxConsecutiveLosses = consecutiveLosses
			}
		} else {
			consecutiveLosses = 0
		}
	}

	return float64(maxConsecutiveLosses) >= rule.Condition.Threshold, nil
}

// evaluatePolicyViolationsRule evaluates a policy violations rule
func (am *AlertManager) evaluatePolicyViolationsRule(rule AlertRule) (bool, error) {
	req := EventQueryRequest{
		Trader:    rule.Condition.Trader,
		Symbol:    rule.Condition.Symbol,
		EventType: "policy_violation",
		StartTime: time.Now().Add(-rule.Condition.TimeWindow),
		EndTime:   time.Now(),
		Limit:     10000,
	}

	events, err := am.auditStore.GetEvents(nil, req)
	if err != nil {
		return false, err
	}

	return float64(len(events)) >= rule.Condition.Threshold, nil
}

// evaluateEmergencyStopRule evaluates an emergency stop rule
func (am *AlertManager) evaluateEmergencyStopRule(rule AlertRule) (bool, error) {
	req := EventQueryRequest{
		Trader:    rule.Condition.Trader,
		EventType: "emergency_stop",
		StartTime: time.Now().Add(-rule.Condition.TimeWindow),
		EndTime:   time.Now(),
		Limit:     10000,
	}

	events, err := am.auditStore.GetEvents(nil, req)
	if err != nil {
		return false, err
	}

	return len(events) > 0, nil
}

// triggerAlert triggers an alert by sending it to all notifiers
func (am *AlertManager) triggerAlert(alert Alert) error {
	// Store alert in audit store
	am.auditStore.StoreAlert(alert)

	// Log alert
	am.logger.LogAlert(alert)

	// Send to all notifiers
	am.mu.RLock()
	notifiers := make([]Notifier, len(am.notifiers))
	copy(notifiers, am.notifiers)
	am.mu.RUnlock()

	for _, notifier := range notifiers {
		if err := notifier.SendAlert(alert); err != nil {
			am.logger.logger.WithError(err).
				WithField("notifier", notifier.Name()).
				WithField("alert_id", alert.ID).
				Error("Failed to send alert")
		}
	}

	return nil
}

// DefaultAlertRules returns default alert rules
func DefaultAlertRules() []AlertRule {
	return []AlertRule{
		{
			ID:          "high_volume_alert",
			Name:        "High Volume Alert",
			Description: "Alert when trading volume exceeds threshold",
			Condition: AlertCondition{
				Type:       "volume_threshold",
				Threshold:  1000000, // 1M tokens
				TimeWindow: time.Hour,
			},
			Severity: AlertSeverityWarning,
			Enabled:  true,
		},
		{
			ID:          "consecutive_losses_alert",
			Name:        "Consecutive Losses Alert",
			Description: "Alert when consecutive losses exceed threshold",
			Condition: AlertCondition{
				Type:       "consecutive_losses",
				Threshold:  5,
				TimeWindow: time.Hour * 24,
			},
			Severity: AlertSeverityError,
			Enabled:  true,
		},
		{
			ID:          "policy_violations_alert",
			Name:        "Policy Violations Alert",
			Description: "Alert when policy violations exceed threshold",
			Condition: AlertCondition{
				Type:       "policy_violations",
				Threshold:  10,
				TimeWindow: time.Hour,
			},
			Severity: AlertSeverityWarning,
			Enabled:  true,
		},
		{
			ID:          "emergency_stop_alert",
			Name:        "Emergency Stop Alert",
			Description: "Alert when emergency stop is triggered",
			Condition: AlertCondition{
				Type:       "emergency_stop",
				Threshold:  1,
				TimeWindow: time.Minute * 5,
			},
			Severity: AlertSeverityCritical,
			Enabled:  true,
		},
	}
}

// Helper function to parse amount (simplified)
func parseAmount(amountStr string) (float64, error) {
	// This is a simplified implementation
	// In real implementation, would parse sdk.Coin properly
	return 0, fmt.Errorf("not implemented")
}

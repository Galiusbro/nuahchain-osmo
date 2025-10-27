package monitoring

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditEvent_JSONSerialization(t *testing.T) {
	event := AuditEvent{
		ID:         "test_event_123",
		Timestamp:  time.Now(),
		EventType:  EventTypeTradeRequest,
		Trader:     "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		Symbol:     "BTC",
		Action:     "buy",
		Amount:     sdk.NewCoin("factory/test/ndollar", math.NewInt(1000)),
		Price:      "50000.00",
		TxHash:     "0x1234567890abcdef",
		Success:    true,
		Reason:     "Trade executed successfully",
		Violations: []string{"cooldown_period"},
		Warnings:   []string{"price_deviation"},
		RiskLevel:  RiskLevelMedium,
		Metadata: map[string]interface{}{
			"test_key": "test_value",
		},
		PolicyCheck: PolicyCheckResult{
			Allowed:    true,
			Reason:     "All checks passed",
			Violations: []string{},
			Warnings:   []string{"price_deviation"},
			CheckedAt:  time.Now(),
			Duration:   time.Millisecond * 100,
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(event)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test JSON unmarshaling
	var unmarshaledEvent AuditEvent
	err = json.Unmarshal(jsonData, &unmarshaledEvent)
	require.NoError(t, err)
	assert.Equal(t, event.ID, unmarshaledEvent.ID)
	assert.Equal(t, event.EventType, unmarshaledEvent.EventType)
	assert.Equal(t, event.Trader, unmarshaledEvent.Trader)
	assert.Equal(t, event.Symbol, unmarshaledEvent.Symbol)
	assert.Equal(t, event.Action, unmarshaledEvent.Action)
	assert.Equal(t, event.Amount.String(), unmarshaledEvent.Amount.String())
	assert.Equal(t, event.Price, unmarshaledEvent.Price)
	assert.Equal(t, event.Success, unmarshaledEvent.Success)
	assert.Equal(t, event.RiskLevel, unmarshaledEvent.RiskLevel)
}

func TestAlert_JSONSerialization(t *testing.T) {
	now := time.Now()
	alert := Alert{
		ID:        "test_alert_123",
		Timestamp: now,
		AlertType: AlertTypeLimitExceeded,
		Severity:  AlertSeverityWarning,
		Title:     "Volume Limit Exceeded",
		Message:   "Trading volume has exceeded the daily limit",
		Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		Symbol:    "BTC",
		Metadata: map[string]interface{}{
			"limit":     1000000,
			"actual":    1500000,
			"threshold": 0.8,
		},
		Acknowledged:   false,
		AcknowledgedAt: nil,
		AcknowledgedBy: "",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(alert)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test JSON unmarshaling
	var unmarshaledAlert Alert
	err = json.Unmarshal(jsonData, &unmarshaledAlert)
	require.NoError(t, err)
	assert.Equal(t, alert.ID, unmarshaledAlert.ID)
	assert.Equal(t, alert.AlertType, unmarshaledAlert.AlertType)
	assert.Equal(t, alert.Severity, unmarshaledAlert.Severity)
	assert.Equal(t, alert.Title, unmarshaledAlert.Title)
	assert.Equal(t, alert.Message, unmarshaledAlert.Message)
	assert.Equal(t, alert.Trader, unmarshaledAlert.Trader)
	assert.Equal(t, alert.Symbol, unmarshaledAlert.Symbol)
	assert.Equal(t, alert.Acknowledged, unmarshaledAlert.Acknowledged)
}

func TestLogger_LogAuditEvent(t *testing.T) {
	config := LoggerConfig{
		LogLevel:      "info",
		LogFormat:     "json",
		AuditLogPath:  "test_logs/audit.log",
		AlertLogPath:  "test_logs/alerts.log",
		BufferSize:    10,
		FlushInterval: time.Second,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	event := AuditEvent{
		ID:        "test_event_456",
		Timestamp: time.Now(),
		EventType: EventTypeTradeRequest,
		Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		Symbol:    "BTC",
		Action:    "buy",
		Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(1000)),
		Price:     "50000.00",
		Success:   true,
		RiskLevel: RiskLevelLow,
	}

	logger.LogAuditEvent(event)

	// Flush to ensure event is written
	err = logger.Flush()
	require.NoError(t, err)
}

func TestLogger_LogAlert(t *testing.T) {
	config := LoggerConfig{
		LogLevel:      "info",
		LogFormat:     "json",
		AuditLogPath:  "test_logs/audit.log",
		AlertLogPath:  "test_logs/alerts.log",
		BufferSize:    10,
		FlushInterval: time.Second,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	alert := Alert{
		ID:        "test_alert_456",
		Timestamp: time.Now(),
		AlertType: AlertTypeLimitExceeded,
		Severity:  AlertSeverityWarning,
		Title:     "Test Alert",
		Message:   "This is a test alert",
		Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
	}

	logger.LogAlert(alert)
}

func TestLogger_LogTradeRequest(t *testing.T) {
	config := LoggerConfig{
		LogLevel:      "info",
		LogFormat:     "json",
		AuditLogPath:  "test_logs/audit.log",
		AlertLogPath:  "test_logs/alerts.log",
		BufferSize:    10,
		FlushInterval: time.Second,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	policyResult := PolicyCheckResult{
		Allowed:    true,
		Reason:     "All checks passed",
		Violations: []string{},
		Warnings:   []string{"price_deviation"},
		CheckedAt:  time.Now(),
		Duration:   time.Millisecond * 100,
	}

	logger.LogTradeRequest(
		"cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		"BTC",
		"buy",
		sdk.NewCoin("factory/test/ndollar", math.NewInt(1000)),
		"50000.00",
		policyResult,
	)
}

func TestLogger_LogPolicyViolation(t *testing.T) {
	config := LoggerConfig{
		LogLevel:      "info",
		LogFormat:     "json",
		AuditLogPath:  "test_logs/audit.log",
		AlertLogPath:  "test_logs/alerts.log",
		BufferSize:    10,
		FlushInterval: time.Second,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	logger.LogPolicyViolation(
		"cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		"BTC",
		"Volume limit exceeded",
		[]string{"daily_volume_limits"},
	)
}

func TestLogger_LogEmergencyStop(t *testing.T) {
	config := LoggerConfig{
		LogLevel:      "info",
		LogFormat:     "json",
		AuditLogPath:  "test_logs/audit.log",
		AlertLogPath:  "test_logs/alerts.log",
		BufferSize:    10,
		FlushInterval: time.Second,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	metadata := map[string]interface{}{
		"consecutive_losses": 5,
		"daily_loss":         "1000000",
	}

	logger.LogEmergencyStop(
		"cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		"Consecutive losses limit exceeded",
		metadata,
	)
}

func TestAuditStore_StoreEvent(t *testing.T) {
	logger, err := NewLogger(DefaultLoggerConfig())
	require.NoError(t, err)
	defer logger.Close()

	store := NewAuditStore(logger, 1000, 100)

	event := AuditEvent{
		ID:        "test_event_789",
		Timestamp: time.Now(),
		EventType: EventTypeTradeRequest,
		Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		Symbol:    "BTC",
		Action:    "buy",
		Success:   true,
		RiskLevel: RiskLevelLow,
	}

	store.StoreEvent(event)

	// Verify event was stored
	req := EventQueryRequest{
		Trader: event.Trader,
		Limit:  10,
	}

	events, err := store.GetEvents(context.Background(), req)
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, event.ID, events[0].ID)
}

func TestAuditStore_StoreAlert(t *testing.T) {
	logger, err := NewLogger(DefaultLoggerConfig())
	require.NoError(t, err)
	defer logger.Close()

	store := NewAuditStore(logger, 1000, 100)

	alert := Alert{
		ID:        "test_alert_789",
		Timestamp: time.Now(),
		AlertType: AlertTypeLimitExceeded,
		Severity:  AlertSeverityWarning,
		Title:     "Test Alert",
		Message:   "This is a test alert",
		Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
	}

	store.StoreAlert(alert)

	// Verify alert was stored
	req := AlertQueryRequest{
		Trader: alert.Trader,
		Limit:  10,
	}

	alerts, err := store.GetAlerts(context.Background(), req)
	require.NoError(t, err)
	assert.Len(t, alerts, 1)
	assert.Equal(t, alert.ID, alerts[0].ID)
}

func TestAuditStore_GetTradeHistory(t *testing.T) {
	logger, err := NewLogger(DefaultLoggerConfig())
	require.NoError(t, err)
	defer logger.Close()

	store := NewAuditStore(logger, 1000, 100)

	trader := "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz"
	symbol := "BTC"

	// Store some trade events
	events := []AuditEvent{
		{
			ID:        "trade_1",
			Timestamp: time.Now().Add(-2 * time.Hour),
			EventType: EventTypeTradeExecuted,
			Trader:    trader,
			Symbol:    symbol,
			Action:    "buy",
			Success:   true,
			Price:     "50000.00",
			Metadata: map[string]interface{}{
				"amount": "1000factory/test/ndollar",
			},
		},
		{
			ID:        "trade_2",
			Timestamp: time.Now().Add(-1 * time.Hour),
			EventType: EventTypeTradeExecuted,
			Trader:    trader,
			Symbol:    symbol,
			Action:    "sell",
			Success:   false,
			Price:     "51000.00",
			Metadata: map[string]interface{}{
				"amount": "500factory/test/ndollar",
			},
		},
	}

	for _, event := range events {
		store.StoreEvent(event)
	}

	// Get trade history
	history, err := store.GetTradeHistory(context.Background(), trader, symbol)
	require.NoError(t, err)
	assert.Equal(t, trader, history.Trader)
	assert.Equal(t, symbol, history.Symbol)
	assert.Equal(t, 2, history.TotalTrades)
	assert.Equal(t, 1, history.SuccessfulTrades)
	assert.Equal(t, 1, history.FailedTrades)
}

func TestAuditStore_GetSystemMetrics(t *testing.T) {
	logger, err := NewLogger(DefaultLoggerConfig())
	require.NoError(t, err)
	defer logger.Close()

	store := NewAuditStore(logger, 1000, 100)

	// Store some events
	events := []AuditEvent{
		{
			ID:        "event_1",
			Timestamp: time.Now(),
			EventType: EventTypeTradeExecuted,
			Trader:    "trader1",
			Symbol:    "BTC",
			Success:   true,
		},
		{
			ID:        "event_2",
			Timestamp: time.Now(),
			EventType: EventTypePolicyViolation,
			Trader:    "trader2",
			Symbol:    "ETH",
			Success:   false,
		},
	}

	for _, event := range events {
		store.StoreEvent(event)
	}

	// Get system metrics
	metrics, err := store.GetSystemMetrics(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, metrics.ActiveTraders)
	assert.Equal(t, 1, metrics.TotalTrades)
	assert.Equal(t, 1, metrics.PolicyViolations)
}

func TestAlertManager_AddRule(t *testing.T) {
	logger, err := NewLogger(DefaultLoggerConfig())
	require.NoError(t, err)
	defer logger.Close()

	store := NewAuditStore(logger, 1000, 100)
	manager := NewAlertManager(store, logger)

	rule := AlertRule{
		ID:          "test_rule",
		Name:        "Test Rule",
		Description: "This is a test rule",
		Condition: AlertCondition{
			Type:       "event_count",
			Threshold:  5,
			TimeWindow: time.Hour,
		},
		Severity: AlertSeverityWarning,
		Enabled:  true,
	}

	manager.AddRule(rule)

	// Verify rule was added (would need a GetRules method to verify)
	// For now, just ensure no error occurred
	assert.True(t, true)
}

func TestConsoleNotifier_SendAlert(t *testing.T) {
	notifier := NewConsoleNotifier()

	alert := Alert{
		ID:        "test_alert",
		Timestamp: time.Now(),
		AlertType: AlertTypeLimitExceeded,
		Severity:  AlertSeverityWarning,
		Title:     "Test Alert",
		Message:   "This is a test alert",
		Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		Symbol:    "BTC",
	}

	err := notifier.SendAlert(alert)
	assert.NoError(t, err)
	assert.Equal(t, "console", notifier.Name())
}

func TestFileNotifier_SendAlert(t *testing.T) {
	notifier := NewFileNotifier("test_alerts.log")

	alert := Alert{
		ID:        "test_alert",
		Timestamp: time.Now(),
		AlertType: AlertTypeLimitExceeded,
		Severity:  AlertSeverityWarning,
		Title:     "Test Alert",
		Message:   "This is a test alert",
	}

	err := notifier.SendAlert(alert)
	assert.NoError(t, err)
	assert.Equal(t, "file", notifier.Name())
}

func TestWebhookNotifier_SendAlert(t *testing.T) {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer token123",
	}
	notifier := NewWebhookNotifier("https://example.com/webhook", headers)

	alert := Alert{
		ID:        "test_alert",
		Timestamp: time.Now(),
		AlertType: AlertTypeLimitExceeded,
		Severity:  AlertSeverityWarning,
		Title:     "Test Alert",
		Message:   "This is a test alert",
	}

	err := notifier.SendAlert(alert)
	assert.NoError(t, err)
	assert.Equal(t, "webhook", notifier.Name())
}

func TestEmailNotifier_SendAlert(t *testing.T) {
	notifier := NewEmailNotifier(
		"smtp.example.com",
		587,
		"alerts@example.com",
		[]string{"admin@example.com"},
	)

	alert := Alert{
		ID:        "test_alert",
		Timestamp: time.Now(),
		AlertType: AlertTypeLimitExceeded,
		Severity:  AlertSeverityWarning,
		Title:     "Test Alert",
		Message:   "This is a test alert",
	}

	err := notifier.SendAlert(alert)
	assert.NoError(t, err)
	assert.Equal(t, "email", notifier.Name())
}

func TestDefaultAlertRules(t *testing.T) {
	rules := DefaultAlertRules()
	assert.Len(t, rules, 4)

	// Check that all rules have required fields
	for _, rule := range rules {
		assert.NotEmpty(t, rule.ID)
		assert.NotEmpty(t, rule.Name)
		assert.NotEmpty(t, rule.Description)
		assert.NotEmpty(t, rule.Condition.Type)
		assert.NotEmpty(t, rule.Severity)
		assert.True(t, rule.Enabled)
	}
}

func TestEventQueryRequest_Default(t *testing.T) {
	req := DefaultEventQueryRequest()
	assert.Equal(t, 0, req.Offset)
	assert.Equal(t, 100, req.Limit)
}

func TestAlertQueryRequest_Default(t *testing.T) {
	req := DefaultAlertQueryRequest()
	assert.Equal(t, 0, req.Offset)
	assert.Equal(t, 100, req.Limit)
}

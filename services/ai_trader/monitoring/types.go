package monitoring

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AuditEvent represents a single audit event in the AI trader system
type AuditEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   EventType              `json:"event_type"`
	Trader      string                 `json:"trader"`
	Symbol      string                 `json:"symbol,omitempty"`
	Action      string                 `json:"action,omitempty"`
	Amount      sdk.Coin               `json:"amount,omitempty"`
	Price       string                 `json:"price,omitempty"`
	TxHash      string                 `json:"tx_hash,omitempty"`
	Success     bool                   `json:"success"`
	Reason      string                 `json:"reason,omitempty"`
	Violations  []string               `json:"violations,omitempty"`
	Warnings    []string               `json:"warnings,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	RiskLevel   RiskLevel              `json:"risk_level"`
	PolicyCheck PolicyCheckResult      `json:"policy_check,omitempty"`
}

// EventType represents the type of audit event
type EventType string

const (
	EventTypeTradeRequest    EventType = "trade_request"
	EventTypeTradeExecuted   EventType = "trade_executed"
	EventTypeTradeRejected   EventType = "trade_rejected"
	EventTypePolicyViolation EventType = "policy_violation"
	EventTypeLimitExceeded   EventType = "limit_exceeded"
	EventTypeEmergencyStop   EventType = "emergency_stop"
	EventTypeConfigChange    EventType = "config_change"
	EventTypeSystemAlert     EventType = "system_alert"
)

// RiskLevel represents the risk level of an event
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// PolicyCheckResult represents the result of policy checks
type PolicyCheckResult struct {
	Allowed    bool          `json:"allowed"`
	Reason     string        `json:"reason,omitempty"`
	Violations []string      `json:"violations,omitempty"`
	Warnings   []string      `json:"warnings,omitempty"`
	CheckedAt  time.Time     `json:"checked_at"`
	Duration   time.Duration `json:"duration"`
}

// Alert represents a system alert
type Alert struct {
	ID             string                 `json:"id"`
	Timestamp      time.Time              `json:"timestamp"`
	AlertType      AlertType              `json:"alert_type"`
	Severity       AlertSeverity          `json:"severity"`
	Title          string                 `json:"title"`
	Message        string                 `json:"message"`
	Trader         string                 `json:"trader,omitempty"`
	Symbol         string                 `json:"symbol,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Acknowledged   bool                   `json:"acknowledged"`
	AcknowledgedAt *time.Time             `json:"acknowledged_at,omitempty"`
	AcknowledgedBy string                 `json:"acknowledged_by,omitempty"`
}

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeLimitExceeded     AlertType = "limit_exceeded"
	AlertTypePolicyViolation   AlertType = "policy_violation"
	AlertTypeEmergencyStop     AlertType = "emergency_stop"
	AlertTypeSystemError       AlertType = "system_error"
	AlertTypeOracleDown        AlertType = "oracle_down"
	AlertTypeHighVolume        AlertType = "high_volume"
	AlertTypeConsecutiveLosses AlertType = "consecutive_losses"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// TradeHistory represents a summary of trading history
type TradeHistory struct {
	Trader           string                 `json:"trader"`
	Symbol           string                 `json:"symbol"`
	TotalTrades      int                    `json:"total_trades"`
	SuccessfulTrades int                    `json:"successful_trades"`
	FailedTrades     int                    `json:"failed_trades"`
	TotalVolume      sdk.Coin               `json:"total_volume"`
	TotalProfit      sdk.Coin               `json:"total_profit"`
	TotalLoss        sdk.Coin               `json:"total_loss"`
	AveragePrice     string                 `json:"average_price"`
	FirstTradeAt     time.Time              `json:"first_trade_at"`
	LastTradeAt      time.Time              `json:"last_trade_at"`
	RiskMetrics      map[string]interface{} `json:"risk_metrics"`
}

// SystemMetrics represents system-wide metrics
type SystemMetrics struct {
	Timestamp        time.Time              `json:"timestamp"`
	ActiveTraders    int                    `json:"active_traders"`
	TotalTrades      int                    `json:"total_trades"`
	TotalVolume      sdk.Coin               `json:"total_volume"`
	ActiveAlerts     int                    `json:"active_alerts"`
	PolicyViolations int                    `json:"policy_violations"`
	EmergencyStops   int                    `json:"emergency_stops"`
	Uptime           time.Duration          `json:"uptime"`
	Metrics          map[string]interface{} `json:"metrics"`
}

// TradeRequest represents a trade request (from risk engine)
type TradeRequest struct {
	Symbol    string    `json:"symbol"`
	Action    string    `json:"action"`
	Amount    sdk.Coin  `json:"amount"`
	Price     string    `json:"price"`
	Timestamp time.Time `json:"timestamp"`
	Trader    string    `json:"trader"`
}

// Helper functions

func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

func calculateRiskLevel(policyResult PolicyCheckResult) RiskLevel {
	if len(policyResult.Violations) > 0 {
		return RiskLevelHigh
	}
	if len(policyResult.Warnings) > 0 {
		return RiskLevelMedium
	}
	return RiskLevelLow
}

func calculateExecutionRiskLevel(success bool, reason string) RiskLevel {
	if !success {
		return RiskLevelHigh
	}
	return RiskLevelLow
}

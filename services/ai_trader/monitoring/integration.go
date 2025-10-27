package monitoring

import (
	"fmt"
	"time"
)

// RiskMonitoringIntegration integrates monitoring with risk management
type RiskMonitoringIntegration struct {
	logger       *Logger
	auditStore   *AuditStore
	alertManager *AlertManager
}

// NewRiskMonitoringIntegration creates a new risk monitoring integration
func NewRiskMonitoringIntegration(logger *Logger, auditStore *AuditStore, alertManager *AlertManager) *RiskMonitoringIntegration {
	return &RiskMonitoringIntegration{
		logger:       logger,
		auditStore:   auditStore,
		alertManager: alertManager,
	}
}

// LogTradeEvaluation logs a trade evaluation from the risk engine
func (rmi *RiskMonitoringIntegration) LogTradeEvaluation(req *TradeRequest, result *PolicyCheckResult) {
	// Create audit event
	event := AuditEvent{
		ID:         generateEventID(),
		Timestamp:  time.Now(),
		EventType:  EventTypeTradeRequest,
		Trader:     req.Trader,
		Symbol:     req.Symbol,
		Action:     req.Action,
		Amount:     req.Amount,
		Price:      req.Price,
		Success:    result.Allowed,
		Reason:     result.Reason,
		Violations: result.Violations,
		Warnings:   result.Warnings,
		RiskLevel:  calculateRiskLevel(*result),
		PolicyCheck: PolicyCheckResult{
			Allowed:    result.Allowed,
			Reason:     result.Reason,
			Violations: result.Violations,
			Warnings:   result.Warnings,
			CheckedAt:  result.CheckedAt,
			Duration:   result.Duration,
		},
	}

	// Log and store event
	rmi.logger.LogAuditEvent(event)
	rmi.auditStore.StoreEvent(event)

	// Check for policy violations that should trigger alerts
	if !result.Allowed && len(result.Violations) > 0 {
		rmi.logPolicyViolationAlert(req, result)
	}

	// Check for warnings that might need attention
	if len(result.Warnings) > 0 {
		rmi.logWarningAlert(req, result)
	}
}

// LogTradeExecution logs a trade execution
func (rmi *RiskMonitoringIntegration) LogTradeExecution(req *TradeRequest, txHash string, success bool, reason string) {
	// Create audit event
	event := AuditEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		EventType: EventTypeTradeExecuted,
		Trader:    req.Trader,
		Symbol:    req.Symbol,
		Action:    req.Action,
		Amount:    req.Amount,
		Price:     req.Price,
		TxHash:    txHash,
		Success:   success,
		Reason:    reason,
		RiskLevel: calculateExecutionRiskLevel(success, reason),
	}

	// Log and store event
	rmi.logger.LogTradeExecution(req.Trader, req.Symbol, req.Action, req.Amount, req.Price, txHash, success, reason)
	rmi.auditStore.StoreEvent(event)

	// Check for execution failures that should trigger alerts
	if !success {
		rmi.logExecutionFailureAlert(req, txHash, reason)
	}
}

// LogEmergencyStop logs an emergency stop event
func (rmi *RiskMonitoringIntegration) LogEmergencyStop(trader, reason string, metadata map[string]interface{}) {
	// Log emergency stop
	rmi.logger.LogEmergencyStop(trader, reason, metadata)

	// Create and store alert
	alert := Alert{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		AlertType: AlertTypeEmergencyStop,
		Severity:  AlertSeverityCritical,
		Title:     "Emergency Stop Triggered",
		Message:   fmt.Sprintf("Emergency stop triggered for trader %s: %s", trader, reason),
		Trader:    trader,
		Metadata:  metadata,
	}

	rmi.alertManager.triggerAlert(alert)
}

// LogLimitExceeded logs a limit exceeded event
func (rmi *RiskMonitoringIntegration) LogLimitExceeded(trader, symbol, limitType string, current, limit interface{}, metadata map[string]interface{}) {
	// Create audit event
	event := AuditEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		EventType: EventTypeLimitExceeded,
		Trader:    trader,
		Symbol:    symbol,
		Reason:    fmt.Sprintf("%s limit exceeded", limitType),
		RiskLevel: RiskLevelHigh,
		Metadata: map[string]interface{}{
			"limit_type": limitType,
			"current":    current,
			"limit":      limit,
		},
	}

	// Add metadata
	for k, v := range metadata {
		event.Metadata[k] = v
	}

	// Log and store event
	rmi.logger.LogAuditEvent(event)
	rmi.auditStore.StoreEvent(event)

	// Create and store alert
	alert := Alert{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		AlertType: AlertTypeLimitExceeded,
		Severity:  AlertSeverityWarning,
		Title:     fmt.Sprintf("%s Limit Exceeded", limitType),
		Message:   fmt.Sprintf("Trader %s exceeded %s limit: %v > %v", trader, limitType, current, limit),
		Trader:    trader,
		Symbol:    symbol,
		Metadata:  event.Metadata,
	}

	rmi.alertManager.triggerAlert(alert)
}

// CheckAlerts checks all alert rules
func (rmi *RiskMonitoringIntegration) CheckAlerts() error {
	return rmi.alertManager.CheckRules()
}

// Helper methods

func (rmi *RiskMonitoringIntegration) logPolicyViolationAlert(req *TradeRequest, result *PolicyCheckResult) {
	alert := Alert{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		AlertType: AlertTypePolicyViolation,
		Severity:  AlertSeverityWarning,
		Title:     "Policy Violation",
		Message:   fmt.Sprintf("Policy violation for trader %s: %s", req.Trader, result.Reason),
		Trader:    req.Trader,
		Symbol:    req.Symbol,
		Metadata: map[string]interface{}{
			"violations": result.Violations,
			"action":     req.Action,
			"amount":     req.Amount.String(),
			"price":      req.Price,
		},
	}

	rmi.alertManager.triggerAlert(alert)
}

func (rmi *RiskMonitoringIntegration) logWarningAlert(req *TradeRequest, result *PolicyCheckResult) {
	alert := Alert{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		AlertType: AlertTypePolicyViolation,
		Severity:  AlertSeverityInfo,
		Title:     "Policy Warning",
		Message:   fmt.Sprintf("Policy warning for trader %s: %s", req.Trader, result.Reason),
		Trader:    req.Trader,
		Symbol:    req.Symbol,
		Metadata: map[string]interface{}{
			"warnings": result.Warnings,
			"action":   req.Action,
			"amount":   req.Amount.String(),
			"price":    req.Price,
		},
	}

	rmi.alertManager.triggerAlert(alert)
}

func (rmi *RiskMonitoringIntegration) logExecutionFailureAlert(req *TradeRequest, txHash, reason string) {
	alert := Alert{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		AlertType: AlertTypeSystemError,
		Severity:  AlertSeverityError,
		Title:     "Trade Execution Failed",
		Message:   fmt.Sprintf("Trade execution failed for trader %s: %s", req.Trader, reason),
		Trader:    req.Trader,
		Symbol:    req.Symbol,
		Metadata: map[string]interface{}{
			"tx_hash": txHash,
			"action":  req.Action,
			"amount":  req.Amount.String(),
			"price":   req.Price,
		},
	}

	rmi.alertManager.triggerAlert(alert)
}

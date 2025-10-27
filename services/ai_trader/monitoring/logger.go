package monitoring

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger provides structured logging for AI trader events
type Logger struct {
	logger     *logrus.Logger
	auditFile  *os.File
	alertFile  *os.File
	mu         sync.RWMutex
	config     LoggerConfig
	buffer     []AuditEvent
	bufferSize int
}

// LoggerConfig contains configuration for the logger
type LoggerConfig struct {
	LogLevel      string        `json:"log_level"`
	LogFormat     string        `json:"log_format"` // json, text
	AuditLogPath  string        `json:"audit_log_path"`
	AlertLogPath  string        `json:"alert_log_path"`
	BufferSize    int           `json:"buffer_size"`
	FlushInterval time.Duration `json:"flush_interval"`
	MaxFileSize   int64         `json:"max_file_size"` // in bytes
	MaxFiles      int           `json:"max_files"`
}

// DefaultLoggerConfig returns default logger configuration
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		LogLevel:      "info",
		LogFormat:     "json",
		AuditLogPath:  "logs/ai_trader_audit.log",
		AlertLogPath:  "logs/ai_trader_alerts.log",
		BufferSize:    1000,
		FlushInterval: 30 * time.Second,
		MaxFileSize:   100 * 1024 * 1024, // 100MB
		MaxFiles:      10,
	}
}

// NewLogger creates a new logger instance
func NewLogger(config LoggerConfig) (*Logger, error) {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %v", err)
	}
	logger.SetLevel(level)

	// Set log format
	if config.LogFormat == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
			FullTimestamp:   true,
		})
	}

	// Create log directories
	if err := os.MkdirAll(filepath.Dir(config.AuditLogPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(config.AlertLogPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create alert log directory: %v", err)
	}

	// Open audit log file
	auditFile, err := os.OpenFile(config.AuditLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %v", err)
	}

	// Open alert log file
	alertFile, err := os.OpenFile(config.AlertLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		auditFile.Close()
		return nil, fmt.Errorf("failed to open alert log file: %v", err)
	}

	l := &Logger{
		logger:     logger,
		auditFile:  auditFile,
		alertFile:  alertFile,
		config:     config,
		buffer:     make([]AuditEvent, 0, config.BufferSize),
		bufferSize: config.BufferSize,
	}

	// Start background flush routine
	go l.flushRoutine()

	return l, nil
}

// LogAuditEvent logs an audit event
func (l *Logger) LogAuditEvent(event AuditEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Add to buffer
	l.buffer = append(l.buffer, event)

	// Log to main logger
	l.logger.WithFields(logrus.Fields{
		"event_type": event.EventType,
		"trader":     event.Trader,
		"symbol":     event.Symbol,
		"action":     event.Action,
		"amount":     event.Amount.String(),
		"price":      event.Price,
		"success":    event.Success,
		"reason":     event.Reason,
		"violations": event.Violations,
		"warnings":   event.Warnings,
		"risk_level": event.RiskLevel,
		"tx_hash":    event.TxHash,
	}).Info("Audit event")

	// Flush if buffer is full
	if len(l.buffer) >= l.bufferSize {
		l.flushBuffer()
	}
}

// LogAlert logs an alert
func (l *Logger) LogAlert(alert Alert) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Log to main logger
	level := logrus.InfoLevel
	switch alert.Severity {
	case AlertSeverityWarning:
		level = logrus.WarnLevel
	case AlertSeverityError:
		level = logrus.ErrorLevel
	case AlertSeverityCritical:
		level = logrus.FatalLevel
	}

	l.logger.WithFields(logrus.Fields{
		"alert_type": alert.AlertType,
		"severity":   alert.Severity,
		"title":      alert.Title,
		"message":    alert.Message,
		"trader":     alert.Trader,
		"symbol":     alert.Symbol,
		"metadata":   alert.Metadata,
	}).Log(level, "Alert")

	// Write to alert log file
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		l.logger.WithError(err).Error("Failed to marshal alert")
		return
	}

	if _, err := l.alertFile.Write(append(alertJSON, '\n')); err != nil {
		l.logger.WithError(err).Error("Failed to write alert to file")
	}
}

// LogTradeRequest logs a trade request
func (l *Logger) LogTradeRequest(trader, symbol, action string, amount interface{}, price string, policyResult PolicyCheckResult) {
	event := AuditEvent{
		ID:          generateEventID(),
		Timestamp:   time.Now(),
		EventType:   EventTypeTradeRequest,
		Trader:      trader,
		Symbol:      symbol,
		Action:      action,
		Price:       price,
		Success:     policyResult.Allowed,
		Reason:      policyResult.Reason,
		Violations:  policyResult.Violations,
		Warnings:    policyResult.Warnings,
		RiskLevel:   calculateRiskLevel(policyResult),
		PolicyCheck: policyResult,
	}

	// Set amount if it's a sdk.Coin
	if coin, ok := amount.(interface{ String() string }); ok {
		event.Metadata = map[string]interface{}{
			"amount": coin.String(),
		}
	}

	l.LogAuditEvent(event)
}

// LogTradeExecution logs a trade execution
func (l *Logger) LogTradeExecution(trader, symbol, action string, amount interface{}, price, txHash string, success bool, reason string) {
	event := AuditEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		EventType: EventTypeTradeExecuted,
		Trader:    trader,
		Symbol:    symbol,
		Action:    action,
		Price:     price,
		TxHash:    txHash,
		Success:   success,
		Reason:    reason,
		RiskLevel: calculateExecutionRiskLevel(success, reason),
	}

	// Set amount if it's a sdk.Coin
	if coin, ok := amount.(interface{ String() string }); ok {
		event.Metadata = map[string]interface{}{
			"amount": coin.String(),
		}
	}

	l.LogAuditEvent(event)
}

// LogPolicyViolation logs a policy violation
func (l *Logger) LogPolicyViolation(trader, symbol, reason string, violations []string) {
	event := AuditEvent{
		ID:         generateEventID(),
		Timestamp:  time.Now(),
		EventType:  EventTypePolicyViolation,
		Trader:     trader,
		Symbol:     symbol,
		Reason:     reason,
		Violations: violations,
		RiskLevel:  RiskLevelHigh,
	}

	l.LogAuditEvent(event)
}

// LogEmergencyStop logs an emergency stop event
func (l *Logger) LogEmergencyStop(trader, reason string, metadata map[string]interface{}) {
	event := AuditEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		EventType: EventTypeEmergencyStop,
		Trader:    trader,
		Reason:    reason,
		RiskLevel: RiskLevelCritical,
		Metadata:  metadata,
	}

	l.LogAuditEvent(event)

	// Also create an alert
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

	l.LogAlert(alert)
}

// Flush flushes the buffer to disk
func (l *Logger) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.flushBuffer()
}

// Close closes the logger and flushes any remaining data
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Flush remaining buffer
	if err := l.flushBuffer(); err != nil {
		return err
	}

	// Close files
	if err := l.auditFile.Close(); err != nil {
		return err
	}
	if err := l.alertFile.Close(); err != nil {
		return err
	}

	return nil
}

// flushBuffer flushes the buffer to the audit log file
func (l *Logger) flushBuffer() error {
	if len(l.buffer) == 0 {
		return nil
	}

	// Write all events to file
	for _, event := range l.buffer {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			l.logger.WithError(err).Error("Failed to marshal audit event")
			continue
		}

		if _, err := l.auditFile.Write(append(eventJSON, '\n')); err != nil {
			l.logger.WithError(err).Error("Failed to write audit event to file")
			return err
		}
	}

	// Clear buffer
	l.buffer = l.buffer[:0]

	return nil
}

// flushRoutine runs in background to periodically flush the buffer
func (l *Logger) flushRoutine() {
	ticker := time.NewTicker(l.config.FlushInterval)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		if len(l.buffer) > 0 {
			l.flushBuffer()
		}
		l.mu.Unlock()
	}
}

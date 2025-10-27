# AI Trader Monitoring and Audit System

This package provides comprehensive monitoring, logging, and alerting capabilities for the AI trader bot system.

## Features

- **Structured Logging**: JSON-formatted audit logs with configurable levels and formats
- **Event Tracking**: Complete audit trail of all trading activities and policy decisions
- **Alert System**: Configurable alerts for policy violations, limit breaches, and system events
- **REST API**: HTTP endpoints for querying events, alerts, and system metrics
- **Risk Integration**: Seamless integration with the risk management engine
- **Multiple Notifiers**: Console, file, webhook, and email notification support

## Components

### 1. Logger (`logger.go`)
- Structured logging with configurable output formats (JSON/Text)
- Buffered logging for performance
- Automatic log rotation and cleanup
- Separate audit and alert log files

### 2. Audit Store (`audit_store.go`)
- In-memory storage for events and alerts
- Query capabilities with filtering and pagination
- Trade history aggregation
- System metrics calculation

### 3. Alert Manager (`alert_manager.go`)
- Rule-based alert system
- Configurable alert conditions
- Multiple notification channels
- Alert acknowledgment support

### 4. REST API (`rest_api.go`)
- HTTP endpoints for monitoring data
- Query parameters for filtering
- JSON responses
- Health check endpoint

### 5. Risk Integration (`integration.go`)
- Integration with risk management engine
- Automatic event logging for policy checks
- Alert generation for violations
- Emergency stop monitoring

## Usage

### Basic Setup

```go
package main

import (
    "github.com/osmosis-labs/osmosis/v30/services/ai_trader/monitoring"
)

func main() {
    // Create logger
    logger, err := monitoring.NewLogger(monitoring.DefaultLoggerConfig())
    if err != nil {
        log.Fatal(err)
    }
    defer logger.Close()

    // Create audit store
    auditStore := monitoring.NewAuditStore(logger, 10000, 1000)

    // Create alert manager
    alertManager := monitoring.NewAlertManager(auditStore, logger)

    // Add notifiers
    consoleNotifier := monitoring.NewConsoleNotifier()
    alertManager.AddNotifier(consoleNotifier)

    // Create REST API
    restAPI := monitoring.NewRESTAPI(auditStore, logger)

    // Start HTTP server
    http.ListenAndServe(":8080", restAPI.GetRouter())
}
```

### Logging Events

```go
// Log a trade request
logger.LogTradeRequest(
    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
    "BTC",
    "buy",
    sdk.NewCoin("factory/test/ndollar", sdk.NewInt(1000)),
    "50000.00",
    policyResult,
)

// Log a policy violation
logger.LogPolicyViolation(
    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
    "BTC",
    "Volume limit exceeded",
    []string{"daily_volume_limits"},
)

// Log an emergency stop
logger.LogEmergencyStop(
    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
    "Consecutive losses limit exceeded",
    map[string]interface{}{
        "consecutive_losses": 5,
        "daily_loss": "1000000",
    },
)
```

### Adding Alert Rules

```go
rule := monitoring.AlertRule{
    ID:          "high_volume_alert",
    Name:        "High Volume Alert",
    Description: "Alert when trading volume exceeds threshold",
    Condition: monitoring.AlertCondition{
        Type:       "volume_threshold",
        Threshold:  1000000, // 1M tokens
        TimeWindow: time.Hour,
    },
    Severity: monitoring.AlertSeverityWarning,
    Enabled:  true,
}

alertManager.AddRule(rule)
```

## REST API Endpoints

### Events
- `GET /api/v1/events` - Get audit events
- `GET /api/v1/events/{id}` - Get specific event

### Alerts
- `GET /api/v1/alerts` - Get alerts
- `POST /api/v1/alerts/{id}/acknowledge` - Acknowledge alert

### History
- `GET /api/v1/history/{trader}` - Get trader history
- `GET /api/v1/history/{trader}/{symbol}` - Get symbol history

### Metrics
- `GET /api/v1/metrics` - Get system metrics
- `GET /api/v1/metrics/trader/{trader}` - Get trader metrics

### Health
- `GET /health` - Health check

### Query Parameters

#### Events
- `trader` - Filter by trader
- `symbol` - Filter by symbol
- `event_type` - Filter by event type
- `risk_level` - Filter by risk level
- `success` - Filter by success status
- `start_time` - Start time (RFC3339)
- `end_time` - End time (RFC3339)
- `offset` - Pagination offset
- `limit` - Pagination limit

#### Alerts
- `trader` - Filter by trader
- `symbol` - Filter by symbol
- `alert_type` - Filter by alert type
- `severity` - Filter by severity
- `acknowledged` - Filter by acknowledgment status
- `start_time` - Start time (RFC3339)
- `end_time` - End time (RFC3339)
- `offset` - Pagination offset
- `limit` - Pagination limit

## Event Types

- `trade_request` - Trade request evaluation
- `trade_executed` - Trade execution result
- `trade_rejected` - Trade rejection
- `policy_violation` - Policy violation
- `limit_exceeded` - Limit exceeded
- `emergency_stop` - Emergency stop triggered
- `config_change` - Configuration change
- `system_alert` - System alert

## Alert Types

- `limit_exceeded` - Limit exceeded
- `policy_violation` - Policy violation
- `emergency_stop` - Emergency stop
- `system_error` - System error
- `oracle_down` - Oracle down
- `high_volume` - High volume
- `consecutive_losses` - Consecutive losses

## Risk Levels

- `low` - Low risk
- `medium` - Medium risk
- `high` - High risk
- `critical` - Critical risk

## Alert Severities

- `info` - Informational
- `warning` - Warning
- `error` - Error
- `critical` - Critical

## Configuration

### Logger Configuration

```go
config := monitoring.LoggerConfig{
    LogLevel:      "info",           // Log level
    LogFormat:     "json",           // Output format (json/text)
    AuditLogPath:  "logs/audit.log", // Audit log file path
    AlertLogPath:  "logs/alerts.log", // Alert log file path
    BufferSize:    1000,             // Buffer size
    FlushInterval: 30 * time.Second, // Flush interval
    MaxFileSize:   100 * 1024 * 1024, // Max file size (100MB)
    MaxFiles:      10,               // Max number of files
}
```

### Default Alert Rules

The system comes with default alert rules:

1. **High Volume Alert** - Triggers when trading volume exceeds threshold
2. **Consecutive Losses Alert** - Triggers when consecutive losses exceed limit
3. **Policy Violations Alert** - Triggers when policy violations exceed threshold
4. **Emergency Stop Alert** - Triggers when emergency stop is activated

## Testing

Run the tests:

```bash
go test ./services/ai_trader/monitoring/... -v
```

Run the example:

```bash
go run ./services/ai_trader/monitoring/example/main.go
```

## Integration with Risk Engine

The monitoring system integrates seamlessly with the risk management engine:

```go
// Create integration
integration := monitoring.NewRiskMonitoringIntegration(logger, auditStore, alertManager)

// Log trade evaluation
integration.LogTradeEvaluation(tradeRequest, policyResult)

// Log trade execution
integration.LogTradeExecution(tradeRequest, txHash, success, reason)

// Log emergency stop
integration.LogEmergencyStop(trader, reason, metadata)
```

## Log Format

### Audit Events (JSON)

```json
{
  "id": "evt_1234567890",
  "timestamp": "2024-01-01T12:00:00Z",
  "event_type": "trade_request",
  "trader": "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
  "symbol": "BTC",
  "action": "buy",
  "amount": "1000factory/test/ndollar",
  "price": "50000.00",
  "success": true,
  "reason": "All checks passed",
  "violations": [],
  "warnings": ["price_deviation"],
  "risk_level": "medium",
  "policy_check": {
    "allowed": true,
    "reason": "All checks passed",
    "violations": [],
    "warnings": ["price_deviation"],
    "checked_at": "2024-01-01T12:00:00Z",
    "duration": "100ms"
  }
}
```

### Alerts (JSON)

```json
{
  "id": "alert_1234567890",
  "timestamp": "2024-01-01T12:00:00Z",
  "alert_type": "limit_exceeded",
  "severity": "warning",
  "title": "Volume Limit Exceeded",
  "message": "Trading volume has exceeded the daily limit",
  "trader": "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
  "symbol": "BTC",
  "metadata": {
    "limit": 1000000,
    "actual": 1500000,
    "threshold": 0.8
  },
  "acknowledged": false
}
```

## Security Considerations

- Log files should be stored securely
- API endpoints should be protected with authentication
- Sensitive data should be masked in logs
- Alert notifications should use secure channels
- Regular log rotation and cleanup

## Performance Considerations

- Use buffered logging for high-frequency events
- Implement log rotation to prevent disk space issues
- Consider using external storage for large-scale deployments
- Monitor memory usage of in-memory audit store
- Use pagination for large result sets


package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/monitoring"
)

func main() {
	// Create logger configuration
	loggerConfig := monitoring.LoggerConfig{
		LogLevel:      "info",
		LogFormat:     "json",
		AuditLogPath:  "logs/ai_trader_audit.log",
		AlertLogPath:  "logs/ai_trader_alerts.log",
		BufferSize:    1000,
		FlushInterval: 30 * time.Second,
		MaxFileSize:   100 * 1024 * 1024, // 100MB
		MaxFiles:      10,
	}

	// Create logger
	logger, err := monitoring.NewLogger(loggerConfig)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Create audit store
	auditStore := monitoring.NewAuditStore(logger, 10000, 1000)

	// Create alert manager
	alertManager := monitoring.NewAlertManager(auditStore, logger)

	// Add default alert rules
	for _, rule := range monitoring.DefaultAlertRules() {
		alertManager.AddRule(rule)
	}

	// Add notifiers
	consoleNotifier := monitoring.NewConsoleNotifier()
	alertManager.AddNotifier(consoleNotifier)

	fileNotifier := monitoring.NewFileNotifier("logs/alerts.log")
	alertManager.AddNotifier(fileNotifier)

	// Create risk monitoring integration
	riskIntegration := monitoring.NewRiskMonitoringIntegration(logger, auditStore, alertManager)

	// Create REST API
	restAPI := monitoring.NewRESTAPI(auditStore, logger)

	// Start HTTP server
	go func() {
		fmt.Println("Starting monitoring server on :8080")
		if err := http.ListenAndServe(":8080", restAPI.GetRouter()); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Simulate some trading activity
	simulateTradingActivity(riskIntegration)

	// Start alert checking routine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := riskIntegration.CheckAlerts(); err != nil {
				log.Printf("Failed to check alerts: %v", err)
			}
		}
	}()

	// Keep the program running
	select {}
}

func simulateTradingActivity(integration *monitoring.RiskMonitoringIntegration) {
	trader := "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz"
	symbol := "BTC"

	// Simulate trade requests
	tradeRequests := []struct {
		action string
		amount sdk.Coin
		price  string
		result monitoring.PolicyCheckResult
	}{
		{
			action: "buy",
			amount: sdk.NewCoin("factory/test/ndollar", math.NewInt(1000)),
			price:  "50000.00",
			result: monitoring.PolicyCheckResult{
				Allowed:    true,
				Reason:     "All checks passed",
				Violations: []string{},
				Warnings:   []string{},
				CheckedAt:  time.Now(),
				Duration:   time.Millisecond * 100,
			},
		},
		{
			action: "sell",
			amount: sdk.NewCoin("factory/test/ndollar", math.NewInt(500)),
			price:  "51000.00",
			result: monitoring.PolicyCheckResult{
				Allowed:    false,
				Reason:     "Cooldown period not met",
				Violations: []string{"cooldown_period"},
				Warnings:   []string{},
				CheckedAt:  time.Now(),
				Duration:   time.Millisecond * 50,
			},
		},
		{
			action: "buy",
			amount: sdk.NewCoin("factory/test/ndollar", math.NewInt(2000)),
			price:  "52000.00",
			result: monitoring.PolicyCheckResult{
				Allowed:    false,
				Reason:     "Volume limit exceeded",
				Violations: []string{"daily_volume_limits"},
				Warnings:   []string{"price_deviation"},
				CheckedAt:  time.Now(),
				Duration:   time.Millisecond * 200,
			},
		},
	}

	for i, req := range tradeRequests {
		// Create trade request
		tradeReq := &monitoring.TradeRequest{
			Symbol:    symbol,
			Action:    req.action,
			Amount:    req.amount,
			Price:     req.price,
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			Trader:    trader,
		}

		// Log trade evaluation
		integration.LogTradeEvaluation(tradeReq, &req.result)

		// If trade was allowed, simulate execution
		if req.result.Allowed {
			txHash := fmt.Sprintf("0x%x", time.Now().UnixNano())
			integration.LogTradeExecution(tradeReq, txHash, true, "Trade executed successfully")
		}

		// Simulate some delay
		time.Sleep(100 * time.Millisecond)
	}

	// Simulate emergency stop
	time.Sleep(2 * time.Second)
	integration.LogEmergencyStop(trader, "Consecutive losses limit exceeded", map[string]interface{}{
		"consecutive_losses": 5,
		"daily_loss":         "1000000",
	})

	// Simulate limit exceeded
	time.Sleep(1 * time.Second)
	integration.LogLimitExceeded(trader, symbol, "daily_volume", 1500000, 1000000, map[string]interface{}{
		"limit_type": "daily_volume",
		"current":    1500000,
		"limit":      1000000,
		"percentage": 150.0,
	})
}

package risk_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/oracle"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/config"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/risk"
)

// MockOracleClient is a mock implementation of oracle.Client
type MockOracleClient struct {
	prices map[string]*oracle.PriceData
}

func (m *MockOracleClient) GetPrice(ctx context.Context, symbol string) (*oracle.PriceData, error) {
	if price, exists := m.prices[symbol]; exists {
		return price, nil
	}
	return nil, fmt.Errorf("price not found for %s", symbol)
}

func (m *MockOracleClient) GetPrices(ctx context.Context, symbols []string) (map[string]*oracle.PriceData, error) {
	result := make(map[string]*oracle.PriceData)
	for _, symbol := range symbols {
		if price, exists := m.prices[symbol]; exists {
			result[symbol] = price
		}
	}
	return result, nil
}

func (m *MockOracleClient) ValidatePrice(price *oracle.PriceData) error {
	if price == nil {
		return fmt.Errorf("price is nil")
	}
	return nil
}

func (m *MockOracleClient) IsPriceStale(price *oracle.PriceData, maxAge time.Duration) bool {
	if price == nil {
		return true
	}
	return time.Since(price.Timestamp) > maxAge
}

func (m *MockOracleClient) Close() error {
	return nil
}

func createTestConfig() *config.Config {
	return &config.Config{
		Bot: config.BotConfig{
			Name:            "test-bot",
			GranterAddress:  "cosmos1granter1234567890abcdefghijklmnopqrstuvwxyz",
			GranteeAddress:  "cosmos1grantee1234567890abcdefghijklmnopqrstuvwxyz",
			ChainID:         "test-chain",
			NodeURL:         "http://localhost:26657",
			TradingInterval: config.DurationString{},
			Enabled:         true,
		},
		Limits: config.TradingLimits{
			MaxDailyVolume:   config.CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(10000000))},
			MaxSingleTrade:   config.CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(1000000))},
			MinSingleTrade:   config.CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(1000))},
			AllowedSymbols:   []string{"BTC", "ETH", "OSMO"},
			MaxTradesPerDay:  100,
			MaxTradesPerHour: 10,
		},
		RiskManagement: config.RiskManagement{
			StopLossPercentage:   0.1,  // 10%
			TakeProfitPercentage: 0.2,  // 20%
			MaxPriceDeviation:    0.05, // 5%
			CoolDownPeriod:       config.NewDurationString(time.Minute * 5),
			EmergencyStop: config.EmergencyStop{
				MaxDailyLoss:         config.CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(5000000))},
				MaxConsecutiveLosses: 5,
				MaxOracleDowntime:    config.NewDurationString(time.Minute * 10),
			},
		},
	}
}

func createMockOracleClient() *MockOracleClient {
	return &MockOracleClient{
		prices: map[string]*oracle.PriceData{
			"BTC": {
				Symbol:     "BTC",
				Value:      "50000.00",
				Source:     "coinbase",
				Timestamp:  time.Now(),
				Confidence: 0.95,
			},
			"ETH": {
				Symbol:     "ETH",
				Value:      "3000.00",
				Source:     "coinbase",
				Timestamp:  time.Now(),
				Confidence: 0.90,
			},
			"OSMO": {
				Symbol:     "OSMO",
				Value:      "1.00",
				Source:     "coinbase",
				Timestamp:  time.Now(),
				Confidence: 0.85,
			},
		},
	}
}

func TestPolicyEngine_EvaluateTrade_ScenarioTable(t *testing.T) {
	cfg := createTestConfig()
	mockOracle := createMockOracleClient()

	// Scenario table for comprehensive testing
	scenarios := []struct {
		name               string
		req                *risk.TradeRequest
		expectedResult     bool
		expectedReason     string
		expectedViolations []string
		setupState         func(*risk.PolicyEngine)
	}{
		{
			name: "valid_buy_trade",
			req: &risk.TradeRequest{
				Symbol:    "BTC",
				Action:    "buy",
				Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(100000)),
				Price:     "50000.00",
				Timestamp: time.Now(),
				Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
			},
			expectedResult:     true,
			expectedReason:     "Trade allowed",
			expectedViolations: []string{},
		},
		{
			name: "invalid_symbol",
			req: &risk.TradeRequest{
				Symbol:    "INVALID",
				Action:    "buy",
				Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(100000)),
				Price:     "50000.00",
				Timestamp: time.Now(),
				Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
			},
			expectedResult:     false,
			expectedReason:     "Trade blocked: 1 violations",
			expectedViolations: []string{"trading_limits: symbol INVALID not in allowed list"},
		},
		{
			name: "amount_too_large",
			req: &risk.TradeRequest{
				Symbol:    "BTC",
				Action:    "buy",
				Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(2000000)), // Exceeds max
				Price:     "50000.00",
				Timestamp: time.Now(),
				Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
			},
			expectedResult:     false,
			expectedReason:     "Trade blocked: 1 violations",
			expectedViolations: []string{"trading_limits: amount 2000000factory/test/ndollar exceeds max single trade 1000000factory/test/ndollar"},
		},
		{
			name: "amount_too_small",
			req: &risk.TradeRequest{
				Symbol:    "BTC",
				Action:    "buy",
				Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(500)), // Below min
				Price:     "50000.00",
				Timestamp: time.Now(),
				Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
			},
			expectedResult:     false,
			expectedReason:     "Trade blocked: 1 violations",
			expectedViolations: []string{"trading_limits: amount 500factory/test/ndollar below min single trade 1000factory/test/ndollar"},
		},
		{
			name: "cooldown_period_not_met",
			req: &risk.TradeRequest{
				Symbol:    "BTC",
				Action:    "buy",
				Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(100000)),
				Price:     "50000.00",
				Timestamp: time.Now(),
				Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
			},
			expectedResult:     false,
			expectedReason:     "Trade blocked: 1 violations",
			expectedViolations: []string{"cooldown_period: cooldown period not met"},
			setupState: func(pe *risk.PolicyEngine) {
				// Record a previous trade to trigger cooldown
				prevReq := &risk.TradeRequest{
					Symbol:    "BTC",
					Action:    "buy",
					Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(50000)),
					Price:     "50000.00",
					Timestamp: time.Now().Add(-2 * time.Minute), // 2 minutes ago
					Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
				}
				pe.RecordTrade(prevReq, true, prevReq.Amount)
			},
		},
		{
			name: "price_deviation_too_high",
			req: &risk.TradeRequest{
				Symbol:    "BTC",
				Action:    "buy",
				Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(100000)),
				Price:     "60000.00", // 20% higher than oracle price (50000)
				Timestamp: time.Now(),
				Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
			},
			expectedResult:     true, // Price deviation is a warning, not a violation
			expectedReason:     "Trade allowed",
			expectedViolations: []string{},
		},
		{
			name: "daily_trades_limit_exceeded",
			req: &risk.TradeRequest{
				Symbol:    "BTC",
				Action:    "buy",
				Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(100000)),
				Price:     "50000.00",
				Timestamp: time.Now(),
				Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
			},
			expectedResult:     false,
			expectedReason:     "Trade blocked: 1 violations",
			expectedViolations: []string{"per_symbol_limits: daily trades count 100 exceeds limit 100"},
			setupState: func(pe *risk.PolicyEngine) {
				// Simulate 100 trades already made today
				for i := 0; i < 100; i++ {
					prevReq := &risk.TradeRequest{
						Symbol:    "BTC",
						Action:    "buy",
						Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(1000)),
						Price:     "50000.00",
						Timestamp: time.Now().Add(-time.Duration(i) * time.Minute),
						Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
					}
					pe.RecordTrade(prevReq, true, prevReq.Amount)
				}
			},
		},
		{
			name: "consecutive_losses_limit_exceeded",
			req: &risk.TradeRequest{
				Symbol:    "BTC",
				Action:    "buy",
				Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(100000)),
				Price:     "50000.00",
				Timestamp: time.Now(),
				Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
			},
			expectedResult:     false,
			expectedReason:     "Emergency stop is active",
			expectedViolations: []string{"emergency_stop"},
			setupState: func(pe *risk.PolicyEngine) {
				// Simulate 5 consecutive losses
				for i := 0; i < 5; i++ {
					prevReq := &risk.TradeRequest{
						Symbol:    "BTC",
						Action:    "buy",
						Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(100000)),
						Price:     "50000.00",
						Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
						Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
					}
					pe.RecordTrade(prevReq, false, prevReq.Amount) // Record as failed trade
				}
			},
		},
		{
			name: "emergency_stop_active",
			req: &risk.TradeRequest{
				Symbol:    "BTC",
				Action:    "buy",
				Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(100000)),
				Price:     "50000.00",
				Timestamp: time.Now(),
				Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
			},
			expectedResult:     false,
			expectedReason:     "Emergency stop is active",
			expectedViolations: []string{"emergency_stop"},
			setupState: func(pe *risk.PolicyEngine) {
				// Trigger emergency stop by exceeding daily loss
				for i := 0; i < 10; i++ {
					prevReq := &risk.TradeRequest{
						Symbol:    "BTC",
						Action:    "sell",
						Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(500000)), // 5M loss per trade
						Price:     "50000.00",
						Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
						Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
					}
					pe.RecordTrade(prevReq, false, prevReq.Amount) // Record as failed trade
				}
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Create fresh engine for each test
			engine := risk.NewPolicyEngine(cfg, mockOracle)

			// Setup state if needed
			if scenario.setupState != nil {
				scenario.setupState(engine)
			}

			// Evaluate trade
			result, err := engine.EvaluateTrade(context.Background(), scenario.req)
			require.NoError(t, err)

			// Verify result
			assert.Equal(t, scenario.expectedResult, result.Allowed, "Expected allowed=%v, got %v", scenario.expectedResult, result.Allowed)
			assert.Equal(t, scenario.expectedReason, result.Reason, "Expected reason=%s, got %s", scenario.expectedReason, result.Reason)

			// Verify violations
			if len(scenario.expectedViolations) > 0 {
				assert.Len(t, result.Violations, len(scenario.expectedViolations), "Expected %d violations, got %d", len(scenario.expectedViolations), len(result.Violations))
				for i, expectedViolation := range scenario.expectedViolations {
					if i < len(result.Violations) {
						assert.Contains(t, result.Violations[i], expectedViolation, "Expected violation to contain '%s', got '%s'", expectedViolation, result.Violations[i])
					}
				}
			} else {
				assert.Empty(t, result.Violations, "Expected no violations, got %v", result.Violations)
			}
		})
	}
}

func TestPolicyEngine_RecordTrade(t *testing.T) {
	cfg := createTestConfig()
	mockOracle := createMockOracleClient()
	engine := risk.NewPolicyEngine(cfg, mockOracle)

	req := &risk.TradeRequest{
		Symbol:    "BTC",
		Action:    "buy",
		Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(100000)),
		Price:     "50000.00",
		Timestamp: time.Now(),
		Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
	}

	// Record successful trade
	engine.RecordTrade(req, true, req.Amount)

	// Verify state was updated
	state := engine.GetState()
	assert.NotNil(t, state.SymbolStats["BTC"])
	assert.Equal(t, 1, state.SymbolStats["BTC"].TradesCount)
	assert.Equal(t, 1, state.SymbolStats["BTC"].DailyTradesCount)
	assert.Equal(t, "50000.00", state.SymbolStats["BTC"].LastPrice)
	assert.Equal(t, 0, state.SymbolStats["BTC"].ConsecutiveLosses) // Should remain 0 for successful trade

	// Record failed trade
	engine.RecordTrade(req, false, req.Amount)

	// Verify consecutive losses increased
	state = engine.GetState()
	assert.Equal(t, 1, state.SymbolStats["BTC"].ConsecutiveLosses)
}

func TestPolicyEngine_ResetEmergencyStop(t *testing.T) {
	cfg := createTestConfig()
	mockOracle := createMockOracleClient()
	engine := risk.NewPolicyEngine(cfg, mockOracle)

	// Trigger emergency stop
	for i := 0; i < 6; i++ {
		req := &risk.TradeRequest{
			Symbol:    "BTC",
			Action:    "sell",
			Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(100000)),
			Price:     "50000.00",
			Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
			Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		}
		engine.RecordTrade(req, false, req.Amount)
	}

	// Verify emergency stop is active
	state := engine.GetState()
	assert.True(t, state.EmergencyStop)

	// Reset emergency stop
	engine.ResetEmergencyStop()

	// Verify emergency stop is reset
	state = engine.GetState()
	assert.False(t, state.EmergencyStop)
}

func TestPolicyEngine_DailyReset(t *testing.T) {
	cfg := createTestConfig()
	mockOracle := createMockOracleClient()
	engine := risk.NewPolicyEngine(cfg, mockOracle)

	// Record some trades
	req := &risk.TradeRequest{
		Symbol:    "BTC",
		Action:    "buy",
		Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(100000)),
		Price:     "50000.00",
		Timestamp: time.Now(),
		Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
	}
	engine.RecordTrade(req, true, req.Amount)

	// Verify initial state
	state := engine.GetState()
	assert.Equal(t, 1, state.SymbolStats["BTC"].DailyTradesCount)
	assert.Equal(t, 1, state.SymbolStats["BTC"].TradesCount)

	// Test that daily reset works by manually calling resetDailyCountersIfNeeded
	// In a real scenario, this would be called automatically when a new day starts
	engine.ResetDailyCounters()

	// Verify daily counters were reset
	state = engine.GetState()
	assert.Equal(t, 0, state.SymbolStats["BTC"].DailyTradesCount) // Should be reset
	assert.Equal(t, 1, state.SymbolStats["BTC"].TradesCount)      // Total should remain
}

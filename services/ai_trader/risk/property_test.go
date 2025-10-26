package risk_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/risk"
)

// Property-based test for volume limits
func TestPolicyEngine_PropertyBasedVolumeLimits(t *testing.T) {
	cfg := createTestConfig()
	mockOracle := createMockOracleClient()
	engine := risk.NewPolicyEngine(cfg, mockOracle)

	// Property: The sum of all allowed trades should never exceed daily volume limit
	maxDailyVolume := cfg.Limits.MaxDailyVolume.Coin.Amount.Int64()
	maxSingleTrade := cfg.Limits.MaxSingleTrade.Coin.Amount.Int64()

	// Use fixed start time to avoid timing issues
	startTime := time.Now()

	// Generate random number of trades (between 1 and 50)
	numTrades := rand.Intn(50) + 1
	totalVolume := int64(0)
	allowedTrades := 0

	t.Logf("Testing with %d random trades", numTrades)

	for i := 0; i < numTrades; i++ {
		// Generate random trade amount
		maxAmount := maxSingleTrade
		if maxDailyVolume-totalVolume < maxSingleTrade {
			maxAmount = maxDailyVolume - totalVolume
		}

		if maxAmount <= 0 {
			break // No more volume available
		}

		// Generate random amount between min and max
		minAmount := cfg.Limits.MinSingleTrade.Coin.Amount.Int64()
		if maxAmount < minAmount {
			break // Can't fit minimum trade
		}

		randomAmount := minAmount + rand.Int63n(maxAmount-minAmount+1)

		req := &risk.TradeRequest{
			Symbol:    "BTC",
			Action:    "buy",
			Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(randomAmount)),
			Price:     "50000.00",
			Timestamp: startTime.Add(time.Duration(i) * time.Hour), // 1 hour between trades to avoid hourly limits
			Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		}

		result, err := engine.EvaluateTrade(context.Background(), req)
		require.NoError(t, err)

		if result.Allowed {
			allowedTrades++
			totalVolume += randomAmount

			// Record the trade
			engine.RecordTrade(req, true, req.Amount)

			// Property: Total volume should never exceed daily limit
			assert.LessOrEqual(t, totalVolume, maxDailyVolume,
				"Total volume %d exceeded daily limit %d after %d trades",
				totalVolume, maxDailyVolume, allowedTrades)
		}
	}

	t.Logf("Completed %d trades, total volume: %d, daily limit: %d",
		allowedTrades, totalVolume, maxDailyVolume)

	// Property: If we have remaining volume, we should be able to make at least one more trade
	if totalVolume < maxDailyVolume {
		remainingVolume := maxDailyVolume - totalVolume
		minTrade := cfg.Limits.MinSingleTrade.Coin.Amount.Int64()

		if remainingVolume >= minTrade {
			req := &risk.TradeRequest{
				Symbol:    "BTC",
				Action:    "buy",
				Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(minTrade)),
				Price:     "50000.00",
				Timestamp: startTime.Add(time.Duration(allowedTrades) * time.Hour), // Avoid hourly limits
				Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
			}

			result, err := engine.EvaluateTrade(context.Background(), req)
			require.NoError(t, err)

			// This trade should be allowed if we have enough remaining volume
			assert.True(t, result.Allowed,
				"Trade with amount %d should be allowed when remaining volume is %d",
				minTrade, remainingVolume)
		}
	}
}

// Property-based test for trade frequency limits
func TestPolicyEngine_PropertyBasedFrequencyLimits(t *testing.T) {
	cfg := createTestConfig()
	mockOracle := createMockOracleClient()
	engine := risk.NewPolicyEngine(cfg, mockOracle)

	maxTradesPerDay := cfg.Limits.MaxTradesPerDay
	// maxTradesPerHour := cfg.Limits.MaxTradesPerHour

	// Property: We should never exceed max trades per day
	tradesToday := 0

	// Generate trades until we hit the limit
	for i := 0; i < maxTradesPerDay+10; i++ { // Try more than the limit
		req := &risk.TradeRequest{
			Symbol:    "BTC",
			Action:    "buy",
			Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(1000)),
			Price:     "50000.00",
			Timestamp: time.Now().Add(time.Duration(i) * 6 * time.Minute), // 6 minutes between trades
			Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		}

		result, err := engine.EvaluateTrade(context.Background(), req)
		require.NoError(t, err)

		if result.Allowed {
			tradesToday++
			engine.RecordTrade(req, true, req.Amount)
		} else {
			// If blocked, it should be due to daily trades limit
			if tradesToday >= maxTradesPerDay {
				assert.Contains(t, result.Violations[0], "daily trades count",
					"Expected daily trades limit violation")
			}
		}
	}

	// Property: We should not exceed maxTradesPerDay allowed trades
	assert.LessOrEqual(t, tradesToday, maxTradesPerDay,
		"Should not exceed %d trades per day, got %d", maxTradesPerDay, tradesToday)

	// Property: We should have made some trades (at least 1)
	assert.Greater(t, tradesToday, 0,
		"Should have made at least 1 trade, got %d", tradesToday)
}

// Property-based test for consecutive losses
func TestPolicyEngine_PropertyBasedConsecutiveLosses(t *testing.T) {
	cfg := createTestConfig()
	mockOracle := createMockOracleClient()
	engine := risk.NewPolicyEngine(cfg, mockOracle)

	maxConsecutiveLosses := cfg.RiskManagement.EmergencyStop.MaxConsecutiveLosses

	// Property: After maxConsecutiveLosses, all trades should be blocked
	for i := 0; i < maxConsecutiveLosses+5; i++ {
		req := &risk.TradeRequest{
			Symbol:    "BTC",
			Action:    "buy",
			Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(1000)),
			Price:     "50000.00",
			Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
			Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		}

		result, err := engine.EvaluateTrade(context.Background(), req)
		require.NoError(t, err)

		if i < maxConsecutiveLosses {
			// First maxConsecutiveLosses trades should be allowed
			assert.True(t, result.Allowed,
				"Trade %d should be allowed before hitting consecutive losses limit", i)
			engine.RecordTrade(req, false, req.Amount) // Record as failed
		} else {
			// After maxConsecutiveLosses, trades should be blocked
			assert.False(t, result.Allowed,
				"Trade %d should be blocked after %d consecutive losses", i, maxConsecutiveLosses)
			assert.Contains(t, result.Violations[0], "emergency_stop",
				"Expected emergency stop due to consecutive losses")
		}
	}
}

// Property-based test for price deviation
func TestPolicyEngine_PropertyBasedPriceDeviation(t *testing.T) {
	cfg := createTestConfig()
	mockOracle := createMockOracleClient()
	engine := risk.NewPolicyEngine(cfg, mockOracle)

	// maxDeviation := cfg.RiskManagement.MaxPriceDeviation
	oraclePrice := 50000.0

	// Use fixed start time to avoid timing issues
	startTime := time.Now()

	// Property: Trades with price deviation within limit should be allowed
	// Trades with price deviation exceeding limit should generate warnings

	testCases := []struct {
		priceMultiplier float64
		expectedAllowed bool
		expectedWarning bool
	}{
		{1.0, true, false},  // Exact oracle price
		{1.01, true, false}, // 1% above
		{0.99, true, false}, // 1% below
		{1.03, true, false}, // 3% above (within 5% limit)
		{0.97, true, false}, // 3% below (within 5% limit)
		{1.05, true, false}, // Exactly at 5% limit
		{0.95, true, false}, // Exactly at -5% limit
		{1.06, true, true},  // 6% above (exceeds limit, warning)
		{0.94, true, true},  // 6% below (exceeds limit, warning)
		{1.10, true, true},  // 10% above (exceeds limit, warning)
		{0.90, true, true},  // 10% below (exceeds limit, warning)
	}

	for i, tc := range testCases {
		price := oraclePrice * tc.priceMultiplier

		req := &risk.TradeRequest{
			Symbol:    "BTC",
			Action:    "buy",
			Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(1000)),
			Price:     fmt.Sprintf("%.2f", price),
			Timestamp: startTime.Add(time.Duration(i) * 6 * time.Minute), // 6 minutes between trades
			Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		}

		result, err := engine.EvaluateTrade(context.Background(), req)
		require.NoError(t, err)

		assert.Equal(t, tc.expectedAllowed, result.Allowed,
			"Trade with price %.2f (%.1f%% deviation) should be allowed=%v",
			price, (tc.priceMultiplier-1)*100, tc.expectedAllowed)

		if tc.expectedWarning {
			assert.NotEmpty(t, result.Warnings,
				"Trade with price %.2f (%.1f%% deviation) should generate warnings",
				price, (tc.priceMultiplier-1)*100)
		}
	}
}

// Property-based test for stop-loss and take-profit
func TestPolicyEngine_PropertyBasedStopLossTakeProfit(t *testing.T) {
	cfg := createTestConfig()
	mockOracle := createMockOracleClient()
	engine := risk.NewPolicyEngine(cfg, mockOracle)

	stopLossPercentage := cfg.RiskManagement.StopLossPercentage
	takeProfitPercentage := cfg.RiskManagement.TakeProfitPercentage
	lastPrice := 50000.0

	// Use fixed start time to avoid timing issues
	startTime := time.Now()

	// First, record a trade to establish last price
	initialReq := &risk.TradeRequest{
		Symbol:    "BTC",
		Action:    "buy",
		Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(1000)),
		Price:     fmt.Sprintf("%.2f", lastPrice),
		Timestamp: startTime.Add(-2 * time.Hour), // 2 hours ago to avoid cooldown
		Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
	}
	engine.RecordTrade(initialReq, true, initialReq.Amount)

	// Property: Trades should be blocked when stop-loss or take-profit is triggered
	testCases := []struct {
		priceMultiplier float64
		expectedAllowed bool
		expectedReason  string
	}{
		{1.0, true, ""},  // Same price
		{1.01, true, ""}, // 1% above
		{0.99, true, ""}, // 1% below
		{1.0 + takeProfitPercentage*0.5, true, ""},             // Half of take-profit
		{1.0 - stopLossPercentage*0.5, true, ""},               // Half of stop-loss
		{1.0 + takeProfitPercentage, false, "take-profit"},     // Exactly take-profit
		{1.0 - stopLossPercentage, false, "stop-loss"},         // Exactly stop-loss
		{1.0 + takeProfitPercentage*1.1, false, "take-profit"}, // Exceeds take-profit
		{1.0 - stopLossPercentage*1.1, false, "stop-loss"},     // Exceeds stop-loss
	}

	for i, tc := range testCases {
		price := lastPrice * tc.priceMultiplier

		req := &risk.TradeRequest{
			Symbol:    "BTC",
			Action:    "buy",
			Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(1000)),
			Price:     fmt.Sprintf("%.2f", price),
			Timestamp: startTime.Add(time.Duration(i) * 6 * time.Minute), // 6 minutes between trades
			Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		}

		result, err := engine.EvaluateTrade(context.Background(), req)
		require.NoError(t, err)

		assert.Equal(t, tc.expectedAllowed, result.Allowed,
			"Trade with price %.2f (%.1f%% change) should be allowed=%v",
			price, (tc.priceMultiplier-1)*100, tc.expectedAllowed)

		if !tc.expectedAllowed && tc.expectedReason != "" {
			assert.Greater(t, len(result.Violations), 0, "Expected violations for price %.2f", price)
			if len(result.Violations) > 0 {
				assert.Contains(t, result.Violations[0], tc.expectedReason,
					"Expected %s violation for price %.2f", tc.expectedReason, price)
			}
		}
	}
}

// Property-based test for cooldown period
func TestPolicyEngine_PropertyBasedCooldownPeriod(t *testing.T) {
	cfg := createTestConfig()
	mockOracle := createMockOracleClient()
	engine := risk.NewPolicyEngine(cfg, mockOracle)

	cooldownDuration, err := cfg.RiskManagement.CoolDownPeriod.ParseDuration()
	require.NoError(t, err)

	// Record initial trade
	initialReq := &risk.TradeRequest{
		Symbol:    "BTC",
		Action:    "buy",
		Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(1000)),
		Price:     "50000.00",
		Timestamp: time.Now(),
		Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
	}
	engine.RecordTrade(initialReq, true, initialReq.Amount)

	// Property: Trades within cooldown period should be blocked
	testCases := []struct {
		timeOffset      time.Duration
		expectedAllowed bool
	}{
		{cooldownDuration / 2, false},      // Half of cooldown period
		{cooldownDuration * 9 / 10, false}, // 90% of cooldown period
		{cooldownDuration, true},           // Exactly cooldown period
		{cooldownDuration * 11 / 10, true}, // 110% of cooldown period
		{cooldownDuration * 2, true},       // Double cooldown period
	}

	for _, tc := range testCases {
		req := &risk.TradeRequest{
			Symbol:    "BTC",
			Action:    "buy",
			Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(1000)),
			Price:     "50000.00",
			Timestamp: time.Now().Add(tc.timeOffset),
			Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
		}

		result, err := engine.EvaluateTrade(context.Background(), req)
		require.NoError(t, err)

		assert.Equal(t, tc.expectedAllowed, result.Allowed,
			"Trade after %v should be allowed=%v (cooldown period: %v)",
			tc.timeOffset, tc.expectedAllowed, cooldownDuration)

		if !tc.expectedAllowed {
			assert.Contains(t, result.Violations[0], "cooldown period",
				"Expected cooldown violation after %v", tc.timeOffset)
		}
	}
}

// Benchmark test for policy evaluation performance
func BenchmarkPolicyEngine_EvaluateTrade(b *testing.B) {
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.EvaluateTrade(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

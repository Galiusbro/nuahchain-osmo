package risk

import (
	"fmt"
	"math/rand"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/config"
)

// PolicyValidator provides additional validation utilities
type PolicyValidator struct {
	config *config.Config
}

// NewPolicyValidator creates a new policy validator
func NewPolicyValidator(cfg *config.Config) *PolicyValidator {
	return &PolicyValidator{
		config: cfg,
	}
}

// ValidateConfig validates the risk management configuration
func (pv *PolicyValidator) ValidateConfig() error {
	limits := pv.config.Limits
	riskMgmt := pv.config.RiskManagement

	// Validate trading limits
	if limits.MaxSingleTrade.Coin.IsLT(limits.MinSingleTrade.Coin) {
		return fmt.Errorf("max single trade must be greater than min single trade")
	}

	if limits.MaxDailyVolume.Coin.IsLT(limits.MaxSingleTrade.Coin) {
		return fmt.Errorf("max daily volume must be greater than max single trade")
	}

	if limits.MaxTradesPerDay < limits.MaxTradesPerHour {
		return fmt.Errorf("max trades per day must be greater than max trades per hour")
	}

	// Validate risk management parameters
	if riskMgmt.StopLossPercentage <= 0 || riskMgmt.StopLossPercentage >= 1 {
		return fmt.Errorf("stop loss percentage must be between 0 and 1")
	}

	if riskMgmt.TakeProfitPercentage <= 0 || riskMgmt.TakeProfitPercentage >= 1 {
		return fmt.Errorf("take profit percentage must be between 0 and 1")
	}

	if riskMgmt.MaxPriceDeviation <= 0 || riskMgmt.MaxPriceDeviation >= 1 {
		return fmt.Errorf("max price deviation must be between 0 and 1")
	}

	// Validate emergency stop parameters
	emergencyStop := riskMgmt.EmergencyStop
	if emergencyStop.MaxConsecutiveLosses <= 0 {
		return fmt.Errorf("max consecutive losses must be positive")
	}

	if emergencyStop.MaxDailyLoss.Coin.IsZero() {
		return fmt.Errorf("max daily loss must be positive")
	}

	return nil
}

// GenerateRandomTradeRequest generates a random trade request for testing
func (pv *PolicyValidator) GenerateRandomTradeRequest() *TradeRequest {
	rand.Seed(time.Now().UnixNano())

	symbols := pv.config.Limits.AllowedSymbols
	if len(symbols) == 0 {
		symbols = []string{"BTC", "ETH", "OSMO"}
	}

	symbol := symbols[rand.Intn(len(symbols))]
	action := []string{"buy", "sell"}[rand.Intn(2)]

	// Generate random amount within limits
	minAmount := pv.config.Limits.MinSingleTrade.Coin.Amount
	maxAmount := pv.config.Limits.MaxSingleTrade.Coin.Amount

	// Generate random amount between min and max
	amountRange := maxAmount.Sub(minAmount)
	randomAmount := minAmount.Add(math.NewInt(rand.Int63n(amountRange.Int64())))

	amount := sdk.NewCoin("factory/test/ndollar", randomAmount)

	// Generate random price
	basePrice := 50000.0 // Base price for BTC
	if symbol == "ETH" {
		basePrice = 3000.0
	} else if symbol == "OSMO" {
		basePrice = 1.0
	}

	priceVariation := 0.1 // ±10% variation
	priceMultiplier := 1.0 + (rand.Float64()-0.5)*2*priceVariation
	price := fmt.Sprintf("%.2f", basePrice*priceMultiplier)

	return &TradeRequest{
		Symbol:    symbol,
		Action:    action,
		Amount:    amount,
		Price:     price,
		Timestamp: time.Now(),
		Trader:    "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz",
	}
}

// CalculateRiskScore calculates a risk score for a trade request
func (pv *PolicyValidator) CalculateRiskScore(req *TradeRequest, state *TradingState) float64 {
	score := 0.0

	// Volume risk (higher volume = higher risk)
	maxVolume := pv.config.Limits.MaxSingleTrade.Coin.Amount
	volumeRatio := float64(req.Amount.Amount.Int64()) / float64(maxVolume.Int64())
	score += volumeRatio * 0.3

	// Frequency risk (more frequent trading = higher risk)
	if stats, exists := state.SymbolStats[req.Symbol]; exists {
		if stats.DailyTradesCount > 0 {
			frequencyRisk := float64(stats.DailyTradesCount) / float64(pv.config.Limits.MaxTradesPerDay)
			score += frequencyRisk * 0.2
		}
	}

	// Consecutive losses risk
	if state.ConsecutiveLosses > 0 {
		lossRisk := float64(state.ConsecutiveLosses) / float64(pv.config.RiskManagement.EmergencyStop.MaxConsecutiveLosses)
		score += lossRisk * 0.3
	}

	// Daily loss risk
	if !state.DailyLoss.IsZero() {
		maxDailyLoss := pv.config.RiskManagement.EmergencyStop.MaxDailyLoss.Coin.Amount
		lossRatio := float64(state.DailyLoss.Amount.Int64()) / float64(maxDailyLoss.Int64())
		score += lossRatio * 0.2
	}

	return score
}

// GetRiskLevel returns the risk level based on the risk score
func (pv *PolicyValidator) GetRiskLevel(score float64) string {
	switch {
	case score < 0.3:
		return "LOW"
	case score < 0.6:
		return "MEDIUM"
	case score < 0.8:
		return "HIGH"
	default:
		return "CRITICAL"
	}
}

// SimulateTradingDay simulates a full trading day with random trades
func (pv *PolicyValidator) SimulateTradingDay(engine *PolicyEngine, numTrades int) (*SimulationResult, error) {
	result := &SimulationResult{
		TotalTrades:   0,
		AllowedTrades: 0,
		BlockedTrades: 0,
		TotalVolume:   sdk.NewCoin("factory/test/ndollar", math.NewInt(0)),
		Violations:    make(map[string]int),
		RiskScores:    make([]float64, 0),
	}

	for i := 0; i < numTrades; i++ {
		req := pv.GenerateRandomTradeRequest()

		// Add some delay between trades
		req.Timestamp = time.Now().Add(time.Duration(i) * time.Minute)

		policyResult, err := engine.EvaluateTrade(nil, req)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate trade %d: %v", i, err)
		}

		result.TotalTrades++

		if policyResult.Allowed {
			result.AllowedTrades++
			if req.Action == "buy" {
				result.TotalVolume = result.TotalVolume.Add(req.Amount)
			}

			// Simulate trade execution (random success/failure)
			success := rand.Float64() > 0.3 // 70% success rate
			engine.RecordTrade(req, success, req.Amount)
		} else {
			result.BlockedTrades++

			// Count violation types
			for _, violation := range policyResult.Violations {
				result.Violations[violation]++
			}
		}

		// Calculate risk score
		state := engine.GetState()
		riskScore := pv.CalculateRiskScore(req, state)
		result.RiskScores = append(result.RiskScores, riskScore)
	}

	return result, nil
}

// SimulationResult represents the result of a trading simulation
type SimulationResult struct {
	TotalTrades   int            `json:"total_trades"`
	AllowedTrades int            `json:"allowed_trades"`
	BlockedTrades int            `json:"blocked_trades"`
	TotalVolume   sdk.Coin       `json:"total_volume"`
	Violations    map[string]int `json:"violations"`
	RiskScores    []float64      `json:"risk_scores"`
	AverageRisk   float64        `json:"average_risk"`
	MaxRisk       float64        `json:"max_risk"`
}

// CalculateStatistics calculates statistics from the simulation result
func (sr *SimulationResult) CalculateStatistics() {
	if len(sr.RiskScores) == 0 {
		return
	}

	sum := 0.0
	max := 0.0

	for _, score := range sr.RiskScores {
		sum += score
		if score > max {
			max = score
		}
	}

	sr.AverageRisk = sum / float64(len(sr.RiskScores))
	sr.MaxRisk = max
}

// PolicyMetrics provides metrics for policy performance
type PolicyMetrics struct {
	TotalEvaluations    int64            `json:"total_evaluations"`
	AllowedTrades       int64            `json:"allowed_trades"`
	BlockedTrades       int64            `json:"blocked_trades"`
	ViolationCounts     map[string]int64 `json:"violation_counts"`
	AverageResponseTime time.Duration    `json:"average_response_time"`
	LastEvaluation      time.Time        `json:"last_evaluation"`
}

// NewPolicyMetrics creates new policy metrics
func NewPolicyMetrics() *PolicyMetrics {
	return &PolicyMetrics{
		ViolationCounts: make(map[string]int64),
	}
}

// RecordEvaluation records an evaluation result
func (pm *PolicyMetrics) RecordEvaluation(allowed bool, violations []string, responseTime time.Duration) {
	pm.TotalEvaluations++
	pm.LastEvaluation = time.Now()

	if allowed {
		pm.AllowedTrades++
	} else {
		pm.BlockedTrades++
		for _, violation := range violations {
			pm.ViolationCounts[violation]++
		}
	}

	// Update average response time
	if pm.TotalEvaluations == 1 {
		pm.AverageResponseTime = responseTime
	} else {
		pm.AverageResponseTime = (pm.AverageResponseTime*time.Duration(pm.TotalEvaluations-1) + responseTime) / time.Duration(pm.TotalEvaluations)
	}
}

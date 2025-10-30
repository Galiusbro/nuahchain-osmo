package risk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	osmomath "github.com/osmosis-labs/osmosis/osmomath"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/oracle"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/trading"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/config"
)

// OracleClient interface for price queries
type OracleClient interface {
	GetPrice(ctx context.Context, symbol string) (*oracle.PriceData, error)
	GetPrices(ctx context.Context, symbols []string) (map[string]*oracle.PriceData, error)
	ValidatePrice(price *oracle.PriceData) error
	IsPriceStale(price *oracle.PriceData, maxAge time.Duration) bool
	Close() error
}

// PolicyEngine represents the risk management policy engine
type PolicyEngine struct {
	config       *config.Config
	oracleClient OracleClient
	state        *TradingState
	mu           sync.RWMutex
}

// TradingState tracks the current trading state for risk management
type TradingState struct {
	// Daily volume tracking
	DailyVolume map[string]sdk.Coin `json:"daily_volume"` // symbol -> volume in NDOLLAR

	// Per-symbol tracking
	SymbolStats map[string]*SymbolStats `json:"symbol_stats"`

	// Last trade timestamps for cool-down
	LastTradeTime map[string]time.Time `json:"last_trade_time"`

	// Consecutive losses tracking
	ConsecutiveLosses int `json:"consecutive_losses"`

	// Daily loss tracking
	DailyLoss sdk.Coin `json:"daily_loss"`

	// Emergency stop flag
	EmergencyStop bool `json:"emergency_stop"`

	// Last reset time for daily counters
	LastReset time.Time `json:"last_reset"`
}

// SymbolStats tracks statistics for a specific symbol
type SymbolStats struct {
	Symbol            string    `json:"symbol"`
	TradesCount       int       `json:"trades_count"`
	DailyTradesCount  int       `json:"daily_trades_count"`
	LastTradeTime     time.Time `json:"last_trade_time"`
	TotalVolume       sdk.Coin  `json:"total_volume"`
	DailyVolume       sdk.Coin  `json:"daily_volume"`
	DailyLoss         sdk.Coin  `json:"daily_loss"`
	ConsecutiveLosses int       `json:"consecutive_losses"`
	AveragePrice      string    `json:"average_price"`
	LastPrice         string    `json:"last_price"`
}

// TradeRequest represents a trade request for policy evaluation
type TradeRequest struct {
	Symbol       string              `json:"symbol"`
	Action       string              `json:"action"` // "buy" or "sell"
	Amount       sdk.Coin            `json:"amount"`
	Price        string              `json:"price"`
	Timestamp    time.Time           `json:"timestamp"`
	Trader       string              `json:"trader"`
	Market       trading.TradeMarket `json:"market,omitempty"`
	PaymentDenom string              `json:"payment_denom,omitempty"`
}

// PolicyResult represents the result of policy evaluation
type PolicyResult struct {
	Allowed    bool     `json:"allowed"`
	Reason     string   `json:"reason"`
	Violations []string `json:"violations"`
	Warnings   []string `json:"warnings"`
}

// NewPolicyEngine creates a new policy engine
func NewPolicyEngine(cfg *config.Config, oracleClient OracleClient) *PolicyEngine {
	return &PolicyEngine{
		config:       cfg,
		oracleClient: oracleClient,
		state: &TradingState{
			DailyVolume:       make(map[string]sdk.Coin),
			SymbolStats:       make(map[string]*SymbolStats),
			LastTradeTime:     make(map[string]time.Time),
			ConsecutiveLosses: 0,
			DailyLoss:         sdk.NewCoin("factory/test/ndollar", math.NewInt(0)),
			EmergencyStop:     false,
			LastReset:         time.Now(),
		},
	}
}

// EvaluateTrade evaluates a trade request against all policies
func (pe *PolicyEngine) EvaluateTrade(ctx context.Context, req *TradeRequest) (*PolicyResult, error) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	// Reset daily counters if needed
	pe.resetDailyCountersIfNeeded(req.Timestamp)

	var violations []string
	var warnings []string

	// 1. Check trading limits
	if err := pe.checkTradingLimits(req); err != nil {
		violations = append(violations, fmt.Sprintf("trading_limits: %v", err))
	}

	// 2. Check per-symbol limits
	if err := pe.checkPerSymbolLimits(req); err != nil {
		violations = append(violations, fmt.Sprintf("per_symbol_limits: %v", err))
	}

	// 3. Check daily volume limits
	if err := pe.checkDailyVolumeLimits(req); err != nil {
		violations = append(violations, fmt.Sprintf("daily_volume_limits: %v", err))
	}

	// 4. Check cool-down period
	if err := pe.checkCoolDownPeriod(req); err != nil {
		violations = append(violations, fmt.Sprintf("cooldown_period: %v", err))
	}

	// 5. Check price deviation
	if err := pe.checkPriceDeviation(ctx, req); err != nil {
		warnings = append(warnings, fmt.Sprintf("price_deviation: %v", err))
	}

	// 6. Check stop-loss and take-profit
	if err := pe.checkStopLossTakeProfit(req); err != nil {
		violations = append(violations, fmt.Sprintf("stop_loss_take_profit: %v", err))
	}

	// 7. Check consecutive losses
	if err := pe.checkConsecutiveLosses(req); err != nil {
		violations = append(violations, fmt.Sprintf("consecutive_losses: %v", err))
	}

	// 8. Check daily loss limits
	if err := pe.checkDailyLossLimits(req); err != nil {
		violations = append(violations, fmt.Sprintf("daily_loss_limits: %v", err))
	}

	// 9. Check emergency stop (only if explicitly triggered)
	if pe.state.EmergencyStop {
		return &PolicyResult{
			Allowed:    false,
			Reason:     "Emergency stop is active",
			Violations: []string{"emergency_stop"},
		}, nil
	}

	allowed := len(violations) == 0
	reason := "Trade allowed"
	if !allowed {
		reason = fmt.Sprintf("Trade blocked: %d violations", len(violations))
	}

	return &PolicyResult{
		Allowed:    allowed,
		Reason:     reason,
		Violations: violations,
		Warnings:   warnings,
	}, nil
}

// RecordTrade records a completed trade for state tracking
func (pe *PolicyEngine) RecordTrade(req *TradeRequest, success bool, actualAmount sdk.Coin) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	// Update symbol stats
	if pe.state.SymbolStats[req.Symbol] == nil {
		pe.state.SymbolStats[req.Symbol] = &SymbolStats{
			Symbol: req.Symbol,
		}
	}

	stats := pe.state.SymbolStats[req.Symbol]
	stats.TradesCount++
	stats.DailyTradesCount++
	stats.LastTradeTime = req.Timestamp
	stats.LastPrice = req.Price

	// Update volumes
	if req.Action == "buy" {
		// Ensure stats.DailyVolume has the correct denomination
		if stats.DailyVolume.Denom == "" {
			stats.DailyVolume = sdk.NewCoin(actualAmount.Denom, math.NewInt(0))
		}
		stats.DailyVolume = stats.DailyVolume.Add(actualAmount)

		currentVolume := pe.state.DailyVolume[req.Symbol]
		if currentVolume.Denom == "" {
			currentVolume = sdk.NewCoin(actualAmount.Denom, math.NewInt(0))
		}
		pe.state.DailyVolume[req.Symbol] = currentVolume.Add(actualAmount)
	}

	// Update last trade time
	pe.state.LastTradeTime[req.Symbol] = req.Timestamp

	// Track consecutive losses
	if !success {
		stats.ConsecutiveLosses++
		// Calculate loss amount (simplified)
		if req.Action == "sell" {
			lossAmount := sdk.NewCoin("factory/test/ndollar", actualAmount.Amount)
			if stats.DailyLoss.Denom == "" {
				stats.DailyLoss = sdk.NewCoin(lossAmount.Denom, math.NewInt(0))
			}
			stats.DailyLoss = stats.DailyLoss.Add(lossAmount)
		}
	} else {
		stats.ConsecutiveLosses = 0
	}

	// Check for emergency stop conditions
	pe.checkEmergencyStopConditions()
}

// GetState returns the current trading state
func (pe *PolicyEngine) GetState() *TradingState {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	// Return a copy to avoid race conditions
	stateCopy := *pe.state
	stateCopy.DailyVolume = make(map[string]sdk.Coin)
	stateCopy.SymbolStats = make(map[string]*SymbolStats)
	stateCopy.LastTradeTime = make(map[string]time.Time)

	for k, v := range pe.state.DailyVolume {
		stateCopy.DailyVolume[k] = v
	}
	for k, v := range pe.state.SymbolStats {
		stateCopy.SymbolStats[k] = v
	}
	for k, v := range pe.state.LastTradeTime {
		stateCopy.LastTradeTime[k] = v
	}

	return &stateCopy
}

// ResetEmergencyStop resets the emergency stop flag
func (pe *PolicyEngine) ResetEmergencyStop() {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.state.EmergencyStop = false
}

// ResetDailyCounters manually resets daily counters (for testing)
func (pe *PolicyEngine) ResetDailyCounters() {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	// Reset daily counters
	pe.state.DailyVolume = make(map[string]sdk.Coin)

	// Reset daily counters for all symbols
	for _, stats := range pe.state.SymbolStats {
		stats.DailyTradesCount = 0
		stats.DailyVolume = sdk.NewCoin("factory/test/ndollar", math.NewInt(0))
		stats.DailyLoss = sdk.NewCoin("factory/test/ndollar", math.NewInt(0))
	}

	pe.state.LastReset = time.Now()
}

// resetDailyCountersIfNeeded resets daily counters if a new day has started
func (pe *PolicyEngine) resetDailyCountersIfNeeded(currentTime time.Time) {
	if currentTime.Truncate(24 * time.Hour).After(pe.state.LastReset.Truncate(24 * time.Hour)) {
		// New day, reset counters
		pe.state.DailyVolume = make(map[string]sdk.Coin)
		pe.state.DailyLoss = sdk.NewCoin("factory/test/ndollar", math.NewInt(0))
		pe.state.ConsecutiveLosses = 0

		// Reset daily counters for all symbols
		for _, stats := range pe.state.SymbolStats {
			stats.DailyTradesCount = 0
			stats.DailyVolume = sdk.NewCoin("factory/test/ndollar", math.NewInt(0))
		}

		pe.state.LastReset = currentTime
	}
}

// checkTradingLimits checks basic trading limits
func (pe *PolicyEngine) checkTradingLimits(req *TradeRequest) error {
	limits := pe.config.Limits

	// Check if symbol is allowed
	allowed := false
	for _, symbol := range limits.AllowedSymbols {
		if symbol == req.Symbol {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("symbol %s not in allowed list", req.Symbol)
	}

	// Check amount limits
	if req.Action == "buy" {
		if req.Amount.IsGTE(limits.MaxSingleTrade.Coin) {
			return fmt.Errorf("amount %s exceeds max single trade %s", req.Amount.String(), limits.MaxSingleTrade.String())
		}
		if req.Amount.IsLT(limits.MinSingleTrade.Coin) {
			return fmt.Errorf("amount %s below min single trade %s", req.Amount.String(), limits.MinSingleTrade.String())
		}
	}

	return nil
}

// checkPerSymbolLimits checks per-symbol trading limits
func (pe *PolicyEngine) checkPerSymbolLimits(req *TradeRequest) error {
	stats := pe.state.SymbolStats[req.Symbol]
	if stats == nil {
		return nil // No previous trades for this symbol
	}

	limits := pe.config.Limits

	// Check daily trades count
	if stats.DailyTradesCount >= limits.MaxTradesPerDay {
		return fmt.Errorf("daily trades count %d exceeds limit %d", stats.DailyTradesCount, limits.MaxTradesPerDay)
	}

	// Check hourly trades count (simplified - using last trade time)
	if !stats.LastTradeTime.IsZero() {
		timeSinceLastTrade := req.Timestamp.Sub(stats.LastTradeTime)
		if timeSinceLastTrade < time.Hour && stats.DailyTradesCount >= limits.MaxTradesPerHour {
			return fmt.Errorf("hourly trades limit exceeded")
		}
	}

	return nil
}

// checkDailyVolumeLimits checks daily volume limits
func (pe *PolicyEngine) checkDailyVolumeLimits(req *TradeRequest) error {
	limits := pe.config.Limits

	// Check total daily volume
	if req.Action == "buy" {
		currentVolume := pe.state.DailyVolume[req.Symbol]
		// Ensure both coins have the same denomination
		if currentVolume.Denom == "" {
			currentVolume = sdk.NewCoin(req.Amount.Denom, math.NewInt(0))
		}
		newVolume := currentVolume.Add(req.Amount)
		if newVolume.IsGTE(limits.MaxDailyVolume.Coin) {
			return fmt.Errorf("daily volume %s would exceed limit %s", newVolume.String(), limits.MaxDailyVolume.String())
		}
	}

	return nil
}

// checkCoolDownPeriod checks cool-down period between trades
func (pe *PolicyEngine) checkCoolDownPeriod(req *TradeRequest) error {
	riskMgmt := pe.config.RiskManagement

	lastTradeTime, exists := pe.state.LastTradeTime[req.Symbol]
	if !exists {
		return nil // No previous trades for this symbol
	}

	cooldownDuration, err := riskMgmt.CoolDownPeriod.ParseDuration()
	if err != nil {
		return fmt.Errorf("invalid cooldown period: %v", err)
	}

	timeSinceLastTrade := req.Timestamp.Sub(lastTradeTime)
	if timeSinceLastTrade < cooldownDuration {
		remaining := cooldownDuration - timeSinceLastTrade
		return fmt.Errorf("cooldown period not met, %s remaining", remaining.String())
	}

	return nil
}

// checkPriceDeviation checks price deviation from oracle
func (pe *PolicyEngine) checkPriceDeviation(ctx context.Context, req *TradeRequest) error {
	riskMgmt := pe.config.RiskManagement

	// Get current price from oracle
	priceData, err := pe.oracleClient.GetPrice(ctx, req.Symbol)
	if err != nil {
		return fmt.Errorf("failed to get oracle price: %v", err)
	}

	// Parse prices
	oraclePrice, err := osmomath.NewDecFromStr(priceData.Value)
	if err != nil {
		return fmt.Errorf("invalid oracle price: %v", err)
	}

	tradePrice, err := osmomath.NewDecFromStr(req.Price)
	if err != nil {
		return fmt.Errorf("invalid trade price: %v", err)
	}

	// Calculate deviation
	deviation := tradePrice.Sub(oraclePrice).Quo(oraclePrice).Abs()
	maxDeviation, err := osmomath.NewDecFromStr(fmt.Sprintf("%.6f", riskMgmt.MaxPriceDeviation))
	if err != nil {
		return fmt.Errorf("invalid max deviation: %v", err)
	}

	if deviation.GT(maxDeviation) {
		return fmt.Errorf("price deviation %.4f exceeds limit %.4f", deviation.MustFloat64(), riskMgmt.MaxPriceDeviation)
	}

	return nil
}

// checkStopLossTakeProfit checks stop-loss and take-profit conditions
func (pe *PolicyEngine) checkStopLossTakeProfit(req *TradeRequest) error {
	riskMgmt := pe.config.RiskManagement
	stats := pe.state.SymbolStats[req.Symbol]

	if stats == nil || stats.LastPrice == "" {
		return nil // No previous price data
	}

	lastPrice, err := osmomath.NewDecFromStr(stats.LastPrice)
	if err != nil {
		return fmt.Errorf("invalid last price: %v", err)
	}

	currentPrice, err := osmomath.NewDecFromStr(req.Price)
	if err != nil {
		return fmt.Errorf("invalid current price: %v", err)
	}

	// Calculate price change percentage
	priceChange := currentPrice.Sub(lastPrice).Quo(lastPrice)

	// Check stop-loss
	stopLossDec, err := osmomath.NewDecFromStr(fmt.Sprintf("%.6f", -riskMgmt.StopLossPercentage))
	if err != nil {
		return fmt.Errorf("invalid stop loss percentage: %v", err)
	}
	if priceChange.LTE(stopLossDec) {
		return fmt.Errorf("stop-loss triggered: price dropped %.4f%%", priceChange.MustFloat64()*100)
	}

	// Check take-profit
	takeProfitDec, err := osmomath.NewDecFromStr(fmt.Sprintf("%.6f", riskMgmt.TakeProfitPercentage))
	if err != nil {
		return fmt.Errorf("invalid take profit percentage: %v", err)
	}
	if priceChange.GTE(takeProfitDec) {
		return fmt.Errorf("take-profit triggered: price increased %.4f%%", priceChange.MustFloat64()*100)
	}

	return nil
}

// checkConsecutiveLosses checks consecutive losses limit
func (pe *PolicyEngine) checkConsecutiveLosses(req *TradeRequest) error {
	emergencyStop := pe.config.RiskManagement.EmergencyStop
	stats := pe.state.SymbolStats[req.Symbol]

	if stats != nil && stats.ConsecutiveLosses >= emergencyStop.MaxConsecutiveLosses {
		return fmt.Errorf("consecutive losses %d exceeds limit %d", stats.ConsecutiveLosses, emergencyStop.MaxConsecutiveLosses)
	}

	return nil
}

// checkDailyLossLimits checks daily loss limits
func (pe *PolicyEngine) checkDailyLossLimits(req *TradeRequest) error {
	emergencyStop := pe.config.RiskManagement.EmergencyStop
	stats := pe.state.SymbolStats[req.Symbol]

	if stats != nil {
		// Ensure DailyLoss has the correct denomination
		if stats.DailyLoss.Denom == "" {
			stats.DailyLoss = sdk.NewCoin(emergencyStop.MaxDailyLoss.Coin.Denom, math.NewInt(0))
		}
		if stats.DailyLoss.IsGTE(emergencyStop.MaxDailyLoss.Coin) {
			return fmt.Errorf("daily loss %s exceeds limit %s", stats.DailyLoss.String(), emergencyStop.MaxDailyLoss.String())
		}
	}

	return nil
}

// checkEmergencyStopConditions checks if emergency stop should be triggered
func (pe *PolicyEngine) checkEmergencyStopConditions() {
	emergencyStop := pe.config.RiskManagement.EmergencyStop

	// Check consecutive losses for any symbol
	for _, stats := range pe.state.SymbolStats {
		if stats.ConsecutiveLosses >= emergencyStop.MaxConsecutiveLosses {
			pe.state.EmergencyStop = true
			return
		}
	}

	// Check daily loss for any symbol
	for _, stats := range pe.state.SymbolStats {
		if stats.DailyLoss.Denom != "" && stats.DailyLoss.IsGTE(emergencyStop.MaxDailyLoss.Coin) {
			pe.state.EmergencyStop = true
			return
		}
	}
}

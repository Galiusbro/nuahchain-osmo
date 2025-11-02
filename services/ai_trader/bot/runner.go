package bot

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/trading"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/config"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/llm"
	md "github.com/osmosis-labs/osmosis/v30/services/ai_trader/marketdata"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/risk"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/shared"
)

// Runner encapsulates the trading loop for a configured bot.
type Runner struct {
	cfg      *config.Config
	market   *md.Service
	decider  *risk.AIDecider
	client   *client.Client
	policy   *risk.PolicyEngine
	interval time.Duration
	symbols  []string
	logf     func(string, ...any)
	mu       sync.Mutex
	closed   bool
}

// NewRunner builds a trading runner using config, market data service, and LLM provider.
func NewRunner(cfg *config.Config, market *md.Service, provider llm.Provider) (*Runner, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if market == nil {
		return nil, fmt.Errorf("market service is required")
	}
	if provider == nil {
		return nil, fmt.Errorf("llm provider is required")
	}
	if !cfg.Bot.Enabled {
		return nil, fmt.Errorf("bot is disabled in configuration")
	}
	interval, err := cfg.Bot.TradingInterval.ParseDuration()
	if err != nil {
		return nil, fmt.Errorf("invalid trading interval: %w", err)
	}
	if interval <= 0 {
		return nil, fmt.Errorf("trading interval must be positive")
	}
	if len(cfg.Limits.AllowedSymbols) == 0 {
		return nil, fmt.Errorf("allowed_symbols must contain at least one entry")
	}
	cli, err := client.NewClient(cfg.Bot.NodeURL)
	if err != nil {
		return nil, err
	}
	decider := risk.NewAIDecider(market, provider)
	if err := applyPerspectiveFromRepo(market, decider, cfg); err != nil {
		return nil, fmt.Errorf("apply perspective: %w", err)
	}
	policy := risk.NewPolicyEngine(cfg, cli.GetOracleClient())
	return &Runner{
		cfg:      cfg,
		market:   market,
		decider:  decider,
		client:   cli,
		policy:   policy,
		interval: interval,
		symbols:  append([]string(nil), cfg.Limits.AllowedSymbols...),
		logf:     log.Printf,
	}, nil
}

// SetLogger overrides the default logger (log.Printf style).
func (r *Runner) SetLogger(fn func(string, ...any)) { r.logf = fn }

// Close releases underlying clients.
func (r *Runner) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	return r.client.Close()
}

// Run starts the trading loop until the context is cancelled.
func (r *Runner) Run(ctx context.Context) error {
	if !r.cfg.Bot.Enabled {
		r.logf("bot %s disabled; exiting", r.cfg.Bot.Name)
		return nil
	}
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	r.logf("bot %s running with interval %s", r.cfg.Bot.Name, r.interval)
	for {
		if err := r.tick(ctx); err != nil {
			r.logf("bot tick error: %v", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (r *Runner) tick(ctx context.Context) error {
	decision, err := r.decider.MakeAIDecision(ctx, r.symbols)
	if err != nil {
		return fmt.Errorf("decision error: %w", err)
	}
	if decision == nil {
		return fmt.Errorf("no decision returned")
	}
	if decision.Action == shared.ActionHold {
		r.logf("decision: HOLD reason=%s", decision.Reason)
		return nil
	}
	if !contains(r.symbols, decision.Symbol) {
		r.logf("decision symbol %s not in allowed set; skipping", decision.Symbol)
		return nil
	}
	coin, err := r.parseAmount(decision)
	if err != nil {
		return fmt.Errorf("parse amount: %w", err)
	}
	price := decision.Price
	if strings.TrimSpace(price) == "" {
		if latest, err := r.market.Latest(ctx, decision.Symbol); err == nil {
			price = latest.Value
		}
	}
	marketType := trading.TradeMarket(decision.Market)
	if marketType == "" {
		marketType = trading.MarketAssets
	}
	req := &risk.TradeRequest{
		Symbol:       decision.Symbol,
		Action:       decision.Action,
		Amount:       coin,
		Price:        price,
		Timestamp:    time.Now().UTC(),
		Trader:       r.cfg.Bot.GranteeAddress,
		Market:       marketType,
		PaymentDenom: decision.PaymentDenom,
	}
	result, err := r.policy.EvaluateTrade(ctx, req)
	if err != nil {
		return fmt.Errorf("policy evaluation: %w", err)
	}
	if !result.Allowed {
		r.logf("policy rejected trade: %s violations=%v", result.Reason, result.Violations)
		r.logAudit("bot.trade_blocked", map[string]any{
			"decision":   decision,
			"violations": result.Violations,
			"warnings":   result.Warnings,
		})
		return nil
	}
	resp, execErr := r.client.ExecuteTradingDecision(ctx, decision, r.cfg.Bot.GranteeAddress, r.cfg.Bot.GranterAddress)
	success := execErr == nil && resp != nil && resp.Success
	if success {
		r.policy.RecordTrade(req, true, coin)
	}
	meta := map[string]any{
		"decision": decision,
		"policy": map[string]any{
			"violations": result.Violations,
			"warnings":   result.Warnings,
		},
		"success": success,
	}
	if resp != nil {
		meta["tx_hash"] = resp.TxHash
		meta["executed_at"] = resp.Timestamp
	}
	if execErr != nil {
		meta["error"] = execErr.Error()
		r.logf("execution error: %v", execErr)
	} else {
		r.logf("executed %s %s %s success=%v", decision.Action, decision.Symbol, decision.Amount, success)
	}
	r.logAudit("bot.trade", meta)
	return nil
}

func (r *Runner) parseAmount(decision *client.TradingDecision) (sdk.Coin, error) {
	denom := decision.PaymentDenom
	if strings.TrimSpace(denom) == "" {
		denom = "factory/test/ndollar"
	}
	amount := strings.TrimSpace(decision.Amount)
	if amount == "" {
		return sdk.Coin{}, fmt.Errorf("empty amount")
	}
	coinStr := amount
	if !strings.Contains(amount, denom) {
		coinStr = fmt.Sprintf("%s%s", amount, denom)
	}
	coin, err := sdk.ParseCoinNormalized(coinStr)
	if err != nil {
		intVal, intErr := sdkmath.NewIntFromString(amount)
		if !intErr || !intVal.IsPositive() {
			return sdk.Coin{}, fmt.Errorf("invalid amount %s: %w", amount, err)
		}
		coin = sdk.NewCoin(denom, intVal)
	}
	return coin, nil
}

func (r *Runner) logAudit(action string, meta map[string]any) {
	repo := r.market.Repository()
	if repo == nil {
		return
	}
	payload, err := json.Marshal(meta)
	if err != nil {
		return
	}
	if err := repo.LogAudit(nil, nil, nil, action, string(payload)); err != nil {
		r.logf("audit log error: %v", err)
	}
}

func contains(arr []string, target string) bool {
	for _, v := range arr {
		if strings.EqualFold(v, target) {
			return true
		}
	}
	return false
}

func applyPerspectiveFromRepo(market *md.Service, decider *risk.AIDecider, cfg *config.Config) error {
	repo := market.Repository()
	if repo == nil {
		return nil
	}
	bot, err := repo.GetBotByName(cfg.Bot.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return nil
	}
	perspective, err := md.ParsePerspective(bot.ConfigJSON)
	if err != nil {
		return err
	}
	decider.SetPerspective(perspective.PreTF, perspective.PreLimit, perspective.TargetTF, perspective.TargetLimit, perspective.PostTF, perspective.PostLimit)
	return nil
}

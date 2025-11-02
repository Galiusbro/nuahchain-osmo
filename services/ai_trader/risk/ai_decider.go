package risk

import (
	"context"
	"encoding/json"
	"time"

	"strings"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/llm"
	md "github.com/osmosis-labs/osmosis/v30/services/ai_trader/marketdata"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/shared"
)

// AIDecider orchestrates data collection and LLM decisioning.
type AIDecider struct {
	Market *md.Service
	LLM    llm.Provider
	// Limits for constraints
	MaxNotionalND string
	MinNotionalND string
	AllowedDenoms []string
	FeeRate       string
	MinConfidence float64

	// Perspective configuration: multi-scale context windows
	PreTF       md.Timeframe // coarser timeframe before target
	TargetTF    md.Timeframe // primary decision timeframe
	PostTF      md.Timeframe // finer timeframe after target
	PreLimit    int          // number of candles for PreTF
	TargetLimit int          // number of candles for TargetTF
	PostLimit   int          // number of candles for PostTF
}

func NewAIDecider(market *md.Service, provider llm.Provider) *AIDecider {
	return &AIDecider{
		Market:        market,
		LLM:           provider,
		MaxNotionalND: "1000000", // defaults; can be overridden by config later
		MinNotionalND: "1000",
		AllowedDenoms: []string{"NDOLLAR", "unuah"},
		FeeRate:       "0.001",
		MinConfidence: 0.55,
		// default perspective: Pre=1h, Target=5m, Post=1m
		PreTF:       md.TF1h,
		TargetTF:    md.TF5m,
		PostTF:      md.TF1m,
		PreLimit:    48,
		TargetLimit: 96,
		PostLimit:   60,
	}
}

// SetPerspective allows the caller to choose the primary timeframe and window sizes.
func (d *AIDecider) SetPerspective(pre md.Timeframe, preLimit int, target md.Timeframe, targetLimit int, post md.Timeframe, postLimit int) {
	if pre != "" {
		d.PreTF = pre
	}
	if target != "" {
		d.TargetTF = target
	}
	if post != "" {
		d.PostTF = post
	}
	if preLimit > 0 {
		d.PreLimit = preLimit
	}
	if targetLimit > 0 {
		d.TargetLimit = targetLimit
	}
	if postLimit > 0 {
		d.PostLimit = postLimit
	}
}

// MakeAIDecision gathers inputs and asks the LLM for a decision, then validates.
func (d *AIDecider) MakeAIDecision(ctx context.Context, symbols []string) (*client.TradingDecision, error) {
	if len(symbols) == 0 {
		return &client.TradingDecision{Action: "hold", Reason: "no symbols"}, nil
	}

	latest := make(map[string]string, len(symbols))
	for _, s := range symbols {
		p, err := d.Market.Latest(ctx, s)
		if err != nil {
			continue
		}
		latest[s] = p.Value
	}

	// Build OHLCV windows across configured timeframes: pre, target, post
	// Keep target timeframe bars separately for PromptInput.OHLCV (flat map)
	ohlcvTarget := map[string][]llm.Candle{}
	for _, s := range symbols {
		fetch := func(tf md.Timeframe, limit int) []llm.Candle {
			bars, err := d.Market.OHLCV(ctx, s, tf, limit)
			if err != nil || len(bars) == 0 {
				return nil
			}
			out := make([]llm.Candle, 0, len(bars))
			for _, b := range bars {
				out = append(out, llm.Candle{T: b.T.Unix(), O: b.O, H: b.H, L: b.L, C: b.C, V: b.V})
			}
			return out
		}
		if tgt := fetch(d.TargetTF, d.TargetLimit); len(tgt) > 0 {
			ohlcvTarget[s] = tgt
		}
		_ = fetch(d.PreTF, d.PreLimit)
		_ = fetch(d.PostTF, d.PostLimit)
	}

	// Compute indicators on target TF and build compact features JSON
	var ctxSummary string
	if len(symbols) > 0 {
		ind, err := d.Market.Indicators(ctx, symbols[0], d.TargetTF, []int{20, 50, 100}, []int{21})
		if err == nil {
			// include a tiny recent slice of closes as sparkline on post TF
			bars, _ := d.Market.OHLCV(ctx, symbols[0], d.PostTF, 30)
			closes := make([]string, 0, len(bars))
			for _, b := range bars {
				closes = append(closes, b.C)
			}
			payload := map[string]any{
				"indicators": ind,
				"closes":     closes,
				"pre_tf":     string(d.PreTF),
				"target_tf":  string(d.TargetTF),
				"post_tf":    string(d.PostTF),
			}
			b, _ := json.Marshal(payload)
			ctxSummary = string(b)
		}
	}

	input := llm.PromptInput{
		Symbols:      symbols,
		NowUTC:       time.Now().UTC(),
		LatestPrices: latest,
		OHLCV:        ohlcvTarget,
		Constraints: llm.Constraints{
			Market:        "assets",
			MaxNotionalND: d.MaxNotionalND,
			MinNotionalND: d.MinNotionalND,
			AllowedDenoms: d.AllowedDenoms,
			FeeRate:       d.FeeRate,
		},
		Context: ctxSummary,
	}

	out, err := d.LLM.GenerateDecision(ctx, input)
	if err != nil {
		return &client.TradingDecision{Action: "hold", Reason: "llm_error"}, nil
	}

	// Persist audit record (prompt + raw response) if repo is available
	// Serialize prompt input
	promptJSONBytes, _ := json.Marshal(input)
	rawContent := ""
	if lr, ok := d.LLM.(interface{ LastRaw() string }); ok {
		rawContent = lr.LastRaw()
	}
	_ = d.Market.SaveDecisionRecord(out.Symbol, out.Action, out.Amount, out.PaymentDenom, out.Market, out.Confidence, out.Rationale, string(promptJSONBytes), rawContent)

	// Basic normalization and validation (extend with real checks later)
	// normalize action to lowercase and allow hold
	switch strings.ToLower(out.Action) {
	case shared.ActionBuy, shared.ActionSell, shared.ActionHold:
		out.Action = strings.ToLower(out.Action)
	default:
		return &client.TradingDecision{Action: "hold", Reason: "invalid_action"}, nil
	}
	// Apply minimum confidence threshold
	if out.Confidence < d.MinConfidence {
		return &client.TradingDecision{Action: "hold", Reason: "low_confidence"}, nil
	}
	if out.Symbol == "" || out.Amount == "" {
		return &client.TradingDecision{Action: "hold", Reason: "incomplete_decision"}, nil
	}
	if out.PaymentDenom == "" {
		out.PaymentDenom = "NDOLLAR"
	}

	return &client.TradingDecision{
		Symbol:       out.Symbol,
		Action:       out.Action,
		Amount:       out.Amount,
		Price:        "", // optional spot snapshot if needed later
		Reason:       out.Rationale,
		Confidence:   float32(out.Confidence),
		Market:       "assets",
		PaymentDenom: out.PaymentDenom,
	}, nil
}

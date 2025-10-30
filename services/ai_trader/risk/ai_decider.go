package risk

import (
	"context"
	"time"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/llm"
	md "github.com/osmosis-labs/osmosis/v30/services/ai_trader/marketdata"
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
}

func NewAIDecider(market *md.Service, provider llm.Provider) *AIDecider {
	return &AIDecider{
		Market:        market,
		LLM:           provider,
		MaxNotionalND: "1000000", // defaults; can be overridden by config later
		MinNotionalND: "1000",
		AllowedDenoms: []string{"NDOLLAR", "unuah"},
		FeeRate:       "0.001",
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

	input := llm.PromptInput{
		Symbols:      symbols,
		NowUTC:       time.Now().UTC(),
		LatestPrices: latest,
		Constraints: llm.Constraints{
			Market:        "assets",
			MaxNotionalND: d.MaxNotionalND,
			MinNotionalND: d.MinNotionalND,
			AllowedDenoms: d.AllowedDenoms,
			FeeRate:       d.FeeRate,
		},
	}

	out, err := d.LLM.GenerateDecision(ctx, input)
	if err != nil {
		return &client.TradingDecision{Action: "hold", Reason: "llm_error"}, nil
	}

	// Basic normalization and validation (extend with real checks later)
	if out.Action != "buy" && out.Action != "sell" {
		return &client.TradingDecision{Action: "hold", Reason: "invalid_action"}, nil
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

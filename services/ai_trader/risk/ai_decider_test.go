package risk

import (
	"context"
	"testing"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/llm"
	md "github.com/osmosis-labs/osmosis/v30/services/ai_trader/marketdata"
)

type stubLLM struct {
	out llm.DecisionOut
	err error
}

func (s stubLLM) GenerateDecision(ctx context.Context, in llm.PromptInput) (llm.DecisionOut, error) {
	return s.out, s.err
}
func (s stubLLM) Name() string { return "stub" }

func TestAIDecider_MakeAIDecision_OK(t *testing.T) {
	market := md.NewService(md.NewYahooFetcher())
	s := stubLLM{out: llm.DecisionOut{Action: "buy", Symbol: "AAPL", Amount: "100000", PaymentDenom: "NDOLLAR", Market: "assets", Confidence: 0.9, Rationale: "test"}}
	d := NewAIDecider(market, s)

	dec, err := d.MakeAIDecision(context.Background(), []string{"AAPL"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if dec.Action != "buy" || dec.Symbol != "AAPL" || dec.Amount != "100000" {
		t.Fatalf("bad decision: %+v", dec)
	}
}

func TestAIDecider_MakeAIDecision_ErrorFallback(t *testing.T) {
	market := md.NewService(md.NewYahooFetcher())
	s := stubLLM{err: assertErr{}}
	d := NewAIDecider(market, s)

	dec, err := d.MakeAIDecision(context.Background(), []string{"AAPL"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if dec.Action != "hold" {
		t.Fatalf("expected hold, got: %+v", dec)
	}
}

type assertErr struct{}

func (assertErr) Error() string { return "boom" }

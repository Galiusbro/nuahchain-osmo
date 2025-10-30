package client_test

import (
	"context"
	"testing"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client"
)

type stubDM struct {
	d   *client.TradingDecision
	err error
}

func (s stubDM) MakeAIDecision(ctx context.Context, symbols []string) (*client.TradingDecision, error) {
	return s.d, s.err
}

func TestDecideAndExecute_HoldSkipsExecution(t *testing.T) {
	c, err := client.NewClient("http://localhost:26657")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer c.Close()

	dm := stubDM{d: &client.TradingDecision{Action: "hold", Reason: "no-op"}}
	res, dec, err := c.DecideAndExecute(context.Background(), dm, []string{"AAPL"}, "g", "r")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !res.Success || dec.Action != "hold" {
		t.Fatalf("unexpected: res=%+v dec=%+v", res, dec)
	}
}

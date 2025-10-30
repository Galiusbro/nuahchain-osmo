package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/llm"
	md "github.com/osmosis-labs/osmosis/v30/services/ai_trader/marketdata"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/risk"
)

// Simple integration example: market data + LLM + decision + execution (hold path will no-op).
func main() {
	// Node URL for execution clients
	nodeURL := "http://localhost:26657"

	// Build market data service (stub Yahoo-like fetcher)
	// Use real Yahoo HTTP fetcher by default
	market := md.NewServiceWithSource("yahoo_http")

	// Build LLM provider (Groq) — key will be read from env GROQ_API_KEY or default
	provider := llm.NewGroq("", "", 12*time.Second)

	// Risk decider using market data and LLM
	decider := risk.NewAIDecider(market, provider)

	// Unified client for execution
	cli, err := client.NewClient(nodeURL)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer cli.Close()

	// Symbols universe to consider
	symbols := []string{"AAPL"}

	// Addresses for authz execution (example placeholders)
	grantee := "nuah1granteeaddress..."
	granter := "nuah1granteraddress..."

	// Decide and (optionally) execute
	res, decision, err := cli.DecideAndExecute(context.Background(), decider, symbols, grantee, granter)
	if err != nil {
		log.Fatalf("DecideAndExecute error: %v", err)
	}

	fmt.Printf("Decision: action=%s symbol=%s amount=%s denom=%s reason=%s\n",
		decision.Action, decision.Symbol, decision.Amount, decision.PaymentDenom, decision.Reason)
	fmt.Printf("Exec result: success=%v at=%s\n", res.Success, res.Timestamp.Format(time.RFC3339))
}

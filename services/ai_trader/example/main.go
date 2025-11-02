package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/bot"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/config"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/llm"
	md "github.com/osmosis-labs/osmosis/v30/services/ai_trader/marketdata"
)

// Simple integration example: market data + trading bot runner (infinite loop).
func main() {
	cfg := config.DefaultConfig()

	if bind := strings.TrimSpace(os.Getenv("AI_TRADER_API_BIND")); bind != "" {
		cfg.API.Bind = bind
	}

	// Build market data service (stub Yahoo-like fetcher)
	// Use real Yahoo HTTP fetcher by default + attach SQLite repo for audit
	market := md.NewServiceWithSource("yahoo_http")
	repo, err := md.NewSQLiteRepository("file:/tmp/ai_trader_md.db?cache=shared&_journal_mode=WAL")
	if err == nil {
		defer repo.Close()
		market = market.WithRepository(repo)
		log.Printf("Repository attached to market service")
	} else {
		log.Printf("WARNING: Failed to create repository: %v", err)
	}

	// Build LLM provider (Groq) — key will be read from env GROQ_API_KEY or default
	provider := llm.NewGroq("", "", 12*time.Second)
	runner, err := bot.NewRunner(cfg, market, provider)
	if err != nil {
		log.Fatalf("failed to create runner: %v", err)
	}
	defer runner.Close()

	// Start minimal REST for market data (for observability)
	mux := http.NewServeMux()
	rateInterval, err := cfg.API.RateInterval.ParseDuration()
	if err != nil {
		rateInterval = time.Minute
	}
	limiter := md.NewRateLimiter(cfg.API.RateLimit, rateInterval)
	log.Printf("Service pointer = %p, repo nil? %v", market, market.Repository() == nil)
	if market.Repository() == nil {
		log.Printf("WARNING: Market service has no repository - REST endpoints requiring DB will fail")
	}
	rest := md.NewREST(market, md.WithRateLimiter(limiter), md.WithCORS(cfg.API.CORSOrigins))
	rest.Register(mux)
	log.Printf("REST API registered on %s", cfg.API.Bind)
	go func() {
		bind := cfg.API.Bind
		cert, key := cfg.API.TLSCertPath, cfg.API.TLSKeyPath
		if strings.TrimSpace(cert) != "" && strings.TrimSpace(key) != "" {
			_ = http.ListenAndServeTLS(bind, cert, key, mux)
		} else {
			_ = http.ListenAndServe(bind, mux)
		}
	}()

	// Start scheduler to backfill and periodically refresh multi-timeframe history
	sch := md.NewScheduler(market, []string{"AAPL"}, []md.Timeframe{md.TF1m, md.TF5m, md.TF1h, md.TF1d}, 15*time.Second)
	sch.Start()
	defer sch.Stop()

	if err := runner.Run(context.Background()); err != nil {
		log.Fatalf("runner stopped: %v", err)
	}
}

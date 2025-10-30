package main

import (
	"context"
	"fmt"
	"time"

	md "github.com/osmosis-labs/osmosis/v30/services/ai_trader/marketdata"
)

func main() {
	repo, err := md.NewSQLiteRepository("file:/tmp/ai_trader_md.db?cache=shared&_journal_mode=WAL")
	if err != nil {
		fmt.Println("sqlite error:", err)
		return
	}
	defer repo.Close()
	svc := md.NewService(md.NewYahooHTTPFetcher(8 * time.Second)).WithRepository(repo)
	p, err := svc.Latest(context.Background(), "AAPL")
	if err != nil {
		fmt.Println("spot error:", err)
		return
	}
	fmt.Printf("spot: symbol=%s price=%s source=%s ts=%d\n", p.Symbol, p.Value, p.Source, p.Timestamp.Unix())

	bars, err := svc.OHLCV(context.Background(), "AAPL", md.TF1m, 5)
	if err != nil {
		fmt.Println("ohlcv error:", err)
		return
	}
	fmt.Printf("ohlcv(1m,5): %d bars, first close=%s last close=%s\n", len(bars), bars[0].C, bars[len(bars)-1].C)
}

package marketdata

import (
	"context"
	"time"
)

// YahooFetcher is a placeholder fetcher that mimics Yahoo-like data source.
// NOTE: Wire actual HTTP calls in production; this stub returns synthetic values for compilation.
type YahooFetcher struct{}

func NewYahooFetcher() *YahooFetcher { return &YahooFetcher{} }

func (y *YahooFetcher) Name() string { return "yahoo" }

func (y *YahooFetcher) GetSpot(ctx context.Context, symbol string) (Price, error) {
	return Price{Symbol: symbol, Value: "100.00", Source: y.Name(), Timestamp: time.Now().UTC()}, nil
}

func (y *YahooFetcher) GetOHLCV(ctx context.Context, symbol string, tf Timeframe, limit int) ([]Candle, error) {
	if limit <= 0 {
		limit = 10
	}
	out := make([]Candle, 0, limit)
	now := time.Now().UTC()
	step := time.Minute
	switch tf {
	case TF5m:
		step = 5 * time.Minute
	case TF1h:
		step = time.Hour
	case TF1d:
		step = 24 * time.Hour
	}
	for i := limit - 1; i >= 0; i-- {
		t := now.Add(-time.Duration(i) * step)
		out = append(out, Candle{T: t, O: "99.5", H: "101.0", L: "98.9", C: "100.0", V: "1000"})
	}
	return out, nil
}

package marketdata

import (
	"context"
	"testing"
	"time"
)

func TestServiceLatestCaches(t *testing.T) {
	svc := NewService(NewYahooFetcher())
	ctx := context.Background()

	p1, err := svc.Latest(ctx, "AAPL")
	if err != nil {
		t.Fatalf("Latest err: %v", err)
	}
	if p1.Symbol != "AAPL" || p1.Value == "" {
		t.Fatalf("unexpected price: %+v", p1)
	}

	// Second call should hit cache (fresh < 30s); timestamps should be equal or very close
	p2, err := svc.Latest(ctx, "AAPL")
	if err != nil {
		t.Fatalf("Latest err2: %v", err)
	}
	if !p2.Timestamp.Equal(p1.Timestamp) {
		t.Fatalf("expected cached timestamp, got %v vs %v", p2.Timestamp, p1.Timestamp)
	}
}

func TestServiceOHLCVProvidesBars(t *testing.T) {
	svc := NewService(NewYahooFetcher())
	ctx := context.Background()

	bars, err := svc.OHLCV(ctx, "AAPL", TF5m, 5)
	if err != nil {
		t.Fatalf("OHLCV err: %v", err)
	}
	if len(bars) != 5 {
		t.Fatalf("expected 5 bars, got %d", len(bars))
	}
	if bars[0].T.IsZero() || bars[0].C == "" {
		t.Fatalf("unexpected first bar: %+v", bars[0])
	}

	// Ensure cached retrieval returns same length quickly
	time.Sleep(5 * time.Millisecond)
	bars2, err := svc.OHLCV(ctx, "AAPL", TF5m, 5)
	if err != nil {
		t.Fatalf("OHLCV err2: %v", err)
	}
	if len(bars2) != 5 {
		t.Fatalf("expected 5 bars again, got %d", len(bars2))
	}
}

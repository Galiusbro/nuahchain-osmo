package marketdata

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fakeFetcher struct {
	spotCalls  int
	ohlcvCalls int
	spot       Price
	ohlcv      map[Timeframe][]Candle
}

func (f *fakeFetcher) Name() string { return "fake" }
func (f *fakeFetcher) GetSpot(ctx context.Context, symbol string) (Price, error) {
	f.spotCalls++
	return f.spot, nil
}
func (f *fakeFetcher) GetOHLCV(ctx context.Context, symbol string, tf Timeframe, limit int) ([]Candle, error) {
	f.ohlcvCalls++
	out := f.ohlcv[tf]
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, nil
}

func tmpDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "md.db")
}

func TestService_OHLCV_UsesRepoIfEnough(t *testing.T) {
	ctx := context.Background()
	f := &fakeFetcher{ohlcv: map[Timeframe][]Candle{}}
	svc := NewService(f)
	dbPath := tmpDB(t)
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("repo: %v", err)
	}
	defer repo.Close()
	svc = svc.WithRepository(repo)

	// Seed repo with 100 1m candles
	now := time.Now().UTC().Add(-100 * time.Minute)
	bars := make([]Candle, 0, 100)
	for i := 0; i < 100; i++ {
		bars = append(bars, Candle{T: now.Add(time.Duration(i) * time.Minute), O: "1", H: "1", L: "1", C: "1", V: ""})
	}
	if err := repo.AppendCandles("AAPL", TF1m, bars); err != nil {
		t.Fatalf("append: %v", err)
	}

	// Request 50 bars: should come from repo (no fetcher call)
	out, err := svc.OHLCV(ctx, "AAPL", TF1m, 50)
	if err != nil {
		t.Fatalf("svc ohlcv: %v", err)
	}
	if len(out) != 50 {
		t.Fatalf("want 50, got %d", len(out))
	}
	if f.ohlcvCalls != 0 {
		t.Fatalf("expected 0 fetch calls, got %d", f.ohlcvCalls)
	}
}

func TestService_OHLCV_FetchesAndSavesWhenMissing(t *testing.T) {
	ctx := context.Background()
	f := &fakeFetcher{ohlcv: map[Timeframe][]Candle{}}
	// Prepare fetcher to return 30 bars for 5m
	now := time.Now().UTC().Add(-150 * time.Minute)
	bars := make([]Candle, 0, 30)
	for i := 0; i < 30; i++ {
		bars = append(bars, Candle{T: now.Add(time.Duration(i) * 5 * time.Minute), O: "2", H: "2", L: "2", C: "2", V: ""})
	}
	f.ohlcv[TF5m] = bars

	svc := NewService(f)
	dbPath := tmpDB(t)
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("repo: %v", err)
	}
	defer func() { repo.Close(); os.Remove(dbPath) }()
	svc = svc.WithRepository(repo)

	// DB empty. Request 20 bars → should fetch once, persist, and return 20
	out, err := svc.OHLCV(ctx, "AAPL", TF5m, 20)
	if err != nil {
		t.Fatalf("svc ohlcv: %v", err)
	}
	if len(out) != 20 {
		t.Fatalf("want 20, got %d", len(out))
	}
	if f.ohlcvCalls != 1 {
		t.Fatalf("expected 1 fetch call, got %d", f.ohlcvCalls)
	}

	// Subsequent request 20 bars → should serve from repo/cache, no new fetch
	out2, err := svc.OHLCV(ctx, "AAPL", TF5m, 20)
	if err != nil {
		t.Fatalf("svc ohlcv 2: %v", err)
	}
	if len(out2) != 20 {
		t.Fatalf("want 20, got %d", len(out2))
	}
	if f.ohlcvCalls != 1 {
		t.Fatalf("expected still 1 fetch call, got %d", f.ohlcvCalls)
	}
}

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

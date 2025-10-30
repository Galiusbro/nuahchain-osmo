package marketdata

import (
	"testing"
	"time"
)

func TestScheduler_Ticks(t *testing.T) {
	svc := NewService(NewYahooFetcher())
	sch := NewScheduler(svc, []string{"AAPL"}, []Timeframe{TF1m}, 50*time.Millisecond)
	sch.Start()
	time.Sleep(160 * time.Millisecond)
	sch.Stop()

	// After a few ticks, latest should have an entry
	if _, ok := svc.store.GetLatest("AAPL"); !ok {
		t.Fatalf("expected cached latest for AAPL")
	}
}

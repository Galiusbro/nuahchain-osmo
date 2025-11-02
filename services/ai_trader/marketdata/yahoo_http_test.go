package marketdata

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

type rtY func(*http.Request) (*http.Response, error)

func (f rtY) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func clientY(fn rtY, timeout time.Duration) *http.Client {
	return &http.Client{Transport: fn, Timeout: timeout}
}

func TestYahooHTTPFetcher_Spot_OK(t *testing.T) {
	y := &YahooHTTPFetcher{client: clientY(rtY(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v8/finance/chart/AAPL" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(stringsReader(`{"chart":{"result":[{"meta":{"regularMarketPrice":267.81},"timestamp":[1],"indicators":{"quote":[{"close":[267.81]}]}}]}}`))}, nil
	}), 2*time.Second), hosts: []string{"http://x"}, userAgents: []string{"test"}}

	p, err := y.GetSpot(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if p.Symbol != "AAPL" || p.Value == "" {
		t.Fatalf("bad price: %+v", p)
	}
}

func TestYahooHTTPFetcher_OHLCV_OK(t *testing.T) {
	y := &YahooHTTPFetcher{client: clientY(rtY(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v8/finance/chart/AAPL" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(stringsReader(`{"chart":{"result":[{"timestamp":[1730246400,1730246460],"indicators":{"quote":[{"open":[1.0,1.1],"high":[1.2,1.3],"low":[0.9,1.0],"close":[1.15,1.2],"volume":[100,120]}]}}]}}`))}, nil
	}), 2*time.Second), hosts: []string{"http://x"}, userAgents: []string{"test"}}

	bars, err := y.GetOHLCV(context.Background(), "AAPL", TF1m, 2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(bars) != 2 || bars[0].C == "" {
		t.Fatalf("bad bars: %+v", bars)
	}
}

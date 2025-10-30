package marketdata

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newTestClient(fn rtFunc, timeout time.Duration) *http.Client {
	return &http.Client{Transport: fn, Timeout: timeout}
}

func TestHTTPFetcher_Spot_OK(t *testing.T) {
	f := &HTTPFetcher{client: newTestClient(rtFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/spot" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(stringsReader(`{"symbol":"AAPL","price":"123.45","source":"test","ts":1730246400}`))}, nil
	}), 2*time.Second), baseURL: "http://x"}

	p, err := f.GetSpot(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if p.Symbol != "AAPL" || p.Value != "123.45" || p.Source != "test" {
		t.Fatalf("bad price: %+v", p)
	}
}

func TestHTTPFetcher_OHLCV_OK(t *testing.T) {
	f := &HTTPFetcher{client: newTestClient(rtFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/ohlcv" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(stringsReader(`[{"t":1730246400,"o":"1","h":"2","l":"0.5","c":"1.5","v":"10"}]`))}, nil
	}), 2*time.Second), baseURL: "http://x"}

	bars, err := f.GetOHLCV(context.Background(), "AAPL", TF1m, 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(bars) != 1 || bars[0].C != "1.5" {
		t.Fatalf("bad bars: %+v", bars)
	}
}

// helper: strings.NewReader with importless alias to keep patch minimal
func stringsReader(s string) *reader { return &reader{data: []byte(s)} }

type reader struct {
	data []byte
	i    int
}

func (r *reader) Read(p []byte) (int, error) {
	if r.i >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.i:])
	r.i += n
	return n, nil
}
func (r *reader) Close() error { return nil }

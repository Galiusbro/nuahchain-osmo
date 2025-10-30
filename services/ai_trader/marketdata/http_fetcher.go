package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPFetcher fetches market data from a Yahoo-like endpoint.
// This is a minimal implementation with retries/backoff and timeouts.
type HTTPFetcher struct {
	client  *http.Client
	baseURL string
	retries int
	backoff time.Duration
}

func NewHTTPFetcher(baseURL string, timeout time.Duration, retries int, backoff time.Duration) *HTTPFetcher {
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	if retries < 0 {
		retries = 0
	}
	if backoff <= 0 {
		backoff = 250 * time.Millisecond
	}
	return &HTTPFetcher{
		client:  &http.Client{Timeout: timeout},
		baseURL: baseURL,
		retries: retries,
		backoff: backoff,
	}
}

type spotResp struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
	Source string `json:"source"`
	Ts     int64  `json:"ts"`
}

func (h *HTTPFetcher) Name() string { return "http" }

func (h *HTTPFetcher) GetSpot(ctx context.Context, symbol string) (Price, error) {
	var lastErr error
	url := fmt.Sprintf("%s/spot?symbol=%s", h.baseURL, symbol)
	for attempt := 0; attempt <= h.retries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return Price{}, err
		}
		resp, err := h.client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(h.backoff)
			continue
		}
		func() { defer resp.Body.Close() }()
		if resp.StatusCode != 200 {
			lastErr = fmt.Errorf("http %d", resp.StatusCode)
			time.Sleep(h.backoff)
			continue
		}
		var sr spotResp
		if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
			return Price{}, err
		}
		ts := time.Unix(sr.Ts, 0).UTC()
		return Price{Symbol: sr.Symbol, Value: sr.Price, Source: sr.Source, Timestamp: ts}, nil
	}
	return Price{}, lastErr
}

type ohlcvResp struct {
	T int64  `json:"t"`
	O string `json:"o"`
	H string `json:"h"`
	L string `json:"l"`
	C string `json:"c"`
	V string `json:"v"`
}

func (h *HTTPFetcher) GetOHLCV(ctx context.Context, symbol string, tf Timeframe, limit int) ([]Candle, error) {
	var lastErr error
	if limit <= 0 {
		limit = 10
	}
	url := fmt.Sprintf("%s/ohlcv?symbol=%s&tf=%s&limit=%d", h.baseURL, symbol, string(tf), limit)
	for attempt := 0; attempt <= h.retries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := h.client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(h.backoff)
			continue
		}
		func() { defer resp.Body.Close() }()
		if resp.StatusCode != 200 {
			lastErr = fmt.Errorf("http %d", resp.StatusCode)
			time.Sleep(h.backoff)
			continue
		}
		var arr []ohlcvResp
		if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
			return nil, err
		}
		out := make([]Candle, 0, len(arr))
		for _, r := range arr {
			out = append(out, Candle{T: time.Unix(r.T, 0).UTC(), O: r.O, H: r.H, L: r.L, C: r.C, V: r.V})
		}
		return out, nil
	}
	return nil, lastErr
}

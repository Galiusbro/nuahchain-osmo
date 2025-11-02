package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// YahooHTTPFetcher calls public Yahoo Finance endpoints.
// Spot:  v7/finance/quote?symbols=SYMB
// OHLCV: v8/finance/chart/SYMB?interval=1m&range=1d  (range will be adjusted by timeframe)
type YahooHTTPFetcher struct {
	client     *http.Client
	hosts      []string
	userAgents []string
	retries    int
	backoff    time.Duration
	rateLimit  time.Duration
	lastAt     time.Time
}

func NewYahooHTTPFetcher(timeout time.Duration) *YahooHTTPFetcher {
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	return &YahooHTTPFetcher{
		client: &http.Client{Timeout: timeout},
		hosts:  []string{"https://query1.finance.yahoo.com", "https://query2.finance.yahoo.com"},
		userAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:91.0) Gecko/20100101 Firefox/91.0",
			"Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15",
		},
		retries:   3,
		backoff:   500 * time.Millisecond,
		rateLimit: 1500 * time.Millisecond,
	}
}

func (y *YahooHTTPFetcher) Name() string { return "yahoo_http" }

// quote response subset
// kept for reference; v7 quote not used currently due to reliability issues
// type yahooQuoteResp struct {
//     QuoteResponse struct {
//         Result []struct {
//             Symbol             string  `json:"symbol"`
//             RegularMarketPrice float64 `json:"regularMarketPrice"`
//         } `json:"result"`
//     } `json:"quoteResponse"`
// }

func (y *YahooHTTPFetcher) GetSpot(ctx context.Context, symbol string) (Price, error) {
	// Use v8 chart endpoint directly for freshest price
	base := y.hosts[rand.Intn(len(y.hosts))]
	u := fmt.Sprintf("%s/v8/finance/chart/%s?interval=1m&range=1d", base, url.PathEscape(symbol))
	var lastErr error
	for i := 0; i <= y.retries; i++ {
		resp, err := y.doGET(ctx, u)
		if err != nil {
			lastErr = err
			time.Sleep(y.jitterBackoff(i))
			continue
		}
		if resp.StatusCode == 429 {
			// Respect Retry-After if present, else exponential backoff with jitter
			ra := resp.Header.Get("Retry-After")
			resp.Body.Close()
			y.sleepRetryAfter(ra, i)
			continue
		}
		if resp.StatusCode != 200 {
			lastErr = fmt.Errorf("yahoo chart http %d", resp.StatusCode)
			resp.Body.Close()
			time.Sleep(y.jitterBackoff(i))
			continue
		}
		var cr yahooChartResp
		if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
			resp.Body.Close()
			return Price{}, err
		}
		resp.Body.Close()
		if len(cr.Chart.Result) == 0 {
			return Price{}, fmt.Errorf("yahoo chart empty")
		}
		r := cr.Chart.Result[0]
		// prefer meta.regularMarketPrice if present, else last close
		val := ""
		if r.Meta.RegularMarketPrice != 0 {
			val = fmt.Sprintf("%f", r.Meta.RegularMarketPrice)
		} else if len(r.Indicators.Quote) > 0 && len(r.Indicators.Quote[0].Close) > 0 {
			last := r.Indicators.Quote[0].Close[len(r.Indicators.Quote[0].Close)-1]
			val = fmt.Sprintf("%f", last)
		}
		if val == "" {
			return Price{}, fmt.Errorf("yahoo chart no price")
		}
		return Price{Symbol: symbol, Value: val, Source: y.Name(), Timestamp: time.Now().UTC()}, nil
	}
	return Price{}, lastErr
}

// chart response subset
type yahooChartResp struct {
	Chart struct {
		Result []struct {
			Meta struct {
				RegularMarketPrice float64 `json:"regularMarketPrice"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []float64 `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

func (y *YahooHTTPFetcher) GetOHLCV(ctx context.Context, symbol string, tf Timeframe, limit int) ([]Candle, error) {
	interval, rng := mapTimeframe(tf)
	if limit <= 0 {
		limit = 50
	}
	// Yahoo returns up to range; we will slice at the end if needed.
	base := y.hosts[rand.Intn(len(y.hosts))]
	u := fmt.Sprintf("%s/v8/finance/chart/%s?interval=%s&range=%s", base, url.PathEscape(symbol), interval, rng)
	var lastErr error
	for i := 0; i <= y.retries; i++ {
		resp, err := y.doGET(ctx, u)
		if err != nil {
			lastErr = err
			time.Sleep(y.jitterBackoff(i))
			continue
		}
		if resp.StatusCode == 429 {
			ra := resp.Header.Get("Retry-After")
			resp.Body.Close()
			y.sleepRetryAfter(ra, i)
			continue
		}
		if resp.StatusCode != 200 {
			lastErr = fmt.Errorf("yahoo chart http %d", resp.StatusCode)
			resp.Body.Close()
			time.Sleep(y.jitterBackoff(i))
			continue
		}
		var cr yahooChartResp
		if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()
		if len(cr.Chart.Result) == 0 {
			return nil, fmt.Errorf("yahoo chart empty")
		}
		r := cr.Chart.Result[0]
		if len(r.Timestamp) == 0 || len(r.Indicators.Quote) == 0 {
			return nil, fmt.Errorf("yahoo chart no data")
		}
		q := r.Indicators.Quote[0]
		n := min(len(r.Timestamp), min(len(q.Open), min(len(q.High), min(len(q.Low), len(q.Close)))))
		out := make([]Candle, 0, n)
		for i := 0; i < n; i++ {
			out = append(out, Candle{
				T: time.Unix(r.Timestamp[i], 0).UTC(),
				O: fmt.Sprintf("%f", q.Open[i]),
				H: fmt.Sprintf("%f", q.High[i]),
				L: fmt.Sprintf("%f", q.Low[i]),
				C: fmt.Sprintf("%f", q.Close[i]),
				V: "",
			})
		}
		if limit < len(out) {
			out = out[len(out)-limit:]
		}
		return out, nil
	}
	return nil, lastErr
}

func mapTimeframe(tf Timeframe) (interval string, rng string) {
	switch tf {
	case TF1m:
		return "1m", "1d"
	case TF5m:
		return "5m", "5d"
	case TF1h:
		// Yahoo prefers 60m for hourly
		return "60m", "1mo"
	case TF1d:
		return "1d", "1y"
	default:
		return "1m", "1d"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// doGET adds headers and basic rate-limiting similar to oracle scraper.
func (y *YahooHTTPFetcher) doGET(ctx context.Context, u string) (*http.Response, error) {
	if since := time.Since(y.lastAt); since < y.rateLimit {
		time.Sleep(y.rateLimit - since)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	// rotate User-Agent per request
	req.Header.Set("User-Agent", y.userAgents[rand.Intn(len(y.userAgents))])
	req.Header.Set("Accept", "application/json,text/plain,*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	resp, err := y.client.Do(req)
	y.lastAt = time.Now()
	return resp, err
}

// jitterBackoff returns exponential backoff with jitter
func (y *YahooHTTPFetcher) jitterBackoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	base := y.backoff * time.Duration(1<<attempt)
	jitter := time.Duration(rand.Int63n(int64(base/2 + 1)))
	return base + jitter
}

// sleepRetryAfter respects Retry-After header when present; falls back to jittered backoff
func (y *YahooHTTPFetcher) sleepRetryAfter(retryAfter string, attempt int) {
	if retryAfter == "" {
		time.Sleep(y.jitterBackoff(attempt))
		return
	}
	if secs, err := time.ParseDuration(retryAfter + "s"); err == nil {
		time.Sleep(secs)
		return
	}
	// Some providers send integer seconds; fallback covered above. If parsing fails, use backoff.
	time.Sleep(y.jitterBackoff(attempt))
}

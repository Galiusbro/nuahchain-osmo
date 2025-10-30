package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// YahooHTTPFetcher calls public Yahoo Finance endpoints.
// Spot:  v7/finance/quote?symbols=SYMB
// OHLCV: v8/finance/chart/SYMB?interval=1m&range=1d  (range will be adjusted by timeframe)
type YahooHTTPFetcher struct {
	client  *http.Client
	baseURL string
}

func NewYahooHTTPFetcher(timeout time.Duration) *YahooHTTPFetcher {
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	return &YahooHTTPFetcher{
		client:  &http.Client{Timeout: timeout},
		baseURL: "https://query1.finance.yahoo.com",
	}
}

func (y *YahooHTTPFetcher) Name() string { return "yahoo_http" }

// quote response subset
type yahooQuoteResp struct {
	QuoteResponse struct {
		Result []struct {
			Symbol             string  `json:"symbol"`
			RegularMarketPrice float64 `json:"regularMarketPrice"`
		} `json:"result"`
	} `json:"quoteResponse"`
}

func (y *YahooHTTPFetcher) GetSpot(ctx context.Context, symbol string) (Price, error) {
	u := fmt.Sprintf("%s/v7/finance/quote?symbols=%s", y.baseURL, url.QueryEscape(symbol))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Price{}, err
	}
	resp, err := y.client.Do(req)
	if err != nil {
		return Price{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return Price{}, fmt.Errorf("yahoo quote http %d", resp.StatusCode)
	}
	var qr yahooQuoteResp
	if err := json.NewDecoder(resp.Body).Decode(&qr); err != nil {
		return Price{}, err
	}
	if len(qr.QuoteResponse.Result) == 0 {
		return Price{}, fmt.Errorf("yahoo quote empty")
	}
	r := qr.QuoteResponse.Result[0]
	return Price{Symbol: r.Symbol, Value: fmt.Sprintf("%f", r.RegularMarketPrice), Source: y.Name(), Timestamp: time.Now().UTC()}, nil
}

// chart response subset
type yahooChartResp struct {
	Chart struct {
		Result []struct {
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
	u := fmt.Sprintf("%s/v8/finance/chart/%s?interval=%s&range=%s", y.baseURL, url.PathEscape(symbol), interval, rng)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := y.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("yahoo chart http %d", resp.StatusCode)
	}
	var cr yahooChartResp
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, err
	}
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
			V: "", // volume optional
		})
	}
	if limit < len(out) {
		out = out[len(out)-limit:]
	}
	return out, nil
}

func mapTimeframe(tf Timeframe) (interval string, rng string) {
	switch tf {
	case TF1m:
		return "1m", "1d"
	case TF5m:
		return "5m", "5d"
	case TF1h:
		return "1h", "1mo"
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

package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// APIClient handles API-based data fetching
type APIClient struct {
	client    *http.Client
	userAgent string
	rateLimit time.Duration
}

// NewAPIClient creates a new API client
func NewAPIClient() *APIClient {
	return &APIClient{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		rateLimit: 1 * time.Second,
	}
}

// YahooFinanceResponse represents Yahoo Finance API response
type YahooFinanceResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				PreviousClose      float64 `json:"previousClose"`
				Currency           string  `json:"currency"`
				ExchangeName       string  `json:"exchangeName"`
				FullExchangeName   string  `json:"fullExchangeName"`
				RegularMarketTime  int64   `json:"regularMarketTime"`
				Timezone           string  `json:"timezone"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"chart"`
}

// PriceData represents price information
type PriceData struct {
	Symbol        string    `json:"symbol"`
	Price         float64   `json:"price"`
	PreviousClose float64   `json:"previousClose"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"changePercent"`
	Currency      string    `json:"currency"`
	Exchange      string    `json:"exchange"`
	Timestamp     time.Time `json:"timestamp"`
	Source        string    `json:"source"`
}

// GetPriceFromYahooFinance fetches price data from Yahoo Finance API
func (c *APIClient) GetPriceFromYahooFinance(symbol string) (*PriceData, error) {
	priceData, err := c.fetchPriceFromYahoo(symbol)
	if err == nil {
		return priceData, nil
	}

	// Attempt automatic resolution via Yahoo search for common 404/no data errors.
	if fallbackSymbol, ok := c.shouldAttemptFallback(symbol, err); ok {
		resolvedPrice, fallbackErr := c.fetchPriceFromYahoo(fallbackSymbol)
		if fallbackErr == nil {
			// Preserve original symbol for downstream consumers but annotate source.
			resolvedPrice.Symbol = symbol
			resolvedPrice.Source = fmt.Sprintf("Yahoo Finance (%s)", fallbackSymbol)
			return resolvedPrice, nil
		}

		// Provide combined context when fallback fetch fails.
		return nil, fmt.Errorf(
			"failed to fetch price for %s (resolved to %s): %w (original error: %v)",
			symbol, fallbackSymbol, fallbackErr, err,
		)
	}

	return nil, err
}

// fetchPriceFromYahoo performs a single Yahoo Finance v8 price request.
func (c *APIClient) fetchPriceFromYahoo(symbol string) (*PriceData, error) {
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s", symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var data YahooFinanceResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(data.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data found for symbol: %s", symbol)
	}

	result := data.Chart.Result[0]
	meta := result.Meta

	// Calculate change and change percent
	change := meta.RegularMarketPrice - meta.PreviousClose
	changePercent := (change / meta.PreviousClose) * 100

	priceData := &PriceData{
		Symbol:        symbol,
		Price:         meta.RegularMarketPrice,
		PreviousClose: meta.PreviousClose,
		Change:        change,
		ChangePercent: changePercent,
		Currency:      meta.Currency,
		Exchange:      meta.FullExchangeName,
		Timestamp:     time.Unix(meta.RegularMarketTime, 0),
		Source:        "Yahoo Finance",
	}

	return priceData, nil
}

// shouldAttemptFallback decides whether we should try resolving the symbol via search.
// It returns the resolved symbol (if found) and a boolean indicating if fallback should proceed.
func (c *APIClient) shouldAttemptFallback(originalSymbol string, fetchErr error) (string, bool) {
	if fetchErr == nil {
		return "", false
	}

	errMsg := fetchErr.Error()
	// Only attempt fallback for 4xx / no data scenarios.
	if !(containsOneOf(errMsg, "HTTP error: 404", "HTTP error: 400", "no data found")) {
		return "", false
	}

	results, err := c.SearchSymbols(originalSymbol)
	if err != nil || len(results) == 0 {
		return "", false
	}

	for _, result := range results {
		rawSymbol, ok := result["symbol"].(string)
		if !ok {
			continue
		}
		candidate := strings.TrimSpace(rawSymbol)
		if candidate == "" {
			continue
		}
		if !strings.EqualFold(candidate, originalSymbol) {
			return candidate, true
		}
	}

	return "", false
}

// containsOneOf checks if the given string contains any of the provided substrings.
func containsOneOf(s string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(s, candidate) {
			return true
		}
	}
	return false
}

// GetHistoricalData fetches historical price data
func (c *APIClient) GetHistoricalData(symbol string, period string) ([]PriceData, error) {
	// Map period to Yahoo Finance range
	rangeMap := map[string]string{
		"1d":  "1d",
		"5d":  "5d",
		"1mo": "1mo",
		"3mo": "3mo",
		"6mo": "6mo",
		"1y":  "1y",
		"2y":  "2y",
		"5y":  "5y",
		"10y": "10y",
		"max": "max",
	}

	rangeStr, exists := rangeMap[period]
	if !exists {
		rangeStr = "1d"
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=%s&interval=1d", symbol, rangeStr)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var data YahooFinanceResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(data.Chart.Result) == 0 {
		return nil, fmt.Errorf("no historical data found for symbol: %s", symbol)
	}

	result := data.Chart.Result[0]
	meta := result.Meta

	var historicalData []PriceData

	if len(result.Indicators.Quote) > 0 {
		quote := result.Indicators.Quote[0]

		for i, timestamp := range result.Timestamp {
			if i < len(quote.Close) && quote.Close[i] > 0 {
				priceData := PriceData{
					Symbol:    symbol,
					Price:     quote.Close[i],
					Currency:  meta.Currency,
					Exchange:  meta.FullExchangeName,
					Timestamp: time.Unix(timestamp, 0),
					Source:    "Yahoo Finance",
				}

				if i < len(quote.Open) {
					priceData.PreviousClose = quote.Open[i]
					priceData.Change = quote.Close[i] - quote.Open[i]
					if quote.Open[i] > 0 {
						priceData.ChangePercent = (priceData.Change / quote.Open[i]) * 100
					}
				}

				historicalData = append(historicalData, priceData)
			}
		}
	}

	return historicalData, nil
}

// SearchSymbols searches for symbols on Yahoo Finance
func (c *APIClient) SearchSymbols(query string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v1/finance/search?q=%s&quotesCount=10", query)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var searchResult struct {
		Quotes []map[string]interface{} `json:"quotes"`
	}

	if err := json.Unmarshal(body, &searchResult); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return searchResult.Quotes, nil
}

// GetDefaultSymbols returns default symbols for different asset types
func GetDefaultSymbols() map[string][]string {
	return map[string][]string{
		"commodities": {
			"GC=F", // Gold
			"SI=F", // Silver
			"CL=F", // Crude Oil
			"NG=F", // Natural Gas
			"HG=F", // Copper
			"PL=F", // Platinum
		},
		"currencies": {
			"EURUSD=X", // EUR/USD
			"GBPUSD=X", // GBP/USD
			"USDJPY=X", // USD/JPY
			"USDCHF=X", // USD/CHF
			"USDCAD=X", // USD/CAD
			"AUDUSD=X", // AUD/USD
		},
		"crypto": {
			"BTC-USD", // Bitcoin
			"ETH-USD", // Ethereum
			"ADA-USD", // Cardano
			"SOL-USD", // Solana
			"DOT-USD", // Polkadot
		},
		"indices": {
			"^GSPC", // S&P 500
			"^DJI",  // Dow Jones
			"^IXIC", // NASDAQ
			"^VIX",  // VIX
		},
		"stocks": {
			"AAPL",  // Apple
			"MSFT",  // Microsoft
			"GOOGL", // Google
			"AMZN",  // Amazon
			"TSLA",  // Tesla
		},
	}
}

// GetRateLimit returns the rate limit duration
func (c *APIClient) GetRateLimit() time.Duration {
	return c.rateLimit
}

// SetRateLimit sets the rate limit duration
func (c *APIClient) SetRateLimit(duration time.Duration) {
	c.rateLimit = duration
}

package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient()

	if client == nil {
		t.Fatal("Expected non-nil APIClient")
	}

	if client.client == nil {
		t.Fatal("Expected non-nil HTTP client")
	}

	if client.userAgent == "" {
		t.Fatal("Expected non-empty user agent")
	}

	if client.rateLimit == 0 {
		t.Fatal("Expected non-zero rate limit")
	}
}

func TestGetPriceFromYahooFinance(t *testing.T) {
	client := NewAPIClient()

	// Test with a well-known stock symbol
	symbol := "AAPL"
	priceData, err := client.GetPriceFromYahooFinance(symbol)

	if err != nil {
		t.Fatalf("Failed to get price data for %s: %v", symbol, err)
	}

	if priceData == nil {
		t.Fatal("Expected non-nil price data")
	}

	if priceData.Symbol != symbol {
		t.Errorf("Expected symbol %s, got %s", symbol, priceData.Symbol)
	}

	if priceData.Price <= 0 {
		t.Errorf("Expected positive price, got %f", priceData.Price)
	}

	if priceData.Currency == "" {
		t.Error("Expected non-empty currency")
	}

	if priceData.Exchange == "" {
		t.Error("Expected non-empty exchange")
	}

	if priceData.Source != "Yahoo Finance" {
		t.Errorf("Expected source 'Yahoo Finance', got %s", priceData.Source)
	}

	// Check if timestamp is recent (within last hour)
	now := time.Now()
	if priceData.Timestamp.After(now) {
		t.Error("Expected timestamp to be in the past")
	}

	if now.Sub(priceData.Timestamp) > time.Hour {
		t.Error("Expected timestamp to be within the last hour")
	}

	t.Logf("Price data for %s: Price=%.2f, Currency=%s, Exchange=%s, Timestamp=%s",
		priceData.Symbol, priceData.Price, priceData.Currency, priceData.Exchange, priceData.Timestamp)
}

func TestGetPriceFromYahooFinanceInvalidSymbol(t *testing.T) {
	client := NewAPIClient()

	// Test with an invalid symbol
	invalidSymbol := "INVALID_SYMBOL_12345"
	_, err := client.GetPriceFromYahooFinance(invalidSymbol)

	if err == nil {
		t.Fatal("Expected error for invalid symbol")
	}

	t.Logf("Expected error for invalid symbol: %v", err)
}

func TestGetHistoricalData(t *testing.T) {
	client := NewAPIClient()

	// Test with a well-known stock symbol and 5-day period
	symbol := "AAPL"
	period := "5d"

	historicalData, err := client.GetHistoricalData(symbol, period)

	if err != nil {
		t.Fatalf("Failed to get historical data for %s: %v", symbol, err)
	}

	if len(historicalData) == 0 {
		t.Fatal("Expected non-empty historical data")
	}

	// Check that we have data for multiple days (should be at least 2-3 days for 5d period)
	if len(historicalData) < 2 {
		t.Errorf("Expected at least 2 data points for 5-day period, got %d", len(historicalData))
	}

	// Check first data point
	firstData := historicalData[0]
	if firstData.Symbol != symbol {
		t.Errorf("Expected symbol %s, got %s", symbol, firstData.Symbol)
	}

	if firstData.Price <= 0 {
		t.Errorf("Expected positive price, got %f", firstData.Price)
	}

	// Check that timestamps are in chronological order
	for i := 1; i < len(historicalData); i++ {
		if historicalData[i].Timestamp.Before(historicalData[i-1].Timestamp) {
			t.Error("Expected timestamps to be in chronological order")
		}
	}

	t.Logf("Historical data for %s (%s): %d data points", symbol, period, len(historicalData))
}

func TestGetHistoricalDataDifferentPeriods(t *testing.T) {
	client := NewAPIClient()
	symbol := "MSFT"

	periods := []string{"1d", "5d", "1mo", "3mo"}

	for _, period := range periods {
		t.Run("period_"+period, func(t *testing.T) {
			historicalData, err := client.GetHistoricalData(symbol, period)

			if err != nil {
				t.Fatalf("Failed to get historical data for %s period %s: %v", symbol, period, err)
			}

			if len(historicalData) == 0 {
				t.Fatal("Expected non-empty historical data")
			}

			t.Logf("Period %s: %d data points", period, len(historicalData))
		})
	}
}

func TestSearchSymbols(t *testing.T) {
	client := NewAPIClient()

	// Test searching for Apple
	query := "Apple"
	results, err := client.SearchSymbols(query)

	if err != nil {
		t.Fatalf("Failed to search symbols for %s: %v", query, err)
	}

	if len(results) == 0 {
		t.Fatal("Expected non-empty search results")
	}

	// Check that we have some results
	if len(results) < 1 {
		t.Errorf("Expected at least 1 search result, got %d", len(results))
	}

	// Check first result has expected fields
	firstResult := results[0]
	if firstResult["symbol"] == nil {
		t.Error("Expected 'symbol' field in search result")
	}

	if firstResult["longname"] == nil {
		t.Error("Expected 'longname' field in search result")
	}

	t.Logf("Search results for '%s': %d results", query, len(results))
	for i, result := range results {
		if i < 3 { // Log first 3 results
			t.Logf("  %d: %v", i+1, result["symbol"])
		}
	}
}

func TestGetDefaultSymbols(t *testing.T) {
	symbols := GetDefaultSymbols()

	// Check that we have all expected categories
	expectedCategories := []string{"commodities", "currencies", "crypto", "indices", "stocks"}
	for _, category := range expectedCategories {
		if _, exists := symbols[category]; !exists {
			t.Errorf("Expected category %s in default symbols", category)
		}
	}

	// Check that each category has symbols
	for category, symbolList := range symbols {
		if len(symbolList) == 0 {
			t.Errorf("Expected non-empty symbol list for category %s", category)
		}

		t.Logf("Category %s: %d symbols", category, len(symbolList))
	}
}

func TestRateLimitSettings(t *testing.T) {
	client := NewAPIClient()

	// Test default rate limit
	defaultRateLimit := client.GetRateLimit()
	if defaultRateLimit != 1*time.Second {
		t.Errorf("Expected default rate limit of 1 second, got %v", defaultRateLimit)
	}

	// Test setting custom rate limit
	customRateLimit := 2 * time.Second
	client.SetRateLimit(customRateLimit)

	if client.GetRateLimit() != customRateLimit {
		t.Errorf("Expected rate limit %v, got %v", customRateLimit, client.GetRateLimit())
	}
}

func TestMultipleSymbols(t *testing.T) {
	client := NewAPIClient()

	// Test multiple well-known symbols
	symbols := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA"}

	for _, symbol := range symbols {
		t.Run("symbol_"+symbol, func(t *testing.T) {
			priceData, err := client.GetPriceFromYahooFinance(symbol)

			if err != nil {
				t.Fatalf("Failed to get price data for %s: %v", symbol, err)
			}

			if priceData.Price <= 0 {
				t.Errorf("Expected positive price for %s, got %f", symbol, priceData.Price)
			}

			t.Logf("%s: $%.2f %s", symbol, priceData.Price, priceData.Currency)

			// Add small delay to respect rate limits
			time.Sleep(client.GetRateLimit())
		})
	}
}

func TestCryptoPrices(t *testing.T) {
	client := NewAPIClient()

	// Test crypto symbols
	cryptoSymbols := []string{"BTC-USD", "ETH-USD", "ADA-USD"}

	for _, symbol := range cryptoSymbols {
		t.Run("crypto_"+symbol, func(t *testing.T) {
			priceData, err := client.GetPriceFromYahooFinance(symbol)

			if err != nil {
				t.Fatalf("Failed to get crypto price data for %s: %v", symbol, err)
			}

			if priceData.Price <= 0 {
				t.Errorf("Expected positive price for %s, got %f", symbol, priceData.Price)
			}

			t.Logf("Crypto %s: $%.2f", symbol, priceData.Price)

			// Add small delay to respect rate limits
			time.Sleep(client.GetRateLimit())
		})
	}
}

func TestCurrencyPairs(t *testing.T) {
	client := NewAPIClient()

	// Test currency pairs
	currencyPairs := []string{"EURUSD=X", "GBPUSD=X", "USDJPY=X"}

	for _, symbol := range currencyPairs {
		t.Run("currency_"+symbol, func(t *testing.T) {
			priceData, err := client.GetPriceFromYahooFinance(symbol)

			if err != nil {
				t.Fatalf("Failed to get currency price data for %s: %v", symbol, err)
			}

			if priceData.Price <= 0 {
				t.Errorf("Expected positive price for %s, got %f", symbol, priceData.Price)
			}

			t.Logf("Currency %s: %.4f", symbol, priceData.Price)

			// Add small delay to respect rate limits
			time.Sleep(client.GetRateLimit())
		})
	}
}

// TestDetailedDataOutput shows detailed data we receive from API
func TestDetailedDataOutput(t *testing.T) {
	client := NewAPIClient()

	// Test with different types of assets
	testCases := []struct {
		name      string
		symbol    string
		assetType string
	}{
		{"Apple Stock", "AAPL", "Stock"},
		{"Bitcoin", "BTC-USD", "Cryptocurrency"},
		{"EUR/USD", "EURUSD=X", "Currency"},
		{"Gold", "GC=F", "Commodity"},
		{"S&P 500", "^GSPC", "Index"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("\n=== Testing %s (%s) ===", tc.name, tc.symbol)

			// Get current price data
			priceData, err := client.GetPriceFromYahooFinance(tc.symbol)
			if err != nil {
				t.Fatalf("Failed to get price data for %s: %v", tc.symbol, err)
			}

			// Display detailed price information
			t.Logf("📊 PRICE DATA:")
			t.Logf("   Symbol: %s", priceData.Symbol)
			t.Logf("   Price: $%.4f", priceData.Price)
			t.Logf("   Previous Close: $%.4f", priceData.PreviousClose)
			t.Logf("   Change: $%.4f", priceData.Change)
			t.Logf("   Change %%: %.2f%%", priceData.ChangePercent)
			t.Logf("   Currency: %s", priceData.Currency)
			t.Logf("   Exchange: %s", priceData.Exchange)
			t.Logf("   Timestamp: %s", priceData.Timestamp.Format("2006-01-02 15:04:05 MST"))
			t.Logf("   Source: %s", priceData.Source)

			// Get historical data for 5 days
			t.Logf("\n📈 HISTORICAL DATA (5 days):")
			historicalData, err := client.GetHistoricalData(tc.symbol, "5d")
			if err != nil {
				t.Logf("   Failed to get historical data: %v", err)
			} else {
				t.Logf("   Total data points: %d", len(historicalData))

				// Show first 3 and last 3 data points
				for i, data := range historicalData {
					if i < 3 || i >= len(historicalData)-3 {
						t.Logf("   [%d] %s: $%.4f (Change: %.2f%%)",
							i+1,
							data.Timestamp.Format("01-02"),
							data.Price,
							data.ChangePercent)
					} else if i == 3 {
						t.Logf("   ... (%d more data points)", len(historicalData)-6)
					}
				}

				// Calculate some statistics
				if len(historicalData) > 0 {
					firstPrice := historicalData[0].Price
					lastPrice := historicalData[len(historicalData)-1].Price
					totalChange := ((lastPrice - firstPrice) / firstPrice) * 100
					t.Logf("   Period change: %.2f%% (from $%.4f to $%.4f)",
						totalChange, firstPrice, lastPrice)
				}
			}

			// Search for related symbols
			t.Logf("\n🔍 SYMBOL SEARCH:")
			searchQuery := tc.name
			if tc.assetType == "Stock" {
				searchQuery = "Apple" // Use Apple for stock search
			} else if tc.assetType == "Cryptocurrency" {
				searchQuery = "Bitcoin"
			}

			searchResults, err := client.SearchSymbols(searchQuery)
			if err != nil {
				t.Logf("   Failed to search symbols: %v", err)
			} else {
				t.Logf("   Found %d related symbols for '%s':", len(searchResults), searchQuery)
				for i, result := range searchResults {
					if i < 5 { // Show first 5 results
						symbol := result["symbol"]
						longname := result["longname"]
						exchange := result["exchange"]
						t.Logf("   [%d] %s - %s (%s)", i+1, symbol, longname, exchange)
					}
				}
			}

            t.Logf("\n%s", strings.Repeat("=", 60))

			// Add delay between requests
			time.Sleep(client.GetRateLimit())
		})
	}
}

// TestRawAPIData shows raw JSON response structure
func TestRawAPIData(t *testing.T) {
	client := NewAPIClient()
	symbol := "AAPL"

	t.Logf("🔍 RAW API DATA for %s:", symbol)

	// Make direct HTTP request to see raw response
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s", symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", client.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := client.client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Parse and display raw data structure
	var rawData map[string]interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	t.Logf("📋 RAW RESPONSE STRUCTURE:")
	t.Logf("   Status Code: %d", resp.StatusCode)
	t.Logf("   Content Length: %d bytes", len(body))

	// Show top-level structure
	if chart, ok := rawData["chart"].(map[string]interface{}); ok {
		t.Logf("   Chart object keys: %v", getKeys(chart))

		if result, ok := chart["result"].([]interface{}); ok && len(result) > 0 {
			if firstResult, ok := result[0].(map[string]interface{}); ok {
				t.Logf("   Result object keys: %v", getKeys(firstResult))

				if meta, ok := firstResult["meta"].(map[string]interface{}); ok {
					t.Logf("   Meta object keys: %v", getKeys(meta))
					t.Logf("   Meta data sample:")
					for key, value := range meta {
						if key == "regularMarketPrice" || key == "currency" || key == "exchangeName" {
							t.Logf("     %s: %v", key, value)
						}
					}
				}
			}
		}
	}

	// Show a sample of the raw JSON (first 500 characters)
	jsonStr := string(body)
	if len(jsonStr) > 500 {
		jsonStr = jsonStr[:500] + "..."
	}
	t.Logf("   Raw JSON sample: %s", jsonStr)
}

// Helper function to get keys from map
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Benchmark test for performance
func BenchmarkGetPriceFromYahooFinance(b *testing.B) {
	client := NewAPIClient()
	symbol := "AAPL"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetPriceFromYahooFinance(symbol)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}

		// Respect rate limits
		time.Sleep(client.GetRateLimit())
	}
}

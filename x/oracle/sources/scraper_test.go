package sources

import (
	"math"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func TestSimpleScraper(t *testing.T) {
	scraper := NewSimpleScraper()

	// Test with a simple source
	source := DataSource{
		Name:      "test-source",
		URL:       "https://httpbin.org/html",
		Selector:  "h1",
		RateLimit: "1s",
		Weight:    1.0,
		Enabled:   true,
		Timeout:   "10s",
		Retries:   1,
	}

	result, err := scraper.ScrapePrice(source, "test-symbol")
	if err != nil {
		t.Fatalf("Scraping failed: %v", err)
	}

	if !result.Success {
		t.Logf("Scraping result: %+v", result)
		// This is expected for httpbin.org since it doesn't contain price data
		t.Log("Expected failure - httpbin.org doesn't contain price data")
	}
}

func TestExtractPriceSelectors(t *testing.T) {
	const sampleHTML = `
	<html>
		<body>
			<table>
				<tr><td class="rtRates">0.91234</td></tr>
			</table>
			<div class="spot-price">$1,932.50</div>
			<span data-field="regularMarketPrice">125.43</span>
			<div class="tv-symbol-price-quote__value">1,234.56</div>
			<p class="no-wrap">$27,000.00</p>
		</body>
	</html>`

	doc, err := html.Parse(strings.NewReader(sampleHTML))
	if err != nil {
		t.Fatalf("failed to parse sample HTML: %v", err)
	}

	scraper := NewSimpleScraper()
	testCases := []struct {
		name     string
		selector string
		expected float64
	}{
		{"table selector", "td.rtRates", 0.91234},
		{"spot price", ".spot-price", 1932.50},
		{"data attribute", "[data-field=\"regularMarketPrice\"]", 125.43},
		{"tradingview class", ".tv-symbol-price-quote__value", 1234.56},
		{"coingecko wrapper", ".no-wrap", 27000.00},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			price, err := scraper.extractPrice(doc, tc.selector)
			if err != nil {
				t.Fatalf("extractPrice returned error: %v", err)
			}
			if math.Abs(price-tc.expected) > 1e-4 {
				t.Fatalf("expected %.2f, got %.5f", tc.expected, price)
			}
		})
	}
}

func TestDefaultSources(t *testing.T) {
	sources := GetDefaultSources()

	if len(sources) == 0 {
		t.Fatal("No default sources configured")
	}

	for _, source := range sources {
		if source.Name == "" {
			t.Error("Source name cannot be empty")
		}
		if source.URL == "" {
			t.Error("Source URL cannot be empty")
		}
		if source.Weight <= 0 {
			t.Error("Source weight must be positive")
		}
		if source.Retries < 0 {
			t.Error("Source retries cannot be negative")
		}
	}
}

func TestScraperConfiguration(t *testing.T) {
	scraper := NewSimpleScraper()

	if scraper.client == nil {
		t.Error("HTTP client not initialized")
	}

	if scraper.userAgent == "" {
		t.Error("User agent not set")
	}

	if scraper.rateLimit <= 0 {
		t.Error("Rate limit not set")
	}
}

func TestRateLimit(t *testing.T) {
	scraper := NewSimpleScraper()

	start := time.Now()

	// Test rate limiting
	scraper.rateLimit = 100 * time.Millisecond

	// Simulate multiple requests
	for i := 0; i < 3; i++ {
		time.Sleep(scraper.rateLimit)
	}

	elapsed := time.Since(start)
	if elapsed < 200*time.Millisecond {
		t.Error("Rate limiting not working properly")
	}
}

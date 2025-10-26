package sources

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// ImprovedScraper handles more sophisticated scraping
type ImprovedScraper struct {
	client    *http.Client
	userAgent string
	rateLimit time.Duration
}

// NewImprovedScraper creates a new improved scraper
func NewImprovedScraper() *ImprovedScraper {
	return &ImprovedScraper{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		rateLimit: 2 * time.Second,
	}
}

// ScrapePriceImproved scrapes price with better parsing
func (s *ImprovedScraper) ScrapePriceImproved(source DataSource, symbol string) (*ScrapingResult, error) {
	if !source.Enabled {
		return &ScrapingResult{
			Symbol:  symbol,
			Source:  source.Name,
			Success: false,
			Error:   "source disabled",
		}, nil
	}

	// Build proper URL based on source
	url := s.buildURL(source, symbol)

	// Create request with better headers
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &ScrapingResult{
			Symbol:  symbol,
			Source:  source.Name,
			Success: false,
			Error:   fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}

	// Set comprehensive headers
	s.setHeaders(req)

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return &ScrapingResult{
			Symbol:  symbol,
			Source:  source.Name,
			Success: false,
			Error:   fmt.Sprintf("request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ScrapingResult{
			Symbol:  symbol,
			Source:  source.Name,
			Success: false,
			Error:   fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
		}, nil
	}

	// Parse HTML with better extraction
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return &ScrapingResult{
			Symbol:  symbol,
			Source:  source.Name,
			Success: false,
			Error:   fmt.Sprintf("failed to parse HTML: %v", err),
		}, nil
	}

	// Extract price using improved method
	price, err := s.extractPriceImproved(doc, source, symbol)
	if err != nil {
		return &ScrapingResult{
			Symbol:  symbol,
			Source:  source.Name,
			Success: false,
			Error:   fmt.Sprintf("failed to extract price: %v", err),
		}, nil
	}

	return &ScrapingResult{
		Symbol:    symbol,
		Price:     price,
		Timestamp: time.Now(),
		Source:    source.Name,
		Success:   true,
	}, nil
}

// buildURL constructs proper URL for each source
func (s *ImprovedScraper) buildURL(source DataSource, symbol string) string {
	baseURL := source.URL

	switch source.Name {
	case "investing.com":
		// Convert symbol to investing.com format
		symbol = strings.ReplaceAll(symbol, "/", "-")
		symbol = strings.ToLower(symbol)
		return strings.ReplaceAll(baseURL, "{symbol}", symbol)

	case "x-rates.com":
		// x-rates.com doesn't use symbol in URL, it's a table
		return baseURL

	case "kitco.com":
		// Kitco is for gold only
		return baseURL

	case "coinmarketcap.com":
		// Convert to CMC format
		symbol = strings.ReplaceAll(symbol, "/", "")
		symbol = strings.ToLower(symbol)
		return strings.ReplaceAll(baseURL, "{symbol}", symbol)

	case "coingecko.com":
		// Convert to CoinGecko format
		symbol = strings.ReplaceAll(symbol, "/", "")
		symbol = strings.ToLower(symbol)
		return strings.ReplaceAll(baseURL, "{symbol}", symbol)

	default:
		return strings.ReplaceAll(baseURL, "{symbol}", symbol)
	}
}

// setHeaders sets comprehensive headers to avoid detection
func (s *ImprovedScraper) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")
}

// extractPriceImproved extracts price with better logic
func (s *ImprovedScraper) extractPriceImproved(doc *html.Node, source DataSource, symbol string) (float64, error) {
	// Different extraction strategies for different sources
	switch source.Name {
	case "x-rates.com":
		return s.extractXratesPrice(doc, symbol)
	case "kitco.com":
		return s.extractKitcoPrice(doc)
	case "investing.com":
		return s.extractInvestingPrice(doc)
	default:
		return s.extractGenericPrice(doc)
	}
}

// extractXratesPrice extracts price from x-rates.com
func (s *ImprovedScraper) extractXratesPrice(doc *html.Node, symbol string) (float64, error) {
	// Look for specific currency rates in the table
	var priceText string
	var walker func(*html.Node)

	walker = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			// Look for price patterns that match currency rates
			if matched, _ := regexp.MatchString(`^\d+\.\d{4}$`, text); matched {
				// Check if this is in a table cell
				parent := n.Parent
				if parent != nil && parent.Data == "td" {
					priceText = text
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}
	walker(doc)

	if priceText == "" {
		return 0, fmt.Errorf("no price found for %s", symbol)
	}

	var price float64
	_, err := fmt.Sscanf(priceText, "%f", &price)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price '%s': %v", priceText, err)
	}

	return price, nil
}

// extractKitcoPrice extracts gold price from kitco.com
func (s *ImprovedScraper) extractKitcoPrice(doc *html.Node) (float64, error) {
	// Look for gold price patterns
	var priceText string
	var walker func(*html.Node)

	walker = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			// Look for gold price patterns (usually 4 decimal places)
			if matched, _ := regexp.MatchString(`^\d{3,4}\.\d{2}$`, text); matched {
				priceText = text
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}
	walker(doc)

	if priceText == "" {
		return 0, fmt.Errorf("no gold price found")
	}

	var price float64
	_, err := fmt.Sscanf(priceText, "%f", &price)
	if err != nil {
		return 0, fmt.Errorf("failed to parse gold price '%s': %v", priceText, err)
	}

	return price, nil
}

// extractInvestingPrice extracts price from investing.com
func (s *ImprovedScraper) extractInvestingPrice(doc *html.Node) (float64, error) {
	// Look for price in investing.com format
	var priceText string
	var walker func(*html.Node)

	walker = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			// Look for price patterns (usually 4-5 decimal places)
			if matched, _ := regexp.MatchString(`^\d+\.\d{4,5}$`, text); matched {
				priceText = text
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}
	walker(doc)

	if priceText == "" {
		return 0, fmt.Errorf("no price found on investing.com")
	}

	var price float64
	_, err := fmt.Sscanf(priceText, "%f", &price)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price '%s': %v", priceText, err)
	}

	return price, nil
}

// extractGenericPrice extracts price using generic method
func (s *ImprovedScraper) extractGenericPrice(doc *html.Node) (float64, error) {
	// Generic price extraction
	var priceText string
	var walker func(*html.Node)

	walker = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			// Look for various price patterns
			patterns := []string{
				`^\d+\.\d{2}$`,           // 123.45
				`^\d+\.\d{4}$`,           // 123.4567
				`^\d+\.\d{5}$`,           // 123.45678
				`^\d{1,3},\d{3}\.\d{2}$`, // 1,234.56
			}

			for _, pattern := range patterns {
				if matched, _ := regexp.MatchString(pattern, text); matched {
					priceText = text
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}
	walker(doc)

	if priceText == "" {
		return 0, fmt.Errorf("no price found")
	}

	// Clean the price text
	priceText = strings.ReplaceAll(priceText, ",", "")

	var price float64
	_, err := fmt.Sscanf(priceText, "%f", &price)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price '%s': %v", priceText, err)
	}

	return price, nil
}

// GetUserAgent returns the user agent string
func (s *ImprovedScraper) GetUserAgent() string {
	return s.userAgent
}

// GetRateLimit returns the rate limit duration
func (s *ImprovedScraper) GetRateLimit() time.Duration {
	return s.rateLimit
}

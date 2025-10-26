package sources

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// SimpleScraper handles basic HTML scraping
type SimpleScraper struct {
	client    *http.Client
	userAgent string
	rateLimit time.Duration
}

// ScrapingResult represents the result of a scraping operation
type ScrapingResult struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

// DataSource defines a data source configuration
type DataSource struct {
	Name      string  `json:"name"`
	URL       string  `json:"url"`
	Selector  string  `json:"selector"`
	RateLimit string  `json:"rate_limit"`
	Weight    float64 `json:"weight"`
	Enabled   bool    `json:"enabled"`
	Timeout   string  `json:"timeout"`
	Retries   int     `json:"retries"`
}

var pricePattern = regexp.MustCompile(`[-+]?\d[\d.,]*`)

// NewSimpleScraper creates a new scraper instance
func NewSimpleScraper() *SimpleScraper {
	return &SimpleScraper{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		rateLimit: 2 * time.Second,
	}
}

// ScrapePrice scrapes price from a given source
func (s *SimpleScraper) ScrapePrice(source DataSource, symbol string) (*ScrapingResult, error) {
	if !source.Enabled {
		return &ScrapingResult{
			Symbol:  symbol,
			Source:  source.Name,
			Success: false,
			Error:   "source disabled",
		}, nil
	}

	// Build URL with symbol (only if URL contains {symbol})
	url := source.URL
	if strings.Contains(source.URL, "{symbol}") {
		url = strings.ReplaceAll(source.URL, "{symbol}", symbol)
	}

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &ScrapingResult{
			Symbol:  symbol,
			Source:  source.Name,
			Success: false,
			Error:   fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}

	// Set comprehensive headers to avoid detection
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("DNT", "1")

	// Execute request with retries
	var resp *http.Response
	for i := 0; i <= source.Retries; i++ {
		resp, err = s.client.Do(req)
		if err == nil {
			break
		}
		if i < source.Retries {
			time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
		}
	}

	if err != nil {
		return &ScrapingResult{
			Symbol:  symbol,
			Source:  source.Name,
			Success: false,
			Error:   fmt.Sprintf("request failed after %d retries: %v", source.Retries, err),
		}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return &ScrapingResult{
			Symbol:  symbol,
			Source:  source.Name,
			Success: false,
			Error:   fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
		}, nil
	}

	bodyReader, closeBody, err := prepareBodyReader(resp)
	if err != nil {
		return &ScrapingResult{
			Symbol:  symbol,
			Source:  source.Name,
			Success: false,
			Error:   fmt.Sprintf("failed to prepare response body: %v", err),
		}, nil
	}
	defer closeBody()

	// Read body
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return &ScrapingResult{
			Symbol:  symbol,
			Source:  source.Name,
			Success: false,
			Error:   fmt.Sprintf("failed to read response body: %v", err),
		}, nil
	}

	// Parse HTML
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return &ScrapingResult{
			Symbol:  symbol,
			Source:  source.Name,
			Success: false,
			Error:   fmt.Sprintf("failed to parse HTML: %v", err),
		}, nil
	}

	// Extract price using selector
	price, err := s.extractPrice(doc, source.Selector)
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

// extractPrice extracts price from HTML using CSS selector
func (s *SimpleScraper) extractPrice(doc *html.Node, selector string) (float64, error) {
	if selector == "" {
		return 0, fmt.Errorf("empty selector")
	}

	// Специальная обработка для investing.com - ищем в текстовых узлах
	if selector == "text" {
		return s.extractPriceFromText(doc)
	}

	for _, sel := range splitSelectors(selector) {
		nodes := findMatchingNodes(doc, sel)
		for _, node := range nodes {
			priceText := strings.TrimSpace(getTextContent(node))
			if priceText == "" {
				continue
			}

			price, err := parsePriceText(priceText)
			if err != nil {
				continue
			}

			if !s.isReasonablePrice(price) {
				continue
			}

			return price, nil
		}
	}

	return 0, fmt.Errorf("no price found for selector %s", selector)
}

// extractPriceFromText ищет цены в текстовых узлах (для investing.com)
func (s *SimpleScraper) extractPriceFromText(doc *html.Node) (float64, error) {
	// Ищем цены в контексте - рядом с ключевыми словами
	priceCandidates := s.findPriceInContext(doc)

	if len(priceCandidates) == 0 {
		return 0, fmt.Errorf("no price found in text nodes")
	}

	// Сортируем по приоритету: основная цена, затем по порядку появления
	sort.Slice(priceCandidates, func(i, j int) bool {
		return priceCandidates[i].Priority > priceCandidates[j].Priority
	})

	return priceCandidates[0].Price, nil
}

// PriceCandidate представляет найденную цену с контекстом
type PriceCandidate struct {
	Price    float64
	Priority int
	Context  string
}

// findPriceInContext ищет цены в контексте ключевых слов
func (s *SimpleScraper) findPriceInContext(doc *html.Node) []PriceCandidate {
	var candidates []PriceCandidate
	var walker func(*html.Node)

	// Ключевые слова для определения основной цены
	priceKeywords := []string{"price", "rate", "last", "current", "bid", "ask", "close"}

	walker = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" && isPriceLike(text) {
				price, err := parsePriceText(text)
				if err == nil && s.isReasonablePrice(price) {
					// Анализируем контекст вокруг цены
					context := s.getPriceContext(n)
					priority := s.calculatePriority(context, priceKeywords)

					candidates = append(candidates, PriceCandidate{
						Price:    price,
						Priority: priority,
						Context:  context,
					})
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}

	walker(doc)
	return candidates
}

// getPriceContext получает контекст вокруг цены
func (s *SimpleScraper) getPriceContext(node *html.Node) string {
	var context strings.Builder

	// Получаем текст родительского элемента
	if node.Parent != nil {
		parentText := getTextContent(node.Parent)
		context.WriteString(parentText)
		context.WriteString(" ")
	}

	// Получаем текст соседних элементов
	if node.Parent != nil {
		for c := node.Parent.FirstChild; c != nil; c = c.NextSibling {
			if c != node {
				text := getTextContent(c)
				if text != "" {
					context.WriteString(text)
					context.WriteString(" ")
				}
			}
		}
	}

	return strings.TrimSpace(context.String())
}

// calculatePriority вычисляет приоритет цены на основе контекста
func (s *SimpleScraper) calculatePriority(context string, keywords []string) int {
	priority := 0
	contextLower := strings.ToLower(context)

	// Высокий приоритет для ключевых слов
	for _, keyword := range keywords {
		if strings.Contains(contextLower, keyword) {
			priority += 10
		}
	}

	// Дополнительные факторы
	if strings.Contains(contextLower, "main") || strings.Contains(contextLower, "primary") {
		priority += 5
	}

	if strings.Contains(contextLower, "real-time") || strings.Contains(contextLower, "live") {
		priority += 3
	}

	// Штраф за нежелательные слова
	if strings.Contains(contextLower, "change") || strings.Contains(contextLower, "percent") {
		priority -= 2
	}

	return priority
}

// getTextContent извлекает весь текст из узла
func getTextContent(n *html.Node) string {
	var text strings.Builder
	var walker func(*html.Node)
	walker = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(strings.TrimSpace(node.Data))
			text.WriteString(" ")
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}
	walker(n)
	return text.String()
}

// isPriceLike проверяет, похоже ли значение на цену
func isPriceLike(text string) bool {
	// Простая проверка: содержит ли текст число с точкой
	return strings.Contains(text, ".") && len(text) > 3 && len(text) < 20
}

// isReasonablePrice checks if a price value is reasonable
func (s *SimpleScraper) isReasonablePrice(price float64) bool {
	if price <= 0 {
		return false
	}

	// Reasonable price ranges for different asset types
	// Forex: 0.5 - 2.0 (major pairs)
	// Crypto: 0.0001 - 100000 (wide range)
	// Stocks: 0.01 - 10000
	// Commodities: 0.1 - 10000

	return price < 1000000
}

func splitSelectors(selector string) []string {
	parts := strings.Split(selector, ",")
	selectors := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			selectors = append(selectors, trimmed)
		}
	}

	if len(selectors) == 0 && selector != "" {
		selectors = append(selectors, selector)
	}

	return selectors
}

func findMatchingNodes(doc *html.Node, selector string) []*html.Node {
	var nodes []*html.Node
	var walker func(*html.Node)

	selector = strings.TrimSpace(selector)
	if selector == "" {
		return nodes
	}

	walker = func(n *html.Node) {
		if matchSelector(n, selector) {
			nodes = append(nodes, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}

	walker(doc)
	return nodes
}

func matchSelector(n *html.Node, selector string) bool {
	if n.Type != html.ElementNode {
		return false
	}

	sel := strings.TrimSpace(selector)
	if sel == "" {
		return false
	}

	if space := strings.LastIndex(sel, " "); space != -1 {
		sel = sel[space+1:]
	}

	if strings.HasPrefix(sel, "#") {
		value, ok := getAttribute(n, "id")
		return ok && value == sel[1:]
	}

	if strings.HasPrefix(sel, ".") {
		return hasClass(n, sel[1:])
	}

	if strings.HasPrefix(sel, "[") && strings.HasSuffix(sel, "]") {
		attr, value := parseAttributeSelector(sel)
		attrVal, ok := getAttribute(n, attr)
		if !ok {
			return false
		}
		if value == "" {
			return true
		}
		return strings.EqualFold(attrVal, value)
	}

	if idx := strings.Index(sel, "["); idx > 0 && strings.HasSuffix(sel, "]") {
		tag := strings.TrimSpace(sel[:idx])
		if tag != "" && !strings.EqualFold(n.Data, tag) {
			return false
		}
		attr, value := parseAttributeSelector(sel[idx:])
		attrVal, ok := getAttribute(n, attr)
		if !ok {
			return false
		}
		if value == "" {
			return true
		}
		return strings.EqualFold(attrVal, value)
	}

	return matchTagAndClasses(n, sel)
}

func matchTagAndClasses(n *html.Node, selector string) bool {
	parts := strings.Split(selector, ".")
	tag := strings.TrimSpace(parts[0])
	if tag != "" && !strings.EqualFold(n.Data, tag) {
		return false
	}

	if len(parts) == 1 {
		return tag != ""
	}

	for _, class := range parts[1:] {
		class = strings.TrimSpace(class)
		if class == "" {
			continue
		}
		if !hasClass(n, class) {
			return false
		}
	}

	return true
}

func hasClass(n *html.Node, className string) bool {
	classAttr, ok := getAttribute(n, "class")
	if !ok {
		return false
	}

	for _, class := range strings.Fields(classAttr) {
		if class == className {
			return true
		}
	}

	return false
}

func getAttribute(n *html.Node, name string) (string, bool) {
	for _, attr := range n.Attr {
		if strings.EqualFold(attr.Key, name) {
			return attr.Val, true
		}
	}

	return "", false
}

func parseAttributeSelector(selector string) (string, string) {
	inner := strings.TrimSpace(selector[1 : len(selector)-1])
	if inner == "" {
		return "", ""
	}

	if strings.Contains(inner, "=") {
		parts := strings.SplitN(inner, "=", 2)
		attr := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")
		return attr, value
	}

	return inner, ""
}

func parsePriceText(text string) (float64, error) {
	matches := pricePattern.FindAllString(text, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("no numeric value found in %q", text)
	}

	for i := len(matches) - 1; i >= 0; i-- {
		normalized := normalizeNumericValue(matches[i])
		if normalized == "" {
			continue
		}
		value, err := strconv.ParseFloat(normalized, 64)
		if err != nil {
			continue
		}
		return value, nil
	}

	return 0, fmt.Errorf("unable to parse price from %q", text)
}

func normalizeNumericValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	sign := ""
	if strings.HasPrefix(value, "+") || strings.HasPrefix(value, "-") {
		sign = value[:1]
		value = value[1:]
	}

	value = strings.Trim(value, ",.")
	if value == "" {
		return ""
	}

	value = strings.ReplaceAll(value, " ", "")

	dotIndex := strings.LastIndex(value, ".")
	commaIndex := strings.LastIndex(value, ",")

	switch {
	case dotIndex == -1 && commaIndex == -1:
		// nothing to normalize
	case dotIndex == -1:
		// only commas -> treat as decimal separator
		value = strings.ReplaceAll(value, ".", "")
		value = strings.ReplaceAll(value, ",", ".")
	case commaIndex == -1:
		// only dots -> if multiple, treat last as decimal separator
		if strings.Count(value, ".") > 1 {
			lastDot := dotIndex
			value = strings.Replace(value[:lastDot], ".", "", -1) + value[lastDot:]
		}
	default:
		if commaIndex > dotIndex {
			// comma acts as decimal separator
			value = strings.ReplaceAll(value, ".", "")
			value = strings.ReplaceAll(value, ",", ".")
		} else {
			// dot is decimal separator, remove thousands commas
			value = strings.ReplaceAll(value, ",", "")
		}
	}

	return sign + value
}

// GetUserAgent returns the user agent string
func (s *SimpleScraper) GetUserAgent() string {
	return s.userAgent
}

// GetRateLimit returns the rate limit duration
func (s *SimpleScraper) GetRateLimit() time.Duration {
	return s.rateLimit
}

// prepareBodyReader returns a reader that transparently decodes common encodings.
func prepareBodyReader(resp *http.Response) (io.Reader, func(), error) {
	encoding := normalizeContentEncoding(resp.Header.Get("Content-Encoding"))

	switch encoding {
	case "", "identity":
		return resp.Body, func() { resp.Body.Close() }, nil
	case "gzip":
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, func() {}, fmt.Errorf("gzip decode init failed: %w", err)
		}
		return reader, func() {
			reader.Close()
			resp.Body.Close()
		}, nil
	case "deflate":
		reader, err := zlib.NewReader(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, func() {}, fmt.Errorf("deflate decode init failed: %w", err)
		}
		return reader, func() {
			reader.Close()
			resp.Body.Close()
		}, nil
	case "br":
		resp.Body.Close()
		return nil, func() {}, fmt.Errorf("brotli compression not supported")
	default:
		resp.Body.Close()
		return nil, func() {}, fmt.Errorf("unsupported content encoding: %s", encoding)
	}
}

func normalizeContentEncoding(header string) string {
	if header == "" {
		return ""
	}

	parts := strings.Split(header, ",")
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			return strings.ToLower(value)
		}
	}

	return ""
}

// GetDefaultSources returns default data source configurations
func GetDefaultSources() []DataSource {
	return []DataSource{
		{
			Name:      "investing.com",
			URL:       "https://www.investing.com/commodities/{symbol}",
			Selector:  "[data-test=\"instrument-price-last\"]",
			RateLimit: "1s",
			Weight:    0.3,
			Enabled:   true,
			Timeout:   "10s",
			Retries:   3,
		},
		{
			Name:      "x-rates.com",
			URL:       "https://www.x-rates.com/table/?from=USD&amount=1",
			Selector:  "td.rtRates",
			RateLimit: "2s",
			Weight:    0.2,
			Enabled:   true,
			Timeout:   "10s",
			Retries:   3,
		},
		{
			Name:      "kitco.com",
			URL:       "https://www.kitco.com/gold-price-today-usa/",
			Selector:  ".CommodityPrice_spotPriceGrid__uiD_5",
			RateLimit: "5s",
			Weight:    0.5,
			Enabled:   true,
			Timeout:   "10s",
			Retries:   3,
		},
		// Ценные ресурсы - Yahoo Finance
		{
			Name:      "yahoo-finance-gold",
			URL:       "https://finance.yahoo.com/quote/GC=F",
			Selector:  ".price.yf-19hyiou",
			RateLimit: "3s",
			Weight:    0.3,
			Enabled:   true,
			Timeout:   "10s",
			Retries:   3,
		},
		{
			Name:      "yahoo-finance-silver",
			URL:       "https://finance.yahoo.com/quote/SI=F",
			Selector:  ".price.yf-19hyiou",
			RateLimit: "3s",
			Weight:    0.3,
			Enabled:   true,
			Timeout:   "10s",
			Retries:   3,
		},
		{
			Name:      "yahoo-finance-oil",
			URL:       "https://finance.yahoo.com/quote/CL=F",
			Selector:  ".price.yf-19hyiou",
			RateLimit: "3s",
			Weight:    0.3,
			Enabled:   true,
			Timeout:   "10s",
			Retries:   3,
		},
		{
			Name:      "yahoo-finance-gas",
			URL:       "https://finance.yahoo.com/quote/NG=F",
			Selector:  ".price.yf-19hyiou",
			RateLimit: "3s",
			Weight:    0.3,
			Enabled:   true,
			Timeout:   "10s",
			Retries:   3,
		},
		{
			Name:      "yahoo-finance-copper",
			URL:       "https://finance.yahoo.com/quote/HG=F",
			Selector:  ".price.yf-19hyiou",
			RateLimit: "3s",
			Weight:    0.3,
			Enabled:   true,
			Timeout:   "10s",
			Retries:   3,
		},
		{
			Name:      "yahoo-finance-platinum",
			URL:       "https://finance.yahoo.com/quote/PL=F",
			Selector:  ".price.yf-19hyiou",
			RateLimit: "3s",
			Weight:    0.3,
			Enabled:   true,
			Timeout:   "10s",
			Retries:   3,
		},
	}
}

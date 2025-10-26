package oracle

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

// Client represents an oracle client for querying prices
type Client struct {
	conn   *grpc.ClientConn
	client types.QueryClient
}

// NewClient creates a new oracle client
func NewClient(nodeURL string) (*Client, error) {
	if nodeURL == "" {
		return nil, fmt.Errorf("node URL is required")
	}

	// Create gRPC connection
	conn, err := grpc.Dial(nodeURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node: %w", err)
	}

	// Create query client
	client := types.NewQueryClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// PriceData represents price information
type PriceData struct {
	Symbol     string    `json:"symbol"`
	Value      string    `json:"value"`
	Source     string    `json:"source"`
	Timestamp  time.Time `json:"timestamp"`
	Confidence float32   `json:"confidence"`
}

// GetPrice queries the current price for a symbol
func (c *Client) GetPrice(ctx context.Context, symbol string) (*PriceData, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	req := &types.QueryPriceRequest{
		Symbol: symbol,
	}

	resp, err := c.client.Price(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to query price for %s: %w", symbol, err)
	}

	if resp.Price == nil {
		return nil, fmt.Errorf("no price data returned for %s", symbol)
	}

	// Convert timestamp from Unix timestamp to time.Time
	timestamp := time.Unix(resp.Price.Timestamp, 0)

	return &PriceData{
		Symbol:     resp.Price.Symbol,
		Value:      resp.Price.Value,
		Source:     resp.Price.Source,
		Timestamp:  timestamp,
		Confidence: resp.Price.Confidence,
	}, nil
}

// GetPrices queries prices for multiple symbols
func (c *Client) GetPrices(ctx context.Context, symbols []string) (map[string]*PriceData, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("at least one symbol is required")
	}

	prices := make(map[string]*PriceData)
	var errors []error

	for _, symbol := range symbols {
		price, err := c.GetPrice(ctx, symbol)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get price for %s: %w", symbol, err))
			continue
		}
		prices[symbol] = price
	}

	if len(errors) > 0 && len(prices) == 0 {
		return nil, fmt.Errorf("failed to get any prices: %v", errors)
	}

	return prices, nil
}

// IsPriceStale checks if a price is older than the specified duration
func (c *Client) IsPriceStale(price *PriceData, maxAge time.Duration) bool {
	if price == nil {
		return true
	}
	return time.Since(price.Timestamp) > maxAge
}

// ValidatePrice checks if a price meets basic validation criteria
func (c *Client) ValidatePrice(price *PriceData) error {
	if price == nil {
		return fmt.Errorf("price is nil")
	}

	if price.Symbol == "" {
		return fmt.Errorf("price symbol is empty")
	}

	if price.Value == "" {
		return fmt.Errorf("price value is empty")
	}

	if price.Confidence < 0 || price.Confidence > 1 {
		return fmt.Errorf("price confidence must be between 0 and 1, got %f", price.Confidence)
	}

	// Check if price is too old (more than 1 hour)
	if time.Since(price.Timestamp) > time.Hour {
		return fmt.Errorf("price is too old: %v", time.Since(price.Timestamp))
	}

	return nil
}

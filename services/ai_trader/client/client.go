package client

import (
	"context"
	"fmt"
	"time"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/authz"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/oracle"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/trading"
)

// Client represents the main AI trader client that combines oracle, trading, and authz functionality
type Client struct {
	oracleClient  *oracle.Client
	tradingClient *trading.Client
	authzClient   *authz.Client
	nodeURL       string
}

// NewClient creates a new AI trader client
func NewClient(nodeURL string) (*Client, error) {
	if nodeURL == "" {
		return nil, fmt.Errorf("node URL is required")
	}

	// Create oracle client
	oracleClient, err := oracle.NewClient(nodeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create oracle client: %w", err)
	}

	// Create trading client
	tradingClient, err := trading.NewClient(nodeURL)
	if err != nil {
		oracleClient.Close()
		return nil, fmt.Errorf("failed to create trading client: %w", err)
	}

	// Create authz client
	authzClient, err := authz.NewClient(nodeURL)
	if err != nil {
		oracleClient.Close()
		tradingClient.Close()
		return nil, fmt.Errorf("failed to create authz client: %w", err)
	}

	return &Client{
		oracleClient:  oracleClient,
		tradingClient: tradingClient,
		authzClient:   authzClient,
		nodeURL:       nodeURL,
	}, nil
}

// Close closes all client connections
func (c *Client) Close() error {
	var errors []error

	if c.oracleClient != nil {
		if err := c.oracleClient.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close oracle client: %w", err))
		}
	}

	if c.tradingClient != nil {
		if err := c.tradingClient.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close trading client: %w", err))
		}
	}

	if c.authzClient != nil {
		if err := c.authzClient.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close authz client: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing clients: %v", errors)
	}

	return nil
}

// GetOracleClient returns the oracle client
func (c *Client) GetOracleClient() *oracle.Client {
	return c.oracleClient
}

// GetTradingClient returns the trading client
func (c *Client) GetTradingClient() *trading.Client {
	return c.tradingClient
}

// GetAuthzClient returns the authz client
func (c *Client) GetAuthzClient() *authz.Client {
	return c.authzClient
}

// TradingDecision represents a trading decision
type TradingDecision struct {
	Symbol       string              `json:"symbol"`
	Action       string              `json:"action"` // "buy", "sell", "hold"
	Amount       string              `json:"amount"`
	Price        string              `json:"price"`
	Reason       string              `json:"reason"`
	Confidence   float32             `json:"confidence"`
	Market       trading.TradeMarket `json:"market,omitempty"`
	PaymentDenom string              `json:"payment_denom,omitempty"` // Optional for assets (defaults to NDOLLAR), used for bonding curve trades
	MinOutput    string              `json:"min_output,omitempty"`    // Optional bonding curve slippage control
}

// DecisionMaker abstracts an AI decision engine (e.g., risk.AIDecider).
// It returns a TradingDecision that can be executed by this client.
type DecisionMaker interface {
	MakeAIDecision(ctx context.Context, symbols []string) (*TradingDecision, error)
}

// ExecuteTradingDecision executes a trading decision using delegated permissions
func (c *Client) ExecuteTradingDecision(ctx context.Context, decision *TradingDecision, grantee string, granter string) (*authz.ExecResponse, error) {
	if decision == nil {
		return nil, fmt.Errorf("trading decision is required")
	}

	if grantee == "" {
		return nil, fmt.Errorf("grantee address is required")
	}

	if granter == "" {
		return nil, fmt.Errorf("granter address is required")
	}

	if decision.Action == "hold" {
		return &authz.ExecResponse{
			Timestamp: time.Now(),
			Success:   true,
		}, nil
	}

	market := decision.Market
	if market == "" {
		market = trading.MarketAssets
	}

	switch market {
	case trading.MarketAssets:
		switch decision.Action {
		case "buy":
			return c.authzClient.ExecuteBuyAsset(ctx, grantee, granter, decision.Symbol, decision.Amount)
		case "sell":
			return c.authzClient.ExecuteSellAsset(ctx, grantee, granter, decision.Symbol, decision.Amount)
		default:
			return nil, fmt.Errorf("invalid trading action: %s", decision.Action)
		}
	case trading.MarketBondingCurve:
		switch decision.Action {
		case "buy":
			return c.authzClient.ExecuteBuyFromCurve(ctx, grantee, granter, decision.Symbol, decision.PaymentDenom, decision.Amount, decision.MinOutput)
		case "sell":
			return c.authzClient.ExecuteSellToCurve(ctx, grantee, granter, decision.Symbol, decision.Amount, decision.PaymentDenom, decision.MinOutput)
		default:
			return nil, fmt.Errorf("invalid trading action: %s", decision.Action)
		}
	default:
		return nil, fmt.Errorf("invalid trading action: %s", decision.Action)
	}
}

// DecideAndExecute asks a DecisionMaker for a decision and executes it via authz.
func (c *Client) DecideAndExecute(ctx context.Context, dm DecisionMaker, symbols []string, grantee, granter string) (*authz.ExecResponse, *TradingDecision, error) {
	if dm == nil {
		return nil, nil, fmt.Errorf("decision maker is required")
	}
	decision, err := dm.MakeAIDecision(ctx, symbols)
	if err != nil {
		return nil, nil, err
	}
	if decision == nil {
		return nil, nil, fmt.Errorf("empty decision")
	}
	// For hold, skip execution but return decision
	if decision.Action == "hold" {
		return &authz.ExecResponse{Timestamp: time.Now(), Success: true}, decision, nil
	}
	res, err := c.ExecuteTradingDecision(ctx, decision, grantee, granter)
	return res, decision, err
}

// GetPriceData gets price data for a symbol
func (c *Client) GetPriceData(ctx context.Context, symbol string) (*oracle.PriceData, error) {
	return c.oracleClient.GetPrice(ctx, symbol)
}

// GetMultiplePriceData gets price data for multiple symbols
func (c *Client) GetMultiplePriceData(ctx context.Context, symbols []string) (map[string]*oracle.PriceData, error) {
	return c.oracleClient.GetPrices(ctx, symbols)
}

// ValidatePriceData validates price data
func (c *Client) ValidatePriceData(price *oracle.PriceData) error {
	return c.oracleClient.ValidatePrice(price)
}

// IsPriceStale checks if price data is stale
func (c *Client) IsPriceStale(price *oracle.PriceData, maxAge time.Duration) bool {
	return c.oracleClient.IsPriceStale(price, maxAge)
}

// ExecuteDirectTrade executes a direct trade (without authz delegation)
func (c *Client) ExecuteDirectTrade(ctx context.Context, trader string, req *trading.TradeRequest) (*trading.TradeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("trade request is required")
	}
	return c.tradingClient.ExecuteTrade(ctx, trader, req)
}

// HealthCheck performs a health check on all clients
func (c *Client) HealthCheck(ctx context.Context) error {
	// Test oracle client by querying a common symbol
	_, err := c.oracleClient.GetPrice(ctx, "BTC")
	if err != nil {
		return fmt.Errorf("oracle client health check failed: %w", err)
	}

	// Note: We can't easily test trading and authz clients without actual transactions
	// In a real implementation, you might want to add more sophisticated health checks

	return nil
}

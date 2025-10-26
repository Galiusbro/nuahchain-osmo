package trading

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
)

// Client represents a trading client for executing buy/sell operations
type Client struct {
	conn   *grpc.ClientConn
	client types.MsgClient
}

// NewClient creates a new trading client
func NewClient(nodeURL string) (*Client, error) {
	if nodeURL == "" {
		return nil, fmt.Errorf("node URL is required")
	}

	// Create gRPC connection
	conn, err := grpc.Dial(nodeURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node: %w", err)
	}

	// Create message client
	client := types.NewMsgClient(conn)

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

// TradeRequest represents a trade request
type TradeRequest struct {
	Symbol string `json:"symbol"`
	Amount string `json:"amount"` // Amount in NDOLLAR for buy, base amount for sell
	Type   string `json:"type"`   // "buy" or "sell"
}

// TradeResponse represents a trade response
type TradeResponse struct {
	Symbol    string    `json:"symbol"`
	Type      string    `json:"type"`
	Amount    string    `json:"amount"`
	Result    string    `json:"result"` // Result amount (base amount for buy, NDOLLAR for sell)
	TxHash    string    `json:"tx_hash"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

// BuyAsset executes a buy asset transaction
func (c *Client) BuyAsset(ctx context.Context, buyer string, symbol string, amountNDOLLAR string) (*TradeResponse, error) {
	if buyer == "" {
		return nil, fmt.Errorf("buyer address is required")
	}
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if amountNDOLLAR == "" {
		return nil, fmt.Errorf("amount is required")
	}

	// Validate buyer address
	_, err := sdk.AccAddressFromBech32(buyer)
	if err != nil {
		return nil, fmt.Errorf("invalid buyer address: %w", err)
	}

	// Create buy message
	msg := types.NewMsgBuyAsset(buyer, symbol, amountNDOLLAR)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid buy message: %w", err)
	}

	// Execute transaction
	resp, err := c.client.BuyAsset(ctx, msg)
	if err != nil {
		return &TradeResponse{
			Symbol:    symbol,
			Type:      "buy",
			Amount:    amountNDOLLAR,
			Timestamp: time.Now(),
			Success:   false,
			Error:     err.Error(),
		}, fmt.Errorf("failed to execute buy transaction: %w", err)
	}

	return &TradeResponse{
		Symbol:    symbol,
		Type:      "buy",
		Amount:    amountNDOLLAR,
		Result:    resp.BaseAmount,
		Timestamp: time.Now(),
		Success:   true,
	}, nil
}

// SellAsset executes a sell asset transaction
func (c *Client) SellAsset(ctx context.Context, seller string, symbol string, baseAmount string) (*TradeResponse, error) {
	if seller == "" {
		return nil, fmt.Errorf("seller address is required")
	}
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if baseAmount == "" {
		return nil, fmt.Errorf("base amount is required")
	}

	// Validate seller address
	_, err := sdk.AccAddressFromBech32(seller)
	if err != nil {
		return nil, fmt.Errorf("invalid seller address: %w", err)
	}

	// Create sell message
	msg := types.NewMsgSellAsset(seller, symbol, baseAmount)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid sell message: %w", err)
	}

	// Execute transaction
	resp, err := c.client.SellAsset(ctx, msg)
	if err != nil {
		return &TradeResponse{
			Symbol:    symbol,
			Type:      "sell",
			Amount:    baseAmount,
			Timestamp: time.Now(),
			Success:   false,
			Error:     err.Error(),
		}, fmt.Errorf("failed to execute sell transaction: %w", err)
	}

	return &TradeResponse{
		Symbol:    symbol,
		Type:      "sell",
		Amount:    baseAmount,
		Result:    resp.Payout_NDOLLAR,
		Timestamp: time.Now(),
		Success:   true,
	}, nil
}

// ExecuteTrade executes a trade based on the request
func (c *Client) ExecuteTrade(ctx context.Context, trader string, req *TradeRequest) (*TradeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("trade request is required")
	}

	switch req.Type {
	case "buy":
		return c.BuyAsset(ctx, trader, req.Symbol, req.Amount)
	case "sell":
		return c.SellAsset(ctx, trader, req.Symbol, req.Amount)
	default:
		return nil, fmt.Errorf("invalid trade type: %s, must be 'buy' or 'sell'", req.Type)
	}
}

// ValidateTradeRequest validates a trade request
func (c *Client) ValidateTradeRequest(req *TradeRequest) error {
	if req == nil {
		return fmt.Errorf("trade request is required")
	}

	if req.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if req.Amount == "" {
		return fmt.Errorf("amount is required")
	}

	if req.Type != "buy" && req.Type != "sell" {
		return fmt.Errorf("invalid trade type: %s, must be 'buy' or 'sell'", req.Type)
	}

	return nil
}

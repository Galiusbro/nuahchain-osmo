package trading

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/shared"
	assetstypes "github.com/osmosis-labs/osmosis/v30/x/assets/types"
	bondingtypes "github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
)

// Client represents a trading client for executing buy/sell operations
type Client struct {
	conn          *grpc.ClientConn
	assetsClient  assetstypes.MsgClient
	bondingClient bondingtypes.MsgClient
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
	assetsClient := assetstypes.NewMsgClient(conn)
	bondingClient := bondingtypes.NewMsgClient(conn)

	return &Client{
		conn:          conn,
		assetsClient:  assetsClient,
		bondingClient: bondingClient,
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
type TradeMarket string

const (
	MarketAssets       TradeMarket = "assets"
	MarketBondingCurve TradeMarket = "bondingcurve"
)

func normalizeMarket(m TradeMarket) TradeMarket {
	if m == "" {
		return MarketAssets
	}
	return m
}

// TradeRequest represents a trade request
type TradeRequest struct {
	Symbol       string      `json:"symbol"`
	Amount       string      `json:"amount"` // For assets: buy amount in NDOLLAR / sell amount in base token; for bonding curve: payment amount (buy) or token amount (sell)
	Type         string      `json:"type"`   // "buy" or "sell"
	Market       TradeMarket `json:"market,omitempty"`
	PaymentDenom string      `json:"payment_denom,omitempty"` // Optional for assets (defaults to NDOLLAR), required for bonding curve when not using default
	MinOutput    string      `json:"min_output,omitempty"`    // For bonding curve trades: min tokens out (buy) or min payment out (sell)
}

// TradeResponse represents a trade response
type TradeResponse struct {
	Symbol       string      `json:"symbol"`
	Type         string      `json:"type"`
	Amount       string      `json:"amount"`
	Result       string      `json:"result"` // Result amount (base amount for buy, NDOLLAR for sell, or bonding curve payout)
	ResultDenom  string      `json:"result_denom,omitempty"`
	PaymentDenom string      `json:"payment_denom,omitempty"`
	Market       TradeMarket `json:"market"`
	TxHash       string      `json:"tx_hash"`
	Timestamp    time.Time   `json:"timestamp"`
	Success      bool        `json:"success"`
	Error        string      `json:"error,omitempty"`
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
	msg := assetstypes.NewMsgBuyAsset(buyer, symbol, amountNDOLLAR)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid buy message: %w", err)
	}

	// Execute transaction
	resp, err := c.assetsClient.BuyAsset(ctx, msg)
	if err != nil {
		return &TradeResponse{
			Symbol:       symbol,
			Type:         "buy",
			Amount:       amountNDOLLAR,
			Timestamp:    time.Now(),
			Market:       MarketAssets,
			PaymentDenom: assetstypes.NDollarDenom,
			Success:      false,
			Error:        err.Error(),
		}, fmt.Errorf("failed to execute buy transaction: %w", err)
	}

	return &TradeResponse{
		Symbol:       symbol,
		Type:         "buy",
		Amount:       amountNDOLLAR,
		Result:       resp.BaseAmount,
		ResultDenom:  symbol,
		PaymentDenom: assetstypes.NDollarDenom,
		Timestamp:    time.Now(),
		Market:       MarketAssets,
		Success:      true,
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
	msg := assetstypes.NewMsgSellAsset(seller, symbol, baseAmount)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid sell message: %w", err)
	}

	// Execute transaction
	resp, err := c.assetsClient.SellAsset(ctx, msg)
	if err != nil {
		return &TradeResponse{
			Symbol:       symbol,
			Type:         "sell",
			Amount:       baseAmount,
			Timestamp:    time.Now(),
			Market:       MarketAssets,
			PaymentDenom: assetstypes.NDollarDenom,
			Success:      false,
			Error:        err.Error(),
		}, fmt.Errorf("failed to execute sell transaction: %w", err)
	}

	return &TradeResponse{
		Symbol:       symbol,
		Type:         "sell",
		Amount:       baseAmount,
		Result:       resp.Payout_NDOLLAR,
		ResultDenom:  assetstypes.NDollarDenom,
		PaymentDenom: assetstypes.NDollarDenom,
		Timestamp:    time.Now(),
		Market:       MarketAssets,
		Success:      true,
	}, nil
}

// BuyFromCurve executes a buy transaction against the bonding curve module.
func (c *Client) BuyFromCurve(ctx context.Context, trader, denom, paymentDenom, paymentAmount, minTokensOut string) (*TradeResponse, error) {
	if trader == "" {
		return nil, fmt.Errorf("trader address is required")
	}
	if denom == "" {
		return nil, fmt.Errorf("denom is required")
	}
	if strings.TrimSpace(paymentAmount) == "" {
		return nil, fmt.Errorf("payment amount is required")
	}

	if strings.TrimSpace(paymentDenom) == "" {
		paymentDenom = assetstypes.NDollarDenom
	}

	if _, err := sdk.AccAddressFromBech32(trader); err != nil {
		return nil, fmt.Errorf("invalid trader address: %w", err)
	}

	msg := &bondingtypes.MsgBuyFromCurve{
		Trader:        trader,
		Denom:         denom,
		PaymentDenom:  paymentDenom,
		PaymentAmount: paymentAmount,
		MinTokensOut:  minTokensOut,
	}

	resp, err := c.bondingClient.BuyFromCurve(ctx, msg)
	if err != nil {
		return &TradeResponse{
			Symbol:       denom,
			Type:         "buy",
			Amount:       paymentAmount,
			Timestamp:    time.Now(),
			Market:       MarketBondingCurve,
			PaymentDenom: paymentDenom,
			Success:      false,
			Error:        err.Error(),
		}, fmt.Errorf("failed to execute bonding curve buy transaction: %w", err)
	}

	return &TradeResponse{
		Symbol:       denom,
		Type:         "buy",
		Amount:       paymentAmount,
		Result:       resp.TokensOut,
		ResultDenom:  denom,
		PaymentDenom: paymentDenom,
		Timestamp:    time.Now(),
		Market:       MarketBondingCurve,
		Success:      true,
	}, nil
}

// SellToCurve executes a sell transaction against the bonding curve module.
func (c *Client) SellToCurve(ctx context.Context, trader, denom, tokenAmount, paymentDenom, minPaymentOut string) (*TradeResponse, error) {
	if trader == "" {
		return nil, fmt.Errorf("trader address is required")
	}
	if denom == "" {
		return nil, fmt.Errorf("denom is required")
	}
	if strings.TrimSpace(tokenAmount) == "" {
		return nil, fmt.Errorf("token amount is required")
	}

	if strings.TrimSpace(paymentDenom) == "" {
		paymentDenom = assetstypes.NDollarDenom
	}

	if _, err := sdk.AccAddressFromBech32(trader); err != nil {
		return nil, fmt.Errorf("invalid trader address: %w", err)
	}

	msg := &bondingtypes.MsgSellToCurve{
		Trader:        trader,
		Denom:         denom,
		TokenAmount:   tokenAmount,
		PaymentDenom:  paymentDenom,
		MinPaymentOut: minPaymentOut,
	}

	resp, err := c.bondingClient.SellToCurve(ctx, msg)
	if err != nil {
		return &TradeResponse{
			Symbol:       denom,
			Type:         "sell",
			Amount:       tokenAmount,
			Timestamp:    time.Now(),
			Market:       MarketBondingCurve,
			PaymentDenom: paymentDenom,
			Success:      false,
			Error:        err.Error(),
		}, fmt.Errorf("failed to execute bonding curve sell transaction: %w", err)
	}

	return &TradeResponse{
		Symbol:       denom,
		Type:         "sell",
		Amount:       tokenAmount,
		Result:       resp.PaymentOut,
		ResultDenom:  paymentDenom,
		PaymentDenom: paymentDenom,
		Timestamp:    time.Now(),
		Market:       MarketBondingCurve,
		Success:      true,
	}, nil
}

// ExecuteTrade executes a trade based on the request
func (c *Client) ExecuteTrade(ctx context.Context, trader string, req *TradeRequest) (*TradeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("trade request is required")
	}

	market := normalizeMarket(req.Market)

	switch market {
	case MarketAssets:
		switch req.Type {
		case shared.ActionBuy:
			return c.BuyAsset(ctx, trader, req.Symbol, req.Amount)
		case shared.ActionSell:
			return c.SellAsset(ctx, trader, req.Symbol, req.Amount)
		default:
			return nil, fmt.Errorf("invalid trade type: %s, must be 'buy' or 'sell'", req.Type)
		}
	case MarketBondingCurve:
		switch req.Type {
		case "buy":
			return c.BuyFromCurve(ctx, trader, req.Symbol, req.PaymentDenom, req.Amount, req.MinOutput)
		case "sell":
			return c.SellToCurve(ctx, trader, req.Symbol, req.Amount, req.PaymentDenom, req.MinOutput)
		default:
			return nil, fmt.Errorf("invalid trade type: %s, must be 'buy' or 'sell'", req.Type)
		}
	default:
		return nil, fmt.Errorf("invalid trade market: %s", market)
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

	market := normalizeMarket(req.Market)
	switch market {
	case MarketAssets:
		return nil
	case MarketBondingCurve:
		return nil
	default:
		return fmt.Errorf("invalid trade market: %s, must be 'assets' or 'bondingcurve'", req.Market)
	}
}

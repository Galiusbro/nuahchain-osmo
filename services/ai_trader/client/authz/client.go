package authz

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
)

// Client represents an authz client for executing delegated operations
type Client struct {
	conn   *grpc.ClientConn
	client authztypes.MsgClient
}

// NewClient creates a new authz client
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
	client := authztypes.NewMsgClient(conn)

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

// ExecRequest represents a request to execute delegated operations
type ExecRequest struct {
	Grantee string    `json:"grantee"`
	Msgs    []sdk.Msg `json:"msgs"`
}

// ExecResponse represents the response from executing delegated operations
type ExecResponse struct {
	TxHash    string    `json:"tx_hash"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Results   [][]byte  `json:"results"`
}

// ExecuteBuyAsset executes a delegated buy asset operation
func (c *Client) ExecuteBuyAsset(ctx context.Context, grantee string, granter string, symbol string, amountNDOLLAR string) (*ExecResponse, error) {
	if grantee == "" {
		return nil, fmt.Errorf("grantee address is required")
	}
	if granter == "" {
		return nil, fmt.Errorf("granter address is required")
	}
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if amountNDOLLAR == "" {
		return nil, fmt.Errorf("amount is required")
	}

	// Validate addresses
	_, err := sdk.AccAddressFromBech32(grantee)
	if err != nil {
		return nil, fmt.Errorf("invalid grantee address: %w", err)
	}

	_, err = sdk.AccAddressFromBech32(granter)
	if err != nil {
		return nil, fmt.Errorf("invalid granter address: %w", err)
	}

	// Create buy asset message with granter as the signer
	buyMsg := types.NewMsgBuyAsset(granter, symbol, amountNDOLLAR)

	// Validate the message
	if err := buyMsg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid buy message: %w", err)
	}

	// Pack the message into Any
	msgAny, err := codectypes.NewAnyWithValue(buyMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to pack message: %w", err)
	}

	// Create MsgExec
	execMsg := &authztypes.MsgExec{
		Grantee: grantee,
		Msgs:    []*codectypes.Any{msgAny},
	}

	// Execute the transaction
	resp, err := c.client.Exec(ctx, execMsg)
	if err != nil {
		return &ExecResponse{
			Timestamp: time.Now(),
			Success:   false,
			Error:     err.Error(),
		}, fmt.Errorf("failed to execute delegated buy transaction: %w", err)
	}

	return &ExecResponse{
		Timestamp: time.Now(),
		Success:   true,
		Results:   resp.Results,
	}, nil
}

// ExecuteSellAsset executes a delegated sell asset operation
func (c *Client) ExecuteSellAsset(ctx context.Context, grantee string, granter string, symbol string, baseAmount string) (*ExecResponse, error) {
	if grantee == "" {
		return nil, fmt.Errorf("grantee address is required")
	}
	if granter == "" {
		return nil, fmt.Errorf("granter address is required")
	}
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if baseAmount == "" {
		return nil, fmt.Errorf("base amount is required")
	}

	// Validate addresses
	_, err := sdk.AccAddressFromBech32(grantee)
	if err != nil {
		return nil, fmt.Errorf("invalid grantee address: %w", err)
	}

	_, err = sdk.AccAddressFromBech32(granter)
	if err != nil {
		return nil, fmt.Errorf("invalid granter address: %w", err)
	}

	// Create sell asset message with granter as the signer
	sellMsg := types.NewMsgSellAsset(granter, symbol, baseAmount)

	// Validate the message
	if err := sellMsg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid sell message: %w", err)
	}

	// Pack the message into Any
	msgAny, err := codectypes.NewAnyWithValue(sellMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to pack message: %w", err)
	}

	// Create MsgExec
	execMsg := &authztypes.MsgExec{
		Grantee: grantee,
		Msgs:    []*codectypes.Any{msgAny},
	}

	// Execute the transaction
	resp, err := c.client.Exec(ctx, execMsg)
	if err != nil {
		return &ExecResponse{
			Timestamp: time.Now(),
			Success:   false,
			Error:     err.Error(),
		}, fmt.Errorf("failed to execute delegated sell transaction: %w", err)
	}

	return &ExecResponse{
		Timestamp: time.Now(),
		Success:   true,
		Results:   resp.Results,
	}, nil
}

// ExecuteMultipleOperations executes multiple delegated operations in a single transaction
func (c *Client) ExecuteMultipleOperations(ctx context.Context, grantee string, msgs []sdk.Msg) (*ExecResponse, error) {
	if grantee == "" {
		return nil, fmt.Errorf("grantee address is required")
	}
	if len(msgs) == 0 {
		return nil, fmt.Errorf("at least one message is required")
	}

	// Validate grantee address
	_, err := sdk.AccAddressFromBech32(grantee)
	if err != nil {
		return nil, fmt.Errorf("invalid grantee address: %w", err)
	}

	// Pack all messages into Any
	var msgAnys []*codectypes.Any
	for _, msg := range msgs {
		msgAny, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to pack message: %w", err)
		}
		msgAnys = append(msgAnys, msgAny)
	}

	// Create MsgExec
	execMsg := &authztypes.MsgExec{
		Grantee: grantee,
		Msgs:    msgAnys,
	}

	// Execute the transaction
	resp, err := c.client.Exec(ctx, execMsg)
	if err != nil {
		return &ExecResponse{
			Timestamp: time.Now(),
			Success:   false,
			Error:     err.Error(),
		}, fmt.Errorf("failed to execute delegated operations: %w", err)
	}

	return &ExecResponse{
		Timestamp: time.Now(),
		Success:   true,
		Results:   resp.Results,
	}, nil
}

// ValidateExecRequest validates an exec request
func (c *Client) ValidateExecRequest(req *ExecRequest) error {
	if req == nil {
		return fmt.Errorf("exec request is required")
	}

	if req.Grantee == "" {
		return fmt.Errorf("grantee address is required")
	}

	if len(req.Msgs) == 0 {
		return fmt.Errorf("at least one message is required")
	}

	// Validate grantee address
	_, err := sdk.AccAddressFromBech32(req.Grantee)
	if err != nil {
		return fmt.Errorf("invalid grantee address: %w", err)
	}

	// Validate each message
	for i, msg := range req.Msgs {
		if msg == nil {
			return fmt.Errorf("message %d is nil", i)
		}
	}

	return nil
}

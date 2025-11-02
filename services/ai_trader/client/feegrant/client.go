package feegrant

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	feegrant "cosmossdk.io/x/feegrant"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Client wraps the feegrant Msg service.
type Client struct {
	conn   *grpc.ClientConn
	client feegrant.MsgClient
}

// NewClient creates a new feegrant client connected to nodeURL.
func NewClient(nodeURL string) (*Client, error) {
	if strings.TrimSpace(nodeURL) == "" {
		return nil, fmt.Errorf("node URL is required")
	}
	conn, err := grpc.Dial(nodeURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node: %w", err)
	}
	return &Client{conn: conn, client: feegrant.NewMsgClient(conn)}, nil
}

// Close releases the underlying gRPC connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GrantBasicAllowance issues a BasicAllowance, optionally restricting allowed messages.
func (c *Client) GrantBasicAllowance(ctx context.Context, granter, grantee string, spendLimit sdk.Coins, expiration time.Time, allowedMsgs []string) (*feegrant.MsgGrantAllowanceResponse, error) {
	granterAddr, err := sdk.AccAddressFromBech32(granter)
	if err != nil {
		return nil, fmt.Errorf("invalid granter address: %w", err)
	}
	granteeAddr, err := sdk.AccAddressFromBech32(grantee)
	if err != nil {
		return nil, fmt.Errorf("invalid grantee address: %w", err)
	}

	basic := &feegrant.BasicAllowance{}
	if !spendLimit.IsZero() {
		basic.SpendLimit = spendLimit
	}
	if !expiration.IsZero() {
		exp := expiration.UTC()
		basic.Expiration = &exp
	}

	var allowance feegrant.FeeAllowanceI = basic
	if len(allowedMsgs) > 0 {
		filtered, err := feegrant.NewAllowedMsgAllowance(basic, allowedMsgs)
		if err != nil {
			return nil, fmt.Errorf("failed to wrap allowed messages: %w", err)
		}
		allowance = filtered
	}

	msg, err := feegrant.NewMsgGrantAllowance(allowance, granterAddr, granteeAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to construct fee grant: %w", err)
	}

	return c.client.GrantAllowance(ctx, msg)
}

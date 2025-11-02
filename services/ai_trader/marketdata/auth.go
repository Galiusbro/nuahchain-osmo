package marketdata

import (
	"context"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authzclient "github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/authz"
	feegrantclient "github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/feegrant"
)

// API key scopes used for simple RBAC checks on REST endpoints.
const (
	ScopeUserAdmin = "user_admin"
	ScopeUserRead  = "user_read"
	ScopeBotTrade  = "bot_trade"
)

// Grant type constants tracked in persistence.
const (
	GrantTypeAuthzTrade = "authz_trade"
)

// AuthzGrantor abstracts on-chain grant issuance so REST can be tested with fakes.
type AuthzGrantor interface {
	GrantTradingAuthz(ctx context.Context, granter, grantee string, expiration time.Time, includeBonding bool) ([]string, error)
}

// FeeGrantResult captures the summary of a fee grant issuance.
type FeeGrantResult struct {
	SpendLimit      string
	AllowedMessages []string
	Expiration      time.Time
}

// FeeGrantor abstracts feegrant issuance for REST + tests.
type FeeGrantor interface {
	GrantFeeAllowance(ctx context.Context, granter, grantee, spendLimit string, expiration time.Time, allowedMsgs []string) (*FeeGrantResult, error)
}

type remoteAuthzGrantor struct {
	nodeURL string
}

func newRemoteAuthzGrantor(nodeURL string) AuthzGrantor {
	clean := strings.TrimSpace(nodeURL)
	if clean == "" {
		return nil
	}
	return &remoteAuthzGrantor{nodeURL: clean}
}

func (g *remoteAuthzGrantor) GrantTradingAuthz(ctx context.Context, granter, grantee string, expiration time.Time, includeBonding bool) ([]string, error) {
	client, err := authzclient.NewClient(g.nodeURL)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.GrantTradingAuthorizations(ctx, granter, grantee, expiration, includeBonding)
}

type remoteFeeGrantor struct {
	nodeURL string
}

func newRemoteFeeGrantor(nodeURL string) FeeGrantor {
	clean := strings.TrimSpace(nodeURL)
	if clean == "" {
		return nil
	}
	return &remoteFeeGrantor{nodeURL: clean}
}

func (g *remoteFeeGrantor) GrantFeeAllowance(ctx context.Context, granter, grantee, spendLimit string, expiration time.Time, allowedMsgs []string) (*FeeGrantResult, error) {
	client, err := feegrantclient.NewClient(g.nodeURL)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	coins := sdk.NewCoins()
	if strings.TrimSpace(spendLimit) != "" {
		parsed, err := sdk.ParseCoinsNormalized(spendLimit)
		if err != nil {
			return nil, fmt.Errorf("invalid fee spend limit: %w", err)
		}
		coins = parsed
	}
	if _, err := client.GrantBasicAllowance(ctx, granter, grantee, coins, expiration, allowedMsgs); err != nil {
		return nil, err
	}
	return &FeeGrantResult{
		SpendLimit:      coins.String(),
		AllowedMessages: allowedMsgs,
		Expiration:      expiration.UTC(),
	}, nil
}

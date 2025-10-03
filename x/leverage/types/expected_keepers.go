package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetSupply(ctx context.Context, denom string) sdk.Coin
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)
}

// UserTokenKeeper defines the expected interface for the usertoken keeper
type UserTokenKeeper interface {
	// GetTokenPrice returns the current price of a token from bonding curve
	GetTokenPrice(ctx sdk.Context, denom string) (math.LegacyDec, error)

	// GetBondingCurveSupply returns the current supply on the bonding curve
	GetBondingCurveSupply(ctx sdk.Context, denom string) (math.Int, error)

	// CalculateTokensFromPayment calculates tokens received for a payment
	CalculateTokensFromPayment(ctx sdk.Context, denom string, currentSupply math.Int, paymentAmount math.Int) math.Int

	// CalculatePayoutFromTokens calculates payout for selling tokens
	CalculatePayoutFromTokens(ctx sdk.Context, denom string, currentSupply math.Int, tokensToSell math.Int) math.Int

	// ExecuteBuyTokens executes a token purchase through bonding curve
	ExecuteBuyTokens(ctx sdk.Context, buyer sdk.AccAddress, denom string, paymentAmount math.Int, paymentDenom string) (math.Int, error)

	// ExecuteSellTokens executes a token sale through bonding curve
	ExecuteSellTokens(ctx sdk.Context, seller sdk.AccAddress, denom string, tokenAmount math.Int) (math.Int, string, error)

	// CheckTokenExists verifies if a user token exists
	CheckTokenExists(ctx sdk.Context, denom string) bool

	// GetTokenSupply returns the current total supply of a token
	GetTokenSupply(ctx sdk.Context, denom string) (math.Int, error)
}

// PriceOracle defines the interface for getting token prices
type PriceOracle interface {
	// GetPrice returns the current price of a token
	GetPrice(ctx sdk.Context, denom string) (math.LegacyDec, error)

	// UpdatePrice updates the price of a token (called by hooks)
	UpdatePrice(ctx sdk.Context, denom string, price math.LegacyDec) error
}

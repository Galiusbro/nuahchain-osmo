package types

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/osmosis-labs/osmosis/osmomath"

	usdoracletypes "github.com/osmosis-labs/osmosis/v30/x/usdoracle/types"
)

// AccountKeeper defines the expected interface for the account module
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
}

// BankKeeper defines the expected interface for the bank module
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetSupply(ctx context.Context, denom string) sdk.Coin
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
}

// USDOracleKeeper defines the expected interface for the USD Oracle module
type USDOracleKeeper interface {
	GetTokenPriceForExchange(ctx context.Context, denom string) (usdoracletypes.TokenPrice, bool)
	GetParams(ctx context.Context) usdoracletypes.Params
	IsTokenSupported(ctx context.Context, denom string) bool
}

// TWAPKeeper defines the expected interface for the TWAP module
type TWAPKeeper interface {
	GetArithmeticTwap(ctx sdk.Context, poolId uint64, baseAsset, quoteAsset string, startTime, endTime time.Time) (osmomath.Dec, error)
	GetGeometricTwap(ctx sdk.Context, poolId uint64, baseAsset, quoteAsset string, startTime, endTime time.Time) (osmomath.Dec, error)
}

// DistributionKeeper defines the expected interface for the distribution module
type DistributionKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}
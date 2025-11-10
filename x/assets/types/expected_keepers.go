package types

import (
	"context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	osmomath "github.com/osmosis-labs/osmosis/osmomath"
	oracletypes "github.com/osmosis-labs/osmosis/v30/x/oracle/types"
	stablecointypes "github.com/osmosis-labs/osmosis/v30/x/stablecoin/types"
)

// BankKeeper defines the expected methods of the bank keeper.
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// OracleKeeper defines the expected methods from the oracle keeper.
type OracleKeeper interface {
	GetPrice(ctx sdk.Context, symbol string) (*oracletypes.Price, bool)
	// GetPriceWithFallback should attempt to fetch price from external API if not in store
	GetPriceWithFallback(ctx sdk.Context, symbol string) (*oracletypes.Price, bool)
	// EnsureFreshPrice must fetch and persist the latest price from the oracle backend
	// before returning it. Implementations should return an error if the price cannot be refreshed.
	EnsureFreshPrice(ctx sdk.Context, symbol string) (*oracletypes.Price, error)
}

// FeesKeeper defines the expected methods from the fees keeper.
type FeesKeeper interface {
	GetTradeFeeRate(ctx sdk.Context) osmomath.Dec
}

// StablecoinKeeper defines the expected methods from the stablecoin keeper.
type StablecoinKeeper interface {
	RecordMint(ctx sdk.Context, amount sdkmath.Int) error
	RecordBurn(ctx sdk.Context, amount sdkmath.Int) error
	GetStats(ctx sdk.Context) stablecointypes.Stats
	GetNDollarDenom(ctx sdk.Context) string
}

package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/osmosis-labs/osmosis/v30/x/mint/types"

	usdoracletypes "github.com/osmosis-labs/osmosis/v30/x/usdoracle/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetAccountsBalances(ctx context.Context) []banktypes.Balance
	GetSupply(ctx context.Context, denom string) sdk.Coin
	GetPaginatedTotalSupply(ctx context.Context, pagination *querytypes.PageRequest) (sdk.Coins, *querytypes.PageResponse, error)
	IterateAccountBalances(ctx context.Context, addr sdk.AccAddress, cb func(coin sdk.Coin) (stop bool))
	IterateAllBalances(ctx context.Context, cb func(address sdk.AccAddress, coin sdk.Coin) (stop bool))
	IterateTotalSupply(ctx context.Context, cb func(coin sdk.Coin) bool)
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)
	IterateAllDenomMetaData(ctx context.Context, cb func(banktypes.Metadata) bool)
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

// MintKeeper defines the expected interface needed to mint/burn tokens.
type MintKeeper interface {
	GetParams(ctx sdk.Context) minttypes.Params
	SetParams(ctx sdk.Context, params minttypes.Params)
	GetMinter(ctx sdk.Context) minttypes.Minter
	SetMinter(ctx sdk.Context, minter minttypes.Minter)
}

// USDOracleKeeper defines the expected interface needed to get USD prices.
type USDOracleKeeper interface {
	GetCurrentPrice(ctx sdk.Context) (usdoracletypes.USDPrice, bool)
	GetPriceHistoryList(ctx sdk.Context, limit uint32) []usdoracletypes.USDPrice
	CalculatePriceDeviation(ctx sdk.Context) (math.LegacyDec, bool)
	IsWithinThreshold(ctx sdk.Context) bool
}

package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v30/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v30/x/tokenfactory/types"
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
	SendCoinsFromModuleToModule(ctx context.Context, senderModule string, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)
}

// TokenFactoryKeeper defines the expected interface for the tokenfactory keeper
type TokenFactoryKeeper interface {
	CreateDenom(ctx context.Context, creatorAddr string, subdenom string) (newTokenDenom string, err error)
	Mint(ctx context.Context, creatorAddr string, coin sdk.Coin, mintToAddress string) error
	Burn(ctx context.Context, creatorAddr string, coin sdk.Coin, burnFromAddress string) error
	ChangeAdmin(ctx context.Context, denom string, newAdmin string, creatorAddr string) error
	GetAuthorityMetadata(ctx context.Context, denom string) (types.DenomAuthorityMetadata, error)
	SetAuthorityMetadata(ctx context.Context, denom string, metadata types.DenomAuthorityMetadata) error
}

// GammKeeper defines the expected interface for the gamm keeper
type GammKeeper interface {
	GetPool(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolI, error)
}

// PoolManagerKeeper defines the expected interface for the poolmanager keeper
type PoolManagerKeeper interface {
	CreatePool(ctx sdk.Context, msg poolmanagertypes.CreatePoolMsg) (uint64, error)
	GetPool(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolI, error)
}

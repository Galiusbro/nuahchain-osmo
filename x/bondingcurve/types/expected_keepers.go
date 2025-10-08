package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v30/x/poolmanager/types"
	usertokentypes "github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, sender sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipient sdk.AccAddress, amt sdk.Coins) error
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

type UserTokenKeeper interface {
	GetToken(ctx sdk.Context, denom string) (usertokentypes.Token, bool)
	UpdateToken(ctx sdk.Context, token usertokentypes.Token) error
}

type PoolManagerKeeper interface {
	CreatePool(ctx sdk.Context, msg poolmanagertypes.CreatePoolMsg) (uint64, error)
}

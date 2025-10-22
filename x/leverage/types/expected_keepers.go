package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	oracletypes "github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

// BankKeeper defines the expected bank keeper.
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// OracleKeeper defines expected oracle methods.
type OracleKeeper interface {
	GetPrice(ctx sdk.Context, symbol string) (*oracletypes.Price, bool)
}

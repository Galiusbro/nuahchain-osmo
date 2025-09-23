package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	rolestypes "github.com/osmosis-labs/osmosis/v30/x/roles/types"
)

// BankKeeper defines the subset of bank keeper methods needed by the treasury module.
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, sender sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipient sdk.AccAddress, amt sdk.Coins) error
}

// RolesKeeper defines the interface for role-based authorization.
type RolesKeeper interface {
	HasRole(ctx sdk.Context, addr sdk.AccAddress, role rolestypes.Role) bool
}

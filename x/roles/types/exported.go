package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// RolesKeeper defines the expected behaviour from the roles keeper when used by other modules.
type RolesKeeper interface {
	HasRole(ctx sdk.Context, addr sdk.AccAddress, role Role) bool
}

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	rolestypes "github.com/osmosis-labs/osmosis/v30/x/roles/types"
)

// RolesKeeper defines the expected interface for role gating.
type RolesKeeper interface {
	HasRole(ctx sdk.Context, addr sdk.AccAddress, role rolestypes.Role) bool
}

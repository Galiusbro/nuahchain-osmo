package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	policytypes "github.com/osmosis-labs/osmosis/v30/x/policy/types"
	rolestypes "github.com/osmosis-labs/osmosis/v30/x/roles/types"
)

// RolesKeeper defines the interface needed from the roles module.
type RolesKeeper interface {
	HasRole(ctx sdk.Context, addr sdk.AccAddress, role rolestypes.Role) bool
}

// PolicyKeeper defines the interface used to query policies.
type PolicyKeeper interface {
	GetPolicy(ctx sdk.Context, policyID uint64) (policytypes.Policy, bool)
}

// TreasuryKeeper defines the interface used to execute payouts.
type TreasuryKeeper interface {
	DisburseClaim(ctx sdk.Context, poolID string, recipient sdk.AccAddress, amount sdk.Coin) error
}

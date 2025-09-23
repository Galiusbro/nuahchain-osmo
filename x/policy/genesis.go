package policy

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/policy/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/policy/types"
)

// InitGenesis initializes module state from genesis data.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState *types.GenesisState) {
	if genState == nil {
		genState = types.DefaultGenesis()
	}

	k.SetParams(ctx, genState.Params)

	if genState.NextPolicyId == 0 {
		genState.NextPolicyId = 1
	}
	k.SetNextPolicyID(ctx, genState.NextPolicyId)

	for _, policy := range genState.Policies {
		k.SetPolicy(ctx, policy)
	}
}

// ExportGenesis exports module state for genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params := k.GetParams(ctx)
	policies := k.GetAllPolicies(ctx)
	nextID := k.GetNextPolicyID(ctx)

	return &types.GenesisState{
		Params:       params,
		NextPolicyId: nextID,
		Policies:     policies,
	}
}

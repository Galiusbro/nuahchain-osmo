package roles

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/roles/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/roles/types"
)

// InitGenesis configures state from genesis data.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState *types.GenesisState) {
	if genState == nil {
		genState = types.DefaultGenesis()
	}

	params := genState.Params
	if params.Authority == "" {
		params.Authority = k.GetAuthority(ctx)
	}

	k.SetParams(ctx, params)

	for _, binding := range genState.Bindings {
		if err := k.SetRoleBinding(ctx, binding); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis produces genesis data from state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params := k.GetParams(ctx)
	bindings := k.GetAllRoleBindings(ctx)

	return &types.GenesisState{
		Params:   params,
		Bindings: bindings,
	}
}

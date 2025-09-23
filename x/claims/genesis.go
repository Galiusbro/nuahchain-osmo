package claims

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/claims/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/claims/types"
)

// InitGenesis initializes module state from genesis data.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState *types.GenesisState) {
	if genState == nil {
		genState = types.DefaultGenesis()
	}

	k.SetParams(ctx, genState.Params)

	if genState.NextClaimId == 0 {
		genState.NextClaimId = 1
	}
	k.SetNextClaimID(ctx, genState.NextClaimId)

	for _, claim := range genState.Claims {
		k.SetClaim(ctx, claim)
	}
}

// ExportGenesis exports the current module state into a genesis structure.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:      k.GetParams(ctx),
		NextClaimId: k.GetNextClaimID(ctx),
		Claims:      k.ExportClaims(ctx),
	}
}

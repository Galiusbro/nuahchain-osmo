package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
)

// InitGenesis sets the initial state for the assets module.
func (k Keeper) InitGenesis(ctx sdk.Context, state *types.GenesisState) {
	if state == nil {
		return
	}

	for _, asset := range state.Assets {
		if asset == nil {
			continue
		}
		k.SetAsset(ctx, asset)
	}
}

// ExportGenesis exports the current state as genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	assets := k.GetAllAssets(ctx)
	return types.NewGenesisState(assets)
}

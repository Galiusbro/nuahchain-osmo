package softpeg

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	softpegkeeper "github.com/osmosis-labs/osmosis/v30/x/softpeg/keeper"
	softpegtypes "github.com/osmosis-labs/osmosis/v30/x/softpeg/types"
)

// InitGenesis initializes the softpeg module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k softpegkeeper.Keeper, genState softpegtypes.GenesisState) {
	// Set all the price data
	for _, elem := range genState.PriceDataList {
		k.SetPriceData(ctx, elem)
	}

	// Set all the community metrics
	for _, elem := range genState.CommunityMetricsList {
		k.SetCommunityMetrics(ctx, elem)
	}

	// Set peg config if provided
	if genState.PegConfig != nil {
		k.SetPegConfig(ctx, *genState.PegConfig)
	}

	// Set params
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the softpeg module's exported genesis.
func ExportGenesis(ctx sdk.Context, k softpegkeeper.Keeper) *softpegtypes.GenesisState {
	genesis := softpegtypes.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	// Export all price data
	genesis.PriceDataList = k.GetAllPriceData(ctx)

	// Export community metrics
	if metrics, found := k.GetLatestCommunityMetrics(ctx); found {
		genesis.CommunityMetricsList = []softpegtypes.CommunityMetrics{metrics}
	}

	// Export peg config
	if config, found := k.GetPegConfig(ctx); found {
		genesis.PegConfig = &config
	}

	return genesis
}

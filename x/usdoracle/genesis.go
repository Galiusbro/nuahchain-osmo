package usdoracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/usdoracle/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/usdoracle/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set module parameters
	k.SetParams(ctx, genState.Params)

	// Set current price if provided
	if genState.CurrentPrice != nil {
		k.SetCurrentPrice(ctx, *genState.CurrentPrice)
	}

	// Set price history
	for _, price := range genState.PriceHistory {
		k.AddPriceHistory(ctx, price)
	}

	// Set price sources
	for _, source := range genState.PriceSources {
		k.SetPriceSource(ctx, source)
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	genesis := types.DefaultGenesis()

	// Export parameters
	genesis.Params = k.GetParams(ctx)

	// Export current price
	if currentPrice, found := k.GetCurrentPrice(ctx); found {
		genesis.CurrentPrice = &currentPrice
	}

	// Export price history (limited to recent entries)
	genesis.PriceHistory = k.GetPriceHistoryList(ctx, 1000) // Last 1000 entries

	// Export price sources
	genesis.PriceSources = k.GetAllPriceSources(ctx)

	return *genesis
}

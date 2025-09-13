package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/exchange/types"
)

// InitGenesis initializes the exchange module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// Set module parameters
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	// Set exchange rates
	for _, exchangeRate := range genState.ExchangeRates {
		if err := k.SetExchangeRate(ctx, exchangeRate); err != nil {
			panic(err)
		}
	}

	// Set daily limits
	for _, dailyLimit := range genState.DailyLimits {
		if err := k.SetDailyLimit(ctx, dailyLimit); err != nil {
			panic(err)
		}
	}

	// Log initialization
	k.Logger().Info("Exchange module genesis initialized",
		"params", genState.Params,
		"exchange_rates_count", len(genState.ExchangeRates),
		"daily_limits_count", len(genState.DailyLimits),
	)
}

// ExportGenesis returns the exchange module's exported genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// Get module parameters
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	// Get all exchange rates
	exchangeRates, err := k.GetAllExchangeRates(ctx)
	if err != nil {
		panic(err)
	}

	// Get all daily limits - simplified for now
	dailyLimits := []types.DailyLimit{}

	// Log export
	k.Logger().Info("Exchange module genesis exported",
		"params", params,
		"exchange_rates_count", len(exchangeRates),
		"daily_limits_count", len(dailyLimits),
	)

	return &types.GenesisState{
		Params:        params,
		ExchangeRates: exchangeRates,
		DailyLimits:   dailyLimits,
	}
}
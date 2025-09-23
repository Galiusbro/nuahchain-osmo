package treasury

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/treasury/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/treasury/types"
)

// InitGenesis initializes module state from genesis data.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState *types.GenesisState) {
	if genState == nil {
		genState = types.DefaultGenesis()
	}

	k.SetParams(ctx, genState.Params)

	for _, pool := range genState.Pools {
		k.SetGenesisPool(ctx, pool)
	}

	for _, balance := range genState.Balances {
		coin := sdk.NewCoin(balance.Balance.Denom, balance.Balance.Amount)
		k.SetGenesisPoolBalance(ctx, balance.PoolId, coin)
	}

	for _, reserve := range genState.Reserves {
		k.SetGenesisReserve(ctx, reserve)
	}
}

// ExportGenesis exports module state to genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:   k.GetParams(ctx),
		Pools:    k.ExportPools(ctx),
		Balances: k.ExportBalances(ctx),
		Reserves: k.ExportReserves(ctx),
	}
}

package keeper

import (
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	k.SetParams(ctx, genState.Params)

	for _, pool := range genState.Pools {
		k.setPool(ctx, pool)
	}

	for _, marginPool := range genState.MarginPools {
		k.setMarginPool(ctx, marginPool)
	}

	for _, position := range genState.MarginPositions {
		k.setMarginPosition(ctx, position)
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	pools := []types.BondingCurvePool{}
	marginPools := []types.MarginPool{}
	marginPositions := []types.MarginPosition{}

	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.TokenPoolKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var pool types.BondingCurvePool
		k.cdc.MustUnmarshal(iterator.Value(), &pool)
		pools = append(pools, pool)
	}

	marginPoolIter := storetypes.KVStorePrefixIterator(store, types.MarginPoolKeyPrefix)
	defer marginPoolIter.Close()

	for ; marginPoolIter.Valid(); marginPoolIter.Next() {
		var marginPool types.MarginPool
		k.cdc.MustUnmarshal(marginPoolIter.Value(), &marginPool)
		marginPools = append(marginPools, marginPool)
	}

	positionIter := storetypes.KVStorePrefixIterator(store, types.MarginPositionKeyPrefix)
	defer positionIter.Close()

	for ; positionIter.Valid(); positionIter.Next() {
		var position types.MarginPosition
		k.cdc.MustUnmarshal(positionIter.Value(), &position)
		marginPositions = append(marginPositions, position)
	}

	return &types.GenesisState{
		Params:          params,
		Pools:           pools,
		MarginPools:     marginPools,
		MarginPositions: marginPositions,
	}
}

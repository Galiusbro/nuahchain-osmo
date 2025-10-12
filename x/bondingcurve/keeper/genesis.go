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

	if genState.GlobalPause != nil {
		k.setGlobalPause(ctx, *genState.GlobalPause)
	}

	for _, pause := range genState.TokenPauses {
		k.setTokenPause(ctx, pause.Denom, pause.Info)
	}

	for _, freeze := range genState.FreezeEntries {
		k.setFreezeInfo(ctx, freeze.TargetType, freeze.Target, freeze.Info)
	}

	if genState.PendingParams != nil {
		k.setPendingParams(ctx, *genState.PendingParams)
	}

	for _, action := range genState.EmergencyActions {
		k.setEmergencyAction(ctx, action)
	}

	if genState.EmergencyActionSeq != 0 {
		k.setEmergencyActionSeq(ctx, genState.EmergencyActionSeq)
	}

	if genState.EmergencyConfig != nil {
		k.setEmergencyConfig(ctx, *genState.EmergencyConfig)
	}

	if genState.ModuleStats != nil {
		k.setModuleStats(ctx, *genState.ModuleStats)
	}
	for _, ts := range genState.TokenStats {
		tsCopy := ts
		k.setTokenStats(ctx, tsCopy)
	}
	if len(genState.LiquidationRecords) > 0 {
		k.setLiquidationRecordsFromGenesis(ctx, genState.LiquidationRecords, genState.LiquidationSeq)
	} else if genState.LiquidationSeq != 0 {
		store := k.getStore(ctx)
		store.Set(types.LiquidationSeqKey, sdk.Uint64ToBigEndian(genState.LiquidationSeq))
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	pools := []types.BondingCurvePool{}
	marginPools := []types.MarginPool{}
	marginPositions := []types.MarginPosition{}
	var globalPause *types.PauseInfo
	tokenPauses := []types.TokenPauseEntry{}
	freezeEntries := []types.FreezeEntry{}
	var pendingParams *types.PendingParams
	emergencyActions := []types.EmergencyAction{}
	var emergencyConfig *types.EmergencyConfig
	emergencySeq := k.getEmergencyActionSeq(ctx)

	store := k.getStore(ctx)
	moduleStats := k.getModuleStats(ctx)
	tokenStats := k.getAllTokenStats(ctx)
	liquidationRecords := k.getAllLiquidationRecords(ctx)
	liquidationSeq := sdk.BigEndianToUint64(store.Get(types.LiquidationSeqKey))
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

	if info, found := k.getGlobalPause(ctx); found {
		globalPause = &info
	}

	pauseIter := storetypes.KVStorePrefixIterator(store, types.TokenPauseKeyPrefix)
	defer pauseIter.Close()
	for ; pauseIter.Valid(); pauseIter.Next() {
		var info types.PauseInfo
		k.cdc.MustUnmarshal(pauseIter.Value(), &info)
		denom := string(pauseIter.Key())
		tokenPauses = append(tokenPauses, types.TokenPauseEntry{Denom: denom, Info: info})
	}

	freezeIter := storetypes.KVStorePrefixIterator(store, types.FreezeKeyPrefix)
	defer freezeIter.Close()
	for ; freezeIter.Valid(); freezeIter.Next() {
		var info types.FreezeInfo
		k.cdc.MustUnmarshal(freezeIter.Value(), &info)
		key := freezeIter.Key()
		if len(key) == 0 {
			continue
		}
		typeKey := types.FreezeTargetType(key[0])
		target := string(key[1:])
		freezeEntries = append(freezeEntries, types.FreezeEntry{TargetType: typeKey, Target: target, Info: info})
	}

	if pending, found := k.getPendingParams(ctx); found {
		pendingCopy := pending
		pendingParams = &pendingCopy
	}

	k.iterateEmergencyActions(ctx, 0, func(action types.EmergencyAction) bool {
		emergencyActions = append(emergencyActions, action)
		return false
	})

	config := k.getEmergencyConfig(ctx)
	emergencyConfig = &config

return &types.GenesisState{
	Params:             params,
	Pools:              pools,
	MarginPools:        marginPools,
	MarginPositions:    marginPositions,
	GlobalPause:        globalPause,
	TokenPauses:        tokenPauses,
	FreezeEntries:      freezeEntries,
	PendingParams:      pendingParams,
	EmergencyActions:   emergencyActions,
	EmergencyConfig:    emergencyConfig,
	EmergencyActionSeq: emergencySeq,
	ModuleStats:        &moduleStats,
	TokenStats:         tokenStats,
	LiquidationRecords: liquidationRecords,
	LiquidationSeq:     liquidationSeq,
}
}

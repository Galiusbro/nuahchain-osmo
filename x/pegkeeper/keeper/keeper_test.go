package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"

	"github.com/osmosis-labs/osmosis/v30/x/pegkeeper/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/pegkeeper/types"
)

func TestKeeperBasic(t *testing.T) {
	// Setup basic test environment
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	transientStoreKey := storetypes.NewTransientStoreKey("transient_test")
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(transientStoreKey, storetypes.StoreTypeTransient, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	paramsSubspace := paramtypes.NewSubspace(cdc,
		codec.NewLegacyAmino(),
		storeKey,
		transientStoreKey,
		"pegkeeper",
	)

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		paramsSubspace,
		"authority",
		nil, // bankKeeper
		nil, // mintKeeper
		nil, // usdOracleKeeper
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize with default params first
	defaultParams := types.DefaultParams()
	k.SetParams(ctx, defaultParams)

	// Test GetParams
	params := k.GetParams(ctx)
	require.NotNil(t, params)
	require.Equal(t, defaultParams, params)

	// Test SetParams with new values
	newParams := types.DefaultParams()
	newParams.Enabled = false // Change something
	k.SetParams(ctx, newParams)

	retrievedParams := k.GetParams(ctx)
	require.Equal(t, newParams, retrievedParams)
	require.False(t, retrievedParams.Enabled)

	// Test GetAuthority
	authority := k.GetAuthority()
	require.Equal(t, "authority", authority)
}
package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v30/x/fees/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/fees/types"
)

func setupKeeper(t *testing.T) (keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	paramsKey := storetypes.NewKVStoreKey(paramtypes.StoreKey)
	transientKey := storetypes.NewTransientStoreKey(paramtypes.TStoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(paramsKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(transientKey, storetypes.StoreTypeTransient, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	paramSubspace := paramtypes.NewSubspace(cdc, legacy.Cdc, paramsKey, transientKey, types.ModuleName)

	k := keeper.NewKeeper(cdc, paramSubspace)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	return k, ctx
}

func TestKeeperParams(t *testing.T) {
	k, ctx := setupKeeper(t)

	params := types.NewParams("0.015")
	require.NoError(t, k.SetParams(ctx, params))

	stored := k.GetParams(ctx)
	require.Equal(t, params, stored)
	require.Equal(t, params.TradeFeeRateDec(), k.GetTradeFeeRate(ctx))
}

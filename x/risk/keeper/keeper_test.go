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
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/risk/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/risk/types"
)

const authority = "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"

func setupKeeper(t *testing.T) (keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	k := keeper.NewKeeper(cdc, storeKey, authority)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	return k, ctx
}

func TestSetAndGetRiskParams(t *testing.T) {
	k, ctx := setupKeeper(t)

	params := &types.RiskParams{
		Symbol:            "BTC",
		MaxLeverage:       "5.0",
		MaintenanceMargin: "0.1",
		InitialMargin:     "0.2",
	}

	require.NoError(t, k.SetRiskParams(ctx, params))

	stored, found := k.GetRiskParams(ctx, "btc")
	require.True(t, found)
	require.Equal(t, params, stored)
}

func TestInitAndExportGenesis(t *testing.T) {
	k, ctx := setupKeeper(t)

	gen := &types.GenesisState{
		RiskParams: []*types.RiskParams{
			{
				Symbol:            "ETH",
				MaxLeverage:       "3",
				MaintenanceMargin: "0.15",
				InitialMargin:     "0.25",
			},
		},
	}

	k.InitGenesis(ctx, gen)

	out := k.ExportGenesis(ctx)
	require.Len(t, out.RiskParams, 1)
	require.Equal(t, "ETH", out.RiskParams[0].Symbol)
}

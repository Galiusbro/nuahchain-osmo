package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
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

	"github.com/osmosis-labs/osmosis/v30/x/stablecoin/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/stablecoin/types"
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

	k := keeper.NewKeeper(cdc, storeKey, mockBankKeeper{}, paramSubspace)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}

type mockBankKeeper struct{}

func (mockBankKeeper) GetBalance(context.Context, sdk.AccAddress, string) sdk.Coin {
	return sdk.NewCoin(types.NDollarDenom, sdkmath.ZeroInt())
}

func TestRecordMintBurnAndStats(t *testing.T) {
	k, ctx := setupKeeper(t)

	require.NoError(t, k.RecordMint(ctx, sdkmath.NewInt(150)))
	require.NoError(t, k.RecordBurn(ctx, sdkmath.NewInt(40)))
	require.NoError(t, k.RecordMint(ctx, sdkmath.NewInt(10)))
	require.NoError(t, k.RecordBurn(ctx, sdkmath.NewInt(5)))

	stats := k.GetStats(ctx)
	minted, ok := sdkmath.NewIntFromString(stats.TotalMinted)
	require.True(t, ok)
	burned, ok := sdkmath.NewIntFromString(stats.TotalBurned)
	require.True(t, ok)
	outstanding, ok := sdkmath.NewIntFromString(stats.Outstanding)
	require.True(t, ok)

	require.Equal(t, sdkmath.NewInt(160), minted)
	require.Equal(t, sdkmath.NewInt(45), burned)
	require.Equal(t, sdkmath.NewInt(115), outstanding)
}

func TestParamsManagement(t *testing.T) {
	k, ctx := setupKeeper(t)

	params := types.NewParams("custom")
	k.SetParams(ctx, params)

	stored := k.GetParams(ctx)
	require.Equal(t, params, stored)
}

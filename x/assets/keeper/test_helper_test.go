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
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	osmomath "github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v30/x/assets/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
	oracletypes "github.com/osmosis-labs/osmosis/v30/x/oracle/types"
	stablecointypes "github.com/osmosis-labs/osmosis/v30/x/stablecoin/types"
)

type dummyBankKeeper struct{}

func (dummyBankKeeper) SendCoinsFromAccountToModule(_ context.Context, _ sdk.AccAddress, _ string, _ sdk.Coins) error {
	return nil
}

func (dummyBankKeeper) SendCoinsFromModuleToModule(_ context.Context, _ string, _ string, _ sdk.Coins) error {
	return nil
}

func (dummyBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, _ string, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}

func (dummyBankKeeper) MintCoins(_ context.Context, _ string, _ sdk.Coins) error {
	return nil
}

func (dummyBankKeeper) BurnCoins(_ context.Context, _ string, _ sdk.Coins) error {
	return nil
}

func (dummyBankKeeper) GetBalance(_ context.Context, _ sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, sdkmath.ZeroInt())
}

type dummyOracleKeeper struct{}

func (dummyOracleKeeper) GetPrice(_ sdk.Context, _ string) (*oracletypes.Price, bool) {
	return nil, false
}

type dummyFeesKeeper struct{}

func (dummyFeesKeeper) GetTradeFeeRate(_ sdk.Context) osmomath.Dec {
	return osmomath.ZeroDec()
}

type dummyStablecoinKeeper struct{}

func (dummyStablecoinKeeper) RecordMint(_ sdk.Context, _ sdkmath.Int) error { return nil }
func (dummyStablecoinKeeper) RecordBurn(_ sdk.Context, _ sdkmath.Int) error { return nil }
func (dummyStablecoinKeeper) GetStats(_ sdk.Context) stablecointypes.Stats {
	return stablecointypes.NewStats(sdkmath.ZeroInt(), sdkmath.ZeroInt())
}

func setupKeeper(t *testing.T) (keeper.Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(cdc, storeKey, dummyBankKeeper{}, dummyOracleKeeper{}, dummyFeesKeeper{}, dummyStablecoinKeeper{})
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	return k, ctx
}

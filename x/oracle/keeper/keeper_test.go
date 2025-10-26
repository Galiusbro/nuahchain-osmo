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
	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

func TestKeeperSetGetPrice(t *testing.T) {
	k, ctx := setupKeeper(t)

	price := &types.Price{
		Symbol: "GOLD",
		Value:  "2000",
	}

	k.SetPrice(ctx, price)

	got, found := k.GetPrice(ctx, "GOLD")
	require.True(t, found)
	require.Equal(t, price, got)

	_, found = k.GetPrice(ctx, "SILVER")
	require.False(t, found)
}

func TestAPIKeeperRealPrice(t *testing.T) {
	// Create APIKeeper
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	priv := secp256k1.GenPrivKey()
	authority := sdk.AccAddress(priv.PubKey().Address()).String()

	// Create APIKeeper instead of regular Keeper
	apiKeeper := keeper.NewAPIKeeper(cdc, storeKey, authority)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	t.Log("🚀 Testing real API price fetching...")

	// Test fetching real price from Yahoo Finance
	symbol := "AAPL"
	t.Logf("📊 Fetching real price for %s...", symbol)

	err := apiKeeper.UpdatePriceFromAPI(ctx, symbol)
	require.NoError(t, err, "Should successfully fetch price from API")

	// Verify price was stored
	price, found := apiKeeper.GetPrice(ctx, symbol)
	require.True(t, found, "Price should be found in store")
	require.NotNil(t, price, "Price should not be nil")
	require.Equal(t, symbol, price.Symbol, "Symbol should match")
	require.Greater(t, price.Value, "0", "Price should be positive")
	require.Equal(t, "yahoo_finance", price.Source, "Source should be yahoo_finance")
	require.Equal(t, float32(1.0), price.Confidence, "Confidence should be 1.0")

	t.Logf("✅ Successfully fetched %s: $%s (source: %s, confidence: %.2f)",
		symbol, price.Value, price.Source, price.Confidence)

	// Test that price is accessible via regular GetPrice method
	regularPrice, found := apiKeeper.GetPrice(ctx, symbol)
	require.True(t, found, "Price should be accessible via regular GetPrice")
	require.Equal(t, price.Value, regularPrice.Value, "Price values should match")

	t.Log("✅ Real API price test completed successfully!")
}

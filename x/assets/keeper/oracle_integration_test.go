package keeper_test

import (
	"testing"
	"time"

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
	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	osmomath "github.com/osmosis-labs/osmosis/osmomath"

	"github.com/osmosis-labs/osmosis/v30/x/assets/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
	oraclekeeper "github.com/osmosis-labs/osmosis/v30/x/oracle/keeper"
	oracletypes "github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

// RealOracleKeeper implements OracleKeeper interface with real API calls
type RealOracleKeeper struct {
	apiKeeper *oraclekeeper.APIKeeper
}

func NewRealOracleKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, authority string) *RealOracleKeeper {
	return &RealOracleKeeper{
		apiKeeper: oraclekeeper.NewAPIKeeper(cdc, storeKey, authority),
	}
}

func (r *RealOracleKeeper) GetPrice(ctx sdk.Context, symbol string) (*oracletypes.Price, bool) {
	// First try to get from store
	price, found := r.apiKeeper.GetPrice(ctx, symbol)
	if found {
		return price, true
	}

	// If not found, try to fetch from API
	err := r.apiKeeper.UpdatePriceFromAPI(ctx, symbol)
	if err != nil {
		return nil, false
	}

	// Try to get from store again
	price, found = r.apiKeeper.GetPrice(ctx, symbol)
	return price, found
}

func (r *RealOracleKeeper) GetPriceWithFallback(ctx sdk.Context, symbol string) (*oracletypes.Price, bool) {
	return r.apiKeeper.GetPriceWithFallback(ctx, symbol)
}

func (r *RealOracleKeeper) EnsureFreshPrice(ctx sdk.Context, symbol string) (*oracletypes.Price, error) {
	return r.apiKeeper.EnsureFreshPrice(ctx, symbol)
}

func setupKeeperWithRealOracle(t *testing.T) (keeper.Keeper, sdk.Context, *RealOracleKeeper) {
	t.Helper()

	// Create stores
	assetsStoreKey := storetypes.NewKVStoreKey(types.StoreKey)
	oracleStoreKey := storetypes.NewKVStoreKey(oracletypes.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(assetsStoreKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(oracleStoreKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Create real oracle keeper
	priv := secp256k1.GenPrivKey()
	authority := sdk.AccAddress(priv.PubKey().Address()).String()
	realOracleKeeper := NewRealOracleKeeper(cdc, oracleStoreKey, authority)

	// Create assets keeper with real oracle
	assetsKeeper := keeper.NewKeeper(cdc, assetsStoreKey, dummyBankKeeper{}, realOracleKeeper, dummyFeesKeeper{}, dummyStablecoinKeeper{})
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	return assetsKeeper, ctx, realOracleKeeper
}

func TestAssetsWithRealOracle(t *testing.T) {
	assetsKeeper, ctx, oracleKeeper := setupKeeperWithRealOracle(t)

	t.Log("🚀 Testing Assets module with real Oracle...")

	// Test 1: Check if oracle can fetch real prices
	t.Run("OraclePriceFetching", func(t *testing.T) {
		symbol := "AAPL"
		t.Logf("📊 Fetching real price for %s...", symbol)

		price, found := oracleKeeper.GetPrice(ctx, symbol)
		require.True(t, found, "Should successfully fetch price from oracle")
		require.NotNil(t, price, "Price should not be nil")
		require.Equal(t, symbol, price.Symbol, "Symbol should match")
		require.Greater(t, price.Value, "0", "Price should be positive")
		require.Equal(t, "yahoo_finance", price.Source, "Source should be yahoo_finance")

		t.Logf("✅ Oracle fetched %s: $%s (source: %s, confidence: %.2f)",
			symbol, price.Value, price.Source, price.Confidence)
	})

	// Test 2: Test Assets module getPrice method
	t.Run("AssetsGetPrice", func(t *testing.T) {
		symbol := "AAPL"
		t.Logf("📊 Testing Assets getPrice for %s...", symbol)

		price, err := assetsKeeper.GetPrice(ctx, symbol)
		require.NoError(t, err, "Assets should successfully get price from oracle")
		require.NotNil(t, price, "Price should not be nil")
		require.Equal(t, symbol, price.Symbol, "Symbol should match")
		require.Greater(t, price.Value, "0", "Price should be positive")

		t.Logf("✅ Assets got price for %s: $%s", symbol, price.Value)
	})

	// Test 3: Test BuyAsset with real price
	t.Run("BuyAssetWithRealPrice", func(t *testing.T) {
		// Create a test account
		priv := secp256k1.GenPrivKey()
		buyer := sdk.AccAddress(priv.PubKey().Address())

		symbol := "AAPL"                    // Use a valid symbol without special characters
		amountND := sdkmath.NewInt(1000000) // 1 NDOLLAR (assuming 6 decimals)

		t.Logf("💰 Testing BuyAsset: %s for %s NDOLLAR...", symbol, amountND.String())

		// This will fail because we don't have real bank operations, but we can test the price fetching
		_, _, err := assetsKeeper.BuyAsset(ctx, buyer, symbol, amountND)

		// We expect this to fail due to bank operations, but the price fetching should work
		if err != nil {
			t.Logf("⚠️ BuyAsset failed (expected due to dummy bank): %v", err)
			// Check if the error is related to price fetching or bank operations
			require.Contains(t, err.Error(), "bank", "Error should be related to bank operations, not price fetching")
		}

		t.Log("✅ Price fetching in BuyAsset worked correctly")
	})

	// Test 4: Test SellAsset with real price
	t.Run("SellAssetWithRealPrice", func(t *testing.T) {
		// Create a test account
		priv := secp256k1.GenPrivKey()
		seller := sdk.AccAddress(priv.PubKey().Address())

		symbol := "AAPL"                                          // Use a valid symbol without special characters
		baseAmount := osmomath.NewDecFromInt(sdkmath.NewInt(100)) // 100 units

		t.Logf("💰 Testing SellAsset: %s for %s units...", symbol, baseAmount.String())

		// This will fail because we don't have real bank operations, but we can test the price fetching
		_, _, err := assetsKeeper.SellAsset(ctx, seller, symbol, baseAmount)

		// We expect this to fail due to insufficient balance, but the price fetching should work
		if err != nil {
			t.Logf("⚠️ SellAsset failed (expected due to insufficient balance): %v", err)
			// Check if the error is related to balance or price fetching
			require.Contains(t, err.Error(), "insufficient", "Error should be related to insufficient balance, not price fetching")
		}

		t.Log("✅ Price fetching in SellAsset worked correctly")
	})

	// Test 5: Test multiple symbols
	t.Run("MultipleSymbols", func(t *testing.T) {
		symbols := []string{"AAPL", "MSFT", "BTC-USD", "EURUSD=X", "GC=F"}

		t.Log("📊 Testing multiple symbols...")
		for _, symbol := range symbols {
			t.Logf("   Testing %s...", symbol)

			price, found := oracleKeeper.GetPrice(ctx, symbol)
			if found {
				t.Logf("   ✅ %s: $%s", symbol, price.Value)
			} else {
				t.Logf("   ⚠️ %s: price not found", symbol)
			}

			// Add delay to respect rate limits
			time.Sleep(1 * time.Second)
		}
	})

	// Test 6: Test price consistency
	t.Run("PriceConsistency", func(t *testing.T) {
		symbol := "AAPL"

		// Get price multiple times to ensure consistency
		price1, found1 := oracleKeeper.GetPrice(ctx, symbol)
		require.True(t, found1, "First price fetch should succeed")

		time.Sleep(1 * time.Second) // Small delay

		price2, found2 := oracleKeeper.GetPrice(ctx, symbol)
		require.True(t, found2, "Second price fetch should succeed")

		// Prices should be the same (cached) or very close (if updated)
		t.Logf("Price 1: $%s", price1.Value)
		t.Logf("Price 2: $%s", price2.Value)

		// They should be the same symbol and source
		require.Equal(t, price1.Symbol, price2.Symbol, "Symbols should match")
		require.Equal(t, price1.Source, price2.Source, "Sources should match")

		t.Log("✅ Price consistency test passed")
	})

	t.Log("✅ Assets with real Oracle integration test completed successfully!")
}

func TestOracleIntegrationFlow(t *testing.T) {
	assetsKeeper, ctx, oracleKeeper := setupKeeperWithRealOracle(t)

	t.Log("🔄 Testing complete Oracle integration flow...")

	// Test the complete flow: Oracle -> Assets -> Price retrieval
	t.Run("CompleteFlow", func(t *testing.T) {
		symbol := "AAPL"

		// Step 1: Oracle fetches price
		t.Log("Step 1: Oracle fetching price...")
		oraclePrice, found := oracleKeeper.GetPrice(ctx, symbol)
		require.True(t, found, "Oracle should fetch price")
		t.Logf("   Oracle price: $%s", oraclePrice.Value)

		// Step 2: Assets module gets price through its interface
		t.Log("Step 2: Assets module getting price...")
		assetsPrice, err := assetsKeeper.GetPrice(ctx, symbol)
		require.NoError(t, err, "Assets should get price successfully")
		t.Logf("   Assets price: $%s", assetsPrice.Value)

		// Step 3: Verify prices match
		require.Equal(t, oraclePrice.Value, assetsPrice.Value, "Prices should match")
		require.Equal(t, oraclePrice.Symbol, assetsPrice.Symbol, "Symbols should match")
		require.Equal(t, oraclePrice.Source, assetsPrice.Source, "Sources should match")

		t.Log("✅ Complete flow test passed - Oracle and Assets are properly integrated!")
	})

	t.Log("✅ Oracle integration flow test completed!")
}

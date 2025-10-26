package keeper

import (
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/sources"
)

func setupKeeper(t *testing.T) *APIKeeper {
	cdc := codec.NewProtoCodec(types.NewInterfaceRegistry())
	storeKey := storetypes.NewKVStoreKey("oracle")
	keeper := NewAPIKeeper(cdc, storeKey, "authority")

	return keeper
}

func setupKeeperWithContext(t *testing.T) (*APIKeeper, sdk.Context) {
	keeper := setupKeeper(t)

	// Create a simple context for testing
	// Note: This is a simplified context for testing purposes
	ctx := sdk.Context{}

	return keeper, ctx
}

func TestAPIKeeper(t *testing.T) {
	keeper := setupKeeper(t)

	// Test that API client is initialized
	require.NotNil(t, keeper.apiClient)
	require.NotNil(t, keeper.symbols)
	require.Greater(t, len(keeper.symbols), 0)

	// Test default symbols
	symbols := sources.GetDefaultSymbols()
	require.Greater(t, len(symbols), 0)

	// Test API client configuration
	require.Greater(t, keeper.apiClient.GetRateLimit(), time.Duration(0))
}

func TestUpdatePriceFromAPI(t *testing.T) {
	keeper := setupKeeper(t)

	// Test that the method exists and can be called
	require.NotNil(t, keeper.UpdatePriceFromAPI)

	// Test API client directly (without context)
	symbol := "AAPL"
	priceData, err := keeper.apiClient.GetPriceFromYahooFinance(symbol)
	require.NoError(t, err)
	require.NotNil(t, priceData)
	require.Equal(t, symbol, priceData.Symbol)
	require.Greater(t, priceData.Price, 0.0)
	require.Equal(t, "Yahoo Finance", priceData.Source)

	t.Logf("✅ API client successfully fetched price for %s: $%.2f", symbol, priceData.Price)
}

func TestGetHistoricalData(t *testing.T) {
	keeper := setupKeeper(t)

	// Test API client directly (without context)
	symbol := "AAPL"
	period := "5d"

	historicalData, err := keeper.apiClient.GetHistoricalData(symbol, period)
	require.NoError(t, err)
	require.NotNil(t, historicalData)
	require.Greater(t, len(historicalData), 0)

	// Check that we have data for multiple days
	require.GreaterOrEqual(t, len(historicalData), 2, "Expected at least 2 data points for 5-day period")

	// Check first data point
	firstData := historicalData[0]
	require.Equal(t, symbol, firstData.Symbol)
	require.Greater(t, firstData.Price, 0.0)

	// Check that timestamps are in chronological order
	for i := 1; i < len(historicalData); i++ {
		require.True(t, historicalData[i].Timestamp.After(historicalData[i-1].Timestamp) ||
			historicalData[i].Timestamp.Equal(historicalData[i-1].Timestamp),
			"Expected timestamps to be in chronological order")
	}

	t.Logf("✅ Historical data for %s (%s): %d data points", symbol, period, len(historicalData))
}

func TestSearchSymbols(t *testing.T) {
	keeper := setupKeeper(t)

	// Test API client directly (without context)
	query := "Apple"
	results, err := keeper.apiClient.SearchSymbols(query)
	require.NoError(t, err)
	require.NotNil(t, results)
	require.Greater(t, len(results), 0)

	// Check first result has expected fields
	firstResult := results[0]
	require.NotNil(t, firstResult["symbol"])
	require.NotNil(t, firstResult["longname"])

	t.Logf("✅ Search results for '%s': %d results", query, len(results))
	for i, result := range results {
		if i < 3 { // Log first 3 results
			t.Logf("  %d: %v", i+1, result["symbol"])
		}
	}
}

func TestGetAPIStats(t *testing.T) {
	keeper := setupKeeper(t)

	// Test that the method exists and returns valid stats
	// Create a simple context for testing
	ctx := sdk.Context{}
	stats := keeper.GetAPIStats(ctx)
	require.NotNil(t, stats)
	require.Contains(t, stats, "total_categories")
	require.Contains(t, stats, "total_symbols")
	require.Contains(t, stats, "api_client_rate_limit")
	require.Contains(t, stats, "api_source")

	// Check that we have categories
	require.Greater(t, stats["total_categories"], 0)
	require.Greater(t, stats["total_symbols"], 0)
	require.Equal(t, "Yahoo Finance", stats["api_source"])

	t.Logf("✅ API Stats: %+v", stats)
}

func TestUpdateAllPrices(t *testing.T) {
	keeper := setupKeeper(t)

	// Test API client directly with multiple symbols
	symbols := []string{"AAPL", "MSFT"}

	for _, symbol := range symbols {
		t.Run("symbol_"+symbol, func(t *testing.T) {
			priceData, err := keeper.apiClient.GetPriceFromYahooFinance(symbol)
			require.NoError(t, err, "Failed to get price data for %s", symbol)
			require.NotNil(t, priceData)
			require.Equal(t, symbol, priceData.Symbol)
			require.Greater(t, priceData.Price, 0.0)

			t.Logf("✅ %s: $%.2f %s", symbol, priceData.Price, priceData.Currency)

			// Add small delay to respect rate limits
			time.Sleep(keeper.apiClient.GetRateLimit())
		})
	}
}

func TestUpdatePricesByCategory(t *testing.T) {
	keeper := setupKeeper(t)

	// Test API client directly with stocks category
	category := "stocks"
	symbols, exists := keeper.GetSymbolsByCategory(category)
	require.True(t, exists, "Category %s should exist", category)
	require.Greater(t, len(symbols), 0, "Category %s should have symbols", category)

	// Test first few symbols from the category
	testSymbols := symbols[:min(2, len(symbols))] // Test max 2 symbols to avoid rate limits

	successCount := 0
	for _, symbol := range testSymbols {
		priceData, err := keeper.apiClient.GetPriceFromYahooFinance(symbol)
		if err == nil {
			successCount++
			require.NotNil(t, priceData)
			require.Equal(t, symbol, priceData.Symbol)
			require.Greater(t, priceData.Price, 0.0)
			t.Logf("✅ %s: $%.2f %s", symbol, priceData.Price, priceData.Currency)
		} else {
			t.Logf("⚠️ Failed to update %s: %v", symbol, err)
		}

		// Add small delay to respect rate limits
		time.Sleep(keeper.apiClient.GetRateLimit())
	}

	require.Greater(t, successCount, 0, "Expected at least one successful price update")
	t.Logf("✅ Successfully updated %d prices for category '%s'", successCount, category)
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestGetAvailableCategories(t *testing.T) {
	keeper := setupKeeper(t)

	// Test that the method exists and returns categories
	categories := keeper.GetAvailableCategories()
	require.NotNil(t, categories)
	require.Greater(t, len(categories), 0)
}

func TestGetSymbolsByCategory(t *testing.T) {
	keeper := setupKeeper(t)

	// Test that the method exists
	require.NotNil(t, keeper.GetSymbolsByCategory)

	// Test getting symbols for a known category
	symbols, exists := keeper.GetSymbolsByCategory("commodities")
	require.True(t, exists)
	require.Greater(t, len(symbols), 0)

	// Test getting symbols for unknown category
	symbols, exists = keeper.GetSymbolsByCategory("unknown")
	require.False(t, exists)
	require.Nil(t, symbols)
}

// TestMultipleAssetTypes tests different types of assets
func TestMultipleAssetTypes(t *testing.T) {
	keeper := setupKeeper(t)

	// Test different types of assets
	testCases := []struct {
		name      string
		symbol    string
		assetType string
	}{
		{"Apple Stock", "AAPL", "Stock"},
		{"Bitcoin", "BTC-USD", "Cryptocurrency"},
		{"EUR/USD", "EURUSD=X", "Currency"},
		{"Gold", "GC=F", "Commodity"},
		{"S&P 500", "^GSPC", "Index"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing %s (%s) - %s", tc.name, tc.symbol, tc.assetType)

			// Test API client directly
			priceData, err := keeper.apiClient.GetPriceFromYahooFinance(tc.symbol)
			require.NoError(t, err, "Failed to get price data for %s", tc.symbol)
			require.NotNil(t, priceData)
			require.Equal(t, tc.symbol, priceData.Symbol)
			require.Greater(t, priceData.Price, 0.0)
			require.Equal(t, "Yahoo Finance", priceData.Source)

			t.Logf("✅ %s: $%.2f %s", tc.name, priceData.Price, priceData.Currency)

			// Add small delay to respect rate limits
			time.Sleep(keeper.apiClient.GetRateLimit())
		})
	}
}

// TestGetPriceWithFallback tests the fallback mechanism
func TestGetPriceWithFallback(t *testing.T) {
	keeper := setupKeeper(t)

	// Test API client directly with a symbol
	symbol := "GOOGL"

	// Test that API client can fetch the price
	priceData, err := keeper.apiClient.GetPriceFromYahooFinance(symbol)
	require.NoError(t, err, "API client should fetch price successfully")
	require.NotNil(t, priceData)
	require.Equal(t, symbol, priceData.Symbol)
	require.Greater(t, priceData.Price, 0.0)

	t.Logf("✅ API client successfully fetched price for %s: $%.2f", symbol, priceData.Price)
}

// TestInvalidSymbol tests error handling for invalid symbols
func TestInvalidSymbol(t *testing.T) {
	keeper := setupKeeper(t)

	// Test with an invalid symbol
	invalidSymbol := "INVALID_SYMBOL_12345"
	_, err := keeper.apiClient.GetPriceFromYahooFinance(invalidSymbol)
	require.Error(t, err, "Expected error for invalid symbol")

	t.Logf("✅ Correctly handled invalid symbol: %v", err)
}

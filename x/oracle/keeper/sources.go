package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/sources"
	oracletypes "github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

// APIKeeper extends the base Keeper with API capabilities
type APIKeeper struct {
	Keeper
	apiClient *sources.APIClient
	symbols   map[string][]string
}

// NewAPIKeeper creates a new APIKeeper instance
func NewAPIKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, authority string) *APIKeeper {
	return &APIKeeper{
		Keeper:    NewKeeper(cdc, storeKey, authority),
		apiClient: sources.NewAPIClient(),
		symbols:   sources.GetDefaultSymbols(),
	}
}

// UpdatePriceFromAPI fetches price from Yahoo Finance API and updates the store
func (k *APIKeeper) UpdatePriceFromAPI(ctx sdk.Context, symbol string) error {
	symbol = EnsureSymbol(symbol)
	if symbol == "" {
		return fmt.Errorf("invalid symbol")
	}

	// Fetch price from Yahoo Finance API
	priceData, err := k.apiClient.GetPriceFromYahooFinance(symbol)
	if err != nil {
		return fmt.Errorf("failed to fetch price from API: %w", err)
	}

	// Create price object with metadata
	price := &oracletypes.Price{
		Symbol:     symbol,
		Value:      fmt.Sprintf("%.8f", priceData.Price),
		Source:     "yahoo_finance",
		Timestamp:  priceData.Timestamp.Unix(),
		Confidence: 1.0, // API data is highly reliable
	}

	// Store the price
	k.SetPrice(ctx, price)

	k.Logger(ctx).Info("Updated price from API",
		"symbol", symbol,
		"price", price.Value,
		"source", priceData.Source,
		"exchange", priceData.Exchange,
		"confidence", price.Confidence,
	)

	return nil
}

// GetHistoricalData fetches historical price data from API
func (k *APIKeeper) GetHistoricalData(ctx sdk.Context, symbol string, period string) ([]sources.PriceData, error) {
	symbol = EnsureSymbol(symbol)
	if symbol == "" {
		return nil, fmt.Errorf("invalid symbol")
	}

	// Fetch historical data from Yahoo Finance API
	historicalData, err := k.apiClient.GetHistoricalData(symbol, period)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical data from API: %w", err)
	}

	return historicalData, nil
}

// SearchSymbols searches for symbols using API
func (k *APIKeeper) SearchSymbols(ctx sdk.Context, query string) ([]map[string]interface{}, error) {
	// Search symbols using Yahoo Finance API
	results, err := k.apiClient.SearchSymbols(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search symbols: %w", err)
	}

	return results, nil
}

// GetPriceWithFallback tries to get price from store, falls back to API if not found
func (k *APIKeeper) GetPriceWithFallback(ctx sdk.Context, symbol string) (*oracletypes.Price, bool) {
	// First try to get from store
	price, found := k.GetPrice(ctx, symbol)
	if found {
		return price, true
	}

	// If not found, try to fetch from API
	err := k.UpdatePriceFromAPI(ctx, symbol)
	if err != nil {
		k.Logger(ctx).Error("Failed to update price from API", "error", err, "symbol", symbol)
		return nil, false
	}

	// Try to get from store again
	price, found = k.GetPrice(ctx, symbol)
	return price, found
}

// UpdateAllPrices updates prices for all configured symbols
func (k *APIKeeper) UpdateAllPrices(ctx sdk.Context, symbols []string) map[string]error {
	results := make(map[string]error)

	for _, symbol := range symbols {
		err := k.UpdatePriceFromAPI(ctx, symbol)
		if err != nil {
			results[symbol] = err
		}
	}

	return results
}

// UpdatePricesByCategory updates prices for all symbols in a specific category
func (k *APIKeeper) UpdatePricesByCategory(ctx sdk.Context, category string) map[string]error {
	results := make(map[string]error)

	symbols, exists := k.symbols[category]
	if !exists {
		results["error"] = fmt.Errorf("category %s not found", category)
		return results
	}

	for _, symbol := range symbols {
		err := k.UpdatePriceFromAPI(ctx, symbol)
		if err != nil {
			results[symbol] = err
		}
	}

	return results
}

// GetAPIStats returns statistics about API performance
func (k *APIKeeper) GetAPIStats(ctx sdk.Context) map[string]interface{} {
	stats := make(map[string]interface{})

	stats["total_categories"] = len(k.symbols)
	stats["total_symbols"] = 0

	for category, symbols := range k.symbols {
		stats[category+"_count"] = len(symbols)
		if total, ok := stats["total_symbols"].(int); ok {
			stats["total_symbols"] = total + len(symbols)
		} else {
			stats["total_symbols"] = len(symbols)
		}
	}

	stats["api_client_rate_limit"] = k.apiClient.GetRateLimit().String()
	stats["api_source"] = "Yahoo Finance"

	return stats
}

// GetAvailableCategories returns list of available symbol categories
func (k *APIKeeper) GetAvailableCategories() []string {
	categories := make([]string, 0, len(k.symbols))
	for category := range k.symbols {
		categories = append(categories, category)
	}
	return categories
}

// GetSymbolsByCategory returns symbols for a specific category
func (k *APIKeeper) GetSymbolsByCategory(category string) ([]string, bool) {
	symbols, exists := k.symbols[category]
	return symbols, exists
}

// GetDefaultSymbols returns all default symbols
func (k *APIKeeper) GetDefaultSymbols() map[string][]string {
	return k.symbols
}

package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v30/x/usdoracle/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		paramstore paramtypes.Subspace
		authority  string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	authority string,
) *Keeper {
	// set KeyTable if it has not already been set
	// Check if subspace is valid before using HasKeyTable
	if ps.Name() != "" && !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramstore: ps,
		authority:  authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetCurrentPrice sets the current USD price
func (k Keeper) SetCurrentPrice(ctx sdk.Context, price types.USDPrice) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&price)
	store.Set(types.CurrentPriceKey, bz)
}

// GetCurrentPrice gets the current USD price
func (k Keeper) GetCurrentPrice(ctx sdk.Context) (types.USDPrice, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CurrentPriceKey)
	if bz == nil {
		return types.USDPrice{}, false
	}

	var price types.USDPrice
	k.cdc.MustUnmarshal(bz, &price)
	return price, true
}

// AddPriceHistory adds a price entry to the history
func (k Keeper) AddPriceHistory(ctx sdk.Context, price types.USDPrice) {
	store := ctx.KVStore(k.storeKey)
	timestamp := price.Timestamp.Unix()
	key := types.GetPriceHistoryKey(timestamp)
	bz := k.cdc.MustMarshal(&price)
	store.Set(key, bz)
}

// GetPriceHistoryList gets price history with pagination
func (k Keeper) GetPriceHistoryList(ctx sdk.Context, limit uint32) []types.USDPrice {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PriceHistoryPrefix)
	defer iterator.Close()

	var prices []types.USDPrice
	count := uint32(0)
	for ; iterator.Valid() && count < limit; iterator.Next() {
		var price types.USDPrice
		k.cdc.MustUnmarshal(iterator.Value(), &price)
		prices = append(prices, price)
		count++
	}

	return prices
}

// SetPriceSource sets a price source configuration
func (k Keeper) SetPriceSource(ctx sdk.Context, source types.PriceSource) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPriceSourceKey(source.Name)
	bz := k.cdc.MustMarshal(&source)
	store.Set(key, bz)
}

// GetPriceSource gets a price source configuration
func (k Keeper) GetPriceSource(ctx sdk.Context, name string) (types.PriceSource, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPriceSourceKey(name)
	bz := store.Get(key)
	if bz == nil {
		return types.PriceSource{}, false
	}

	var source types.PriceSource
	k.cdc.MustUnmarshal(bz, &source)
	return source, true
}

// GetAllPriceSources gets all price source configurations
func (k Keeper) GetAllPriceSources(ctx sdk.Context) []types.PriceSource {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PriceSourcesPrefix)
	defer iterator.Close()

	var sources []types.PriceSource
	for ; iterator.Valid(); iterator.Next() {
		var source types.PriceSource
		k.cdc.MustUnmarshal(iterator.Value(), &source)
		sources = append(sources, source)
	}

	return sources
}

// CalculatePriceDeviation calculates the deviation from USD peg (1.0)
func (k Keeper) CalculatePriceDeviation(ctx sdk.Context) (math.LegacyDec, bool) {
	price, found := k.GetCurrentPrice(ctx)
	if !found {
		return math.LegacyDec{}, false
	}

	// Calculate deviation from 1.0 USD
	one := math.LegacyOneDec()
	deviation := price.Price.Sub(one).Abs()
	return deviation, true
}

// IsWithinThreshold checks if current price is within acceptable threshold
func (k Keeper) IsWithinThreshold(ctx sdk.Context) bool {
	params := k.GetParams(ctx)
	deviation, found := k.CalculatePriceDeviation(ctx)
	if !found {
		return false
	}

	return deviation.LTE(params.PriceDeviationThreshold)
}

// UpdatePriceFromSources aggregates prices from multiple sources and updates current price
func (k Keeper) UpdatePriceFromSources(ctx sdk.Context, sourcePrices []types.USDPrice) error {
	if len(sourcePrices) == 0 {
		return fmt.Errorf("no source prices provided")
	}

	// Get all price sources for weights
	sources := k.GetAllPriceSources(ctx)
	sourceWeights := make(map[string]math.LegacyDec)
	for _, source := range sources {
		if source.Enabled {
			sourceWeights[source.Name] = source.Weight
		}
	}

	// Calculate weighted average
	var weightedSum math.LegacyDec
	var totalWeight math.LegacyDec

	for _, price := range sourcePrices {
		if weight, exists := sourceWeights[price.Source]; exists {
			weightedSum = weightedSum.Add(price.Price.Mul(weight))
			totalWeight = totalWeight.Add(weight)
		}
	}

	if totalWeight.IsZero() {
		return fmt.Errorf("no valid sources with weights")
	}

	weightedPrice := weightedSum.Quo(totalWeight)

	// Create new price entry
	newPrice := types.USDPrice{
		Price:       weightedPrice,
		Timestamp:   time.Now(),
		Source:      "aggregated",
		BlockHeight: ctx.BlockHeight(),
	}

	// Update current price and add to history
	k.SetCurrentPrice(ctx, newPrice)
	k.AddPriceHistory(ctx, newPrice)

	return nil
}

// GetTokenPriceForExchange returns the price for a specific token (for Exchange module)
func (k Keeper) GetTokenPriceForExchange(ctx context.Context, denom string) (types.TokenPrice, bool) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// For now, return a mock implementation
	// In a real implementation, this would query token-specific prices
	currentPrice, found := k.GetCurrentPrice(sdkCtx)
	if !found {
		return types.TokenPrice{}, false
	}

	tokenPrice := types.TokenPrice{
		Denom:     denom,
		Price:     currentPrice.Price,
		Timestamp: currentPrice.Timestamp,
	}

	return tokenPrice, true
}



// IsTokenSupported checks if a token is supported for price feeds
func (k Keeper) IsTokenSupported(ctx context.Context, denom string) bool {
	params := k.GetParams(ctx)

	// Check if token is in supported tokens list
	for _, token := range params.SupportedTokens {
		if token.Denom == denom {
			return true
		}
	}
	return false
}

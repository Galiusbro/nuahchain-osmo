package keeper

import (
	"context"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

// Keeper manages access to oracle state.
type Keeper struct {
	cdc       codec.BinaryCodec
	storeKey  storetypes.StoreKey
	authority string
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, authority string) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// Authority returns the keeper authority address.
func (k Keeper) Authority() string {
	return k.authority
}

// StoreKey returns the keeper store key.
func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.storeKey
}

// SetPrice sets the price for a given symbol.
func (k Keeper) SetPrice(ctx sdk.Context, price *types.Price) {
	if price == nil {
		return
	}
	cleanSymbol := EnsureSymbol(price.Symbol)
	if cleanSymbol == "" {
		return
	}

	priceCopy := *price
	priceCopy.Symbol = cleanSymbol

	store := ctx.KVStore(k.storeKey)
	store.Set(types.PriceKey(priceCopy.Symbol), k.cdc.MustMarshal(&priceCopy))
}

// GetPrice retrieves a price by symbol.
func (k Keeper) GetPrice(ctx sdk.Context, symbol string) (*types.Price, bool) {
	symbol = EnsureSymbol(symbol)
	if symbol == "" {
		return nil, false
	}
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PriceKey(symbol))
	if bz == nil {
		return nil, false
	}

	var price types.Price
	k.cdc.MustUnmarshal(bz, &price)
	return &price, true
}

// GetPriceWithContext retrieves a price by symbol using context.Context.
func (k Keeper) GetPriceWithContext(ctx context.Context, symbol string) (*types.Price, bool) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return k.GetPrice(sdkCtx, symbol)
}

// SetPriceHistory stores a price history entry
func (k Keeper) SetPriceHistory(ctx sdk.Context, entry *types.PriceHistoryEntry) {
	if entry == nil {
		return
	}
	cleanSymbol := EnsureSymbol(entry.Symbol)
	if cleanSymbol == "" {
		return
	}

	entryCopy := *entry
	entryCopy.Symbol = cleanSymbol

	store := ctx.KVStore(k.storeKey)
	store.Set(types.PriceHistoryKey(entryCopy.Symbol, entryCopy.Timestamp), k.cdc.MustMarshal(&entryCopy))
}

// GetPriceHistory retrieves price history for a symbol within a time range
func (k Keeper) GetPriceHistory(ctx sdk.Context, symbol string, startTime, endTime int64, limit int32) ([]*types.PriceHistoryEntry, error) {
	symbol = EnsureSymbol(symbol)
	if symbol == "" {
		return nil, fmt.Errorf("invalid symbol")
	}

	store := ctx.KVStore(k.storeKey)
	prefix := types.PriceHistoryPrefix(symbol)
	iterator := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	defer iterator.Close()

	var entries []*types.PriceHistoryEntry
	count := int32(0)

	for ; iterator.Valid() && count < limit; iterator.Next() {
		var entry types.PriceHistoryEntry
		k.cdc.MustUnmarshal(iterator.Value(), &entry)

		// Filter by time range
		if entry.Timestamp >= startTime && entry.Timestamp <= endTime {
			entries = append(entries, &entry)
			count++
		}
	}

	return entries, nil
}

// GetLatestPriceHistory retrieves the most recent price history entries for a symbol
func (k Keeper) GetLatestPriceHistory(ctx sdk.Context, symbol string, limit int32) ([]*types.PriceHistoryEntry, error) {
	symbol = EnsureSymbol(symbol)
	if symbol == "" {
		return nil, fmt.Errorf("invalid symbol")
	}

	store := ctx.KVStore(k.storeKey)
	prefix := types.PriceHistoryPrefix(symbol)
	iterator := store.ReverseIterator(prefix, storetypes.PrefixEndBytes(prefix))
	defer iterator.Close()

	var entries []*types.PriceHistoryEntry
	count := int32(0)

	for ; iterator.Valid() && count < limit; iterator.Next() {
		var entry types.PriceHistoryEntry
		k.cdc.MustUnmarshal(iterator.Value(), &entry)
		entries = append(entries, &entry)
		count++
	}

	return entries, nil
}

// EnsureSymbol sanitizes the provided symbol string.
func EnsureSymbol(symbol string) string {
	return strings.TrimSpace(symbol)
}

// GetAllPrices returns every stored price entry.
func (k Keeper) GetAllPrices(ctx sdk.Context) []*types.Price {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PriceKeyPrefix)
	defer iterator.Close()

	var prices []*types.Price
	for ; iterator.Valid(); iterator.Next() {
		var price types.Price
		k.cdc.MustUnmarshal(iterator.Value(), &price)
		priceCopy := price
		prices = append(prices, &priceCopy)
	}

	return prices
}

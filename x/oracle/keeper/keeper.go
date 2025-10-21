package keeper

import (
	"context"
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

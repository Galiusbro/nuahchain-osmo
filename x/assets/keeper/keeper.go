package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v30/x/assets/types"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
)

// Keeper provides methods for managing assets state.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

// NewKeeper returns a new Keeper instance.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// Logger returns a module-scoped logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetAsset stores or updates the provided asset.
func (k Keeper) SetAsset(ctx sdk.Context, asset *types.Asset) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(asset)
	store.Set(types.AssetKey(asset.Symbol), bz)
}

// GetAsset retrieves an asset by symbol.
func (k Keeper) GetAsset(ctx sdk.Context, symbol string) (*types.Asset, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.AssetKey(symbol))
	if bz == nil {
		return nil, false
	}

	var asset types.Asset
	k.cdc.MustUnmarshal(bz, &asset)
	return &asset, true
}

// GetAllAssets returns all stored assets.
func (k Keeper) GetAllAssets(ctx sdk.Context) []*types.Asset {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.AssetKeyPrefix)
	defer iterator.Close()

	var assets []*types.Asset
	for ; iterator.Valid(); iterator.Next() {
		var asset types.Asset
		k.cdc.MustUnmarshal(iterator.Value(), &asset)
		assetCopy := asset // avoid referencing loop variable
		assets = append(assets, &assetCopy)
	}

	return assets
}

// RemoveAsset deletes an asset by symbol.
func (k Keeper) RemoveAsset(ctx sdk.Context, symbol string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.AssetKey(symbol))
}

// EnsureAsset returns an existing asset or creates a new default entry if missing.
func (k Keeper) EnsureAsset(ctx sdk.Context, symbol string) (*types.Asset, bool, error) {
	cleanSymbol := strings.TrimSpace(symbol)
	if cleanSymbol == "" {
		return nil, false, fmt.Errorf("symbol cannot be empty")
	}

	if asset, found := k.GetAsset(ctx, cleanSymbol); found {
		return asset, false, nil
	}

	newAsset := &types.Asset{
		Symbol:   cleanSymbol,
		Name:     cleanSymbol,
		Type:     "unknown",
		Decimals: 2,
		Status:   "active",
	}

	k.SetAsset(ctx, newAsset)

	return newAsset, true, nil
}

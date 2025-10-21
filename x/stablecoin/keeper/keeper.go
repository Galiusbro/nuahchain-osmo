package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/stablecoin/types"

	storetypes "cosmossdk.io/store/types"
)

// Keeper provides access to stablecoin statistics.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

// NewKeeper creates a new stablecoin keeper instance.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// RecordMint increases the total minted counter.
func (k Keeper) RecordMint(ctx sdk.Context, amount sdkmath.Int) error {
	if amount.IsNegative() {
		return fmt.Errorf("mint amount cannot be negative")
	}
	if amount.IsZero() {
		return nil
	}

	current := k.getTotalMinted(ctx)
	newAmount := current.Add(amount)
	k.setTotalMinted(ctx, newAmount)
	return nil
}

// RecordBurn increases the total burned counter.
func (k Keeper) RecordBurn(ctx sdk.Context, amount sdkmath.Int) error {
	if amount.IsNegative() {
		return fmt.Errorf("burn amount cannot be negative")
	}
	if amount.IsZero() {
		return nil
	}

	current := k.getTotalBurned(ctx)
	newAmount := current.Add(amount)
	k.setTotalBurned(ctx, newAmount)
	return nil
}

// GetStats returns the current stablecoin statistics.
func (k Keeper) GetStats(ctx sdk.Context) types.Stats {
	minted := k.getTotalMinted(ctx)
	burned := k.getTotalBurned(ctx)
	return types.NewStats(minted, burned)
}

// InitGenesis initializes module state from genesis data.
func (k Keeper) InitGenesis(ctx sdk.Context, state *types.GenesisState) {
	stats := state.GetStats()
	if stats == nil {
		stats = &types.Stats{}
	}

	minted := parseInt(stats.TotalMinted)
	burned := parseInt(stats.TotalBurned)

	k.setTotalMinted(ctx, minted)
	k.setTotalBurned(ctx, burned)
}

// ExportGenesis exports current module state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	stats := k.GetStats(ctx)
	return &types.GenesisState{Stats: &stats}
}

func (k Keeper) getTotalMinted(ctx sdk.Context) sdkmath.Int {
	return k.getInt(ctx, types.TotalMintedKey)
}

func (k Keeper) setTotalMinted(ctx sdk.Context, amount sdkmath.Int) {
	k.setInt(ctx, types.TotalMintedKey, amount)
}

func (k Keeper) getTotalBurned(ctx sdk.Context) sdkmath.Int {
	return k.getInt(ctx, types.TotalBurnedKey)
}

func (k Keeper) setTotalBurned(ctx sdk.Context, amount sdkmath.Int) {
	k.setInt(ctx, types.TotalBurnedKey, amount)
}

func (k Keeper) getInt(ctx sdk.Context, key []byte) sdkmath.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)
	if bz == nil {
		return sdkmath.ZeroInt()
	}
	val, ok := sdkmath.NewIntFromString(string(bz))
	if !ok {
		panic("invalid stored integer value")
	}
	return val
}

func (k Keeper) setInt(ctx sdk.Context, key []byte, amount sdkmath.Int) {
	store := ctx.KVStore(k.storeKey)
	store.Set(key, []byte(amount.String()))
}

func parseInt(str string) sdkmath.Int {
	if str == "" {
		return sdkmath.ZeroInt()
	}
	val, ok := sdkmath.NewIntFromString(str)
	if !ok {
		panic("invalid integer string")
	}
	return val
}

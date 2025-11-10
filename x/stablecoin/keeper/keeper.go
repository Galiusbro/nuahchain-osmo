package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	osmomath "github.com/osmosis-labs/osmosis/osmomath"
	feetypes "github.com/osmosis-labs/osmosis/v30/x/fees/types"
	"github.com/osmosis-labs/osmosis/v30/x/stablecoin/types"
)

// Keeper provides access to stablecoin statistics.
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	bankKeeper types.BankKeeper
	paramstore paramtypes.Subspace
}

// NewKeeper creates a new stablecoin keeper instance.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, bankKeeper types.BankKeeper, ps paramtypes.Subspace) Keeper {
	if bankKeeper == nil {
		panic("bank keeper cannot be nil")
	}

	if ps.Name() != "" && !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		bankKeeper: bankKeeper,
		paramstore: ps,
	}
}

// GetNDollarDenom returns the actual NDOLLAR denom by searching through bank metadata.
// It looks for a denom with "NDOLLAR" as display name or alias.
// Returns the base denom (factory/.../ndollar format) if found, otherwise returns types.NDollarDenom as fallback.
func (k Keeper) GetNDollarDenom(ctx sdk.Context) string {
	// Try to find NDOLLAR metadata by iterating through all denom metadata
	metadata := k.bankKeeper.GetAllDenomMetaData(ctx)
	for _, meta := range metadata {
		// Check if this is NDOLLAR by looking at display name or symbol
		if meta.Display == "ndollar" || meta.Symbol == "NDOLLAR" || meta.Display == "NDOLLAR" {
			return meta.Base
		}
		// Also check aliases
		for _, unit := range meta.DenomUnits {
			for _, alias := range unit.Aliases {
				if alias == "NDOLLAR" || alias == "ndollar" {
					return meta.Base
				}
			}
		}
	}
	// Fallback to constant if not found (for backward compatibility)
	return types.NDollarDenom
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
	if state == nil {
		stats := types.NewStats(sdkmath.ZeroInt(), sdkmath.ZeroInt())
		k.setTotalMinted(ctx, parseInt(stats.TotalMinted))
		k.SetParams(ctx, types.DefaultParams())
		return
	}

	stats := state.GetStats()
	if stats == nil {
		stats = &types.Stats{}
	}

	minted := parseInt(stats.TotalMinted)
	burned := parseInt(stats.TotalBurned)

	k.setTotalMinted(ctx, minted)
	k.setTotalBurned(ctx, burned)

	if state.Params != nil {
		k.SetParams(ctx, *state.Params)
	} else {
		k.SetParams(ctx, types.DefaultParams())
	}
}

// ExportGenesis exports current module state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	stats := k.GetStats(ctx)
	params := k.GetParams(ctx)
	return &types.GenesisState{Stats: &stats, Params: &params}
}

func (k Keeper) getReserveBalance(ctx sdk.Context) sdkmath.Int {
	addr := authtypes.NewModuleAddress(feetypes.ModuleName)
	ndollarDenom := k.GetNDollarDenom(ctx)
	coin := k.bankKeeper.GetBalance(ctx, addr, ndollarDenom)
	return coin.Amount
}

func (k Keeper) coverageMetrics(ctx sdk.Context) (sdkmath.Int, sdkmath.Int, string) {
	stats := k.GetStats(ctx)
	outstanding := parseInt(stats.Outstanding)
	reserve := k.getReserveBalance(ctx)
	ratio := osmomath.ZeroDec()

	if outstanding.IsPositive() {
		reserveDec := osmomath.NewDecFromInt(reserve)
		outstandingDec := osmomath.NewDecFromInt(outstanding)
		if !outstandingDec.IsZero() {
			ratio = reserveDec.Quo(outstandingDec)
		}
	}

	return outstanding, reserve, ratio.String()
}

// GetParams returns the module parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	if k.paramstore.Name() == "" {
		return types.DefaultParams()
	}

	var params types.Params
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams stores the module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if err := params.Validate(); err != nil {
		panic(err)
	}

	if k.paramstore.Name() == "" {
		return
	}

	k.paramstore.SetParamSet(ctx, &params)
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

// BuyNDollar converts unuah to NDOLLAR at 1:1 ratio
// Returns the NDOLLAR denom and the amount minted
func (k Keeper) BuyNDollar(ctx sdk.Context, buyer sdk.AccAddress, unuahAmount sdkmath.Int) (string, sdkmath.Int, error) {
	if unuahAmount.IsNil() || !unuahAmount.IsPositive() {
		return "", sdkmath.Int{}, fmt.Errorf("unuah amount must be positive")
	}

	// Get the real NDOLLAR denom
	ndollarDenom := k.GetNDollarDenom(ctx)
	if ndollarDenom == "" {
		return "", sdkmath.Int{}, fmt.Errorf("NDOLLAR denom not found")
	}

	// Create coins
	unuahCoin := sdk.NewCoin("unuah", unuahAmount)
	ndollarCoin := sdk.NewCoin(ndollarDenom, unuahAmount) // 1:1 conversion

	// Transfer unuah from buyer to module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, buyer, types.ModuleName, sdk.NewCoins(unuahCoin)); err != nil {
		return "", sdkmath.Int{}, fmt.Errorf("failed to send unuah: %w", err)
	}

	// Mint NDOLLAR to module
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(ndollarCoin)); err != nil {
		return "", sdkmath.Int{}, fmt.Errorf("failed to mint NDOLLAR: %w", err)
	}

	// Send NDOLLAR from module to buyer
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, buyer, sdk.NewCoins(ndollarCoin)); err != nil {
		return "", sdkmath.Int{}, fmt.Errorf("failed to send NDOLLAR: %w", err)
	}

	// Record the mint in statistics
	if err := k.RecordMint(ctx, unuahAmount); err != nil {
		return "", sdkmath.Int{}, fmt.Errorf("failed to record mint: %w", err)
	}

	return ndollarDenom, unuahAmount, nil
}

// SellNDollar converts NDOLLAR back to unuah at 1:1 ratio
// Returns the unuah amount received
func (k Keeper) SellNDollar(ctx sdk.Context, seller sdk.AccAddress, ndollarAmount sdkmath.Int) (sdkmath.Int, error) {
	if ndollarAmount.IsNil() || !ndollarAmount.IsPositive() {
		return sdkmath.Int{}, fmt.Errorf("ndollar amount must be positive")
	}

	// Get the real NDOLLAR denom
	ndollarDenom := k.GetNDollarDenom(ctx)
	if ndollarDenom == "" {
		return sdkmath.Int{}, fmt.Errorf("NDOLLAR denom not found")
	}

	// Create coins
	ndollarCoin := sdk.NewCoin(ndollarDenom, ndollarAmount)
	unuahCoin := sdk.NewCoin("unuah", ndollarAmount) // 1:1 conversion

	// Transfer NDOLLAR from seller to module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, seller, types.ModuleName, sdk.NewCoins(ndollarCoin)); err != nil {
		return sdkmath.Int{}, fmt.Errorf("failed to send NDOLLAR: %w", err)
	}

	// Burn NDOLLAR from module
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(ndollarCoin)); err != nil {
		return sdkmath.Int{}, fmt.Errorf("failed to burn NDOLLAR: %w", err)
	}

	// Send unuah from module to seller
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seller, sdk.NewCoins(unuahCoin)); err != nil {
		return sdkmath.Int{}, fmt.Errorf("failed to send unuah: %w", err)
	}

	// Record the burn in statistics
	if err := k.RecordBurn(ctx, ndollarAmount); err != nil {
		return sdkmath.Int{}, fmt.Errorf("failed to record burn: %w", err)
	}

	return ndollarAmount, nil
}

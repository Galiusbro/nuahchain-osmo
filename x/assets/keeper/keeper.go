package keeper

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	osmomath "github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
	feetypes "github.com/osmosis-labs/osmosis/v30/x/fees/types"
	oracletypes "github.com/osmosis-labs/osmosis/v30/x/oracle/types"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
)

// Keeper provides methods for managing assets state.
type Keeper struct {
	cdc          codec.BinaryCodec
	storeKey     storetypes.StoreKey
	bankKeeper   types.BankKeeper
	oracleKeeper types.OracleKeeper
	feesKeeper   types.FeesKeeper
	stableKeeper types.StablecoinKeeper
}

// NewKeeper returns a new Keeper instance.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, bankKeeper types.BankKeeper, oracleKeeper types.OracleKeeper, feesKeeper types.FeesKeeper, stableKeeper types.StablecoinKeeper) Keeper {
	return Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		bankKeeper:   bankKeeper,
		oracleKeeper: oracleKeeper,
		feesKeeper:   feesKeeper,
		stableKeeper: stableKeeper,
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

// BuyAsset executes a purchase of the given asset using NDOLLAR.
func (k Keeper) BuyAsset(ctx sdk.Context, buyer sdk.AccAddress, symbol string, amountND sdkmath.Int) (sdk.Coin, osmomath.Dec, error) {
	if amountND.IsNil() || !amountND.IsPositive() {
		return sdk.Coin{}, osmomath.Dec{}, fmt.Errorf("ndollar amount must be positive")
	}

	_, _, err := k.EnsureAsset(ctx, symbol)
	if err != nil {
		return sdk.Coin{}, osmomath.Dec{}, err
	}

	price, err := k.getPrice(ctx, symbol)
	if err != nil {
		return sdk.Coin{}, osmomath.Dec{}, err
	}

	priceDec, err := osmomath.NewDecFromStr(price.Value)
	if err != nil {
		return sdk.Coin{}, osmomath.Dec{}, fmt.Errorf("invalid price value: %w", err)
	}
	if priceDec.IsZero() {
		return sdk.Coin{}, osmomath.Dec{}, fmt.Errorf("price cannot be zero")
	}

	feeRate := k.feesKeeper.GetTradeFeeRate(ctx)
	amountNDDec := osmomath.NewDecFromInt(amountND)
	feeInt := feeRate.Mul(amountNDDec).TruncateInt()
	if feeInt.GTE(amountND) {
		feeInt = amountND.Sub(sdkmath.OneInt())
	}

	netNDInt := amountND.Sub(feeInt)
	if !netNDInt.IsPositive() {
		return sdk.Coin{}, osmomath.Dec{}, fmt.Errorf("purchase amount too small after fees")
	}

	netNDDec := osmomath.NewDecFromInt(netNDInt)
	baseAmountDec := netNDDec.Quo(priceDec)
	if baseAmountDec.IsZero() {
		return sdk.Coin{}, osmomath.Dec{}, fmt.Errorf("purchase amount too small")
	}

	microFactor := osmomath.NewDecFromInt(types.AssetPrecisionFactor())
	assetCoinAmount := baseAmountDec.Mul(microFactor).TruncateInt()
	if !assetCoinAmount.IsPositive() {
		return sdk.Coin{}, osmomath.Dec{}, fmt.Errorf("purchase amount too small")
	}

	assetDenom := types.AssetDenom(symbol)
	assetCoin := sdk.NewCoin(assetDenom, assetCoinAmount)
	fullPayment := sdk.NewCoin(types.NDollarDenom, amountND)

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, buyer, types.ModuleName, sdk.NewCoins(fullPayment)); err != nil {
		return sdk.Coin{}, osmomath.Dec{}, err
	}

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(assetCoin)); err != nil {
		return sdk.Coin{}, osmomath.Dec{}, err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, buyer, sdk.NewCoins(assetCoin)); err != nil {
		return sdk.Coin{}, osmomath.Dec{}, err
	}

	if feeInt.IsPositive() {
		feeCoin := sdk.NewCoin(types.NDollarDenom, feeInt)
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, feetypes.ModuleName, sdk.NewCoins(feeCoin)); err != nil {
			return sdk.Coin{}, osmomath.Dec{}, err
		}
	}

	if netNDInt.IsPositive() {
		netCoin := sdk.NewCoin(types.NDollarDenom, netNDInt)
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(netCoin)); err != nil {
			return sdk.Coin{}, osmomath.Dec{}, err
		}
		if err := k.stableKeeper.RecordBurn(ctx, netNDInt); err != nil {
			return sdk.Coin{}, osmomath.Dec{}, err
		}
	}

	return assetCoin, baseAmountDec, nil
}

func (k Keeper) getPrice(ctx sdk.Context, symbol string) (*oracletypes.Price, error) {
	price, found := k.oracleKeeper.GetPrice(ctx, strings.TrimSpace(symbol))
	if !found {
		return nil, fmt.Errorf("price for %s not found", symbol)
	}
	return price, nil
}

// SellAsset executes the sale of the given asset amount for NDOLLAR.
func (k Keeper) SellAsset(ctx sdk.Context, seller sdk.AccAddress, symbol string, baseAmount osmomath.Dec) (sdk.Coin, sdkmath.Int, error) {
	if !baseAmount.IsPositive() {
		return sdk.Coin{}, sdkmath.Int{}, fmt.Errorf("base amount must be positive")
	}

	microFactor := osmomath.NewDecFromInt(types.AssetPrecisionFactor())
	assetAmount := baseAmount.Mul(microFactor).TruncateInt()
	if !assetAmount.IsPositive() {
		return sdk.Coin{}, sdkmath.Int{}, fmt.Errorf("base amount too small")
	}

	assetDenom := types.AssetDenom(symbol)
	assetCoin := sdk.NewCoin(assetDenom, assetAmount)

	balance := k.bankKeeper.GetBalance(ctx, seller, assetDenom)
	if balance.Amount.LT(assetAmount) {
		return sdk.Coin{}, sdkmath.Int{}, fmt.Errorf("insufficient asset balance")
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, seller, types.ModuleName, sdk.NewCoins(assetCoin)); err != nil {
		return sdk.Coin{}, sdkmath.Int{}, err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(assetCoin)); err != nil {
		return sdk.Coin{}, sdkmath.Int{}, err
	}

	price, err := k.getPrice(ctx, symbol)
	if err != nil {
		return sdk.Coin{}, sdkmath.Int{}, err
	}

	priceDec, err := osmomath.NewDecFromStr(price.Value)
	if err != nil {
		return sdk.Coin{}, sdkmath.Int{}, fmt.Errorf("invalid price value: %w", err)
	}

	payoutDec := baseAmount.Mul(priceDec)
	payoutInt := payoutDec.TruncateInt()
	if !payoutInt.IsPositive() {
		return sdk.Coin{}, sdkmath.Int{}, fmt.Errorf("payout too small")
	}

	feeRate := k.feesKeeper.GetTradeFeeRate(ctx)
	feeInt := payoutDec.Mul(feeRate).TruncateInt()
	if feeInt.GTE(payoutInt) {
		feeInt = payoutInt.Sub(sdkmath.OneInt())
	}

	netInt := payoutInt.Sub(feeInt)
	if netInt.IsNegative() {
		return sdk.Coin{}, sdkmath.Int{}, fmt.Errorf("payout too small after fees")
	}

	totalCoin := sdk.NewCoin(types.NDollarDenom, payoutInt)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(totalCoin)); err != nil {
		return sdk.Coin{}, sdkmath.Int{}, err
	}

	if feeInt.IsPositive() {
		feeCoin := sdk.NewCoin(types.NDollarDenom, feeInt)
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, feetypes.ModuleName, sdk.NewCoins(feeCoin)); err != nil {
			return sdk.Coin{}, sdkmath.Int{}, err
		}
	}

	netCoin := sdk.NewCoin(types.NDollarDenom, netInt)
	if netInt.IsPositive() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seller, sdk.NewCoins(netCoin)); err != nil {
			return sdk.Coin{}, sdkmath.Int{}, err
		}
	}

	if err := k.stableKeeper.RecordMint(ctx, payoutInt); err != nil {
		return sdk.Coin{}, sdkmath.Int{}, err
	}

	return netCoin, netInt, nil
}

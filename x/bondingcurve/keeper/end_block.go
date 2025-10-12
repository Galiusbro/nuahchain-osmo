package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"

	appparams "github.com/osmosis-labs/osmosis/v30/app/params"
	"github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
	balancertypes "github.com/osmosis-labs/osmosis/v30/x/gamm/pool-models/balancer"
)

// EndBlocker processes bonding curve pools and activates DEX liquidity when thresholds are met.
func (k Keeper) EndBlocker(ctx sdk.Context) error {
	if err := k.applyPendingParams(ctx); err != nil {
		return err
	}

	if err := k.ProcessLiquidations(ctx); err != nil {
		return err
	}

	params := k.GetParams(ctx)

	iterator := storetypes.KVStorePrefixIterator(k.getStore(ctx), types.TokenPoolKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var pool types.BondingCurvePool
		k.cdc.MustUnmarshal(iterator.Value(), &pool)

		if pool.DexActivated {
			continue
		}

		tokensSold := pool.TokensSoldDec()
		maxSupply := params.MaxSupplyDec()
		// Activate DEX when 85% of max supply is sold
		activationThreshold := maxSupply.Mul(osmomath.MustNewDecFromStr("0.85"))
		if tokensSold.LT(activationThreshold) {
			continue
		}

		if err := k.activateDex(ctx, params, pool); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) activateDex(ctx sdk.Context, params types.Params, pool types.BondingCurvePool) error {
	token, found := k.userTokenKeeper.GetToken(ctx, pool.Denom)
	if !found {
		return types.ErrInvalidToken
	}

	providerStr := params.BondingCurveWallet
	if providerStr == "" {
		providerStr = token.Creator
	}

	if providerStr == "" {
		return fmt.Errorf("no valid provider address found: AiCeoWallet='%s', PlatformWallet='%s', Creator='%s'",
			token.Distribution.AiCeoWallet, token.Distribution.PlatformWallet, token.Creator)
	}

	providerAddr, err := sdk.AccAddressFromBech32(providerStr)
	if err != nil {
		return fmt.Errorf("invalid provider address '%s': %w", providerStr, err)
	}

	liquidityAmount := token.Distribution.BondingCurveSupplyInt()
	if liquidityAmount.IsZero() {
		liquidityAmount = params.MaxSupplyDec().TruncateInt()
	}

	if liquidityAmount.IsZero() {
		return types.ErrInvalidAmount
	}

	providerBalance := k.bankKeeper.GetBalance(sdk.WrapSDKContext(ctx), providerAddr, pool.Denom)
	if providerBalance.Amount.LT(liquidityAmount) {
		return types.ErrInsufficientLiquidity
	}

	quoteDenom := params.QuoteDenom
	if quoteDenom == "" {
		quoteDenom = appparams.BaseCoinUnit
	}

	providerQuote := k.bankKeeper.GetBalance(sdk.WrapSDKContext(ctx), providerAddr, quoteDenom)
	if providerQuote.Amount.LT(liquidityAmount) {
		return types.ErrInsufficientLiquidity
	}

	poolAssets := []balancertypes.PoolAsset{
		{
			Token:  sdk.NewCoin(pool.Denom, liquidityAmount),
			Weight: osmomath.OneInt(),
		},
		{
			Token:  sdk.NewCoin(quoteDenom, liquidityAmount),
			Weight: osmomath.OneInt(),
		},
	}

	poolParams := balancertypes.PoolParams{
		SwapFee: osmomath.ZeroDec(),
		ExitFee: osmomath.ZeroDec(),
	}

	msg := balancertypes.NewMsgCreateBalancerPool(providerAddr, poolParams, poolAssets, "")

	poolID, err := k.poolManager.CreatePool(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to create balancer pool: %w", err)
	}

	pool.MarkDexActivated(poolID, providerStr)
	k.setPool(ctx, pool)

	token.State.CurveCompleted = true
	token.State.DexTradingEnabled = true
	token.State.SoftLockEnabled = false
	if err := k.userTokenKeeper.UpdateToken(ctx, token); err != nil {
		return err
	}

	return nil
}

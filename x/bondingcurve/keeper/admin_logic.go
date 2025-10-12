package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
)

func (k Keeper) ensureModuleActive(ctx sdk.Context, denom string, participant sdk.AccAddress) error {
	now := ctx.BlockTime()

	if global, ok := k.getGlobalPause(ctx); ok && global.IsActive(now) {
		return types.ErrModulePaused
	}

	if denom != "" {
		if tokenPause, ok := k.getTokenPause(ctx, denom); ok && tokenPause.IsActive(now) {
			return types.ErrTokenPaused
		}
		if freezeInfo, ok := k.getFreezeInfo(ctx, types.FreezeTargetType_FREEZE_TARGET_TYPE_TOKEN, denom); ok && freezeInfo.IsFrozen(now) {
			return types.ErrTokenPaused
		}
	}

	if participant != nil {
		if freezeInfo, ok := k.getFreezeInfo(ctx, types.FreezeTargetType_FREEZE_TARGET_TYPE_ADDRESS, participant.String()); ok && freezeInfo.IsFrozen(now) {
			return types.ErrAddressFrozen
		}
	}

	return nil
}

func (k Keeper) forceLiquidation(ctx sdk.Context, positionID uint64) (bool, error) {
	position, found := k.GetMarginPosition(ctx, positionID)
	if !found || position.Status != types.PositionStatus_POSITION_STATUS_OPEN {
		return false, nil
	}

	marginPool := k.ensureMarginPool(ctx, position.Denom)
	updatedPool, priceInfo, err := k.updatePriceInfo(ctx, marginPool)
	if err != nil {
		return false, err
	}

	markPrice := priceInfo.MarkPrice
	if !markPrice.IsPositive() {
		params := k.GetParams(ctx)
		pool := k.ensurePool(ctx, position.Denom)
		markPrice = pool.LastPriceDec()
		if !markPrice.IsPositive() {
			markPrice = types.CalculatePrice(pool.TokensSoldDec(), params)
		}
		if !markPrice.IsPositive() {
			markPrice = params.StartPriceDec()
		}
	}

	poolAfter, _, err := k.executeFullLiquidation(ctx, updatedPool, position, markPrice, nil)
	if err != nil {
		return false, err
	}
	k.setMarginPool(ctx, poolAfter)

	return true, nil
}

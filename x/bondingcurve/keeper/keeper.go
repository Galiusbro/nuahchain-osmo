package keeper

import (
	"fmt"
	"sort"
	"strings"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/osmomath"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
)

type Keeper struct {
	cdc             codec.BinaryCodec
	storeKey        storetypes.StoreKey
	paramstore      paramtypes.Subspace
	accountKeeper   types.AccountKeeper
	bankKeeper      types.BankKeeper
	userTokenKeeper types.UserTokenKeeper
	poolManager     types.PoolManagerKeeper
}

const (
	maxLiquidationsPerBlock   = 50
	maxPositionsConsidered    = 250
	liquidationUrgencyBuffer  = 0.05
	liquidationAlertThreshold = 25
)

var (
	priceVolatilityAlertThreshold = osmomath.NewDecWithPrec(3, 1) // 30% deviation alert
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	userTokenKeeper types.UserTokenKeeper,
	poolManager types.PoolManagerKeeper,
) Keeper {
	if ps.Name() != "" && !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		paramstore:      ps,
		accountKeeper:   accountKeeper,
		bankKeeper:      bankKeeper,
		userTokenKeeper: userTokenKeeper,
		poolManager:     poolManager,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	if k.paramstore.Name() == "" {
		return types.DefaultParams()
	}

	var params types.Params
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if err := params.Validate(); err != nil {
		panic(err)
	}
	if k.paramstore.Name() == "" {
		return
	}
	k.paramstore.SetParamSet(ctx, &params)
}

func (k Keeper) getStore(ctx sdk.Context) storetypes.KVStore {
	return ctx.KVStore(k.storeKey)
}

func (k Keeper) GetPool(ctx sdk.Context, denom string) (types.BondingCurvePool, bool) {
	store := k.getStore(ctx)
	bz := store.Get(types.TokenPoolKey(denom))
	if bz == nil {
		return types.BondingCurvePool{}, false
	}
	var pool types.BondingCurvePool
	k.cdc.MustUnmarshal(bz, &pool)
	return pool, true
}

func (k Keeper) setPool(ctx sdk.Context, pool types.BondingCurvePool) {
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(&pool)
	store.Set(types.TokenPoolKey(pool.Denom), bz)
}

func (k Keeper) SetPool(ctx sdk.Context, pool types.BondingCurvePool) {
	k.setPool(ctx, pool)
}

func (k Keeper) GetMarginPool(ctx sdk.Context, denom string) (types.MarginPool, bool) {
	store := k.getStore(ctx)
	bz := store.Get(types.MarginPoolKey(denom))
	if bz == nil {
		return types.MarginPool{}, false
	}
	var pool types.MarginPool
	k.cdc.MustUnmarshal(bz, &pool)
	return pool, true
}

func (k Keeper) setMarginPool(ctx sdk.Context, pool types.MarginPool) {
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(&pool)
	store.Set(types.MarginPoolKey(pool.Denom), bz)
}

func (k Keeper) SetMarginPool(ctx sdk.Context, pool types.MarginPool) {
	k.setMarginPool(ctx, pool)
}

func (k Keeper) ensurePool(ctx sdk.Context, denom string) types.BondingCurvePool {
	pool, exists := k.GetPool(ctx, denom)
	if exists {
		return pool
	}
	pool = types.BondingCurvePool{
		Denom:             denom,
		TokensSold:        "0",
		ReserveNuah:       "0",
		ReserveNdollar:    "0",
		LastPrice:         "0",
		DexPoolId:         0,
		DexActivated:      false,
		LiquidityProvider: "",
	}
	k.setPool(ctx, pool)
	return pool
}

func (k Keeper) ensureMarginPool(ctx sdk.Context, denom string) types.MarginPool {
	marginPool, exists := k.GetMarginPool(ctx, denom)
	if exists {
		// ensure defaults populated
		if marginPool.MaintenanceMarginRatio == "" {
			marginPool.SetMaintenanceMarginRatio(types.DefaultMaintenanceMarginRatio)
		}
		if marginPool.NextPositionId == 0 {
			marginPool.NextPositionId = 1
		}
		if marginPool.InsuranceFund == "" {
			marginPool.SetInsuranceFund(osmomath.ZeroDec())
		}
		if marginPool.TotalLiquidationFees == "" {
			marginPool.SetTotalLiquidationFees(osmomath.ZeroDec())
		}
		if marginPool.TotalBadDebt == "" {
			marginPool.SetTotalBadDebt(osmomath.ZeroDec())
		}
		if marginPool.LastMarkPrice == "" {
			marginPool.SetLastMarkPrice(osmomath.ZeroDec())
		}
		if marginPool.LastTwapPrice == "" {
			marginPool.SetLastTwapPrice(osmomath.ZeroDec())
		}
		k.setMarginPool(ctx, marginPool)
		return marginPool
	}

	available := osmomath.ZeroDec()
	pool, poolExists := k.GetPool(ctx, denom)
	params := k.GetParams(ctx)
	if poolExists {
		available = available.Add(pool.ReserveNdollarDec())
		price := pool.LastPriceDec()
		if price.IsZero() {
			price = types.CalculatePrice(pool.TokensSoldDec(), params)
		}
		if price.IsPositive() {
			available = available.Add(pool.ReserveNuahDec().Mul(price))
		}
	}
	markPrice := osmomath.ZeroDec()
	if poolExists {
		markPrice = pool.LastPriceDec()
		if markPrice.IsZero() {
			markPrice = types.CalculatePrice(pool.TokensSoldDec(), params)
		}
	}

	marginPool = types.MarginPool{
		Denom:                  denom,
		TotalCollateral:        osmomath.ZeroDec().String(),
		AvailableLiquidity:     available.String(),
		TotalLongExposure:      osmomath.ZeroDec().String(),
		TotalShortExposure:     osmomath.ZeroDec().String(),
		MaintenanceMarginRatio: types.DefaultMaintenanceMarginRatio.String(),
		NextPositionId:         1,
		LastFundingTimestamp:   uint64(ctx.BlockTime().Unix()),
		CumulativeFundingRate:  types.DefaultFundingRate.String(),
		InsuranceFund:          osmomath.ZeroDec().String(),
		TotalLiquidationFees:   osmomath.ZeroDec().String(),
		TotalBadDebt:           osmomath.ZeroDec().String(),
		LiquidationsPaused:     false,
		LastMarkPrice:          markPrice.String(),
		LastTwapPrice:          markPrice.String(),
		TotalLiquidations:      0,
		LastLiquidationHeight:  0,
	}

	k.setMarginPool(ctx, marginPool)
	return marginPool
}

func (k Keeper) GetMarginPosition(ctx sdk.Context, id uint64) (types.MarginPosition, bool) {
	store := k.getStore(ctx)
	bz := store.Get(types.MarginPositionKey(id))
	if bz == nil {
		return types.MarginPosition{}, false
	}
	var position types.MarginPosition
	k.cdc.MustUnmarshal(bz, &position)
	return position, true
}

func (k Keeper) setMarginPosition(ctx sdk.Context, position types.MarginPosition) {
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(&position)
	store.Set(types.MarginPositionKey(position.Id), bz)
	if position.Trader != "" {
		if trader, err := sdk.AccAddressFromBech32(position.Trader); err == nil {
			store.Set(types.MarginPositionTraderIndexKey(trader, position.Id), []byte{})
		}
	}
}

func (k Keeper) deleteMarginPosition(ctx sdk.Context, id uint64) {
	store := k.getStore(ctx)
	if position, found := k.GetMarginPosition(ctx, id); found {
		if position.Trader != "" {
			if trader, err := sdk.AccAddressFromBech32(position.Trader); err == nil {
				store.Delete(types.MarginPositionTraderIndexKey(trader, id))
			}
		}
	}
	store.Delete(types.MarginPositionKey(id))
}

type liquidationCandidate struct {
	Position  types.MarginPosition
	MarkPrice osmomath.Dec
	TwapPrice osmomath.Dec
	Distance  osmomath.Dec
}

type liquidationResult struct {
	liquidationType  string
	payout           osmomath.Dec
	liquidatorReward osmomath.Dec
	insurancePayout  osmomath.Dec
	badDebt          osmomath.Dec
}

func (k Keeper) ProcessLiquidations(ctx sdk.Context) error {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.MarginPositionKeyPrefix)
	defer iterator.Close()

	candidates := make([]liquidationCandidate, 0, 16)
	considered := 0
	volatilityAlerts := make(map[string]bool)

	for ; iterator.Valid(); iterator.Next() {
		if considered >= maxPositionsConsidered {
			break
		}
		var position types.MarginPosition
		k.cdc.MustUnmarshal(iterator.Value(), &position)
		if position.Status != types.PositionStatus_POSITION_STATUS_OPEN {
			continue
		}

		marginPool := k.ensureMarginPool(ctx, position.Denom)
		updatedPool, priceInfo, err := k.updatePriceInfo(ctx, marginPool)
		if err != nil {
			continue
		}

		marginPool = updatedPool

		if priceInfo.TwapPrice.IsPositive() && !priceInfo.IsCircuitBreakerTriggered() && !volatilityAlerts[position.Denom] {
			deviation := priceInfo.MarkPrice.Sub(priceInfo.TwapPrice).Abs().Quo(priceInfo.TwapPrice)
			if deviation.GTE(priceVolatilityAlertThreshold) {
				ctx.EventManager().EmitEvent(sdk.NewEvent(
					types.EventTypePriceVolatility,
					sdk.NewAttribute(types.AttributeKeyDenom, position.Denom),
					sdk.NewAttribute(types.AttributeKeyPrice, priceInfo.MarkPrice.String()),
					sdk.NewAttribute(types.AttributeKeyVolatility, deviation.String()),
				))
				volatilityAlerts[position.Denom] = true
			}
		}

		if marginPool.LiquidationsPaused {
			if priceInfo.IsCircuitBreakerTriggered() {
				continue
			}
			marginPool.LiquidationsPaused = false
			k.setMarginPool(ctx, marginPool)
		}

		if priceInfo.IsCircuitBreakerTriggered() {
			marginPool.LiquidationsPaused = true
			marginPool.LastLiquidationHeight = uint64(ctx.BlockHeight())
			k.setMarginPool(ctx, marginPool)
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeCircuitBreaker,
				sdk.NewAttribute(types.AttributeKeyDenom, position.Denom),
				sdk.NewAttribute(types.AttributeKeyPrice, priceInfo.MarkPrice.String()),
				sdk.NewAttribute(types.AttributeKeyPaused, "true"),
			))
			continue
		}

		ratio := k.liquidationRatio(position, priceInfo.MarkPrice)
		if ratio.IsNegative() {
			continue
		}

		candidates = append(candidates, liquidationCandidate{
			Position:  position,
			MarkPrice: priceInfo.MarkPrice,
			TwapPrice: priceInfo.TwapPrice,
			Distance:  ratio,
		})
		considered++
	}

	if len(candidates) == 0 {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Distance.GT(candidates[j].Distance)
	})

	if len(candidates) > maxPositionsConsidered {
		candidates = candidates[:maxPositionsConsidered]
	}

	processed := 0
	denomLiquidations := make(map[string]int)
	for _, candidate := range candidates {
		if processed >= maxLiquidationsPerBlock {
			break
		}

		marginPool := k.ensureMarginPool(ctx, candidate.Position.Denom)
		updatedPool, ok, err := k.handleLiquidationCandidate(ctx, marginPool, candidate)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		k.setMarginPool(ctx, updatedPool)
		denomLiquidations[candidate.Position.Denom]++
		processed++
	}

	if processed > 0 {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeLiquidationBatch,
			sdk.NewAttribute(types.AttributeKeyPositionsProcessed, fmt.Sprintf("%d", processed)),
		))
		if processed >= liquidationAlertThreshold {
			details := make([]string, 0, len(denomLiquidations))
			for denom, count := range denomLiquidations {
				details = append(details, fmt.Sprintf("%s:%d", denom, count))
			}
			sort.Strings(details)
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeLiquidationAlert,
				sdk.NewAttribute(types.AttributeKeyPositionsProcessed, fmt.Sprintf("%d", processed)),
				sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", liquidationAlertThreshold)),
				sdk.NewAttribute(types.AttributeKeyDenom, strings.Join(details, ",")),
			))
		}
	}

	return nil
}

func (k Keeper) updatePriceInfo(ctx sdk.Context, pool types.MarginPool) (types.MarginPool, types.PriceInfo, error) {
	bondingPool := k.ensurePool(ctx, pool.Denom)
	params := k.GetParams(ctx)
	markPrice := bondingPool.LastPriceDec()
	if !markPrice.IsPositive() {
		markPrice = types.CalculatePrice(bondingPool.TokensSoldDec(), params)
	}
	if !markPrice.IsPositive() {
		return pool, types.PriceInfo{}, types.ErrPriceUnavailable
	}

	prevTwap := pool.LastTwapPriceDec()
	if prevTwap.IsZero() {
		prevTwap = markPrice
	}
	alpha := types.TwapSmoothingFactor
	if alpha.IsZero() {
		alpha = osmomath.NewDecWithPrec(2, 1)
	}
	if alpha.GT(osmomath.OneDec()) {
		alpha = osmomath.OneDec()
	}
	twap := prevTwap.Mul(osmomath.OneDec().Sub(alpha)).Add(markPrice.Mul(alpha))

	pool.SetLastMarkPrice(markPrice)
	pool.SetLastTwapPrice(twap)
	k.setMarginPool(ctx, pool)

	return pool, types.PriceInfo{MarkPrice: markPrice, TwapPrice: twap}, nil
}

func (k Keeper) liquidationRatio(position types.MarginPosition, markPrice osmomath.Dec) osmomath.Dec {
	liqPrice := position.LiquidationPriceDec()
	if !markPrice.IsPositive() {
		return osmomath.OneDec()
	}
	if !liqPrice.IsPositive() {
		return osmomath.OneDec()
	}

	var delta osmomath.Dec
	if position.Type == types.PositionType_POSITION_TYPE_LONG {
		delta = liqPrice.Sub(markPrice)
	} else {
		delta = markPrice.Sub(liqPrice)
	}
	if delta.IsNegative() {
		return delta.Quo(liqPrice)
	}
	return delta.Quo(liqPrice)
}

func (k Keeper) handleLiquidationCandidate(ctx sdk.Context, pool types.MarginPool, candidate liquidationCandidate) (types.MarginPool, bool, error) {
	position, found := k.GetMarginPosition(ctx, candidate.Position.Id)
	if !found || position.Status != types.PositionStatus_POSITION_STATUS_OPEN {
		return pool, false, nil
	}

	ratio := k.liquidationRatio(position, candidate.MarkPrice)
	if ratio.IsNegative() {
		return pool, false, nil
	}

	shouldPartial := position.PositionSizeDec().GTE(types.MaxPartialPositionSize) && ratio.LTE(types.PartialLiquidationBuffer)
	if shouldPartial {
		updatedPool, _, err := k.executePartialLiquidation(ctx, pool, position, candidate.MarkPrice, nil)
		return updatedPool, true, err
	}

	updatedPool, _, err := k.executeFullLiquidation(ctx, pool, position, candidate.MarkPrice, nil)
	return updatedPool, true, err
}

func (k Keeper) executeFullLiquidation(ctx sdk.Context, pool types.MarginPool, position types.MarginPosition, markPrice osmomath.Dec, liquidator sdk.AccAddress) (types.MarginPool, liquidationResult, error) {
	collateral := position.CollateralAmountDec()
	positionSize := position.PositionSizeDec()
	pnl := position.CalculatePnL(markPrice)
	fee := positionSize.Mul(types.DefaultMarginFeeRate)
	penalty := collateral.Mul(types.LiquidationPenaltyRate)
	liquidatorReward := penalty.Mul(types.LiquidatorIncentiveRate)
	insuranceShare := penalty.Sub(liquidatorReward)

	payout := collateral.Add(pnl).Sub(fee).Sub(penalty)
	badDebt := osmomath.ZeroDec()
	if payout.IsNegative() {
		badDebt = payout.Neg()
		payout = osmomath.ZeroDec()
	}

	insuranceFund := pool.InsuranceFundDec().Add(insuranceShare)
	if badDebt.IsPositive() {
		cover := minDec(badDebt, insuranceFund)
		insuranceFund = insuranceFund.Sub(cover)
		badDebt = badDebt.Sub(cover)
	}
	pool.SetInsuranceFund(insuranceFund)
	if badDebt.IsPositive() {
		pool.SetTotalBadDebt(pool.TotalBadDebtDec().Add(badDebt))
	}

	pool.SetTotalLiquidationFees(pool.TotalLiquidationFeesDec().Add(penalty))
	pool.SetTotalCollateral(k.clampDec(pool.TotalCollateralDec().Sub(collateral)))
	pool.SetAvailableLiquidity(pool.AvailableLiquidityDec().Add(positionSize).Add(fee).Add(penalty))

	if position.Type == types.PositionType_POSITION_TYPE_LONG {
		pool.SetTotalLongExposure(k.clampDec(pool.TotalLongExposureDec().Sub(positionSize)))
	} else {
		pool.SetTotalShortExposure(k.clampDec(pool.TotalShortExposureDec().Sub(positionSize)))
	}

	pool.TotalLiquidations++
	pool.LastLiquidationHeight = uint64(ctx.BlockHeight())

	position.SetStatus(types.PositionStatus_POSITION_STATUS_LIQUIDATED)
	position.SetRealizedPnl(position.RealizedPnLDec().Add(pnl))
	position.SetPositionSize(osmomath.ZeroDec())
	position.SetCollateralAmount(osmomath.ZeroDec())
	position.SetMaintenanceMargin(osmomath.ZeroDec())
	position.SetLiquidationPrice(osmomath.ZeroDec())
	position.SetLastMarkPrice(markPrice)
	k.setMarginPosition(ctx, position)

	traderAddr, err := sdk.AccAddressFromBech32(position.Trader)
	if err != nil {
		return pool, liquidationResult{}, err
	}

	if payoutCoin, err := types.DecToCoin(payout, position.CollateralDenom); err == nil && payoutCoin.Amount.IsPositive() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(sdk.WrapSDKContext(ctx), types.ModuleName, traderAddr, sdk.NewCoins(payoutCoin)); err != nil {
			return pool, liquidationResult{}, err
		}
	}

	if len(liquidator) > 0 {
		if rewardCoin, err := types.DecToCoin(liquidatorReward, position.CollateralDenom); err == nil && rewardCoin.Amount.IsPositive() {
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(sdk.WrapSDKContext(ctx), types.ModuleName, liquidator, sdk.NewCoins(rewardCoin)); err != nil {
				return pool, liquidationResult{}, err
			}
		}
	}

	attributes := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeKeyTrader, position.Trader),
		sdk.NewAttribute(types.AttributeKeyDenom, position.Denom),
		sdk.NewAttribute(types.AttributeKeyPositionId, fmt.Sprintf("%d", position.Id)),
		sdk.NewAttribute(types.AttributeKeyLiquidationType, "full"),
		sdk.NewAttribute(types.AttributeKeyPayoutAmount, payout.String()),
		sdk.NewAttribute(types.AttributeKeyRealizedPnL, pnl.String()),
		sdk.NewAttribute(types.AttributeKeyFeesPaid, fee.String()),
		sdk.NewAttribute(types.AttributeKeyLiquidationPenalty, penalty.String()),
		sdk.NewAttribute(types.AttributeKeyLiquidatorReward, liquidatorReward.String()),
		sdk.NewAttribute(types.AttributeKeyInsurancePayout, insuranceShare.String()),
		sdk.NewAttribute(types.AttributeKeyBadDebtCovered, badDebt.String()),
	}
	if len(liquidator) > 0 {
		attributes = append(attributes, sdk.NewAttribute(types.AttributeKeyLiquidator, liquidator.String()))
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCloseMargin,
		attributes...,
	))

	return pool, liquidationResult{
		liquidationType:  "full",
		payout:           payout,
		liquidatorReward: liquidatorReward,
		insurancePayout:  insuranceShare,
		badDebt:          badDebt,
	}, nil
}

func (k Keeper) executePartialLiquidation(ctx sdk.Context, pool types.MarginPool, position types.MarginPosition, markPrice osmomath.Dec, liquidator sdk.AccAddress) (types.MarginPool, liquidationResult, error) {
	collateral := position.CollateralAmountDec()
	positionSize := position.PositionSizeDec()
	fraction := types.PartialLiquidationFraction
	if fraction.GT(osmomath.OneDec()) {
		fraction = osmomath.OneDec()
	}
	partialCollateral := collateral.Mul(fraction)
	partialPnL := position.CalculatePnL(markPrice).Mul(fraction)
	fee := positionSize.Mul(types.DefaultMarginFeeRate).Mul(fraction)
	penalty := partialCollateral.Mul(types.LiquidationPenaltyRate)
	liquidatorReward := penalty.Mul(types.LiquidatorIncentiveRate)
	insuranceShare := penalty.Sub(liquidatorReward)

	remainingCollateral := collateral.Add(partialPnL).Sub(penalty).Sub(fee)
	if remainingCollateral.IsNegative() {
		remainingCollateral = osmomath.ZeroDec()
	}
	newPositionSize := positionSize.Sub(positionSize.Mul(fraction))
	if newPositionSize.IsNegative() {
		newPositionSize = osmomath.ZeroDec()
	}
	reduction := positionSize.Sub(newPositionSize)

	payout := partialCollateral.Add(partialPnL).Sub(fee).Sub(penalty)
	badDebt := osmomath.ZeroDec()
	if payout.IsNegative() {
		badDebt = payout.Neg()
		payout = osmomath.ZeroDec()
	} else {
		remainingCollateral = remainingCollateral.Sub(payout)
		if remainingCollateral.IsNegative() {
			remainingCollateral = osmomath.ZeroDec()
		}
	}

	insuranceFund := pool.InsuranceFundDec().Add(insuranceShare)
	if badDebt.IsPositive() {
		cover := minDec(badDebt, insuranceFund)
		insuranceFund = insuranceFund.Sub(cover)
		badDebt = badDebt.Sub(cover)
	}
	pool.SetInsuranceFund(insuranceFund)
	if badDebt.IsPositive() {
		pool.SetTotalBadDebt(pool.TotalBadDebtDec().Add(badDebt))
	}

	pool.SetTotalLiquidationFees(pool.TotalLiquidationFeesDec().Add(penalty))
	pool.SetTotalCollateral(k.clampDec(pool.TotalCollateralDec().Sub(collateral.Sub(remainingCollateral))))
	pool.SetAvailableLiquidity(pool.AvailableLiquidityDec().Add(reduction).Add(fee).Add(penalty))

	if position.Type == types.PositionType_POSITION_TYPE_LONG {
		pool.SetTotalLongExposure(k.clampDec(pool.TotalLongExposureDec().Sub(reduction)))
	} else {
		pool.SetTotalShortExposure(k.clampDec(pool.TotalShortExposureDec().Sub(reduction)))
	}
	pool.TotalLiquidations++
	pool.LastLiquidationHeight = uint64(ctx.BlockHeight())

	maintMargin := types.CalculateMaintenanceMargin(newPositionSize)
	buffer := types.PartialLiquidationBuffer
	if buffer.IsNegative() {
		buffer = osmomath.ZeroDec()
	}
	if buffer.GTE(osmomath.OneDec()) {
		buffer = osmomath.NewDecWithPrec(5, 2) // fallback 5%
	}

	var newLiquidationPrice osmomath.Dec
	if position.Type == types.PositionType_POSITION_TYPE_LONG {
		newLiquidationPrice = markPrice.Mul(osmomath.OneDec().Sub(buffer))
		if newLiquidationPrice.IsNegative() {
			newLiquidationPrice = markPrice
		}
	} else {
		newLiquidationPrice = markPrice.Mul(osmomath.OneDec().Add(buffer))
	}

	position.SetCollateralAmount(remainingCollateral)
	position.SetPositionSize(newPositionSize)
	position.SetMaintenanceMargin(maintMargin)
	position.SetLiquidationPrice(newLiquidationPrice)
	position.SetRealizedPnl(position.RealizedPnLDec().Add(partialPnL))
	position.SetLastMarkPrice(markPrice)
	k.setMarginPosition(ctx, position)

	traderAddr, err := sdk.AccAddressFromBech32(position.Trader)
	if err != nil {
		return pool, liquidationResult{}, err
	}

	if payoutCoin, err := types.DecToCoin(payout, position.CollateralDenom); err == nil && payoutCoin.Amount.IsPositive() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(sdk.WrapSDKContext(ctx), types.ModuleName, traderAddr, sdk.NewCoins(payoutCoin)); err != nil {
			return pool, liquidationResult{}, err
		}
	}

	if len(liquidator) > 0 {
		if rewardCoin, err := types.DecToCoin(liquidatorReward, position.CollateralDenom); err == nil && rewardCoin.Amount.IsPositive() {
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(sdk.WrapSDKContext(ctx), types.ModuleName, liquidator, sdk.NewCoins(rewardCoin)); err != nil {
				return pool, liquidationResult{}, err
			}
		}
	}

	attributes := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeKeyTrader, position.Trader),
		sdk.NewAttribute(types.AttributeKeyDenom, position.Denom),
		sdk.NewAttribute(types.AttributeKeyPositionId, fmt.Sprintf("%d", position.Id)),
		sdk.NewAttribute(types.AttributeKeyLiquidationType, "partial"),
		sdk.NewAttribute(types.AttributeKeyPayoutAmount, payout.String()),
		sdk.NewAttribute(types.AttributeKeyRealizedPnL, partialPnL.String()),
		sdk.NewAttribute(types.AttributeKeyFeesPaid, fee.String()),
		sdk.NewAttribute(types.AttributeKeyLiquidationPenalty, penalty.String()),
		sdk.NewAttribute(types.AttributeKeyLiquidatorReward, liquidatorReward.String()),
		sdk.NewAttribute(types.AttributeKeyInsurancePayout, insuranceShare.String()),
		sdk.NewAttribute(types.AttributeKeyBadDebtCovered, badDebt.String()),
		sdk.NewAttribute(types.AttributeKeyPositionSize, newPositionSize.String()),
	}
	if len(liquidator) > 0 {
		attributes = append(attributes, sdk.NewAttribute(types.AttributeKeyLiquidator, liquidator.String()))
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCloseMargin,
		attributes...,
	))

	return pool, liquidationResult{
		liquidationType:  "partial",
		payout:           payout,
		liquidatorReward: liquidatorReward,
		insurancePayout:  insuranceShare,
		badDebt:          badDebt,
	}, nil
}

func (k Keeper) clampDec(dec osmomath.Dec) osmomath.Dec {
	if dec.IsNegative() {
		return osmomath.ZeroDec()
	}
	return dec
}

func minDec(a, b osmomath.Dec) osmomath.Dec {
	if a.LT(b) {
		return a
	}
	return b
}

func (k Keeper) GetModuleAddress() sdk.AccAddress {
	return k.accountKeeper.GetModuleAddress(types.ModuleName)
}

func (k Keeper) updateTokenState(ctx sdk.Context, denom string, sold osmomath.Dec, currentPrice osmomath.Dec) error {
	token, found := k.userTokenKeeper.GetToken(ctx, denom)
	if !found {
		return types.ErrInvalidToken
	}

	token.State.BondingCurveSold = sold.String()
	token.State.CurrentPrice = currentPrice.String()
	supply := osmomath.NewDecFromInt(token.Distribution.BondingCurveSupplyInt())
	if sold.GTE(supply) {
		token.State.CurveCompleted = true
	}

	return k.userTokenKeeper.UpdateToken(ctx, token)
}

func (k Keeper) bondingCurveWallet(ctx sdk.Context) (sdk.AccAddress, error) {
	params := k.GetParams(ctx)
	if params.BondingCurveWallet == "" {
		return k.GetModuleAddress(), nil
	}
	addr, err := sdk.AccAddressFromBech32(params.BondingCurveWallet)
	if err != nil {
		return sdk.AccAddress{}, err
	}
	return addr, nil
}

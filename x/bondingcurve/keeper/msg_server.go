package keeper

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/osmomath"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (s msgServer) BuyFromCurve(goCtx context.Context, msg *types.MsgBuyFromCurve) (*types.MsgBuyFromCurveResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	trader, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return nil, err
	}

	paymentDec, err := osmomath.NewDecFromStr(msg.PaymentAmount)
	if err != nil || !paymentDec.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	params := s.GetParams(ctx)

	if msg.PaymentDenom != params.QuoteDenom && msg.PaymentDenom != sdk.DefaultBondDenom {
		return nil, types.ErrInvalidPaymentDenom
	}

	pool := s.ensurePool(ctx, msg.Denom)

	tokensSold := pool.TokensSoldDec()
	maxSupply := params.MaxSupplyDec()

	tokensOutDec := types.IntegrateBuyAmount(tokensSold, paymentDec, params)
	if !tokensOutDec.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	if tokensSold.Add(tokensOutDec).GT(maxSupply) {
		return nil, types.ErrMaxSupplyReached
	}

	paymentCoin, err := types.DecToCoin(paymentDec, msg.PaymentDenom)
	if err != nil {
		return nil, err
	}

	if paymentCoin.Amount.IsZero() {
		return nil, types.ErrInvalidAmount
	}

	err = s.bankKeeper.SendCoinsFromAccountToModule(sdk.WrapSDKContext(ctx), trader, types.ModuleName, sdk.NewCoins(paymentCoin))
	if err != nil {
		return nil, err
	}

	tokensOutCoin, err := types.DecToCoin(tokensOutDec, msg.Denom)
	if err != nil {
		return nil, err
	}

	actualTokens := types.CoinToDec(tokensOutCoin)
	if msg.MinTokensOut != "" {
		minTokensDec, err := osmomath.NewDecFromStr(msg.MinTokensOut)
		if err != nil {
			return nil, types.ErrInvalidAmount
		}
		if actualTokens.LT(minTokensDec) {
			return nil, types.ErrMinTokensNotMet
		}
	}

	err = s.distributeTokens(ctx, trader, tokensOutCoin)
	if err != nil {
		return nil, err
	}

	// update pool
	pool.SetTokensSold(tokensSold.Add(actualTokens))

	actualPayment := types.CoinToDec(paymentCoin)
	if msg.PaymentDenom == params.QuoteDenom {
		pool.SetReserveNdollar(pool.ReserveNdollarDec().Add(actualPayment))
	} else {
		pool.SetReserveNuah(pool.ReserveNuahDec().Add(actualPayment))
	}

	lastPrice := types.CalculatePrice(pool.TokensSoldDec(), params)
	pool.SetLastPrice(lastPrice)

	s.setPool(ctx, pool)

	if err := s.updateTokenState(ctx, msg.Denom, pool.TokensSoldDec(), lastPrice); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeBuy,
		sdk.NewAttribute(types.AttributeKeyTrader, msg.Trader),
		sdk.NewAttribute(types.AttributeKeyDenom, msg.Denom),
		sdk.NewAttribute(types.AttributeKeyPaymentDenom, msg.PaymentDenom),
		sdk.NewAttribute(types.AttributeKeyPaymentAmount, actualPayment.String()),
		sdk.NewAttribute(types.AttributeKeyTokensOut, actualTokens.String()),
		sdk.NewAttribute(types.AttributeKeyPrice, lastPrice.String()),
	))

	return &types.MsgBuyFromCurveResponse{
		TokensOut: actualTokens.String(),
		PricePaid: actualPayment.String(),
	}, nil
}

func (s msgServer) SellToCurve(goCtx context.Context, msg *types.MsgSellToCurve) (*types.MsgSellToCurveResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	trader, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return nil, err
	}

	tokenAmountDec, err := osmomath.NewDecFromStr(msg.TokenAmount)
	if err != nil || !tokenAmountDec.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	params := s.GetParams(ctx)

	if msg.PaymentDenom != params.QuoteDenom && msg.PaymentDenom != sdk.DefaultBondDenom {
		return nil, types.ErrInvalidPaymentDenom
	}

	pool, found := s.GetPool(ctx, msg.Denom)
	if !found {
		return nil, types.ErrPoolNotFound
	}

	tokensSold := pool.TokensSoldDec()
	if tokenAmountDec.GT(tokensSold) {
		return nil, types.ErrInvalidAmount
	}

	paymentOutDec := types.IntegrateSellAmount(tokensSold, tokenAmountDec, params)
	if !paymentOutDec.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	paymentCoin, err := types.DecToCoin(paymentOutDec, msg.PaymentDenom)
	if err != nil {
		return nil, err
	}

	if paymentCoin.Amount.IsZero() {
		return nil, types.ErrInvalidAmount
	}

	actualPayment := types.CoinToDec(paymentCoin)

	tokensCoin, err := types.DecToCoin(tokenAmountDec, msg.Denom)
	if err != nil {
		return nil, err
	}

	actualTokens := types.CoinToDec(tokensCoin)

	if actualTokens.IsZero() {
		return nil, types.ErrInvalidAmount
	}

	if actualTokens.GT(tokensSold) {
		return nil, types.ErrInvalidAmount
	}

	if msg.MinPaymentOut != "" {
		minPayment, err := osmomath.NewDecFromStr(msg.MinPaymentOut)
		if err != nil {
			return nil, types.ErrInvalidAmount
		}
		if actualPayment.LT(minPayment) {
			return nil, types.ErrMinPaymentNotMet
		}
	}

	// ensure reserves sufficient
	if msg.PaymentDenom == params.QuoteDenom {
		if pool.ReserveNdollarDec().LT(actualPayment) {
			return nil, types.ErrInsufficientLiquidity
		}
	} else {
		if pool.ReserveNuahDec().LT(actualPayment) {
			return nil, types.ErrInsufficientLiquidity
		}
	}

	// transfer tokens from trader back to module wallet
	err = s.receiveTokens(ctx, trader, tokensCoin)
	if err != nil {
		return nil, err
	}

	// pay out quote
	err = s.bankKeeper.SendCoinsFromModuleToAccount(sdk.WrapSDKContext(ctx), types.ModuleName, trader, sdk.NewCoins(paymentCoin))
	if err != nil {
		return nil, err
	}

	// update pool
	pool.SetTokensSold(tokensSold.Sub(actualTokens))
	if msg.PaymentDenom == params.QuoteDenom {
		pool.SetReserveNdollar(pool.ReserveNdollarDec().Sub(actualPayment))
	} else {
		pool.SetReserveNuah(pool.ReserveNuahDec().Sub(actualPayment))
	}

	lastPrice := types.CalculatePrice(pool.TokensSoldDec(), params)
	pool.SetLastPrice(lastPrice)
	s.setPool(ctx, pool)

	if err := s.updateTokenState(ctx, msg.Denom, pool.TokensSoldDec(), lastPrice); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeSell,
		sdk.NewAttribute(types.AttributeKeyTrader, msg.Trader),
		sdk.NewAttribute(types.AttributeKeyDenom, msg.Denom),
		sdk.NewAttribute(types.AttributeKeyTokensIn, actualTokens.String()),
		sdk.NewAttribute(types.AttributeKeyPaymentDenom, msg.PaymentDenom),
		sdk.NewAttribute(types.AttributeKeyPaymentOut, actualPayment.String()),
		sdk.NewAttribute(types.AttributeKeyPrice, lastPrice.String()),
	))

	return &types.MsgSellToCurveResponse{
		PaymentOut:    actualPayment.String(),
		PriceReceived: actualPayment.String(),
	}, nil
}

func (s msgServer) OpenMarginPosition(goCtx context.Context, msg *types.MsgOpenMarginPosition) (*types.MsgOpenMarginPositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	trader, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return nil, err
	}

	if msg.PositionType == types.PositionType_POSITION_TYPE_UNSPECIFIED {
		return nil, types.ErrInvalidParams
	}

	if !types.ValidateLeverage(msg.Leverage) {
		return nil, types.ErrInvalidLeverage
	}

	collateralDec, err := osmomath.NewDecFromStr(msg.CollateralAmount)
	if err != nil || !collateralDec.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	positionSize := types.PositionSize(collateralDec, msg.Leverage)
	if !positionSize.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	if msg.MinPositionSize != "" {
		minSize, err := osmomath.NewDecFromStr(msg.MinPositionSize)
		if err != nil {
			return nil, types.ErrInvalidAmount
		}
		if positionSize.LT(minSize) {
			return nil, types.ErrMinPositionNotMet
		}
	}

	params := s.GetParams(ctx)
	if msg.CollateralDenom != params.QuoteDenom && msg.CollateralDenom != sdk.DefaultBondDenom {
		return nil, types.ErrUnsupportedDenom
	}

	pool := s.ensurePool(ctx, msg.Denom)
	marginPool := s.ensureMarginPool(ctx, msg.Denom)

	available := marginPool.AvailableLiquidityDec()
	if positionSize.GT(available) {
		return nil, types.ErrInsufficientLiquidity
	}

	newAvailable := available.Sub(positionSize)
	if newAvailable.IsNegative() {
		return nil, types.ErrInsufficientLiquidity
	}

	// Determine price source based on message or default to bonding curve
	priceSource := msg.PriceSource
	if priceSource == types.PriceSource_PRICE_SOURCE_UNSPECIFIED {
		priceSource = types.PriceSource_PRICE_SOURCE_BONDING_CURVE // Default to bonding curve
	}

	var entryPrice osmomath.Dec

	switch priceSource {
	case types.PriceSource_PRICE_SOURCE_DEX_POOL:
		var priceErr error
		entryPrice, priceErr = s.getDexPoolPrice(ctx, pool, params)
		if priceErr != nil {
			return nil, fmt.Errorf("failed to get DEX pool price: %w", priceErr)
		}
	case types.PriceSource_PRICE_SOURCE_BONDING_CURVE:
		fallthrough
	default:
		entryPrice = pool.LastPriceDec()
		if !entryPrice.IsPositive() {
			entryPrice = types.CalculatePrice(pool.TokensSoldDec(), params)
		}
		if !entryPrice.IsPositive() {
			entryPrice = params.StartPriceDec()
		}
	}

	if !entryPrice.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	liquidationPrice, err := types.CalculateLiquidationPrice(entryPrice, msg.Leverage, msg.PositionType)
	if err != nil {
		return nil, types.ErrInvalidLiquidation
	}

	maintenanceMargin := types.CalculateMaintenanceMargin(positionSize)
	if collateralDec.LT(maintenanceMargin) {
		return nil, types.ErrMarginInsufficient
	}

	collateralCoin, err := types.DecToCoin(collateralDec, msg.CollateralDenom)
	if err != nil {
		return nil, err
	}
	if collateralCoin.Amount.IsZero() {
		return nil, types.ErrInvalidAmount
	}

	if err := s.bankKeeper.SendCoinsFromAccountToModule(sdk.WrapSDKContext(ctx), trader, types.ModuleName, sdk.NewCoins(collateralCoin)); err != nil {
		return nil, err
	}

	positionID := marginPool.NextPositionId
	marginPool.NextPositionId++
	marginPool.SetTotalCollateral(marginPool.TotalCollateralDec().Add(collateralDec))
	marginPool.SetAvailableLiquidity(newAvailable)

	if msg.PositionType == types.PositionType_POSITION_TYPE_LONG {
		marginPool.SetTotalLongExposure(marginPool.TotalLongExposureDec().Add(positionSize))
	} else {
		marginPool.SetTotalShortExposure(marginPool.TotalShortExposureDec().Add(positionSize))
	}
	marginPool.LastFundingTimestamp = uint64(ctx.BlockTime().Unix())
	s.setMarginPool(ctx, marginPool)

	position := types.MarginPosition{
		Id:                positionID,
		Trader:            msg.Trader,
		Denom:             msg.Denom,
		CollateralDenom:   msg.CollateralDenom,
		CollateralAmount:  collateralDec.String(),
		PositionSize:      positionSize.String(),
		EntryPrice:        entryPrice.String(),
		Leverage:          msg.Leverage,
		Type:              msg.PositionType,
		CreatedAt:         uint64(ctx.BlockTime().Unix()),
		LiquidationPrice:  liquidationPrice.String(),
		MaintenanceMargin: maintenanceMargin.String(),
		FundingFeeAccrued: types.DefaultFundingRate.String(),
		Status:            types.PositionStatus_POSITION_STATUS_OPEN,
		RealizedPnl:       osmomath.ZeroDec().String(),
		LastMarkPrice:     entryPrice.String(),
		PriceSource:       priceSource,
	}

	s.setMarginPosition(ctx, position)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeOpenMargin,
		sdk.NewAttribute(types.AttributeKeyTrader, msg.Trader),
		sdk.NewAttribute(types.AttributeKeyDenom, msg.Denom),
		sdk.NewAttribute(types.AttributeKeyCollateralDenom, msg.CollateralDenom),
		sdk.NewAttribute(types.AttributeKeyCollateralAmount, collateralDec.String()),
		sdk.NewAttribute(types.AttributeKeyLeverage, fmt.Sprintf("%d", msg.Leverage)),
		sdk.NewAttribute(types.AttributeKeyPositionSize, positionSize.String()),
		sdk.NewAttribute(types.AttributeKeyEntryPrice, entryPrice.String()),
		sdk.NewAttribute(types.AttributeKeyLiquidationPrice, liquidationPrice.String()),
		sdk.NewAttribute(types.AttributeKeyPositionId, fmt.Sprintf("%d", positionID)),
	))

	return &types.MsgOpenMarginPositionResponse{
		PositionId:       positionID,
		PositionSize:     positionSize.String(),
		EntryPrice:       entryPrice.String(),
		LiquidationPrice: liquidationPrice.String(),
	}, nil
}

func (s msgServer) CloseMarginPosition(goCtx context.Context, msg *types.MsgCloseMarginPosition) (*types.MsgCloseMarginPositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	trader, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return nil, err
	}

	position, found := s.GetMarginPosition(ctx, msg.PositionId)
	if !found {
		return nil, types.ErrPositionNotFound
	}

	if position.Trader != msg.Trader {
		return nil, types.ErrUnauthorizedPosition
	}

	if position.Status != types.PositionStatus_POSITION_STATUS_OPEN {
		return nil, types.ErrPositionClosed
	}

	params := s.GetParams(ctx)
	pool := s.ensurePool(ctx, position.Denom)
	marginPool := s.ensureMarginPool(ctx, position.Denom)

	entryPrice := position.EntryPriceDec()
	if !entryPrice.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	// Get current price based on position's price source
	var currentPrice osmomath.Dec

	switch position.PriceSource {
	case types.PriceSource_PRICE_SOURCE_DEX_POOL:
		var priceErr error
		currentPrice, priceErr = s.getDexPoolPrice(ctx, pool, params)
		if priceErr != nil {
			return nil, fmt.Errorf("failed to get DEX pool price for closing position: %w", priceErr)
		}
	case types.PriceSource_PRICE_SOURCE_BONDING_CURVE:
		fallthrough
	default:
		currentPrice = pool.LastPriceDec()
		if !currentPrice.IsPositive() {
			currentPrice = types.CalculatePrice(pool.TokensSoldDec(), params)
		}
		if !currentPrice.IsPositive() {
			currentPrice = params.StartPriceDec()
		}
	}

	if !currentPrice.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	positionSize := position.PositionSizeDec()
	if !positionSize.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	collateral := position.CollateralAmountDec()
	if !collateral.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	priceDiff := currentPrice.Sub(entryPrice)
	if position.Type == types.PositionType_POSITION_TYPE_SHORT {
		priceDiff = entryPrice.Sub(currentPrice)
	}

	if priceDiff.IsZero() {
		priceDiff = osmomath.ZeroDec()
	}

	denomFactor := entryPrice
	if !denomFactor.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	percentage := priceDiff.Quo(denomFactor)
	pnl := percentage.Mul(positionSize)

	fee := positionSize.Mul(types.DefaultMarginFeeRate)
	payoutDec := collateral.Add(pnl).Sub(fee)
	if payoutDec.IsNegative() {
		payoutDec = osmomath.ZeroDec()
	}

	var payoutCoin sdk.Coin
	actualPayoutDec := osmomath.ZeroDec()
	if payoutDec.IsPositive() {
		coin, err := types.DecToCoin(payoutDec, position.CollateralDenom)
		if err == nil {
			payoutCoin = coin
			actualPayoutDec = types.CoinToDec(coin)
		}
	}

	if msg.MinPayout != "" {
		minPayout, err := osmomath.NewDecFromStr(msg.MinPayout)
		if err != nil {
			return nil, types.ErrInvalidAmount
		}
		if actualPayoutDec.LT(minPayout) {
			return nil, types.ErrMinPaymentNotMet
		}
	}

	updatedCollateral := marginPool.TotalCollateralDec().Sub(collateral)
	if updatedCollateral.IsNegative() {
		updatedCollateral = osmomath.ZeroDec()
	}
	marginPool.SetTotalCollateral(updatedCollateral)
	marginPool.SetAvailableLiquidity(marginPool.AvailableLiquidityDec().Add(positionSize).Add(fee))

	if position.Type == types.PositionType_POSITION_TYPE_LONG {
		updatedLong := marginPool.TotalLongExposureDec().Sub(positionSize)
		if updatedLong.IsNegative() {
			updatedLong = osmomath.ZeroDec()
		}
		marginPool.SetTotalLongExposure(updatedLong)
	} else {
		updatedShort := marginPool.TotalShortExposureDec().Sub(positionSize)
		if updatedShort.IsNegative() {
			updatedShort = osmomath.ZeroDec()
		}
		marginPool.SetTotalShortExposure(updatedShort)
	}

	marginPool.LastFundingTimestamp = uint64(ctx.BlockTime().Unix())
	s.setMarginPool(ctx, marginPool)

	position.SetStatus(types.PositionStatus_POSITION_STATUS_CLOSED)
	position.SetPositionSize(osmomath.ZeroDec())
	position.SetCollateralAmount(osmomath.ZeroDec())
	position.SetMaintenanceMargin(osmomath.ZeroDec())
	position.SetLiquidationPrice(osmomath.ZeroDec())
	position.SetRealizedPnl(position.RealizedPnLDec().Add(pnl))
	position.SetLastMarkPrice(currentPrice)
	s.setMarginPosition(ctx, position)

	if payoutCoin.Amount.IsPositive() {
		if err := s.bankKeeper.SendCoinsFromModuleToAccount(sdk.WrapSDKContext(ctx), types.ModuleName, trader, sdk.NewCoins(payoutCoin)); err != nil {
			return nil, err
		}
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCloseMargin,
		sdk.NewAttribute(types.AttributeKeyTrader, msg.Trader),
		sdk.NewAttribute(types.AttributeKeyDenom, position.Denom),
		sdk.NewAttribute(types.AttributeKeyPositionId, fmt.Sprintf("%d", msg.PositionId)),
		sdk.NewAttribute(types.AttributeKeyPayoutAmount, actualPayoutDec.String()),
		sdk.NewAttribute(types.AttributeKeyRealizedPnL, pnl.String()),
		sdk.NewAttribute(types.AttributeKeyFeesPaid, fee.String()),
	))

	return &types.MsgCloseMarginPositionResponse{
		PayoutAmount: actualPayoutDec.String(),
		RealizedPnl:  pnl.String(),
		FeesPaid:     fee.String(),
	}, nil
}

func (s msgServer) distributeTokens(ctx sdk.Context, trader sdk.AccAddress, coin sdk.Coin) error {
	wallet, err := s.bondingCurveWallet(ctx)
	if err != nil {
		return err
	}

	if wallet.Equals(s.GetModuleAddress()) {
		return s.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, trader, sdk.NewCoins(coin))
	}

	return s.bankKeeper.SendCoins(sdk.WrapSDKContext(ctx), wallet, trader, sdk.NewCoins(coin))
}

func (s msgServer) receiveTokens(ctx sdk.Context, trader sdk.AccAddress, coin sdk.Coin) error {
	wallet, err := s.bondingCurveWallet(ctx)
	if err != nil {
		return err
	}

	if wallet.Equals(s.GetModuleAddress()) {
		return s.bankKeeper.SendCoinsFromAccountToModule(ctx, trader, types.ModuleName, sdk.NewCoins(coin))
	}

	return s.bankKeeper.SendCoins(sdk.WrapSDKContext(ctx), trader, wallet, sdk.NewCoins(coin))
}

// getDexPoolPrice retrieves the current price from the DEX pool
func (s msgServer) getDexPoolPrice(ctx sdk.Context, pool types.BondingCurvePool, params types.Params) (osmomath.Dec, error) {
	// Check if DEX pool is activated
	if !pool.DexActivated || pool.DexPoolId == 0 {
		return osmomath.ZeroDec(), fmt.Errorf("DEX pool not activated for denom %s", pool.Denom)
	}

	// Get quote denom
	quoteDenom := params.QuoteDenom
	if quoteDenom == "" {
		quoteDenom = "unuah" // Default quote denom
	}

	// Get spot price from DEX pool using pool manager
	spotPrice, err := s.poolManager.RouteCalculateSpotPrice(
		ctx,
		pool.DexPoolId,
		quoteDenom, // quote asset (e.g., "unuah")
		pool.Denom, // base asset (e.g., "factory/nuah1.../ttt")
	)
	if err != nil {
		return osmomath.ZeroDec(), fmt.Errorf("failed to calculate spot price from DEX pool %d: %w", pool.DexPoolId, err)
	}

	// Convert BigDec to Dec
	price := spotPrice.Dec()
	if !price.IsPositive() {
		return osmomath.ZeroDec(), fmt.Errorf("DEX pool price is not positive: %s", price.String())
	}

	return price, nil
}

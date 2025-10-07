package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/leverage/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// OpenPosition opens a new leverage position
func (k msgServer) OpenPosition(goCtx context.Context, msg *types.MsgOpenPosition) (*types.MsgOpenPositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate trader address
	traderAddr, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return nil, fmt.Errorf("invalid trader address: %w", err)
	}

	// Validate parameters
	params := k.GetParams(ctx)

	// Check max leverage
	if msg.Leverage.GT(params.MaxLeverage) {
		return nil, types.ErrMaxLeverageExceeded
	}

	// Check min collateral
	if msg.Collateral.Amount.LT(params.MinCollateralAmount) {
		return nil, types.ErrMinCollateralNotMet
	}

	// Validate collateral denomination
	if !k.ValidateCollateralDenom(ctx, msg.Collateral.Denom) {
		return nil, types.ErrInvalidCollateralDenom
	}

	// Validate token denomination
	if !k.ValidateTokenDenom(ctx, msg.TokenDenom) {
		return nil, types.ErrTokenNotSupported
	}

	// Get current token price
	currentPrice, err := k.GetTokenPrice(ctx, msg.TokenDenom)
	if err != nil {
		return nil, fmt.Errorf("failed to get token price: %w", err)
	}

	// Check slippage protection
	if msg.Side == types.PositionSideLong && currentPrice.GT(msg.MaxPrice) {
		return nil, types.ErrSlippageExceeded
	}
	if msg.Side == types.PositionSideShort && currentPrice.LT(msg.MinPrice) {
		return nil, types.ErrSlippageExceeded
	}

	// Calculate position size with proper precision handling
	collateralDec := math.LegacyNewDecFromInt(msg.Collateral.Amount)
	positionValue := collateralDec.Mul(msg.Leverage)
	positionSizeDec := positionValue.Quo(currentPrice)

	// Check if position size is too small (less than 1 token)
	if positionSizeDec.LT(math.LegacyOneDec()) {
		return nil, fmt.Errorf("position size too small: %s tokens (minimum 1 token required)", positionSizeDec.String())
	}

	// Use RoundInt() instead of TruncateInt() for better precision
	positionSize := positionSizeDec.RoundInt()

	// Check max position size
	if positionSize.GT(params.MaxPositionSize) {
		return nil, types.ErrMaxPositionSizeExceeded
	}

	// Calculate trading fee
	tradingFeeAmount := positionValue.Mul(params.TradingFee).TruncateInt()

	// Calculate liquidation price (will be updated after we know actual position size)
	var liquidationPrice math.LegacyDec

	// Transfer collateral from trader to module
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if moduleAddr == nil {
		return nil, fmt.Errorf("module address not found")
	}

	// Transfer collateral + trading fee
	totalAmount := msg.Collateral.Amount.Add(tradingFeeAmount)
	collateralCoin := sdk.NewCoin(msg.Collateral.Denom, totalAmount)

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, traderAddr, types.ModuleName, sdk.NewCoins(collateralCoin))
	if err != nil {
		return nil, fmt.Errorf("failed to transfer collateral: %w", err)
	}

	// Generate position ID first
	positionID := k.GetNextPositionID(ctx)

	// For leverage trading, we simulate buying tokens with borrowed funds
	// In a real implementation, this would involve a lending protocol
	var tokensReceived math.Int
	borrowAmount := math.ZeroInt()
	var borrowErr error
	if msg.Side == types.PositionSideLong {
		// For LONG: actually borrow additional base currency and buy tokens
		borrowDec := positionValue.Sub(collateralDec)
		if borrowDec.IsPositive() {
			borrowAmount = borrowDec.TruncateInt()
		}

		var borrowID string
		if borrowAmount.IsPositive() {
			borrowID, borrowErr = k.lendingKeeper.BorrowTokens(ctx, traderAddr, msg.Collateral.Denom, borrowAmount, positionID)
			if borrowErr != nil {
				return nil, fmt.Errorf("failed to borrow collateral for LONG position: %w", borrowErr)
			}

			borrowCoin := sdk.NewCoin(msg.Collateral.Denom, borrowAmount)
			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, traderAddr, types.ModuleName, sdk.NewCoins(borrowCoin)); err != nil {
				_ = k.lendingKeeper.RepayTokens(ctx, traderAddr, borrowID, borrowAmount)
				return nil, fmt.Errorf("failed to move borrowed collateral to module: %w", err)
			}
		}

		totalPurchase := msg.Collateral.Amount.Add(borrowAmount)
		purchaseTokens, err := k.userTokenKeeper.ExecuteBuyTokens(ctx, moduleAddr, msg.TokenDenom, totalPurchase, msg.Collateral.Denom)
		if err != nil {
			// Refund collateral (including trading fee) and repay borrow if needed
			refund := sdk.NewCoin(msg.Collateral.Denom, totalAmount)
			_ = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, traderAddr, sdk.NewCoins(refund))
			if borrowAmount.IsPositive() {
				borrowCoin := sdk.NewCoin(msg.Collateral.Denom, borrowAmount)
				_ = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, traderAddr, sdk.NewCoins(borrowCoin))
				if borrowID != "" {
					_ = k.lendingKeeper.RepayTokens(ctx, traderAddr, borrowID, borrowAmount)
				}
			}
			return nil, fmt.Errorf("failed to buy tokens for LONG position: %w", err)
		}

		tokensReceived = purchaseTokens
		liquidationPrice = k.CalculateLiquidationPrice(ctx, msg.Collateral.Amount, tokensReceived, currentPrice, msg.Side)
	} else {
		// For SHORT: Borrow tokens from lending pool and sell them immediately
		borrowID, err := k.lendingKeeper.BorrowTokens(ctx, traderAddr, msg.TokenDenom, positionSize, positionID)
		if err != nil {
			return nil, fmt.Errorf("failed to borrow tokens for SHORT position: %w", err)
		}

		// Sell borrowed tokens immediately to get proceeds
		_, _, err = k.userTokenKeeper.ExecuteSellTokens(ctx, traderAddr, msg.TokenDenom, positionSize)
		if err != nil {
			// If sell fails, repay the borrowed tokens
			_ = k.lendingKeeper.RepayTokens(ctx, traderAddr, borrowID, positionSize)
			return nil, fmt.Errorf("failed to sell borrowed tokens: %w", err)
		}

		// For SHORT positions, the position size is the amount of tokens borrowed (negative exposure)
		// The proceeds are what we get from selling, but the position size is the borrowed amount
		tokensReceived = positionSize

		// Calculate liquidation price for SHORT position
		liquidationPrice = k.CalculateLiquidationPrice(ctx, msg.Collateral.Amount, tokensReceived, currentPrice, msg.Side)

		// Store borrow ID in position for later repayment
		// We'll add this field to Position struct later if needed
	}

	// Create position
	position := types.Position{
		Id:               positionID,
		Trader:           msg.Trader,
		TokenDenom:       msg.TokenDenom,
		CollateralDenom:  msg.Collateral.Denom,
		Side:             msg.Side,
		Size_:            tokensReceived,
		Collateral:       msg.Collateral.Amount,
		Leverage:         msg.Leverage,
		EntryPrice:       currentPrice,
		LiquidationPrice: liquidationPrice,
		UnrealizedPnl:    math.ZeroInt(),
		Status:           types.PositionStatusOpen,
		CreatedAt:        ctx.BlockTime().Unix(),
		UpdatedAt:        ctx.BlockTime().Unix(),
	}

	// Store position
	k.SetPosition(ctx, position)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"open_position",
			sdk.NewAttribute("position_id", positionID),
			sdk.NewAttribute("trader", msg.Trader),
			sdk.NewAttribute("token_denom", msg.TokenDenom),
			sdk.NewAttribute("side", msg.Side.String()),
			sdk.NewAttribute("size", tokensReceived.String()),
			sdk.NewAttribute("collateral", msg.Collateral.Amount.String()),
			sdk.NewAttribute("leverage", msg.Leverage.String()),
			sdk.NewAttribute("entry_price", currentPrice.String()),
			sdk.NewAttribute("liquidation_price", liquidationPrice.String()),
		),
	})

	return &types.MsgOpenPositionResponse{
		PositionId: positionID,
		Position:   position,
	}, nil
}

// ClosePosition closes an existing position
func (k msgServer) ClosePosition(goCtx context.Context, msg *types.MsgClosePosition) (*types.MsgClosePositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate trader address
	traderAddr, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return nil, fmt.Errorf("invalid trader address: %w", err)
	}

	// Get position
	position, found := k.GetPosition(ctx, msg.PositionId)
	if !found {
		return nil, types.ErrPositionNotFound
	}

	// Check ownership
	if position.Trader != msg.Trader {
		return nil, types.ErrUnauthorized
	}

	// Check if position is open
	if position.Status != types.PositionStatusOpen {
		return nil, types.ErrPositionAlreadyClosed
	}

	// Get current token price
	currentPrice, err := k.GetTokenPrice(ctx, position.TokenDenom)
	if err != nil {
		return nil, fmt.Errorf("failed to get token price: %w", err)
	}

	// Check slippage protection
	if position.Side == types.PositionSideLong && currentPrice.LT(msg.MinPrice) {
		return nil, types.ErrSlippageExceeded
	}
	if position.Side == types.PositionSideShort && currentPrice.GT(msg.MaxPrice) {
		return nil, types.ErrSlippageExceeded
	}

	// Calculate realized PnL
	k.UpdatePositionPnL(ctx, &position, currentPrice)
	realizedPnL := position.UnrealizedPnl

	// Handle position closing based on side
	var collateralToReturn math.Int

	if position.Side == types.PositionSideLong {
		// For LONG: sell held tokens, repay any borrowed collateral, return remaining proceeds
		moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
		if moduleAddr == nil {
			return nil, fmt.Errorf("module address not found")
		}

		// Transfer tokens from module to trader for sale
		tokenCoin := sdk.NewCoin(position.TokenDenom, position.Size_)
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, traderAddr, sdk.NewCoins(tokenCoin))
		if err != nil {
			return nil, fmt.Errorf("failed to transfer position tokens: %w", err)
		}

		// Sell tokens back through bonding curve
		payout, payoutDenom, err := k.userTokenKeeper.ExecuteSellTokens(ctx, traderAddr, position.TokenDenom, position.Size_)
		if err != nil {
			// Return tokens to module in case of failure
			_ = k.bankKeeper.SendCoinsFromAccountToModule(ctx, traderAddr, types.ModuleName, sdk.NewCoins(tokenCoin))
			return nil, fmt.Errorf("failed to sell tokens for LONG position: %w", err)
		}

		// Move proceeds back to leverage module for accounting
		payoutCoin := sdk.NewCoin(payoutDenom, payout)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, traderAddr, types.ModuleName, sdk.NewCoins(payoutCoin)); err != nil {
			return nil, fmt.Errorf("failed to capture sale proceeds: %w", err)
		}

		// Repay any borrowed collateral using sale proceeds
		debtPaid := math.ZeroInt()
		if borrowID, found := k.lendingKeeper.GetBorrowIDByLeveragePosition(ctx, position.Id); found {
			borrowPos, found := k.lendingKeeper.GetBorrowPosition(ctx, borrowID)
			if found {
				debt := borrowPos.GetTotalDebt()
				if debt.IsPositive() {
					debtCoin := sdk.NewCoin(position.CollateralDenom, debt)
					if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, traderAddr, sdk.NewCoins(debtCoin)); err != nil {
						return nil, fmt.Errorf("failed to provision repayment funds: %w", err)
					}

					if err := k.lendingKeeper.RepayTokens(ctx, traderAddr, borrowID, debt); err != nil {
						return nil, fmt.Errorf("failed to repay borrowed collateral: %w", err)
					}
					debtPaid = debt
				}
			}
		}

		// Remaining proceeds are collateral plus realized PnL
		collateralToReturn = payout.Sub(debtPaid)
		if collateralToReturn.IsNegative() {
			collateralToReturn = math.ZeroInt()
		}
		realizedPnL = collateralToReturn.Sub(position.Collateral)
		position.UnrealizedPnl = math.ZeroInt()
	} else {
		// For SHORT: Buy back tokens to repay the borrowed amount
		// Find the associated borrow position
		if borrowID, found := k.lendingKeeper.GetBorrowIDByLeveragePosition(ctx, position.Id); found {
			// Get borrow position to know how much to repay
			borrowPos, found := k.lendingKeeper.GetBorrowPosition(ctx, borrowID)
			if found {
				// Calculate cost to buy back borrowed tokens
				buybackCost := math.LegacyNewDecFromInt(borrowPos.BorrowedAmount).Mul(currentPrice).TruncateInt()

				// Buy back tokens to repay loan
				_, err = k.userTokenKeeper.ExecuteBuyTokens(ctx, traderAddr, position.TokenDenom, buybackCost, position.CollateralDenom)
				if err != nil {
					return nil, fmt.Errorf("failed to buy back tokens for SHORT position: %w", err)
				}

				// Repay the borrowed tokens
				err = k.lendingKeeper.RepayTokens(ctx, traderAddr, borrowID, borrowPos.BorrowedAmount)
				if err != nil {
					return nil, fmt.Errorf("failed to repay borrowed tokens: %w", err)
				}
			}
		}

		// Calculate final collateral to return (original collateral + PnL - buyback cost)
		collateralToReturn = position.Collateral.Add(realizedPnL)
		if collateralToReturn.IsNegative() {
			collateralToReturn = math.ZeroInt()
		}
	}

	// Return collateral to trader
	if collateralToReturn.IsPositive() {
		collateralCoin := sdk.NewCoin(position.CollateralDenom, collateralToReturn)
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, traderAddr, sdk.NewCoins(collateralCoin))
		if err != nil {
			return nil, fmt.Errorf("failed to return collateral: %w", err)
		}
	}

	// Update position status
	position.Status = types.PositionStatusClosed
	position.UpdatedAt = ctx.BlockTime().Unix()
	k.SetPosition(ctx, position)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"close_position",
			sdk.NewAttribute("position_id", msg.PositionId),
			sdk.NewAttribute("trader", msg.Trader),
			sdk.NewAttribute("realized_pnl", realizedPnL.String()),
			sdk.NewAttribute("collateral_returned", collateralToReturn.String()),
			sdk.NewAttribute("close_price", currentPrice.String()),
		),
	})

	return &types.MsgClosePositionResponse{
		RealizedPnl:        realizedPnL,
		CollateralReturned: sdk.NewCoin(position.CollateralDenom, collateralToReturn),
	}, nil
}

// AddCollateral adds collateral to an existing position
func (k msgServer) AddCollateral(goCtx context.Context, msg *types.MsgAddCollateral) (*types.MsgAddCollateralResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate trader address
	traderAddr, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return nil, fmt.Errorf("invalid trader address: %w", err)
	}

	// Get position
	position, found := k.GetPosition(ctx, msg.PositionId)
	if !found {
		return nil, types.ErrPositionNotFound
	}

	// Check ownership
	if position.Trader != msg.Trader {
		return nil, types.ErrUnauthorized
	}

	// Check if position is open
	if position.Status != types.PositionStatusOpen {
		return nil, types.ErrPositionAlreadyClosed
	}

	// Check collateral denomination matches
	if msg.Amount.Denom != position.CollateralDenom {
		return nil, types.ErrInvalidCollateralDenom
	}

	// Transfer additional collateral
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, traderAddr, types.ModuleName, sdk.NewCoins(msg.Amount))
	if err != nil {
		return nil, fmt.Errorf("failed to transfer collateral: %w", err)
	}

	// Update position
	position.Collateral = position.Collateral.Add(msg.Amount.Amount)

	// Recalculate liquidation price
	position.LiquidationPrice = k.CalculateLiquidationPrice(ctx, position.Collateral, position.Size_, position.EntryPrice, position.Side)
	position.UpdatedAt = ctx.BlockTime().Unix()

	k.SetPosition(ctx, position)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"add_collateral",
			sdk.NewAttribute("position_id", msg.PositionId),
			sdk.NewAttribute("trader", msg.Trader),
			sdk.NewAttribute("amount_added", msg.Amount.Amount.String()),
			sdk.NewAttribute("new_collateral", position.Collateral.String()),
			sdk.NewAttribute("new_liquidation_price", position.LiquidationPrice.String()),
		),
	})

	return &types.MsgAddCollateralResponse{
		NewLiquidationPrice: position.LiquidationPrice,
	}, nil
}

// RemoveCollateral removes collateral from an existing position
func (k msgServer) RemoveCollateral(goCtx context.Context, msg *types.MsgRemoveCollateral) (*types.MsgRemoveCollateralResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate trader address
	traderAddr, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return nil, fmt.Errorf("invalid trader address: %w", err)
	}

	// Get position
	position, found := k.GetPosition(ctx, msg.PositionId)
	if !found {
		return nil, types.ErrPositionNotFound
	}

	// Check ownership
	if position.Trader != msg.Trader {
		return nil, types.ErrUnauthorized
	}

	// Check if position is open
	if position.Status != types.PositionStatusOpen {
		return nil, types.ErrPositionAlreadyClosed
	}

	// Check collateral denomination matches
	if msg.Amount.Denom != position.CollateralDenom {
		return nil, types.ErrInvalidCollateralDenom
	}

	// Check if there's enough collateral to remove
	if msg.Amount.Amount.GTE(position.Collateral) {
		return nil, types.ErrInsufficientCollateral
	}

	// Calculate new collateral amount
	newCollateral := position.Collateral.Sub(msg.Amount.Amount)

	// Check if new liquidation price would be safe
	newLiquidationPrice := k.CalculateLiquidationPrice(ctx, newCollateral, position.Size_, position.EntryPrice, position.Side)

	// Get current price to ensure position won't be immediately liquidated
	currentPrice, err := k.GetTokenPrice(ctx, position.TokenDenom)
	if err != nil {
		return nil, fmt.Errorf("failed to get token price: %w", err)
	}

	// Check if position would be liquidatable after removing collateral
	if position.Side == types.PositionSideLong && currentPrice.LTE(newLiquidationPrice) {
		return nil, fmt.Errorf("removing collateral would make position liquidatable")
	}
	if position.Side == types.PositionSideShort && currentPrice.GTE(newLiquidationPrice) {
		return nil, fmt.Errorf("removing collateral would make position liquidatable")
	}

	// Transfer collateral back to trader
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, traderAddr, sdk.NewCoins(msg.Amount))
	if err != nil {
		return nil, fmt.Errorf("failed to return collateral: %w", err)
	}

	// Update position
	position.Collateral = newCollateral
	position.LiquidationPrice = newLiquidationPrice
	position.UpdatedAt = ctx.BlockTime().Unix()

	k.SetPosition(ctx, position)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"remove_collateral",
			sdk.NewAttribute("position_id", msg.PositionId),
			sdk.NewAttribute("trader", msg.Trader),
			sdk.NewAttribute("amount_removed", msg.Amount.Amount.String()),
			sdk.NewAttribute("new_collateral", position.Collateral.String()),
			sdk.NewAttribute("new_liquidation_price", position.LiquidationPrice.String()),
		),
	})

	return &types.MsgRemoveCollateralResponse{
		NewLiquidationPrice: newLiquidationPrice,
	}, nil
}

// LiquidatePosition liquidates a position that has reached liquidation threshold
func (k msgServer) LiquidatePosition(goCtx context.Context, msg *types.MsgLiquidatePosition) (*types.MsgLiquidatePositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate liquidator address
	liquidatorAddr, err := sdk.AccAddressFromBech32(msg.Liquidator)
	if err != nil {
		return nil, fmt.Errorf("invalid liquidator address: %w", err)
	}

	// Get position
	position, found := k.GetPosition(ctx, msg.PositionId)
	if !found {
		return nil, types.ErrPositionNotFound
	}

	// Check if position is open
	if position.Status != types.PositionStatusOpen {
		return nil, types.ErrPositionAlreadyClosed
	}

	// Get current token price
	currentPrice, err := k.GetTokenPrice(ctx, position.TokenDenom)
	if err != nil {
		return nil, fmt.Errorf("failed to get token price: %w", err)
	}

	// Check if position is liquidatable
	if !k.IsPositionLiquidatable(ctx, position, currentPrice) {
		return nil, types.ErrPositionNotLiquidatable
	}

	// Calculate liquidation parameters
	params := k.GetParams(ctx)
	liquidationFeeAmount := math.LegacyNewDecFromInt(position.Collateral).Mul(params.LiquidationFee).TruncateInt()

	// Execute liquidation trade
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	if position.Side == types.PositionSideLong {
		// For LONG: Sell all tokens
		_, _, err = k.userTokenKeeper.ExecuteSellTokens(ctx, moduleAddr, position.TokenDenom, position.Size_)
		if err != nil {
			return nil, fmt.Errorf("failed to execute liquidation sell: %w", err)
		}
	} else {
		// For SHORT: Buy back tokens to repay the borrowed amount
		if borrowID, found := k.lendingKeeper.GetBorrowIDByLeveragePosition(ctx, position.Id); found {
			// Get borrow position to know how much to repay
			borrowPos, found := k.lendingKeeper.GetBorrowPosition(ctx, borrowID)
			if found {
				// Calculate cost to buy back borrowed tokens
				buybackCost := math.LegacyNewDecFromInt(borrowPos.GetTotalDebt()).Mul(currentPrice).TruncateInt()

				// Buy back tokens using liquidator funds
				_, err = k.userTokenKeeper.ExecuteBuyTokens(ctx, liquidatorAddr, position.TokenDenom, buybackCost, position.CollateralDenom)
				if err != nil {
					return nil, fmt.Errorf("failed to execute liquidation buyback: %w", err)
				}

				// Repay the borrowed tokens (liquidator pays the debt)
				err = k.lendingKeeper.RepayTokens(ctx, liquidatorAddr, borrowID, borrowPos.GetTotalDebt())
				if err != nil {
					return nil, fmt.Errorf("failed to repay borrowed tokens during liquidation: %w", err)
				}
			}
		}
	}

	// Calculate remaining collateral after liquidation
	remainingCollateral := position.Collateral.Sub(liquidationFeeAmount)
	if remainingCollateral.IsNegative() {
		remainingCollateral = math.ZeroInt()
		liquidationFeeAmount = position.Collateral
	}

	// Pay liquidation reward to liquidator
	if liquidationFeeAmount.IsPositive() {
		rewardCoin := sdk.NewCoin(position.CollateralDenom, liquidationFeeAmount)
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidatorAddr, sdk.NewCoins(rewardCoin))
		if err != nil {
			return nil, fmt.Errorf("failed to pay liquidation reward: %w", err)
		}
	}

	// Return remaining collateral to position owner (if any)
	var remainingCollateralCoin sdk.Coin
	if remainingCollateral.IsPositive() {
		remainingCollateralCoin = sdk.NewCoin(position.CollateralDenom, remainingCollateral)
		traderAddr, _ := sdk.AccAddressFromBech32(position.Trader)
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, traderAddr, sdk.NewCoins(remainingCollateralCoin))
		if err != nil {
			return nil, fmt.Errorf("failed to return remaining collateral: %w", err)
		}
	} else {
		remainingCollateralCoin = sdk.NewCoin(position.CollateralDenom, math.ZeroInt())
	}

	// Update position status
	position.Status = types.PositionStatusLiquidated
	position.UpdatedAt = ctx.BlockTime().Unix()
	k.SetPosition(ctx, position)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"liquidate_position",
			sdk.NewAttribute("position_id", msg.PositionId),
			sdk.NewAttribute("liquidator", msg.Liquidator),
			sdk.NewAttribute("trader", position.Trader),
			sdk.NewAttribute("liquidation_price", currentPrice.String()),
			sdk.NewAttribute("liquidation_reward", liquidationFeeAmount.String()),
			sdk.NewAttribute("remaining_collateral", remainingCollateral.String()),
		),
	})

	return &types.MsgLiquidatePositionResponse{
		LiquidationReward:   sdk.NewCoin(position.CollateralDenom, liquidationFeeAmount),
		RemainingCollateral: remainingCollateralCoin,
	}, nil
}

// UpdateParams updates the module parameters
func (k msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if k.GetAuthority() != msg.Authority {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", k.GetAuthority(), msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	k.SetParams(ctx, msg.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}

// ProvideLiquidity provides liquidity to a lending pool
func (k msgServer) ProvideLiquidity(goCtx context.Context, msg *types.MsgProvideLiquidity) (*types.MsgProvideLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate provider address
	providerAddr, err := sdk.AccAddressFromBech32(msg.Provider)
	if err != nil {
		return nil, fmt.Errorf("invalid provider address: %w", err)
	}

	// Validate amount
	if msg.Amount.Amount.IsZero() || msg.Amount.Amount.IsNegative() {
		return nil, fmt.Errorf("invalid amount: %s", msg.Amount.Amount.String())
	}

	// Get lending keeper
	lendingKeeper := k.GetLendingKeeper()

	// Provide liquidity
	err = lendingKeeper.ProvideLiquidity(ctx, providerAddr, msg.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to provide liquidity: %w", err)
	}

	// Get the updated liquidity provider record to return share tokens
	lp, found := lendingKeeper.GetLiquidityProvider(ctx, msg.Provider, msg.Amount.Denom)
	if !found {
		return nil, fmt.Errorf("liquidity provider record not found")
	}

	return &types.MsgProvideLiquidityResponse{
		ShareTokens: lp.ShareTokens,
	}, nil
}

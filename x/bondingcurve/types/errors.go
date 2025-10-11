package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidPaymentDenom     = errorsmod.Register(ModuleName, 1, "invalid payment denom")
	ErrInsufficientLiquidity   = errorsmod.Register(ModuleName, 2, "insufficient liquidity")
	ErrMaxSupplyReached        = errorsmod.Register(ModuleName, 3, "max bonding curve supply reached")
	ErrMinTokensNotMet         = errorsmod.Register(ModuleName, 4, "minimum tokens not met")
	ErrMinPaymentNotMet        = errorsmod.Register(ModuleName, 5, "minimum payment not met")
	ErrInvalidAmount           = errorsmod.Register(ModuleName, 6, "invalid amount")
	ErrPoolNotFound            = errorsmod.Register(ModuleName, 7, "bonding curve pool not found")
	ErrUnsupportedDenom        = errorsmod.Register(ModuleName, 8, "unsupported denom")
	ErrInvalidToken            = errorsmod.Register(ModuleName, 9, "invalid token")
	ErrInvalidParams           = errorsmod.Register(ModuleName, 10, "invalid bonding curve params")
	ErrInvalidLeverage         = errorsmod.Register(ModuleName, 11, "invalid leverage")
	ErrMinPositionNotMet       = errorsmod.Register(ModuleName, 12, "minimum position size not met")
	ErrPositionNotFound        = errorsmod.Register(ModuleName, 13, "margin position not found")
	ErrUnauthorizedPosition    = errorsmod.Register(ModuleName, 14, "unauthorized margin position access")
	ErrMarginInsufficient      = errorsmod.Register(ModuleName, 15, "insufficient collateral for maintenance margin")
	ErrInvalidLiquidation      = errorsmod.Register(ModuleName, 16, "invalid liquidation price")
	ErrPriceUnavailable        = errorsmod.Register(ModuleName, 17, "price data unavailable")
	ErrLiquidationsPaused      = errorsmod.Register(ModuleName, 18, "liquidations currently paused")
	ErrCircuitBreaker          = errorsmod.Register(ModuleName, 19, "circuit breaker triggered")
	ErrPositionClosed          = errorsmod.Register(ModuleName, 20, "margin position not open")
	ErrPositionNotLiquidatable = errorsmod.Register(ModuleName, 21, "margin position not eligible for liquidation")
)

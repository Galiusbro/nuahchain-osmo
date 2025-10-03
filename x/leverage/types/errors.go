package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/leverage module sentinel errors
var (
	ErrInvalidAddress          = sdkerrors.Register(ModuleName, 1, "invalid address")
	ErrUnauthorized            = sdkerrors.Register(ModuleName, 2, "unauthorized")
	ErrInvalidAmount           = sdkerrors.Register(ModuleName, 3, "invalid amount")
	ErrPositionNotFound        = sdkerrors.Register(ModuleName, 4, "position not found")
	ErrInvalidRequest          = sdkerrors.Register(ModuleName, 5, "invalid request")
	ErrInvalidLeverage         = sdkerrors.Register(ModuleName, 6, "invalid leverage")
	ErrInvalidPositionSide     = sdkerrors.Register(ModuleName, 7, "invalid position side")
	ErrInsufficientCollateral  = sdkerrors.Register(ModuleName, 8, "insufficient collateral")
	ErrPositionAlreadyClosed   = sdkerrors.Register(ModuleName, 9, "position already closed")
	ErrPositionNotLiquidatable = sdkerrors.Register(ModuleName, 10, "position not liquidatable")
	ErrInvalidPrice            = sdkerrors.Register(ModuleName, 11, "invalid price")
	ErrSlippageExceeded        = sdkerrors.Register(ModuleName, 12, "slippage exceeded")
	ErrMaxLeverageExceeded     = sdkerrors.Register(ModuleName, 13, "max leverage exceeded")
	ErrMinCollateralNotMet     = sdkerrors.Register(ModuleName, 14, "minimum collateral not met")
	ErrMaxPositionSizeExceeded = sdkerrors.Register(ModuleName, 15, "max position size exceeded")
	ErrInsufficientLiquidity   = sdkerrors.Register(ModuleName, 16, "insufficient liquidity")
	ErrTokenNotSupported       = sdkerrors.Register(ModuleName, 17, "token not supported")
	ErrInvalidCollateralDenom  = sdkerrors.Register(ModuleName, 18, "invalid collateral denomination")

	// Lending errors
	ErrBorrowPositionNotFound    = sdkerrors.Register(ModuleName, 19, "borrow position not found")
	ErrLendingPoolNotFound       = sdkerrors.Register(ModuleName, 20, "lending pool not found")
	ErrInvalidInterestRate       = sdkerrors.Register(ModuleName, 21, "invalid interest rate")
	ErrBorrowLimitExceeded       = sdkerrors.Register(ModuleName, 22, "borrow limit exceeded")
	ErrRepaymentExceedsDebt      = sdkerrors.Register(ModuleName, 23, "repayment amount exceeds debt")
	ErrCannotBorrowOwnToken      = sdkerrors.Register(ModuleName, 24, "cannot borrow own collateral token")
	ErrLiquidityProviderNotFound = sdkerrors.Register(ModuleName, 25, "liquidity provider not found")
	ErrInsufficientShares        = sdkerrors.Register(ModuleName, 26, "insufficient share tokens")
	ErrInvalidPositionID         = sdkerrors.Register(ModuleName, 27, "invalid position ID")
	ErrInvalidTraderAddress      = sdkerrors.Register(ModuleName, 28, "invalid trader address")
	ErrInvalidTokenDenom         = sdkerrors.Register(ModuleName, 29, "invalid token denom")
)

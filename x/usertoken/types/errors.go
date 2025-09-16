package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/usertoken module sentinel errors
var (
	ErrInvalidAddress     = sdkerrors.Register(ModuleName, 1, "invalid address")
	ErrUnauthorized       = sdkerrors.Register(ModuleName, 2, "unauthorized")
	ErrInvalidAmount      = sdkerrors.Register(ModuleName, 3, "invalid amount")
	ErrTokenNotFound      = sdkerrors.Register(ModuleName, 4, "token not found")
	ErrInvalidRequest     = sdkerrors.Register(ModuleName, 5, "invalid request")
	ErrInvalidSubdenom    = sdkerrors.Register(ModuleName, 6, "invalid subdenom")
	ErrInvalidName        = sdkerrors.Register(ModuleName, 7, "invalid name")
	ErrInvalidSymbol      = sdkerrors.Register(ModuleName, 8, "invalid symbol")
	ErrInvalidDecimals    = sdkerrors.Register(ModuleName, 9, "invalid decimals")
	ErrTokenAlreadyExists = sdkerrors.Register(ModuleName, 10, "token already exists")
	ErrInsufficientFunds  = sdkerrors.Register(ModuleName, 11, "insufficient funds")
	ErrLBPNotActive       = sdkerrors.Register(ModuleName, 12, "LBP not active")
	ErrLBPAlreadyActive   = sdkerrors.Register(ModuleName, 13, "LBP already active")
)

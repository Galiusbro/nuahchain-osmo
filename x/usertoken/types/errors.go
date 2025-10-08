package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrTokenExists         = errorsmod.Register(ModuleName, 1, "token already exists")
	ErrNameExists          = errorsmod.Register(ModuleName, 2, "token name already in use")
	ErrSymbolExists        = errorsmod.Register(ModuleName, 3, "token symbol already in use")
	ErrInvalidName         = errorsmod.Register(ModuleName, 4, "invalid token name")
	ErrInvalidSymbol       = errorsmod.Register(ModuleName, 5, "invalid token symbol")
	ErrInvalidImage        = errorsmod.Register(ModuleName, 6, "invalid token image URL")
	ErrInvalidDescription  = errorsmod.Register(ModuleName, 7, "invalid token description")
	ErrInvalidAddress      = errorsmod.Register(ModuleName, 8, "invalid address")
	ErrParamValidation     = errorsmod.Register(ModuleName, 9, "invalid module parameters")
	ErrFounderClaimExpired = errorsmod.Register(ModuleName, 10, "founder claim period expired")
	ErrFounderAlreadyClaim = errorsmod.Register(ModuleName, 11, "founder tranche already resolved")
	ErrTokenNotFound       = errorsmod.Register(ModuleName, 12, "token not found")
	ErrDistributionConfig  = errorsmod.Register(ModuleName, 13, "distribution configuration error")
	ErrTokenFactory        = errorsmod.Register(ModuleName, 14, "token factory error")
	ErrInsufficientFee     = errorsmod.Register(ModuleName, 15, "insufficient creation fee")
	ErrParamAddresses      = errorsmod.Register(ModuleName, 16, "module parameter addresses not configured")
)

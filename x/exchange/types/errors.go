package types

import (
	"cosmossdk.io/errors"
)

// x/exchange module sentinel errors
var (
	ErrInvalidSigner              = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrInvalidExchangeAmount      = errors.Register(ModuleName, 1101, "invalid exchange amount")
	ErrExchangeAmountTooSmall     = errors.Register(ModuleName, 1102, "exchange amount below minimum threshold")
	ErrExchangeAmountTooLarge     = errors.Register(ModuleName, 1103, "exchange amount above maximum threshold")
	ErrDailyLimitExceeded         = errors.Register(ModuleName, 1104, "daily exchange limit exceeded")
	ErrUnsupportedToken           = errors.Register(ModuleName, 1105, "token not supported for exchange")
	ErrExchangeRateNotFound       = errors.Register(ModuleName, 1106, "exchange rate not found for token")
	ErrPriceDeviationTooHigh      = errors.Register(ModuleName, 1107, "price deviation between Oracle and TWAP too high")
	ErrExchangeDisabled           = errors.Register(ModuleName, 1108, "exchange functionality is disabled")
	ErrInvalidMinOutput           = errors.Register(ModuleName, 1109, "minimum output amount not met")
	ErrInsufficientBalance        = errors.Register(ModuleName, 1110, "insufficient balance for exchange")
	ErrInvalidTreasuryAddress     = errors.Register(ModuleName, 1111, "invalid treasury address")
	ErrExchangeFeeCalculation     = errors.Register(ModuleName, 1112, "error calculating exchange fee")
	ErrOracleDataStale            = errors.Register(ModuleName, 1113, "oracle data is stale")
	ErrTWAPDataNotAvailable       = errors.Register(ModuleName, 1114, "TWAP data not available")
	ErrInvalidParams              = errors.Register(ModuleName, 1115, "invalid module parameters")
	ErrInvalidAuthority           = errors.Register(ModuleName, 1116, "invalid authority for governance message")
	ErrTokenAlreadySupported      = errors.Register(ModuleName, 1117, "token is already in supported tokens registry")
	ErrTokenNotSupported          = errors.Register(ModuleName, 1118, "token is not in supported tokens registry")
)
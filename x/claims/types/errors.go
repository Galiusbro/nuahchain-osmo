package types

import "cosmossdk.io/errors"

var (
	ErrClaimNotFound         = errors.Register(ModuleName, 3100, "claim not found")
	ErrUnauthorized          = errors.Register(ModuleName, 3101, "unauthorized")
	ErrInvalidPolicy         = errors.Register(ModuleName, 3102, "invalid policy")
	ErrInvalidStatusChange   = errors.Register(ModuleName, 3103, "invalid claim status change")
	ErrClaimAlreadyResolved  = errors.Register(ModuleName, 3104, "claim already resolved")
	ErrUnsupportedPayout     = errors.Register(ModuleName, 3105, "payout unavailable")
	ErrMaxOpenClaimsExceeded = errors.Register(ModuleName, 3106, "too many open claims for policy")
)

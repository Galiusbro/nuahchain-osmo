package types

import (
	"cosmossdk.io/errors"
)

// x/limitedaccount module sentinel errors
var (
	ErrInvalidParams        = errors.Register(ModuleName, 1, "invalid params")
	ErrDailyLimitExceeded   = errors.Register(ModuleName, 2, "daily transaction limit exceeded")
	ErrInvalidAddress       = errors.Register(ModuleName, 3, "invalid address")
	ErrDuplicateAddress     = errors.Register(ModuleName, 4, "duplicate address")
	ErrAccountAlreadyExists = errors.Register(ModuleName, 5, "account already exists")
	ErrAccountNotFound      = errors.Register(ModuleName, 6, "account not found")
)

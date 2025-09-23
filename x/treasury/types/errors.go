package types

import "cosmossdk.io/errors"

var (
	ErrPoolNotFound       = errors.Register(ModuleName, 4100, "treasury pool not found")
	ErrUnauthorized       = errors.Register(ModuleName, 4101, "unauthorized")
	ErrInvalidDeposit     = errors.Register(ModuleName, 4102, "invalid deposit")
	ErrInsufficientFunds  = errors.Register(ModuleName, 4103, "insufficient treasury funds")
	ErrInvalidPoolRequest = errors.Register(ModuleName, 4104, "invalid pool request")
)

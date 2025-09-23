package types

import (
	"fmt"

	"cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrUnauthorized     = errors.Register(ModuleName, 1100, "unauthorized")
	ErrRoleNotFound     = errors.Register(ModuleName, 1101, "role not found")
	ErrInvalidAuthority = errors.Register(ModuleName, 1102, "invalid authority")
)

// ErrUnknownRole formats an error for unsupported roles.
func ErrUnknownRole(role Role) error {
	return errors.Wrapf(sdkerrors.ErrInvalidType, "unknown role %d", role)
}

// ErrAddressMissingRole signals that an address lacks a required role.
func ErrAddressMissingRole(address string, role Role) error {
	return errors.Wrapf(ErrRoleNotFound, "address %s missing role %s", address, role.String())
}

// ValidateAuthority ensures at least one authority value is provided.
func ValidateAuthority(authority string) error {
	if authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	return nil
}

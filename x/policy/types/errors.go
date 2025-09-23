package types

import (
	"fmt"

	"cosmossdk.io/errors"
)

var (
	ErrPolicyNotFound      = errors.Register(ModuleName, 1200, "policy not found")
	ErrPolicyInactive      = errors.Register(ModuleName, 1201, "policy inactive")
	ErrUnauthorized        = errors.Register(ModuleName, 1202, "unauthorized")
	ErrInvalidPolicyType   = errors.Register(ModuleName, 1203, "invalid policy type")
	ErrPolicyAlreadyClosed = errors.Register(ModuleName, 1204, "policy already closed")
)

// ErrPolicyOwnerMismatch signals an owner mismatch for a policy.
func ErrPolicyOwnerMismatch(expected, actual string) error {
	return errors.Wrapf(ErrUnauthorized, "expected owner %s, got %s", expected, actual)
}

// ErrPolicyStatusTransition indicates an invalid status update.
func ErrPolicyStatusTransition(from PolicyStatus, to PolicyStatus) error {
	return errors.Wrapf(ErrUnauthorized, "cannot transition policy from %s to %s", from.String(), to.String())
}

// ErrPolicyMissingAttribute signals that a required attribute is missing.
func ErrPolicyMissingAttribute(key string) error {
	return fmt.Errorf("policy missing attribute %s", key)
}

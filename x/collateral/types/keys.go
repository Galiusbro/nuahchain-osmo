package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name.
	ModuleName = "collateral"
	// StoreKey defines the module store key.
	StoreKey = ModuleName
	// RouterKey defines the message routing key.
	RouterKey = ModuleName
	// MemStoreKey defines the in-memory store key.
	MemStoreKey = "mem_collateral"
)

var (
	// CollateralKeyPrefix stores per-owner collateral positions.
	CollateralKeyPrefix = []byte{0x01}
)

// NormalizeDenom trims and uppercases denom strings for storage consistency.
func NormalizeDenom(denom string) string {
	return strings.TrimSpace(strings.ToLower(denom))
}

// OwnerPrefix returns the prefix store key for the provided owner address.
func OwnerPrefix(owner sdk.AccAddress) []byte {
	return append(CollateralKeyPrefix, owner.Bytes()...)
}

// PositionKey returns the store key for a specific owner and denom combination.
func PositionKey(owner sdk.AccAddress, denom string) []byte {
	key := make([]byte, 0, len(CollateralKeyPrefix)+len(owner.Bytes())+len(denom)+1)
	key = append(key, OwnerPrefix(owner)...)
	key = append(key, 0x00)
	key = append(key, []byte(NormalizeDenom(denom))...)
	return key
}

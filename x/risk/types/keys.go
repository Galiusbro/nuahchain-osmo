package types

import "strings"

const (
	// ModuleName defines the module name.
	ModuleName = "risk"
	// StoreKey defines the primary store key.
	StoreKey = ModuleName
	// RouterKey defines the module routing key.
	RouterKey = ModuleName
	// MemStoreKey defines the in-memory store key.
	MemStoreKey = "mem_risk"
)

var (
	// RiskParamsKeyPrefix is the prefix for storing risk parameters by symbol.
	RiskParamsKeyPrefix = []byte{0x01}
)

// RiskParamsKey returns the storage key for the provided asset symbol.
func RiskParamsKey(symbol string) []byte {
	return append(RiskParamsKeyPrefix, []byte(NormalizeSymbol(symbol))...)
}

// NormalizeSymbol canonicalizes asset symbols for storage.
func NormalizeSymbol(symbol string) string {
	return strings.ToUpper(strings.TrimSpace(symbol))
}

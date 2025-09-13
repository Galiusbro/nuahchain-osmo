package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "pegkeeper"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_pegkeeper"

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// ParamsKey defines the key to store the params
	ParamsKey = "params"
)

// KVStore keys
var (
	// PegStateKey stores the current peg state
	PegStateKey = []byte{0x01}

	// StabilizationParamsKey stores stabilization parameters
	StabilizationParamsKey = []byte{0x02}

	// LastAdjustmentKey stores the last adjustment timestamp
	LastAdjustmentKey = []byte{0x03}

	// AdjustmentHistoryPrefix stores historical adjustments
	AdjustmentHistoryPrefix = []byte{0x04}
)

// GetAdjustmentHistoryKey returns the key for storing adjustment history
func GetAdjustmentHistoryKey(timestamp int64) []byte {
	return append(AdjustmentHistoryPrefix, sdk.Uint64ToBigEndian(uint64(timestamp))...)
}

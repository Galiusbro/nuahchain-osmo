package types

const (
	// ModuleName defines the module name
	ModuleName = "freeaccount"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore keys
var (
	// FreeAccountPrefix is the prefix for free account addresses
	FreeAccountPrefix = []byte{0x01}
)
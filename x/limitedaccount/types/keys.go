package types

const (
	// ModuleName defines the module name
	ModuleName = "limitedaccount"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_limitedaccount"
)

var (
	// LimitedAccountPrefix is the prefix for limited account keys
	LimitedAccountPrefix = []byte{0x01}
	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x02}
)

// LimitedAccountKey returns the store key to retrieve a LimitedAccount from the index fields
func LimitedAccountKey(address string) []byte {
	return append(LimitedAccountPrefix, []byte(address)...)
}

package types

const (
	// ModuleName defines the module name
	ModuleName = "usertoken"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_usertoken"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// Store key prefixes
var (
	// UserTokenKeyPrefix is the prefix to retrieve all UserToken
	UserTokenKeyPrefix = []byte{0x01}
	// FounderTrancheKeyPrefix is the prefix to retrieve all FounderTranche
	FounderTrancheKeyPrefix = []byte{0x02}
)

// UserTokenKey returns the store key to retrieve a UserToken from the index fields
func UserTokenKey(denom string) []byte {
	return append(UserTokenKeyPrefix, []byte(denom)...)
}

// FounderTrancheKey returns the store key to retrieve a FounderTranche from the index fields
func FounderTrancheKey(denom string) []byte {
	return append(FounderTrancheKeyPrefix, []byte(denom)...)
}

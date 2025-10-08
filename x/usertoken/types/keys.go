package types

const (
	// ModuleName defines the module name
	ModuleName = "usertoken"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for usertoken
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_usertoken"
)

// Key prefixes for the module KV-store.
var (
	TokenKeyPrefix          = []byte{0x01}
	TokenByNamePrefix       = []byte{0x02}
	TokenBySymbolPrefix     = []byte{0x03}
	FounderDeadlineQueueKey = []byte{0x04}
	TokenCreatorIndexPrefix = []byte{0x05}
)

// TokenKey returns the primary key for a token record.
func TokenKey(denom string) []byte {
	return append(TokenKeyPrefix, []byte(denom)...)
}

// NameIndexKey returns the key that indexes a token by its name.
func NameIndexKey(name string) []byte {
	return append(TokenByNamePrefix, []byte(name)...)
}

// SymbolIndexKey returns the key that indexes a token by its symbol.
func SymbolIndexKey(symbol string) []byte {
	return append(TokenBySymbolPrefix, []byte(symbol)...)
}

// CreatorIndexKey indexes a token by creator address.
func CreatorIndexKey(creator string, denom string) []byte {
	key := append([]byte{}, TokenCreatorIndexPrefix...)
	key = append(key, []byte(creator)...)
	key = append(key, 0x00)
	return append(key, []byte(denom)...)
}

// FounderDeadlineKey queues a token by its founder claim deadline.
func FounderDeadlineKey(deadline uint64, denom string) []byte {
	key := make([]byte, 8)
	for i := uint(0); i < 8; i++ {
		key[7-i] = byte(deadline >> (i * 8))
	}
	key = append(FounderDeadlineQueueKey, key...)
	key = append(key, 0x00)
	return append(key, []byte(denom)...)
}

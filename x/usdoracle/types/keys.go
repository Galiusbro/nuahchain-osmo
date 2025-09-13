package types

const (
	// ModuleName defines the module name
	ModuleName = "usdoracle"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_usdoracle"
)

var (
	// CurrentPriceKey is the key for storing the current USD price
	CurrentPriceKey = []byte{0x01}

	// PriceHistoryPrefix is the prefix for storing price history
	PriceHistoryPrefix = []byte{0x02}

	// PriceSourcesPrefix is the prefix for storing price sources
	PriceSourcesPrefix = []byte{0x03}

	// ParamsKey is the key for storing module parameters
	ParamsKey = []byte{0x04}
)

// GetPriceHistoryKey returns the key for a specific price history entry
func GetPriceHistoryKey(timestamp int64) []byte {
	key := make([]byte, len(PriceHistoryPrefix)+8)
	copy(key, PriceHistoryPrefix)
	// Convert timestamp to bytes (big endian for proper ordering)
	for i := 0; i < 8; i++ {
		key[len(PriceHistoryPrefix)+i] = byte(timestamp >> (8 * (7 - i)))
	}
	return key
}

// GetPriceSourceKey returns the key for a specific price source
func GetPriceSourceKey(name string) []byte {
	return append(PriceSourcesPrefix, []byte(name)...)
}

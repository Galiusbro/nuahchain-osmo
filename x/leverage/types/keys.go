package types

const (
	// ModuleName defines the module name.
	ModuleName = "leverage"
	// StoreKey defines the primary module store key.
	StoreKey = ModuleName
	// RouterKey defines the routing key for messages.
	RouterKey = ModuleName
	// MemStoreKey defines the in-memory store key.
	MemStoreKey = "mem_leverage"
	// NDollarDenom defines the denom used for NDOLLAR accounting.
	NDollarDenom = "NDOLLAR"
)

var (
	// PositionKeyPrefix stores positions by id.
	PositionKeyPrefix = []byte{0x01}
	// OwnerPositionsPrefix stores position ids per owner.
	OwnerPositionsPrefix = []byte{0x02}
	// NextPositionIDKey stores the next position identifier.
	NextPositionIDKey = []byte{0x03}
	// PositionQuotePrefix stores quote amounts per position.
	PositionQuotePrefix = []byte{0x04}
)

// PositionKey returns key for a given position id.
func PositionKey(id uint64) []byte {
	bz := make([]byte, 8)
	for i := uint(0); i < 8; i++ {
		bz[7-i] = byte(id >> (i * 8))
	}
	return append(PositionKeyPrefix, bz...)
}

// OwnerPositionsKey returns prefix key for owner addresses.
func OwnerPositionsKey(owner []byte) []byte {
	return append(OwnerPositionsPrefix, owner...)
}

// PositionQuoteKey returns key for quote tracking.
func PositionQuoteKey(id uint64) []byte {
	return append(PositionQuotePrefix, PositionKey(id)[len(PositionKeyPrefix):]...)
}

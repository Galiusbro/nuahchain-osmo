package types

const (
	ModuleName  = "claims"
	StoreKey    = ModuleName
	RouterKey   = ModuleName
	MemStoreKey = "mem_claims"
)

var (
	ClaimKeyPrefix = []byte{0x01}
	NextClaimIDKey = []byte{0x02}
)

// ClaimKey returns the store key for a claim id.
func ClaimKey(id uint64) []byte {
	return append(ClaimKeyPrefix, Uint64ToBytes(id)...)
}

// Uint64ToBytes encodes a uint64 in big endian.
func Uint64ToBytes(id uint64) []byte {
	return []byte{
		byte(id >> 56),
		byte(id >> 48),
		byte(id >> 40),
		byte(id >> 32),
		byte(id >> 24),
		byte(id >> 16),
		byte(id >> 8),
		byte(id),
	}
}

// BytesToUint64 decodes a big-endian uint64.
func BytesToUint64(b []byte) uint64 {
	if len(b) != 8 {
		return 0
	}

	return uint64(b[0])<<56 |
		uint64(b[1])<<48 |
		uint64(b[2])<<40 |
		uint64(b[3])<<32 |
		uint64(b[4])<<24 |
		uint64(b[5])<<16 |
		uint64(b[6])<<8 |
		uint64(b[7])
}

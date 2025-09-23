package types

const (
	ModuleName  = "policy"
	StoreKey    = ModuleName
	RouterKey   = ModuleName
	MemStoreKey = "mem_policy"
)

var (
	PolicyKeyPrefix = []byte{0x01}
	NextPolicyIDKey = []byte{0x02}
)

// PolicyKey returns the store key for a policy id.
func PolicyKey(id uint64) []byte {
	return append(PolicyKeyPrefix, Uint64ToBytes(id)...)
}

// Uint64ToBytes encodes a uint64 into big-endian bytes.
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

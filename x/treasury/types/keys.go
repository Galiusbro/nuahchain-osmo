package types

const (
	ModuleName  = "treasury"
	StoreKey    = ModuleName
	RouterKey   = ModuleName
	MemStoreKey = "mem_treasury"
)

var (
	PoolKeyPrefix    = []byte{0x01}
	BalanceKeyPrefix = []byte{0x02}
	ReserveKeyPrefix = []byte{0x03}
)

// TreasuryPoolKey returns the store key for a pool.
func TreasuryPoolKey(poolID string) []byte {
	return append(PoolKeyPrefix, []byte(poolID)...)
}

// PoolBalanceKey returns the store key for a pool balance entry.
func PoolBalanceKey(poolID, denom string) []byte {
	key := append([]byte{}, BalanceKeyPrefix...)
	key = append(key, []byte(poolID)...)
	key = append(key, byte(0))
	key = append(key, []byte(denom)...)
	return key
}

// PoolReserveKey returns the store key for a reserve entry.
func PoolReserveKey(poolID, denom string) []byte {
	key := append([]byte{}, ReserveKeyPrefix...)
	key = append(key, []byte(poolID)...)
	key = append(key, byte(0))
	key = append(key, []byte(denom)...)
	return key
}

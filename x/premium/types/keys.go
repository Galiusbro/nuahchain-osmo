package types

const (
	ModuleName  = "premium"
	StoreKey    = ModuleName
	RouterKey   = ModuleName
	MemStoreKey = "mem_premium"
)

var (
	PremiumPlanKeyPrefix    = []byte{0x01}
	PremiumPaymentKeyPrefix = []byte{0x02}
	OverdueKeyPrefix        = []byte{0x03}
	NextPlanIDKey           = []byte{0x10}
	NextPaymentIDKey        = []byte{0x11}
)

func PremiumPlanKey(id uint64) []byte {
	return append(PremiumPlanKeyPrefix, Uint64ToBytes(id)...)
}

func PremiumPaymentKey(id uint64) []byte {
	return append(PremiumPaymentKeyPrefix, Uint64ToBytes(id)...)
}

func OverdueKey(planID uint64) []byte {
	return append(OverdueKeyPrefix, Uint64ToBytes(planID)...)
}

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

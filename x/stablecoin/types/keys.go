package types

const (
	// ModuleName defines the module name.
	ModuleName = "stablecoin"
	// StoreKey defines the primary module store key.
	StoreKey = ModuleName
	// RouterKey defines the module routing key (unused).
	RouterKey = ModuleName
	// MemStoreKey defines the in-memory store key.
	MemStoreKey = "mem_stablecoin"
	// NDollarDenom defines the denom used for NDOLLAR accounting.
	NDollarDenom = "NDOLLAR"
)

var (
	// TotalMintedKey stores the total minted NDOLLAR amount.
	TotalMintedKey = []byte{0x01}
	// TotalBurnedKey stores the total burned NDOLLAR amount.
	TotalBurnedKey = []byte{0x02}
)

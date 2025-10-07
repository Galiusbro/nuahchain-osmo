package types

const (
	// ModuleName defines the module name
	ModuleName = "leverage"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_leverage"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// Store key prefixes
var (
	// PositionKeyPrefix is the prefix to retrieve all Position
	PositionKeyPrefix = []byte{0x01}
	// ParamsKeyPrefix is the prefix to retrieve module params
	ParamsKeyPrefix = []byte{0x02}
	// NextPositionIDKeyPrefix is the prefix for the next position ID
	NextPositionIDKeyPrefix = []byte{0x03}

	// Lending system prefixes
	// LendingPoolKeyPrefix is the prefix to retrieve all LendingPool
	LendingPoolKeyPrefix = []byte{0x06}
	// BorrowPositionKeyPrefix is the prefix to retrieve all BorrowPosition
	BorrowPositionKeyPrefix = []byte{0x07}
	// LiquidityProviderKeyPrefix is the prefix to retrieve all LiquidityProvider
	LiquidityProviderKeyPrefix = []byte{0x08}
	// NextBorrowIDKeyPrefix is the prefix for the next borrow position ID
	NextBorrowIDKeyPrefix = []byte{0x09}
	// LeverageBorrowIndexKeyPrefix maps leverage position IDs to borrow IDs
	LeverageBorrowIndexKeyPrefix = []byte{0x0B}
)

// PositionKey returns the store key to retrieve a Position from the index fields
func PositionKey(positionID string) []byte {
	return append(PositionKeyPrefix, []byte(positionID)...)
}

// TraderPositionPrefix returns the prefix for positions by trader
func TraderPositionPrefix(trader string) []byte {
	return append([]byte{0x04}, []byte(trader)...)
}

// TraderPositionKey returns the store key for a position by trader and position ID
func TraderPositionKey(trader, positionID string) []byte {
	prefix := TraderPositionPrefix(trader)
	return append(prefix, []byte(positionID)...)
}

// TokenPositionPrefix returns the prefix for positions by token denom
func TokenPositionPrefix(tokenDenom string) []byte {
	return append([]byte{0x05}, []byte(tokenDenom)...)
}

// TokenPositionKey returns the store key for a position by token denom and position ID
func TokenPositionKey(tokenDenom, positionID string) []byte {
	prefix := TokenPositionPrefix(tokenDenom)
	return append(prefix, []byte(positionID)...)
}

// Lending system key functions

// LendingPoolKey returns the store key to retrieve a LendingPool
func LendingPoolKey(denom string) []byte {
	return append(LendingPoolKeyPrefix, []byte(denom)...)
}

// BorrowPositionKey returns the store key to retrieve a BorrowPosition
func BorrowPositionKey(borrowID string) []byte {
	return append(BorrowPositionKeyPrefix, []byte(borrowID)...)
}

// BorrowerPositionPrefix returns the prefix for borrow positions by borrower
func BorrowerPositionPrefix(borrower string) []byte {
	return append([]byte{0x0A}, []byte(borrower)...)
}

// BorrowerPositionKey returns the store key for a borrow position by borrower and borrow ID
func BorrowerPositionKey(borrower, borrowID string) []byte {
	prefix := BorrowerPositionPrefix(borrower)
	return append(prefix, []byte(borrowID)...)
}

// LeverageBorrowIndexKey returns the store key mapping a leverage position ID to its borrow ID
func LeverageBorrowIndexKey(leveragePositionID string) []byte {
	return append(LeverageBorrowIndexKeyPrefix, []byte(leveragePositionID)...)
}

// LiquidityProviderKey returns the store key for a liquidity provider
func LiquidityProviderKey(provider, denom string) []byte {
	prefix := append(LiquidityProviderKeyPrefix, []byte(provider)...)
	return append(prefix, []byte(denom)...)
}

// LiquidityProviderPrefix returns the prefix for liquidity providers by provider address
func LiquidityProviderPrefix(provider string) []byte {
	return append(LiquidityProviderKeyPrefix, []byte(provider)...)
}

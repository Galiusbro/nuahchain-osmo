package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "exchange"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_exchange"
)

// KVStore keys
var (
	// ParamsKey is the key for module parameters
	ParamsKey = collections.NewPrefix(1)
	
	// ExchangeRateKeyPrefix is the prefix for exchange rates
	ExchangeRateKeyPrefix = collections.NewPrefix(2)
	
	// DailyLimitKeyPrefix is the prefix for daily limits
	DailyLimitKeyPrefix = collections.NewPrefix(3)
	
	// ExchangeTransactionKeyPrefix is the prefix for exchange transactions
	ExchangeTransactionKeyPrefix = collections.NewPrefix(4)
)

// GetExchangeRateKey returns the key for a specific denom's exchange rate
func GetExchangeRateKey(denom string) []byte {
	return append(ExchangeRateKeyPrefix, []byte(denom)...)
}

// GetDailyLimitKey returns the key for a specific address's daily limit
func GetDailyLimitKey(address, date string) []byte {
	key := append(DailyLimitKeyPrefix, []byte(address)...)
	return append(key, []byte(date)...)
}

// GetExchangeTransactionKey returns the key for a specific exchange transaction
func GetExchangeTransactionKey(txID string) []byte {
	return append(ExchangeTransactionKeyPrefix, []byte(txID)...)
}
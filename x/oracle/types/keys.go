package types

import (
	"encoding/binary"
)

const (
	// ModuleName defines the oracle module name.
	ModuleName = "oracle"
	// StoreKey defines the primary module store key.
	StoreKey = ModuleName
	// RouterKey defines the message routing key.
	RouterKey = ModuleName
	// MemStoreKey defines the in-memory store key.
	MemStoreKey = "mem_oracle"

	// TypeMsgSetPrice is the type string for MsgSetPrice.
	TypeMsgSetPrice = "set_price"

	// EventTypePriceUpdated is emitted when a price is updated.
	EventTypePriceUpdated = "oracle.price_updated"

	// AttributeKeySymbol records the asset symbol in events.
	AttributeKeySymbol = "symbol"
	// AttributeKeyValue records the price value in events.
	AttributeKeyValue = "value"
	// AttributeKeyAuthority records the authority address in events.
	AttributeKeyAuthority = "authority"
)

var (
	// PriceKeyPrefix stores prices keyed by symbol.
	PriceKeyPrefix = []byte{0x01}
	// PriceHistoryKeyPrefix stores price history keyed by symbol and timestamp.
	PriceHistoryKeyPrefix = []byte{0x02}
)

// PriceKey returns the KV-store key for a symbol.
func PriceKey(symbol string) []byte {
	return append(PriceKeyPrefix, []byte(symbol)...)
}

// PriceHistoryKey returns the KV-store key for price history by symbol and timestamp.
func PriceHistoryKey(symbol string, timestamp int64) []byte {
	// prefix | symbol | 0x00 | big-endian timestamp (8 bytes)
	key := append(PriceHistoryKeyPrefix, []byte(symbol)...)
	key = append(key, 0x00)
	var ts [8]byte
	binary.BigEndian.PutUint64(ts[:], uint64(timestamp))
	key = append(key, ts[:]...)
	return key
}

// PriceHistoryPrefix returns the prefix for all price history entries of a symbol.
func PriceHistoryPrefix(symbol string) []byte {
	// prefix | symbol | 0x00 — start of the time-sorted keyspace
	p := append(PriceHistoryKeyPrefix, []byte(symbol)...)
	p = append(p, 0x00)
	return p
}

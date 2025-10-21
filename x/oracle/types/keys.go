package types

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
)

// PriceKey returns the KV-store key for a symbol.
func PriceKey(symbol string) []byte {
	return append(PriceKeyPrefix, []byte(symbol)...)
}

package types

const (
	// ModuleName defines the module name.
	ModuleName = "assets"
	// StoreKey defines the primary module store key.
	StoreKey = ModuleName
	// RouterKey defines the message routing key.
	RouterKey = ModuleName
	// MemStoreKey defines the in-memory store key.
	MemStoreKey = "mem_assets"

	// TypeMsgEnsureAsset is the type string for MsgEnsureAsset.
	TypeMsgEnsureAsset = "ensure_asset"

	// EventTypeAssetCreated is emitted when a new asset is registered.
	EventTypeAssetCreated = "assets.asset_created"
	// AttributeKeySymbol stores the asset symbol in events.
	AttributeKeySymbol = "symbol"
	// AttributeKeyCreator stores the message creator in events.
	AttributeKeyCreator = "creator"
)

var (
	// AssetKeyPrefix is used to store assets by their symbol string key.
	AssetKeyPrefix = []byte{0x01}
)

// KeyPrefix converts a string into a byte slice prefix for KVStore iteration.
func KeyPrefix(p string) []byte {
	return []byte(p)
}

// AssetKey returns the key used to store an asset by its symbol.
func AssetKey(symbol string) []byte {
	return append(AssetKeyPrefix, []byte(symbol)...)
}

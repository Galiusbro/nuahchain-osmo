package types

import (
	"strings"

	sdkmath "cosmossdk.io/math"
)

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
	// TypeMsgBuyAsset is the type string for MsgBuyAsset.
	TypeMsgBuyAsset = "buy_asset"
	// TypeMsgSellAsset is the type string for MsgSellAsset.
	TypeMsgSellAsset = "sell_asset"

	// EventTypeAssetCreated is emitted when a new asset is registered.
	EventTypeAssetCreated = "assets.asset_created"
	// EventTypeAssetBought is emitted when an asset is purchased.
	EventTypeAssetBought = "assets.asset_bought"
	// EventTypeAssetSold is emitted when an asset is sold.
	EventTypeAssetSold = "assets.asset_sold"
	// AttributeKeySymbol stores the asset symbol in events.
	AttributeKeySymbol = "symbol"
	// AttributeKeyCreator stores the message creator in events.
	AttributeKeyCreator = "creator"
	// AttributeKeyBuyer stores the buyer address in events.
	AttributeKeyBuyer = "buyer"
	// AttributeKeySeller stores the seller address in events.
	AttributeKeySeller = "seller"
	// AttributeKeyAmountNDOLLAR stores the NDOLLAR amount used for purchase.
	AttributeKeyAmountNDOLLAR = "amount_ndollar"
	// AttributeKeyBaseAmount stores the purchased/sold asset amount.
	AttributeKeyBaseAmount = "base_amount"
	// AttributeKeyPayoutNDOLLAR stores the NDOLLAR payout on sale.
	AttributeKeyPayoutNDOLLAR = "payout_ndollar"

	// NDollarDenom defines the NDOLLAR denom used for trades.
	NDollarDenom = "NDOLLAR"

	// AssetDenomExponent defines the decimal exponent used for asset tokens (1e6).
	AssetDenomExponent = 6
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

// AssetDenom returns the bank denom for an asset symbol.
func AssetDenom(symbol string) string {
	return "asset/" + strings.ToUpper(symbol)
}

// AssetPrecisionFactor returns the integer scaling factor for asset amounts.
func AssetPrecisionFactor() sdkmath.Int {
	return sdkmath.NewIntWithDecimal(1, AssetDenomExponent)
}

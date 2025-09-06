package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "softpeg"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_softpeg"

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// Event types
	EventTypePriceUpdate     = "price_update"
	EventTypePriceAlert      = "price_alert"
	EventTypeCommunityUpdate = "community_update"

	// Attribute keys
	AttributeKeyPrice      = "price"
	AttributeKeyConfidence = "confidence"
	AttributeKeyTimestamp  = "timestamp"
	AttributeKeyDeviation  = "deviation"
	AttributeKeySeverity   = "severity"
	AttributeKeyAlertType  = "alert_type"
)

// Store key prefixes
var (
	// PriceDataPrefix is the prefix for price data storage
	PriceDataPrefix = []byte{0x01}

	// CommunityMetricsPrefix is the prefix for community metrics storage
	CommunityMetricsPrefix = []byte{0x02}

	// PegConfigPrefix is the prefix for peg configuration storage
	PegConfigPrefix = []byte{0x03}

	// AlertsKeyPrefix is the prefix for alerts storage
	AlertsKeyPrefix = []byte{0x04}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x05}
)

// GetPriceDataKey returns the store key for price data at a specific timestamp
func GetPriceDataKey(timestamp int64) []byte {
	return append(PriceDataPrefix, sdk.Uint64ToBigEndian(uint64(timestamp))...)
}

// PriceDataKey returns the store key for price data
func PriceDataKey(timestamp int64) []byte {
	return append(PriceDataPrefix, sdk.Uint64ToBigEndian(uint64(timestamp))...)
}

// GetCommunityMetricsKey returns the store key for community metrics at a specific timestamp
func GetCommunityMetricsKey(timestamp int64) []byte {
	return append(CommunityMetricsPrefix, sdk.Uint64ToBigEndian(uint64(timestamp))...)
}

// CommunityMetricsKey returns the store key for community metrics
func CommunityMetricsKey(timestamp int64) []byte {
	return append(CommunityMetricsPrefix, sdk.Uint64ToBigEndian(uint64(timestamp))...)
}

// GetAlertKey returns the store key for a specific alert
func GetAlertKey(alertID string) []byte {
	return append(AlertsKeyPrefix, []byte(alertID)...)
}

// PegConfigKey returns the store key for peg configuration
func PegConfigKey() []byte {
	return PegConfigPrefix
}

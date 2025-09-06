package types

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
)

// Store key prefixes
var (
	// PriceDataKeyPrefix is the prefix for price data storage
	PriceDataKeyPrefix = []byte{0x01}

	// CommunityMetricsKeyPrefix is the prefix for community metrics storage
	CommunityMetricsKeyPrefix = []byte{0x02}

	// PegConfigKeyPrefix is the prefix for peg configuration storage
	PegConfigKeyPrefix = []byte{0x03}

	// AlertsKeyPrefix is the prefix for alerts storage
	AlertsKeyPrefix = []byte{0x04}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x05}
)

// GetPriceDataKey returns the store key for price data at a specific timestamp
func GetPriceDataKey(timestamp int64) []byte {
	return append(PriceDataKeyPrefix, sdk.Uint64ToBigEndian(uint64(timestamp))...)
}

// GetCommunityMetricsKey returns the store key for community metrics at a specific timestamp
func GetCommunityMetricsKey(timestamp int64) []byte {
	return append(CommunityMetricsKeyPrefix, sdk.Uint64ToBigEndian(uint64(timestamp))...)
}

// GetAlertKey returns the store key for a specific alert
func GetAlertKey(alertID string) []byte {
	return append(AlertsKeyPrefix, []byte(alertID)...)
}
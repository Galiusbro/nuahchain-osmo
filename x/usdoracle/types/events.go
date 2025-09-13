package types

// Event types
const (
	EventTypeUSDPriceUpdate     = "usd_price_update"
	EventTypePriceSourcesUpdate = "price_sources_update"
)

// Event attributes
const (
	AttributeKeyPrice        = "price"
	AttributeKeySource       = "source"
	AttributeKeyTimestamp    = "timestamp"
	AttributeKeySourcesCount = "sources_count"
)
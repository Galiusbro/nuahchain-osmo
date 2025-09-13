package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:        DefaultParams(),
		CurrentPrice:  nil,
		PriceHistory:  []USDPrice{},
		PriceSources:  DefaultPriceSources(),
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// Validate params
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate current price if set
	if gs.CurrentPrice != nil {
		if err := validateUSDPrice(*gs.CurrentPrice); err != nil {
			return fmt.Errorf("invalid current price: %w", err)
		}
	}

	// Validate price history
	for i, price := range gs.PriceHistory {
		if err := validateUSDPrice(price); err != nil {
			return fmt.Errorf("invalid price history entry %d: %w", i, err)
		}
	}

	// Validate price sources
	sourceNames := make(map[string]bool)
	for i, source := range gs.PriceSources {
		if err := validatePriceSource(source); err != nil {
			return fmt.Errorf("invalid price source %d: %w", i, err)
		}
		
		// Check for duplicate source names
		if sourceNames[source.Name] {
			return fmt.Errorf("duplicate price source name: %s", source.Name)
		}
		sourceNames[source.Name] = true
	}

	return nil
}

// DefaultPriceSources returns default price source configurations
func DefaultPriceSources() []PriceSource {
	return []PriceSource{
		{
			Name:    "coingecko",
			Weight:  math.LegacyMustNewDecFromStr("0.6"),
			Enabled: true,
			Url:     "https://api.coingecko.com/api/v3/simple/price?ids=usd-coin&vs_currencies=usd",
		},
		{
			Name:    "binance",
			Weight:  math.LegacyMustNewDecFromStr("0.4"),
			Enabled: true,
			Url:     "https://api.binance.com/api/v3/ticker/price?symbol=USDCUSDT",
		},
	}
}

func validateUSDPrice(price USDPrice) error {
	if price.Price.IsNegative() {
		return fmt.Errorf("price cannot be negative")
	}

	if price.Price.IsZero() {
		return fmt.Errorf("price cannot be zero")
	}

	if price.Source == "" {
		return fmt.Errorf("price source cannot be empty")
	}

	if price.Timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}

	if price.Timestamp.After(time.Now().Add(time.Hour)) {
		return fmt.Errorf("timestamp cannot be in the future")
	}

	return nil
}

func validatePriceSource(source PriceSource) error {
	if source.Name == "" {
		return fmt.Errorf("source name cannot be empty")
	}

	if source.Weight.IsNegative() {
		return fmt.Errorf("source weight cannot be negative")
	}

	if source.Weight.IsZero() {
		return fmt.Errorf("source weight cannot be zero")
	}

	if source.Weight.GT(math.LegacyOneDec()) {
		return fmt.Errorf("source weight cannot exceed 1.0")
	}

	if source.Url == "" {
		return fmt.Errorf("source URL cannot be empty")
	}

	return nil
}
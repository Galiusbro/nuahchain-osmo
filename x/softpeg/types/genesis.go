package types

import "fmt"

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:               DefaultParams(),
		PriceDataList:        []PriceData{},
		CommunityMetricsList: []CommunityMetrics{},
		PegConfig:            nil,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Validate params
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate price data list
	priceDataTimestamps := make(map[int64]bool)
	for _, priceData := range gs.PriceDataList {
		if err := priceData.Validate(); err != nil {
			return fmt.Errorf("invalid price data: %w", err)
		}
		// Check for duplicate timestamps
		if priceDataTimestamps[priceData.Timestamp] {
			return fmt.Errorf("duplicate price data timestamp: %d", priceData.Timestamp)
		}
		priceDataTimestamps[priceData.Timestamp] = true
	}

	// Validate community metrics list
	metricsTimestamps := make(map[int64]bool)
	for _, metrics := range gs.CommunityMetricsList {
		if err := metrics.Validate(); err != nil {
			return fmt.Errorf("invalid community metrics: %w", err)
		}
		// Check for duplicate timestamps
		if metricsTimestamps[metrics.Timestamp] {
			return fmt.Errorf("duplicate community metrics timestamp: %d", metrics.Timestamp)
		}
		metricsTimestamps[metrics.Timestamp] = true
	}

	// Validate peg config if present
	if gs.PegConfig != nil {
		if err := gs.PegConfig.Validate(); err != nil {
			return fmt.Errorf("invalid peg config: %w", err)
		}
	}

	return nil
}

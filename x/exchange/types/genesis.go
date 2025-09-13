package types

import (
	"fmt"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:        DefaultParams(),
		ExchangeRates: []ExchangeRate{},
		DailyLimits:   []DailyLimit{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Validate params
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate exchange rates
	seenDenoms := make(map[string]bool)
	for i, rate := range gs.ExchangeRates {
		if err := validateExchangeRate(rate); err != nil {
			return fmt.Errorf("invalid exchange rate at index %d: %w", i, err)
		}

		if seenDenoms[rate.Denom] {
			return fmt.Errorf("duplicate exchange rate for denom %s", rate.Denom)
		}
		seenDenoms[rate.Denom] = true
	}

	// Validate daily limits
	seenLimits := make(map[string]bool)
	for i, limit := range gs.DailyLimits {
		if err := validateDailyLimit(limit); err != nil {
			return fmt.Errorf("invalid daily limit at index %d: %w", i, err)
		}

		limitKey := fmt.Sprintf("%s-%s", limit.Address, limit.Date)
		if seenLimits[limitKey] {
			return fmt.Errorf("duplicate daily limit for address %s on date %s", limit.Address, limit.Date)
		}
		seenLimits[limitKey] = true
	}

	return nil
}

// validateExchangeRate validates an exchange rate
func validateExchangeRate(rate ExchangeRate) error {
	if rate.Denom == "" {
		return fmt.Errorf("exchange rate denom cannot be empty")
	}

	if rate.Rate.IsNegative() || rate.Rate.IsZero() {
		return fmt.Errorf("exchange rate must be positive: %s", rate.Rate)
	}

	if rate.LastUpdated.IsZero() {
		return fmt.Errorf("exchange rate last updated time cannot be zero")
	}

	return nil
}

// validateDailyLimit validates a daily limit
func validateDailyLimit(limit DailyLimit) error {
	if limit.Address == "" {
		return fmt.Errorf("daily limit address cannot be empty")
	}

	if limit.Date == "" {
		return fmt.Errorf("daily limit date cannot be empty")
	}

	if limit.TotalExchangedUsd.IsNegative() {
		return fmt.Errorf("total exchanged USD must be non-negative: %s", limit.TotalExchangedUsd)
	}

	return nil
}
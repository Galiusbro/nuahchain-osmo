package types

import (
	"fmt"
	"strings"
)

// DefaultGenesis returns default genesis data.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Prices: []*Price{},
	}
}

// Validate validates the genesis state.
func (gs *GenesisState) Validate() error {
	if gs == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}

	seen := make(map[string]struct{})
	for idx, price := range gs.Prices {
		if price == nil {
			return fmt.Errorf("price at index %d is nil", idx)
		}
		symbol := strings.TrimSpace(price.Symbol)
		if symbol == "" {
			return fmt.Errorf("price at index %d has empty symbol", idx)
		}
		if strings.TrimSpace(price.Value) == "" {
			return fmt.Errorf("price at index %d has empty value", idx)
		}
		if _, ok := seen[symbol]; ok {
			return fmt.Errorf("duplicate price symbol %s", price.Symbol)
		}
		seen[symbol] = struct{}{}
	}

	return nil
}

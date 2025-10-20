package types

import (
	"fmt"
)

// NewGenesisState creates a new GenesisState instance.
func NewGenesisState(assets []*Asset) *GenesisState {
	return &GenesisState{
		Assets: assets,
	}
}

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Assets: []*Asset{},
	}
}

// Validate validates the genesis state data.
func (gs *GenesisState) Validate() error {
	if gs == nil {
		return fmt.Errorf("genesis state is nil")
	}

	seen := make(map[string]struct{})
	for idx, asset := range gs.Assets {
		if asset == nil {
			return fmt.Errorf("asset at index %d is nil", idx)
		}
		if asset.Symbol == "" {
			return fmt.Errorf("asset at index %d has empty symbol", idx)
		}
		if _, exists := seen[asset.Symbol]; exists {
			return fmt.Errorf("duplicate asset symbol %s", asset.Symbol)
		}
		seen[asset.Symbol] = struct{}{}
	}

	return nil
}

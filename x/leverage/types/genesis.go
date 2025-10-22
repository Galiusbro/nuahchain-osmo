package types

import "fmt"

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		NextPositionId: 1,
		Positions:      []*Position{},
	}
}

// Validate validates the genesis state.
func (gs *GenesisState) Validate() error {
	if gs == nil {
		return nil
	}
	for _, pos := range gs.Positions {
		if pos == nil {
			return fmt.Errorf("position cannot be nil")
		}
		if pos.Id == 0 {
			return fmt.Errorf("position id must be positive")
		}
		if pos.Owner == "" {
			return fmt.Errorf("owner cannot be empty")
		}
		if pos.Symbol == "" {
			return fmt.Errorf("symbol cannot be empty")
		}
	}
	return nil
}

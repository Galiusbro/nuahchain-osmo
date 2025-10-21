package types

import "fmt"

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	params := DefaultParams()
	return &GenesisState{Params: &params}
}

// Validate validates genesis state fields.
func (gs *GenesisState) Validate() error {
	if gs == nil {
		return nil
	}
	if gs.Params == nil {
		return fmt.Errorf("params cannot be nil")
	}
	return gs.Params.Validate()
}

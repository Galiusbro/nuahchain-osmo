package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	stats := NewStats(sdkmath.ZeroInt(), sdkmath.ZeroInt())
	return &GenesisState{Stats: &stats}
}

// Validate validates the genesis state contents.
func (gs *GenesisState) Validate() error {
	if gs == nil || gs.Stats == nil {
		return fmt.Errorf("stats cannot be nil")
	}

	if _, ok := sdkmath.NewIntFromString(gs.Stats.TotalMinted); !ok {
		return fmt.Errorf("invalid total minted")
	}
	if _, ok := sdkmath.NewIntFromString(gs.Stats.TotalBurned); !ok {
		return fmt.Errorf("invalid total burned")
	}

	return nil
}

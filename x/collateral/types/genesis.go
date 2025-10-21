package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState defines the initial state for the collateral module.
type GenesisState struct {
	Positions []*Position `json:"positions" yaml:"positions"`
}

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{Positions: []*Position{}}
}

// Validate validates the genesis state contents.
func (gs *GenesisState) Validate() error {
	if gs == nil {
		return nil
	}
	for _, position := range gs.Positions {
		if position == nil {
			return fmt.Errorf("position cannot be nil")
		}
		if _, err := sdk.AccAddressFromBech32(position.Owner); err != nil {
			return fmt.Errorf("invalid owner address: %w", err)
		}
		if NormalizeDenom(position.Denom) == "" {
			return fmt.Errorf("denom cannot be empty")
		}
		amount, ok := sdkmath.NewIntFromString(position.Amount)
		if !ok {
			return fmt.Errorf("invalid amount for %s:%s", position.Owner, position.Denom)
		}
		if amount.IsNegative() {
			return fmt.Errorf("amount cannot be negative")
		}
	}
	return nil
}

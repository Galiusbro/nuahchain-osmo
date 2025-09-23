package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:      DefaultParams(),
		NextClaimId: 1,
		Claims:      []Claim{},
	}
}

// Validate performs basic genesis state validation.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	seen := make(map[uint64]struct{})

	for _, claim := range gs.Claims {
		if claim.Id == 0 {
			return fmt.Errorf("claim id must be positive")
		}
		if _, exists := seen[claim.Id]; exists {
			return fmt.Errorf("duplicate claim id %d", claim.Id)
		}
		seen[claim.Id] = struct{}{}

		if claim.PolicyId == 0 {
			return fmt.Errorf("claim %d missing policy id", claim.Id)
		}
		if claim.Claimant == "" {
			return fmt.Errorf("claim %d missing claimant", claim.Id)
		}
		if _, err := sdk.AccAddressFromBech32(claim.Claimant); err != nil {
			return fmt.Errorf("claim %d invalid claimant: %w", claim.Id, err)
		}
		if claim.Amount.Amount.IsNegative() {
			return fmt.Errorf("claim %d amount negative", claim.Id)
		}
	}

	return nil
}

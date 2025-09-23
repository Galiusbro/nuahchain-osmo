package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:       DefaultParams(),
		NextPolicyId: 1,
		Policies:     []Policy{},
	}
}

// Validate performs basic genesis validation.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	seen := make(map[uint64]struct{})

	for _, policy := range gs.Policies {
		if policy.Id == 0 {
			return fmt.Errorf("policy id must be positive")
		}

		if _, exists := seen[policy.Id]; exists {
			return fmt.Errorf("duplicate policy id %d", policy.Id)
		}
		seen[policy.Id] = struct{}{}

		if policy.Owner == "" {
			return fmt.Errorf("policy owner missing for id %d", policy.Id)
		}

		if _, err := sdk.AccAddressFromBech32(policy.Owner); err != nil {
			return fmt.Errorf("invalid owner for policy %d: %w", policy.Id, err)
		}

		if policy.PolicyType == "" {
			return fmt.Errorf("policy type missing for id %d", policy.Id)
		}

		if policy.StartTime != nil && policy.EndTime != nil && policy.EndTime.Before(*policy.StartTime) {
			return fmt.Errorf("policy %d end time before start time", policy.Id)
		}

		if policy.EndTime != nil && policy.EndTime.Before(time.Unix(0, 0)) {
			return fmt.Errorf("policy %d end time invalid", policy.Id)
		}
	}

	return nil
}

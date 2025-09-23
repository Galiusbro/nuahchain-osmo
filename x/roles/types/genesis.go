package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:   DefaultParams(),
		Bindings: []RoleBinding{},
	}
}

// Validate performs basic genesis state validation.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	seen := make(map[string]struct{})

	for _, binding := range gs.Bindings {
		if binding.Address == "" {
			return fmt.Errorf("role binding address cannot be empty")
		}

		if _, err := sdk.AccAddressFromBech32(binding.Address); err != nil {
			return fmt.Errorf("invalid address %s: %w", binding.Address, err)
		}

		key := binding.Address
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate role binding for %s", key)
		}
		seen[key] = struct{}{}

		roleSet := make(map[Role]struct{})
		for _, role := range binding.Roles {
			if role == Role_ROLE_UNSPECIFIED {
				return fmt.Errorf("binding for %s includes unspecified role", binding.Address)
			}

			if _, exists := roleSet[role]; exists {
				return fmt.Errorf("binding for %s includes duplicate role %s", binding.Address, role.String())
			}
			roleSet[role] = struct{}{}
		}
	}

	return nil
}

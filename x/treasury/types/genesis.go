package types

import (
	"fmt"
)

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:   DefaultParams(),
		Pools:    []TreasuryPool{},
		Balances: []PoolBalance{},
		Reserves: []PoolReserves{},
	}
}

// Validate performs basic validation of the genesis state.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	pools := make(map[string]struct{})
	for _, pool := range gs.Pools {
		if pool.Id == "" {
			return fmt.Errorf("treasury pool id cannot be empty")
		}
		if _, exists := pools[pool.Id]; exists {
			return fmt.Errorf("duplicate treasury pool id %s", pool.Id)
		}
		pools[pool.Id] = struct{}{}
	}

	for _, balance := range gs.Balances {
		if balance.PoolId == "" {
			return fmt.Errorf("pool balance missing pool id")
		}
		if _, exists := pools[balance.PoolId]; !exists {
			return fmt.Errorf("balance references unknown pool %s", balance.PoolId)
		}
		if balance.Balance.Amount.IsNegative() {
			return fmt.Errorf("pool %s balance for %s is negative", balance.PoolId, balance.Balance.Denom)
		}
	}

	for _, reserve := range gs.Reserves {
		if reserve.PoolId == "" {
			return fmt.Errorf("reserve missing pool id")
		}
		if reserve.Denom == "" {
			return fmt.Errorf("reserve denomination missing for pool %s", reserve.PoolId)
		}
	}

	return nil
}

package types

// GenesisState defines the risk module genesis data.
type GenesisState struct {
	RiskParams []*RiskParams `json:"risk_params" yaml:"risk_params"`
}

// DefaultGenesis returns the default empty genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		RiskParams: []*RiskParams{},
	}
}

// Validate ensures the genesis state is valid.
func (gs *GenesisState) Validate() error {
	if gs == nil {
		return nil
	}

	for _, params := range gs.RiskParams {
		if err := ValidateRiskParams(params); err != nil {
			return err
		}
	}

	return nil
}

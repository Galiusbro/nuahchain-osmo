package types

// GenesisState defines the risk module genesis data.
type GenesisState struct {
	Params     *Params       `json:"params" yaml:"params"`
	RiskParams []*RiskParams `json:"risk_params" yaml:"risk_params"`
}

// DefaultGenesis returns the default empty genesis state.
func DefaultGenesis() *GenesisState {
	params := DefaultParams()
	return &GenesisState{
		Params:     &params,
		RiskParams: []*RiskParams{},
	}
}

// Validate ensures the genesis state is valid.
func (gs *GenesisState) Validate() error {
	if gs == nil {
		return nil
	}
	if gs.Params != nil {
		if err := gs.Params.Validate(); err != nil {
			return err
		}
	}

	for _, params := range gs.RiskParams {
		if err := ValidateRiskParams(params); err != nil {
			return err
		}
	}

	return nil
}

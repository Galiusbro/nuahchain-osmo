package types

import (
	"fmt"

	osmomath "github.com/osmosis-labs/osmosis/osmomath"
)

// ValidateRiskParams performs basic validation on a RiskParams object.
func ValidateRiskParams(params *RiskParams) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}

	if NormalizeSymbol(params.Symbol) == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	if _, err := osmomath.NewDecFromStr(params.MaxLeverage); err != nil {
		return fmt.Errorf("invalid max leverage: %w", err)
	}
	if _, err := osmomath.NewDecFromStr(params.MaintenanceMargin); err != nil {
		return fmt.Errorf("invalid maintenance margin: %w", err)
	}
	if _, err := osmomath.NewDecFromStr(params.InitialMargin); err != nil {
		return fmt.Errorf("invalid initial margin: %w", err)
	}

	return nil
}

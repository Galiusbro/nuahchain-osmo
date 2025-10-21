package types

import (
	"fmt"

	osmomath "github.com/osmosis-labs/osmosis/osmomath"
)

// DefaultTradeFeeRate defines the default trade fee (0.3%).
const DefaultTradeFeeRate = "0.003"

// NewParams returns a new Params instance.
func NewParams(tradeFeeRate string) Params {
	return Params{TradeFeeRate: tradeFeeRate}
}

// DefaultParams returns the default module parameters.
func DefaultParams() Params {
	return Params{TradeFeeRate: DefaultTradeFeeRate}
}

// Validate performs basic validation of parameters.
func (p Params) Validate() error {
	if _, err := osmomath.NewDecFromStr(p.TradeFeeRate); err != nil {
		return fmt.Errorf("invalid trade fee rate: %w", err)
	}

	dec := osmomath.MustNewDecFromStr(p.TradeFeeRate)
	if dec.IsNegative() {
		return fmt.Errorf("trade fee rate cannot be negative")
	}
	if dec.GTE(osmomath.OneDec()) {
		return fmt.Errorf("trade fee rate must be less than 1")
	}

	return nil
}

// TradeFeeRateDec returns the trade fee rate as a decimal.
func (p Params) TradeFeeRateDec() osmomath.Dec {
	dec, err := osmomath.NewDecFromStr(p.TradeFeeRate)
	if err != nil {
		return osmomath.ZeroDec()
	}
	return dec
}

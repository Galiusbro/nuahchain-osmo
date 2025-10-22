package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	osmomath "github.com/osmosis-labs/osmosis/osmomath"
)

// DefaultTradeFeeRate defines the default trade fee (0.3%).
const DefaultTradeFeeRate = "0.003"

var (
	KeyParamName = []byte("ParamName")
)

// ParamKeyTable returns the parameter key table for the module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyParamName, &p.ParamName, validateParamName),
	}
}

func validateParamName(i interface{}) error {
	value, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return validateTradeFee(value)
}

func validateTradeFee(value string) error {
	dec, err := osmomath.NewDecFromStr(value)
	if err != nil {
		return fmt.Errorf("invalid trade fee rate: %w", err)
	}
	if dec.IsNegative() {
		return fmt.Errorf("trade fee rate cannot be negative")
	}
	if dec.GTE(osmomath.OneDec()) {
		return fmt.Errorf("trade fee rate must be less than 1")
	}
	return nil
}

// NewParams returns a new Params instance.
func NewParams(tradeFeeRate string) Params {
	return Params{ParamName: tradeFeeRate}
}

// DefaultParams returns the default module parameters.
func DefaultParams() Params {
	return Params{ParamName: DefaultTradeFeeRate}
}

// Validate performs basic validation of parameters.
func (p Params) Validate() error {
	return validateTradeFee(p.ParamName)
}

// TradeFeeRateDec returns the trade fee rate as a decimal.
func (p Params) TradeFeeRateDec() osmomath.Dec {
	dec, err := osmomath.NewDecFromStr(p.ParamName)
	if err != nil {
		return osmomath.ZeroDec()
	}
	return dec
}

package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyAllowedDenoms             = []byte("AllowedDenoms")
	KeyPaymentGracePeriodSeconds = []byte("PaymentGracePeriodSeconds")
	KeyAllowPartialPayments      = []byte("AllowPartialPayments")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams constructs Params.
func NewParams(denoms []string, grace uint64, allowPartial bool) Params {
	return Params{
		AllowedDenoms:             denoms,
		PaymentGracePeriodSeconds: grace,
		AllowPartialPayments:      allowPartial,
	}
}

// DefaultParams returns default params.
func DefaultParams() Params {
	return NewParams([]string{"stake"}, 0, false)
}

// ParamSetPairs implements ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAllowedDenoms, &p.AllowedDenoms, validateAllowedDenoms),
		paramtypes.NewParamSetPair(KeyPaymentGracePeriodSeconds, &p.PaymentGracePeriodSeconds, validatePositiveUint64),
		paramtypes.NewParamSetPair(KeyAllowPartialPayments, &p.AllowPartialPayments, validateBool),
	}
}

// Validate validates params.
func (p Params) Validate() error {
	if err := validateAllowedDenoms(p.AllowedDenoms); err != nil {
		return err
	}
	if err := validatePositiveUint64(p.PaymentGracePeriodSeconds); err != nil {
		return err
	}
	return nil
}

func validateAllowedDenoms(i interface{}) error {
	denoms, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	for _, denom := range denoms {
		if denom == "" {
			return fmt.Errorf("denom cannot be empty")
		}
	}
	return nil
}

func validatePositiveUint64(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateBool(i interface{}) error {
	if _, ok := i.(bool); !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

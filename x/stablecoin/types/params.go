package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const DefaultParamName = "default"

var (
	KeyParamName = []byte("ParamName")
)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

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
	if value == "" {
		return fmt.Errorf("param name cannot be empty")
	}
	return nil
}

func NewParams(name string) Params {
	return Params{ParamName: name}
}

func DefaultParams() Params {
	return NewParams(DefaultParamName)
}

func (p Params) Validate() error {
	return validateParamName(p.ParamName)
}

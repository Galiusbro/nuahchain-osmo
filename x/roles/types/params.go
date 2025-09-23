package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyAuthority = []byte("Authority")
)

// ParamKeyTable returns the parameter key table for the roles module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance.
func NewParams(authority string) Params {
	return Params{Authority: authority}
}

// DefaultParams provides default module parameters.
func DefaultParams() Params {
	return NewParams("")
}

// ParamSetPairs implements the ParamSet interface required by the parameter store.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAuthority, &p.Authority, validateAuthorityParam),
	}
}

// Validate validates the parameter set.
func (p Params) Validate() error {
	return validateAuthorityParam(p.Authority)
}

func validateAuthorityParam(i interface{}) error {
	authority, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if authority == "" {
		return nil
	}

	// Validate bech32 address format.
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		return fmt.Errorf("invalid authority address: %w", err)
	}

	return nil
}

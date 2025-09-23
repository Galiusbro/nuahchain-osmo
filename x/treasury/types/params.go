package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyAuthority            = []byte("Authority")
	KeyDefaultPoolID        = []byte("DefaultPoolID")
	KeyEnableIBCWithdrawals = []byte("EnableIBCWithdrawals")
)

// ParamKeyTable returns the parameter key table for the treasury module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams constructs a Params instance.
func NewParams(authority, defaultPool string, ibcEnabled bool) Params {
	return Params{
		Authority:            authority,
		DefaultPoolId:        defaultPool,
		EnableIbcWithdrawals: ibcEnabled,
	}
}

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	return NewParams("", "", false)
}

// ParamSetPairs implements ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAuthority, &p.Authority, validateAuthority),
		paramtypes.NewParamSetPair(KeyDefaultPoolID, &p.DefaultPoolId, validateStringOptional),
		paramtypes.NewParamSetPair(KeyEnableIBCWithdrawals, &p.EnableIbcWithdrawals, validateBool),
	}
}

// Validate validates the parameter set.
func (p Params) Validate() error {
	if err := validateAuthority(p.Authority); err != nil {
		return err
	}
	if err := validateStringOptional(p.DefaultPoolId); err != nil {
		return err
	}
	if err := validateBool(p.EnableIbcWithdrawals); err != nil {
		return err
	}
	return nil
}

func validateAuthority(i interface{}) error {
	authority, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if authority == "" {
		return nil
	}
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		return fmt.Errorf("invalid authority address: %w", err)
	}
	return nil
}

func validateStringOptional(i interface{}) error {
	if _, ok := i.(string); !ok {
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

package types

import (
	"fmt"
	"strings"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyMaxOpenClaimsPerPolicy  = []byte("MaxOpenClaimsPerPolicy")
	KeyAutoApprovalPolicyTypes = []byte("AutoApprovalPolicyTypes")
)

// ParamKeyTable returns the parameter key table for the claims module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams constructs a Params instance.
func NewParams(maxOpen uint64, autoApproval []string) Params {
	return Params{
		MaxOpenClaimsPerPolicy:  maxOpen,
		AutoApprovalPolicyTypes: autoApproval,
	}
}

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	return NewParams(0, nil)
}

// ParamSetPairs implements the ParamSet interface.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxOpenClaimsPerPolicy, &p.MaxOpenClaimsPerPolicy, validateMaxOpenClaims),
		paramtypes.NewParamSetPair(KeyAutoApprovalPolicyTypes, &p.AutoApprovalPolicyTypes, validatePolicyTypeList),
	}
}

// Validate performs basic validation of params.
func (p Params) Validate() error {
	if err := validateMaxOpenClaims(p.MaxOpenClaimsPerPolicy); err != nil {
		return err
	}
	if err := validatePolicyTypeList(p.AutoApprovalPolicyTypes); err != nil {
		return err
	}
	return nil
}

func validateMaxOpenClaims(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validatePolicyTypeList(i interface{}) error {
	types, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, t := range types {
		if strings.TrimSpace(t) == "" {
			return fmt.Errorf("policy type entry cannot be empty")
		}
	}
	return nil
}

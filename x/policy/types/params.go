package types

import (
	"fmt"
	"strings"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyAllowedPolicyTypes    = []byte("AllowedPolicyTypes")
	KeyDefaultPolicyDuration = []byte("DefaultPolicyDuration")
	KeyDefaultTreasuryPoolID = []byte("DefaultTreasuryPoolID")
)

const (
	DefaultPolicyDurationDays = uint64(365)
)

// ParamKeyTable returns the parameter key table for the module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams constructs a new Params instance.
func NewParams(allowedTypes []string, duration uint64, poolID string) Params {
	return Params{
		AllowedPolicyTypes:        allowedTypes,
		DefaultPolicyDurationDays: duration,
		DefaultTreasuryPoolId:     poolID,
	}
}

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	return NewParams([]string{"custom"}, DefaultPolicyDurationDays, "")
}

// ParamSetPairs defines parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAllowedPolicyTypes, &p.AllowedPolicyTypes, validateAllowedPolicyTypes),
		paramtypes.NewParamSetPair(KeyDefaultPolicyDuration, &p.DefaultPolicyDurationDays, validateDefaultDuration),
		paramtypes.NewParamSetPair(KeyDefaultTreasuryPoolID, &p.DefaultTreasuryPoolId, validateDefaultTreasuryPoolID),
	}
}

// Validate validates params.
func (p Params) Validate() error {
	if err := validateAllowedPolicyTypes(p.AllowedPolicyTypes); err != nil {
		return err
	}
	if err := validateDefaultDuration(p.DefaultPolicyDurationDays); err != nil {
		return err
	}
	if err := validateDefaultTreasuryPoolID(p.DefaultTreasuryPoolId); err != nil {
		return err
	}
	return nil
}

func validateAllowedPolicyTypes(i interface{}) error {
	list, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, item := range list {
		if strings.TrimSpace(item) == "" {
			return fmt.Errorf("policy type cannot be empty")
		}
	}
	return nil
}

func validateDefaultDuration(i interface{}) error {
	val, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if val == 0 {
		return fmt.Errorf("default policy duration must be greater than zero")
	}
	return nil
}

func validateDefaultTreasuryPoolID(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

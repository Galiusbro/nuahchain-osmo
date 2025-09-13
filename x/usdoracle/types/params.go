package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	// Default parameter values
	DefaultEnabled                 = true
	DefaultUpdateInterval          = uint64(60)                           // 60 seconds
	DefaultPriceDeviationThreshold = math.LegacyMustNewDecFromStr("0.05") // 5%
	DefaultAdmin                   = ""
)

var (
	// Parameter keys
	KeyEnabled                 = []byte("Enabled")
	KeyAdmin                   = []byte("Admin")
	KeyUpdateInterval          = []byte("UpdateInterval")
	KeyPriceDeviationThreshold = []byte("PriceDeviationThreshold")
	KeySupportedTokens         = []byte("SupportedTokens")
	KeyPriceSources            = []byte("PriceSources")
	KeyMinSources              = []byte("MinSources")
	KeyMaxPriceAge             = []byte("MaxPriceAge")
	KeyAutoTokenRegistration   = []byte("AutoTokenRegistration")
	KeyDefaultUpdateFrequency  = []byte("DefaultUpdateFrequency")
	KeyDefaultMaxDeviation     = []byte("DefaultMaxDeviation")
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	enabled bool,
	admin string,
	updateInterval uint64,
	priceDeviationThreshold math.LegacyDec,
) Params {
	return Params{
		Enabled:                 enabled,
		Admin:                   admin,
		UpdateInterval:          updateInterval,
		PriceDeviationThreshold: priceDeviationThreshold,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultEnabled,
		DefaultAdmin,
		DefaultUpdateInterval,
		DefaultPriceDeviationThreshold,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyEnabled, &p.Enabled, validateEnabled),
		paramtypes.NewParamSetPair(KeyAdmin, &p.Admin, validateAdmin),
		paramtypes.NewParamSetPair(KeyUpdateInterval, &p.UpdateInterval, validateUpdateInterval),
		paramtypes.NewParamSetPair(KeyPriceDeviationThreshold, &p.PriceDeviationThreshold, validatePriceDeviationThreshold),
		paramtypes.NewParamSetPair(KeySupportedTokens, &p.SupportedTokens, validateSupportedTokens),
		paramtypes.NewParamSetPair(KeyPriceSources, &p.PriceSources, validatePriceSources),
		paramtypes.NewParamSetPair(KeyMinSources, &p.MinSources, validateMinSources),
		paramtypes.NewParamSetPair(KeyMaxPriceAge, &p.MaxPriceAge, validateMaxPriceAge),
		paramtypes.NewParamSetPair(KeyAutoTokenRegistration, &p.AutoTokenRegistration, validateAutoTokenRegistration),
		paramtypes.NewParamSetPair(KeyDefaultUpdateFrequency, &p.DefaultUpdateFrequency, validateDefaultUpdateFrequency),
		paramtypes.NewParamSetPair(KeyDefaultMaxDeviation, &p.DefaultMaxDeviation, validateDefaultMaxDeviation),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateEnabled(p.Enabled); err != nil {
		return err
	}

	if err := validateAdmin(p.Admin); err != nil {
		return err
	}

	if err := validateUpdateInterval(p.UpdateInterval); err != nil {
		return err
	}

	if err := validatePriceDeviationThreshold(p.PriceDeviationThreshold); err != nil {
		return err
	}

	if err := validateSupportedTokens(p.SupportedTokens); err != nil {
		return err
	}

	if err := validatePriceSources(p.PriceSources); err != nil {
		return err
	}

	if err := validateMinSources(p.MinSources); err != nil {
		return err
	}

	if err := validateMaxPriceAge(p.MaxPriceAge); err != nil {
		return err
	}

	if err := validateAutoTokenRegistration(p.AutoTokenRegistration); err != nil {
		return err
	}

	if err := validateDefaultUpdateFrequency(p.DefaultUpdateFrequency); err != nil {
		return err
	}

	if err := validateDefaultMaxDeviation(p.DefaultMaxDeviation); err != nil {
		return err
	}

	return nil
}

func validateEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateAdmin(i interface{}) error {
	admin, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Admin can be empty (will be set later)
	if admin != "" {
		// TODO: Add address validation if needed
	}

	return nil
}

func validateUpdateInterval(i interface{}) error {
	interval, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if interval == 0 {
		return fmt.Errorf("update interval must be positive")
	}

	if interval > uint64(24*time.Hour.Seconds()) {
		return fmt.Errorf("update interval cannot exceed 24 hours")
	}

	return nil
}

func validatePriceDeviationThreshold(i interface{}) error {
	threshold, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if threshold.IsNegative() {
		return fmt.Errorf("price deviation threshold cannot be negative")
	}

	if threshold.GT(math.LegacyOneDec()) {
		return fmt.Errorf("price deviation threshold cannot exceed 100%%")
	}

	return nil
}

func validateSupportedTokens(i interface{}) error {
	_, ok := i.([]SupportedToken)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validatePriceSources(i interface{}) error {
	_, ok := i.([]PriceSource)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateMinSources(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateMaxPriceAge(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateAutoTokenRegistration(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateDefaultUpdateFrequency(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateDefaultMaxDeviation(i interface{}) error {
	if i == nil {
		return nil
	}
	// Check if it's a pointer to LegacyDec
	if ptr, ok := i.(*math.LegacyDec); ok {
		if ptr == nil {
			return nil // nil pointer is valid
		}
		return nil
	}
	// Check if it's a direct LegacyDec value
	_, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

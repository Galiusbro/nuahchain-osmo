package types

import (
	"fmt"
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	// KeyMaxDeviationThreshold is the key for max deviation threshold
	KeyMaxDeviationThreshold = []byte("MaxDeviationThreshold")
	// KeyAdjustmentFactor is the key for adjustment factor
	KeyAdjustmentFactor = []byte("AdjustmentFactor")
	// KeyMinAdjustmentInterval is the key for min adjustment interval
	KeyMinAdjustmentInterval = []byte("MinAdjustmentInterval")
	// KeyMaxSupplyChangePerAdjustment is the key for max supply change per adjustment
	KeyMaxSupplyChangePerAdjustment = []byte("MaxSupplyChangePerAdjustment")
	// KeyOracleModule is the key for oracle module
	KeyOracleModule = []byte("OracleModule")
	// KeyEnabled is the key for enabled flag
	KeyEnabled = []byte("Enabled")
	// KeyTargetDenom is the key for target denomination
	KeyTargetDenom = []byte("TargetDenom")
	// KeyReferenceDenom is the key for reference denomination
	KeyReferenceDenom = []byte("ReferenceDenom")
	// KeyTargetPrice is the key for target price
	KeyTargetPrice = []byte("TargetPrice")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	maxDeviationThreshold string,
	adjustmentFactor string,
	minAdjustmentInterval time.Duration,
	maxSupplyChangePerAdjustment string,
	oracleModule string,
	enabled bool,
	targetDenom string,
	referenceDenom string,
	targetPrice string,
) Params {
	return Params{
		MaxDeviationThreshold:        maxDeviationThreshold,
		AdjustmentFactor:             adjustmentFactor,
		MinAdjustmentInterval:        int64(minAdjustmentInterval.Seconds()),
		MaxSupplyChangePerAdjustment: maxSupplyChangePerAdjustment,
		OracleModule:                 oracleModule,
		Enabled:                      enabled,
		TargetDenom:                  targetDenom,
		ReferenceDenom:               referenceDenom,
		TargetPrice:                  targetPrice,
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return NewParams(
		"0.05",      // 5% max deviation
		"0.1",       // 10% adjustment factor
		time.Hour,   // 1 hour min interval
		"0.02",      // 2% max supply change
		"usdoracle", // oracle module
		true,        // enabled
		"nuah",      // target denom
		"usd",       // reference denom
		"1.0",       // target price
	)
}

// Validate validates parameters
func (p Params) Validate() error {
	if err := validateMaxDeviationThreshold(p.MaxDeviationThreshold); err != nil {
		return err
	}
	if err := validateAdjustmentFactor(p.AdjustmentFactor); err != nil {
		return err
	}
	if err := validateMinAdjustmentInterval(p.MinAdjustmentInterval); err != nil {
		return err
	}
	if err := validateMaxSupplyChangePerAdjustment(p.MaxSupplyChangePerAdjustment); err != nil {
		return err
	}
	if err := validateOracleModule(p.OracleModule); err != nil {
		return err
	}
	if err := validateTargetDenom(p.TargetDenom); err != nil {
		return err
	}
	if err := validateReferenceDenom(p.ReferenceDenom); err != nil {
		return err
	}
	if err := validateTargetPrice(p.TargetPrice); err != nil {
		return err
	}
	return nil
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxDeviationThreshold, &p.MaxDeviationThreshold, validateMaxDeviationThreshold),
		paramtypes.NewParamSetPair(KeyAdjustmentFactor, &p.AdjustmentFactor, validateAdjustmentFactor),
		paramtypes.NewParamSetPair(KeyMinAdjustmentInterval, &p.MinAdjustmentInterval, validateMinAdjustmentInterval),
		paramtypes.NewParamSetPair(KeyMaxSupplyChangePerAdjustment, &p.MaxSupplyChangePerAdjustment, validateMaxSupplyChangePerAdjustment),
		paramtypes.NewParamSetPair(KeyOracleModule, &p.OracleModule, validateOracleModule),
		paramtypes.NewParamSetPair(KeyEnabled, &p.Enabled, validateEnabled),
		paramtypes.NewParamSetPair(KeyTargetDenom, &p.TargetDenom, validateTargetDenom),
		paramtypes.NewParamSetPair(KeyReferenceDenom, &p.ReferenceDenom, validateReferenceDenom),
		paramtypes.NewParamSetPair(KeyTargetPrice, &p.TargetPrice, validateTargetPrice),
	}
}

func validateMaxDeviationThreshold(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if len(v) == 0 {
		return fmt.Errorf("max deviation threshold cannot be empty")
	}
	return nil
}

func validateAdjustmentFactor(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if len(v) == 0 {
		return fmt.Errorf("adjustment factor cannot be empty")
	}
	return nil
}

func validateMinAdjustmentInterval(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return fmt.Errorf("min adjustment interval must be positive")
	}
	return nil
}

func validateMaxSupplyChangePerAdjustment(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if len(v) == 0 {
		return fmt.Errorf("max supply change per adjustment cannot be empty")
	}
	return nil
}

func validateOracleModule(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if len(v) == 0 {
		return fmt.Errorf("oracle module cannot be empty")
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

func validateTargetDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if len(v) == 0 {
		return fmt.Errorf("target denom cannot be empty")
	}
	return nil
}

func validateReferenceDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if len(v) == 0 {
		return fmt.Errorf("reference denom cannot be empty")
	}
	return nil
}

func validateTargetPrice(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if len(v) == 0 {
		return fmt.Errorf("target price cannot be empty")
	}
	return nil
}

package types

import (
	"fmt"
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	yaml "gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

// Parameter store keys
var (
	KeyTargetPrice        = []byte("TargetPrice")
	KeyDeviationThreshold = []byte("DeviationThreshold")
	KeyAlertThreshold     = []byte("AlertThreshold")
	KeyMonitoringEnabled  = []byte("MonitoringEnabled")
	KeyUpdateInterval     = []byte("UpdateInterval")
	KeyMinLiquidity       = []byte("MinLiquidity")
	KeyMaxDeviation       = []byte("MaxDeviation")
	KeyTrustDecayRate     = []byte("TrustDecayRate")
	KeyMinConfidence      = []byte("MinConfidence")
	KeyMaxAlertsPerHour   = []byte("MaxAlertsPerHour")
	KeyFeedbackCooldown   = []byte("FeedbackCooldown")
	KeyPriceDataRetention = []byte("PriceDataRetention")
	KeyCommunityWeight    = []byte("CommunityWeight")
	KeyLiquidityWeight    = []byte("LiquidityWeight")
	KeyVolumeWeight       = []byte("VolumeWeight")
)

// Default parameter values
var (
	DefaultTargetPrice        = osmomath.OneDec()                  // 1.0 USD
	DefaultDeviationThreshold = osmomath.MustNewDecFromStr("0.05") // 5%
	DefaultAlertThreshold     = osmomath.MustNewDecFromStr("0.03") // 3%
	DefaultMonitoringEnabled  = true
	DefaultUpdateInterval     = time.Minute * 5                     // 5 minutes
	DefaultMinLiquidity       = osmomath.NewInt(100000)             // 100k NUAH
	DefaultMaxDeviation       = osmomath.MustNewDecFromStr("0.10")  // 10%
	DefaultTrustDecayRate     = osmomath.MustNewDecFromStr("0.001") // 0.1% per hour
	DefaultMinConfidence      = osmomath.MustNewDecFromStr("0.70")  // 70%
	DefaultMaxAlertsPerHour   = int64(10)
	DefaultFeedbackCooldown   = time.Hour * 1                      // 1 hour
	DefaultPriceDataRetention = time.Hour * 24 * 30                // 30 days
	DefaultCommunityWeight    = osmomath.MustNewDecFromStr("0.30") // 30%
	DefaultLiquidityWeight    = osmomath.MustNewDecFromStr("0.40") // 40%
	DefaultVolumeWeight       = osmomath.MustNewDecFromStr("0.30") // 30%
)

// NewParams creates a new Params instance
func NewParams(
	targetPrice, deviationThreshold, alertThreshold osmomath.Dec,
	monitoringEnabled bool,
	updateInterval time.Duration,
	minLiquidity osmomath.Int,
	maxDeviation, trustDecayRate, minConfidence osmomath.Dec,
	maxAlertsPerHour int64,
	feedbackCooldown, priceDataRetention time.Duration,
	communityWeight, liquidityWeight, volumeWeight osmomath.Dec,
) Params {
	return Params{
		TargetPrice:        targetPrice,
		DeviationThreshold: deviationThreshold,
		AlertThreshold:     alertThreshold,
		MonitoringEnabled:  monitoringEnabled,
		UpdateInterval:     updateInterval,
		MinLiquidity:       minLiquidity,
		MaxDeviation:       maxDeviation,
		TrustDecayRate:     trustDecayRate,
		MinConfidence:      minConfidence,
		MaxAlertsPerHour:   maxAlertsPerHour,
		FeedbackCooldown:   feedbackCooldown,
		PriceDataRetention: priceDataRetention,
		CommunityWeight:    communityWeight,
		LiquidityWeight:    liquidityWeight,
		VolumeWeight:       volumeWeight,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultTargetPrice,
		DefaultDeviationThreshold,
		DefaultAlertThreshold,
		DefaultMonitoringEnabled,
		DefaultUpdateInterval,
		DefaultMinLiquidity,
		DefaultMaxDeviation,
		DefaultTrustDecayRate,
		DefaultMinConfidence,
		DefaultMaxAlertsPerHour,
		DefaultFeedbackCooldown,
		DefaultPriceDataRetention,
		DefaultCommunityWeight,
		DefaultLiquidityWeight,
		DefaultVolumeWeight,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyTargetPrice, &p.TargetPrice, validateTargetPrice),
		paramtypes.NewParamSetPair(KeyDeviationThreshold, &p.DeviationThreshold, validateDeviationThreshold),
		paramtypes.NewParamSetPair(KeyAlertThreshold, &p.AlertThreshold, validateAlertThreshold),
		paramtypes.NewParamSetPair(KeyMonitoringEnabled, &p.MonitoringEnabled, validateMonitoringEnabled),
		paramtypes.NewParamSetPair(KeyUpdateInterval, &p.UpdateInterval, validateUpdateInterval),
		paramtypes.NewParamSetPair(KeyMinLiquidity, &p.MinLiquidity, validateMinLiquidity),
		paramtypes.NewParamSetPair(KeyMaxDeviation, &p.MaxDeviation, validateMaxDeviation),
		paramtypes.NewParamSetPair(KeyTrustDecayRate, &p.TrustDecayRate, validateTrustDecayRate),
		paramtypes.NewParamSetPair(KeyMinConfidence, &p.MinConfidence, validateMinConfidence),
		paramtypes.NewParamSetPair(KeyMaxAlertsPerHour, &p.MaxAlertsPerHour, validateMaxAlertsPerHour),
		paramtypes.NewParamSetPair(KeyFeedbackCooldown, &p.FeedbackCooldown, validateFeedbackCooldown),
		paramtypes.NewParamSetPair(KeyPriceDataRetention, &p.PriceDataRetention, validatePriceDataRetention),
		paramtypes.NewParamSetPair(KeyCommunityWeight, &p.CommunityWeight, validateCommunityWeight),
		paramtypes.NewParamSetPair(KeyLiquidityWeight, &p.LiquidityWeight, validateLiquidityWeight),
		paramtypes.NewParamSetPair(KeyVolumeWeight, &p.VolumeWeight, validateVolumeWeight),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateTargetPrice(p.TargetPrice); err != nil {
		return err
	}
	if err := validateDeviationThreshold(p.DeviationThreshold); err != nil {
		return err
	}
	if err := validateAlertThreshold(p.AlertThreshold); err != nil {
		return err
	}
	if err := validateUpdateInterval(p.UpdateInterval); err != nil {
		return err
	}
	if err := validateMinLiquidity(p.MinLiquidity); err != nil {
		return err
	}
	if err := validateMaxDeviation(p.MaxDeviation); err != nil {
		return err
	}
	if err := validateTrustDecayRate(p.TrustDecayRate); err != nil {
		return err
	}
	if err := validateMinConfidence(p.MinConfidence); err != nil {
		return err
	}
	if err := validateMaxAlertsPerHour(p.MaxAlertsPerHour); err != nil {
		return err
	}
	if err := validateFeedbackCooldown(p.FeedbackCooldown); err != nil {
		return err
	}
	if err := validatePriceDataRetention(p.PriceDataRetention); err != nil {
		return err
	}

	// Validate weights sum to 1.0
	totalWeight := p.CommunityWeight.Add(p.LiquidityWeight).Add(p.VolumeWeight)
	if !totalWeight.Equal(osmomath.OneDec()) {
		return fmt.Errorf("weights must sum to 1.0, got %s", totalWeight.String())
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validation functions
func validateTargetPrice(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() || v.IsZero() {
		return fmt.Errorf("target price must be positive: %s", v)
	}

	return nil
}

func validateDeviationThreshold(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() || v.GT(osmomath.OneDec()) {
		return fmt.Errorf("deviation threshold must be between 0 and 1: %s", v)
	}

	return nil
}

func validateAlertThreshold(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() || v.GT(osmomath.OneDec()) {
		return fmt.Errorf("alert threshold must be between 0 and 1: %s", v)
	}

	return nil
}

func validateMonitoringEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateUpdateInterval(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("update interval must be positive: %s", v)
	}

	return nil
}

func validateMinLiquidity(i interface{}) error {
	v, ok := i.(osmomath.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("min liquidity must be non-negative: %s", v)
	}

	return nil
}

func validateMaxDeviation(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() || v.GT(osmomath.OneDec()) {
		return fmt.Errorf("max deviation must be between 0 and 1: %s", v)
	}

	return nil
}

func validateTrustDecayRate(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() || v.GT(osmomath.OneDec()) {
		return fmt.Errorf("trust decay rate must be between 0 and 1: %s", v)
	}

	return nil
}

func validateMinConfidence(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() || v.GT(osmomath.OneDec()) {
		return fmt.Errorf("min confidence must be between 0 and 1: %s", v)
	}

	return nil
}

func validateMaxAlertsPerHour(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 0 {
		return fmt.Errorf("max alerts per hour must be non-negative: %d", v)
	}

	return nil
}

func validateFeedbackCooldown(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 0 {
		return fmt.Errorf("feedback cooldown must be non-negative: %s", v)
	}

	return nil
}

func validatePriceDataRetention(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("price data retention must be positive: %s", v)
	}

	return nil
}

func validateCommunityWeight(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() || v.GT(osmomath.OneDec()) {
		return fmt.Errorf("community weight must be between 0 and 1: %s", v)
	}

	return nil
}

func validateLiquidityWeight(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() || v.GT(osmomath.OneDec()) {
		return fmt.Errorf("liquidity weight must be between 0 and 1: %s", v)
	}

	return nil
}

func validateVolumeWeight(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() || v.GT(osmomath.OneDec()) {
		return fmt.Errorf("volume weight must be between 0 and 1: %s", v)
	}

	return nil
}

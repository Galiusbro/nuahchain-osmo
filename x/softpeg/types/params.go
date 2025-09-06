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
	KeyTargetPrice           = []byte("TargetPrice")
	KeyDeviationThreshold    = []byte("DeviationThreshold")
	KeyAlertThreshold        = []byte("AlertThreshold")
	KeyMonitoringEnabled     = []byte("MonitoringEnabled")
	KeyUpdateInterval        = []byte("UpdateInterval")
	KeyMinLiquidity          = []byte("MinLiquidity")
	KeyMaxDeviation          = []byte("MaxDeviation")
	KeyTrustDecayRate        = []byte("TrustDecayRate")
	KeyMinConfidence         = []byte("MinConfidence")
	KeyMaxAlertsPerHour      = []byte("MaxAlertsPerHour")
	KeyFeedbackCooldown      = []byte("FeedbackCooldown")
	KeyPriceDataRetention    = []byte("PriceDataRetention")
	KeyCommunityWeight       = []byte("CommunityWeight")
	KeyLiquidityWeight       = []byte("LiquidityWeight")
	KeyVolumeWeight          = []byte("VolumeWeight")
)

// Default parameter values
var (
	DefaultTargetPrice        = osmomath.OneDec()                    // 1.0 USD
	DefaultDeviationThreshold = osmomath.MustNewDecFromStr("0.05")   // 5%
	DefaultAlertThreshold     = osmomath.MustNewDecFromStr("0.03")   // 3%
	DefaultMonitoringEnabled  = true
	DefaultUpdateInterval     = time.Minute * 5                      // 5 minutes
	DefaultMinLiquidity       = osmomath.NewInt(100000)              // 100k NUAH
	DefaultMaxDeviation       = osmomath.MustNewDecFromStr("0.10")   // 10%
	DefaultTrustDecayRate     = osmomath.MustNewDecFromStr("0.001")  // 0.1% per hour
	DefaultMinConfidence      = osmomath.MustNewDecFromStr("0.70")   // 70%
	DefaultMaxAlertsPerHour   = int64(10)
	DefaultFeedbackCooldown   = time.Hour * 1                       // 1 hour
	DefaultPriceDataRetention = time.Hour * 24 * 30                  // 30 days
	DefaultCommunityWeight    = osmomath.MustNewDecFromStr("0.30")   // 30%
	DefaultLiquidityWeight    = osmomath.MustNewDecFromStr("0.40")   // 40%
	DefaultVolumeWeight       = osmomath.MustNewDecFromStr("0.30")   // 30%
)

// Params defines the parameters for the softpeg module.
type Params struct {
	TargetPrice        osmomath.Dec  `protobuf:"bytes,1,opt,name=target_price,json=targetPrice,proto3,customtype=github.com/osmosis-labs/osmosis/osmomath.Dec" json:"target_price" yaml:"target_price"`
	DeviationThreshold osmomath.Dec  `protobuf:"bytes,2,opt,name=deviation_threshold,json=deviationThreshold,proto3,customtype=github.com/osmosis-labs/osmosis/osmomath.Dec" json:"deviation_threshold" yaml:"deviation_threshold"`
	AlertThreshold     osmomath.Dec  `protobuf:"bytes,3,opt,name=alert_threshold,json=alertThreshold,proto3,customtype=github.com/osmosis-labs/osmosis/osmomath.Dec" json:"alert_threshold" yaml:"alert_threshold"`
	MonitoringEnabled  bool          `protobuf:"varint,4,opt,name=monitoring_enabled,json=monitoringEnabled,proto3" json:"monitoring_enabled" yaml:"monitoring_enabled"`
	UpdateInterval     time.Duration `protobuf:"bytes,5,opt,name=update_interval,json=updateInterval,proto3,stdduration" json:"update_interval" yaml:"update_interval"`
	MinLiquidity       osmomath.Int  `protobuf:"bytes,6,opt,name=min_liquidity,json=minLiquidity,proto3,customtype=github.com/osmosis-labs/osmosis/osmomath.Int" json:"min_liquidity" yaml:"min_liquidity"`
	MaxDeviation       osmomath.Dec  `protobuf:"bytes,7,opt,name=max_deviation,json=maxDeviation,proto3,customtype=github.com/osmosis-labs/osmosis/osmomath.Dec" json:"max_deviation" yaml:"max_deviation"`
	TrustDecayRate     osmomath.Dec  `protobuf:"bytes,8,opt,name=trust_decay_rate,json=trustDecayRate,proto3,customtype=github.com/osmosis-labs/osmosis/osmomath.Dec" json:"trust_decay_rate" yaml:"trust_decay_rate"`
	MinConfidence      osmomath.Dec  `protobuf:"bytes,9,opt,name=min_confidence,json=minConfidence,proto3,customtype=github.com/osmosis-labs/osmosis/osmomath.Dec" json:"min_confidence" yaml:"min_confidence"`
	MaxAlertsPerHour   int64         `protobuf:"varint,10,opt,name=max_alerts_per_hour,json=maxAlertsPerHour,proto3" json:"max_alerts_per_hour" yaml:"max_alerts_per_hour"`
	FeedbackCooldown   time.Duration `protobuf:"bytes,11,opt,name=feedback_cooldown,json=feedbackCooldown,proto3,stdduration" json:"feedback_cooldown" yaml:"feedback_cooldown"`
	PriceDataRetention time.Duration `protobuf:"bytes,12,opt,name=price_data_retention,json=priceDataRetention,proto3,stdduration" json:"price_data_retention" yaml:"price_data_retention"`
	CommunityWeight    osmomath.Dec  `protobuf:"bytes,13,opt,name=community_weight,json=communityWeight,proto3,customtype=github.com/osmosis-labs/osmosis/osmomath.Dec" json:"community_weight" yaml:"community_weight"`
	LiquidityWeight    osmomath.Dec  `protobuf:"bytes,14,opt,name=liquidity_weight,json=liquidityWeight,proto3,customtype=github.com/osmosis-labs/osmosis/osmomath.Dec" json:"liquidity_weight" yaml:"liquidity_weight"`
	VolumeWeight       osmomath.Dec  `protobuf:"bytes,15,opt,name=volume_weight,json=volumeWeight,proto3,customtype=github.com/osmosis-labs/osmosis/osmomath.Dec" json:"volume_weight" yaml:"volume_weight"`
}

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
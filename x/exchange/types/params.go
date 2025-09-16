package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyEnabled                 = []byte("Enabled")
	KeyAdmin                   = []byte("Admin")
	KeyMinExchangeAmountUSD    = []byte("MinExchangeAmountUSD")
	KeyMaxExchangeAmountUSD    = []byte("MaxExchangeAmountUSD")
	KeyDailyLimitUSD           = []byte("DailyLimitUSD")
	KeyPriceDeviationThreshold = []byte("PriceDeviationThreshold")
	KeyTreasuryAddresses       = []byte("TreasuryAddresses")
	KeyExchangeFee             = []byte("ExchangeFee")
	KeySupportedTokens         = []byte("SupportedTokens")
)

// Default parameter values
var (
	DefaultEnabled                 = true
	DefaultAdmin                   = ""
	DefaultMinExchangeAmountUSD    = math.LegacyNewDec(10)           // $10
	DefaultMaxExchangeAmountUSD    = math.LegacyNewDec(100000)       // $100k
	DefaultDailyLimitUSD           = math.LegacyNewDec(1000000)      // $1M
	DefaultPriceDeviationThreshold = math.LegacyNewDecWithPrec(2, 2) // 2%
	DefaultTreasuryAddresses       = []string{}
	DefaultExchangeFee             = math.LegacyNewDecWithPrec(1, 3) // 0.1%
	DefaultSupportedTokens         = []string{"ueth", "ubtc", "uusdc", "uusdt", "uatom", "uosmo", "usol"}
)

// NewParams creates a new Params instance
func NewParams(
	enabled bool,
	admin string,
	minExchangeAmountUSD,
	maxExchangeAmountUSD,
	dailyLimitUSD,
	priceDeviationThreshold,
	exchangeFee math.LegacyDec,
	treasuryAddresses []string,
	supportedTokens []string,
) Params {
	return Params{
		Enabled:                 enabled,
		Admin:                   admin,
		MinExchangeAmountUsd:    minExchangeAmountUSD,
		MaxExchangeAmountUsd:    maxExchangeAmountUSD,
		DailyLimitUsd:           dailyLimitUSD,
		PriceDeviationThreshold: priceDeviationThreshold,
		TreasuryAddresses:       treasuryAddresses,
		ExchangeFee:             exchangeFee,
		SupportedTokens:         supportedTokens,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultEnabled,
		DefaultAdmin,
		DefaultMinExchangeAmountUSD,
		DefaultMaxExchangeAmountUSD,
		DefaultDailyLimitUSD,
		DefaultPriceDeviationThreshold,
		DefaultExchangeFee,
		DefaultTreasuryAddresses,
		DefaultSupportedTokens,
	)
}

// ParamKeyTable returns the parameter key table for use with the sdk.Params module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateEnabled(p.Enabled); err != nil {
		return err
	}

	if err := validateAdmin(p.Admin); err != nil {
		return err
	}

	if err := validateMinExchangeAmountUSD(p.MinExchangeAmountUsd); err != nil {
		return err
	}

	if err := validateMaxExchangeAmountUSD(p.MaxExchangeAmountUsd); err != nil {
		return err
	}

	if err := validateDailyLimitUSD(p.DailyLimitUsd); err != nil {
		return err
	}

	if err := validatePriceDeviationThreshold(p.PriceDeviationThreshold); err != nil {
		return err
	}

	if err := validateTreasuryAddresses(p.TreasuryAddresses); err != nil {
		return err
	}

	if err := validateExchangeFee(p.ExchangeFee); err != nil {
		return err
	}

	if err := validateSupportedTokens(p.SupportedTokens); err != nil {
		return err
	}

	// Validate that min < max
	if p.MinExchangeAmountUsd.GTE(p.MaxExchangeAmountUsd) {
		return fmt.Errorf("min exchange amount (%s) must be less than max exchange amount (%s)",
			p.MinExchangeAmountUsd, p.MaxExchangeAmountUsd)
	}

	return nil
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyEnabled, &p.Enabled, validateEnabled),
		paramtypes.NewParamSetPair(KeyAdmin, &p.Admin, validateAdmin),
		paramtypes.NewParamSetPair(KeyMinExchangeAmountUSD, &p.MinExchangeAmountUsd, validateMinExchangeAmountUSD),
		paramtypes.NewParamSetPair(KeyMaxExchangeAmountUSD, &p.MaxExchangeAmountUsd, validateMaxExchangeAmountUSD),
		paramtypes.NewParamSetPair(KeyDailyLimitUSD, &p.DailyLimitUsd, validateDailyLimitUSD),
		paramtypes.NewParamSetPair(KeyPriceDeviationThreshold, &p.PriceDeviationThreshold, validatePriceDeviationThreshold),
		paramtypes.NewParamSetPair(KeyTreasuryAddresses, &p.TreasuryAddresses, validateTreasuryAddresses),
		paramtypes.NewParamSetPair(KeyExchangeFee, &p.ExchangeFee, validateExchangeFee),
		paramtypes.NewParamSetPair(KeySupportedTokens, &p.SupportedTokens, validateSupportedTokens),
	}
}

// Validation functions
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

	if admin != "" {
		_, err := sdk.AccAddressFromBech32(admin)
		if err != nil {
			return fmt.Errorf("invalid admin address: %w", err)
		}
	}

	return nil
}

func validateMinExchangeAmountUSD(i interface{}) error {
	amount, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if amount.IsNegative() {
		return fmt.Errorf("min exchange amount must be non-negative: %s", amount)
	}

	return nil
}

func validateMaxExchangeAmountUSD(i interface{}) error {
	amount, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if amount.IsNegative() {
		return fmt.Errorf("max exchange amount must be non-negative: %s", amount)
	}

	return nil
}

func validateDailyLimitUSD(i interface{}) error {
	amount, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if amount.IsNegative() {
		return fmt.Errorf("daily limit must be non-negative: %s", amount)
	}

	return nil
}

func validatePriceDeviationThreshold(i interface{}) error {
	threshold, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if threshold.IsNegative() || threshold.GT(math.LegacyOneDec()) {
		return fmt.Errorf("price deviation threshold must be between 0 and 1: %s", threshold)
	}

	return nil
}

func validateTreasuryAddresses(i interface{}) error {
	addresses, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, addr := range addresses {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return fmt.Errorf("invalid treasury address %s: %w", addr, err)
		}
	}

	return nil
}

func validateExchangeFee(i interface{}) error {
	fee, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if fee.IsNegative() || fee.GT(math.LegacyOneDec()) {
		return fmt.Errorf("exchange fee must be between 0 and 1: %s", fee)
	}

	return nil
}

func validateSupportedTokens(i interface{}) error {
	v, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(v) == 0 {
		return fmt.Errorf("supported tokens list cannot be empty")
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, token := range v {
		if token == "" {
			return fmt.Errorf("supported token denomination cannot be empty")
		}
		if seen[token] {
			return fmt.Errorf("duplicate token denomination: %s", token)
		}
		seen[token] = true
	}

	return nil
}

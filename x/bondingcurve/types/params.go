package types

import (
	"fmt"
	"time"

	"github.com/osmosis-labs/osmosis/osmomath"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyBondingCurveWallet = []byte("BondingCurveWallet")
	KeyQuoteDenom         = []byte("QuoteDenom")
	KeyMaxSupply          = []byte("MaxSupply")
	KeyStartPrice         = []byte("StartPrice")
	KeyEndPrice           = []byte("EndPrice")
	KeyTokenCreationFee   = []byte("TokenCreationFee")
	KeyFounderClaimPeriod = []byte("FounderClaimPeriod")
	KeyMaxLeverage        = []byte("MaxLeverage")
	KeyLiquidationPenalty = []byte("LiquidationPenalty")
	KeyProtocolFeeRate    = []byte("ProtocolFeeRate")
	KeyMinCollateralRatio = []byte("MinCollateralRatio")
)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	wallet,
	quoteDenom,
	maxSupply,
	startPrice,
	endPrice string,
	tokenCreationFee sdk.Coin,
	founderClaim time.Duration,
	maxLeverage uint32,
	liquidationPenalty osmomath.Dec,
	protocolFeeRate osmomath.Dec,
	minCollateralRatio osmomath.Dec,
) Params {
	return Params{
		BondingCurveWallet: wallet,
		QuoteDenom:         quoteDenom,
		MaxSupply:          maxSupply,
		StartPrice:         startPrice,
		EndPrice:           endPrice,
		TokenCreationFee:   tokenCreationFee,
		FounderClaimPeriod: founderClaim,
		MaxLeverage:        maxLeverage,
		LiquidationPenalty: liquidationPenalty.String(),
		ProtocolFeeRate:    protocolFeeRate.String(),
		MinCollateralRatio: minCollateralRatio.String(),
	}
}

func DefaultParams() Params {
	return Params{
		BondingCurveWallet: "",
		QuoteDenom:         "undollar",
		MaxSupply:          "30000000.0",
		StartPrice:         "0.0002",
		EndPrice:           "1.0",
		TokenCreationFee:   sdk.NewInt64Coin("undollar", 5_000_000),
		FounderClaimPeriod: time.Hour,
		MaxLeverage:        MaxMarginLeverage,
		LiquidationPenalty: osmomath.NewDecWithPrec(5, 2).String(), // 5%
		ProtocolFeeRate:    osmomath.NewDecWithPrec(3, 3).String(), // 0.3%
		MinCollateralRatio: osmomath.NewDecWithPrec(1, 1).String(), // 10%
	}
}

func (p Params) Validate() error {
	if p.BondingCurveWallet != "" {
		if _, err := sdk.AccAddressFromBech32(p.BondingCurveWallet); err != nil {
			return fmt.Errorf("invalid bonding curve wallet: %w", err)
		}
	}

	if err := sdk.ValidateDenom(p.QuoteDenom); err != nil {
		return fmt.Errorf("invalid quote denom: %w", err)
	}

	if _, err := osmomath.NewDecFromStr(p.MaxSupply); err != nil {
		return fmt.Errorf("invalid max supply: %w", err)
	}

	startPrice, err := osmomath.NewDecFromStr(p.StartPrice)
	if err != nil {
		return fmt.Errorf("invalid start price: %w", err)
	}

	endPrice, err := osmomath.NewDecFromStr(p.EndPrice)
	if err != nil {
		return fmt.Errorf("invalid end price: %w", err)
	}

	if !endPrice.GT(startPrice) {
		return fmt.Errorf("end price must be greater than start price")
	}

	if err := validateCoin(p.TokenCreationFee); err != nil {
		return err
	}

	if err := validateDuration(p.FounderClaimPeriod); err != nil {
		return err
	}

	if err := validateMaxLeverage(p.MaxLeverage); err != nil {
		return err
	}

	if err := validatePositiveDecRange(p.LiquidationPenalty, "liquidation penalty", osmomath.ZeroDec(), osmomath.OneDec()); err != nil {
		return err
	}

	if err := validatePositiveDecRange(p.ProtocolFeeRate, "protocol fee rate", osmomath.ZeroDec(), osmomath.MustNewDecFromStr("0.1")); err != nil {
		return err
	}

	if err := validatePositiveDecRange(p.MinCollateralRatio, "min collateral ratio", osmomath.NewDecWithPrec(5, 2), osmomath.OneDec()); err != nil {
		return err
	}

	return nil
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyBondingCurveWallet, &p.BondingCurveWallet, validateWallet),
		paramtypes.NewParamSetPair(KeyQuoteDenom, &p.QuoteDenom, validateDenom),
		paramtypes.NewParamSetPair(KeyMaxSupply, &p.MaxSupply, validateDecString),
		paramtypes.NewParamSetPair(KeyStartPrice, &p.StartPrice, validateDecString),
		paramtypes.NewParamSetPair(KeyEndPrice, &p.EndPrice, validateDecString),
		paramtypes.NewParamSetPair(KeyTokenCreationFee, &p.TokenCreationFee, validateCoin),
		paramtypes.NewParamSetPair(KeyFounderClaimPeriod, &p.FounderClaimPeriod, validateDuration),
		paramtypes.NewParamSetPair(KeyMaxLeverage, &p.MaxLeverage, validateMaxLeverageParam),
		paramtypes.NewParamSetPair(KeyLiquidationPenalty, &p.LiquidationPenalty, validateDecString),
		paramtypes.NewParamSetPair(KeyProtocolFeeRate, &p.ProtocolFeeRate, validateDecString),
		paramtypes.NewParamSetPair(KeyMinCollateralRatio, &p.MinCollateralRatio, validateDecString),
	}
}

func validateWallet(i any) error {
	if addr, ok := i.(string); ok {
		if addr == "" {
			return nil
		}
		_, err := sdk.AccAddressFromBech32(addr)
		return err
	}
	return fmt.Errorf("invalid parameter type: %T", i)
}

func validateDenom(i any) error {
	denom, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return sdk.ValidateDenom(denom)
}

func validateDecString(i any) error {
	value, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	_, err := osmomath.NewDecFromStr(value)
	return err
}

func validateCoin(i any) error {
	coin, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if err := coin.Validate(); err != nil {
		return fmt.Errorf("invalid coin: %w", err)
	}
	return nil
}

func validateDuration(i any) error {
	switch value := i.(type) {
	case time.Duration:
		if value <= 0 {
			return fmt.Errorf("duration must be positive")
		}
		return nil
	case *time.Duration:
		if value == nil {
			return fmt.Errorf("duration pointer cannot be nil")
		}
		return validateDuration(*value)
	default:
		return fmt.Errorf("invalid parameter type: %T", i)
	}
}

func validateMaxLeverage(i any) error {
	value, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return validateMaxLeverageValue(value)
}

func validateMaxLeverageParam(i any) error {
	value, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return validateMaxLeverageValue(value)
}

func validateMaxLeverageValue(value uint32) error {
	if value == 0 {
		return fmt.Errorf("max leverage must be greater than zero")
	}
	if value > MaxMarginLeverage {
		return fmt.Errorf("max leverage cannot exceed %d", MaxMarginLeverage)
	}
	return nil
}

func validatePositiveDecRange(valueStr, name string, min, max osmomath.Dec) error {
	value, err := osmomath.NewDecFromStr(valueStr)
	if err != nil {
		return fmt.Errorf("invalid %s: %w", name, err)
	}
	if !value.GT(min) {
		return fmt.Errorf("%s must be greater than %s", name, min.String())
	}
	if value.GT(max) {
		return fmt.Errorf("%s must be less than or equal to %s", name, max.String())
	}
	return nil
}

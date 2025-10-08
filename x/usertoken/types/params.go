package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	appparams "github.com/osmosis-labs/osmosis/v30/app/params"
)

var (
	KeyTokenCreationFee   = []byte("TokenCreationFee")
	KeyFounderClaimPeriod = []byte("FounderClaimPeriod")
	KeyBondingCurveWallet = []byte("BondingCurveWallet")
	KeyPlatformWallet     = []byte("PlatformWallet")
	KeyReferralWallet     = []byte("ReferralWallet")
	KeyAiCeoWallet        = []byte("AiCeoWallet")
)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	creationFee sdk.Coin,
	founderClaimPeriod uint64,
	bondingCurveWallet, platformWallet, referralWallet, aiCeoWallet string,
) Params {
	return Params{
		TokenCreationFee:   creationFee,
		FounderClaimPeriod: founderClaimPeriod,
		BondingCurveWallet: bondingCurveWallet,
		PlatformWallet:     platformWallet,
		ReferralWallet:     referralWallet,
		AiCeoWallet:        aiCeoWallet,
	}
}

func DefaultParams() Params {
	return Params{
		TokenCreationFee:   sdk.NewInt64Coin(appparams.BaseCoinUnit, 0),
		FounderClaimPeriod: 3600,
	}
}

func (p Params) Validate() error {
	if err := validateTokenCreationFee(p.TokenCreationFee); err != nil {
		return err
	}
	if err := validateFounderClaimPeriod(p.FounderClaimPeriod); err != nil {
		return err
	}
	if err := validateOptionalAddress(p.BondingCurveWallet); err != nil {
		return fmt.Errorf("bonding curve wallet: %w", err)
	}
	if err := validateOptionalAddress(p.PlatformWallet); err != nil {
		return fmt.Errorf("platform wallet: %w", err)
	}
	if err := validateOptionalAddress(p.ReferralWallet); err != nil {
		return fmt.Errorf("referral wallet: %w", err)
	}
	if err := validateOptionalAddress(p.AiCeoWallet); err != nil {
		return fmt.Errorf("ai ceo wallet: %w", err)
	}
	return nil
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyTokenCreationFee, &p.TokenCreationFee, validateTokenCreationFee),
		paramtypes.NewParamSetPair(KeyFounderClaimPeriod, &p.FounderClaimPeriod, validateFounderClaimPeriod),
		paramtypes.NewParamSetPair(KeyBondingCurveWallet, &p.BondingCurveWallet, validateOptionalAddressParam),
		paramtypes.NewParamSetPair(KeyPlatformWallet, &p.PlatformWallet, validateOptionalAddressParam),
		paramtypes.NewParamSetPair(KeyReferralWallet, &p.ReferralWallet, validateOptionalAddressParam),
		paramtypes.NewParamSetPair(KeyAiCeoWallet, &p.AiCeoWallet, validateOptionalAddressParam),
	}
}

func validateTokenCreationFee(i interface{}) error {
	coin, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if coin.IsNil() || coin.IsZero() {
		return nil
	}

	if err := sdk.ValidateDenom(coin.Denom); err != nil {
		return err
	}

	if coin.Amount.IsNegative() {
		return fmt.Errorf("token creation fee cannot be negative")
	}

	return nil
}

func validateFounderClaimPeriod(i interface{}) error {
	period, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if period == 0 {
		return fmt.Errorf("founder claim period must be positive")
	}
	return nil
}

func validateOptionalAddress(addr string) error {
	if addr == "" {
		return nil
	}
	if _, err := sdk.AccAddressFromBech32(addr); err != nil {
		return err
	}
	return nil
}

func validateOptionalAddressParam(i interface{}) error {
	addr, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return validateOptionalAddress(addr)
}

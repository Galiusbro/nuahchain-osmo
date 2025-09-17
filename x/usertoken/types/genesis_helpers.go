package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// DefaultGenesis returns the default usertoken genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:     DefaultParams(),
		UserTokens: []*UserToken{},
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		FounderTranchePrice:    math.LegacyMustNewDecFromStr("0.00005"), // 0.00005 N$
		FounderTrancheAmount:   math.NewInt(10_000_000),                 // 10M tokens
		BondingCurveStartPrice: math.LegacyMustNewDecFromStr("0.0002"),  // 0.0002 N$
		BondingCurveEndPrice:   math.LegacyMustNewDecFromStr("1.0"),     // 1.0 N$
		BondingCurveMaxSupply:  math.NewInt(30_000_000),                 // 30M tokens
		MinCreatorPurchase:     math.LegacyMustNewDecFromStr("500"),     // 500 N$ minimum
		AiCeoWallet:            "",                                      // Will be set during chain initialization
		ReferralWallet:         "",                                      // Will be set during chain initialization
		PlatformFeeWallet:      "",                                      // Will be set during chain initialization
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}

// Validate validates the parameters
func (p Params) Validate() error {
	if p.FounderTranchePrice.IsNegative() {
		return fmt.Errorf("founder tranche price cannot be negative: %s", p.FounderTranchePrice)
	}
	if p.FounderTrancheAmount.IsNegative() {
		return fmt.Errorf("founder tranche amount cannot be negative: %s", p.FounderTrancheAmount)
	}
	if p.BondingCurveStartPrice.IsNegative() {
		return fmt.Errorf("bonding curve start price cannot be negative: %s", p.BondingCurveStartPrice)
	}
	if p.BondingCurveEndPrice.IsNegative() {
		return fmt.Errorf("bonding curve end price cannot be negative: %s", p.BondingCurveEndPrice)
	}
	if p.BondingCurveMaxSupply.IsNegative() {
		return fmt.Errorf("bonding curve max supply cannot be negative: %s", p.BondingCurveMaxSupply)
	}
	if p.BondingCurveStartPrice.GTE(p.BondingCurveEndPrice) {
		return fmt.Errorf("bonding curve start price must be less than end price: %s >= %s", p.BondingCurveStartPrice, p.BondingCurveEndPrice)
	}
	if p.MinCreatorPurchase.IsNegative() {
		return fmt.Errorf("minimum creator purchase cannot be negative: %s", p.MinCreatorPurchase)
	}
	// Note: Wallet addresses can be empty strings initially and will be validated when set
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  FounderTranchePrice: %s
  FounderTrancheAmount: %s
  BondingCurveStartPrice: %s
  BondingCurveEndPrice: %s
  BondingCurveMaxSupply: %s`,
		p.FounderTranchePrice, p.FounderTrancheAmount, p.BondingCurveStartPrice, p.BondingCurveEndPrice, p.BondingCurveMaxSupply)
}

package types

import (
	"fmt"

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
)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(wallet, quoteDenom, maxSupply, startPrice, endPrice string) Params {
	return Params{
		BondingCurveWallet: wallet,
		QuoteDenom:         quoteDenom,
		MaxSupply:          maxSupply,
		StartPrice:         startPrice,
		EndPrice:           endPrice,
	}
}

func DefaultParams() Params {
	return Params{
		BondingCurveWallet: "",
		QuoteDenom:         "undollar",
		MaxSupply:          "30000000.0",
		StartPrice:         "0.0002",
		EndPrice:           "1.0",
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

	return nil
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyBondingCurveWallet, &p.BondingCurveWallet, validateWallet),
		paramtypes.NewParamSetPair(KeyQuoteDenom, &p.QuoteDenom, validateDenom),
		paramtypes.NewParamSetPair(KeyMaxSupply, &p.MaxSupply, validateDecString),
		paramtypes.NewParamSetPair(KeyStartPrice, &p.StartPrice, validateDecString),
		paramtypes.NewParamSetPair(KeyEndPrice, &p.EndPrice, validateDecString),
	}
}

func validateWallet(i interface{}) error {
	if addr, ok := i.(string); ok {
		if addr == "" {
			return nil
		}
		_, err := sdk.AccAddressFromBech32(addr)
		return err
	}
	return fmt.Errorf("invalid parameter type: %T", i)
}

func validateDenom(i interface{}) error {
	denom, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return sdk.ValidateDenom(denom)
}

func validateDecString(i interface{}) error {
	value, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	_, err := osmomath.NewDecFromStr(value)
	return err
}

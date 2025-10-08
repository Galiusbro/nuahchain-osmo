package types

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/osmomath"
)

func MustIntFromString(value string) osmomath.Int {
	if value == "" {
		return osmomath.ZeroInt()
	}
	intVal, ok := osmomath.NewIntFromString(value)
	if !ok {
		panic(fmt.Sprintf("invalid integer string: %s", value))
	}
	return intVal
}

func (d *TokenDistribution) SetTotalSupply(amount osmomath.Int) {
	d.TotalSupply = amount.String()
}

func (d *TokenDistribution) SetBondingCurveSupply(amount osmomath.Int) {
	d.BondingCurveSupply = amount.String()
}

func (d *TokenDistribution) SetPlatformAllocation(amount osmomath.Int) {
	d.PlatformWallet = amount.String()
}

func (d *TokenDistribution) SetReferralAllocation(amount osmomath.Int) {
	d.ReferralWallet = amount.String()
}

func (d *TokenDistribution) SetAiCeoAllocation(amount osmomath.Int) {
	d.AiCeoWallet = amount.String()
}

func (d *TokenDistribution) SetFounderReserved(amount osmomath.Int) {
	d.FounderReserved = amount.String()
}

func (d TokenDistribution) FounderReservedInt() osmomath.Int {
	return MustIntFromString(d.FounderReserved)
}

func (d TokenDistribution) BondingCurveSupplyInt() osmomath.Int {
	return MustIntFromString(d.BondingCurveSupply)
}

func (d TokenDistribution) PlatformAllocationInt() osmomath.Int {
	return MustIntFromString(d.PlatformWallet)
}

func (d TokenDistribution) ReferralAllocationInt() osmomath.Int {
	return MustIntFromString(d.ReferralWallet)
}

func (d TokenDistribution) AiCeoAllocationInt() osmomath.Int {
	return MustIntFromString(d.AiCeoWallet)
}

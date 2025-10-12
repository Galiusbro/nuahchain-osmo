package types

import (
	"github.com/osmosis-labs/osmosis/osmomath"
)

func MustNewDecFromStr(value string) osmomath.Dec {
	if value == "" {
		return osmomath.ZeroDec()
	}
	dec, err := osmomath.NewDecFromStr(value)
	if err != nil {
		panic(err)
	}
	return dec
}

func (p Params) MaxSupplyDec() osmomath.Dec {
	return MustNewDecFromStr(p.MaxSupply)
}

func (p Params) StartPriceDec() osmomath.Dec {
	return MustNewDecFromStr(p.StartPrice)
}

func (p Params) EndPriceDec() osmomath.Dec {
	return MustNewDecFromStr(p.EndPrice)
}

func (p Params) LiquidationPenaltyDec() osmomath.Dec {
	return MustNewDecFromStr(p.LiquidationPenalty)
}

func (p Params) ProtocolFeeRateDec() osmomath.Dec {
	return MustNewDecFromStr(p.ProtocolFeeRate)
}

func (p Params) MinCollateralRatioDec() osmomath.Dec {
	return MustNewDecFromStr(p.MinCollateralRatio)
}

func (p Params) PriceSlope() osmomath.Dec {
	maxSupply := p.MaxSupplyDec()
	if maxSupply.IsZero() {
		return osmomath.ZeroDec()
	}
	return p.EndPriceDec().Sub(p.StartPriceDec()).Quo(maxSupply)
}

func CalculatePrice(tokensSold osmomath.Dec, params Params) osmomath.Dec {
	return params.StartPriceDec().Add(params.PriceSlope().Mul(tokensSold))
}

func IntegrateBuyAmount(tokensSold osmomath.Dec, payment osmomath.Dec, params Params) osmomath.Dec {
	slope := params.PriceSlope()
	base := params.StartPriceDec().Add(slope.Mul(tokensSold))

	// Solve quadratic: (base) * x + 0.5 * slope * x^2 = payment
	if slope.IsZero() {
		if base.IsZero() {
			return osmomath.ZeroDec()
		}
		return payment.Quo(base)
	}

	a := slope.Quo(osmomath.NewDec(2))
	b := base

	discriminant := b.Mul(b).Add(a.Mul(osmomath.NewDec(4)).Mul(payment))

	root := osmomath.MustMonotonicSqrt(discriminant)
	return root.Sub(b).Quo(a.Mul(osmomath.NewDec(2)))
}

func IntegrateSellAmount(tokensSold osmomath.Dec, tokensIn osmomath.Dec, params Params) osmomath.Dec {
	slope := params.PriceSlope()
	base := params.StartPriceDec().Add(slope.Mul(tokensSold))

	// payment = base * x - 0.5 * slope * x^2
	if slope.IsZero() {
		return base.Mul(tokensIn)
	}

	return base.Mul(tokensIn).Sub(slope.Quo(osmomath.NewDec(2)).Mul(tokensIn).Mul(tokensIn))
}

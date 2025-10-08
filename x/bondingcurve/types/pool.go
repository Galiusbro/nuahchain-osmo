package types

import "github.com/osmosis-labs/osmosis/osmomath"

func (p BondingCurvePool) TokensSoldDec() osmomath.Dec {
	return MustNewDecFromStr(p.TokensSold)
}

func (p BondingCurvePool) ReserveNuahDec() osmomath.Dec {
	return MustNewDecFromStr(p.ReserveNuah)
}

func (p BondingCurvePool) ReserveNdollarDec() osmomath.Dec {
	return MustNewDecFromStr(p.ReserveNdollar)
}

func (p BondingCurvePool) LastPriceDec() osmomath.Dec {
	return MustNewDecFromStr(p.LastPrice)
}

func (p *BondingCurvePool) SetTokensSold(amount osmomath.Dec) {
	p.TokensSold = amount.String()
}

func (p *BondingCurvePool) SetReserveNuah(amount osmomath.Dec) {
	p.ReserveNuah = amount.String()
}

func (p *BondingCurvePool) SetReserveNdollar(amount osmomath.Dec) {
	p.ReserveNdollar = amount.String()
}

func (p *BondingCurvePool) SetLastPrice(price osmomath.Dec) {
	p.LastPrice = price.String()
}

func (p *BondingCurvePool) MarkDexActivated(poolID uint64, provider string) {
	p.DexPoolId = poolID
	p.DexActivated = true
	p.LiquidityProvider = provider
}

package types

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/osmomath"
)

const (
	MinMarginLeverage = uint32(1)
	MaxMarginLeverage = uint32(100)
)

var (
	DefaultMaintenanceMarginRatio = osmomath.MustNewDecFromStr("0.01")    // 1% maintenance margin
	LiquidationBufferFactor       = osmomath.MustNewDecFromStr("0.9")     // 10% safety buffer
	DefaultMarginFeeRate          = osmomath.NewDecWithPrec(5, 3)         // 0.5% fee
	DefaultFundingRate            = osmomath.ZeroDec()                    // placeholder for v1
	MarginPricePrecision          = osmomath.NewDecWithPrec(1, 6)         // helper to ensure precision
	LiquidationPenaltyRate        = osmomath.NewDecWithPrec(5, 2)         // 5% liquidation penalty on collateral
	LiquidatorIncentiveRate       = osmomath.NewDecWithPrec(3, 1)         // 30% of penalty goes to liquidator incentives
	PartialLiquidationBuffer      = osmomath.NewDecWithPrec(2, 2)         // 2% buffer for partial liquidation window
	PartialLiquidationFraction    = osmomath.NewDecWithPrec(5, 1)         // default 50% partial close
	MaxPartialPositionSize        = osmomath.MustNewDecFromStr("1000000") // threshold for partial liquidation consideration
	CircuitBreakerThreshold       = osmomath.NewDecWithPrec(5, 1)         // 50% deviation triggers circuit breaker
	TwapSmoothingFactor           = osmomath.NewDecWithPrec(2, 1)         // 0.2 smoothing factor
)

func ValidateLeverage(leverage uint32) bool {
	return leverage >= MinMarginLeverage && leverage <= MaxMarginLeverage
}

func PositionSize(collateral osmomath.Dec, leverage uint32) osmomath.Dec {
	return collateral.Mul(osmomath.NewDec(int64(leverage)))
}

func CalculateMaintenanceMargin(positionSize osmomath.Dec) osmomath.Dec {
	return positionSize.Mul(DefaultMaintenanceMarginRatio)
}

func CalculateLiquidationPrice(entryPrice osmomath.Dec, leverage uint32, positionType PositionType) (osmomath.Dec, error) {
	if !ValidateLeverage(leverage) {
		return osmomath.ZeroDec(), fmt.Errorf("invalid leverage")
	}

	if !entryPrice.IsPositive() {
		return osmomath.ZeroDec(), fmt.Errorf("entry price must be positive")
	}

	invLeverage := osmomath.NewDec(1).Quo(osmomath.NewDec(int64(leverage)))
	buffer := invLeverage.Mul(LiquidationBufferFactor)

	switch positionType {
	case PositionType_POSITION_TYPE_LONG:
		return entryPrice.Mul(osmomath.OneDec().Sub(buffer)), nil
	case PositionType_POSITION_TYPE_SHORT:
		return entryPrice.Mul(osmomath.OneDec().Add(buffer)), nil
	default:
		return osmomath.ZeroDec(), fmt.Errorf("unknown position type")
	}
}

func (m MarginPool) TotalCollateralDec() osmomath.Dec {
	return MustNewDecFromStr(m.TotalCollateral)
}

func (m MarginPool) AvailableLiquidityDec() osmomath.Dec {
	return MustNewDecFromStr(m.AvailableLiquidity)
}

func (m MarginPool) TotalLongExposureDec() osmomath.Dec {
	return MustNewDecFromStr(m.TotalLongExposure)
}

func (m MarginPool) TotalShortExposureDec() osmomath.Dec {
	return MustNewDecFromStr(m.TotalShortExposure)
}

func (m MarginPool) MaintenanceMarginRatioDec() osmomath.Dec {
	if m.MaintenanceMarginRatio == "" {
		return DefaultMaintenanceMarginRatio
	}
	return MustNewDecFromStr(m.MaintenanceMarginRatio)
}

func (m MarginPool) CumulativeFundingRateDec() osmomath.Dec {
	return MustNewDecFromStr(m.CumulativeFundingRate)
}

func (m MarginPool) InsuranceFundDec() osmomath.Dec {
	return MustNewDecFromStr(m.InsuranceFund)
}

func (m MarginPool) TotalLiquidationFeesDec() osmomath.Dec {
	return MustNewDecFromStr(m.TotalLiquidationFees)
}

func (m MarginPool) TotalBadDebtDec() osmomath.Dec {
	return MustNewDecFromStr(m.TotalBadDebt)
}

func (m MarginPool) LastMarkPriceDec() osmomath.Dec {
	return MustNewDecFromStr(m.LastMarkPrice)
}

func (m MarginPool) LastTwapPriceDec() osmomath.Dec {
	return MustNewDecFromStr(m.LastTwapPrice)
}

func (m *MarginPool) SetTotalCollateral(val osmomath.Dec) {
	m.TotalCollateral = val.String()
}

func (m *MarginPool) SetAvailableLiquidity(val osmomath.Dec) {
	m.AvailableLiquidity = val.String()
}

func (m *MarginPool) SetTotalLongExposure(val osmomath.Dec) {
	m.TotalLongExposure = val.String()
}

func (m *MarginPool) SetTotalShortExposure(val osmomath.Dec) {
	m.TotalShortExposure = val.String()
}

func (m *MarginPool) SetMaintenanceMarginRatio(val osmomath.Dec) {
	m.MaintenanceMarginRatio = val.String()
}

func (m *MarginPool) SetCumulativeFundingRate(val osmomath.Dec) {
	m.CumulativeFundingRate = val.String()
}

func (m *MarginPool) SetInsuranceFund(val osmomath.Dec) {
	m.InsuranceFund = val.String()
}

func (m *MarginPool) SetTotalLiquidationFees(val osmomath.Dec) {
	m.TotalLiquidationFees = val.String()
}

func (m *MarginPool) SetTotalBadDebt(val osmomath.Dec) {
	m.TotalBadDebt = val.String()
}

func (m *MarginPool) SetLastMarkPrice(val osmomath.Dec) {
	m.LastMarkPrice = val.String()
}

func (m *MarginPool) SetLastTwapPrice(val osmomath.Dec) {
	m.LastTwapPrice = val.String()
}

func (mp MarginPosition) CollateralAmountDec() osmomath.Dec {
	return MustNewDecFromStr(mp.CollateralAmount)
}

func (mp MarginPosition) PositionSizeDec() osmomath.Dec {
	return MustNewDecFromStr(mp.PositionSize)
}

func (mp MarginPosition) EntryPriceDec() osmomath.Dec {
	return MustNewDecFromStr(mp.EntryPrice)
}

func (mp MarginPosition) LiquidationPriceDec() osmomath.Dec {
	return MustNewDecFromStr(mp.LiquidationPrice)
}

func (mp MarginPosition) MaintenanceMarginDec() osmomath.Dec {
	return MustNewDecFromStr(mp.MaintenanceMargin)
}

func (mp MarginPosition) RealizedPnLDec() osmomath.Dec {
	return MustNewDecFromStr(mp.RealizedPnl)
}

func (mp MarginPosition) LastMarkPriceDec() osmomath.Dec {
	return MustNewDecFromStr(mp.LastMarkPrice)
}

func (mp *MarginPosition) SetMaintenanceMargin(amount osmomath.Dec) {
	mp.MaintenanceMargin = amount.String()
}

func (mp *MarginPosition) SetLiquidationPrice(price osmomath.Dec) {
	mp.LiquidationPrice = price.String()
}

func (mp *MarginPosition) SetPositionSize(size osmomath.Dec) {
	mp.PositionSize = size.String()
}

func (mp *MarginPosition) SetCollateralAmount(amount osmomath.Dec) {
	mp.CollateralAmount = amount.String()
}

func (mp *MarginPosition) SetRealizedPnl(amount osmomath.Dec) {
	mp.RealizedPnl = amount.String()
}

func (mp *MarginPosition) SetLastMarkPrice(price osmomath.Dec) {
	mp.LastMarkPrice = price.String()
}

func (mp *MarginPosition) SetStatus(status PositionStatus) {
	mp.Status = status
}

func (mp MarginPosition) CalculatePnL(markPrice osmomath.Dec) osmomath.Dec {
	entryPrice := mp.EntryPriceDec()
	if !entryPrice.IsPositive() || !markPrice.IsPositive() {
		return osmomath.ZeroDec()
	}

	priceDiff := markPrice.Sub(entryPrice)
	if mp.Type == PositionType_POSITION_TYPE_SHORT {
		priceDiff = entryPrice.Sub(markPrice)
	}

	return priceDiff.Quo(entryPrice).Mul(mp.PositionSizeDec())
}

type PriceInfo struct {
	MarkPrice osmomath.Dec
	TwapPrice osmomath.Dec
}

func (pi PriceInfo) IsCircuitBreakerTriggered() bool {
	if pi.TwapPrice.IsZero() {
		return false
	}
	deviation := pi.MarkPrice.Sub(pi.TwapPrice).Abs().Quo(pi.TwapPrice)
	return deviation.GTE(CircuitBreakerThreshold)
}

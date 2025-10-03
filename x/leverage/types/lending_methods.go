package types

import (
	"cosmossdk.io/math"
)

// CalculateUtilizationRate calculates the utilization rate of a lending pool
func (pool LendingPool) CalculateUtilizationRate() math.LegacyDec {
	if pool.TotalSupply.IsZero() {
		return math.LegacyZeroDec()
	}

	return math.LegacyNewDecFromInt(pool.TotalBorrowed).Quo(math.LegacyNewDecFromInt(pool.TotalSupply))
}

// GetTotalDebt returns the total debt (borrowed amount + accrued interest)
func (bp BorrowPosition) GetTotalDebt() math.Int {
	return bp.BorrowedAmount.Add(bp.AccruedInterest)
}

// Validate performs basic validation on LendingPool
func (pool LendingPool) Validate() error {
	if pool.Denom == "" {
		return ErrInvalidTokenDenom
	}
	if pool.TotalSupply.IsNegative() {
		return ErrInvalidAmount
	}
	if pool.TotalBorrowed.IsNegative() {
		return ErrInvalidAmount
	}
	if pool.TotalBorrowed.GT(pool.TotalSupply) {
		return ErrInsufficientLiquidity
	}
	return nil
}

// Validate performs basic validation on BorrowPosition
func (bp BorrowPosition) Validate() error {
	if bp.Id == "" {
		return ErrInvalidPositionID
	}
	if bp.Borrower == "" {
		return ErrInvalidTraderAddress
	}
	if bp.TokenDenom == "" {
		return ErrInvalidTokenDenom
	}
	if bp.BorrowedAmount.IsNegative() {
		return ErrInvalidAmount
	}
	if bp.AccruedInterest.IsNegative() {
		return ErrInvalidAmount
	}
	return nil
}

// InterestRateModel defines how interest rates are calculated
type InterestRateModel struct {
	BaseRate       math.LegacyDec `json:"base_rate"`
	Multiplier     math.LegacyDec `json:"multiplier"`
	JumpMultiplier math.LegacyDec `json:"jump_multiplier"`
	Kink           math.LegacyDec `json:"kink"`
}

// CalculateInterestRate calculates the current interest rate based on utilization
func (model InterestRateModel) CalculateInterestRate(utilizationRate math.LegacyDec) math.LegacyDec {
	if utilizationRate.LTE(model.Kink) {
		// Below kink: rate = baseRate + (utilizationRate * multiplier)
		return model.BaseRate.Add(utilizationRate.Mul(model.Multiplier))
	} else {
		// Above kink: rate = baseRate + (kink * multiplier) + ((utilizationRate - kink) * jumpMultiplier)
		normalRate := model.BaseRate.Add(model.Kink.Mul(model.Multiplier))
		excessUtilization := utilizationRate.Sub(model.Kink)
		jumpRate := excessUtilization.Mul(model.JumpMultiplier)
		return normalRate.Add(jumpRate)
	}
}

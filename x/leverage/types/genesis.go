package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// DefaultGenesis returns the default leverage genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:             DefaultParams(),
		Positions:          []Position{},
		NextPositionId:     1,
		LendingPools:       []LendingPool{},
		BorrowPositions:    []BorrowPosition{},
		LiquidityProviders: []LiquidityProvider{},
		NextBorrowId:       1,
	}
}

// DefaultParams returns default parameters
func DefaultParams() LeverageParams {
	return LeverageParams{
		MaxLeverage:            math.LegacyNewDec(100),                 // 100x max leverage
		MaintenanceMargin:      math.LegacyNewDecWithPrec(1, 2),        // 1% maintenance margin
		LiquidationFee:         math.LegacyNewDecWithPrec(5, 3),        // 0.5% liquidation fee
		TradingFee:             math.LegacyNewDecWithPrec(1, 3),        // 0.1% trading fee
		MaxPositionSize:        math.NewInt(1_000_000_000_000_000_000), // 1B tokens max position
		MinCollateralAmount:    math.NewInt(1_000_000),                 // 1 token min collateral (6 decimals)
		BaseInterestRate:       math.LegacyNewDecWithPrec(2, 2),        // 2% base interest rate
		InterestRateMultiplier: math.LegacyNewDecWithPrec(10, 2),       // 10% interest rate multiplier
		MaxInterestRate:        math.LegacyNewDecWithPrec(50, 2),       // 50% max interest rate
		MaxBorrowRatio:         math.LegacyNewDecWithPrec(80, 2),       // 80% max borrow ratio
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate positions
	positionIDs := make(map[string]bool)
	for _, position := range gs.Positions {
		if err := position.Validate(); err != nil {
			return fmt.Errorf("invalid position %s: %w", position.Id, err)
		}

		// Check for duplicate position IDs
		if positionIDs[position.Id] {
			return fmt.Errorf("duplicate position ID: %s", position.Id)
		}
		positionIDs[position.Id] = true
	}

	// Validate next position ID
	if gs.NextPositionId == 0 {
		return fmt.Errorf("next position ID cannot be zero")
	}

	return nil
}

// Validate validates the parameters
func (p LeverageParams) Validate() error {
	if p.MaxLeverage.IsNil() || p.MaxLeverage.LTE(math.LegacyOneDec()) {
		return fmt.Errorf("max leverage must be greater than 1")
	}

	if p.MaintenanceMargin.IsNil() || p.MaintenanceMargin.LTE(math.LegacyZeroDec()) || p.MaintenanceMargin.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("maintenance margin must be between 0 and 1")
	}

	if p.LiquidationFee.IsNil() || p.LiquidationFee.LT(math.LegacyZeroDec()) || p.LiquidationFee.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("liquidation fee must be between 0 and 1")
	}

	if p.TradingFee.IsNil() || p.TradingFee.LT(math.LegacyZeroDec()) || p.TradingFee.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("trading fee must be between 0 and 1")
	}

	if p.MaxPositionSize.IsNil() || p.MaxPositionSize.LTE(math.ZeroInt()) {
		return fmt.Errorf("max position size must be positive")
	}

	if p.MinCollateralAmount.IsNil() || p.MinCollateralAmount.LTE(math.ZeroInt()) {
		return fmt.Errorf("min collateral amount must be positive")
	}

	if p.BaseInterestRate.IsNil() || p.BaseInterestRate.IsNegative() {
		return fmt.Errorf("base interest rate cannot be negative")
	}

	if p.InterestRateMultiplier.IsNil() || p.InterestRateMultiplier.IsNegative() {
		return fmt.Errorf("interest rate multiplier cannot be negative")
	}

	if p.MaxInterestRate.IsNil() || p.MaxInterestRate.LTE(math.LegacyZeroDec()) {
		return fmt.Errorf("max interest rate must be positive")
	}

	if p.MaxBorrowRatio.IsNil() || p.MaxBorrowRatio.LTE(math.LegacyZeroDec()) || p.MaxBorrowRatio.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("max borrow ratio must be between 0 and 1")
	}

	return nil
}

// Validate validates a position
func (p Position) Validate() error {
	if p.Id == "" {
		return fmt.Errorf("position ID cannot be empty")
	}

	if p.Trader == "" {
		return fmt.Errorf("trader address cannot be empty")
	}

	if p.TokenDenom == "" {
		return fmt.Errorf("token denom cannot be empty")
	}

	if p.CollateralDenom == "" {
		return fmt.Errorf("collateral denom cannot be empty")
	}

	if p.Side == PositionSideUnspecified {
		return fmt.Errorf("position side cannot be unspecified")
	}

	if p.Size_.IsNil() || p.Size_.LTE(math.ZeroInt()) {
		return fmt.Errorf("position size must be positive")
	}

	if p.Collateral.IsNil() || p.Collateral.LTE(math.ZeroInt()) {
		return fmt.Errorf("collateral must be positive")
	}

	if p.Leverage.IsNil() || p.Leverage.LTE(math.LegacyOneDec()) {
		return fmt.Errorf("leverage must be greater than 1")
	}

	if p.EntryPrice.IsNil() || p.EntryPrice.LTE(math.LegacyZeroDec()) {
		return fmt.Errorf("entry price must be positive")
	}

	if p.LiquidationPrice.IsNil() || p.LiquidationPrice.LTE(math.LegacyZeroDec()) {
		return fmt.Errorf("liquidation price must be positive")
	}

	if p.Status == PositionStatusUnspecified {
		return fmt.Errorf("position status cannot be unspecified")
	}

	if p.CreatedAt <= 0 {
		return fmt.Errorf("created at must be positive")
	}

	if p.UpdatedAt <= 0 {
		return fmt.Errorf("updated at must be positive")
	}

	return nil
}

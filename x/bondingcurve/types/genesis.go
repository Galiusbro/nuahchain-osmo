package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// DefaultGenesis returns default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:          DefaultParams(),
		Pools:           []BondingCurvePool{},
		MarginPools:     []MarginPool{},
		MarginPositions: []MarginPosition{},
		TokenPauses:     []TokenPauseEntry{},
		FreezeEntries:   []FreezeEntry{},
		EmergencyActions: []EmergencyAction{},
	}
}

// Validate performs basic validation of genesis data.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	seenDenoms := make(map[string]struct{})
	for _, pool := range gs.Pools {
		if _, exists := seenDenoms[pool.Denom]; exists {
			return ErrInvalidToken
		}
		seenDenoms[pool.Denom] = struct{}{}
	}

	marginDenoms := make(map[string]struct{})
	for _, marginPool := range gs.MarginPools {
		if marginPool.Denom == "" {
			return fmt.Errorf("margin pool denom cannot be empty")
		}
		if _, exists := marginDenoms[marginPool.Denom]; exists {
			return fmt.Errorf("duplicate margin pool denom: %s", marginPool.Denom)
		}
		marginDenoms[marginPool.Denom] = struct{}{}
		if _, err := osmomath.NewDecFromStr(marginPool.AvailableLiquidity); err != nil {
			return fmt.Errorf("invalid margin pool liquidity: %w", err)
		}
		if _, err := osmomath.NewDecFromStr(marginPool.TotalCollateral); err != nil {
			return fmt.Errorf("invalid margin pool collateral: %w", err)
		}
		if _, err := osmomath.NewDecFromStr(marginPool.TotalLongExposure); err != nil {
			return fmt.Errorf("invalid margin pool long exposure: %w", err)
		}
		if _, err := osmomath.NewDecFromStr(marginPool.TotalShortExposure); err != nil {
			return fmt.Errorf("invalid margin pool short exposure: %w", err)
		}
		if _, err := osmomath.NewDecFromStr(marginPool.InsuranceFund); err != nil {
			return fmt.Errorf("invalid margin pool insurance fund: %w", err)
		}
		if _, err := osmomath.NewDecFromStr(marginPool.TotalLiquidationFees); err != nil {
			return fmt.Errorf("invalid margin pool liquidation fees: %w", err)
		}
		if _, err := osmomath.NewDecFromStr(marginPool.TotalBadDebt); err != nil {
			return fmt.Errorf("invalid margin pool bad debt: %w", err)
		}
		if _, err := osmomath.NewDecFromStr(marginPool.LastMarkPrice); err != nil {
			return fmt.Errorf("invalid margin pool last mark price: %w", err)
		}
		if _, err := osmomath.NewDecFromStr(marginPool.LastTwapPrice); err != nil {
			return fmt.Errorf("invalid margin pool TWAP price: %w", err)
		}
		if marginPool.NextPositionId == 0 {
			return fmt.Errorf("margin pool %s must have next_position_id > 0", marginPool.Denom)
		}
	}

	positionIDs := make(map[uint64]struct{})
	for _, position := range gs.MarginPositions {
		if position.Id == 0 {
			return fmt.Errorf("margin position id must be > 0")
		}
		if _, exists := positionIDs[position.Id]; exists {
			return fmt.Errorf("duplicate margin position id: %d", position.Id)
		}
		positionIDs[position.Id] = struct{}{}
		if _, err := osmomath.NewDecFromStr(position.CollateralAmount); err != nil {
			return fmt.Errorf("invalid collateral amount: %w", err)
		}
		if _, err := osmomath.NewDecFromStr(position.PositionSize); err != nil {
			return fmt.Errorf("invalid position size: %w", err)
		}
		if _, err := osmomath.NewDecFromStr(position.EntryPrice); err != nil {
			return fmt.Errorf("invalid entry price: %w", err)
		}
		if _, err := osmomath.NewDecFromStr(position.LiquidationPrice); err != nil {
			return fmt.Errorf("invalid liquidation price: %w", err)
		}
		if _, err := osmomath.NewDecFromStr(position.MaintenanceMargin); err != nil {
			return fmt.Errorf("invalid maintenance margin: %w", err)
		}
		if position.RealizedPnl != "" {
			if _, err := osmomath.NewDecFromStr(position.RealizedPnl); err != nil {
				return fmt.Errorf("invalid realized pnl: %w", err)
			}
		}
		if position.LastMarkPrice != "" {
			if _, err := osmomath.NewDecFromStr(position.LastMarkPrice); err != nil {
				return fmt.Errorf("invalid last mark price: %w", err)
			}
		}
		if position.Denom == "" || position.Trader == "" {
			return fmt.Errorf("margin position must include denom and trader")
		}
		if _, exists := marginDenoms[position.Denom]; !exists {
			return fmt.Errorf("missing margin pool for denom %s", position.Denom)
		}
		if !ValidateLeverage(position.Leverage) {
			return fmt.Errorf("invalid leverage in position id %d", position.Id)
		}
		switch position.Status {
		case PositionStatus_POSITION_STATUS_UNSPECIFIED,
			PositionStatus_POSITION_STATUS_OPEN,
			PositionStatus_POSITION_STATUS_CLOSED,
			PositionStatus_POSITION_STATUS_LIQUIDATED:
		default:
			return fmt.Errorf("invalid position status for id %d", position.Id)
		}
	}

	pauseDenoms := make(map[string]struct{})
	for _, pause := range gs.TokenPauses {
		if err := sdk.ValidateDenom(pause.Denom); err != nil {
			return fmt.Errorf("invalid token pause denom: %w", err)
		}
		if _, exists := pauseDenoms[pause.Denom]; exists {
			return fmt.Errorf("duplicate token pause denom: %s", pause.Denom)
		}
		pauseDenoms[pause.Denom] = struct{}{}
	}

	freezeKeys := make(map[string]struct{})
	for _, freeze := range gs.FreezeEntries {
		if freeze.Target == "" {
			return fmt.Errorf("freeze target cannot be empty")
		}
		key := fmt.Sprintf("%d:%s", freeze.TargetType, freeze.Target)
		if _, exists := freezeKeys[key]; exists {
			return fmt.Errorf("duplicate freeze entry for target %s", freeze.Target)
		}
		switch freeze.TargetType {
		case FreezeTargetType_FREEZE_TARGET_TYPE_ADDRESS:
			if _, err := sdk.AccAddressFromBech32(freeze.Target); err != nil {
				return fmt.Errorf("invalid freeze address: %w", err)
			}
		case FreezeTargetType_FREEZE_TARGET_TYPE_TOKEN:
			if err := sdk.ValidateDenom(freeze.Target); err != nil {
				return fmt.Errorf("invalid freeze denom: %w", err)
			}
		default:
			return fmt.Errorf("invalid freeze target type: %d", freeze.TargetType)
		}
		freezeKeys[key] = struct{}{}
	}

	actionIDs := make(map[uint64]struct{})
	for _, action := range gs.EmergencyActions {
		if action.Id == 0 {
			return fmt.Errorf("emergency action id must be > 0")
		}
		if _, exists := actionIDs[action.Id]; exists {
			return fmt.Errorf("duplicate emergency action id: %d", action.Id)
		}
		actionIDs[action.Id] = struct{}{}
		if len(action.Signers) == 0 {
			return fmt.Errorf("emergency action %d must include signers", action.Id)
		}
	}

	if gs.PendingParams != nil {
		if err := gs.PendingParams.Params.Validate(); err != nil {
			return fmt.Errorf("invalid pending params: %w", err)
		}
	}

	if gs.EmergencyConfig != nil {
		if gs.EmergencyConfig.Threshold == 0 {
			return fmt.Errorf("emergency config threshold must be > 0")
		}
		if len(gs.EmergencyConfig.Signers) > 0 && gs.EmergencyConfig.Threshold > uint32(len(gs.EmergencyConfig.Signers)) {
			return fmt.Errorf("emergency config threshold exceeds signer count")
		}
	}

	return nil
}

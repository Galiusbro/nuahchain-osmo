package keeper_test

import (
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/osmomath"

	"github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
)

func (s *KeeperTestSuite) moduleAuthority() string {
	return s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName).String()
}

func (s *KeeperTestSuite) signers() []string {
	return []string{s.creator.String()}
}

func (s *KeeperTestSuite) TestEmergencyPauseBlocksBuys() {
	authority := s.moduleAuthority()

	_, err := s.msgServer.SetEmergencyPause(sdk.WrapSDKContext(s.Ctx), &types.MsgSetEmergencyPause{
		Authority: authority,
		Paused:    true,
		Signers:   s.signers(),
	})
	s.Require().NoError(err)

	_, err = s.msgServer.BuyFromCurve(s.Ctx, &types.MsgBuyFromCurve{
		Trader:        s.trader.String(),
		Denom:         s.denom,
		PaymentDenom:  s.quoteDenom,
		PaymentAmount: "10.0",
	})
	s.Require().ErrorIs(err, types.ErrModulePaused)

	_, err = s.msgServer.SetEmergencyPause(sdk.WrapSDKContext(s.Ctx), &types.MsgSetEmergencyPause{
		Authority: authority,
		Paused:    false,
		Signers:   s.signers(),
	})
	s.Require().NoError(err)

	_, err = s.msgServer.BuyFromCurve(s.Ctx, &types.MsgBuyFromCurve{
		Trader:        s.trader.String(),
		Denom:         s.denom,
		PaymentDenom:  s.quoteDenom,
		PaymentAmount: "10.0",
	})
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestFreezeAddressPreventsActions() {
	authority := s.moduleAuthority()

	_, err := s.msgServer.SetFreeze(sdk.WrapSDKContext(s.Ctx), &types.MsgSetFreeze{
		Authority:  authority,
		TargetType: types.FreezeTargetType_FREEZE_TARGET_TYPE_ADDRESS,
		Target:     s.trader.String(),
		Frozen:     true,
		Signers:    s.signers(),
	})
	s.Require().NoError(err)

	_, err = s.msgServer.BuyFromCurve(s.Ctx, &types.MsgBuyFromCurve{
		Trader:        s.trader.String(),
		Denom:         s.denom,
		PaymentDenom:  s.quoteDenom,
		PaymentAmount: "5.0",
	})
	s.Require().ErrorIs(err, types.ErrAddressFrozen)
}

func (s *KeeperTestSuite) TestUpdateParamsAppliesAfterDelay() {
	authority := s.moduleAuthority()
	params := s.App.BondingCurveKeeper.GetParams(s.Ctx)
	params.MaxLeverage = 50

	_, err := s.msgServer.UpdateParams(sdk.WrapSDKContext(s.Ctx), &types.MsgUpdateParams{
		Authority: authority,
		Params:    params,
	})
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(types.SensitiveParamChangeDelay).Add(time.Second))
	s.Require().NoError(s.App.BondingCurveKeeper.EndBlocker(s.Ctx))

	updated := s.App.BondingCurveKeeper.GetParams(s.Ctx)
	s.Require().Equal(uint32(50), updated.MaxLeverage)
}

func (s *KeeperTestSuite) TestForceLiquidationClosesPosition() {
	authority := s.moduleAuthority()
	s.setupMarginLiquidity(math.NewInt(1_000_000_000), osmomath.MustNewDecFromStr("1.0"))

	resp, err := s.msgServer.OpenMarginPosition(sdk.WrapSDKContext(s.Ctx), &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "1000.0",
		Leverage:         2,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
	})
	s.Require().NoError(err)
	s.Require().NotZero(resp.PositionId)

	_, err = s.msgServer.ForceLiquidation(sdk.WrapSDKContext(s.Ctx), &types.MsgForceLiquidation{
		Authority:   authority,
		PositionIds: []uint64{resp.PositionId},
		Signers:     s.signers(),
	})
	s.Require().NoError(err)

	position, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, resp.PositionId)
	s.Require().True(found)
	s.Require().Equal(types.PositionStatus_POSITION_STATUS_LIQUIDATED, position.Status)
}

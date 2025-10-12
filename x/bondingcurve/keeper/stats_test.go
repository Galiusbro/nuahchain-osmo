package keeper_test

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"

	keeper "github.com/osmosis-labs/osmosis/v30/x/bondingcurve/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
)

func (s *KeeperTestSuite) TestModuleStatsUpdates() {
	paymentAmount := "100.0"

	_, err := s.msgServer.BuyFromCurve(sdk.WrapSDKContext(s.Ctx), &types.MsgBuyFromCurve{
		Trader:        s.trader.String(),
		Denom:         s.denom,
		PaymentDenom:  s.quoteDenom,
		PaymentAmount: paymentAmount,
	})
	s.Require().NoError(err)

	genesis := s.App.BondingCurveKeeper.ExportGenesis(s.Ctx)
	s.Require().NotNil(genesis.ModuleStats)

	expected := osmomath.MustNewDecFromStr(paymentAmount)
	actualBuy := osmomath.MustNewDecFromStr(genesis.ModuleStats.TotalBuyVolume)
	s.Require().True(actualBuy.Equal(expected))

	found := false
	for _, ts := range genesis.TokenStats {
		if ts.Denom == s.denom {
			totalVolume := osmomath.MustNewDecFromStr(ts.TotalVolume)
			s.Require().True(totalVolume.Equal(expected))
			found = true
			break
		}
	}
	s.Require().True(found, "token stats for %s not found", s.denom)
}

func (s *KeeperTestSuite) TestMarginPositionsQuery() {
	s.setupMarginLiquidity(math.NewInt(1_000_000_000), osmomath.OneDec())

	_, err := s.msgServer.OpenMarginPosition(sdk.WrapSDKContext(s.Ctx), &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "1000.0",
		Leverage:         2,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
	})
	s.Require().NoError(err)

	querySrv := keeper.NewQueryServerImpl(*s.App.BondingCurveKeeper)
	resp, err := querySrv.MarginPositions(sdk.WrapSDKContext(s.Ctx), &types.QueryMarginPositionsRequest{
		Denom: s.denom,
		Limit: 10,
	})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Len(resp.Positions, 1)
	s.Require().NotNil(resp.Summary)
	s.Require().Equal(uint64(1), resp.Summary.OpenPositions)

	totalLong := osmomath.MustNewDecFromStr(resp.Summary.TotalLongExposure)
	s.Require().False(totalLong.IsZero())
}

func (s *KeeperTestSuite) TestLiquidationRecords() {
	s.setupMarginLiquidity(math.NewInt(2_000_000_000), osmomath.OneDec())

	openResp, err := s.msgServer.OpenMarginPosition(sdk.WrapSDKContext(s.Ctx), &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "1000.0",
		Leverage:         5,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
	})
	s.Require().NoError(err)

	position, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, openResp.PositionId)
	s.Require().True(found)

	liquidationPrice := position.LiquidationPriceDec()
	newPrice := liquidationPrice.Mul(osmomath.MustNewDecFromStr("0.90"))

	pool, poolFound := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(poolFound)
	pool.SetLastPrice(newPrice)
	s.App.BondingCurveKeeper.SetPool(s.Ctx, pool)

	marginPool, poolOK := s.App.BondingCurveKeeper.GetMarginPool(s.Ctx, s.denom)
	s.Require().True(poolOK)
	marginPool.SetLastMarkPrice(newPrice)
	marginPool.SetLastTwapPrice(newPrice)
	s.App.BondingCurveKeeper.SetMarginPool(s.Ctx, marginPool)

	err = s.App.BondingCurveKeeper.ProcessLiquidations(s.Ctx)
	s.Require().NoError(err)

	genesis := s.App.BondingCurveKeeper.ExportGenesis(s.Ctx)
	s.Require().NotNil(genesis.ModuleStats)
	s.Require().Greater(genesis.ModuleStats.TotalLiquidations, uint64(0))
	s.Require().NotEmpty(genesis.LiquidationRecords)

	record := genesis.LiquidationRecords[len(genesis.LiquidationRecords)-1]
	s.Require().Equal(openResp.PositionId, record.PositionId)
	s.Require().Equal(s.denom, record.Denom)
}

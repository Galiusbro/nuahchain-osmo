package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	keeper "github.com/osmosis-labs/osmosis/v30/x/leverage/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/leverage/types"
	leveragetypes "github.com/osmosis-labs/osmosis/v30/x/leverage/types"
	usertokenkeeper "github.com/osmosis-labs/osmosis/v30/x/usertoken/keeper"
	usertokentypes "github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

type FullFlowSuite struct {
	suite.Suite
	apptesting.KeeperTestHelper
}

func TestFullFlowSuite(t *testing.T) {
	suite.Run(t, new(FullFlowSuite))
}

func (s *FullFlowSuite) SetupTest() {
	s.KeeperTestHelper.Setup()
}

func (s *FullFlowSuite) TestFullFlowIntegration() {
	ctx := s.Ctx
	app := s.App

	params := app.UserTokenKeeper.GetParams(ctx)
	params.PlatformFeeWallet = s.TestAccs[1].String()
	params.ReferralWallet = s.TestAccs[2].String()
	params.AiCeoWallet = s.TestAccs[1].String()
	app.UserTokenKeeper.SetParams(ctx, params)

	userTokenSrv := usertokenkeeper.NewMsgServerImpl(*app.UserTokenKeeper, app.UserTokenKeeper.GetAuthority())
	leverageSrv := keeper.NewMsgServerImpl(*app.LeverageKeeper)

	creator := s.TestAccs[0]
	traderLong := s.TestAccs[1]
	traderShort := s.TestAccs[2]
	baseProvider := s.TestAccs[3]
	tokenProvider := s.TestAccs[4]

	createResp, err := userTokenSrv.CreateUserToken(ctx, &usertokentypes.MsgCreateUserToken{
		Creator:  creator.String(),
		Subdenom: "leverage-flow",
		Name:     "Leverage Flow Token",
		Symbol:   "LFT",
		Decimals: 6,
	})
	s.Require().NoError(err)
	userTokenDenom := createResp.Denom

    usertokenModuleAddr := app.AccountKeeper.GetModuleAddress(usertokentypes.ModuleName)
    s.Require().NotNil(usertokenModuleAddr)

    tokenSeed := sdk.NewCoin(userTokenDenom, math.NewInt(5_000_000_000))
    s.Require().NoError(app.BankKeeper.MintCoins(ctx, usertokentypes.ModuleName, sdk.NewCoins(tokenSeed)))
    s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, usertokentypes.ModuleName, tokenProvider, sdk.NewCoins(tokenSeed)))

	lendingKeeper := app.LeverageKeeper.GetLendingKeeper()
	s.Require().NoError(lendingKeeper.ProvideLiquidity(ctx, tokenProvider, tokenSeed))

	baseDenom := "unuah"
	baseLiquidity := sdk.NewCoin(baseDenom, math.NewInt(20_000_000_000))
	s.FundAcc(baseProvider, sdk.NewCoins(baseLiquidity))
	s.Require().NoError(lendingKeeper.ProvideLiquidity(ctx, baseProvider, baseLiquidity))

    baseReserve := sdk.NewCoin(baseDenom, math.NewInt(50_000_000_000))
    s.Require().NoError(app.BankKeeper.MintCoins(ctx, usertokentypes.ModuleName, sdk.NewCoins(baseReserve)))
    baseReserveBalance := app.BankKeeper.GetBalance(ctx, usertokenModuleAddr, baseDenom).Amount
    s.Require().True(baseReserveBalance.GTE(baseReserve.Amount), "usertoken module should hold base reserve")

	// Long position lifecycle
	longCollateral := math.NewInt(2_000_000_000)

	leverageParams := app.LeverageKeeper.GetParams(ctx)
	longLeverage := math.LegacyNewDec(2)

	longFee := math.LegacyNewDecFromInt(longCollateral).Mul(longLeverage).Mul(leverageParams.TradingFee).TruncateInt()
	longFunding := longCollateral.Add(longFee).Add(math.OneInt())
	s.FundAcc(traderLong, sdk.NewCoins(sdk.NewCoin(baseDenom, longFunding)))
	initialLongBalance := app.BankKeeper.GetBalance(ctx, traderLong, baseDenom).Amount

	longResp, err := leverageSrv.OpenPosition(ctx, &types.MsgOpenPosition{
		Trader:     traderLong.String(),
		TokenDenom: userTokenDenom,
		Collateral: sdk.NewCoin(baseDenom, longCollateral),
		Leverage:   longLeverage,
		Side:       types.PositionSideLong,
		MinPrice:   math.LegacyZeroDec(),
		MaxPrice:   math.LegacyNewDec(1_000_000_000),
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(longResp.PositionId)

	longPos, found := app.LeverageKeeper.GetPosition(ctx, longResp.PositionId)
	s.Require().True(found)
	s.Require().Equal(types.PositionSideLong, longPos.Side)
	s.Require().True(longPos.Size_.GT(math.ZeroInt()))

	borrowID, found := lendingKeeper.GetBorrowIDByLeveragePosition(ctx, longPos.Id)
	s.Require().True(found)
	borrowPos, found := lendingKeeper.GetBorrowPosition(ctx, borrowID)
	s.Require().True(found)
	s.Require().Equal(baseDenom, borrowPos.TokenDenom)
	s.Require().True(borrowPos.BorrowedAmount.GT(math.ZeroInt()))

	_, err = leverageSrv.ClosePosition(ctx, &types.MsgClosePosition{
		Trader:     traderLong.String(),
		PositionId: longPos.Id,
		MinPrice:   math.LegacyZeroDec(),
		MaxPrice:   math.LegacyNewDec(1_000_000_000),
	})
	s.Require().NoError(err)

	closedLong, found := app.LeverageKeeper.GetPosition(ctx, longPos.Id)
	s.Require().True(found)
	s.Require().Equal(types.PositionStatusClosed, closedLong.Status)
	_, found = lendingKeeper.GetBorrowIDByLeveragePosition(ctx, longPos.Id)
	s.Require().False(found)

	positionValue := math.LegacyNewDecFromInt(longCollateral).Mul(longLeverage)
	longTradingFee := positionValue.Mul(leverageParams.TradingFee).TruncateInt()
	finalLongBalance := app.BankKeeper.GetBalance(ctx, traderLong, baseDenom).Amount
	longDelta := absDiff(initialLongBalance, finalLongBalance.Add(longTradingFee))
	s.Require().True(longDelta.Equal(math.ZeroInt()), "unexpected long balance delta: %s", longDelta.String())

	leverageModuleAddr := app.AccountKeeper.GetModuleAddress(leveragetypes.ModuleName)
	leverTokenBalance := app.BankKeeper.GetBalance(ctx, leverageModuleAddr, userTokenDenom).Amount
	s.Require().True(leverTokenBalance.IsZero())

	// Short position lifecycle
	shortCollateral := math.NewInt(3_000_000_000)

	shortLeverage := math.LegacyNewDec(3)
	shortFee := math.LegacyNewDecFromInt(shortCollateral).Mul(shortLeverage).Mul(leverageParams.TradingFee).TruncateInt()
	shortFunding := shortCollateral.Add(shortFee).Add(math.OneInt())
	s.FundAcc(traderShort, sdk.NewCoins(sdk.NewCoin(baseDenom, shortFunding)))
	initialShortBalance := app.BankKeeper.GetBalance(ctx, traderShort, baseDenom).Amount

	shortResp, err := leverageSrv.OpenPosition(ctx, &types.MsgOpenPosition{
		Trader:     traderShort.String(),
		TokenDenom: userTokenDenom,
		Collateral: sdk.NewCoin(baseDenom, shortCollateral),
		Leverage:   shortLeverage,
		Side:       types.PositionSideShort,
		MinPrice:   math.LegacyZeroDec(),
		MaxPrice:   math.LegacyNewDec(1_000_000_000),
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(shortResp.PositionId)

	shortPos, found := app.LeverageKeeper.GetPosition(ctx, shortResp.PositionId)
	s.Require().True(found)
	s.Require().Equal(types.PositionSideShort, shortPos.Side)

	shortBorrowID, found := lendingKeeper.GetBorrowIDByLeveragePosition(ctx, shortPos.Id)
	s.Require().True(found)
	shortBorrowPos, found := lendingKeeper.GetBorrowPosition(ctx, shortBorrowID)
	s.Require().True(found)
	s.Require().Equal(userTokenDenom, shortBorrowPos.TokenDenom)
	s.Require().True(shortBorrowPos.BorrowedAmount.GT(math.ZeroInt()))

	_, err = leverageSrv.ClosePosition(ctx, &types.MsgClosePosition{
		Trader:     traderShort.String(),
		PositionId: shortPos.Id,
		MinPrice:   math.LegacyZeroDec(),
		MaxPrice:   math.LegacyNewDec(1_000_000_000),
	})
	s.Require().NoError(err)

	closedShort, found := app.LeverageKeeper.GetPosition(ctx, shortPos.Id)
	s.Require().True(found)
	s.Require().Equal(types.PositionStatusClosed, closedShort.Status)
	_, found = lendingKeeper.GetBorrowIDByLeveragePosition(ctx, shortPos.Id)
	s.Require().False(found)

	shortValue := math.LegacyNewDecFromInt(shortCollateral).Mul(shortLeverage)
	shortTradingFee := shortValue.Mul(leverageParams.TradingFee).TruncateInt()
	finalShortBalance := app.BankKeeper.GetBalance(ctx, traderShort, baseDenom).Amount
	shortDelta := absDiff(initialShortBalance, finalShortBalance.Add(shortTradingFee))
	s.Require().True(shortDelta.Equal(math.ZeroInt()), "unexpected short balance delta: %s", shortDelta.String())

	basePool, found := lendingKeeper.GetLendingPool(ctx, baseDenom)
	s.Require().True(found)
	s.Require().Equal(baseLiquidity.Amount, basePool.AvailableLiquidity)

	tokenPool, found := lendingKeeper.GetLendingPool(ctx, userTokenDenom)
	s.Require().True(found)
	s.Require().Equal(tokenSeed.Amount, tokenPool.AvailableLiquidity)

	moduleBaseBalance := app.BankKeeper.GetBalance(ctx, leverageModuleAddr, baseDenom).Amount
	expectedFees := longTradingFee.Add(shortTradingFee)
	s.Require().Equal(expectedFees, moduleBaseBalance)
}

func absDiff(a, b math.Int) math.Int {
	if a.GT(b) {
		return a.Sub(b)
	}
	return b.Sub(a)
}

package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	"github.com/osmosis-labs/osmosis/osmomath"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v30/app/params"
	bondingkeeper "github.com/osmosis-labs/osmosis/v30/x/bondingcurve/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
	usertokenkeeper "github.com/osmosis-labs/osmosis/v30/x/usertoken/keeper"
	usertokentypes "github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	msgServer       types.MsgServer
	userTokenServer usertokentypes.MsgServer
	creator         sdk.AccAddress
	trader          sdk.AccAddress
	denom           string
	quoteDenom      string
	curveWallet     sdk.AccAddress
	initialQuote    sdk.Coin
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.Ctx = s.Ctx.WithBlockTime(time.Unix(1_700_000_000, 0))

	// ensure we have enough accounts for wallet assignments
	for len(s.TestAccs) < 7 {
		extra := apptesting.CreateRandomAccounts(1)[0]
		s.TestAccs = append(s.TestAccs, extra)
	}

	s.creator = s.TestAccs[0]
	s.trader = s.TestAccs[1]
	s.quoteDenom = appparams.BaseCoinUnit
	s.curveWallet = s.TestAccs[5]

	bondingParams := s.App.BondingCurveKeeper.GetParams(s.Ctx)
	bondingParams.BondingCurveWallet = s.curveWallet.String()
	bondingParams.QuoteDenom = s.quoteDenom
	bondingParams.MaxSupply = "30000000.0"
	bondingParams.StartPrice = "0.0002"
	bondingParams.EndPrice = "1.0"
	s.App.BondingCurveKeeper.SetParams(s.Ctx, bondingParams)

	userParams := s.App.UserTokenKeeper.GetParams(s.Ctx)
	userParams.BondingCurveWallet = bondingParams.BondingCurveWallet
	userParams.PlatformWallet = s.TestAccs[2].String()
	userParams.ReferralWallet = s.TestAccs[3].String()
	userParams.AiCeoWallet = s.TestAccs[4].String()
	s.App.UserTokenKeeper.SetParams(s.Ctx, userParams)

	s.msgServer = bondingkeeper.NewMsgServerImpl(*s.App.BondingCurveKeeper)
	s.userTokenServer = usertokenkeeper.NewMsgServerImpl(*s.App.UserTokenKeeper)

	resp, err := s.userTokenServer.CreateToken(sdk.WrapSDKContext(s.Ctx), usertokentypes.NewMsgCreateToken(
		s.creator.String(),
		"TestToken",
		"TTT",
		"",
		"bonding curve token",
	))
	s.Require().NoError(err)
	s.denom = resp.Denom

	// Fund trader with quote asset
	s.FundAcc(s.trader, sdk.NewCoins(sdk.NewCoin(s.quoteDenom, sdkmath.NewInt(1_000_000_000_000))))
	s.initialQuote = s.App.BankKeeper.GetBalance(s.Ctx, s.trader, s.quoteDenom)

	// ensure bonding curve wallet holds the minted curve supply
	balance := s.App.BankKeeper.GetBalance(s.Ctx, s.curveWallet, s.denom)
	s.Require().True(balance.Amount.GT(sdkmath.ZeroInt()))

	// Fund curve wallet with quote tokens for DEX activation
	s.FundAcc(s.curveWallet, sdk.NewCoins(sdk.NewCoin(s.quoteDenom, sdkmath.NewInt(1_000_000_000_000))))

	// Fund curve wallet with more tokens for DEX activation (need 30M tokens)
	s.FundAcc(s.curveWallet, sdk.NewCoins(sdk.NewCoin(s.denom, sdkmath.NewInt(30_000_000))))
}

func (s *KeeperTestSuite) setupMarginLiquidity(available sdkmath.Int, price osmomath.Dec) {
	pool := types.BondingCurvePool{
		Denom:             s.denom,
		TokensSold:        "0",
		ReserveNuah:       "0",
		ReserveNdollar:    osmomath.NewDecFromInt(available).String(),
		LastPrice:         price.String(),
		DexPoolId:         0,
		DexActivated:      false,
		LiquidityProvider: "",
	}
	s.App.BondingCurveKeeper.SetPool(s.Ctx, pool)
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin(s.quoteDenom, available.MulRaw(10))))
}

func (s *KeeperTestSuite) TestBuyFromCurve() {
	payment := "100.0"
	resp, err := s.msgServer.BuyFromCurve(s.Ctx, &types.MsgBuyFromCurve{
		Trader:        s.trader.String(),
		Denom:         s.denom,
		PaymentDenom:  s.quoteDenom,
		PaymentAmount: payment,
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(resp.TokensOut)

	balance := s.App.BankKeeper.GetBalance(s.Ctx, s.trader, s.denom)
	s.Require().True(balance.Amount.GT(sdkmath.ZeroInt()))

	pool, found := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(found)
	sold := pool.TokensSoldDec()
	s.Require().True(sold.GT(osmomath.ZeroDec()))
}

func (s *KeeperTestSuite) TestSellToCurve() {
	// buy first
	_, err := s.msgServer.BuyFromCurve(s.Ctx, &types.MsgBuyFromCurve{
		Trader:        s.trader.String(),
		Denom:         s.denom,
		PaymentDenom:  s.quoteDenom,
		PaymentAmount: "200.0",
	})
	s.Require().NoError(err)

	balanceAfterBuy := s.App.BankKeeper.GetBalance(s.Ctx, s.trader, s.quoteDenom)
	poolBefore, found := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(found)
	soldBefore := poolBefore.TokensSoldDec()

	// determine half of balance to sell
	balance := s.App.BankKeeper.GetBalance(s.Ctx, s.trader, s.denom)
	sellAmount := balance.Amount.QuoRaw(2)
	s.Require().True(sellAmount.GT(sdkmath.ZeroInt()))

	_, err = s.msgServer.SellToCurve(s.Ctx, &types.MsgSellToCurve{
		Trader:       s.trader.String(),
		Denom:        s.denom,
		TokenAmount:  types.CoinToDec(sdk.NewCoin(s.denom, sellAmount)).String(),
		PaymentDenom: s.quoteDenom,
	})
	s.Require().NoError(err)

	newBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.trader, s.quoteDenom)
	s.Require().True(newBalance.Amount.GT(balanceAfterBuy.Amount))

	poolAfter, found := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(found)
	s.Require().True(poolAfter.TokensSoldDec().LT(soldBefore))
}

func (s *KeeperTestSuite) TestBuyExceedsSupply() {
	_, err := s.msgServer.BuyFromCurve(s.Ctx, &types.MsgBuyFromCurve{
		Trader:        s.trader.String(),
		Denom:         s.denom,
		PaymentDenom:  s.quoteDenom,
		PaymentAmount: "100000000.0",
	})
	s.Require().Error(err)
}
func (s *KeeperTestSuite) TestDexActivation() {
	params := s.App.BondingCurveKeeper.GetParams(s.Ctx)
	params.MaxSupply = "1.0"
	params.QuoteDenom = s.quoteDenom
	params.BondingCurveWallet = s.curveWallet.String()
	s.App.BondingCurveKeeper.SetParams(s.Ctx, params)

	token, found := s.App.UserTokenKeeper.GetToken(s.Ctx, s.denom)
	s.Require().True(found)
	token.Distribution.SetBondingCurveSupply(osmomath.NewInt(1))

	if token.Distribution.AiCeoWallet == "" {
		token.Distribution.AiCeoWallet = s.TestAccs[4].String()
	}
	if token.Distribution.PlatformWallet == "" {
		token.Distribution.PlatformWallet = s.curveWallet.String()
	}
	s.Require().NoError(s.App.UserTokenKeeper.UpdateToken(s.Ctx, token))

	provider := s.curveWallet
	s.FundAcc(provider, sdk.NewCoins(sdk.NewCoin(s.denom, sdkmath.NewInt(1_000_000_000_000))))
	s.FundAcc(provider, sdk.NewCoins(sdk.NewCoin(s.quoteDenom, sdkmath.NewInt(1_000_000_000_000))))

	pool := types.BondingCurvePool{
		Denom:          s.denom,
		TokensSold:     "1.0",
		ReserveNuah:    "0",
		ReserveNdollar: "0",
		LastPrice:      "0",
	}
	s.App.BondingCurveKeeper.SetPool(s.Ctx, pool)

	s.Require().NoError(s.App.BondingCurveKeeper.EndBlocker(s.Ctx))

	pool, found = s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(found)
	s.Require().True(pool.DexActivated)
	s.Require().NotZero(pool.DexPoolId)

	token, _ = s.App.UserTokenKeeper.GetToken(s.Ctx, s.denom)
	s.Require().True(token.State.DexTradingEnabled)
	s.Require().False(token.State.SoftLockEnabled)
}

func (s *KeeperTestSuite) TestOpenMarginPositionSuccess() {
	s.setupMarginLiquidity(sdkmath.NewInt(1_000_000), osmomath.OneDec())

	resp, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "100.0",
		Leverage:         5,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
		MinPositionSize:  "400.0",
	})
	s.Require().NoError(err)
	s.Require().EqualValues(1, resp.PositionId)

	position, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, resp.PositionId)
	s.Require().True(found)
	s.Require().Equal(s.trader.String(), position.Trader)
	s.Require().Equal(types.PositionType_POSITION_TYPE_LONG, position.Type)

	marginPool, found := s.App.BondingCurveKeeper.GetMarginPool(s.Ctx, s.denom)
	s.Require().True(found)
	collateralDec := osmomath.MustNewDecFromStr("100.0")
	positionSize := types.PositionSize(collateralDec, 5)
	s.Require().True(marginPool.TotalCollateralDec().Equal(collateralDec))
	s.Require().True(marginPool.TotalLongExposureDec().Equal(positionSize))
	s.Require().True(marginPool.AvailableLiquidityDec().Equal(osmomath.NewDecFromInt(sdkmath.NewInt(1_000_000)).Sub(positionSize)))
}

func (s *KeeperTestSuite) TestOpenMarginPositionInvalidLeverage() {
	s.setupMarginLiquidity(sdkmath.NewInt(100_000), osmomath.OneDec())

	_, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "50.0",
		Leverage:         types.MaxMarginLeverage + 1,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
	})
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrInvalidLeverage)
}

func (s *KeeperTestSuite) TestCloseMarginPositionRealizesPnL() {
	s.setupMarginLiquidity(sdkmath.NewInt(1_000_000), osmomath.OneDec())

	openResp, err := s.msgServer.OpenMarginPosition(sdk.WrapSDKContext(s.Ctx), &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "200.0",
		Leverage:         4,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
	})
	s.Require().NoError(err)

	pool, found := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(found)
	pool.LastPrice = osmomath.MustNewDecFromStr("1.25").String()
	s.App.BondingCurveKeeper.SetPool(s.Ctx, pool)

	closeResp, err := s.msgServer.CloseMarginPosition(sdk.WrapSDKContext(s.Ctx), &types.MsgCloseMarginPosition{
		Trader:     s.trader.String(),
		PositionId: openResp.PositionId,
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(closeResp.PayoutAmount)

	position, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, openResp.PositionId)
	s.Require().True(found)
	s.Require().Equal(types.PositionStatus_POSITION_STATUS_CLOSED, position.Status)
	s.Require().True(position.PositionSizeDec().IsZero())
	s.Require().True(position.CollateralAmountDec().IsZero())

	marginPool, found := s.App.BondingCurveKeeper.GetMarginPool(s.Ctx, s.denom)
	s.Require().True(found)
	s.Require().True(marginPool.TotalLongExposureDec().IsZero())
}

func (s *KeeperTestSuite) TestCloseMarginPositionMinPayout() {
	s.setupMarginLiquidity(sdkmath.NewInt(500_000), osmomath.OneDec())

	openResp, err := s.msgServer.OpenMarginPosition(sdk.WrapSDKContext(s.Ctx), &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "150.0",
		Leverage:         3,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
	})
	s.Require().NoError(err)

	_, err = s.msgServer.CloseMarginPosition(sdk.WrapSDKContext(s.Ctx), &types.MsgCloseMarginPosition{
		Trader:     s.trader.String(),
		PositionId: openResp.PositionId,
		MinPayout:  "1000.0",
	})
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrMinPaymentNotMet)
}

func (s *KeeperTestSuite) TestProcessLiquidationsFull() {
	s.setupMarginLiquidity(sdkmath.NewInt(2_000_000), osmomath.OneDec())

	openResp, err := s.msgServer.OpenMarginPosition(sdk.WrapSDKContext(s.Ctx), &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "500.0",
		Leverage:         20,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
	})
	s.Require().NoError(err)

	// Prime TWAP with stable price
	pool, found := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(found)
	pool.LastPrice = osmomath.OneDec().String()
	s.App.BondingCurveKeeper.SetPool(s.Ctx, pool)
	s.Require().NoError(s.App.BondingCurveKeeper.ProcessLiquidations(s.Ctx))

	// Drop price below liquidation threshold
	pool.LastPrice = osmomath.MustNewDecFromStr("0.40").String()
	s.App.BondingCurveKeeper.SetPool(s.Ctx, pool)
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	s.Require().NoError(s.App.BondingCurveKeeper.ProcessLiquidations(s.Ctx))
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	s.Require().NoError(s.App.BondingCurveKeeper.ProcessLiquidations(s.Ctx))

	position, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, openResp.PositionId)
	s.Require().True(found)
	s.Require().Equal(types.PositionStatus_POSITION_STATUS_LIQUIDATED, position.Status)
	s.Require().True(position.PositionSizeDec().IsZero())

	marginPool, found := s.App.BondingCurveKeeper.GetMarginPool(s.Ctx, s.denom)
	s.Require().True(found)
	s.Require().True(marginPool.TotalLongExposureDec().IsZero())
	s.Require().True(marginPool.TotalCollateralDec().IsZero())
	s.Require().True(marginPool.InsuranceFundDec().IsZero())
	s.Require().Equal(uint64(1), marginPool.TotalLiquidations)
	s.Require().True(marginPool.TotalBadDebtDec().GT(osmomath.ZeroDec()))
}

func (s *KeeperTestSuite) TestProcessLiquidationsPartial() {
	s.setupMarginLiquidity(sdkmath.NewInt(20_000_000), osmomath.OneDec())

	openResp, err := s.msgServer.OpenMarginPosition(sdk.WrapSDKContext(s.Ctx), &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "20000.0",
		Leverage:         60,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
	})
	s.Require().NoError(err)

	pool, found := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(found)
	pool.LastPrice = osmomath.OneDec().String()
	s.App.BondingCurveKeeper.SetPool(s.Ctx, pool)
	s.Require().NoError(s.App.BondingCurveKeeper.ProcessLiquidations(s.Ctx))

	pool.LastPrice = osmomath.MustNewDecFromStr("0.97").String()
	s.App.BondingCurveKeeper.SetPool(s.Ctx, pool)
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	s.Require().NoError(s.App.BondingCurveKeeper.ProcessLiquidations(s.Ctx))

	position, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, openResp.PositionId)
	s.Require().True(found)
	s.Require().Equal(types.PositionStatus_POSITION_STATUS_OPEN, position.Status)
	originalSize := types.PositionSize(osmomath.MustNewDecFromStr("20000.0"), 60)
	s.Require().True(position.PositionSizeDec().LT(originalSize))
	s.Require().True(position.PositionSizeDec().GT(osmomath.ZeroDec()))
	s.Require().True(position.CollateralAmountDec().LT(osmomath.MustNewDecFromStr("20000")))

	marginPool, found := s.App.BondingCurveKeeper.GetMarginPool(s.Ctx, s.denom)
	s.Require().True(found)
	s.Require().True(marginPool.TotalLongExposureDec().Equal(position.PositionSizeDec()))
	s.Require().True(marginPool.InsuranceFundDec().IsZero())
	s.Require().Equal(uint64(1), marginPool.TotalLiquidations)
	s.Require().True(marginPool.TotalCollateralDec().Equal(position.CollateralAmountDec()))
	s.Require().True(marginPool.TotalLiquidationFeesDec().GT(osmomath.ZeroDec()))
}

func (s *KeeperTestSuite) TestProcessLiquidationsCircuitBreaker() {
	s.setupMarginLiquidity(sdkmath.NewInt(5_000_000), osmomath.OneDec())

	openResp, err := s.msgServer.OpenMarginPosition(sdk.WrapSDKContext(s.Ctx), &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "1000.0",
		Leverage:         15,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
	})
	s.Require().NoError(err)

	pool, _ := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	pool.LastPrice = osmomath.OneDec().String()
	s.App.BondingCurveKeeper.SetPool(s.Ctx, pool)
	s.Require().NoError(s.App.BondingCurveKeeper.ProcessLiquidations(s.Ctx))

	pool.LastPrice = osmomath.MustNewDecFromStr("5.0").String()
	s.App.BondingCurveKeeper.SetPool(s.Ctx, pool)
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	s.Require().NoError(s.App.BondingCurveKeeper.ProcessLiquidations(s.Ctx))

	marginPool, found := s.App.BondingCurveKeeper.GetMarginPool(s.Ctx, s.denom)
	s.Require().True(found)
	s.Require().True(marginPool.LiquidationsPaused)

	position, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, openResp.PositionId)
	s.Require().True(found)
	s.Require().Equal(types.PositionStatus_POSITION_STATUS_OPEN, position.Status)

	pool.LastPrice = osmomath.MustNewDecFromStr("1.1").String()
	s.App.BondingCurveKeeper.SetPool(s.Ctx, pool)
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	s.Require().NoError(s.App.BondingCurveKeeper.ProcessLiquidations(s.Ctx))

	marginPool, _ = s.App.BondingCurveKeeper.GetMarginPool(s.Ctx, s.denom)
	s.Require().False(marginPool.LiquidationsPaused)
}

func (s *KeeperTestSuite) TestProcessLiquidationsStressBatching() {
	s.setupMarginLiquidity(sdkmath.NewInt(50_000_000), osmomath.OneDec())

	var positionIDs []uint64
	for i := 0; i < 60; i++ {
		collateral := fmt.Sprintf("%d.0", 500+i*5)
		resp, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
			Trader:           s.trader.String(),
			Denom:            s.denom,
			CollateralDenom:  s.quoteDenom,
			CollateralAmount: collateral,
			Leverage:         25,
			PositionType:     types.PositionType_POSITION_TYPE_LONG,
		})
		s.Require().NoError(err)
		positionIDs = append(positionIDs, resp.PositionId)
	}

	pool, _ := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	pool.LastPrice = osmomath.MustNewDecFromStr("0.9").String()
	s.App.BondingCurveKeeper.SetPool(s.Ctx, pool)
	s.Require().NoError(s.App.BondingCurveKeeper.ProcessLiquidations(s.Ctx))

	pool.LastPrice = osmomath.MustNewDecFromStr("0.4").String()
	s.App.BondingCurveKeeper.SetPool(s.Ctx, pool)
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	s.Require().NoError(s.App.BondingCurveKeeper.ProcessLiquidations(s.Ctx))

	liquidatedFirst := 0
	for _, id := range positionIDs {
		pos, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, id)
		s.Require().True(found)
		if pos.Status == types.PositionStatus_POSITION_STATUS_LIQUIDATED {
			liquidatedFirst++
		}
	}
	s.Require().True(liquidatedFirst < len(positionIDs))

	for i := 0; i < 3; i++ {
		s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
		s.Require().NoError(s.App.BondingCurveKeeper.ProcessLiquidations(s.Ctx))
	}

	finalLiquidated := 0
	for _, id := range positionIDs {
		pos, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, id)
		s.Require().True(found)
		if pos.Status == types.PositionStatus_POSITION_STATUS_LIQUIDATED {
			finalLiquidated++
		}
	}
	s.Require().Equal(len(positionIDs), finalLiquidated)
}

func (s *KeeperTestSuite) TestBuyAllTokensFromCurve() {
	// Test buying a large amount of tokens to test price progression
	// We'll buy in multiple transactions to avoid hitting limits
	params := s.App.BondingCurveKeeper.GetParams(s.Ctx)
	maxSupply := params.MaxSupplyDec()
	startPrice := params.StartPriceDec()

	// Buy tokens in multiple transactions to test price progression
	totalTokensBought := osmomath.ZeroDec()
	paymentAmount := "1000000.0" // 1 million per transaction (reasonable amount)

	// Try to buy tokens in multiple transactions
	for i := 0; i < 20; i++ { // Try up to 20 transactions
		// Get current pool state before transaction
		pool, found := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
		var currentPrice, currentSold osmomath.Dec

		if found {
			currentPrice = pool.LastPriceDec()
			currentSold = pool.TokensSoldDec()
		} else {
			// Pool doesn't exist yet, will be created on first transaction
			// Use the start price from parameters
			currentPrice = params.StartPriceDec()
			currentSold = osmomath.ZeroDec()
		}

		fmt.Printf("Transaction %d: Starting with price=%s, sold=%s\n",
			i+1, currentPrice.String(), currentSold.String())

		resp, err := s.msgServer.BuyFromCurve(s.Ctx, &types.MsgBuyFromCurve{
			Trader:        s.trader.String(),
			Denom:         s.denom,
			PaymentDenom:  s.quoteDenom,
			PaymentAmount: paymentAmount,
		})

		if err != nil {
			// If we hit max supply, that's expected
			if err.Error() == "max bonding curve supply reached" {
				fmt.Printf("Transaction %d: Hit max supply limit\n", i+1)
				break
			}
			// If it's another error, fail the test
			s.Require().NoError(err)
		}

		if resp != nil {
			s.Require().NotEmpty(resp.TokensOut)

			tokensOut, err := osmomath.NewDecFromStr(resp.TokensOut)
			s.Require().NoError(err)
			totalTokensBought = totalTokensBought.Add(tokensOut)

			// Get updated pool state after transaction
			pool, found = s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
			s.Require().True(found)
			newPrice := pool.LastPriceDec()
			newSold := pool.TokensSoldDec()

			fmt.Printf("Transaction %d: Bought %s tokens for %s, new price=%s, total sold=%s\n",
				i+1, tokensOut.String(), paymentAmount, newPrice.String(), newSold.String())

			// Calculate price change
			priceChange := newPrice.Sub(currentPrice)
			if currentPrice.GT(osmomath.ZeroDec()) {
				priceChangePercent := priceChange.Quo(currentPrice).Mul(osmomath.NewDec(100))
				fmt.Printf("Transaction %d: Price change: %s (%s%%)\n",
					i+1, priceChange.String(), priceChangePercent.String())
			} else {
				fmt.Printf("Transaction %d: Price change: %s (new price from zero)\n",
					i+1, priceChange.String())
			}
		}

		// Check if we're close to max supply
		if totalTokensBought.GTE(maxSupply.Mul(osmomath.MustNewDecFromStr("0.95"))) {
			fmt.Printf("Transaction %d: Reached 95%% of max supply, stopping\n", i+1)
			break
		}
	}

	// Verify we bought a significant amount of tokens
	s.Require().True(totalTokensBought.GT(osmomath.ZeroDec()), "Should have bought some tokens")

	// Check the price progression
	pool, found := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(found)
	finalPrice := pool.LastPriceDec()
	finalSold := pool.TokensSoldDec()

	// Print final statistics
	fmt.Printf("\n=== FINAL STATISTICS ===\n")
	fmt.Printf("Total tokens bought: %s\n", totalTokensBought.String())
	fmt.Printf("Max supply: %s\n", maxSupply.String())
	fmt.Printf("Final price: %s\n", finalPrice.String())
	fmt.Printf("Start price: %s\n", startPrice.String())
	fmt.Printf("Final sold: %s\n", finalSold.String())

	// Calculate final metrics
	supplyPercentage := totalTokensBought.Quo(maxSupply)
	priceIncrease := finalPrice.Quo(startPrice)
	priceIncreasePercent := priceIncrease.Sub(osmomath.OneDec()).Mul(osmomath.NewDec(100))

	fmt.Printf("Supply percentage: %s%%\n", supplyPercentage.Mul(osmomath.NewDec(100)).String())
	fmt.Printf("Price increase: %sx (%s%%)\n", priceIncrease.String(), priceIncreasePercent.String())
	fmt.Printf("========================\n\n")

	// Final price should be significantly higher than start price
	s.Require().True(finalPrice.GT(startPrice),
		fmt.Sprintf("Final price %s should be higher than start price %s", finalPrice.String(), startPrice.String()))

	// Price should have increased significantly
	s.Require().True(priceIncrease.GTE(osmomath.MustNewDecFromStr("5")),
		fmt.Sprintf("Price should have increased at least 5x. Start: %s, Final: %s, Increase: %s",
			startPrice.String(), finalPrice.String(), priceIncrease.String()))

	// Verify we bought a reasonable percentage of max supply
	s.Require().True(supplyPercentage.GTE(osmomath.MustNewDecFromStr("0.01")),
		fmt.Sprintf("Should have bought at least 1%% of max supply. Bought: %s, Max: %s, Percentage: %s",
			totalTokensBought.String(), maxSupply.String(), supplyPercentage.String()))

	// If we bought a significant amount, check if DEX activation should be triggered
	if totalTokensBought.GTE(maxSupply.Mul(osmomath.MustNewDecFromStr("0.99"))) {
		s.Require().NoError(s.App.BondingCurveKeeper.EndBlocker(s.Ctx))

		pool, found = s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
		s.Require().True(found)
		s.Require().True(pool.DexActivated)

		// Verify token state changes
		token, found := s.App.UserTokenKeeper.GetToken(s.Ctx, s.denom)
		s.Require().True(found)
		s.Require().True(token.State.DexTradingEnabled)
		s.Require().False(token.State.SoftLockEnabled)
	}
}

func (s *KeeperTestSuite) TestDexActivationAndTrading() {
	// Test DEX activation before and after curve completion
	// Also test buy/sell operations with detailed logging
	params := s.App.BondingCurveKeeper.GetParams(s.Ctx)
	maxSupply := params.MaxSupplyDec()
	startPrice := params.StartPriceDec()
	endPrice := params.EndPriceDec()

	fmt.Printf("\n=== DEX ACTIVATION AND TRADING TEST ===\n")
	fmt.Printf("Max supply: %s\n", maxSupply.String())
	fmt.Printf("Start price: %s\n", startPrice.String())
	fmt.Printf("End price: %s\n", endPrice.String())

	// Check initial DEX state
	token, found := s.App.UserTokenKeeper.GetToken(s.Ctx, s.denom)
	s.Require().True(found)
	fmt.Printf("Initial DEX state: DexTradingEnabled=%t, SoftLockEnabled=%t\n",
		token.State.DexTradingEnabled, token.State.SoftLockEnabled)

	// Phase 1: Buy tokens to activate DEX (reach 99% of max supply)
	fmt.Printf("\n--- PHASE 1: ACTIVATING DEX ---\n")
	totalTokensBought := osmomath.ZeroDec()
	paymentAmount := "1000000.0" // 1 million per transaction

	for i := 0; i < 20; i++ {
		pool, found := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
		var currentPrice, currentSold osmomath.Dec

		if found {
			currentPrice = pool.LastPriceDec()
			currentSold = pool.TokensSoldDec()
		} else {
			currentPrice = params.StartPriceDec()
			currentSold = osmomath.ZeroDec()
		}

		fmt.Printf("Buy %d: Starting with price=%s, sold=%s\n",
			i+1, currentPrice.String(), currentSold.String())

		resp, err := s.msgServer.BuyFromCurve(s.Ctx, &types.MsgBuyFromCurve{
			Trader:        s.trader.String(),
			Denom:         s.denom,
			PaymentDenom:  s.quoteDenom,
			PaymentAmount: paymentAmount,
		})

		if err != nil {
			if err.Error() == "max bonding curve supply reached" {
				fmt.Printf("Buy %d: Hit max supply limit\n", i+1)
				break
			}
			s.Require().NoError(err)
		}

		if resp != nil {
			s.Require().NotEmpty(resp.TokensOut)

			tokensOut, err := osmomath.NewDecFromStr(resp.TokensOut)
			s.Require().NoError(err)
			totalTokensBought = totalTokensBought.Add(tokensOut)

			// Get updated pool state
			pool, found = s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
			s.Require().True(found)
			newPrice := pool.LastPriceDec()
			newSold := pool.TokensSoldDec()

			fmt.Printf("Buy %d: Bought %s tokens, new price=%s, total sold=%s\n",
				i+1, tokensOut.String(), newPrice.String(), newSold.String())
		}

		// Check if we're close to max supply (85% for DEX activation)
		if totalTokensBought.GTE(maxSupply.Mul(osmomath.MustNewDecFromStr("0.85"))) {
			fmt.Printf("Buy %d: Reached 85%% of max supply, stopping\n", i+1)
			break
		}
	}

	// Check DEX state after buying
	fmt.Printf("\n--- CHECKING DEX STATE AFTER BUYING ---\n")
	pool, found := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(found)
	fmt.Printf("Pool DEX activated: %t\n", pool.DexActivated)

	token, found = s.App.UserTokenKeeper.GetToken(s.Ctx, s.denom)
	s.Require().True(found)
	fmt.Printf("Token DEX state: DexTradingEnabled=%t, SoftLockEnabled=%t\n",
		token.State.DexTradingEnabled, token.State.SoftLockEnabled)

	// Check curve wallet balance before DEX activation
	fmt.Printf("\n--- CHECKING CURVE WALLET BALANCE ---\n")
	curveWalletBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.curveWallet, s.denom)
	curveWalletQuoteBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.curveWallet, s.quoteDenom)
	fmt.Printf("Curve wallet %s balance: %s\n", s.denom, curveWalletBalance.Amount.String())
	fmt.Printf("Curve wallet %s balance: %s\n", s.quoteDenom, curveWalletQuoteBalance.Amount.String())

	// Trigger DEX activation
	fmt.Printf("\n--- TRIGGERING DEX ACTIVATION ---\n")
	err := s.App.BondingCurveKeeper.EndBlocker(s.Ctx)
	if err != nil {
		fmt.Printf("DEX activation error: %s\n", err.Error())
	} else {
		fmt.Printf("DEX activation successful!\n")
	}

	// Check DEX state after activation
	pool, found = s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(found)
	fmt.Printf("Pool DEX activated after EndBlocker: %t\n", pool.DexActivated)

	token, found = s.App.UserTokenKeeper.GetToken(s.Ctx, s.denom)
	s.Require().True(found)
	fmt.Printf("Token DEX state after EndBlocker: DexTradingEnabled=%t, SoftLockEnabled=%t\n",
		token.State.DexTradingEnabled, token.State.SoftLockEnabled)

	// Phase 2: Test selling tokens back to curve
	fmt.Printf("\n--- PHASE 2: TESTING SELL OPERATIONS ---\n")

	// Check trader's token balance
	traderBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.trader, s.denom)
	fmt.Printf("Trader token balance: %s\n", traderBalance.Amount.String())

	// Sell some tokens back
	sellAmount := "1000000.0" // 1 million tokens
	fmt.Printf("Selling %s tokens back to curve...\n", sellAmount)

	sellResp, err := s.msgServer.SellToCurve(s.Ctx, &types.MsgSellToCurve{
		Trader:        s.trader.String(),
		Denom:         s.denom,
		TokenAmount:   sellAmount,
		PaymentDenom:  s.quoteDenom,
		MinPaymentOut: "0",
	})

	if err != nil {
		fmt.Printf("Sell error: %s\n", err.Error())
	} else {
		s.Require().NotEmpty(sellResp.PaymentOut)
		fmt.Printf("Sold %s tokens for %s %s\n",
			sellAmount, sellResp.PaymentOut, s.quoteDenom)

		// Check updated pool state
		pool, found = s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
		s.Require().True(found)
		fmt.Printf("Pool after sell: price=%s, sold=%s\n",
			pool.LastPriceDec().String(), pool.TokensSoldDec().String())
	}

	// Phase 3: Test buying more tokens after DEX activation
	fmt.Printf("\n--- PHASE 3: TESTING BUY AFTER DEX ACTIVATION ---\n")

	// Check if curve is still active after DEX activation
	token, found = s.App.UserTokenKeeper.GetToken(s.Ctx, s.denom)
	s.Require().True(found)
	fmt.Printf("Token state after DEX activation: CurveCompleted=%t, DexTradingEnabled=%t\n",
		token.State.CurveCompleted, token.State.DexTradingEnabled)

	// Try to buy more tokens from curve (should still work!)
	buyResp, err := s.msgServer.BuyFromCurve(s.Ctx, &types.MsgBuyFromCurve{
		Trader:        s.trader.String(),
		Denom:         s.denom,
		PaymentDenom:  s.quoteDenom,
		PaymentAmount: "500000.0", // 500k payment
	})

	if err != nil {
		fmt.Printf("Buy after DEX activation error: %s\n", err.Error())
	} else {
		s.Require().NotEmpty(buyResp.TokensOut)
		fmt.Printf("Bought %s tokens for 500000.0 after DEX activation\n", buyResp.TokensOut)

		// Check final pool state
		pool, found = s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
		s.Require().True(found)
		fmt.Printf("Final pool state: price=%s, sold=%s, DEX activated=%t\n",
			pool.LastPriceDec().String(), pool.TokensSoldDec().String(), pool.DexActivated)
	}

	// Phase 4: Test trading through DEX pool
	fmt.Printf("\n--- PHASE 4: TESTING DEX POOL TRADING ---\n")

	// Get DEX pool ID
	pool, found = s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(found)
	fmt.Printf("DEX Pool ID: %d\n", pool.DexPoolId)

	if pool.DexPoolId > 0 {
		// Try to get pool info from Osmosis
		poolInfo, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, pool.DexPoolId)
		if err != nil {
			fmt.Printf("Error getting DEX pool info: %s\n", err.Error())
		} else {
			fmt.Printf("DEX Pool found: ID=%d, Type=%T\n", poolInfo.GetId(), poolInfo)

			// Try to perform a swap through the DEX pool
			// This would require implementing a swap message, which is complex
			// For now, we'll just verify the pool exists and is accessible
			fmt.Printf("DEX Pool is accessible and ready for trading!\n")
		}
	} else {
		fmt.Printf("DEX Pool ID is 0 - pool not created\n")
	}

	// Phase 5: Test margin trading after DEX activation
	fmt.Printf("\n--- PHASE 5: TESTING MARGIN TRADING AFTER DEX ACTIVATION ---\n")

	// Check if margin trading is available
	fmt.Printf("Testing margin trading with leverage after DEX activation...\n")

	// Get current pool state before opening margin position
	pool, found = s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(found)
	fmt.Printf("Pool state before margin position: price=%s, sold=%s, DEX activated=%t\n",
		pool.LastPriceDec().String(), pool.TokensSoldDec().String(), pool.DexActivated)

	// Try to open a margin position
	marginResp, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "1000.0", // 1000 collateral
		Leverage:         10,       // 10x leverage
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
		MinPositionSize:  "0",
	})

	if err != nil {
		fmt.Printf("Margin trading error: %s\n", err.Error())
	} else {
		fmt.Printf("Margin position opened successfully! Position ID: %d\n", marginResp.PositionId)

		// Check position details
		position, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, marginResp.PositionId)
		if found {
			fmt.Printf("Position details: Trader=%s, Type=%s, Leverage=%d\n",
				position.Trader, position.Type.String(), position.Leverage)
			fmt.Printf("Position entry price: %s\n", position.EntryPrice)
			fmt.Printf("Position liquidation price: %s\n", position.LiquidationPrice)
		}

		// Check margin pool
		marginPool, found := s.App.BondingCurveKeeper.GetMarginPool(s.Ctx, s.denom)
		if found {
			fmt.Printf("Margin pool: Available=%s, TotalCollateral=%s\n",
				marginPool.AvailableLiquidity, marginPool.TotalCollateral)
		}

		// CRITICAL: Check where the price comes from
		fmt.Printf("\n--- PRICE SOURCE ANALYSIS ---\n")
		fmt.Printf("Pool LastPrice: %s\n", pool.LastPriceDec().String())
		fmt.Printf("Pool TokensSold: %s\n", pool.TokensSoldDec().String())
		fmt.Printf("DEX Pool ID: %d\n", pool.DexPoolId)
		fmt.Printf("DEX Activated: %t\n", pool.DexActivated)

		// Check if DEX pool price is different
		if pool.DexPoolId > 0 {
			dexPoolInfo, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, pool.DexPoolId)
			if err == nil {
				fmt.Printf("DEX Pool exists: ID=%d, Type=%T\n", dexPoolInfo.GetId(), dexPoolInfo)
				fmt.Printf("MARGIN TRADING USES BONDING CURVE PRICE, NOT DEX PRICE!\n")
			}
		}
	}

	// Final DEX state check
	token, found = s.App.UserTokenKeeper.GetToken(s.Ctx, s.denom)
	s.Require().True(found)
	fmt.Printf("Final token state: DexTradingEnabled=%t, SoftLockEnabled=%t\n",
		token.State.DexTradingEnabled, token.State.SoftLockEnabled)

	fmt.Printf("=== END DEX ACTIVATION AND TRADING TEST ===\n\n")
}

func (s *KeeperTestSuite) TestMarginTradingWithDexPool() {
	// Test margin trading using DEX pool price source
	fmt.Printf("\n=== MARGIN TRADING WITH DEX POOL TEST ===\n")

	// Setup: Activate DEX first
	params := s.App.BondingCurveKeeper.GetParams(s.Ctx)
	maxSupply := params.MaxSupplyDec()

	// Buy tokens to activate DEX (reach 85% of max supply)
	fmt.Printf("Activating DEX by buying tokens...\n")
	totalTokensBought := osmomath.ZeroDec()
	paymentAmount := "1000000.0"

	for i := 0; i < 20; i++ {
		buyResp, err := s.msgServer.BuyFromCurve(s.Ctx, &types.MsgBuyFromCurve{
			Trader:        s.trader.String(),
			Denom:         s.denom,
			PaymentDenom:  s.quoteDenom,
			PaymentAmount: paymentAmount,
		})

		if err != nil {
			fmt.Printf("Buy error at transaction %d: %s\n", i+1, err.Error())
			break
		}

		tokensOut, _ := osmomath.NewDecFromStr(buyResp.TokensOut)
		totalTokensBought = totalTokensBought.Add(tokensOut)

		// Check if we've reached 85% of max supply
		if totalTokensBought.GTE(maxSupply.Mul(osmomath.MustNewDecFromStr("0.85"))) {
			fmt.Printf("Reached 85%% of max supply, stopping\n")
			break
		}
	}

	// Trigger DEX activation
	fmt.Printf("Triggering DEX activation...\n")
	s.Require().NoError(s.App.BondingCurveKeeper.EndBlocker(s.Ctx))

	// Verify DEX is activated
	pool, found := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(found)
	s.Require().True(pool.DexActivated)
	s.Require().NotZero(pool.DexPoolId)

	fmt.Printf("DEX activated! Pool ID: %d\n", pool.DexPoolId)

	// Test 1: Open margin position with bonding curve price
	fmt.Printf("\n--- TEST 1: MARGIN POSITION WITH BONDING CURVE PRICE ---\n")

	bondingCurveResp, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "1000.0",
		Leverage:         5,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
		MinPositionSize:  "0",
		PriceSource:      types.PriceSource_PRICE_SOURCE_BONDING_CURVE,
	})

	s.Require().NoError(err)
	fmt.Printf("Bonding curve position opened: ID=%d\n", bondingCurveResp.PositionId)

	// Test 2: Open margin position with DEX pool price
	fmt.Printf("\n--- TEST 2: MARGIN POSITION WITH DEX POOL PRICE ---\n")

	dexResp, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "1000.0",
		Leverage:         5,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
		MinPositionSize:  "0",
		PriceSource:      types.PriceSource_PRICE_SOURCE_DEX_POOL,
	})

	if err != nil {
		fmt.Printf("DEX margin position error: %s\n", err.Error())
	} else {
		fmt.Printf("DEX position opened: ID=%d\n", dexResp.PositionId)

		// Compare prices
		bondingPosition, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, bondingCurveResp.PositionId)
		s.Require().True(found)

		dexPosition, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, dexResp.PositionId)
		s.Require().True(found)

		fmt.Printf("Bonding curve entry price: %s\n", bondingPosition.EntryPrice)
		fmt.Printf("DEX pool entry price: %s\n", dexPosition.EntryPrice)
		fmt.Printf("Price source comparison: Bonding=%s, DEX=%s\n",
			bondingPosition.PriceSource.String(), dexPosition.PriceSource.String())
	}

	fmt.Printf("=== END MARGIN TRADING WITH DEX POOL TEST ===\n\n")
}

func (s *KeeperTestSuite) TestDexMarginTradingComprehensive() {
	// Comprehensive test for DEX margin trading - similar to bonding curve tests
	fmt.Printf("\n=== COMPREHENSIVE DEX MARGIN TRADING TEST ===\n")

	// Setup: Activate DEX first
	params := s.App.BondingCurveKeeper.GetParams(s.Ctx)
	maxSupply := params.MaxSupplyDec()

	// Buy tokens to activate DEX (reach 85% of max supply)
	fmt.Printf("Activating DEX by buying tokens...\n")
	totalTokensBought := osmomath.ZeroDec()
	paymentAmount := "1000000.0"

	for i := 0; i < 20; i++ {
		buyResp, err := s.msgServer.BuyFromCurve(s.Ctx, &types.MsgBuyFromCurve{
			Trader:        s.trader.String(),
			Denom:         s.denom,
			PaymentDenom:  s.quoteDenom,
			PaymentAmount: paymentAmount,
		})

		if err != nil {
			fmt.Printf("Buy error at transaction %d: %s\n", i+1, err.Error())
			break
		}

		tokensOut, _ := osmomath.NewDecFromStr(buyResp.TokensOut)
		totalTokensBought = totalTokensBought.Add(tokensOut)

		// Check if we've reached 85% of max supply
		if totalTokensBought.GTE(maxSupply.Mul(osmomath.MustNewDecFromStr("0.85"))) {
			fmt.Printf("Reached 85%% of max supply, stopping\n")
			break
		}
	}

	// Trigger DEX activation
	fmt.Printf("Triggering DEX activation...\n")
	s.Require().NoError(s.App.BondingCurveKeeper.EndBlocker(s.Ctx))

	// Verify DEX is activated
	pool, found := s.App.BondingCurveKeeper.GetPool(s.Ctx, s.denom)
	s.Require().True(found)
	s.Require().True(pool.DexActivated)
	s.Require().NotZero(pool.DexPoolId)

	fmt.Printf("DEX activated! Pool ID: %d\n", pool.DexPoolId)

	// Test 1: Open DEX margin position with different leverage levels
	fmt.Printf("\n--- TEST 1: DEX MARGIN POSITIONS WITH DIFFERENT LEVERAGE ---\n")

	// Test 1x leverage (minimum)
	fmt.Printf("Testing 1x leverage (minimum)...\n")
	minLeverageResp, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "500.0",
		Leverage:         types.MinMarginLeverage, // 1x
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
		PriceSource:      types.PriceSource_PRICE_SOURCE_DEX_POOL,
	})
	s.Require().NoError(err)
	fmt.Printf("1x leverage position opened: ID=%d\n", minLeverageResp.PositionId)

	// Test 10x leverage (medium)
	fmt.Printf("Testing 10x leverage (medium)...\n")
	mediumLeverageResp, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "1000.0",
		Leverage:         10,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
		PriceSource:      types.PriceSource_PRICE_SOURCE_DEX_POOL,
	})
	s.Require().NoError(err)
	fmt.Printf("10x leverage position opened: ID=%d\n", mediumLeverageResp.PositionId)

	// Test 100x leverage (maximum)
	fmt.Printf("Testing 100x leverage (maximum)...\n")
	maxLeverageResp, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "2000.0",
		Leverage:         types.MaxMarginLeverage, // 100x
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
		PriceSource:      types.PriceSource_PRICE_SOURCE_DEX_POOL,
	})
	s.Require().NoError(err)
	fmt.Printf("100x leverage position opened: ID=%d\n", maxLeverageResp.PositionId)

	// Test 2: Verify position details and price sources
	fmt.Printf("\n--- TEST 2: VERIFY POSITION DETAILS ---\n")

	// Check 1x position
	minPosition, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, minLeverageResp.PositionId)
	s.Require().True(found)
	s.Require().Equal(types.PriceSource_PRICE_SOURCE_DEX_POOL, minPosition.PriceSource)
	s.Require().Equal(types.MinMarginLeverage, minPosition.Leverage)
	fmt.Printf("1x position: PriceSource=%s, Leverage=%d, EntryPrice=%s\n",
		minPosition.PriceSource.String(), minPosition.Leverage, minPosition.EntryPrice)

	// Check 10x position
	mediumPosition, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, mediumLeverageResp.PositionId)
	s.Require().True(found)
	s.Require().Equal(types.PriceSource_PRICE_SOURCE_DEX_POOL, mediumPosition.PriceSource)
	s.Require().Equal(uint32(10), mediumPosition.Leverage)
	fmt.Printf("10x position: PriceSource=%s, Leverage=%d, EntryPrice=%s\n",
		mediumPosition.PriceSource.String(), mediumPosition.Leverage, mediumPosition.EntryPrice)

	// Check 100x position
	maxPosition, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, maxLeverageResp.PositionId)
	s.Require().True(found)
	s.Require().Equal(types.PriceSource_PRICE_SOURCE_DEX_POOL, maxPosition.PriceSource)
	s.Require().Equal(types.MaxMarginLeverage, maxPosition.Leverage)
	fmt.Printf("100x position: PriceSource=%s, Leverage=%d, EntryPrice=%s\n",
		maxPosition.PriceSource.String(), maxPosition.Leverage, maxPosition.EntryPrice)

	// Test 3: Test closing DEX margin positions
	fmt.Printf("\n--- TEST 3: CLOSING DEX MARGIN POSITIONS ---\n")

	// Close 1x position
	fmt.Printf("Closing 1x leverage position...\n")
	closeResp1, err := s.msgServer.CloseMarginPosition(s.Ctx, &types.MsgCloseMarginPosition{
		Trader:     s.trader.String(),
		PositionId: minLeverageResp.PositionId,
	})
	s.Require().NoError(err)
	fmt.Printf("1x position closed: Payout=%s, PnL=%s\n",
		closeResp1.PayoutAmount, closeResp1.RealizedPnl)

	// Close 10x position
	fmt.Printf("Closing 10x leverage position...\n")
	closeResp2, err := s.msgServer.CloseMarginPosition(s.Ctx, &types.MsgCloseMarginPosition{
		Trader:     s.trader.String(),
		PositionId: mediumLeverageResp.PositionId,
	})
	s.Require().NoError(err)
	fmt.Printf("10x position closed: Payout=%s, PnL=%s\n",
		closeResp2.PayoutAmount, closeResp2.RealizedPnl)

	// Test 4: Test invalid leverage for DEX
	fmt.Printf("\n--- TEST 4: TESTING INVALID LEVERAGE FOR DEX ---\n")

	// Test leverage below minimum
	_, err = s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "100.0",
		Leverage:         0, // Below minimum
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
		PriceSource:      types.PriceSource_PRICE_SOURCE_DEX_POOL,
	})
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrInvalidLeverage)
	fmt.Printf("Below minimum leverage correctly rejected\n")

	// Test leverage above maximum
	_, err = s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "100.0",
		Leverage:         types.MaxMarginLeverage + 1, // Above maximum
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
		PriceSource:      types.PriceSource_PRICE_SOURCE_DEX_POOL,
	})
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrInvalidLeverage)
	fmt.Printf("Above maximum leverage correctly rejected\n")

	// Test 5: Compare DEX vs Bonding Curve prices
	fmt.Printf("\n--- TEST 5: COMPARING DEX VS BONDING CURVE PRICES ---\n")

	// Open positions with both price sources
	bondingResp, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "1000.0",
		Leverage:         5,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
		PriceSource:      types.PriceSource_PRICE_SOURCE_BONDING_CURVE,
	})
	s.Require().NoError(err)

	dexResp, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "1000.0",
		Leverage:         5,
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
		PriceSource:      types.PriceSource_PRICE_SOURCE_DEX_POOL,
	})
	s.Require().NoError(err)

	// Compare prices
	bondingPos, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, bondingResp.PositionId)
	s.Require().True(found)

	dexPos, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, dexResp.PositionId)
	s.Require().True(found)

	fmt.Printf("Price comparison:\n")
	fmt.Printf("  Bonding Curve: %s (source: %s)\n",
		bondingPos.EntryPrice, bondingPos.PriceSource.String())
	fmt.Printf("  DEX Pool: %s (source: %s)\n",
		dexPos.EntryPrice, dexPos.PriceSource.String())

	// Calculate price difference
	bondingPrice, _ := osmomath.NewDecFromStr(bondingPos.EntryPrice)
	dexPrice, _ := osmomath.NewDecFromStr(dexPos.EntryPrice)
	priceDiff := dexPrice.Sub(bondingPrice)
	priceDiffPercent := priceDiff.Quo(bondingPrice).Mul(osmomath.MustNewDecFromStr("100"))

	fmt.Printf("  Price difference: %s (%s%%)\n",
		priceDiff.String(), priceDiffPercent.String())

	fmt.Printf("=== END COMPREHENSIVE DEX MARGIN TRADING TEST ===\n\n")
}

func (s *KeeperTestSuite) TestMaxLeverage100x() {
	s.setupMarginLiquidity(sdkmath.NewInt(10_000_000), osmomath.OneDec())

	// Test maximum leverage of 100x
	resp, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "1000.0",
		Leverage:         types.MaxMarginLeverage, // 100x
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
	})
	s.Require().NoError(err)
	s.Require().EqualValues(1, resp.PositionId)

	position, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, resp.PositionId)
	s.Require().True(found)
	s.Require().Equal(types.MaxMarginLeverage, position.Leverage)

	// Position size should be 100x the collateral
	expectedPositionSize := osmomath.MustNewDecFromStr("1000.0").Mul(osmomath.NewDec(int64(types.MaxMarginLeverage)))
	actualPositionSize := position.PositionSizeDec()
	s.Require().True(actualPositionSize.Equal(expectedPositionSize))

	// Check margin pool updates
	marginPool, found := s.App.BondingCurveKeeper.GetMarginPool(s.Ctx, s.denom)
	s.Require().True(found)
	s.Require().True(marginPool.TotalLongExposureDec().Equal(actualPositionSize))
}

func (s *KeeperTestSuite) TestMinLeverage1x() {
	s.setupMarginLiquidity(sdkmath.NewInt(1_000_000), osmomath.OneDec())

	// Test minimum leverage of 1x
	resp, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "500.0",
		Leverage:         types.MinMarginLeverage, // 1x
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
	})
	s.Require().NoError(err)
	s.Require().EqualValues(1, resp.PositionId)

	position, found := s.App.BondingCurveKeeper.GetMarginPosition(s.Ctx, resp.PositionId)
	s.Require().True(found)
	s.Require().Equal(types.MinMarginLeverage, position.Leverage)

	// Position size should equal the collateral (1x leverage)
	expectedPositionSize := osmomath.MustNewDecFromStr("500.0")
	actualPositionSize := position.PositionSizeDec()
	s.Require().True(actualPositionSize.Equal(expectedPositionSize))

	// Check margin pool updates
	marginPool, found := s.App.BondingCurveKeeper.GetMarginPool(s.Ctx, s.denom)
	s.Require().True(found)
	s.Require().True(marginPool.TotalLongExposureDec().Equal(actualPositionSize))
}

func (s *KeeperTestSuite) TestLeverageBelowMinimum() {
	s.setupMarginLiquidity(sdkmath.NewInt(1_000_000), osmomath.OneDec())

	// Test leverage below minimum (should fail)
	_, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "500.0",
		Leverage:         0, // Below minimum
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
	})
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrInvalidLeverage)
}

func (s *KeeperTestSuite) TestLeverageAboveMaximum() {
	s.setupMarginLiquidity(sdkmath.NewInt(1_000_000), osmomath.OneDec())

	// Test leverage above maximum (should fail)
	_, err := s.msgServer.OpenMarginPosition(s.Ctx, &types.MsgOpenMarginPosition{
		Trader:           s.trader.String(),
		Denom:            s.denom,
		CollateralDenom:  s.quoteDenom,
		CollateralAmount: "500.0",
		Leverage:         types.MaxMarginLeverage + 1, // Above maximum
		PositionType:     types.PositionType_POSITION_TYPE_LONG,
	})
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrInvalidLeverage)
}

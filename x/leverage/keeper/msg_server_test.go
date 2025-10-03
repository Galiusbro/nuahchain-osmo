package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/leverage/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/leverage/types"
	usertokentypes "github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

type LeverageTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient    types.QueryClient
	msgServer      types.MsgServer
	leverageKeeper *keeper.Keeper
}

func (s *LeverageTestSuite) SetupTest() {
	s.Setup()

	// Get the leverage keeper from app keepers
	s.leverageKeeper = s.App.LeverageKeeper
	s.msgServer = keeper.NewMsgServerImpl(*s.leverageKeeper)
	s.queryClient = types.NewQueryClient(s.QueryHelper)
}

func TestLeverageTestSuite(t *testing.T) {
	suite.Run(t, new(LeverageTestSuite))
}

func (s *LeverageTestSuite) TestOpenPosition() {
	// Create test addresses
	trader := s.TestAccs[0]

	// Create a test user token first
	userTokenDenom := s.CreateUserToken("testtoken", "TEST", 6, trader)

	// Fund trader with collateral (NDollar)
	collateralDenom := "factory/" + s.TestAccs[1].String() + "/ndollar"
	collateralAmount := math.NewInt(10000_000_000) // 10000 NDollar
	s.FundAcc(trader, sdk.NewCoins(sdk.NewCoin(collateralDenom, collateralAmount)))

	// Test opening a LONG position
	msg := &types.MsgOpenPosition{
		Trader:     trader.String(),
		TokenDenom: userTokenDenom,
		Collateral: sdk.NewCoin(collateralDenom, math.NewInt(100_000_000)), // 100 NDollar
		Leverage:   math.LegacyNewDec(5),                                   // 5x leverage
		Side:       types.PositionSideLong,
		MinPrice:   math.LegacyZeroDec(),
		MaxPrice:   math.LegacyNewDec(1000), // High max price for slippage protection
	}

	// Execute the message
	resp, err := s.msgServer.OpenPosition(s.Ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotEmpty(resp.PositionId)

	// Verify position was created
	position, found := s.leverageKeeper.GetPosition(s.Ctx, resp.PositionId)
	s.Require().True(found)
	s.Require().Equal(trader.String(), position.Trader)
	s.Require().Equal(userTokenDenom, position.TokenDenom)
	s.Require().Equal(collateralDenom, position.CollateralDenom)
	s.Require().Equal(types.PositionSideLong, position.Side)
	s.Require().Equal(types.PositionStatusOpen, position.Status)
	s.Require().True(position.Size_.IsPositive())
	s.Require().Equal(msg.Collateral.Amount, position.Collateral)
	s.Require().Equal(msg.Leverage, position.Leverage)

	// Verify collateral was transferred from trader
	traderBalance := s.App.BankKeeper.GetBalance(s.Ctx, trader, collateralDenom)
	expectedBalance := collateralAmount.Sub(msg.Collateral.Amount)
	s.Require().True(traderBalance.Amount.LTE(expectedBalance)) // LTE because of trading fees
}

func (s *LeverageTestSuite) TestOpenPositionInvalidToken() {
	trader := s.TestAccs[0]

	// Try to open position with non-existent token
	msg := &types.MsgOpenPosition{
		Trader:     trader.String(),
		TokenDenom: "factory/invalid/token",
		Collateral: sdk.NewCoin("unuah", math.NewInt(100_000_000)),
		Leverage:   math.LegacyNewDec(2),
		Side:       types.PositionSideLong,
		MinPrice:   math.LegacyZeroDec(),
		MaxPrice:   math.LegacyNewDec(1000),
	}

	_, err := s.msgServer.OpenPosition(s.Ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "token not supported")
}

func (s *LeverageTestSuite) TestOpenPositionInsufficientCollateral() {
	trader := s.TestAccs[0]
	userTokenDenom := s.CreateUserToken("testtoken", "TEST", 6, trader)

	// Don't fund the trader - they have no collateral
	msg := &types.MsgOpenPosition{
		Trader:     trader.String(),
		TokenDenom: userTokenDenom,
		Collateral: sdk.NewCoin("unuah", math.NewInt(100_000_000)),
		Leverage:   math.LegacyNewDec(2),
		Side:       types.PositionSideLong,
		MinPrice:   math.LegacyZeroDec(),
		MaxPrice:   math.LegacyNewDec(1000),
	}

	_, err := s.msgServer.OpenPosition(s.Ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "insufficient funds")
}

func (s *LeverageTestSuite) TestClosePosition() {
	// First open a position
	trader := s.TestAccs[0]
	userTokenDenom := s.CreateUserToken("testtoken", "TEST", 6, trader)

	collateralDenom := "unuah"
	collateralAmount := math.NewInt(10000_000_000)
	s.FundAcc(trader, sdk.NewCoins(sdk.NewCoin(collateralDenom, collateralAmount)))

	openMsg := &types.MsgOpenPosition{
		Trader:     trader.String(),
		TokenDenom: userTokenDenom,
		Collateral: sdk.NewCoin(collateralDenom, math.NewInt(100_000_000)),
		Leverage:   math.LegacyNewDec(2),
		Side:       types.PositionSideLong,
		MinPrice:   math.LegacyZeroDec(),
		MaxPrice:   math.LegacyNewDec(1000),
	}

	openResp, err := s.msgServer.OpenPosition(s.Ctx, openMsg)
	s.Require().NoError(err)

	// Now close the position
	closeMsg := &types.MsgClosePosition{
		Trader:     trader.String(),
		PositionId: openResp.PositionId,
		MinPrice:   math.LegacyZeroDec(),
		MaxPrice:   math.LegacyNewDec(1000),
	}

	closeResp, err := s.msgServer.ClosePosition(s.Ctx, closeMsg)
	s.Require().NoError(err)
	s.Require().NotNil(closeResp)

	// Verify position is closed
	position, found := s.leverageKeeper.GetPosition(s.Ctx, openResp.PositionId)
	s.Require().True(found)
	s.Require().Equal(types.PositionStatusClosed, position.Status)
}

func (s *LeverageTestSuite) TestClosePositionUnauthorized() {
	// Open position with one trader
	trader1 := s.TestAccs[0]
	trader2 := s.TestAccs[1]

	userTokenDenom := s.CreateUserToken("testtoken", "TEST", 6, trader1)

	collateralAmount := math.NewInt(10000_000_000)
	s.FundAcc(trader1, sdk.NewCoins(sdk.NewCoin("unuah", collateralAmount)))

	openMsg := &types.MsgOpenPosition{
		Trader:     trader1.String(),
		TokenDenom: userTokenDenom,
		Collateral: sdk.NewCoin("unuah", math.NewInt(100_000_000)),
		Leverage:   math.LegacyNewDec(2),
		Side:       types.PositionSideLong,
		MinPrice:   math.LegacyZeroDec(),
		MaxPrice:   math.LegacyNewDec(1000),
	}

	openResp, err := s.msgServer.OpenPosition(s.Ctx, openMsg)
	s.Require().NoError(err)

	// Try to close with different trader
	closeMsg := &types.MsgClosePosition{
		Trader:     trader2.String(), // Different trader!
		PositionId: openResp.PositionId,
		MinPrice:   math.LegacyZeroDec(),
		MaxPrice:   math.LegacyNewDec(1000),
	}

	_, err = s.msgServer.ClosePosition(s.Ctx, closeMsg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unauthorized")
}

func (s *LeverageTestSuite) TestAddCollateral() {
	// First open a position
	trader := s.TestAccs[0]
	userTokenDenom := s.CreateUserToken("testtoken", "TEST", 6, trader)

	collateralDenom := "unuah"
	collateralAmount := math.NewInt(10000_000_000)
	s.FundAcc(trader, sdk.NewCoins(sdk.NewCoin(collateralDenom, collateralAmount)))

	openMsg := &types.MsgOpenPosition{
		Trader:     trader.String(),
		TokenDenom: userTokenDenom,
		Collateral: sdk.NewCoin(collateralDenom, math.NewInt(100_000_000)),
		Leverage:   math.LegacyNewDec(2),
		Side:       types.PositionSideLong,
		MinPrice:   math.LegacyZeroDec(),
		MaxPrice:   math.LegacyNewDec(1000),
	}

	openResp, err := s.msgServer.OpenPosition(s.Ctx, openMsg)
	s.Require().NoError(err)

	// Get initial liquidation price
	initialPosition, _ := s.leverageKeeper.GetPosition(s.Ctx, openResp.PositionId)
	initialLiqPrice := initialPosition.LiquidationPrice

	// Add collateral
	addMsg := &types.MsgAddCollateral{
		Trader:     trader.String(),
		PositionId: openResp.PositionId,
		Amount:     sdk.NewCoin(collateralDenom, math.NewInt(50_000_000)), // Add 50 more
	}

	addResp, err := s.msgServer.AddCollateral(s.Ctx, addMsg)
	s.Require().NoError(err)
	s.Require().NotNil(addResp)

	// Verify collateral was added and liquidation price improved
	updatedPosition, found := s.leverageKeeper.GetPosition(s.Ctx, openResp.PositionId)
	s.Require().True(found)
	s.Require().Equal(
		openMsg.Collateral.Amount.Add(addMsg.Amount.Amount),
		updatedPosition.Collateral,
	)

	// For LONG positions, liquidation price should decrease (improve) when collateral is added
	s.Require().True(updatedPosition.LiquidationPrice.LT(initialLiqPrice))
}

func (s *LeverageTestSuite) TestLeverageParams() {
	// Test getting default params
	params := s.leverageKeeper.GetParams(s.Ctx)
	s.Require().Equal(math.LegacyNewDec(100), params.MaxLeverage)                // 100x max leverage
	s.Require().Equal(math.LegacyNewDecWithPrec(1, 2), params.MaintenanceMargin) // 1%
	s.Require().Equal(math.LegacyNewDecWithPrec(5, 3), params.LiquidationFee)    // 0.5%
	s.Require().Equal(math.LegacyNewDecWithPrec(1, 3), params.TradingFee)        // 0.1%
}

func (s *LeverageTestSuite) TestMaxLeverageValidation() {
	trader := s.TestAccs[0]
	userTokenDenom := s.CreateUserToken("testtoken", "TEST", 6, trader)

	collateralAmount := math.NewInt(10000_000_000)
	s.FundAcc(trader, sdk.NewCoins(sdk.NewCoin("unuah", collateralAmount)))

	// Try to open position with leverage > 100x
	msg := &types.MsgOpenPosition{
		Trader:     trader.String(),
		TokenDenom: userTokenDenom,
		Collateral: sdk.NewCoin("unuah", math.NewInt(100_000_000)),
		Leverage:   math.LegacyNewDec(150), // 150x leverage - should fail
		Side:       types.PositionSideLong,
		MinPrice:   math.LegacyZeroDec(),
		MaxPrice:   math.LegacyNewDec(1000),
	}

	_, err := s.msgServer.OpenPosition(s.Ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "max leverage exceeded")
}

func (s *LeverageTestSuite) TestQueryPosition() {
	// First create a position
	trader := s.TestAccs[0]
	userTokenDenom := s.CreateUserToken("testtoken", "TEST", 6, trader)

	collateralAmount := math.NewInt(10000_000_000)
	s.FundAcc(trader, sdk.NewCoins(sdk.NewCoin("unuah", collateralAmount)))

	openMsg := &types.MsgOpenPosition{
		Trader:     trader.String(),
		TokenDenom: userTokenDenom,
		Collateral: sdk.NewCoin("unuah", math.NewInt(100_000_000)),
		Leverage:   math.LegacyNewDec(3),
		Side:       types.PositionSideLong,
		MinPrice:   math.LegacyZeroDec(),
		MaxPrice:   math.LegacyNewDec(1000),
	}

	openResp, err := s.msgServer.OpenPosition(s.Ctx, openMsg)
	s.Require().NoError(err)

	// Query the position
	queryReq := &types.QueryPositionRequest{
		PositionId: openResp.PositionId,
	}

	queryResp, err := s.queryClient.Position(s.Ctx, queryReq)
	s.Require().NoError(err)
	s.Require().NotNil(queryResp)
	s.Require().Equal(openResp.PositionId, queryResp.Position.Id)
	s.Require().Equal(trader.String(), queryResp.Position.Trader)
	s.Require().Equal(userTokenDenom, queryResp.Position.TokenDenom)
}

func (s *LeverageTestSuite) TestQueryPositionsByTrader() {
	trader := s.TestAccs[0]
	userTokenDenom := s.CreateUserToken("testtoken", "TEST", 6, trader)

	collateralAmount := math.NewInt(10000_000_000)
	s.FundAcc(trader, sdk.NewCoins(sdk.NewCoin("unuah", collateralAmount)))

	// Create multiple positions
	for i := 0; i < 3; i++ {
		openMsg := &types.MsgOpenPosition{
			Trader:     trader.String(),
			TokenDenom: userTokenDenom,
			Collateral: sdk.NewCoin("unuah", math.NewInt(500_000_000)),
			Leverage:   math.LegacyNewDec(2),
			Side:       types.PositionSideLong,
			MinPrice:   math.LegacyZeroDec(),
			MaxPrice:   math.LegacyNewDec(1000),
		}

		_, err := s.msgServer.OpenPosition(s.Ctx, openMsg)
		s.Require().NoError(err)
	}

	// Query positions by trader
	queryReq := &types.QueryPositionsByTraderRequest{
		Trader: trader.String(),
	}

	queryResp, err := s.queryClient.PositionsByTrader(s.Ctx, queryReq)
	s.Require().NoError(err)
	s.Require().NotNil(queryResp)
	s.Require().Len(queryResp.Positions, 3)

	// All positions should belong to the trader
	for _, position := range queryResp.Positions {
		s.Require().Equal(trader.String(), position.Trader)
		s.Require().Equal(types.PositionStatusOpen, position.Status)
	}
}

// Helper method to create a user token for testing
func (s *LeverageTestSuite) CreateUserToken(subdenom, symbol string, decimals uint32, creator sdk.AccAddress) string {
	// Create a real user token through the usertoken module
	denom := "factory/" + creator.String() + "/" + subdenom

	// Create the token in usertoken module
	userToken := usertokentypes.UserToken{
		Denom:                denom,
		Creator:              creator.String(),
		Name:                 symbol,
		Symbol:               symbol,
		MaxSupply:            math.NewInt(1000000000000), // 1 trillion max supply
		CurrentSupply:        math.NewInt(1000000000000), // 1 trillion current supply
		FounderTokensClaimed: math.NewInt(0),
		LbpActive:            false,
		LbpStartTime:         0,
	}

	s.App.UserTokenKeeper.SetUserToken(s.Ctx, userToken)

	return denom
}

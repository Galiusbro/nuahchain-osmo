package keeper_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/usertoken/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	msgServer types.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.Ctx = s.Ctx.WithBlockTime(time.Unix(1_700_000_000, 0))

	for len(s.TestAccs) < 5 {
		extra := apptesting.CreateRandomAccounts(1)[0]
		s.TestAccs = append(s.TestAccs, extra)
	}

	params := s.App.UserTokenKeeper.GetParams(s.Ctx)
	params.BondingCurveWallet = s.TestAccs[1].String()
	params.PlatformWallet = s.TestAccs[2].String()
	params.ReferralWallet = s.TestAccs[3].String()
	params.AiCeoWallet = s.TestAccs[4].String()
	params.FounderClaimPeriod = 3600
	s.App.UserTokenKeeper.SetParams(s.Ctx, params)

	s.msgServer = keeper.NewMsgServerImpl(*s.App.UserTokenKeeper)
}

func (s *KeeperTestSuite) TestCreateTokenSuccess() {
	creator := s.TestAccs[0]

	resp, err := s.msgServer.CreateToken(sdk.WrapSDKContext(s.Ctx), types.NewMsgCreateToken(
		creator.String(),
		"Nuah Dollar",
		"NDLR",
		"https://example.com/logo.png",
		"Test token",
	))
	s.Require().NoError(err)

	expectedDenom := fmt.Sprintf("factory/%s/%s", creator.String(), strings.ToLower("NDLR"))
	s.Require().Equal(expectedDenom, resp.Denom)

	token, found := s.App.UserTokenKeeper.GetToken(s.Ctx, resp.Denom)
	s.Require().True(found)
	s.Require().Equal("Nuah Dollar", token.Name)
	s.Require().Equal("NDLR", token.Symbol)
	s.Require().Equal(uint64(s.Ctx.BlockTime().Unix()), token.CreatedAt)
	s.Require().False(token.Distribution.FounderClaimed)
	s.Require().Equal(osmomath.NewInt(100_000_000).String(), token.Distribution.TotalSupply)
	s.Require().Equal(osmomath.NewInt(30_000_000).String(), token.Distribution.BondingCurveSupply)
	s.Require().Equal(osmomath.NewInt(10_000_000).String(), token.Distribution.PlatformWallet)
	s.Require().Equal(osmomath.NewInt(10_000_000).String(), token.Distribution.ReferralWallet)
	s.Require().Equal(osmomath.NewInt(40_000_000).String(), token.Distribution.AiCeoWallet)
	s.Require().Equal(osmomath.NewInt(10_000_000).String(), token.Distribution.FounderReserved)

	expectedDeadline := uint64(s.Ctx.BlockTime().Add(time.Hour).Unix())
	s.Require().Equal(expectedDeadline, token.Distribution.FounderClaimDeadline)
	state := token.State
	s.Require().Equal("0", state.BondingCurveSold)
	s.Require().True(state.SoftLockEnabled)

	params := s.App.UserTokenKeeper.GetParams(s.Ctx)

	bondingAddr, err := sdk.AccAddressFromBech32(params.BondingCurveWallet)
	s.Require().NoError(err)
	platformAddr, err := sdk.AccAddressFromBech32(params.PlatformWallet)
	s.Require().NoError(err)
	referralAddr, err := sdk.AccAddressFromBech32(params.ReferralWallet)
	s.Require().NoError(err)
	aiAddr, err := sdk.AccAddressFromBech32(params.AiCeoWallet)
	s.Require().NoError(err)

	bondingBalance := s.App.BankKeeper.GetBalance(s.Ctx, bondingAddr, resp.Denom)
	s.Require().Equal(osmomath.NewInt(30_000_000), bondingBalance.Amount)
	platformBalance := s.App.BankKeeper.GetBalance(s.Ctx, platformAddr, resp.Denom)
	s.Require().Equal(osmomath.NewInt(10_000_000), platformBalance.Amount)
	referralBalance := s.App.BankKeeper.GetBalance(s.Ctx, referralAddr, resp.Denom)
	s.Require().Equal(osmomath.NewInt(10_000_000), referralBalance.Amount)
	aiBalance := s.App.BankKeeper.GetBalance(s.Ctx, aiAddr, resp.Denom)
	s.Require().Equal(osmomath.NewInt(40_000_000), aiBalance.Amount)

	creatorBalance := s.App.BankKeeper.GetBalance(s.Ctx, creator, resp.Denom)
	s.Require().True(creatorBalance.IsZero())
}

func (s *KeeperTestSuite) TestCreateTokenUniqueness() {
	creator := s.TestAccs[0]

	_, err := s.msgServer.CreateToken(sdk.WrapSDKContext(s.Ctx), types.NewMsgCreateToken(
		creator.String(), "First Token", "ONE", "", "first",
	))
	s.Require().NoError(err)

	_, err = s.msgServer.CreateToken(sdk.WrapSDKContext(s.Ctx), types.NewMsgCreateToken(
		creator.String(), "First Token", "TWO", "", "duplicate name",
	))
	s.Require().ErrorIs(err, types.ErrNameExists)

	_, err = s.msgServer.CreateToken(sdk.WrapSDKContext(s.Ctx), types.NewMsgCreateToken(
		creator.String(), "Second Token", "ONE", "", "duplicate symbol",
	))
	s.Require().ErrorIs(err, types.ErrSymbolExists)
}

func (s *KeeperTestSuite) TestFounderAllocationExpires() {
	creator := s.TestAccs[0]

	resp, err := s.msgServer.CreateToken(sdk.WrapSDKContext(s.Ctx), types.NewMsgCreateToken(
		creator.String(), "Expiry Token", "EXP", "", "founder expires",
	))
	s.Require().NoError(err)

	params := s.App.UserTokenKeeper.GetParams(s.Ctx)
	bondingAddr, err := sdk.AccAddressFromBech32(params.BondingCurveWallet)
	s.Require().NoError(err)

	// Advance time past deadline and run end blocker
	deadline := s.Ctx.BlockTime().Add(time.Hour + time.Second)
	s.Ctx = s.Ctx.WithBlockTime(deadline)
	s.Require().NoError(s.App.UserTokenKeeper.EndBlocker(sdk.WrapSDKContext(s.Ctx)))

	token, found := s.App.UserTokenKeeper.GetToken(s.Ctx, resp.Denom)
	s.Require().True(found)
	s.Require().True(token.Distribution.FounderClaimed)
	s.Require().Equal(osmomath.ZeroInt().String(), token.Distribution.FounderReserved)
	s.Require().Equal(osmomath.NewInt(40_000_000).String(), token.Distribution.BondingCurveSupply)

	bondingBalance := s.App.BankKeeper.GetBalance(s.Ctx, bondingAddr, resp.Denom)
	s.Require().Equal(osmomath.NewInt(40_000_000), bondingBalance.Amount)
}

package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/tokenfactory/types"
)

// TestSmokeCreateAndMint ensures a denom can be created and minted through MsgCreateDenom/MsgMint.
func (s *KeeperTestSuite) TestSmokeCreateAndMint() {
	creator := s.TestAccs[0]
	recipient := s.TestAccs[1]
	subdenom := "smoke"

	createResp, err := s.msgServer.CreateDenom(s.Ctx, types.NewMsgCreateDenom(creator.String(), subdenom))
	s.Require().NoError(err)

	expectedDenom, err := types.GetTokenDenom(creator.String(), subdenom)
	s.Require().NoError(err)
	s.Require().Equal(expectedDenom, createResp.GetNewTokenDenom())

	authorityMetadata, err := s.App.TokenFactoryKeeper.GetAuthorityMetadata(s.Ctx, expectedDenom)
	s.Require().NoError(err)
	s.Require().Equal(creator.String(), authorityMetadata.Admin)

	initialSupply := sdk.NewInt64Coin(expectedDenom, 1_000_000)
	_, err = s.msgServer.Mint(s.Ctx, types.NewMsgMint(creator.String(), initialSupply))
	s.Require().NoError(err)

	creatorBalance := s.App.BankKeeper.GetBalance(s.Ctx, creator, expectedDenom)
	s.Require().Equal(initialSupply, creatorBalance)

	totalSupply := s.App.BankKeeper.GetSupply(s.Ctx, expectedDenom)
	s.Require().Equal(initialSupply.Amount, totalSupply.Amount)

	recipientMint := sdk.NewInt64Coin(expectedDenom, 500_000)
	_, err = s.msgServer.Mint(s.Ctx, types.NewMsgMintTo(creator.String(), recipientMint, recipient.String()))
	s.Require().NoError(err)

	finalSupply := s.App.BankKeeper.GetSupply(s.Ctx, expectedDenom)
	s.Require().Equal(initialSupply.Amount.Add(recipientMint.Amount), finalSupply.Amount)

	recipientBalance := s.App.BankKeeper.GetBalance(s.Ctx, recipient, expectedDenom)
	s.Require().Equal(recipientMint, recipientBalance)
	s.Require().Equal(initialSupply, s.App.BankKeeper.GetBalance(s.Ctx, creator, expectedDenom))

	_, err = s.msgServer.Mint(s.Ctx, types.NewMsgMint(recipient.String(), sdk.NewInt64Coin(expectedDenom, 1)))
	s.Require().ErrorIs(err, types.ErrUnauthorized)
}

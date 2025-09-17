package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/usertoken/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

type VestingTestSuite struct {
	apptesting.KeeperTestHelper

	msgServer types.MsgServer
}

func (suite *VestingTestSuite) SetupTest() {
	suite.Setup()
	suite.msgServer = keeper.NewMsgServerImpl(*suite.App.UserTokenKeeper, suite.App.UserTokenKeeper.GetAuthority())
}

func TestVestingTestSuite(t *testing.T) {
	suite.Run(t, new(VestingTestSuite))
}

func (suite *VestingTestSuite) TestCreateVestingAccount() {
	// Setup fresh context for this test
	suite.Setup()
	suite.msgServer = keeper.NewMsgServerImpl(*suite.App.UserTokenKeeper, suite.App.UserTokenKeeper.GetAuthority())

	// Fund the creator account with tokens
	creatorAddr, _ := sdk.AccAddressFromBech32("nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d")
	fundCoins := sdk.NewCoins(sdk.NewCoin("factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/mytoken", sdkmath.NewInt(10000000)))
	suite.FundAcc(creatorAddr, fundCoins)

	testCases := []struct {
		name        string
		msg         *types.MsgCreateVestingAccount
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid vesting account creation",
			msg: &types.MsgCreateVestingAccount{
				Creator:   "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				ToAddress: "nuah1v4musx0tts90ka7nf8taae3c9k3mna663r6trc",
				Amount:    []sdk.Coin{{Denom: "factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/mytoken", Amount: sdkmath.NewInt(1000000)}},
				EndTime:   time.Now().Add(365 * 24 * time.Hour).Unix(),
				Delayed:   false,
			},
			expectError: false,
		},
		{
			name: "invalid from address",
			msg: &types.MsgCreateVestingAccount{
				Creator:   "invalid_address",
				ToAddress: "nuah1335z3d325yemqa9lzk6kl4el9a05v45hy7ven5",
				Amount:    []sdk.Coin{{Denom: "factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/mytoken", Amount: sdkmath.NewInt(1000000)}},
				EndTime:   time.Now().Add(365 * 24 * time.Hour).Unix(),
				Delayed:   false,
			},
			expectError: true,
			errorMsg:    "invalid creator address",
		},
		{
			name: "invalid to address",
			msg: &types.MsgCreateVestingAccount{
				Creator:   "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				ToAddress: "invalid_address",
				Amount:    []sdk.Coin{{Denom: "factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/mytoken", Amount: sdkmath.NewInt(1000000)}},
				EndTime:   time.Now().Add(365 * 24 * time.Hour).Unix(),
				Delayed:   false,
			},
			expectError: true,
			errorMsg:    "invalid to_address",
		},
		{
			name: "empty amount",
			msg: &types.MsgCreateVestingAccount{
				Creator:   "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				ToAddress: "nuah1ls0meskk3sc0jppxf89sa7c9ew2g2mmgk0r4k7",
				Amount:    []sdk.Coin{},
				EndTime:   time.Now().Add(365 * 24 * time.Hour).Unix(),
				Delayed:   false,
			},
			expectError: true,
			errorMsg:    "amount cannot be empty",
		},
		{
			name: "end time in the past",
			msg: &types.MsgCreateVestingAccount{
				Creator:   "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				ToAddress: "nuah1g5schkujd763ez99wdr54d388duzvx0x2lduyq",
				Amount:    []sdk.Coin{{Denom: "factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/mytoken", Amount: sdkmath.NewInt(1000000)}},
				EndTime:   time.Now().Add(-24 * time.Hour).Unix(),
				Delayed:   false,
			},
			expectError: true,
			errorMsg:    "end_time must be in the future",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.CreateVestingAccount(suite.Ctx, tc.msg)

			if tc.expectError {
				suite.Require().Error(err)
				if tc.errorMsg != "" {
					suite.Require().Contains(err.Error(), tc.errorMsg)
				}
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *VestingTestSuite) TestCreateVestingAccountValidation() {
	// Setup fresh context for this test
	suite.Setup()
	suite.msgServer = keeper.NewMsgServerImpl(*suite.App.UserTokenKeeper, suite.App.UserTokenKeeper.GetAuthority())

	// Test message validation
	msg := &types.MsgCreateVestingAccount{
		Creator:   "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
		ToAddress: "nuah1lqam4xhdqrczfc949539l9r8zfyfxel2y2q8j4",
		Amount:    []sdk.Coin{{Denom: "factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/mytoken", Amount: sdkmath.NewInt(1000000)}},
		EndTime:   time.Now().Add(365 * 24 * time.Hour).Unix(),
		Delayed:   false,
	}

	err := msg.ValidateBasic()
	suite.Require().NoError(err)

	// Test invalid message
	invalidMsg := &types.MsgCreateVestingAccount{
		Creator:   "",
		ToAddress: "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
		Amount:    []sdk.Coin{{Denom: "factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/mytoken", Amount: sdkmath.NewInt(1000000)}},
		EndTime:   time.Now().Add(365 * 24 * time.Hour).Unix(),
		Delayed:   false,
	}

	err = invalidMsg.ValidateBasic()
	suite.Require().Error(err)
}

func (suite *VestingTestSuite) TestCreateVestingAccountEvents() {
	// Setup fresh context for this test
	suite.Setup()
	suite.msgServer = keeper.NewMsgServerImpl(*suite.App.UserTokenKeeper, suite.App.UserTokenKeeper.GetAuthority())

	// Fund the creator account with tokens
	creatorAddr, _ := sdk.AccAddressFromBech32("nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d")
	fundCoins := sdk.NewCoins(sdk.NewCoin("factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/mytoken", sdkmath.NewInt(10000000)))
	suite.FundAcc(creatorAddr, fundCoins)

	// Test that proper events are emitted
	msg := &types.MsgCreateVestingAccount{
		Creator:   "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
		ToAddress: "nuah1335z3d325yemqa9lzk6kl4el9a05v45hy7ven5",
		Amount:    []sdk.Coin{{Denom: "factory/nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d/mytoken", Amount: sdkmath.NewInt(1000000)}},
		EndTime:   time.Now().Add(365 * 24 * time.Hour).Unix(),
		Delayed:   false,
	}

	// Create context with event manager
	ctx := suite.Ctx.WithEventManager(sdk.NewEventManager())

	_, err := suite.msgServer.CreateVestingAccount(ctx, msg)
	suite.Require().NoError(err)

	// Check that events were emitted
	events := ctx.EventManager().Events()
	suite.Require().Greater(len(events), 0)

	// Check for specific event type
	found := false
	for _, event := range events {
		if event.Type == "create_vesting_account" {
			found = true
			break
		}
	}
	suite.Require().True(found, "create_vesting_account event should be emitted")
}

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/usertoken/keeper"
	usertokentypes "github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

type UpdateParamsTestSuite struct {
	apptesting.KeeperTestHelper

	msgServer usertokentypes.MsgServer
}

func (suite *UpdateParamsTestSuite) SetupTest() {
	suite.Setup()
	suite.msgServer = keeper.NewMsgServerImpl(*suite.App.UserTokenKeeper, suite.App.UserTokenKeeper.GetAuthority())
}

func TestUpdateParamsTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateParamsTestSuite))
}

func (suite *UpdateParamsTestSuite) TestUpdateParams() {
	testCases := []struct {
		name      string
		authority string
		newParams usertokentypes.Params
		expectErr bool
	}{
		{
			name:      "valid authority and params",
			authority: suite.App.UserTokenKeeper.GetAuthority(),
			newParams: usertokentypes.Params{
				FounderTranchePrice:  math.LegacyNewDec(1000), // Change from 500 to 1000
				FounderTrancheAmount: math.NewInt(10000000),
				AiCeoWallet:          "osmo1test",
			},
			expectErr: false,
		},
		{
			name:      "invalid authority",
			authority: "invalid_authority",
			newParams: usertokentypes.Params{
				FounderTranchePrice:  math.LegacyNewDec(1000),
				FounderTrancheAmount: math.NewInt(10000000),
				AiCeoWallet:          "osmo1test",
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create the message
			msg := &usertokentypes.MsgUpdateParams{
				Authority: tc.authority,
				Params:    tc.newParams,
			}

			// Execute the message
			_, err := suite.msgServer.UpdateParams(suite.Ctx, msg)

			if tc.expectErr {
				require.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)

				// Verify the params were updated
				updatedParams := suite.App.UserTokenKeeper.GetParams(suite.Ctx)
				require.Equal(suite.T(), tc.newParams.FounderTranchePrice, updatedParams.FounderTranchePrice)
				require.Equal(suite.T(), tc.newParams.FounderTrancheAmount, updatedParams.FounderTrancheAmount)
				require.Equal(suite.T(), tc.newParams.AiCeoWallet, updatedParams.AiCeoWallet)
			}
		})
	}
}

func (suite *UpdateParamsTestSuite) TestUpdateParamsChangePrice() {
	// Test specifically changing the founder tranche price from 0.00005 to another value
	initialParams := suite.App.UserTokenKeeper.GetParams(suite.Ctx)
	require.Equal(suite.T(), math.LegacyMustNewDecFromStr("0.00005"), initialParams.FounderTranchePrice)

	// Update to new price (change from 0.00005 to 0.0001)
	newPrice := math.LegacyMustNewDecFromStr("0.0001")
	newParams := usertokentypes.Params{
		FounderTranchePrice:    newPrice,
		FounderTrancheAmount:   initialParams.FounderTrancheAmount,
		BondingCurveStartPrice: initialParams.BondingCurveStartPrice,
		BondingCurveEndPrice:   initialParams.BondingCurveEndPrice,
		BondingCurveMaxSupply:  initialParams.BondingCurveMaxSupply,
		MinCreatorPurchase:     initialParams.MinCreatorPurchase,
		AiCeoWallet:            initialParams.AiCeoWallet,
		ReferralWallet:         initialParams.ReferralWallet,
		PlatformFeeWallet:      initialParams.PlatformFeeWallet,
	}

	msg := &usertokentypes.MsgUpdateParams{
		Authority: suite.App.UserTokenKeeper.GetAuthority(),
		Params:    newParams,
	}

	_, err := suite.msgServer.UpdateParams(suite.Ctx, msg)
	require.NoError(suite.T(), err)

	// Verify the price was changed
	updatedParams := suite.App.UserTokenKeeper.GetParams(suite.Ctx)
	require.Equal(suite.T(), newPrice, updatedParams.FounderTranchePrice)
	require.NotEqual(suite.T(), math.LegacyMustNewDecFromStr("0.00005"), updatedParams.FounderTranchePrice)
}

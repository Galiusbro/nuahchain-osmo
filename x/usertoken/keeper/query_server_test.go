package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/usertoken/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

type QueryServerTestSuite struct {
	apptesting.KeeperTestHelper

	queryServer types.QueryServer
}

func (suite *QueryServerTestSuite) SetupTest() {
	suite.Setup()
	suite.queryServer = keeper.NewQueryServerImpl(*suite.App.UserTokenKeeper)
}

func TestQueryServerTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

func (suite *QueryServerTestSuite) TestQueryParams() {
	tests := []struct {
		name string
		req  *types.QueryParamsRequest
		err  error
	}{
		{
			name: "valid request",
			req:  &types.QueryParamsRequest{},
			err:  nil,
		},
		{
			name: "nil request",
			req:  nil,
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			queryServer := suite.queryServer
			resp, err := queryServer.Params(suite.Ctx, tt.req)

			if tt.err != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tt.err.Error(), err.Error())
				suite.Require().Nil(resp)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().NotNil(resp.Params)
			}
		})
	}
}

func (suite *QueryServerTestSuite) TestQueryUserToken() {
	// Create a test user token first
	testToken := types.UserToken{
		Denom:                "factory/osmo1abc/testtoken",
		Creator:              suite.TestAccs[0].String(),
		CurrentSupply:        math.NewInt(1000000),
		FounderTokensClaimed: math.NewInt(0),
		LbpActive:            false,
		LbpStartTime:         0,
	}
	suite.App.UserTokenKeeper.SetUserToken(suite.Ctx, testToken)

	tests := []struct {
		name string
		req  *types.QueryUserTokenRequest
		err  error
	}{
		{
			name: "valid request",
			req:  &types.QueryUserTokenRequest{Denom: testToken.Denom},
			err:  nil,
		},
		{
			name: "nil request",
			req:  nil,
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
		{
			name: "empty denom",
			req:  &types.QueryUserTokenRequest{Denom: ""},
			err:  status.Error(codes.InvalidArgument, "denom cannot be empty"),
		},
		{
			name: "token not found",
			req:  &types.QueryUserTokenRequest{Denom: "factory/osmo1abc/nonexistent"},
			err:  status.Error(codes.NotFound, "user token not found"),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			queryServer := suite.queryServer
			resp, err := queryServer.UserToken(suite.Ctx, tt.req)

			if tt.err != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tt.err.Error(), err.Error())
				suite.Require().Nil(resp)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().NotNil(resp.UserToken)
				suite.Require().Equal(testToken.Denom, resp.UserToken.Denom)
				suite.Require().Equal(testToken.Creator, resp.UserToken.Creator)
			}
		})
	}
}

func (suite *QueryServerTestSuite) TestQueryUserTokens() {
	// Create multiple test user tokens
	testTokens := []types.UserToken{
		{
			Denom:                "factory/osmo1abc/token1",
			Creator:              suite.TestAccs[0].String(),
			CurrentSupply:        math.NewInt(1000000),
			FounderTokensClaimed: math.NewInt(0),
			LbpActive:            false,
			LbpStartTime:         0,
		},
		{
			Denom:                "factory/osmo1def/token2",
			Creator:              suite.TestAccs[1].String(),
			CurrentSupply:        math.NewInt(2000000),
			FounderTokensClaimed: math.NewInt(0),
			LbpActive:            false,
			LbpStartTime:         0,
		},
	}

	for _, token := range testTokens {
		suite.App.UserTokenKeeper.SetUserToken(suite.Ctx, token)
	}

	tests := []struct {
		name string
		req  *types.QueryUserTokensRequest
		err  error
	}{
		{
			name: "valid request",
			req:  &types.QueryUserTokensRequest{},
			err:  nil,
		},
		{
			name: "nil request",
			req:  nil,
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			queryServer := suite.queryServer
			resp, err := queryServer.UserTokens(suite.Ctx, tt.req)

			if tt.err != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tt.err.Error(), err.Error())
				suite.Require().Nil(resp)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().NotNil(resp.UserTokens)
				// Should return at least the tokens we created
				suite.Require().GreaterOrEqual(len(resp.UserTokens), len(testTokens))
			}
		})
	}
}

func (suite *QueryServerTestSuite) TestQueryBondingCurvePrice() {
	// Create a test user token first
	testToken := types.UserToken{
		Denom:                "factory/osmo1abc/pricetoken",
		Creator:              suite.TestAccs[0].String(),
		CurrentSupply:        math.NewInt(5000000), // 5M tokens
		FounderTokensClaimed: math.NewInt(0),
		LbpActive:            false,
		LbpStartTime:         0,
	}
	suite.App.UserTokenKeeper.SetUserToken(suite.Ctx, testToken)

	tests := []struct {
		name string
		req  *types.QueryBondingCurvePriceRequest
		err  error
	}{
		{
			name: "valid request",
			req:  &types.QueryBondingCurvePriceRequest{Denom: testToken.Denom},
			err:  nil,
		},
		{
			name: "nil request",
			req:  nil,
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
		{
			name: "empty denom",
			req:  &types.QueryBondingCurvePriceRequest{Denom: ""},
			err:  status.Error(codes.InvalidArgument, "denom cannot be empty"),
		},
		{
			name: "token not found",
			req:  &types.QueryBondingCurvePriceRequest{Denom: "factory/osmo1abc/nonexistent"},
			err:  nil, // GetTokenSupply doesn't return error for non-existent tokens, returns zero supply
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			queryServer := suite.queryServer
			resp, err := queryServer.BondingCurvePrice(suite.Ctx, tt.req)

			if tt.err != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tt.err.Error(), err.Error())
				suite.Require().Nil(resp)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().True(resp.Price.IsPositive())
				// Price should be between min (0.0002) and max (1.0)
				minPrice := math.LegacyNewDecWithPrec(2, 4) // 0.0002
				maxPrice := math.LegacyOneDec()             // 1.0
				suite.Require().True(resp.Price.GTE(minPrice))
				suite.Require().True(resp.Price.LTE(maxPrice))
			}
		})
	}
}

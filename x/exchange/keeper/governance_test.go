package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v30/x/exchange/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/exchange/types"
	usdoracletypes "github.com/osmosis-labs/osmosis/v30/x/usdoracle/types"
	apptesting "github.com/osmosis-labs/osmosis/v30/app/apptesting"
)

// GovernanceTestSuite tests governance functionality for Exchange module
type GovernanceTestSuite struct {
	apptesting.KeeperTestHelper

	keeper *keeper.Keeper
	ctx    sdk.Context
}

func TestGovernanceTestSuite(t *testing.T) {
	suite.Run(t, new(GovernanceTestSuite))
}

func (s *GovernanceTestSuite) SetupTest() {
	s.Setup()
	s.ctx = s.Ctx

	// Initialize Exchange keeper
	s.keeper = s.App.ExchangeKeeper

	// Setup USD Oracle with supported tokens
	s.setupUSDOracle()

	// Setup Exchange module with supported tokens
	s.setupExchangeParams()
}

func (s *GovernanceTestSuite) setupUSDOracle() {
	// Setup USD Oracle with default supported tokens
	usdOracleParams := usdoracletypes.Params{
		SupportedTokens: []usdoracletypes.SupportedToken{
			{Denom: "ibc/ETH", Symbol: "ETH", Name: "Ethereum", Enabled: true, Decimals: 18, MinUpdateFrequency: 60, MaxPriceDeviation: osmomath.NewDecWithPrec(5, 2)},
			{Denom: "ibc/BTC", Symbol: "BTC", Name: "Bitcoin", Enabled: true, Decimals: 8, MinUpdateFrequency: 60, MaxPriceDeviation: osmomath.NewDecWithPrec(5, 2)},
			{Denom: "ibc/USDC", Symbol: "USDC", Name: "USD Coin", Enabled: true, Decimals: 6, MinUpdateFrequency: 60, MaxPriceDeviation: osmomath.NewDecWithPrec(2, 2)},
			{Denom: "ibc/ATOM", Symbol: "ATOM", Name: "Cosmos", Enabled: true, Decimals: 6, MinUpdateFrequency: 60, MaxPriceDeviation: osmomath.NewDecWithPrec(5, 2)},
		},
		Enabled:                 true,
		UpdateInterval:          60,                            // 1 minute
		PriceDeviationThreshold: osmomath.NewDecWithPrec(5, 2), // 5%
		MinSources:              1,
		MaxPriceAge:             300, // 5 minutes
	}
	s.App.USDOracleKeeper.SetParams(s.ctx, usdOracleParams)
}

func (s *GovernanceTestSuite) setupExchangeParams() {
	// Setup Exchange module with supported tokens that match USD Oracle
	exchangeParams := types.DefaultParams()
	exchangeParams.SupportedTokens = []string{"ibc/ETH", "ibc/BTC", "ibc/USDC"}
	err := s.keeper.SetParams(s.ctx, exchangeParams)
	s.Require().NoError(err)
}

// TestAddSupportedTokenSuccess tests successful addition of a supported token by governance
func (s *GovernanceTestSuite) TestAddSupportedTokenSuccess() {
	// Get governance authority (should be gov module address)
	govAuthority := s.App.AccountKeeper.GetModuleAddress("gov").String()

	// Create message to add supported token
	msg := &types.MsgAddSupportedToken{
		Authority: govAuthority,
		Denom:     "ibc/ATOM",
	}

	// Execute message
	msgServer := keeper.NewMsgServerImpl(*s.keeper)
	_, err := msgServer.AddSupportedToken(s.ctx, msg)
	s.Require().NoError(err)

	// Verify token was added to supported tokens list
	params, err := s.keeper.GetParams(s.ctx)
	s.Require().NoError(err)

	found := false
	for _, token := range params.SupportedTokens {
		if token == "ibc/ATOM" {
			found = true
			break
		}
	}
	s.Require().True(found, "Token ibc/ATOM should be in supported tokens list")
}

// TestAddSupportedTokenInvalidAuthority tests rejection when non-governance tries to add token
func (s *GovernanceTestSuite) TestAddSupportedTokenInvalidAuthority() {
	// Use regular user address instead of governance
	userAddr := s.TestAccs[0].String()

	// Create message to add supported token with invalid authority
	msg := &types.MsgAddSupportedToken{
		Authority: userAddr,
		Denom:     "ibc/ATOM",
	}

	// Execute message - should fail
	msgServer := keeper.NewMsgServerImpl(*s.keeper)
	_, err := msgServer.AddSupportedToken(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid authority")

	// Verify token was NOT added to supported tokens list
	params, err := s.keeper.GetParams(s.ctx)
	s.Require().NoError(err)

	found := false
	for _, token := range params.SupportedTokens {
		if token == "ibc/ATOM" {
			found = true
			break
		}
	}
	s.Require().False(found, "Token ibc/ATOM should NOT be in supported tokens list")
}

// TestAddSupportedTokenAlreadyExists tests rejection when trying to add existing token
func (s *GovernanceTestSuite) TestAddSupportedTokenAlreadyExists() {
	// Get governance authority
	govAuthority := s.App.AccountKeeper.GetModuleAddress("gov").String()

	// Try to add token that already exists in default params
	msg := &types.MsgAddSupportedToken{
		Authority: govAuthority,
		Denom:     "ibc/ETH", // This token is already in default supported tokens
	}

	// Execute message - should fail
	msgServer := keeper.NewMsgServerImpl(*s.keeper)
	_, err := msgServer.AddSupportedToken(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "already supported")
}

// TestRemoveSupportedTokenSuccess tests successful removal of a supported token by governance
func (s *GovernanceTestSuite) TestRemoveSupportedTokenSuccess() {
	// Get governance authority
	govAuthority := s.App.AccountKeeper.GetModuleAddress("gov").String()

	// First verify token exists in supported tokens
	params, err := s.keeper.GetParams(s.ctx)
	s.Require().NoError(err)

	found := false
	for _, token := range params.SupportedTokens {
		if token == "ibc/ETH" {
			found = true
			break
		}
	}
	s.Require().True(found, "Token ibc/ETH should be in supported tokens list initially")

	// Create message to remove supported token
	msg := &types.MsgRemoveSupportedToken{
		Authority: govAuthority,
		Denom:     "ibc/ETH",
	}

	// Execute message
	msgServer := keeper.NewMsgServerImpl(*s.keeper)
	_, err = msgServer.RemoveSupportedToken(s.ctx, msg)
	s.Require().NoError(err)

	// Verify token was removed from supported tokens list
	params, err = s.keeper.GetParams(s.ctx)
	s.Require().NoError(err)

	found = false
	for _, token := range params.SupportedTokens {
		if token == "ibc/ETH" {
			found = true
			break
		}
	}
	s.Require().False(found, "Token ibc/ETH should NOT be in supported tokens list after removal")
}

// TestRemoveSupportedTokenInvalidAuthority tests rejection when non-governance tries to remove token
func (s *GovernanceTestSuite) TestRemoveSupportedTokenInvalidAuthority() {
	// Use regular user address instead of governance
	userAddr := s.TestAccs[0].String()

	// Create message to remove supported token with invalid authority
	msg := &types.MsgRemoveSupportedToken{
		Authority: userAddr,
		Denom:     "ibc/ETH",
	}

	// Execute message - should fail
	msgServer := keeper.NewMsgServerImpl(*s.keeper)
	_, err := msgServer.RemoveSupportedToken(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid authority")

	// Verify token is still in supported tokens list
	params, err := s.keeper.GetParams(s.ctx)
	s.Require().NoError(err)

	found := false
	for _, token := range params.SupportedTokens {
		if token == "ibc/ETH" {
			found = true
			break
		}
	}
	s.Require().True(found, "Token ibc/ETH should still be in supported tokens list")
}

// TestRemoveSupportedTokenNotExists tests rejection when trying to remove non-existent token
func (s *GovernanceTestSuite) TestRemoveSupportedTokenNotExists() {
	// Get governance authority
	govAuthority := s.App.AccountKeeper.GetModuleAddress("gov").String()

	// Try to remove token that doesn't exist
	msg := &types.MsgRemoveSupportedToken{
		Authority: govAuthority,
		Denom:     "ibc/NONEXISTENT",
	}

	// Execute message - should fail
	msgServer := keeper.NewMsgServerImpl(*s.keeper)
	_, err := msgServer.RemoveSupportedToken(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "not in supported tokens registry")
}

// TestGovernanceOnlyCanManageTokens tests that only governance can manage supported tokens
func (s *GovernanceTestSuite) TestGovernanceOnlyCanManageTokens() {
	// Test multiple unauthorized addresses
	unauthorizedAddresses := []string{
		s.TestAccs[0].String(), // Regular user
		s.TestAccs[1].String(), // Another user
		s.App.AccountKeeper.GetModuleAddress("exchange").String(), // Exchange module itself
		s.App.AccountKeeper.GetModuleAddress("bank").String(),     // Bank module
	}

	for _, unauthorizedAddr := range unauthorizedAddresses {
		// Test AddSupportedToken with unauthorized address
		addMsg := &types.MsgAddSupportedToken{
			Authority: unauthorizedAddr,
			Denom:     "ibc/UNAUTHORIZED",
		}

		msgServer := keeper.NewMsgServerImpl(*s.keeper)
		_, err := msgServer.AddSupportedToken(s.ctx, addMsg)
		s.Require().Error(err, "AddSupportedToken should fail for unauthorized address: %s", unauthorizedAddr)
		s.Require().Contains(err.Error(), "invalid authority")

		// Test RemoveSupportedToken with unauthorized address
		removeMsg := &types.MsgRemoveSupportedToken{
			Authority: unauthorizedAddr,
			Denom:     "ibc/ETH",
		}

		_, err = msgServer.RemoveSupportedToken(s.ctx, removeMsg)
		s.Require().Error(err, "RemoveSupportedToken should fail for unauthorized address: %s", unauthorizedAddr)
		s.Require().Contains(err.Error(), "invalid authority")
	}

	// Verify that only governance authority works
	govAuthority := s.App.AccountKeeper.GetModuleAddress("gov").String()

	// Test successful addition with governance authority
	addMsg := &types.MsgAddSupportedToken{
		Authority: govAuthority,
		Denom:     "ibc/GOVERNANCE_ONLY",
	}

	msgServer := keeper.NewMsgServerImpl(*s.keeper)
	_, err := msgServer.AddSupportedToken(s.ctx, addMsg)
	s.Require().NoError(err, "AddSupportedToken should succeed with governance authority")

	// Verify token was added
	params, err := s.keeper.GetParams(s.ctx)
	s.Require().NoError(err)

	found := false
	for _, token := range params.SupportedTokens {
		if token == "ibc/GOVERNANCE_ONLY" {
			found = true
			break
		}
	}
	s.Require().True(found, "Token should be added by governance")

	// Test successful removal with governance authority
	removeMsg := &types.MsgRemoveSupportedToken{
		Authority: govAuthority,
		Denom:     "ibc/GOVERNANCE_ONLY",
	}

	_, err = msgServer.RemoveSupportedToken(s.ctx, removeMsg)
	s.Require().NoError(err, "RemoveSupportedToken should succeed with governance authority")

	// Verify token was removed
	params, err = s.keeper.GetParams(s.ctx)
	s.Require().NoError(err)

	found = false
	for _, token := range params.SupportedTokens {
		if token == "ibc/GOVERNANCE_ONLY" {
			found = true
			break
		}
	}
	s.Require().False(found, "Token should be removed by governance")
}
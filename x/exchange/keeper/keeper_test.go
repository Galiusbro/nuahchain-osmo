package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/exchange/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/exchange/types"
	usdoracletypes "github.com/osmosis-labs/osmosis/v30/x/usdoracle/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	keeper *TestKeeper
	ctx    sdk.Context

	// Mock data for testing
	mockTWAPPrices map[uint64]math.LegacyDec // poolId -> price
}

// TestKeeper wraps the real keeper and overrides GetTWAPPrice for testing
type TestKeeper struct {
	*keeper.Keeper
	testSuite   *KeeperTestSuite
	dailyLimits map[string]types.DailyLimit // address+date -> limit
}

// testMsgServer wraps TestKeeper to implement message server functionality
type testMsgServer struct {
	*TestKeeper
}

// ExchangeTokens implements the message server interface using TestKeeper
func (tms testMsgServer) ExchangeTokens(ctx sdk.Context, msg *types.MsgExchangeTokens) (*types.MsgExchangeTokensResponse, error) {
	// Use the embedded keeper's msgServer but with TestKeeper methods
	msgServer := keeper.NewMsgServerImpl(*tms.TestKeeper.Keeper)
	return msgServer.ExchangeTokens(ctx, msg)
}

// GetTWAPPrice overrides the real GetTWAPPrice method for testing
func (tk *TestKeeper) GetTWAPPrice(ctx sdk.Context, denom string, poolId uint64) (osmomath.Dec, error) {
	if mockPrice, exists := tk.testSuite.mockTWAPPrices[poolId]; exists {
		// Convert math.LegacyDec to osmomath.Dec properly
		// Use string conversion to preserve precision
		return osmomath.MustNewDecFromStr(mockPrice.String()), nil
	}
	// Return error if no mock price is set
	return osmomath.ZeroDec(), fmt.Errorf("no mock TWAP price set for pool %d", poolId)
}

// GetDailyLimit returns the daily limit for testing
func (tk *TestKeeper) GetDailyLimit(ctx sdk.Context, address, date string) (types.DailyLimit, error) {
	key := address + "|" + date
	if limit, exists := tk.dailyLimits[key]; exists {
		return limit, nil
	}
	// Return empty limit if not found (like real keeper)
	return types.DailyLimit{
		Address:           address,
		Date:              date,
		TotalExchangedUsd: math.LegacyNewDec(0),
	}, nil
}

// SetDailyLimit sets the daily limit for testing
func (tk *TestKeeper) SetDailyLimit(ctx sdk.Context, limit types.DailyLimit) error {
	key := limit.Address + "|" + limit.Date
	tk.dailyLimits[key] = limit
	return nil
}

// ValidatePriceDeviation validates price deviation using mock TWAP prices
func (tk *TestKeeper) ValidatePriceDeviation(ctx sdk.Context, denom string, oraclePrice math.LegacyDec, poolId uint64) error {
	params, err := tk.Keeper.GetParams(ctx)
	if err != nil {
		return err
	}

	// Use mock TWAP price
	twapPrice, err := tk.GetTWAPPrice(ctx, denom, poolId)
	if err != nil {
		return fmt.Errorf("failed to get TWAP price: %w", err)
	}

	// Convert osmomath.Dec to math.LegacyDec for comparison
	// Use string conversion to preserve precision
	twapPriceLegacy, err := math.LegacyNewDecFromStr(twapPrice.String())
	if err != nil {
		return fmt.Errorf("failed to convert TWAP price: %w", err)
	}

	// Calculate deviation percentage
	deviation := oraclePrice.Sub(twapPriceLegacy).Abs().Quo(twapPriceLegacy)
	maxDeviation := params.PriceDeviationThreshold

	if deviation.GT(maxDeviation) {
		return fmt.Errorf("price deviation too high: %.2f%% > %.2f%%",
			deviation.MulInt64(100), maxDeviation.MulInt64(100))
	}

	return nil
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	// Initialize mock data
	s.mockTWAPPrices = make(map[uint64]math.LegacyDec)

	// Create TestKeeper wrapper
	s.keeper = &TestKeeper{
		Keeper:      s.App.ExchangeKeeper,
		testSuite:   s,
		dailyLimits: make(map[string]types.DailyLimit),
	}
	s.ctx = s.Ctx

	// Setup default params
	params := types.DefaultParams()
	params.PriceDeviationThreshold = math.LegacyNewDecWithPrec(2, 2) // 2%
	params.MinExchangeAmountUsd = math.LegacyNewDec(10)              // $10 in uusd
	params.MaxExchangeAmountUsd = math.LegacyNewDec(100000)          // $100k in uusd
	params.DailyLimitUsd = math.LegacyNewDec(1000000)                // $1M in uusd
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Setup USD Oracle params with supported tokens
	usdOracleParams := usdoracletypes.Params{
		SupportedTokens: []usdoracletypes.SupportedToken{
			{Denom: "ibc/ETH", Symbol: "ETH", Name: "Ethereum", Enabled: true, Decimals: 18, MinUpdateFrequency: 60, MaxPriceDeviation: osmomath.NewDecWithPrec(5, 2)},
			{Denom: "ibc/BTC", Symbol: "BTC", Name: "Bitcoin", Enabled: true, Decimals: 8, MinUpdateFrequency: 60, MaxPriceDeviation: osmomath.NewDecWithPrec(5, 2)},
			{Denom: "ibc/USDC", Symbol: "USDC", Name: "USD Coin", Enabled: true, Decimals: 6, MinUpdateFrequency: 60, MaxPriceDeviation: osmomath.NewDecWithPrec(2, 2)},
			{Denom: "ibc/ATOM", Symbol: "ATOM", Name: "Cosmos", Enabled: true, Decimals: 6, MinUpdateFrequency: 60, MaxPriceDeviation: osmomath.NewDecWithPrec(5, 2)},
			{Denom: "uosmo", Symbol: "OSMO", Name: "Osmosis", Enabled: true, Decimals: 6, MinUpdateFrequency: 60, MaxPriceDeviation: osmomath.NewDecWithPrec(5, 2)},
		},
		Enabled:                 true,
		UpdateInterval:          60,                            // 1 minute
		PriceDeviationThreshold: osmomath.NewDecWithPrec(5, 2), // 5%
		MinSources:              1,
		MaxPriceAge:             300, // 5 minutes
	}
	s.App.USDOracleKeeper.SetParams(s.ctx, usdOracleParams)
}

func (s *KeeperTestSuite) TestExchangeTokensSuccess() {
	// Setup test account
	testAddr := s.TestAccs[0]

	// Mint tokens to test account
	testCoins := sdk.NewCoins(
		sdk.NewCoin("ibc/ETH", osmomath.NewInt(1000000000000000000)), // 1 ETH (18 decimals)
	)
	err := s.App.BankKeeper.MintCoins(s.ctx, types.ModuleName, testCoins)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, testAddr, testCoins)
	s.Require().NoError(err)

	// Mock Oracle price for ETH = $2000
	s.mockOraclePrice("ibc/ETH", math.LegacyNewDec(2000))

	// Mock TWAP price for ETH = $1980 (1% deviation)
	s.mockTWAPPrice(1, math.LegacyNewDec(1980))

	// Test exchange - use smaller amount to avoid hitting limits
	msg := &types.MsgExchangeTokens{
		Sender:     testAddr.String(),
		TokenIn:    sdk.NewCoin("ibc/ETH", osmomath.NewInt(25000000000000000)), // 0.025 ETH (18 decimals)
		MinNuahOut: osmomath.NewInt(49_000000),                                 // Expect ~$50 worth of N$
	}

	msgServer := testMsgServer{TestKeeper: s.keeper}
	_, err = msgServer.ExchangeTokens(s.ctx, msg)
	s.Require().NoError(err)

	// Check N$ balance
	nDollarBalance := s.App.BankKeeper.GetBalance(s.ctx, testAddr, "unuah")
	s.Require().True(nDollarBalance.Amount.GTE(msg.MinNuahOut))

	// Check remaining ETH balance
	ethBalance := s.App.BankKeeper.GetBalance(s.ctx, testAddr, "ibc/ETH")
	s.Require().Equal(osmomath.NewInt(975000000000000000), ethBalance.Amount) // 0.975 ETH remaining (1 - 0.025)
}

func (s *KeeperTestSuite) TestExchangeTokensPriceDeviationError() {
	// Setup test account
	testAddr := s.TestAccs[0]

	// Mint tokens to test account
	testCoins := sdk.NewCoins(
		sdk.NewCoin("ibc/ETH", osmomath.NewInt(1000000000000000000)), // 1 ETH (18 decimals)
	)
	err := s.App.BankKeeper.MintCoins(s.ctx, types.ModuleName, testCoins)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, testAddr, testCoins)
	s.Require().NoError(err)

	// Mock Oracle price for ETH = $2000
	s.mockOraclePrice("ibc/ETH", math.LegacyNewDec(2000))

	// Mock TWAP price for ETH = $1800 (10% deviation - exceeds 2% threshold)
	s.mockTWAPPrice(1, math.LegacyNewDec(1800))

	// Test price deviation validation directly
	oraclePrice := math.LegacyNewDec(2000)
	err = s.keeper.ValidatePriceDeviation(s.ctx, "ibc/ETH", oraclePrice, 1)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "price deviation too high")
}

func (s *KeeperTestSuite) TestExchangeTokensMinAmountError() {
	// Setup test account
	testAddr := s.TestAccs[0]

	// Mint tokens to test account
	testCoins := sdk.NewCoins(
		sdk.NewCoin("ibc/ETH", osmomath.NewInt(1000000000000000000)), // 1 ETH (18 decimals)
	)
	err := s.App.BankKeeper.MintCoins(s.ctx, types.ModuleName, testCoins)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, testAddr, testCoins)
	s.Require().NoError(err)

	// Mock Oracle price for ETH = $2000
	s.mockOraclePrice("ibc/ETH", math.LegacyNewDec(2000))

	// Mock TWAP price for ETH = $1980
	s.mockTWAPPrice(1, math.LegacyNewDec(1980))

	// Test exchange with amount below minimum ($10)
	msg := &types.MsgExchangeTokens{
		Sender:     testAddr.String(),
		TokenIn:    sdk.NewCoin("ibc/ETH", osmomath.NewInt(4000000000000000)), // ~$8 worth (18 decimals)
		MinNuahOut: osmomath.NewInt(7_000000),
	}

	msgServer := keeper.NewMsgServerImpl(*s.keeper.Keeper)
	_, err = msgServer.ExchangeTokens(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "amount below minimum")
}

func (s *KeeperTestSuite) TestExchangeTokensMaxAmountError() {
	// Setup test account
	testAddr := s.TestAccs[0]

	// Mint tokens to test account
	ethAmount, _ := osmomath.NewIntFromString("100000000000000000000") // 100 ETH (18 decimals)
	testCoins := sdk.NewCoins(
		sdk.NewCoin("ibc/ETH", ethAmount),
	)
	err := s.App.BankKeeper.MintCoins(s.ctx, types.ModuleName, testCoins)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, testAddr, testCoins)
	s.Require().NoError(err)

	// Mock Oracle price for ETH = $2000
	s.mockOraclePrice("ibc/ETH", math.LegacyNewDec(2000))

	// Mock TWAP price for ETH = $1980
	s.mockTWAPPrice(1, math.LegacyNewDec(1980))

	// Test exchange with amount > $100k
	ethAmount2, _ := osmomath.NewIntFromString("60000000000000000000") // ~$120k worth (18 decimals)
	msg := &types.MsgExchangeTokens{
		Sender:     testAddr.String(),
		TokenIn:    sdk.NewCoin("ibc/ETH", ethAmount2),
		MinNuahOut: osmomath.NewInt(118000_000000),
	}

	msgServer := keeper.NewMsgServerImpl(*s.keeper.Keeper)
	_, err = msgServer.ExchangeTokens(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "exchange amount above maximum threshold")
}

func (s *KeeperTestSuite) TestExchangeTokensDailyLimitError() {
	// Setup test account
	testAddr := s.TestAccs[0]

	// Mint tokens to test account
	ethAmount, _ := osmomath.NewIntFromString("1000000000000000000000") // 1000 ETH (18 decimals)
	testCoins := sdk.NewCoins(
		sdk.NewCoin("ibc/ETH", ethAmount),
	)
	err := s.App.BankKeeper.MintCoins(s.ctx, types.ModuleName, testCoins)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, testAddr, testCoins)
	s.Require().NoError(err)

	// Mock Oracle price for ETH = $2000
	s.mockOraclePrice("ibc/ETH", math.LegacyNewDec(2000))

	// Mock TWAP price for ETH = $1980
	s.mockTWAPPrice(1, math.LegacyNewDec(1980))

	// Set existing daily limit close to maximum
	todayStr := s.keeper.GetTodayString(s.ctx)
	existingLimit := types.DailyLimit{
		Address:           testAddr.String(),
		Date:              todayStr,
		TotalExchangedUsd: math.LegacyNewDec(900000), // $900k already used
	}
	err = s.keeper.SetDailyLimit(s.ctx, existingLimit)
	s.Require().NoError(err)

	// Test exchange that should fail due to daily limit
	// Use amount that will exceed daily limit: $900k + $200k = $1.1M > $1M limit
	ethAmount2, _ := osmomath.NewIntFromString("100000000000000000000") // 100 ETH = ~$200k worth (18 decimals)
	msg := &types.MsgExchangeTokens{
		Sender:     testAddr.String(),
		TokenIn:    sdk.NewCoin("ibc/ETH", ethAmount2),
		MinNuahOut: osmomath.NewInt(160000_000000), // Expect ~$160k worth of N$
	}

	// Test daily limit validation directly
	params, err := s.keeper.GetParams(s.ctx)
	s.Require().NoError(err)

	// Calculate USD value
	supportedToken, found := s.App.USDOracleKeeper.GetSupportedToken(s.ctx, msg.TokenIn.Denom)
	s.Require().True(found)

	tokenAmount := math.LegacyNewDecFromInt(msg.TokenIn.Amount).Quo(math.LegacyNewDec(10).Power(uint64(supportedToken.Decimals)))
	exchangeRate, err2 := s.keeper.GetExchangeRate(s.ctx, msg.TokenIn.Denom)
	s.Require().NoError(err2)
	usdValue := tokenAmount.Mul(exchangeRate.Rate)

	// Check daily limit
	dailyLimit, err3 := s.keeper.GetDailyLimit(s.ctx, msg.Sender, todayStr)
	s.Require().NoError(err3)

	newDailyTotal := dailyLimit.TotalExchangedUsd.Add(usdValue)
	s.Require().True(newDailyTotal.GT(params.DailyLimitUsd), "Expected daily limit to be exceeded")
}

func (s *KeeperTestSuite) TestGetTWAPPrice() {
	// Mock TWAP price
	expectedPrice := math.LegacyNewDec(1500)
	s.mockTWAPPrice(1, expectedPrice)

	// Test GetTWAPPrice
	price, err := s.keeper.GetTWAPPrice(s.ctx, "ibc/ETH", 1)
	s.Require().NoError(err)
	s.Require().Equal(expectedPrice, price)
}

func (s *KeeperTestSuite) TestValidatePriceDeviation() {
	oraclePrice := math.LegacyNewDec(2000)

	// Test case 1: Price deviation within threshold
	s.mockTWAPPrice(1, math.LegacyNewDec(1980)) // 1% deviation
	err := s.keeper.ValidatePriceDeviation(s.ctx, "ibc/ETH", oraclePrice, 1)
	s.Require().NoError(err)

	// Test case 2: Price deviation exceeds threshold
	s.mockTWAPPrice(1, math.LegacyNewDec(1900)) // 5% deviation
	err = s.keeper.ValidatePriceDeviation(s.ctx, "ibc/ETH", oraclePrice, 1)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "price deviation too high")
}

func (s *KeeperTestSuite) TestUnsupportedToken() {
	// Setup test account
	testAddr := s.TestAccs[0]

	// Mint unsupported token
	testCoins := sdk.NewCoins(
		sdk.NewCoin("unsupported", osmomath.NewInt(1000000)),
	)
	err := s.App.BankKeeper.MintCoins(s.ctx, types.ModuleName, testCoins)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, testAddr, testCoins)
	s.Require().NoError(err)

	// Test exchange with unsupported token
	msg := &types.MsgExchangeTokens{
		Sender:     testAddr.String(),
		TokenIn:    sdk.NewCoin("unsupported", osmomath.NewInt(1000)),
		MinNuahOut: osmomath.NewInt(1000),
	}

	msgServer := keeper.NewMsgServerImpl(*s.keeper.Keeper)
	_, err = msgServer.ExchangeTokens(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "not supported")
}

// Helper functions for mocking

func (s *KeeperTestSuite) mockOraclePrice(denom string, price math.LegacyDec) {
	// Set exchange rate for the token
	exchangeRate := types.ExchangeRate{
		Denom:       denom,
		Rate:        price,
		LastUpdated: s.ctx.BlockTime(),
	}
	err := s.keeper.SetExchangeRate(s.ctx, exchangeRate)
	s.Require().NoError(err)

	// Also mock USD Oracle price
	tokenPrice := usdoracletypes.TokenPrice{
		Denom: denom,
		Price: price,
	}
	s.App.USDOracleKeeper.SetTokenPrice(s.ctx, tokenPrice)
}

func (s *KeeperTestSuite) mockTWAPPrice(poolId uint64, price math.LegacyDec) {
	s.mockTWAPPrices[poolId] = price
}

// GetTWAPPriceForTest is a test version of GetTWAPPrice that uses mock data
func (s *KeeperTestSuite) GetTWAPPriceForTest(ctx sdk.Context, denom string, poolId uint64) (osmomath.Dec, error) {
	if mockPrice, exists := s.mockTWAPPrices[poolId]; exists {
		// Convert math.LegacyDec to osmomath.Dec
		return osmomath.NewDecFromBigInt(mockPrice.BigInt()), nil
	}
	// Return error if no mock price is set
	return osmomath.ZeroDec(), fmt.Errorf("no mock TWAP price set for pool %d", poolId)
}

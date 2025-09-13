package usdoracle_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/usdoracle/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/usdoracle/types"
)

type IntegrationTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func (s *IntegrationTestSuite) SetupTest() {
	s.Setup()
	s.queryClient = types.NewQueryClient(s.QueryHelper)
	s.msgServer = keeper.NewMsgServerImpl(*s.App.USDOracleKeeper)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) TestFullPriceUpdateFlow() {
	s.SetupTest()

	// Test setting up price sources
	source1 := types.PriceSource{
		Name:    "coinbase",
		Weight:  math.LegacyNewDec(1),
		Enabled: true,
		Url:     "https://api.coinbase.com/v2/exchange-rates?currency=USD",
	}

	source2 := types.PriceSource{
		Name:    "binance",
		Weight:  math.LegacyNewDec(1),
		Enabled: true,
		Url:     "https://api.binance.com/api/v3/ticker/price?symbol=USDUSDT",
	}

	// Set price sources
	s.App.USDOracleKeeper.SetPriceSource(s.Ctx, source1)
	s.App.USDOracleKeeper.SetPriceSource(s.Ctx, source2)

	// Verify sources are set
	allSources := s.App.USDOracleKeeper.GetAllPriceSources(s.Ctx)
	s.Require().Len(allSources, 2)

	// Test price updates
	price1 := types.USDPrice{
		Price:       math.LegacyNewDec(100),
		Timestamp:   time.Now(),
		Source:      "coinbase",
		BlockHeight: s.Ctx.BlockHeight(),
	}

	price2 := types.USDPrice{
		Price:       math.LegacyNewDecWithPrec(10005, 2), // 100.05
		Timestamp:   time.Now().Add(time.Second),
		Source:      "binance",
		BlockHeight: s.Ctx.BlockHeight(),
	}

	// Add price history
	s.App.USDOracleKeeper.AddPriceHistory(s.Ctx, price1)
	s.App.USDOracleKeeper.AddPriceHistory(s.Ctx, price2)

	// Set current price
	s.App.USDOracleKeeper.SetCurrentPrice(s.Ctx, price2)

	// Verify current price
	currentPrice, found := s.App.USDOracleKeeper.GetCurrentPrice(s.Ctx)
	s.Require().True(found)
	s.Require().Equal(price2, currentPrice)

	// Test price history
	history := s.App.USDOracleKeeper.GetPriceHistoryList(s.Ctx, 10)
	s.Require().Len(history, 2)
}

func (s *IntegrationTestSuite) TestQueryCurrentPrice() {
	s.SetupTest()

	// Set a current price
	price := types.USDPrice{
		Price:       math.LegacyNewDec(100),
		Timestamp:   time.Now(),
		Source:      "test-source",
		BlockHeight: s.Ctx.BlockHeight(),
	}
	s.App.USDOracleKeeper.SetCurrentPrice(s.Ctx, price)

	// Query current price
	resp, err := s.queryClient.GetUSDPrice(sdk.WrapSDKContext(s.Ctx), &types.QueryGetUSDPriceRequest{})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal(price, resp.Price)
}

func (s *IntegrationTestSuite) TestQueryPriceHistory() {
	s.SetupTest()

	// Add some price history
	for i := 0; i < 5; i++ {
		price := types.USDPrice{
			Price:       math.LegacyNewDec(int64(100 + i)),
			Timestamp:   time.Now().Add(time.Duration(i) * time.Second),
			Source:      "test-source",
			BlockHeight: s.Ctx.BlockHeight() + int64(i),
		}
		s.App.USDOracleKeeper.AddPriceHistory(s.Ctx, price)
	}

	// Query price history
	resp, err := s.queryClient.GetPriceHistory(sdk.WrapSDKContext(s.Ctx), &types.QueryGetPriceHistoryRequest{
		Pagination: &query.PageRequest{Limit: 3},
	})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Len(resp.Prices, 3)
}

func (s *IntegrationTestSuite) TestQueryPriceSources() {
	s.SetupTest()

	// Add price sources
	sources := []types.PriceSource{
		{
			Name:    "source1",
			Weight:  math.LegacyNewDec(1),
			Enabled: true,
			Url:     "https://api1.test.com",
		},
		{
			Name:    "source2",
			Weight:  math.LegacyNewDec(2),
			Enabled: false,
			Url:     "https://api2.test.com",
		},
	}

	for _, source := range sources {
		s.App.USDOracleKeeper.SetPriceSource(s.Ctx, source)
	}

	// Query all price sources
	// Note: PriceSources query is not available in the current implementation
	// We'll test individual source retrieval instead
	source1Retrieved, found := s.App.USDOracleKeeper.GetPriceSource(s.Ctx, "source1")
	s.Require().True(found)
	s.Require().Equal(sources[0], source1Retrieved)

}

func (s *IntegrationTestSuite) TestQueryParams() {
	s.SetupTest()

	// Set custom params
	params := types.Params{
		Enabled:                 true,
		Admin:                   "test-admin",
		UpdateInterval:          120,
		PriceDeviationThreshold: math.LegacyNewDecWithPrec(10, 2), // 10%
	}
	s.App.USDOracleKeeper.SetParams(s.Ctx, params)

	// Query params
	resp, err := s.queryClient.Params(sdk.WrapSDKContext(s.Ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal(params, resp.Params)
}

func (s *IntegrationTestSuite) TestPriceDeviationCalculation() {
	s.SetupTest()

	// Set initial price
	initialPrice := types.USDPrice{
		Price:       math.LegacyNewDec(100),
		Timestamp:   time.Now(),
		Source:      "test-source",
		BlockHeight: s.Ctx.BlockHeight(),
	}
	s.App.USDOracleKeeper.SetCurrentPrice(s.Ctx, initialPrice)

	// Test deviation calculation
	deviation, found := s.App.USDOracleKeeper.CalculatePriceDeviation(s.Ctx)
	s.Require().True(found)
	s.Require().True(deviation.GTE(math.LegacyZeroDec()))

	// Test threshold checking
	withinThreshold := s.App.USDOracleKeeper.IsWithinThreshold(s.Ctx)
	s.Require().True(withinThreshold)
}

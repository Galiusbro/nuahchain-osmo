package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	"github.com/osmosis-labs/osmosis/v30/x/usdoracle/testutil"
	"github.com/osmosis-labs/osmosis/v30/x/usdoracle/types"
)

type KeeperTestSuite struct {
	testutil.UsdOracleTestSuite
}

func TestKeeperSuite(t *testing.T) {
	testutil.TestKeeperTestSuite(t)
}

func (s *KeeperTestSuite) TestSetGetCurrentPrice() {
	s.SetupTest()

	// Test setting and getting current price
	price := types.USDPrice{
		Price:       math.LegacyNewDec(100),
		Timestamp:   time.Now(),
		Source:      "test-source",
		BlockHeight: 1,
	}
	s.Keeper.SetCurrentPrice(s.Ctx, price)

	retrievedPrice, found := s.Keeper.GetCurrentPrice(s.Ctx)
	require.True(s.T(), found)
	require.Equal(s.T(), price, retrievedPrice)
}

func (s *KeeperTestSuite) TestAddPriceHistory() {
	s.SetupTest()

	// Test adding price history
	price1 := types.USDPrice{
		Price:       math.LegacyNewDec(100),
		Timestamp:   time.Now(),
		Source:      "test-source-1",
		BlockHeight: 1,
	}
	price2 := types.USDPrice{
		Price:       math.LegacyNewDec(101),
		Timestamp:   time.Now().Add(time.Second),
		Source:      "test-source-2",
		BlockHeight: 2,
	}

	s.Keeper.AddPriceHistory(s.Ctx, price1)
	s.Keeper.AddPriceHistory(s.Ctx, price2)

	history := s.Keeper.GetPriceHistoryList(s.Ctx, 10)
	require.Len(s.T(), history, 2)
	require.Equal(s.T(), price1, history[0])
	require.Equal(s.T(), price2, history[1])
}

func (s *KeeperTestSuite) TestSetGetPriceSource() {
	s.SetupTest()

	// Test setting and getting price source
	source := types.PriceSource{
		Name:    "test-source",
		Weight:  math.LegacyNewDec(1),
		Enabled: true,
		Url:     "https://api.test.com",
	}

	s.Keeper.SetPriceSource(s.Ctx, source)

	retrievedSource, found := s.Keeper.GetPriceSource(s.Ctx, "test-source")
	require.True(s.T(), found)
	require.Equal(s.T(), source, retrievedSource)
}

func (s *KeeperTestSuite) TestGetAllPriceSources() {
	s.SetupTest()

	// Test getting all price sources
	source1 := types.PriceSource{
		Name:    "source1",
		Weight:  math.LegacyNewDec(1),
		Enabled: true,
		Url:     "https://api1.test.com",
	}
	source2 := types.PriceSource{
		Name:    "source2",
		Weight:  math.LegacyNewDec(2),
		Enabled: false,
		Url:     "https://api2.test.com",
	}

	s.Keeper.SetPriceSource(s.Ctx, source1)
	s.Keeper.SetPriceSource(s.Ctx, source2)

	allSources := s.Keeper.GetAllPriceSources(s.Ctx)
	require.Len(s.T(), allSources, 2)
}

func (s *KeeperTestSuite) TestParams() {
	s.SetupTest()

	// Test setting and getting params
	params := types.Params{
		Enabled:                 true,
		Admin:                   "test-admin",
		UpdateInterval:          60,
		PriceDeviationThreshold: math.LegacyNewDecWithPrec(5, 2),
	}

	s.Keeper.SetParams(s.Ctx, params)

	retrievedParams := s.Keeper.GetParams(s.Ctx)
	require.Equal(s.T(), params, retrievedParams)
}

func (s *KeeperTestSuite) TestCalculatePriceDeviation() {
	s.SetupTest()

	// Set current price
	currentPrice := types.USDPrice{
		Price:       math.LegacyNewDec(100),
		Timestamp:   time.Now(),
		Source:      "test-source",
		BlockHeight: 1,
	}
	s.Keeper.SetCurrentPrice(s.Ctx, currentPrice)

	// Test deviation calculation
	deviation, found := s.Keeper.CalculatePriceDeviation(s.Ctx)
	require.True(s.T(), found)
	require.True(s.T(), deviation.GTE(math.LegacyZeroDec()))
}

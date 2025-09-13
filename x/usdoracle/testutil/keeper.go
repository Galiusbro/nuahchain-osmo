package testutil

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/usdoracle/keeper"
)

// UsdOracleTestSuite is a test suite for the USD Oracle module
type UsdOracleTestSuite struct {
	apptesting.KeeperTestHelper

	Keeper *keeper.Keeper
}

// SetupTest sets up the test environment
func (s *UsdOracleTestSuite) SetupTest() {
	s.Setup()
	s.Keeper = s.App.USDOracleKeeper
}

// TestKeeperTestSuite runs the test suite
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UsdOracleTestSuite))
}

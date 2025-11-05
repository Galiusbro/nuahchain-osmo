package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/usertoken/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

type TestSuite struct {
	apptesting.KeeperTestHelper
}

func TestCreateTokenDirect(t *testing.T) {
	// Setup test environment similar to tests
	testSuite := &TestSuite{}
	testSuite.SetT(t)
	testSuite.Setup()
	testSuite.Ctx = testSuite.Ctx.WithBlockTime(time.Unix(1_700_000_000, 0))

	// Ensure we have enough test accounts
	for len(testSuite.TestAccs) < 5 {
		extra := apptesting.CreateRandomAccounts(1)[0]
		testSuite.TestAccs = append(testSuite.TestAccs, extra)
	}

	// Set params like in tests
	params := testSuite.App.UserTokenKeeper.GetParams(testSuite.Ctx)
	params.BondingCurveWallet = testSuite.TestAccs[1].String()
	params.PlatformWallet = testSuite.TestAccs[2].String()
	params.ReferralWallet = testSuite.TestAccs[3].String()
	params.AiCeoWallet = testSuite.TestAccs[4].String()
	params.FounderClaimPeriod = 3600
	testSuite.App.UserTokenKeeper.SetParams(testSuite.Ctx, params)

	// Create msg server
	msgServer := keeper.NewMsgServerImpl(*testSuite.App.UserTokenKeeper)

	// Create token
	creator := testSuite.TestAccs[0]
	resp, err := msgServer.CreateToken(testSuite.Ctx, types.NewMsgCreateToken(
		creator.String(),
		"Direct Test Token",
		"DTT",
		"https://example.com/token.png",
		"A token created directly via keeper",
	))

	if err != nil {
		t.Fatalf("Error creating token: %v", err)
	}

	// Output result
	result := map[string]interface{}{
		"success": true,
		"denom":   resp.Denom,
		"creator": creator.String(),
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonBytes))
}

func main() {
	// Run as a test
	testing.Main(func(pat, str string) (bool, error) {
		return true, nil
	}, []testing.InternalTest{
		{Name: "TestCreateTokenDirect", F: TestCreateTokenDirect},
	}, []testing.InternalBenchmark{}, []testing.InternalExample{})
}

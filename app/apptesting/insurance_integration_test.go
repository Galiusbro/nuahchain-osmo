package apptesting

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v30/app/params"
	claimstypes "github.com/osmosis-labs/osmosis/v30/x/claims/types"
	policytypes "github.com/osmosis-labs/osmosis/v30/x/policy/types"
	premiumtypes "github.com/osmosis-labs/osmosis/v30/x/premium/types"
	rolestypes "github.com/osmosis-labs/osmosis/v30/x/roles/types"
	treasurytypes "github.com/osmosis-labs/osmosis/v30/x/treasury/types"
)

type InsuranceKeeperTestSuite struct {
	KeeperTestHelper
}

func TestInsuranceKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(InsuranceKeeperTestSuite))
}

func (s *InsuranceKeeperTestSuite) SetupTest() {
	s.Setup()
}

func (s *InsuranceKeeperTestSuite) TestLifecycle() {
	ctx := s.Ctx.WithBlockTime(time.Now().UTC())
	s.Ctx = ctx

	authority := s.TestAccs[0]
	recipient := s.TestAccs[1]

	baseCoin := sdk.NewInt64Coin(params.BaseCoinUnit, 2_000_000)
	s.FundAcc(authority, sdk.NewCoins(baseCoin))
	s.FundAcc(recipient, sdk.NewCoins(baseCoin))

	rolesKeeper := s.App.RolesKeeper
	rolesParams := rolestypes.NewParams(authority.String())
	rolesKeeper.SetParams(ctx, rolesParams)

	treasuryKeeper := s.App.TreasuryKeeper
	treasuryParams := treasurytypes.NewParams(authority.String(), "", false)
	treasuryKeeper.SetParams(ctx, treasuryParams)

	premiumKeeper := s.App.PremiumKeeper
	premiumKeeper.SetParams(ctx, premiumtypes.NewParams([]string{params.BaseCoinUnit}, 0, false))

	claimsKeeper := s.App.ClaimsKeeper
	claimsKeeper.SetParams(ctx, claimstypes.Params{MaxOpenClaimsPerPolicy: 5})

	// Assign insurer, treasury manager, and claims reviewer roles to authority
	err := rolesKeeper.AssignRoles(ctx, authority, authority, []rolestypes.Role{
		rolestypes.Role_ROLE_INSURER,
		rolestypes.Role_ROLE_TREASURY_MANAGER,
		rolestypes.Role_ROLE_CLAIMS_REVIEWER,
	})
	s.Require().NoError(err)

	poolID := "pool-test"
	err = treasuryKeeper.CreateTreasuryPool(ctx, authority, treasurytypes.TreasuryPool{
		Id:          poolID,
		Description: "test pool",
		Manager:     authority.String(),
		PolicyTypes: []string{"custom"},
	})
	s.Require().NoError(err)

	deposit := sdk.NewInt64Coin(params.BaseCoinUnit, 1_000_000)
	err = treasuryKeeper.DepositToTreasury(ctx, authority, poolID, deposit)
	s.Require().NoError(err)

	policyKeeper := s.App.PolicyKeeper
	start := time.Now().UTC()
	policy, err := policyKeeper.CreatePolicy(ctx, authority, "custom", []policytypes.PolicyAttribute{
		{Key: "coverage", Value: "basic"},
	}, &start, nil, poolID, []string{"integration"})
	s.Require().NoError(err)

	schedule := premiumtypes.PremiumSchedule{ScheduleType: "periodic", PeriodSeconds: 3600, MaxPayments: 0}
	plan, err := premiumKeeper.CreatePremiumPlan(ctx, authority, policy.Id, authority, sdk.NewInt64Coin(params.BaseCoinUnit, 50_000), schedule, poolID)
	s.Require().NoError(err)

	_, err = premiumKeeper.RecordPremiumPayment(ctx, authority, plan.Id, sdk.NewInt64Coin(params.BaseCoinUnit, 50_000))
	s.Require().NoError(err)

	claim, err := claimsKeeper.SubmitClaim(ctx, authority, policy.Id, sdk.NewInt64Coin(params.BaseCoinUnit, 25_000), "test claim", nil)
	s.Require().NoError(err)

	_, err = claimsKeeper.ReviewClaim(ctx, authority, claim.Id, claimstypes.ClaimStatus_CLAIM_STATUS_APPROVED, "approved")
	s.Require().NoError(err)

	_, err = claimsKeeper.ExecuteClaimPayout(ctx, authority, claim.Id, recipient)
	s.Require().NoError(err)

	updatedClaim, found := claimsKeeper.GetClaim(ctx, claim.Id)
	s.Require().True(found)
	s.Require().Equal(claimstypes.ClaimStatus_CLAIM_STATUS_PAID, updatedClaim.Status)
}

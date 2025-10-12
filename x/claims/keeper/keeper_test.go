package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/claims/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/claims/types"
	policytypes "github.com/osmosis-labs/osmosis/v30/x/policy/types"
	rolestypes "github.com/osmosis-labs/osmosis/v30/x/roles/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	keeper         keeper.Keeper
	rolesKeeper    *MockRolesKeeper
	policyKeeper   *MockPolicyKeeper
	treasuryKeeper *MockTreasuryKeeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	// Create mock keepers
	s.rolesKeeper = NewMockRolesKeeper()
	s.policyKeeper = NewMockPolicyKeeper()
	s.treasuryKeeper = NewMockTreasuryKeeper()

	// Create claims keeper with mocks
	s.keeper = keeper.NewKeeper(
		s.App.AppCodec(),
		s.App.GetKey(types.StoreKey),
		s.App.GetSubspace(types.ModuleName),
		s.rolesKeeper,
		s.policyKeeper,
		s.treasuryKeeper,
	)

	// Set up default parameters
	params := types.DefaultParams()
	params.MaxOpenClaimsPerPolicy = 5
	params.AutoApprovalPolicyTypes = []string{} // No auto-approval by default
	s.keeper.SetParams(s.Ctx, params)

}

func (s *KeeperTestSuite) TestSubmitClaim() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	claimant := s.TestAccs[0]
	policyID := uint64(1)
	amount := sdk.NewCoin("unuah", math.NewInt(1000000))
	description := "Test claim for car accident"
	evidence := []types.ClaimEvidence{
		{Uri: "https://photos.com/accident1.jpg", Notes: "Photo of damaged car"},
		{Uri: "https://docs.com/police_report.pdf", Notes: "Police report"},
	}

	// Set up mock policy
	s.policyKeeper.SetPolicy(policyID, policytypes.Policy{
		Id:             policyID,
		Owner:          claimant.String(),
		PolicyType:     "auto",
		Status:         policytypes.PolicyStatus_POLICY_STATUS_ACTIVE,
		TreasuryPoolId: "pool-001",
	})

	// Submit claim
	claim, err := keeper.SubmitClaim(ctx, claimant, policyID, amount, description, evidence)
	s.Require().NoError(err)
	s.Require().NotNil(claim)

	// Verify claim details
	s.Require().Greater(claim.Id, uint64(0), "Claim ID should be positive")
	s.Require().Equal(uint64(1), claim.Id, "First claim should have ID 1, but got %d", claim.Id)
	s.Require().Equal(policyID, claim.PolicyId)
	s.Require().Equal(claimant.String(), claim.Claimant)
	s.Require().Equal(amount, claim.Amount)
	s.Require().Equal(description, claim.Description)

	s.Require().Equal(types.ClaimStatus_CLAIM_STATUS_PENDING, claim.Status)

	s.Require().Len(claim.Evidence, 2)
	s.Require().Equal("pool-001", claim.TreasuryPoolId)

	// Verify evidence
	for i, expectedEvidence := range evidence {
		s.Require().Equal(expectedEvidence.Uri, claim.Evidence[i].Uri)
		s.Require().Equal(expectedEvidence.Notes, claim.Evidence[i].Notes)
	}
}

func (s *KeeperTestSuite) TestSubmitClaimWithAutoApproval() {
	ctx := s.Ctx
	keeper := s.keeper

	// Enable auto-approval for this test
	params := s.keeper.GetParams(ctx)
	params.AutoApprovalPolicyTypes = []string{"auto"}
	s.keeper.SetParams(ctx, params)

	// Test data
	claimant := s.TestAccs[0]
	policyID := uint64(1)
	amount := sdk.NewCoin("unuah", math.NewInt(1000000))
	description := "Auto claim"

	// Set up mock policy with auto-approval type
	s.policyKeeper.SetPolicy(policyID, policytypes.Policy{
		Id:             policyID,
		Owner:          claimant.String(),
		PolicyType:     "auto", // This type is in auto-approval list
		Status:         policytypes.PolicyStatus_POLICY_STATUS_ACTIVE,
		TreasuryPoolId: "pool-001",
	})

	// Submit claim
	claim, err := keeper.SubmitClaim(ctx, claimant, policyID, amount, description, nil)
	s.Require().NoError(err)

	// Verify auto-approval
	s.Require().Equal(types.ClaimStatus_CLAIM_STATUS_APPROVED, claim.Status)
	s.Require().NotNil(claim.Decision)
	s.Require().Equal("auto", claim.Decision.Reviewer)
	s.Require().Equal(types.ClaimStatus_CLAIM_STATUS_APPROVED, claim.Decision.Status)
	s.Require().Equal("auto-approved", claim.Decision.Reason)
}

func (s *KeeperTestSuite) TestSubmitClaimInvalidPolicy() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	claimant := s.TestAccs[0]
	policyID := uint64(999) // Non-existent policy
	amount := sdk.NewCoin("unuah", math.NewInt(1000000))

	// Submit claim with non-existent policy
	_, err := keeper.SubmitClaim(ctx, claimant, policyID, amount, "Test claim", nil)
	s.Require().Error(err)
	s.Require().Equal(types.ErrInvalidPolicy, err)
}

func (s *KeeperTestSuite) TestSubmitClaimInactivePolicy() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	claimant := s.TestAccs[0]
	policyID := uint64(1)
	amount := sdk.NewCoin("unuah", math.NewInt(1000000))

	// Set up mock policy with inactive status
	s.policyKeeper.SetPolicy(policyID, policytypes.Policy{
		Id:             policyID,
		Owner:          claimant.String(),
		PolicyType:     "auto",
		Status:         policytypes.PolicyStatus_POLICY_STATUS_CANCELLED, // Inactive
		TreasuryPoolId: "pool-001",
	})

	// Submit claim
	_, err := keeper.SubmitClaim(ctx, claimant, policyID, amount, "Test claim", nil)
	s.Require().Error(err)
	s.Require().Equal(types.ErrInvalidPolicy, err)
}

func (s *KeeperTestSuite) TestSubmitClaimInvalidAmount() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	claimant := s.TestAccs[0]
	policyID := uint64(1)
	zeroAmount := sdk.NewCoin("unuah", math.NewInt(0))                     // Zero amount
	negativeAmount := sdk.Coin{Denom: "unuah", Amount: math.NewInt(-1000)} // Negative amount (invalid)

	// Test zero amount
	_, err := keeper.SubmitClaim(ctx, claimant, policyID, zeroAmount, "Test claim", nil)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "claim amount must be positive")

	// Test negative amount
	_, err = keeper.SubmitClaim(ctx, claimant, policyID, negativeAmount, "Test claim", nil)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "claim amount must be positive")
}

func (s *KeeperTestSuite) TestReviewClaim() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	claimant := s.TestAccs[0]
	reviewer := s.TestAccs[1]
	policyID := uint64(1)
	amount := sdk.NewCoin("unuah", math.NewInt(1000000))

	// Set up mock policy
	s.policyKeeper.SetPolicy(policyID, policytypes.Policy{
		Id:             policyID,
		Owner:          claimant.String(),
		PolicyType:     "auto",
		Status:         policytypes.PolicyStatus_POLICY_STATUS_ACTIVE,
		TreasuryPoolId: "pool-001",
	})

	// Set up reviewer role
	s.rolesKeeper.SetHasRole(true)

	// Submit claim first
	claim, err := keeper.SubmitClaim(ctx, claimant, policyID, amount, "Test claim", nil)
	s.Require().NoError(err)

	// Review claim - approve
	updatedClaim, err := keeper.ReviewClaim(ctx, reviewer, claim.Id, types.ClaimStatus_CLAIM_STATUS_APPROVED, "Approved based on evidence")
	s.Require().NoError(err)
	s.Require().Equal(types.ClaimStatus_CLAIM_STATUS_APPROVED, updatedClaim.Status)
	s.Require().NotNil(updatedClaim.Decision)
	s.Require().Equal(reviewer.String(), updatedClaim.Decision.Reviewer)
	s.Require().Equal("Approved based on evidence", updatedClaim.Decision.Reason)

	// Review claim - reject (submit new claim first)
	claim2, err := keeper.SubmitClaim(ctx, claimant, policyID, amount, "Test claim 2", nil)
	s.Require().NoError(err)

	updatedClaim2, err := keeper.ReviewClaim(ctx, reviewer, claim2.Id, types.ClaimStatus_CLAIM_STATUS_REJECTED, "Insufficient evidence")
	s.Require().NoError(err)
	s.Require().Equal(types.ClaimStatus_CLAIM_STATUS_REJECTED, updatedClaim2.Status)
	s.Require().Equal("Insufficient evidence", updatedClaim2.Decision.Reason)
}

func (s *KeeperTestSuite) TestReviewClaimUnauthorized() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	claimant := s.TestAccs[0]
	unauthorizedUser := s.TestAccs[2]
	policyID := uint64(1)
	amount := sdk.NewCoin("unuah", math.NewInt(1000000))

	// Set up mock policy
	s.policyKeeper.SetPolicy(policyID, policytypes.Policy{
		Id:             policyID,
		Owner:          claimant.String(),
		PolicyType:     "auto",
		Status:         policytypes.PolicyStatus_POLICY_STATUS_ACTIVE,
		TreasuryPoolId: "pool-001",
	})

	// Set up unauthorized user (no roles)
	s.rolesKeeper.SetUserAuthorized(unauthorizedUser, false)

	// Submit claim first
	claim, err := keeper.SubmitClaim(ctx, claimant, policyID, amount, "Test claim", nil)
	s.Require().NoError(err)

	// Try to review claim without authorization
	_, err = keeper.ReviewClaim(ctx, unauthorizedUser, claim.Id, types.ClaimStatus_CLAIM_STATUS_APPROVED, "Test")
	s.Require().Error(err)
	s.Require().Equal(types.ErrUnauthorized, err)
}

func (s *KeeperTestSuite) TestAddClaimEvidence() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	claimant := s.TestAccs[0]
	reviewer := s.TestAccs[1]
	policyID := uint64(1)
	amount := sdk.NewCoin("unuah", math.NewInt(1000000))

	// Set up mock policy
	s.policyKeeper.SetPolicy(policyID, policytypes.Policy{
		Id:             policyID,
		Owner:          claimant.String(),
		PolicyType:     "auto",
		Status:         policytypes.PolicyStatus_POLICY_STATUS_ACTIVE,
		TreasuryPoolId: "pool-001",
	})

	// Set up reviewer role
	s.rolesKeeper.SetHasRole(true)

	// Submit claim first
	claim, err := keeper.SubmitClaim(ctx, claimant, policyID, amount, "Test claim", nil)
	s.Require().NoError(err)

	// Add evidence
	newEvidence := types.ClaimEvidence{
		Uri:   "https://docs.com/additional_evidence.pdf",
		Notes: "Additional medical report",
	}

	updatedClaim, err := keeper.AddClaimEvidence(ctx, reviewer, claim.Id, newEvidence)
	s.Require().NoError(err)
	s.Require().Len(updatedClaim.Evidence, 1)
	s.Require().Equal(newEvidence.Uri, updatedClaim.Evidence[0].Uri)
	s.Require().Equal(newEvidence.Notes, updatedClaim.Evidence[0].Notes)
}

func (s *KeeperTestSuite) TestExecuteClaimPayout() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	claimant := s.TestAccs[0]
	treasuryManager := s.TestAccs[1]
	policyID := uint64(1)
	amount := sdk.NewCoin("unuah", math.NewInt(1000000))

	// Set up mock policy
	s.policyKeeper.SetPolicy(policyID, policytypes.Policy{
		Id:             policyID,
		Owner:          claimant.String(),
		PolicyType:     "auto",
		Status:         policytypes.PolicyStatus_POLICY_STATUS_ACTIVE,
		TreasuryPoolId: "pool-001",
	})

	// Set up treasury manager role
	s.rolesKeeper.SetHasRole(true)

	// Submit and approve claim
	claim, err := keeper.SubmitClaim(ctx, claimant, policyID, amount, "Test claim", nil)
	s.Require().NoError(err)

	// Approve claim
	claim, err = keeper.ReviewClaim(ctx, treasuryManager, claim.Id, types.ClaimStatus_CLAIM_STATUS_APPROVED, "Approved")
	s.Require().NoError(err)

	// Execute payout
	updatedClaim, err := keeper.ExecuteClaimPayout(ctx, treasuryManager, claim.Id, claimant)
	s.Require().NoError(err)
	s.Require().Equal(types.ClaimStatus_CLAIM_STATUS_PAID, updatedClaim.Status)
	s.Require().NotNil(updatedClaim.ResolvedAt)

	// Verify treasury keeper was called
	s.Require().True(s.treasuryKeeper.DisburseClaimCalled)
	s.Require().Equal("pool-001", s.treasuryKeeper.LastPoolID)
	s.Require().Equal(claimant, s.treasuryKeeper.LastRecipient)
	s.Require().Equal(amount, s.treasuryKeeper.LastAmount)
}

func (s *KeeperTestSuite) TestExecuteClaimPayoutUnauthorized() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	claimant := s.TestAccs[0]
	unauthorizedUser := s.TestAccs[2]
	policyID := uint64(1)
	amount := sdk.NewCoin("unuah", math.NewInt(1000000))

	// Set up mock policy
	s.policyKeeper.SetPolicy(policyID, policytypes.Policy{
		Id:             policyID,
		Owner:          claimant.String(),
		PolicyType:     "auto",
		Status:         policytypes.PolicyStatus_POLICY_STATUS_ACTIVE,
		TreasuryPoolId: "pool-001",
	})

	// Set up unauthorized user (no roles)
	s.rolesKeeper.SetUserAuthorized(unauthorizedUser, false)

	// Submit and approve claim
	claim, err := keeper.SubmitClaim(ctx, claimant, policyID, amount, "Test claim", nil)
	s.Require().NoError(err)

	claim, err = keeper.ReviewClaim(ctx, s.TestAccs[1], claim.Id, types.ClaimStatus_CLAIM_STATUS_APPROVED, "Approved")
	s.Require().NoError(err)

	// Try to execute payout without authorization
	_, err = keeper.ExecuteClaimPayout(ctx, unauthorizedUser, claim.Id, claimant)
	s.Require().Error(err)
	s.Require().Equal(types.ErrUnauthorized, err)
}

func (s *KeeperTestSuite) TestExecuteClaimPayoutNotApproved() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	claimant := s.TestAccs[0]
	treasuryManager := s.TestAccs[1]
	policyID := uint64(1)
	amount := sdk.NewCoin("unuah", math.NewInt(1000000))

	// Set up mock policy
	s.policyKeeper.SetPolicy(policyID, policytypes.Policy{
		Id:             policyID,
		Owner:          claimant.String(),
		PolicyType:     "auto",
		Status:         policytypes.PolicyStatus_POLICY_STATUS_ACTIVE,
		TreasuryPoolId: "pool-001",
	})

	// Set up treasury manager role
	s.rolesKeeper.SetHasRole(true)

	// Submit claim (but don't approve it)
	claim, err := keeper.SubmitClaim(ctx, claimant, policyID, amount, "Test claim", nil)
	s.Require().NoError(err)

	// Try to execute payout on pending claim
	_, err = keeper.ExecuteClaimPayout(ctx, treasuryManager, claim.Id, claimant)
	s.Require().Error(err)
	s.Require().Equal(types.ErrInvalidStatusChange, err)
}

func (s *KeeperTestSuite) TestGetClaim() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	claimant := s.TestAccs[0]
	policyID := uint64(1)
	amount := sdk.NewCoin("unuah", math.NewInt(1000000))

	// Set up mock policy
	s.policyKeeper.SetPolicy(policyID, policytypes.Policy{
		Id:             policyID,
		Owner:          claimant.String(),
		PolicyType:     "auto",
		Status:         policytypes.PolicyStatus_POLICY_STATUS_ACTIVE,
		TreasuryPoolId: "pool-001",
	})

	// Submit claim
	claim, err := keeper.SubmitClaim(ctx, claimant, policyID, amount, "Test claim", nil)
	s.Require().NoError(err)

	// Get claim
	retrievedClaim, found := keeper.GetClaim(ctx, claim.Id)
	s.Require().True(found)
	s.Require().Equal(claim.Id, retrievedClaim.Id)
	s.Require().Equal(claim.PolicyId, retrievedClaim.PolicyId)
	s.Require().Equal(claim.Claimant, retrievedClaim.Claimant)

	// Get non-existent claim
	_, found = keeper.GetClaim(ctx, 999)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestGetAllClaims() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	claimant := s.TestAccs[0]
	policyID := uint64(1)
	amount := sdk.NewCoin("unuah", math.NewInt(1000000))

	// Set up mock policy
	s.policyKeeper.SetPolicy(policyID, policytypes.Policy{
		Id:             policyID,
		Owner:          claimant.String(),
		PolicyType:     "auto",
		Status:         policytypes.PolicyStatus_POLICY_STATUS_ACTIVE,
		TreasuryPoolId: "pool-001",
	})

	// Submit multiple claims
	claims := make([]types.Claim, 3)
	for i := 0; i < 3; i++ {
		claim, err := keeper.SubmitClaim(ctx, claimant, policyID, amount, fmt.Sprintf("Test claim %d", i+1), nil)
		s.Require().NoError(err)
		claims[i] = claim
	}

	// Get all claims
	allClaims, _, err := keeper.GetClaims(ctx, &types.QueryClaimsRequest{}, nil)
	s.Require().NoError(err)
	s.Require().Len(allClaims, 3)

	// Verify claims
	for i, claim := range allClaims {
		s.Require().Equal(claims[i].Id, claim.Id)
		s.Require().Equal(claims[i].PolicyId, claim.PolicyId)
		s.Require().Equal(claims[i].Claimant, claim.Claimant)
	}
}

func (s *KeeperTestSuite) TestMaxOpenClaimsPerPolicy() {
	ctx := s.Ctx
	keeper := s.keeper

	// Set max open claims to 2
	params := types.DefaultParams()
	params.MaxOpenClaimsPerPolicy = 2
	keeper.SetParams(ctx, params)

	// Test data
	claimant := s.TestAccs[0]
	policyID := uint64(1)
	amount := sdk.NewCoin("unuah", math.NewInt(1000000))

	// Set up mock policy
	s.policyKeeper.SetPolicy(policyID, policytypes.Policy{
		Id:             policyID,
		Owner:          claimant.String(),
		PolicyType:     "auto",
		Status:         policytypes.PolicyStatus_POLICY_STATUS_ACTIVE,
		TreasuryPoolId: "pool-001",
	})

	// Submit 2 claims (should succeed)
	_, err := keeper.SubmitClaim(ctx, claimant, policyID, amount, "Claim 1", nil)
	s.Require().NoError(err)
	_, err = keeper.SubmitClaim(ctx, claimant, policyID, amount, "Claim 2", nil)
	s.Require().NoError(err)

	// Submit 3rd claim (should fail)
	_, err = keeper.SubmitClaim(ctx, claimant, policyID, amount, "Claim 3", nil)
	s.Require().Error(err)
	s.Require().Equal(types.ErrMaxOpenClaimsExceeded, err)
}

// Mock implementations

type MockRolesKeeper struct {
	hasRole         bool
	authorizedUsers map[string]bool
}

func NewMockRolesKeeper() *MockRolesKeeper {
	return &MockRolesKeeper{
		hasRole:         true,
		authorizedUsers: make(map[string]bool),
	}
}

func (m *MockRolesKeeper) SetHasRole(hasRole bool) {
	m.hasRole = hasRole
}

func (m *MockRolesKeeper) SetUserAuthorized(addr sdk.AccAddress, authorized bool) {
	m.authorizedUsers[addr.String()] = authorized
}

func (m *MockRolesKeeper) HasRole(ctx sdk.Context, addr sdk.AccAddress, role rolestypes.Role) bool {
	if authorized, exists := m.authorizedUsers[addr.String()]; exists {
		return authorized
	}
	return m.hasRole
}

type MockPolicyKeeper struct {
	policies map[uint64]policytypes.Policy
}

func NewMockPolicyKeeper() *MockPolicyKeeper {
	return &MockPolicyKeeper{
		policies: make(map[uint64]policytypes.Policy),
	}
}

func (m *MockPolicyKeeper) SetPolicy(id uint64, policy policytypes.Policy) {
	m.policies[id] = policy
}

func (m *MockPolicyKeeper) GetPolicy(ctx sdk.Context, id uint64) (policytypes.Policy, bool) {
	policy, found := m.policies[id]
	return policy, found
}

type MockTreasuryKeeper struct {
	DisburseClaimCalled bool
	LastPoolID          string
	LastRecipient       sdk.AccAddress
	LastAmount          sdk.Coin
}

func NewMockTreasuryKeeper() *MockTreasuryKeeper {
	return &MockTreasuryKeeper{}
}

func (m *MockTreasuryKeeper) DisburseClaim(ctx sdk.Context, poolID string, recipient sdk.AccAddress, amount sdk.Coin) error {
	m.DisburseClaimCalled = true
	m.LastPoolID = poolID
	m.LastRecipient = recipient
	m.LastAmount = amount
	return nil
}

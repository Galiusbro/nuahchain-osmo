package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/policy/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/policy/types"
	roleskeeper "github.com/osmosis-labs/osmosis/v30/x/roles/keeper"
	rolestypes "github.com/osmosis-labs/osmosis/v30/x/roles/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	keeper      keeper.Keeper
	rolesKeeper roleskeeper.Keeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	// Create roles keeper with authority
	authority := s.TestAccs[0].String()
	s.rolesKeeper = roleskeeper.NewKeeper(s.App.AppCodec(), s.App.GetKey(rolestypes.StoreKey), s.App.GetSubspace(rolestypes.ModuleName), authority)

	// Set up roles parameters
	rolesParams := rolestypes.NewParams(authority)
	s.rolesKeeper.SetParams(s.Ctx, rolesParams)

	// Create policy keeper
	s.keeper = keeper.NewKeeper(s.App.AppCodec(), s.App.GetKey(types.StoreKey), s.App.GetSubspace(types.ModuleName), s.rolesKeeper)

	// Set up policy parameters to allow the policy types we're testing
	params := types.NewParams(
		[]string{"auto", "home", "life", "custom"}, // Allow common policy types
		365,        // Default duration in days
		"pool-001", // Default treasury pool ID
	)
	s.keeper.SetParams(s.Ctx, params)
}

func (s *KeeperTestSuite) TestCreatePolicy() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	owner := s.TestAccs[0]
	policyType := "auto"
	startTime := time.Now()
	endTime := startTime.Add(365 * 24 * time.Hour) // 1 year
	treasuryPoolID := "pool-001"
	tags := []string{"auto", "vehicle"}

	// Create policy attributes
	attributes := []types.PolicyAttribute{
		{Key: "vehicle_make", Value: "Toyota"},
		{Key: "vehicle_model", Value: "Camry"},
		{Key: "vehicle_year", Value: "2020"},
		{Key: "driver_age", Value: "30"},
	}

	// Create policy
	policy, err := keeper.CreatePolicy(ctx, owner, policyType, attributes, &startTime, &endTime, treasuryPoolID, tags)
	s.Require().NoError(err, "Should be able to create policy")
	s.Require().NotNil(policy, "Policy should not be nil")

	// Verify policy details
	s.Require().Equal(uint64(1), policy.Id, "Policy ID should be auto-generated")
	s.Require().Equal(owner.String(), policy.Owner, "Policy owner should match")
	s.Require().Equal(policyType, policy.PolicyType, "Policy type should match")
	s.Require().Equal(types.PolicyStatus_POLICY_STATUS_ACTIVE, policy.Status, "Policy should be active")
	s.Require().Equal(treasuryPoolID, policy.TreasuryPoolId, "Treasury pool ID should match")
	s.Require().Len(policy.Attributes, len(attributes), "Policy should have correct number of attributes")
	s.Require().Len(policy.Tags, len(tags), "Policy should have correct number of tags")

	// Verify attributes
	for i, expectedAttr := range attributes {
		s.Require().Equal(expectedAttr.Key, policy.Attributes[i].Key, "Attribute key should match")
		s.Require().Equal(expectedAttr.Value, policy.Attributes[i].Value, "Attribute value should match")
	}

	// Verify tags
	for i, expectedTag := range tags {
		s.Require().Equal(expectedTag, policy.Tags[i], "Tag should match")
	}
}

func (s *KeeperTestSuite) TestGetPolicy() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	owner := s.TestAccs[0]
	policyType := "home"
	startTime := time.Now()
	endTime := startTime.Add(365 * 24 * time.Hour)
	treasuryPoolID := "pool-002"
	tags := []string{"home", "property"}

	// Create policy
	policy, err := keeper.CreatePolicy(ctx, owner, policyType, nil, &startTime, &endTime, treasuryPoolID, tags)
	s.Require().NoError(err)

	// Get policy
	retrievedPolicy, found := keeper.GetPolicy(ctx, policy.Id)
	s.Require().True(found, "Policy should be found")
	s.Require().NotNil(retrievedPolicy, "Policy should not be nil")
	s.Require().Equal(policy.Id, retrievedPolicy.Id, "Policy ID should match")

	// Test getting non-existent policy
	nonExistentPolicy, found := keeper.GetPolicy(ctx, 999999)
	s.Require().False(found, "Non-existent policy should not be found")
	s.Require().Equal(types.Policy{}, nonExistentPolicy, "Non-existent policy should be empty")
}

func (s *KeeperTestSuite) TestUpdatePolicyStatus() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	owner := s.TestAccs[0]
	authority := s.TestAccs[1] // Different account for authority
	policyType := "life"
	startTime := time.Now()
	endTime := startTime.Add(365 * 24 * time.Hour)
	treasuryPoolID := "pool-003"
	tags := []string{"life", "insurance"}

	// Create policy
	policy, err := keeper.CreatePolicy(ctx, owner, policyType, nil, &startTime, &endTime, treasuryPoolID, tags)
	s.Require().NoError(err)

	// Assign insurer role to authority
	authorityAddr := s.TestAccs[0] // Use the same authority that was set up in SetupTest
	err = s.rolesKeeper.AssignRoles(ctx, authorityAddr, authority, []rolestypes.Role{rolestypes.Role_ROLE_INSURER})
	s.Require().NoError(err)

	// Update status to expired
	updatedPolicy, err := keeper.UpdatePolicyStatus(ctx, authority, policy.Id, types.PolicyStatus_POLICY_STATUS_EXPIRED)
	s.Require().NoError(err, "Should be able to update policy status")

	// Verify status update
	s.Require().Equal(types.PolicyStatus_POLICY_STATUS_EXPIRED, updatedPolicy.Status, "Policy status should be updated")

	// Test updating non-existent policy
	_, err = keeper.UpdatePolicyStatus(ctx, authority, 999999, types.PolicyStatus_POLICY_STATUS_CANCELLED)
	s.Require().Error(err, "Should error when updating non-existent policy")
	s.Require().Contains(err.Error(), "policy not found", "Error should indicate policy not found")
}

func (s *KeeperTestSuite) TestGetAllPolicies() {
	ctx := s.Ctx
	keeper := s.keeper

	// Create multiple policies
	policyTypes := []string{"auto", "home", "life"}
	startTime := time.Now()
	endTime := startTime.Add(365 * 24 * time.Hour)
	treasuryPoolID := "pool-001"
	tags := []string{"test"}

	createdPolicies := make([]types.Policy, len(policyTypes))
	for i, policyType := range policyTypes {
		owner := s.TestAccs[i]
		policy, err := keeper.CreatePolicy(ctx, owner, policyType, nil, &startTime, &endTime, treasuryPoolID, tags)
		s.Require().NoError(err, "Should be able to create policy %s", policyType)
		createdPolicies[i] = policy
	}

	// Get all policies
	allPolicies := keeper.GetAllPolicies(ctx)
	s.Require().Len(allPolicies, len(policyTypes), "Should return all policies")

	// Verify all policies are present
	for _, expectedPolicy := range createdPolicies {
		found := false
		for _, policy := range allPolicies {
			if policy.Id == expectedPolicy.Id {
				found = true
				s.Require().Equal(expectedPolicy.Owner, policy.Owner, "Policy owner should match")
				s.Require().Equal(expectedPolicy.PolicyType, policy.PolicyType, "Policy type should match")
				break
			}
		}
		s.Require().True(found, "Policy %d should be found in list", expectedPolicy.Id)
	}
}

func (s *KeeperTestSuite) TestGetPoliciesByOwner() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	owner1 := s.TestAccs[0]
	owner2 := s.TestAccs[1]

	// Create policies for different owners
	startTime := time.Now()
	endTime := startTime.Add(365 * 24 * time.Hour)
	treasuryPoolID := "pool-001"
	tags := []string{"test"}

	// Owner 1 policies
	_, err := keeper.CreatePolicy(ctx, owner1, "auto", nil, &startTime, &endTime, treasuryPoolID, tags)
	s.Require().NoError(err)
	_, err = keeper.CreatePolicy(ctx, owner1, "home", nil, &startTime, &endTime, treasuryPoolID, tags)
	s.Require().NoError(err)

	// Owner 2 policies
	_, err = keeper.CreatePolicy(ctx, owner2, "life", nil, &startTime, &endTime, treasuryPoolID, tags)
	s.Require().NoError(err)

	// Get policies for owner 1
	owner1Filter := &types.PolicyFilter{Owner: owner1.String()}
	owner1Policies, _, err := keeper.GetPolicies(ctx, owner1Filter, nil)
	s.Require().NoError(err)
	s.Require().Len(owner1Policies, 2, "Owner 1 should have 2 policies")

	// Verify owner 1 policies
	for _, policy := range owner1Policies {
		s.Require().Equal(owner1.String(), policy.Owner, "Policy should belong to owner 1")
	}

	// Get policies for owner 2
	owner2Filter := &types.PolicyFilter{Owner: owner2.String()}
	owner2Policies, _, err := keeper.GetPolicies(ctx, owner2Filter, nil)
	s.Require().NoError(err)
	s.Require().Len(owner2Policies, 1, "Owner 2 should have 1 policy")

	// Verify owner 2 policies
	for _, policy := range owner2Policies {
		s.Require().Equal(owner2.String(), policy.Owner, "Policy should belong to owner 2")
	}

	// Get policies for non-existent owner
	nonExistentFilter := &types.PolicyFilter{Owner: "non-existent"}
	nonExistentPolicies, _, err := keeper.GetPolicies(ctx, nonExistentFilter, nil)
	s.Require().NoError(err)
	s.Require().Len(nonExistentPolicies, 0, "Non-existent owner should have no policies")
}

func (s *KeeperTestSuite) TestGetPoliciesByType() {
	ctx := s.Ctx
	keeper := s.keeper

	// Create policies of different types
	startTime := time.Now()
	endTime := startTime.Add(365 * 24 * time.Hour)
	treasuryPoolID := "pool-001"
	tags := []string{"test"}

	policyTypes := []string{"auto", "home", "life", "auto", "home"}

	for i, policyType := range policyTypes {
		owner := s.TestAccs[i%len(s.TestAccs)]
		_, err := keeper.CreatePolicy(ctx, owner, policyType, nil, &startTime, &endTime, treasuryPoolID, tags)
		s.Require().NoError(err, "Should be able to create policy %s", policyType)
	}

	// Get auto policies
	autoFilter := &types.PolicyFilter{PolicyType: "auto"}
	autoPolicies, _, err := keeper.GetPolicies(ctx, autoFilter, nil)
	s.Require().NoError(err)
	s.Require().Len(autoPolicies, 2, "Should have 2 auto policies")

	// Verify all are auto policies
	for _, policy := range autoPolicies {
		s.Require().Equal("auto", policy.PolicyType, "Policy should be auto type")
	}

	// Get home policies
	homeFilter := &types.PolicyFilter{PolicyType: "home"}
	homePolicies, _, err := keeper.GetPolicies(ctx, homeFilter, nil)
	s.Require().NoError(err)
	s.Require().Len(homePolicies, 2, "Should have 2 home policies")

	// Verify all are home policies
	for _, policy := range homePolicies {
		s.Require().Equal("home", policy.PolicyType, "Policy should be home type")
	}

	// Get life policies
	lifeFilter := &types.PolicyFilter{PolicyType: "life"}
	lifePolicies, _, err := keeper.GetPolicies(ctx, lifeFilter, nil)
	s.Require().NoError(err)
	s.Require().Len(lifePolicies, 1, "Should have 1 life policy")

	// Verify all are life policies
	for _, policy := range lifePolicies {
		s.Require().Equal("life", policy.PolicyType, "Policy should be life type")
	}

	// Get non-existent type
	nonExistentFilter := &types.PolicyFilter{PolicyType: "non-existent"}
	nonExistentPolicies, _, err := keeper.GetPolicies(ctx, nonExistentFilter, nil)
	s.Require().NoError(err)
	s.Require().Len(nonExistentPolicies, 0, "Non-existent type should have no policies")
}

func (s *KeeperTestSuite) TestPolicyAttributes() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	owner := s.TestAccs[0]
	policyType := "auto"
	startTime := time.Now()
	endTime := startTime.Add(365 * 24 * time.Hour)
	treasuryPoolID := "pool-004"
	tags := []string{"auto", "vehicle"}

	// Create policy with attributes
	attributes := []types.PolicyAttribute{
		{Key: "vehicle_make", Value: "Honda"},
		{Key: "vehicle_model", Value: "Civic"},
		{Key: "vehicle_year", Value: "2021"},
		{Key: "driver_age", Value: "25"},
		{Key: "driving_record", Value: "clean"},
	}

	policy, err := keeper.CreatePolicy(ctx, owner, policyType, attributes, &startTime, &endTime, treasuryPoolID, tags)
	s.Require().NoError(err)

	// Get policy and verify attributes
	retrievedPolicy, found := keeper.GetPolicy(ctx, policy.Id)
	s.Require().True(found, "Policy should be found")
	s.Require().Len(retrievedPolicy.Attributes, len(attributes), "Policy should have correct number of attributes")

	// Verify each attribute
	for i, expectedAttr := range attributes {
		s.Require().Equal(expectedAttr.Key, retrievedPolicy.Attributes[i].Key, "Attribute %d key should match", i)
		s.Require().Equal(expectedAttr.Value, retrievedPolicy.Attributes[i].Value, "Attribute %d value should match", i)
	}

	// Test finding specific attribute
	vehicleMakeFound := false
	for _, attr := range retrievedPolicy.Attributes {
		if attr.Key == "vehicle_make" {
			s.Require().Equal("Honda", attr.Value, "Vehicle make should match")
			vehicleMakeFound = true
			break
		}
	}
	s.Require().True(vehicleMakeFound, "Vehicle make attribute should be found")

	// Test finding non-existent attribute
	nonExistentFound := false
	for _, attr := range retrievedPolicy.Attributes {
		if attr.Key == "non-existent" {
			nonExistentFound = true
			break
		}
	}
	s.Require().False(nonExistentFound, "Non-existent attribute should not be found")
}

func (s *KeeperTestSuite) TestPolicyStatusTransitions() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	owner := s.TestAccs[0]
	authority := s.TestAccs[1] // Different account for authority
	policyType := "auto"
	startTime := time.Now()
	endTime := startTime.Add(365 * 24 * time.Hour)
	treasuryPoolID := "pool-005"
	tags := []string{"auto", "test"}

	// Create policy (not used directly, but needed for setup)
	_, err := keeper.CreatePolicy(ctx, owner, policyType, nil, &startTime, &endTime, treasuryPoolID, tags)
	s.Require().NoError(err)

	// Assign insurer role to authority
	authorityAddr := s.TestAccs[0] // Use the same authority that was set up in SetupTest
	err = s.rolesKeeper.AssignRoles(ctx, authorityAddr, authority, []rolestypes.Role{rolestypes.Role_ROLE_INSURER})
	s.Require().NoError(err)

	// Test valid status transitions - create a new policy for each transition
	validTransitions := []types.PolicyStatus{
		types.PolicyStatus_POLICY_STATUS_EXPIRED,
		types.PolicyStatus_POLICY_STATUS_CLAIMED,
		types.PolicyStatus_POLICY_STATUS_CANCELLED,
	}

	for i, status := range validTransitions {
		// Create a new policy for each transition
		newPolicy, err := keeper.CreatePolicy(ctx, owner, policyType, nil, &startTime, &endTime, treasuryPoolID, tags)
		s.Require().NoError(err, "Should be able to create policy for transition %d", i+1)

		updatedPolicy, err := keeper.UpdatePolicyStatus(ctx, authority, newPolicy.Id, status)
		s.Require().NoError(err, "Should be able to transition to %s", status.String())

		// Verify status
		s.Require().Equal(status, updatedPolicy.Status, "Policy status should be %s", status.String())
	}
}

func (s *KeeperTestSuite) TestPolicyOwnership() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	owner := s.TestAccs[0]
	authority := s.TestAccs[1] // Different account for authority
	policyType := "home"
	startTime := time.Now()
	endTime := startTime.Add(365 * 24 * time.Hour)
	treasuryPoolID := "pool-006"
	tags := []string{"home", "property"}

	// Create policy
	policy, err := keeper.CreatePolicy(ctx, owner, policyType, nil, &startTime, &endTime, treasuryPoolID, tags)
	s.Require().NoError(err)

	// Assign insurer role to authority
	authorityAddr := s.TestAccs[0] // Use the same authority that was set up in SetupTest
	err = s.rolesKeeper.AssignRoles(ctx, authorityAddr, authority, []rolestypes.Role{rolestypes.Role_ROLE_INSURER})
	s.Require().NoError(err)

	// Test that authority can update the policy
	updatedPolicy, err := keeper.UpdatePolicyStatus(ctx, authority, policy.Id, types.PolicyStatus_POLICY_STATUS_CANCELLED)
	s.Require().NoError(err, "Authority should be able to update policy")

	// Verify the update
	s.Require().Equal(types.PolicyStatus_POLICY_STATUS_CANCELLED, updatedPolicy.Status, "Policy should be cancelled")
}

func (s *KeeperTestSuite) TestPolicySerialization() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	owner := s.TestAccs[0]
	policyType := "life"
	startTime := time.Now()
	endTime := startTime.Add(365 * 24 * time.Hour)
	treasuryPoolID := "pool-007"
	tags := []string{"life", "insurance"}

	attributes := []types.PolicyAttribute{
		{Key: "beneficiary", Value: s.TestAccs[1].String()},
		{Key: "medical_history", Value: "clean"},
	}

	// Create policy
	policy, err := keeper.CreatePolicy(ctx, owner, policyType, attributes, &startTime, &endTime, treasuryPoolID, tags)
	s.Require().NoError(err)

	// Test that the policy can be serialized and deserialized
	codec := s.App.AppCodec()
	encoded, err := codec.Marshal(&policy)
	s.Require().NoError(err, "Should be able to marshal policy")

	var decoded types.Policy
	err = codec.Unmarshal(encoded, &decoded)
	s.Require().NoError(err, "Should be able to unmarshal policy")

	// Verify deserialized policy matches original
	s.Require().Equal(policy.Id, decoded.Id, "Policy ID should match")
	s.Require().Equal(policy.Owner, decoded.Owner, "Policy owner should match")
	s.Require().Equal(policy.PolicyType, decoded.PolicyType, "Policy type should match")
	s.Require().Equal(policy.Status, decoded.Status, "Policy status should match")
	s.Require().Equal(policy.TreasuryPoolId, decoded.TreasuryPoolId, "Treasury pool ID should match")
	s.Require().Len(decoded.Attributes, len(policy.Attributes), "Policy attributes should match")
	s.Require().Len(decoded.Tags, len(policy.Tags), "Policy tags should match")
}

func (s *KeeperTestSuite) TestPolicyValidation() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test invalid policy data
	startTime := time.Now()
	endTime := startTime.Add(365 * 24 * time.Hour)
	treasuryPoolID := "pool-001"
	tags := []string{"test"}

	// Test empty policy type
	_, err := keeper.CreatePolicy(ctx, s.TestAccs[0], "", nil, &startTime, &endTime, treasuryPoolID, tags)
	s.Require().Error(err, "Should error with empty policy type")

	// Test invalid time range (end before start)
	invalidEndTime := startTime.Add(-24 * time.Hour)
	_, err = keeper.CreatePolicy(ctx, s.TestAccs[0], "auto", nil, &startTime, &invalidEndTime, treasuryPoolID, tags)
	s.Require().Error(err, "Should error with invalid time range")
}

func (s *KeeperTestSuite) TestConcurrentPolicyOperations() {
	ctx := s.Ctx
	keeper := s.keeper

	// Create multiple policies concurrently (simulated by sequential operations)
	startTime := time.Now()
	endTime := startTime.Add(365 * 24 * time.Hour)
	treasuryPoolID := "pool-001"
	tags := []string{"test"}

	createdPolicies := make([]types.Policy, 10)
	for i := 0; i < 10; i++ {
		owner := s.TestAccs[i%len(s.TestAccs)]
		policyType := []string{"auto", "home", "life"}[i%3]

		policy, err := keeper.CreatePolicy(ctx, owner, policyType, nil, &startTime, &endTime, treasuryPoolID, tags)
		s.Require().NoError(err, "Should be able to create policy %d", i+1)
		createdPolicies[i] = policy
	}

	// Verify all policies were created
	allPolicies := keeper.GetAllPolicies(ctx)
	s.Require().Len(allPolicies, 10, "Should have created 10 policies")

	// Verify each policy
	for i, policy := range allPolicies {
		s.Require().Equal(uint64(i+1), policy.Id, "Policy ID should match")
		s.Require().Equal(types.PolicyStatus_POLICY_STATUS_ACTIVE, policy.Status, "Policy should be active")
	}
}

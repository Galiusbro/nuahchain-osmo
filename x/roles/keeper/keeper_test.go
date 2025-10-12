package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/roles/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/roles/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	keeper keeper.Keeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	authority := s.TestAccs[0].String()
	s.keeper = keeper.NewKeeper(s.App.AppCodec(), s.App.GetKey(types.StoreKey), s.App.GetSubspace(types.ModuleName), authority)

	// Set the authority in params so GetAuthority() returns the correct value
	params := types.DefaultParams()
	params.Authority = authority
	s.keeper.SetParams(s.Ctx, params)
}

func (s *KeeperTestSuite) TestAssignRole() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	address := s.TestAccs[0]
	authority := s.TestAccs[0] // Use same address as authority for testing
	role := types.Role_ROLE_INSURER

	// Test assigning a role
	err := keeper.AssignRoles(ctx, authority, address, []types.Role{role})
	s.Require().NoError(err)

	// Verify role was assigned
	hasRole := keeper.HasRole(ctx, address, role)
	s.Require().True(hasRole, "Address should have the assigned role")

	// Test assigning the same role again (should not error)
	err = keeper.AssignRoles(ctx, authority, address, []types.Role{role})
	s.Require().NoError(err, "Assigning the same role should not error")

	// Test assigning multiple roles
	roles := []types.Role{
		types.Role_ROLE_POLICY_HOLDER,
		types.Role_ROLE_CLAIMS_REVIEWER,
		types.Role_ROLE_ORACLE,
	}

	err = keeper.AssignRoles(ctx, authority, address, roles)
	s.Require().NoError(err, "Should be able to assign multiple roles")

	// Verify all roles are assigned
	for _, r := range roles {
		hasRole := keeper.HasRole(ctx, address, r)
		s.Require().True(hasRole, "Address should have role %s", r.String())
	}
}

func (s *KeeperTestSuite) TestRevokeRole() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	address := s.TestAccs[0]
	authority := s.TestAccs[0] // Use same address as authority for testing
	role := types.Role_ROLE_INSURER

	// First assign a role
	err := keeper.AssignRoles(ctx, authority, address, []types.Role{role})
	s.Require().NoError(err)

	// Verify role is assigned
	hasRole := keeper.HasRole(ctx, address, role)
	s.Require().True(hasRole, "Address should have the assigned role")

	// Revoke the role
	err = keeper.RevokeRoles(ctx, authority, address, []types.Role{role})
	s.Require().NoError(err)

	// Verify role is revoked
	hasRole = keeper.HasRole(ctx, address, role)
	s.Require().False(hasRole, "Address should not have the revoked role")

	// Test revoking a role that doesn't exist (should error because no role binding exists)
	err = keeper.RevokeRoles(ctx, authority, address, []types.Role{types.Role_ROLE_TREASURY_MANAGER})
	s.Require().Error(err, "Revoking non-existent role should error when no role binding exists")
	s.Require().Equal(types.ErrRoleNotFound, err, "Should return role not found error")
}

func (s *KeeperTestSuite) TestHasRole() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	address := s.TestAccs[0]
	authority := s.TestAccs[0] // Use same address as authority for testing
	role := types.Role_ROLE_INSURER

	// Initially should not have the role
	hasRole := keeper.HasRole(ctx, address, role)
	s.Require().False(hasRole, "Address should not have role initially")

	// Assign the role
	err := keeper.AssignRoles(ctx, authority, address, []types.Role{role})
	s.Require().NoError(err)

	// Now should have the role
	hasRole = keeper.HasRole(ctx, address, role)
	s.Require().True(hasRole, "Address should have role after assignment")
}

func (s *KeeperTestSuite) TestGetRoles() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	address := s.TestAccs[0]
	authority := s.TestAccs[0] // Use same address as authority for testing
	expectedRoles := []types.Role{
		types.Role_ROLE_INSURER,
		types.Role_ROLE_POLICY_HOLDER,
		types.Role_ROLE_CLAIMS_REVIEWER,
	}

	// Assign multiple roles
	err := keeper.AssignRoles(ctx, authority, address, expectedRoles)
	s.Require().NoError(err)

	// Get role binding
	binding, found := keeper.GetRoleBinding(ctx, address)
	s.Require().True(found, "Role binding should exist")
	s.Require().Len(binding.Roles, len(expectedRoles), "Should return correct number of roles")

	// Verify all expected roles are present
	for _, expectedRole := range expectedRoles {
		found := false
		for _, role := range binding.Roles {
			if role == expectedRole {
				found = true
				break
			}
		}
		s.Require().True(found, "Role %s should be present", expectedRole.String())
	}
}

func (s *KeeperTestSuite) TestGetRoleBindings() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	addresses := []sdk.AccAddress{
		s.TestAccs[0],
		s.TestAccs[1],
		s.TestAccs[2],
	}
	authority := s.TestAccs[0] // Use same address as authority for testing

	// Assign different roles to different addresses
	roleAssignments := map[string][]types.Role{
		addresses[0].String(): {types.Role_ROLE_INSURER, types.Role_ROLE_TREASURY_MANAGER},
		addresses[1].String(): {types.Role_ROLE_POLICY_HOLDER, types.Role_ROLE_CLAIMS_REVIEWER},
		addresses[2].String(): {types.Role_ROLE_ORACLE},
	}

	// Assign roles
	for _, address := range addresses {
		roles := roleAssignments[address.String()]
		err := keeper.AssignRoles(ctx, authority, address, roles)
		s.Require().NoError(err)
	}

	// Get all role bindings
	roleBindings := keeper.GetAllRoleBindings(ctx)
	s.Require().Len(roleBindings, len(addresses), "Should return correct number of role bindings")

	// Verify each role binding
	for _, binding := range roleBindings {
		expectedRoles, exists := roleAssignments[binding.Address]
		s.Require().True(exists, "Address %s should have role assignments", binding.Address)
		s.Require().Len(binding.Roles, len(expectedRoles), "Address %s should have correct number of roles", binding.Address)

		// Verify all expected roles are present
		for _, expectedRole := range expectedRoles {
			found := false
			for _, role := range binding.Roles {
				if role == expectedRole {
					found = true
					break
				}
			}
			s.Require().True(found, "Address %s should have role %s", binding.Address, expectedRole.String())
		}
	}
}

func (s *KeeperTestSuite) TestHasAnyRole() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	address := s.TestAccs[0]
	authority := s.TestAccs[0] // Use same address as authority for testing
	roles := []types.Role{
		types.Role_ROLE_INSURER,
		types.Role_ROLE_POLICY_HOLDER,
		types.Role_ROLE_CLAIMS_REVIEWER,
	}

	// Initially should not have any roles
	hasAnyRole := s.hasAnyRole(ctx, keeper, address, roles)
	s.Require().False(hasAnyRole, "Address should not have any roles initially")

	// Assign one role
	err := keeper.AssignRoles(ctx, authority, address, []types.Role{roles[0]})
	s.Require().NoError(err)

	// Now should have at least one role
	hasAnyRole = s.hasAnyRole(ctx, keeper, address, roles)
	s.Require().True(hasAnyRole, "Address should have at least one role after assignment")

	// Test with empty roles list
	hasAnyRole = s.hasAnyRole(ctx, keeper, address, []types.Role{})
	s.Require().False(hasAnyRole, "Empty roles list should return false")
}

func (s *KeeperTestSuite) TestHasAllRoles() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	address := s.TestAccs[0]
	authority := s.TestAccs[0] // Use same address as authority for testing
	roles := []types.Role{
		types.Role_ROLE_INSURER,
		types.Role_ROLE_POLICY_HOLDER,
		types.Role_ROLE_CLAIMS_REVIEWER,
	}

	// Initially should not have all roles
	hasAllRoles := s.hasAllRoles(ctx, keeper, address, roles)
	s.Require().False(hasAllRoles, "Address should not have all roles initially")

	// Assign one role
	err := keeper.AssignRoles(ctx, authority, address, []types.Role{roles[0]})
	s.Require().NoError(err)

	// Should not have all roles yet
	hasAllRoles = s.hasAllRoles(ctx, keeper, address, roles)
	s.Require().False(hasAllRoles, "Address should not have all roles with only one assigned")

	// Assign all roles
	err = keeper.AssignRoles(ctx, authority, address, roles)
	s.Require().NoError(err)

	// Now should have all roles
	hasAllRoles = s.hasAllRoles(ctx, keeper, address, roles)
	s.Require().True(hasAllRoles, "Address should have all roles after assignment")

	// Test with empty roles list
	hasAllRoles = s.hasAllRoles(ctx, keeper, address, []types.Role{})
	s.Require().True(hasAllRoles, "Empty roles list should return true")
}

func (s *KeeperTestSuite) TestClearRoles() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	address := s.TestAccs[0]
	authority := s.TestAccs[0] // Use same address as authority for testing
	roles := []types.Role{
		types.Role_ROLE_INSURER,
		types.Role_ROLE_POLICY_HOLDER,
		types.Role_ROLE_CLAIMS_REVIEWER,
	}

	// Assign multiple roles
	err := keeper.AssignRoles(ctx, authority, address, roles)
	s.Require().NoError(err)

	// Verify roles are assigned
	for _, role := range roles {
		hasRole := keeper.HasRole(ctx, address, role)
		s.Require().True(hasRole, "Address should have role %s", role.String())
	}

	// Clear all roles by revoking all roles
	err = keeper.RevokeRoles(ctx, authority, address, roles)
	s.Require().NoError(err)

	// Verify no roles are assigned
	for _, role := range roles {
		hasRole := keeper.HasRole(ctx, address, role)
		s.Require().False(hasRole, "Address should not have role %s after clearing", role.String())
	}

	// Verify GetRoleBinding returns empty or not found
	binding, found := keeper.GetRoleBinding(ctx, address)
	if found {
		s.Require().Len(binding.Roles, 0, "Address should have no roles after clearing")
	}
}

func (s *KeeperTestSuite) TestRoleEnumValues() {
	// Test that all role enum values are valid
	validRoles := []types.Role{
		types.Role_ROLE_UNSPECIFIED,
		types.Role_ROLE_INSURER,
		types.Role_ROLE_POLICY_HOLDER,
		types.Role_ROLE_CLAIMS_REVIEWER,
		types.Role_ROLE_ORACLE,
		types.Role_ROLE_TREASURY_MANAGER,
	}

	for _, role := range validRoles {
		s.Require().NotEmpty(role.String(), "Role %d should have a string representation", role)
	}
}

func (s *KeeperTestSuite) TestConcurrentRoleOperations() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	addresses := []sdk.AccAddress{
		s.TestAccs[0],
		s.TestAccs[1],
		s.TestAccs[2],
	}
	authority := s.TestAccs[0] // Use same address as authority for testing

	// Assign roles concurrently (simulated by sequential operations)
	for i, address := range addresses {
		role := types.Role(i + 1) // Skip ROLE_UNSPECIFIED
		err := keeper.AssignRoles(ctx, authority, address, []types.Role{role})
		s.Require().NoError(err, "Should be able to assign role %s to address %s", role.String(), address.String())
	}

	// Verify all assignments worked
	for i, address := range addresses {
		role := types.Role(i + 1)
		hasRole := keeper.HasRole(ctx, address, role)
		s.Require().True(hasRole, "Address %s should have role %s", address.String(), role.String())
	}
}

func (s *KeeperTestSuite) TestRoleBindingSerialization() {
	ctx := s.Ctx
	keeper := s.keeper

	// Test data
	address := s.TestAccs[0]
	authority := s.TestAccs[0] // Use same address as authority for testing
	roles := []types.Role{
		types.Role_ROLE_INSURER,
		types.Role_ROLE_POLICY_HOLDER,
		types.Role_ROLE_CLAIMS_REVIEWER,
	}

	// Assign roles
	err := keeper.AssignRoles(ctx, authority, address, roles)
	s.Require().NoError(err)

	// Get role bindings and verify serialization
	roleBindings := keeper.GetAllRoleBindings(ctx)
	s.Require().Len(roleBindings, 1, "Should have one role binding")

	binding := roleBindings[0]
	s.Require().Equal(address.String(), binding.Address, "Address should match")
	s.Require().Len(binding.Roles, len(roles), "Should have correct number of roles")

	// Test that the binding can be serialized and deserialized
	codec := s.App.AppCodec()
	encoded, err := codec.Marshal(&binding)
	s.Require().NoError(err, "Should be able to marshal role binding")

	var decoded types.RoleBinding
	err = codec.Unmarshal(encoded, &decoded)
	s.Require().NoError(err, "Should be able to unmarshal role binding")

	s.Require().Equal(binding.Address, decoded.Address, "Decoded address should match")
	s.Require().Len(decoded.Roles, len(binding.Roles), "Decoded roles should have same length")
}

// Helper method to check if address has any of the given roles
func (s *KeeperTestSuite) hasAnyRole(ctx sdk.Context, keeper keeper.Keeper, address sdk.AccAddress, roles []types.Role) bool {
	for _, role := range roles {
		if keeper.HasRole(ctx, address, role) {
			return true
		}
	}
	return false
}

// Helper method to check if address has all of the given roles
func (s *KeeperTestSuite) hasAllRoles(ctx sdk.Context, keeper keeper.Keeper, address sdk.AccAddress, roles []types.Role) bool {
	for _, role := range roles {
		if !keeper.HasRole(ctx, address, role) {
			return false
		}
	}
	return true
}

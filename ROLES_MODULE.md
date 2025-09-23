# Roles Module Documentation

## Overview

The Roles module provides a comprehensive role-based access control (RBAC) system for the insurance blockchain platform. It manages user permissions and authorization across all insurance operations, ensuring that only authorized users can perform specific actions within the system.

## Purpose

The Roles module serves as the foundation for security and access control in the insurance ecosystem. It defines and manages five distinct roles, each with specific permissions and responsibilities:

1. **ROLE_INSURER** - Insurance companies that create and manage policies
2. **ROLE_POLICY_HOLDER** - Individuals or entities that purchase insurance policies
3. **ROLE_CLAIMS_REVIEWER** - Authorized personnel who review and process claims
4. **ROLE_ORACLE** - External data providers for claim verification
5. **ROLE_TREASURY_MANAGER** - Administrators who manage treasury pools and financial operations

## Key Features

### Role Management
- **Role Assignment**: Assign one or multiple roles to blockchain addresses
- **Role Revocation**: Remove roles from addresses when needed
- **Role Verification**: Check if an address has specific role permissions
- **Authority Management**: Update module authority for administrative control

### Security Features
- **Authority-based Control**: Only authorized addresses can assign/revoke roles
- **Event Emission**: All role changes are logged as blockchain events
- **Persistent Storage**: Role bindings are stored permanently on-chain
- **Validation**: Comprehensive validation of addresses and role types

## Core Functions

### Keeper Functions

#### AssignRoles
```go
func (k Keeper) AssignRoles(ctx sdk.Context, authority sdk.AccAddress, addr sdk.AccAddress, roles []types.Role) error
```
Assigns one or more roles to a specified address. Only the module authority can perform this operation.

**Parameters:**
- `authority`: The address performing the role assignment (must match module authority)
- `addr`: The target address receiving the roles
- `roles`: Array of roles to assign

**Events Emitted:**
- `EventTypeRoleAssigned` for each role assigned

#### RevokeRoles
```go
func (k Keeper) RevokeRoles(ctx sdk.Context, authority sdk.AccAddress, addr sdk.AccAddress, roles []types.Role) error
```
Removes one or more roles from a specified address.

**Parameters:**
- `authority`: The address performing the role revocation
- `addr`: The target address losing the roles
- `roles`: Array of roles to revoke

**Events Emitted:**
- `EventTypeRoleRevoked` for each role revoked

#### HasRole
```go
func (k Keeper) HasRole(ctx sdk.Context, addr sdk.AccAddress, role types.Role) bool
```
Checks if an address has a specific role. This is the primary authorization function used by other modules.

**Returns:** `true` if the address has the role, `false` otherwise

#### GetRoleBinding
```go
func (k Keeper) GetRoleBinding(ctx sdk.Context, addr sdk.AccAddress) (types.RoleBinding, bool)
```
Retrieves all roles assigned to a specific address.

**Returns:** RoleBinding struct and existence flag

#### GetAllRoleBindings
```go
func (k Keeper) GetAllRoleBindings(ctx sdk.Context) []types.RoleBinding
```
Returns all role bindings in the system for administrative purposes.

### Message Types

#### MsgAssignRoles
```protobuf
message MsgAssignRoles {
  string authority = 1;
  string address = 2;
  repeated Role roles = 3;
}
```

#### MsgRevokeRoles
```protobuf
message MsgRevokeRoles {
  string authority = 1;
  string address = 2;
  repeated Role roles = 3;
}
```

#### MsgUpdateAuthority
```protobuf
message MsgUpdateAuthority {
  string authority = 1;
  string new_authority = 2;
}
```

## CLI Commands

### Transaction Commands

#### Assign Roles
```bash
nuahd tx roles assign [authority] [address] [role ...] --from [wallet]
```

**Example:**
```bash
# Assign INSURER role to an address
nuahd tx roles assign cosmos1authority... cosmos1insurer... ROLE_INSURER --from authority-wallet

# Assign multiple roles
nuahd tx roles assign cosmos1authority... cosmos1user... ROLE_POLICY_HOLDER ROLE_CLAIMS_REVIEWER --from authority-wallet
```

#### Revoke Roles
```bash
nuahd tx roles revoke [authority] [address] [role ...] --from [wallet]
```

**Example:**
```bash
# Revoke INSURER role
nuahd tx roles revoke cosmos1authority... cosmos1insurer... ROLE_INSURER --from authority-wallet
```

#### Update Authority
```bash
nuahd tx roles update-authority [authority] [new-authority] --from [wallet]
```

### Query Commands

#### Query Roles by Address
```bash
nuahd query roles roles-by-address [address]
```

#### Query All Role Bindings
```bash
nuahd query roles all-bindings
```

#### Query Module Parameters
```bash
nuahd query roles params
```

## Usage Examples from Integration Test

### 1. Setting Up Roles for Insurance Workflow

```go
// Step 1: Assign TREASURY_MANAGER role
err := s.App.RolesKeeper.AssignRoles(
    s.Ctx,
    authority,
    treasuryManager,
    []rolestypes.Role{rolestypes.Role_ROLE_TREASURY_MANAGER},
)
require.NoError(s.T(), err)

// Step 2: Assign INSURER role
err = s.App.RolesKeeper.AssignRoles(
    s.Ctx,
    authority,
    insurer,
    []rolestypes.Role{rolestypes.Role_ROLE_INSURER},
)
require.NoError(s.T(), err)

// Step 3: Assign POLICY_HOLDER role
err = s.App.RolesKeeper.AssignRoles(
    s.Ctx,
    authority,
    policyHolder,
    []rolestypes.Role{rolestypes.Role_ROLE_POLICY_HOLDER},
)
require.NoError(s.T(), err)
```

### 2. Role Verification in Other Modules

```go
// Treasury module checking TREASURY_MANAGER role
func (k Keeper) assertAuthority(ctx sdk.Context, authority string) error {
    expected := k.GetAuthority(ctx)
    if expected == "" {
        // Role-based checking
        addr, err := sdk.AccAddressFromBech32(authority)
        if err != nil {
            return err
        }
        if !k.rolesKeeper.HasRole(ctx, addr, rolestypes.Role_ROLE_TREASURY_MANAGER) {
            return types.ErrUnauthorized
        }
        return nil
    }
    // Direct authority checking
    if authority != expected {
        return types.ErrUnauthorized
    }
    return nil
}
```

### 3. Role Revocation

```go
// Revoke roles when access is no longer needed
err := s.App.RolesKeeper.RevokeRoles(
    s.Ctx,
    authority,
    oldInsurer,
    []rolestypes.Role{rolestypes.Role_ROLE_INSURER},
)
require.NoError(s.T(), err)

// Verify role was revoked
hasRole := s.App.RolesKeeper.HasRole(
    s.Ctx,
    oldInsurer,
    rolestypes.Role_ROLE_INSURER,
)
require.False(s.T(), hasRole)
```

## Integration with Other Modules

The Roles module is integrated with all other insurance modules:

- **Treasury Module**: Checks `ROLE_TREASURY_MANAGER` for pool operations
- **Policy Module**: Validates `ROLE_INSURER` for policy creation and `ROLE_POLICY_HOLDER` for policy purchases
- **Premium Module**: Ensures `ROLE_POLICY_HOLDER` for premium payments
- **Claims Module**: Verifies `ROLE_POLICY_HOLDER` for claim submission and `ROLE_CLAIMS_REVIEWER` for claim processing

## Events

### Role Assignment Event
```go
sdk.NewEvent(
    types.EventTypeRoleAssigned,
    sdk.NewAttribute(types.AttributeKeyAuthority, authority.String()),
    sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
    sdk.NewAttribute(types.AttributeKeyRole, role.String()),
)
```

### Role Revocation Event
```go
sdk.NewEvent(
    types.EventTypeRoleRevoked,
    sdk.NewAttribute(types.AttributeKeyAuthority, authority.String()),
    sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
    sdk.NewAttribute(types.AttributeKeyRole, role.String()),
)
```

## Error Handling

- **ErrUnauthorized**: Returned when an unauthorized address attempts role operations
- **ErrUnknownRole**: Returned when an invalid role is specified
- **ErrRoleNotFound**: Returned when trying to revoke a role that doesn't exist

## Best Practices

1. **Principle of Least Privilege**: Assign only the minimum roles necessary for each address
2. **Regular Audits**: Periodically review role assignments using `GetAllRoleBindings`
3. **Event Monitoring**: Monitor role assignment/revocation events for security
4. **Authority Management**: Carefully manage the module authority address
5. **Role Validation**: Always validate role requirements in dependent modules

## Security Considerations

- The module authority has complete control over role assignments
- Role changes are permanent until explicitly revoked
- All role operations are logged as blockchain events
- Role verification is performed on every authorized operation
- The system supports multiple roles per address for flexible permissions
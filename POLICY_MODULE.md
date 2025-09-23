# Policy Module Documentation

## Overview

The Policy module is a core component of the insurance system that manages insurance policies throughout their lifecycle. It handles policy creation, attribute management, status updates, and cancellation operations. The module integrates with the Treasury and Roles modules to ensure proper authorization and financial management.

## Purpose

The Policy module serves as the central registry for all insurance policies in the system. It provides:

- **Policy Lifecycle Management**: Complete management from creation to expiration or cancellation
- **Attribute Management**: Flexible key-value attribute system for policy customization
- **Status Tracking**: Real-time policy status monitoring and updates
- **Integration Support**: Seamless integration with Treasury pools and role-based access control

## Key Features

### 1. Policy Management
- Create new insurance policies with customizable attributes
- Update policy attributes during the policy lifecycle
- Cancel policies with proper authorization and reason tracking
- Query policies by various filters (owner, type, status)

### 2. Status Management
- Track policy status through predefined states (ACTIVE, EXPIRED, CLAIMED, CANCELLED)
- Administrative status updates with proper authorization
- Automatic status transitions based on policy conditions

### 3. Flexible Attributes
- Key-value attribute system for policy customization
- Support for policy-specific data and metadata
- Attribute replacement and incremental updates

### 4. Security & Access Control
- Integration with Roles module for authorization
- Authority-based operations for administrative functions
- Owner-based access control for policy operations

## Core Functions

### Keeper Functions

#### Policy Management Functions

**`CreatePolicy(ctx, msg)`**
- Creates a new insurance policy
- Validates policy parameters and treasury pool association
- Assigns unique policy ID and sets initial status
- Emits policy creation events

**`UpdatePolicyAttributes(ctx, msg)`**
- Updates policy attributes with new key-value pairs
- Supports both replacement and incremental updates
- Requires proper authorization for modifications
- Maintains attribute history and versioning

**`CancelPolicy(ctx, msg)`**
- Cancels an active policy with specified reason
- Updates policy status to CANCELLED
- Records cancellation timestamp and reason
- Triggers necessary cleanup operations

**`UpdatePolicyStatus(ctx, msg)`**
- Administrative function to update policy status
- Supports all status transitions (ACTIVE, EXPIRED, CLAIMED, CANCELLED)
- Requires authority permissions for execution
- Maintains status change audit trail

#### Query Functions

**`GetPolicy(ctx, policyId)`**
- Retrieves a specific policy by ID
- Returns complete policy information including attributes
- Handles policy existence validation

**`GetPolicies(ctx, filter)`**
- Queries policies with optional filtering
- Supports filtering by owner, policy type, and status
- Returns paginated results for large datasets

**`GetParams(ctx)`**
- Retrieves module parameters and configuration
- Returns current policy module settings

## Message Types

### Transaction Messages

#### MsgCreatePolicy
```go
type MsgCreatePolicy struct {
    Owner          string            // Policy owner address
    PolicyType     string            // Type of insurance policy
    Attributes     []PolicyAttribute // Policy-specific attributes
    TreasuryPoolId string            // Associated treasury pool
    StartTime      *time.Time        // Policy start time
    EndTime        *time.Time        // Policy end time
    Tags           []string          // Policy tags for categorization
}
```

#### MsgUpdatePolicyAttributes
```go
type MsgUpdatePolicyAttributes struct {
    Authority  string            // Authority address for authorization
    PolicyId   uint64            // Target policy ID
    Attributes []PolicyAttribute // New or updated attributes
    Replace    bool              // Whether to replace all attributes
}
```

#### MsgCancelPolicy
```go
type MsgCancelPolicy struct {
    Authority string // Authority address for authorization
    PolicyId  uint64 // Policy ID to cancel
    Reason    string // Cancellation reason
}
```

#### MsgUpdatePolicyStatus
```go
type MsgUpdatePolicyStatus struct {
    Authority string       // Authority address for authorization
    PolicyId  uint64       // Target policy ID
    Status    PolicyStatus // New policy status
}
```

### Data Structures

#### Policy
```go
type Policy struct {
    Id             uint64            // Unique policy identifier
    Owner          string            // Policy owner address
    PolicyType     string            // Insurance policy type
    Status         PolicyStatus      // Current policy status
    StartTime      *time.Time        // Policy effective start time
    EndTime        *time.Time        // Policy expiration time
    Attributes     []PolicyAttribute // Policy attributes
    TreasuryPoolId string            // Associated treasury pool
    PremiumPlanId  string            // Associated premium plan
    Tags           []string          // Policy categorization tags
}
```

#### PolicyAttribute
```go
type PolicyAttribute struct {
    Key   string // Attribute key
    Value string // Attribute value
}
```

#### PolicyStatus (Enum)
- `POLICY_STATUS_UNSPECIFIED`: Default/uninitialized status
- `POLICY_STATUS_ACTIVE`: Policy is active and in effect
- `POLICY_STATUS_EXPIRED`: Policy has reached its end time
- `POLICY_STATUS_CLAIMED`: Policy has been claimed
- `POLICY_STATUS_CANCELLED`: Policy has been cancelled

#### PolicyFilter
```go
type PolicyFilter struct {
    Owner      string       // Filter by policy owner
    PolicyType string       // Filter by policy type
    Status     PolicyStatus // Filter by policy status
}
```

## CLI Commands

### Transaction Commands

#### Create Policy
```bash
osmosisd tx policy create-policy [owner] [policy-type] [attributes] [treasury-pool-id] [start-time] [end-time] [tags] --from [key]
```

#### Update Policy Attributes
```bash
osmosisd tx policy update-attributes [policy-id] [attributes] [replace] --from [authority]
```

#### Cancel Policy
```bash
osmosisd tx policy cancel-policy [policy-id] [reason] --from [authority]
```

#### Update Policy Status
```bash
osmosisd tx policy update-status [policy-id] [status] --from [authority]
```

### Query Commands

#### Query Policy
```bash
osmosisd query policy policy [policy-id]
```

#### Query Policies
```bash
osmosisd query policy policies [--owner] [--policy-type] [--status]
```

#### Query Parameters
```bash
osmosisd query policy params
```

## Usage Examples

### Example 1: Creating a Life Insurance Policy

From the comprehensive integration test:

```bash
# Create a life insurance policy
osmosisd tx policy create-policy \
  $POLICY_HOLDER_ADDR \
  "life_insurance" \
  "coverage_amount:100000,beneficiary:spouse" \
  "pool_1" \
  "2024-01-01T00:00:00Z" \
  "2034-01-01T00:00:00Z" \
  "individual,term_life" \
  --from policy_holder \
  --chain-id testing \
  --yes
```

### Example 2: Updating Policy Attributes

```bash
# Update policy attributes (add medical exam results)
osmosisd tx policy update-attributes \
  1 \
  "medical_exam_date:2024-02-15,health_score:excellent" \
  false \
  --from insurer \
  --chain-id testing \
  --yes
```

### Example 3: Querying Policies

```bash
# Query all active policies for a specific owner
osmosisd query policy policies \
  --owner $POLICY_HOLDER_ADDR \
  --status POLICY_STATUS_ACTIVE

# Query specific policy details
osmosisd query policy policy 1
```

### Example 4: Policy Cancellation

```bash
# Cancel a policy due to non-payment
osmosisd tx policy cancel-policy \
  1 \
  "Non-payment of premiums" \
  --from insurer \
  --chain-id testing \
  --yes
```

## Integration with Other Modules

### Roles Module Integration
- **Authorization**: Policy operations require appropriate roles (Insurer, Policy_Holder)
- **Access Control**: Administrative functions restricted to authorized roles
- **Permission Validation**: Role-based permission checks for sensitive operations

### Treasury Module Integration
- **Pool Association**: Policies must be associated with valid treasury pools
- **Financial Backing**: Treasury pools provide financial backing for policies
- **Reserve Management**: Policy creation affects treasury pool reserves

### Premium Module Integration
- **Premium Plans**: Policies can be associated with premium payment plans
- **Payment Tracking**: Premium payments linked to policy status
- **Billing Integration**: Policy attributes affect premium calculations

### Claims Module Integration
- **Claim Processing**: Policy status affects claim eligibility
- **Coverage Validation**: Policy attributes determine claim coverage
- **Status Updates**: Successful claims update policy status to CLAIMED

## Events

The Policy module emits the following events:

### PolicyCreated
```go
{
  "type": "policy_created",
  "attributes": [
    {"key": "policy_id", "value": "1"},
    {"key": "owner", "value": "osmo1..."},
    {"key": "policy_type", "value": "life_insurance"},
    {"key": "treasury_pool_id", "value": "pool_1"}
  ]
}
```

### PolicyAttributesUpdated
```go
{
  "type": "policy_attributes_updated",
  "attributes": [
    {"key": "policy_id", "value": "1"},
    {"key": "updated_by", "value": "osmo1..."},
    {"key": "attribute_count", "value": "2"}
  ]
}
```

### PolicyCancelled
```go
{
  "type": "policy_cancelled",
  "attributes": [
    {"key": "policy_id", "value": "1"},
    {"key": "cancelled_by", "value": "osmo1..."},
    {"key": "reason", "value": "Non-payment of premiums"}
  ]
}
```

### PolicyStatusUpdated
```go
{
  "type": "policy_status_updated",
  "attributes": [
    {"key": "policy_id", "value": "1"},
    {"key": "old_status", "value": "POLICY_STATUS_ACTIVE"},
    {"key": "new_status", "value": "POLICY_STATUS_EXPIRED"},
    {"key": "updated_by", "value": "osmo1..."}
  ]
}
```

## Error Handling

### Common Errors

- **ErrPolicyNotFound**: Policy with specified ID does not exist
- **ErrUnauthorized**: Insufficient permissions for the operation
- **ErrInvalidPolicyType**: Unsupported or invalid policy type
- **ErrInvalidTreasuryPool**: Associated treasury pool does not exist
- **ErrPolicyExpired**: Operation not allowed on expired policy
- **ErrPolicyAlreadyCancelled**: Cannot modify already cancelled policy
- **ErrInvalidTimeRange**: Start time must be before end time
- **ErrInvalidAttributes**: Malformed or invalid policy attributes

### Error Response Format
```go
{
  "code": 1001,
  "message": "policy not found: policy ID 123 does not exist",
  "details": "The specified policy ID was not found in the system"
}
```

## Best Practices

### Policy Creation
1. **Validate Treasury Pool**: Ensure the associated treasury pool exists and has sufficient reserves
2. **Set Appropriate Times**: Use realistic start and end times for policy coverage
3. **Use Descriptive Attributes**: Include relevant policy details in attributes
4. **Tag Appropriately**: Use consistent tagging for policy categorization

### Attribute Management
1. **Use Structured Keys**: Follow consistent naming conventions for attribute keys
2. **Validate Values**: Ensure attribute values are properly formatted
3. **Consider Privacy**: Avoid storing sensitive information in attributes
4. **Version Control**: Track attribute changes for audit purposes

### Status Management
1. **Follow Lifecycle**: Respect the natural policy lifecycle progression
2. **Document Changes**: Always provide clear reasons for status updates
3. **Coordinate with Claims**: Ensure status changes align with claim processing
4. **Monitor Expiration**: Implement automated expiration handling

### Security Considerations
1. **Role Verification**: Always verify caller roles before operations
2. **Input Validation**: Validate all input parameters thoroughly
3. **Access Control**: Implement proper access control for sensitive operations
4. **Audit Trail**: Maintain comprehensive audit logs for all policy changes

## Security Considerations

### Access Control
- Policy creation requires Policy_Holder role
- Administrative operations require Insurer or higher authority
- Owner-based access control for policy queries
- Authority validation for all administrative functions

### Data Protection
- Sensitive policy information protected through role-based access
- Attribute encryption for confidential data
- Secure storage of policy documents and metadata
- Privacy-preserving query mechanisms

### Operational Security
- Transaction signing required for all state changes
- Multi-signature support for high-value policies
- Rate limiting for policy creation and updates
- Comprehensive audit logging for compliance

### Integration Security
- Secure communication with Treasury module
- Validated treasury pool associations
- Protected role verification with Roles module
- Encrypted inter-module data exchange
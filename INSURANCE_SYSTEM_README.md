# Insurance System Documentation

## Overview

This is a comprehensive blockchain-based insurance system built on the Cosmos SDK, providing a complete end-to-end insurance platform with role-based access control, policy management, premium processing, treasury management, and claims handling.

## System Architecture

The insurance system consists of six interconnected modules that work together to provide a complete insurance platform:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Roles Module  │    │  Policy Module  │    │ Premium Module  │
│                 │    │                 │    │                 │
│ • User Roles    │◄──►│ • Policy Mgmt   │◄──►│ • Payment Plans │
│ • Permissions   │    │ • Attributes    │    │ • Schedules     │
│ • Access Control│    │ • Status Mgmt   │    │ • Overdue Mgmt  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Treasury Module │    │  Claims Module  │    │   Bank Module   │
│                 │    │                 │    │                 │
│ • Pool Mgmt     │◄──►│ • Claim Process │◄──►│ • Token Transfers│
│ • Fund Tracking │    │ • Evidence Mgmt │    │ • Balance Mgmt  │
│ • Contributions │    │ • Payout Exec   │    │ • Fee Handling  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Module Overview

### 1. Roles Module
**Purpose**: Manages user roles and permissions throughout the insurance system.

**Key Features**:
- Role assignment and management (INSURER, UNDERWRITER, CLAIMS_ADJUSTER, etc.)
- Permission-based access control
- Role hierarchy and inheritance
- Administrative role management

**Core Functions**:
- `AssignRole()` - Assign roles to users
- `RevokeRole()` - Remove roles from users
- `HasRole()` - Check if user has specific role
- `GetUserRoles()` - Retrieve all roles for a user

### 2. Policy Module
**Purpose**: Handles insurance policy creation, management, and lifecycle.

**Key Features**:
- Policy creation with flexible attributes
- Policy status management (ACTIVE, EXPIRED, CANCELLED, CLAIMED)
- Time-based policy validity
- Integration with treasury pools

**Core Functions**:
- `CreatePolicy()` - Create new insurance policies
- `UpdatePolicyAttributes()` - Modify policy details
- `CancelPolicy()` - Cancel existing policies
- `UpdatePolicyStatus()` - Change policy status

### 3. Premium Module
**Purpose**: Manages premium payment plans, schedules, and payment processing.

**Key Features**:
- Flexible premium payment schedules
- Payment tracking and history
- Overdue payment management
- Integration with treasury pools

**Core Functions**:
- `CreatePremiumPlan()` - Create payment plans
- `RecordPremiumPayment()` - Process premium payments
- `MarkPremiumOverdue()` - Handle overdue payments
- `GetPremiumPlan()` - Retrieve payment plan details

### 4. Treasury Module
**Purpose**: Manages insurance fund pools, contributions, and financial operations.

**Key Features**:
- Multiple treasury pool management
- Contribution tracking from premiums
- Fund allocation and distribution
- Balance monitoring and reporting

**Core Functions**:
- `CreateTreasuryPool()` - Create new fund pools
- `AddContribution()` - Add funds to pools
- `WithdrawFunds()` - Withdraw funds for payouts
- `GetPoolBalance()` - Check pool balances

### 5. Claims Module
**Purpose**: Handles the complete claims lifecycle from submission to payout.

**Key Features**:
- Claim submission with evidence support
- Review and approval workflow
- Evidence management system
- Automated payout execution

**Core Functions**:
- `SubmitClaim()` - Submit new claims
- `ReviewClaim()` - Review and approve/reject claims
- `AddClaimEvidence()` - Add supporting evidence
- `ExecuteClaimPayout()` - Process approved payouts

### 6. Bank Module (Cosmos SDK)
**Purpose**: Provides core token transfer and balance management functionality.

**Integration Points**:
- Premium payment processing
- Claim payout execution
- Treasury fund management
- Fee collection and distribution

## Module Interactions

### Policy Creation Flow
```
1. User (with INSURER role) creates policy
   ├── Roles Module: Validates INSURER role
   ├── Policy Module: Creates policy record
   └── Treasury Module: Links to treasury pool

2. Premium plan creation (optional)
   ├── Premium Module: Creates payment schedule
   └── Policy Module: Links premium plan to policy
```

### Premium Payment Flow
```
1. Premium payment submitted
   ├── Premium Module: Validates payment plan
   ├── Bank Module: Transfers tokens
   └── Treasury Module: Records contribution

2. Payment processing
   ├── Premium Module: Updates payment status
   └── Policy Module: Maintains policy active status
```

### Claims Processing Flow
```
1. Claim submission
   ├── Claims Module: Validates claim details
   ├── Policy Module: Verifies policy status
   └── Roles Module: Validates claimant permissions

2. Claim review
   ├── Roles Module: Validates reviewer permissions
   ├── Claims Module: Records review decision
   └── Treasury Module: Reserves funds (if approved)

3. Claim payout
   ├── Claims Module: Executes payout
   ├── Treasury Module: Withdraws funds
   └── Bank Module: Transfers tokens to claimant
```

## Key Workflows

### 1. Insurance Policy Lifecycle

#### Policy Creation
```bash
# 1. Create treasury pool (admin)
osmosisd tx treasury create-pool "Auto Insurance Pool" --from admin

# 2. Create policy (insurer)
osmosisd tx policy create-policy \
  --policy-type "auto" \
  --attributes "coverage:comprehensive,deductible:500" \
  --start-time "2024-01-01T00:00:00Z" \
  --end-time "2024-12-31T23:59:59Z" \
  --treasury-pool-id 1 \
  --tags "auto,comprehensive" \
  --from insurer

# 3. Create premium plan (insurer)
osmosisd tx premium create-plan \
  --policy-id 1 \
  --payer osmo1customer... \
  --amount 100uosmo \
  --schedule "monthly:2592000:12" \
  --treasury-pool-id 1 \
  --from insurer
```

#### Premium Payments
```bash
# Customer makes premium payment
osmosisd tx premium record-payment \
  --plan-id 1 \
  --amount 100uosmo \
  --from customer
```

#### Claims Process
```bash
# 1. Submit claim
osmosisd tx claims submit-claim \
  1 \
  5000uosmo \
  "Vehicle accident damage" \
  "https://ipfs.io/evidence1" \
  "Police report and photos" \
  --from customer

# 2. Review claim (claims adjuster)
osmosisd tx claims review-claim \
  1 \
  approved \
  "Valid claim with sufficient evidence" \
  --from adjuster

# 3. Execute payout (admin)
osmosisd tx claims execute-payout \
  1 \
  osmo1customer... \
  --from admin
```

### 2. Role Management Workflow

```bash
# Assign roles to users
osmosisd tx roles assign-role osmo1user... INSURER --from admin
osmosisd tx roles assign-role osmo1user... UNDERWRITER --from admin
osmosisd tx roles assign-role osmo1user... CLAIMS_ADJUSTER --from admin

# Query user roles
osmosisd query roles user-roles osmo1user...

# Check specific role
osmosisd query roles has-role osmo1user... INSURER
```

### 3. Treasury Management Workflow

```bash
# Create treasury pools
osmosisd tx treasury create-pool "Auto Insurance Pool" --from admin
osmosisd tx treasury create-pool "Health Insurance Pool" --from admin

# Query pool status
osmosisd query treasury pool 1
osmosisd query treasury pools

# Query contributions
osmosisd query treasury contributions --pool-id 1
```

## Configuration

### Module Parameters

Each module has configurable parameters that can be adjusted through governance:

#### Roles Module
- `max_roles_per_user`: Maximum roles a user can have
- `role_assignment_fee`: Fee for role assignment operations

#### Policy Module
- `max_policy_duration`: Maximum policy validity period
- `min_policy_amount`: Minimum policy coverage amount
- `policy_creation_fee`: Fee for creating policies

#### Premium Module
- `max_payment_schedule`: Maximum number of payments in a schedule
- `overdue_grace_period`: Grace period before marking payments overdue
- `payment_processing_fee`: Fee for processing premium payments

#### Treasury Module
- `min_pool_balance`: Minimum balance required in treasury pools
- `withdrawal_fee`: Fee for fund withdrawals
- `contribution_fee`: Fee for adding contributions

#### Claims Module
- `max_claim_amount`: Maximum claimable amount per policy
- `claim_review_period`: Maximum time for claim review
- `evidence_size_limit`: Maximum size for evidence attachments

### Governance Parameters

```bash
# Query module parameters
osmosisd query roles params
osmosisd query policy params
osmosisd query premium params
osmosisd query treasury params
osmosisd query claims params

# Update parameters through governance
osmosisd tx gov submit-proposal param-change proposal.json --from proposer
```

## API Reference

### REST Endpoints

#### Roles Module
- `GET /osmosis/roles/v1beta1/roles` - List all available roles
- `GET /osmosis/roles/v1beta1/users/{address}/roles` - Get user roles
- `GET /osmosis/roles/v1beta1/roles/{role}/users` - Get users with role

#### Policy Module
- `GET /osmosis/policy/v1beta1/policies` - List policies
- `GET /osmosis/policy/v1beta1/policies/{id}` - Get policy by ID
- `GET /osmosis/policy/v1beta1/policies/owner/{address}` - Get policies by owner

#### Premium Module
- `GET /osmosis/premium/v1beta1/plans` - List premium plans
- `GET /osmosis/premium/v1beta1/plans/{id}` - Get plan by ID
- `GET /osmosis/premium/v1beta1/payments` - List payments

#### Treasury Module
- `GET /osmosis/treasury/v1beta1/pools` - List treasury pools
- `GET /osmosis/treasury/v1beta1/pools/{id}` - Get pool by ID
- `GET /osmosis/treasury/v1beta1/contributions` - List contributions

#### Claims Module
- `GET /osmosis/claims/v1beta1/claims` - List claims
- `GET /osmosis/claims/v1beta1/claims/{id}` - Get claim by ID
- `GET /osmosis/claims/v1beta1/claims/policy/{id}` - Get claims by policy

### gRPC Services

Each module exposes gRPC services for programmatic access:

```protobuf
// Example: Claims service
service Msg {
  rpc SubmitClaim(MsgSubmitClaim) returns (MsgSubmitClaimResponse);
  rpc ReviewClaim(MsgReviewClaim) returns (MsgReviewClaimResponse);
  rpc AddClaimEvidence(MsgAddClaimEvidence) returns (MsgAddClaimEvidenceResponse);
  rpc ExecuteClaimPayout(MsgExecuteClaimPayout) returns (MsgExecuteClaimPayoutResponse);
}

service Query {
  rpc Claim(QueryClaimRequest) returns (QueryClaimResponse);
  rpc Claims(QueryClaimsRequest) returns (QueryClaimsResponse);
  rpc Params(QueryParamsRequest) returns (QueryParamsResponse);
}
```

## Events and Monitoring

### System Events

The insurance system emits comprehensive events for monitoring and integration:

#### Policy Events
- `policy_created` - New policy created
- `policy_updated` - Policy attributes updated
- `policy_cancelled` - Policy cancelled
- `policy_expired` - Policy reached end date

#### Premium Events
- `premium_plan_created` - New payment plan created
- `premium_payment_recorded` - Payment processed
- `premium_overdue` - Payment marked overdue

#### Claims Events
- `claim_submitted` - New claim submitted
- `claim_reviewed` - Claim reviewed and decided
- `claim_evidence_added` - Evidence added to claim
- `claim_paid_out` - Claim payout executed

#### Treasury Events
- `treasury_pool_created` - New pool created
- `contribution_added` - Funds added to pool
- `funds_withdrawn` - Funds withdrawn from pool

### Monitoring Integration

```bash
# Subscribe to events
osmosisd query txs --events 'claim_submitted.claim_id=1'
osmosisd query txs --events 'policy_created.owner=osmo1abc...'

# Monitor specific modules
osmosisd query txs --events 'message.module=claims'
osmosisd query txs --events 'message.module=policy'
```

## Security Considerations

### Access Control
- Role-based permissions enforced at module level
- Multi-signature requirements for high-value operations
- Time-locked operations for sensitive functions
- Audit trails for all administrative actions

### Financial Security
- Treasury pool balance validation
- Maximum claim amount limits
- Premium payment verification
- Automated fraud detection patterns

### Data Integrity
- Immutable on-chain storage
- Cryptographic evidence verification
- Timestamp validation for time-sensitive operations
- Cross-module data consistency checks

## Development and Testing

### Running Tests
```bash
# Run all module tests
make test

# Run specific module tests
go test ./x/roles/...
go test ./x/policy/...
go test ./x/premium/...
go test ./x/treasury/...
go test ./x/claims/...

# Run integration tests
go test ./tests/integration/...
```

### Local Development
```bash
# Initialize local chain
osmosisd init mynode --chain-id testing

# Add test accounts
osmosisd keys add admin
osmosisd keys add insurer
osmosisd keys add customer
osmosisd keys add adjuster

# Start local node
osmosisd start
```

### Docker Deployment
```bash
# Build Docker image
docker build -t insurance-chain .

# Run container
docker run -p 26657:26657 -p 1317:1317 insurance-chain
```

## Troubleshooting

### Common Issues

#### Role Assignment Failures
```bash
# Check if user has admin role
osmosisd query roles has-role osmo1admin... ADMIN

# Verify role exists
osmosisd query roles roles
```

#### Policy Creation Issues
```bash
# Verify treasury pool exists
osmosisd query treasury pool 1

# Check user has INSURER role
osmosisd query roles has-role osmo1user... INSURER
```

#### Premium Payment Failures
```bash
# Check premium plan status
osmosisd query premium plan 1

# Verify account balance
osmosisd query bank balances osmo1user...
```

#### Claims Processing Issues
```bash
# Verify policy is active
osmosisd query policy policy 1

# Check reviewer permissions
osmosisd query roles has-role osmo1reviewer... CLAIMS_ADJUSTER
```

### Error Codes

| Module | Error Code | Description |
|--------|------------|-------------|
| Roles | 1001 | Role not found |
| Roles | 1002 | Unauthorized role assignment |
| Policy | 2001 | Policy not found |
| Policy | 2002 | Policy not active |
| Premium | 3001 | Premium plan not found |
| Premium | 3002 | Payment overdue |
| Treasury | 4001 | Treasury pool not found |
| Treasury | 4002 | Insufficient funds |
| Claims | 5001 | Claim not found |
| Claims | 5002 | Claim already reviewed |

## Contributing

### Development Guidelines
1. Follow Cosmos SDK conventions
2. Write comprehensive tests for all functionality
3. Document all public APIs
4. Use semantic versioning for releases
5. Maintain backward compatibility

### Code Review Process
1. Create feature branch from main
2. Implement changes with tests
3. Submit pull request with detailed description
4. Address review feedback
5. Merge after approval

### Release Process
1. Update version numbers
2. Generate changelog
3. Create release tag
4. Build and publish binaries
5. Update documentation

## Support and Resources

### Documentation
- [Cosmos SDK Documentation](https://docs.cosmos.network/)
- [Module Development Guide](./docs/module-development.md)
- [API Reference](./docs/api-reference.md)

### Community
- GitHub Issues: Report bugs and feature requests
- Discord: Real-time community support
- Forum: Technical discussions and proposals

### Professional Support
- Enterprise support available
- Custom development services
- Training and consultation

---

*This insurance system provides a complete, production-ready blockchain-based insurance platform with comprehensive functionality for policy management, premium processing, claims handling, and treasury operations.*
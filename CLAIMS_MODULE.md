# Claims Module Documentation

## Overview

The Claims module is a critical component of the insurance system that handles the entire claims lifecycle - from submission to payout. It provides functionality for policyholders to submit claims, for authorized reviewers to evaluate them, and for executing approved payouts from treasury pools.

## Purpose

The Claims module serves as the core claims processing engine, enabling:
- **Claim Submission**: Policyholders can submit claims against their active policies
- **Evidence Management**: Support for attaching and managing claim evidence
- **Review Process**: Authorized reviewers can approve or reject claims
- **Payout Execution**: Automated execution of approved claim payouts
- **Status Tracking**: Complete lifecycle tracking of claims

## Key Features

### 1. Claim Management
- Submit claims with detailed descriptions and evidence
- Automatic policy validation and eligibility checks
- Support for multiple evidence attachments
- Comprehensive claim status tracking

### 2. Review System
- Role-based access control for claim reviewers
- Structured decision recording with reasons
- Status transitions (pending → approved/rejected → paid)
- Audit trail for all review decisions

### 3. Evidence Handling
- Multiple evidence types support (URIs, documents, photos)
- Post-submission evidence addition capability
- Evidence metadata and notes
- Secure evidence storage references

### 4. Payout Processing
- Automated payout execution for approved claims
- Treasury pool integration for fund disbursement
- Flexible payout address specification
- Transaction hash recording for audit

### 5. Security & Access Control
- Role-based permissions for reviewers and administrators
- Policy ownership validation
- Authority checks for sensitive operations
- Comprehensive error handling

## Core Functions

### Keeper Functions

#### Claim Management
- **`SubmitClaim(ctx, claimant, policyId, amount, description, evidence)`**
  - Creates a new claim against an active policy
  - Validates policy existence and status
  - Implements auto-approval logic for eligible claims
  - Returns claim ID for tracking

- **`ReviewClaim(ctx, authority, claimId, decision, reason)`**
  - Allows authorized reviewers to approve/reject claims
  - Records decision with timestamp and reason
  - Updates claim status and triggers events
  - Requires appropriate reviewer role

#### Evidence Operations
- **`AddClaimEvidence(ctx, authority, claimId, evidence)`**
  - Adds additional evidence to existing claims
  - Supports post-submission evidence collection
  - Maintains evidence chain for audit purposes

#### Payout Execution
- **`ExecuteClaimPayout(ctx, authority, claimId, payoutAddress)`**
  - Executes approved claim payouts
  - Transfers funds from treasury pool to specified address
  - Updates claim status to PAID
  - Records transaction details

#### Query Functions
- **`GetClaim(ctx, claimId)`** - Retrieve claim by ID
- **`GetClaims(ctx, filters)`** - List claims with optional filtering
- **`GetClaimsByPolicy(ctx, policyId)`** - Get all claims for a policy
- **`GetClaimsByClaimant(ctx, claimant)`** - Get claims by claimant address
- **`GetClaimsByStatus(ctx, status)`** - Filter claims by status

## Message Types

### Transaction Messages

#### MsgSubmitClaim
```go
type MsgSubmitClaim struct {
    Claimant    string          // Address of the claimant
    PolicyId    uint64          // ID of the policy being claimed against
    Amount      types.Coin      // Claim amount
    Description string          // Detailed claim description
    Evidence    []ClaimEvidence // Initial evidence attachments
}
```

#### MsgReviewClaim
```go
type MsgReviewClaim struct {
    Authority string      // Address of the reviewer (must have appropriate role)
    ClaimId   uint64      // ID of the claim being reviewed
    Decision  ClaimStatus // APPROVED or REJECTED
    Reason    string      // Reason for the decision
}
```

#### MsgAddClaimEvidence
```go
type MsgAddClaimEvidence struct {
    Authority string        // Address with authority to add evidence
    ClaimId   uint64        // ID of the claim
    Evidence  ClaimEvidence // Evidence to be added
}
```

#### MsgExecuteClaimPayout
```go
type MsgExecuteClaimPayout struct {
    Authority     string // Address with payout authority
    ClaimId       uint64 // ID of the approved claim
    PayoutAddress string // Address to receive the payout
}
```

### Data Structures

#### Claim
```go
type Claim struct {
    Id             uint64        // Unique claim identifier
    PolicyId       uint64        // Associated policy ID
    Claimant       string        // Address of the claimant
    Reporter       string        // Address of the claim reporter (if different)
    Amount         types.Coin    // Claimed amount
    Description    string        // Detailed claim description
    Evidence       []ClaimEvidence // Attached evidence
    Status         ClaimStatus   // Current claim status
    Decision       *ClaimDecision // Review decision (if reviewed)
    SubmittedAt    time.Time     // Submission timestamp
    ResolvedAt     *time.Time    // Resolution timestamp (if resolved)
    TreasuryPoolId uint64        // Treasury pool for payout
}
```

#### ClaimEvidence
```go
type ClaimEvidence struct {
    Uri   string // URI/URL to evidence (IPFS, HTTP, etc.)
    Notes string // Additional notes about the evidence
}
```

#### ClaimDecision
```go
type ClaimDecision struct {
    Reviewer  string      // Address of the reviewer
    Status    ClaimStatus // Decision (APPROVED/REJECTED)
    Reason    string      // Reason for the decision
    DecidedAt time.Time   // Decision timestamp
}
```

#### ClaimStatus
```go
type ClaimStatus int32

const (
    CLAIM_STATUS_UNSPECIFIED ClaimStatus = 0
    CLAIM_STATUS_PENDING     ClaimStatus = 1 // Awaiting review
    CLAIM_STATUS_APPROVED    ClaimStatus = 2 // Approved for payout
    CLAIM_STATUS_REJECTED    ClaimStatus = 3 // Rejected
    CLAIM_STATUS_PAID        ClaimStatus = 4 // Payout executed
)
```

## CLI Commands

### Transaction Commands

#### Submit Claim
```bash
# Submit a new claim
osmosisd tx claims submit-claim [policy-id] [amount] [description] [evidence-uri] [evidence-notes] --from [claimant]

# Example
osmosisd tx claims submit-claim 1 1000uosmo "Car accident claim" "https://ipfs.io/evidence1" "Police report and photos" --from alice
```

#### Review Claim
```bash
# Review a claim (requires reviewer role)
osmosisd tx claims review-claim [claim-id] [decision] [reason] --from [reviewer]

# Example - Approve
osmosisd tx claims review-claim 1 approved "Valid claim with sufficient evidence" --from reviewer1

# Example - Reject
osmosisd tx claims review-claim 2 rejected "Insufficient evidence provided" --from reviewer1
```

#### Add Evidence
```bash
# Add additional evidence to a claim
osmosisd tx claims add-evidence [claim-id] [evidence-uri] [evidence-notes] --from [authority]

# Example
osmosisd tx claims add-evidence 1 "https://ipfs.io/additional-evidence" "Medical report" --from reviewer1
```

#### Execute Payout
```bash
# Execute payout for approved claim
osmosisd tx claims execute-payout [claim-id] [payout-address] --from [authority]

# Example
osmosisd tx claims execute-payout 1 osmo1abc...xyz --from admin
```

### Query Commands

#### Query Claim
```bash
# Get claim by ID
osmosisd query claims claim [claim-id]

# Example
osmosisd query claims claim 1
```

#### Query Claims
```bash
# List all claims
osmosisd query claims claims

# Filter by policy ID
osmosisd query claims claims --policy-id 1

# Filter by claimant
osmosisd query claims claims --claimant osmo1abc...xyz

# Filter by status
osmosisd query claims claims --status pending
```

## Usage Examples

### Example from Integration Test

```go
// Submit a claim
submitMsg := &types.MsgSubmitClaim{
    Claimant:    "osmo1claimant123",
    PolicyId:    1,
    Amount:      sdk.NewCoin("uosmo", sdk.NewInt(5000)),
    Description: "Vehicle damage from accident",
    Evidence: []types.ClaimEvidence{
        {
            Uri:   "https://ipfs.io/QmHash1",
            Notes: "Police report",
        },
        {
            Uri:   "https://ipfs.io/QmHash2", 
            Notes: "Damage photos",
        },
    },
}

// Execute submission
res, err := msgServer.SubmitClaim(ctx, submitMsg)
claimId := res.ClaimId

// Review the claim
reviewMsg := &types.MsgReviewClaim{
    Authority: "osmo1reviewer456",
    ClaimId:   claimId,
    Decision:  types.CLAIM_STATUS_APPROVED,
    Reason:    "Valid claim with comprehensive evidence",
}

_, err = msgServer.ReviewClaim(ctx, reviewMsg)

// Execute payout
payoutMsg := &types.MsgExecuteClaimPayout{
    Authority:     "osmo1admin789",
    ClaimId:       claimId,
    PayoutAddress: "osmo1claimant123",
}

_, err = msgServer.ExecuteClaimPayout(ctx, payoutMsg)
```

## Integration with Other Modules

### Roles Module
- **Reviewer Authorization**: Validates reviewer roles before allowing claim reviews
- **Admin Permissions**: Checks admin roles for payout execution and evidence addition
- **Access Control**: Enforces role-based permissions throughout the claims process

### Policy Module
- **Policy Validation**: Verifies policy existence and active status before claim submission
- **Coverage Verification**: Ensures claim amount doesn't exceed policy coverage
- **Policy Linking**: Maintains relationship between claims and policies

### Treasury Module
- **Fund Management**: Integrates with treasury pools for claim payouts
- **Balance Verification**: Ensures sufficient funds before payout execution
- **Transaction Recording**: Records payout transactions for audit purposes

### Bank Module
- **Token Transfers**: Executes actual token transfers for claim payouts
- **Balance Updates**: Updates account balances after payout execution
- **Transaction Fees**: Handles transaction fees for payout operations

## Events

The Claims module emits the following events:

### ClaimSubmitted
```go
{
    "type": "claim_submitted",
    "attributes": [
        {"key": "claim_id", "value": "1"},
        {"key": "policy_id", "value": "1"},
        {"key": "claimant", "value": "osmo1abc..."},
        {"key": "amount", "value": "1000uosmo"}
    ]
}
```

### ClaimReviewed
```go
{
    "type": "claim_reviewed",
    "attributes": [
        {"key": "claim_id", "value": "1"},
        {"key": "reviewer", "value": "osmo1reviewer..."},
        {"key": "decision", "value": "approved"},
        {"key": "reason", "value": "Valid claim"}
    ]
}
```

### ClaimPaidOut
```go
{
    "type": "claim_paid_out",
    "attributes": [
        {"key": "claim_id", "value": "1"},
        {"key": "amount", "value": "1000uosmo"},
        {"key": "recipient", "value": "osmo1recipient..."},
        {"key": "tx_hash", "value": "ABC123..."}
    ]
}
```

## Error Handling

### Common Errors
- `ErrClaimNotFound`: Claim with specified ID doesn't exist
- `ErrPolicyNotFound`: Referenced policy doesn't exist
- `ErrPolicyNotActive`: Policy is not in active status
- `ErrInvalidClaimAmount`: Claim amount exceeds policy coverage
- `ErrUnauthorizedReviewer`: Reviewer lacks required permissions
- `ErrClaimAlreadyReviewed`: Attempt to review already decided claim
- `ErrClaimNotApproved`: Attempt to payout non-approved claim
- `ErrInsufficientFunds`: Treasury pool lacks sufficient funds
- `ErrInvalidEvidence`: Evidence format or content is invalid

### Error Response Format
```go
type ErrorResponse struct {
    Code    uint32 `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

## Best Practices

### For Claimants
1. **Complete Documentation**: Provide comprehensive evidence with initial claim submission
2. **Accurate Information**: Ensure all claim details are accurate and verifiable
3. **Timely Submission**: Submit claims promptly after incidents occur
4. **Evidence Quality**: Use high-quality, clear evidence (photos, documents, reports)

### For Reviewers
1. **Thorough Review**: Carefully examine all evidence before making decisions
2. **Clear Reasoning**: Provide detailed reasons for all decisions
3. **Timely Processing**: Process claims within reasonable timeframes
4. **Consistent Standards**: Apply consistent evaluation criteria across claims

### For Administrators
1. **Role Management**: Properly assign and manage reviewer roles
2. **Treasury Monitoring**: Ensure adequate funds in treasury pools
3. **Audit Trails**: Maintain comprehensive audit trails for all operations
4. **Security Practices**: Follow security best practices for sensitive operations

## Security Considerations

### Access Control
- All reviewer operations require appropriate role validation
- Payout execution requires admin-level permissions
- Evidence addition is restricted to authorized users
- Policy ownership is validated for claim submissions

### Data Integrity
- All claim data is immutably stored on-chain
- Evidence references are cryptographically secured
- Decision records include timestamps and reviewer identification
- Payout transactions are permanently recorded

### Financial Security
- Treasury pool balance validation before payouts
- Multi-signature requirements for large payouts (configurable)
- Automatic fraud detection for suspicious patterns
- Rate limiting for claim submissions per policy

### Audit and Compliance
- Complete audit trail for all claim operations
- Immutable decision records with reasoning
- Transparent payout execution with transaction hashes
- Comprehensive event logging for external monitoring
# Premium Module Documentation

## Overview

The Premium module manages insurance premium plans, payments, and overdue tracking within the insurance system. It handles the creation of premium payment schedules, recording of payments, and management of overdue premiums for insurance policies.

## Purpose

The Premium module serves as the financial backbone for insurance premium management, providing:
- **Premium Plan Management**: Creation and management of premium payment schedules
- **Payment Processing**: Recording and tracking premium payments
- **Overdue Management**: Monitoring and marking overdue premium payments
- **Financial Integration**: Seamless integration with treasury pools for premium collection

## Key Features

### 1. Premium Plan Management
- Create premium plans with flexible payment schedules
- Support for different payment frequencies (monthly, quarterly, annual)
- Integration with policy and treasury systems
- Automatic calculation of due dates

### 2. Payment Processing
- Record premium payments from policyholders
- Update payment status and calculate next due dates
- Track payment history and statistics
- Automatic plan status updates

### 3. Overdue Management
- Mark premium plans as overdue
- Track overdue reasons and timestamps
- Administrative controls for overdue management
- Integration with policy status updates

### 4. Security & Access Control
- Role-based access control using the Roles module
- Authority validation for administrative operations
- Secure payment recording and plan management

## Core Functions

### Keeper Functions

#### Premium Plan Management
- **`CreatePremiumPlan(ctx, authority, policyId, payer, amount, schedule, treasuryPoolId)`**: Creates a new premium plan for a policy
- **`GetPremiumPlan(ctx, planId)`**: Retrieves a premium plan by ID
- **`GetAllPremiumPlans(ctx)`**: Retrieves all premium plans
- **`GetPremiumPlansByPolicy(ctx, policyId)`**: Gets premium plans for a specific policy
- **`GetPremiumPlansByPayer(ctx, payer)`**: Gets premium plans for a specific payer

#### Payment Operations
- **`RecordPremiumPayment(ctx, payer, planId, amount)`**: Records a premium payment
- **`GetPremiumPayment(ctx, paymentId)`**: Retrieves a payment record by ID
- **`GetPaymentsByPlan(ctx, planId)`**: Gets all payments for a premium plan
- **`GetPaymentsByPayer(ctx, payer)`**: Gets all payments by a specific payer

#### Overdue Management
- **`MarkPremiumOverdue(ctx, authority, planId, reason)`**: Marks a premium plan as overdue
- **`GetOverduePremiums(ctx)`**: Retrieves all overdue premium records
- **`IsOverdue(ctx, planId)`**: Checks if a premium plan is overdue

#### Administrative Functions
- **`GetParams(ctx)`**: Retrieves module parameters
- **`SetParams(ctx, params)`**: Sets module parameters
- **`GetAuthority()`**: Returns the module authority address

## Message Types

### Transaction Messages

#### MsgCreatePremiumPlan
Creates a new premium plan for an insurance policy.

**Fields:**
- `authority`: Address with authority to create premium plans (typically insurer)
- `policy_id`: ID of the associated insurance policy
- `payer`: Address of the premium payer
- `amount`: Premium amount per payment
- `schedule`: Payment schedule configuration
- `treasury_pool_id`: ID of the treasury pool for premium collection

#### MsgRecordPremiumPayment
Records a premium payment made by a policyholder.

**Fields:**
- `payer`: Address of the payment sender
- `plan_id`: ID of the premium plan
- `amount`: Payment amount

#### MsgMarkPremiumOverdue
Marks a premium plan as overdue (administrative function).

**Fields:**
- `authority`: Address with authority to mark overdue
- `plan_id`: ID of the premium plan
- `reason`: Reason for marking as overdue

### Data Structures

#### PremiumPlan
Represents a premium payment plan for an insurance policy.

**Fields:**
- `id`: Unique plan identifier
- `policy_id`: Associated policy ID
- `payer`: Premium payer address
- `amount`: Premium amount per payment
- `schedule`: Payment schedule configuration
- `status`: Plan status (ACTIVE, SUSPENDED, COMPLETED, CANCELLED)
- `next_due_time`: Next payment due date
- `payments_made`: Number of payments completed
- `last_payment_time`: Timestamp of last payment
- `treasury_pool_id`: Associated treasury pool ID

#### PremiumSchedule
Defines the payment schedule for a premium plan.

**Fields:**
- `schedule_type`: Type of schedule (e.g., PERIODIC)
- `period_seconds`: Payment period in seconds
- `max_payments`: Maximum number of payments (0 for unlimited)

#### PremiumPayment
Records a premium payment transaction.

**Fields:**
- `id`: Unique payment identifier
- `plan_id`: Associated premium plan ID
- `payer`: Payment sender address
- `amount`: Payment amount
- `paid_at`: Payment timestamp
- `tx_hash`: Transaction hash

#### PremiumOverdue
Tracks overdue premium information.

**Fields:**
- `plan_id`: Associated premium plan ID
- `due_at`: When the premium became overdue
- `reason`: Reason for being overdue

#### PremiumPlanStatus
Enumeration of premium plan statuses:
- `PREMIUM_PLAN_STATUS_UNSPECIFIED`: Default/unspecified status
- `PREMIUM_PLAN_STATUS_ACTIVE`: Plan is active and accepting payments
- `PREMIUM_PLAN_STATUS_SUSPENDED`: Plan is temporarily suspended
- `PREMIUM_PLAN_STATUS_COMPLETED`: All payments completed
- `PREMIUM_PLAN_STATUS_CANCELLED`: Plan has been cancelled

## CLI Commands

### Transaction Commands

#### Create Premium Plan
```bash
nuahchaind tx premium create-plan [authority] [policy-id] [payer] [amount] [schedule] [treasury-pool-id] --from [key]
```

#### Record Premium Payment
```bash
nuahchaind tx premium record-payment [plan-id] [amount] --from [payer]
```

#### Mark Premium Overdue
```bash
nuahchaind tx premium mark-overdue [authority] [plan-id] [reason] --from [key]
```

### Query Commands

#### Query Premium Plan
```bash
nuahchaind query premium plan [plan-id]
```

#### Query Premium Plans
```bash
nuahchaind query premium plans
```

#### Query Premium Payments
```bash
nuahchaind query premium payments [plan-id]
```

#### Query Module Parameters
```bash
nuahchaind query premium params
```

## Usage Examples

### Creating a Premium Plan
```go
// From integration test
msg := &premiumtypes.MsgCreatePremiumPlan{
    Authority:      insurerAddr,
    PolicyId:       policyId,
    Payer:          policyHolderAddr,
    Amount:         sdk.NewCoin("stake", sdk.NewInt(1000)),
    Schedule: premiumtypes.PremiumSchedule{
        ScheduleType:   "PERIODIC",
        PeriodSeconds:  2592000, // 30 days
        MaxPayments:    12,      // 12 monthly payments
    },
    TreasuryPoolId: treasuryPoolId,
}

res, err := msgServer.CreatePremiumPlan(ctx, msg)
if err != nil {
    return err
}
planId := res.PlanId
```

### Recording a Premium Payment
```go
// Record payment from policyholder
paymentMsg := &premiumtypes.MsgRecordPremiumPayment{
    Payer:  policyHolderAddr,
    PlanId: planId,
    Amount: sdk.NewCoin("stake", sdk.NewInt(1000)),
}

paymentRes, err := msgServer.RecordPremiumPayment(ctx, paymentMsg)
if err != nil {
    return err
}

fmt.Printf("Payment recorded with ID: %d\n", paymentRes.PaymentId)
fmt.Printf("Next due time: %s\n", paymentRes.NextDueTime)
```

### Marking Premium as Overdue
```go
// Mark premium as overdue (insurer authority required)
overdueMsg := &premiumtypes.MsgMarkPremiumOverdue{
    Authority: insurerAddr,
    PlanId:    planId,
    Reason:    "Payment not received within grace period",
}

_, err := msgServer.MarkPremiumOverdue(ctx, overdueMsg)
if err != nil {
    return err
}
```

### Querying Premium Information
```go
// Query premium plan
plan, err := keeper.GetPremiumPlan(ctx, planId)
if err != nil {
    return err
}

// Query payments for a plan
payments, err := keeper.GetPaymentsByPlan(ctx, planId)
if err != nil {
    return err
}

// Check if premium is overdue
isOverdue := keeper.IsOverdue(ctx, planId)
```

## Integration with Other Modules

### Roles Module
- **Authority Validation**: Validates that only authorized addresses (insurers) can create premium plans and mark overdue
- **Access Control**: Ensures proper role-based access to administrative functions

### Policy Module
- **Policy Association**: Premium plans are linked to specific insurance policies
- **Status Synchronization**: Premium status affects policy status and vice versa

### Treasury Module
- **Premium Collection**: Premium payments are deposited into specified treasury pools
- **Financial Management**: Integration with treasury for premium fund management

### Bank Module
- **Payment Processing**: Handles the actual transfer of premium payments
- **Balance Management**: Manages account balances for premium transactions

## Events

The Premium module emits the following events:

### PremiumPlanCreated
Emitted when a new premium plan is created.
- `plan_id`: ID of the created plan
- `policy_id`: Associated policy ID
- `payer`: Premium payer address
- `amount`: Premium amount

### PremiumPaymentRecorded
Emitted when a premium payment is recorded.
- `payment_id`: ID of the payment record
- `plan_id`: Associated premium plan ID
- `payer`: Payment sender
- `amount`: Payment amount

### PremiumMarkedOverdue
Emitted when a premium is marked as overdue.
- `plan_id`: ID of the overdue plan
- `reason`: Reason for being overdue

## Error Handling

Common error scenarios and their handling:

### Invalid Authority
- **Error**: Unauthorized access to administrative functions
- **Handling**: Validate authority using Roles module before execution

### Invalid Payment Amount
- **Error**: Payment amount doesn't match expected premium
- **Handling**: Validate payment amount against plan requirements

### Plan Not Found
- **Error**: Attempting to operate on non-existent premium plan
- **Handling**: Verify plan existence before operations

### Insufficient Funds
- **Error**: Payer doesn't have sufficient balance for premium payment
- **Handling**: Check account balance before processing payment

## Best Practices

### For Developers
1. **Authority Validation**: Always validate authority before administrative operations
2. **Payment Verification**: Verify payment amounts and schedules before processing
3. **Status Management**: Keep premium plan status synchronized with policy status
4. **Error Handling**: Implement comprehensive error handling for all operations

### For Insurers
1. **Schedule Planning**: Design premium schedules that align with policy terms
2. **Overdue Management**: Implement timely overdue marking processes
3. **Treasury Integration**: Ensure proper treasury pool configuration for premium collection
4. **Payment Monitoring**: Regularly monitor premium payment status

### for Policyholders
1. **Timely Payments**: Make premium payments before due dates to avoid overdue status
2. **Amount Verification**: Ensure payment amounts match premium requirements
3. **Status Monitoring**: Regularly check premium plan status and payment history

## Security Considerations

### Access Control
- Premium plan creation requires insurer authority
- Overdue marking requires administrative privileges
- Payment recording is open to payers but validated

### Financial Security
- All premium payments are processed through the secure Bank module
- Treasury integration ensures proper fund management
- Payment validation prevents incorrect amounts

### Data Integrity
- Premium plans are immutable once created (except status updates)
- Payment records provide complete audit trail
- Overdue tracking maintains accountability

### Validation
- Comprehensive input validation for all operations
- Authority checks for administrative functions
- Balance verification for payment processing
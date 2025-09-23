# Treasury Module Documentation

## Overview

The Treasury module is a core component of the insurance system that manages financial pools for different types of insurance policies. It provides secure fund management, deposit/withdrawal operations, and reserve management capabilities with role-based access control.

## Purpose

The Treasury module serves as the financial backbone of the insurance system by:
- Creating and managing treasury pools for different insurance policy types
- Handling secure deposits and withdrawals of funds
- Managing pool reserves and minimum reserve ratios
- Providing financial oversight and control mechanisms
- Integrating with the Roles module for access control

## Key Features

### 1. Pool Management
- **Treasury Pool Creation**: Create dedicated pools for specific insurance policy types
- **Pool Updates**: Modify pool descriptions, managers, and supported policy types
- **Multi-Pool Support**: Support for multiple independent treasury pools

### 2. Financial Operations
- **Secure Deposits**: Accept deposits from authorized users into specific pools
- **Controlled Withdrawals**: Authority-controlled withdrawals with recipient specification
- **Balance Tracking**: Real-time tracking of pool balances and reserves

### 3. Reserve Management
- **Minimum Reserves**: Set and enforce minimum reserve ratios for each pool
- **Risk Management**: Ensure pools maintain adequate liquidity for claims
- **Multi-Denomination Support**: Handle reserves in different token denominations

### 4. Security & Access Control
- **Role-Based Access**: Integration with Roles module for permission management
- **Authority Validation**: Strict authority checks for sensitive operations
- **Audit Trail**: Complete transaction logging and event emission

## Core Functions

### Keeper Functions

#### Pool Management
- **`CreateTreasuryPool`**: Creates a new treasury pool with specified parameters
- **`UpdateTreasuryPool`**: Updates existing pool configuration
- **`GetTreasuryPool`**: Retrieves pool information by ID
- **`GetAllTreasuryPools`**: Lists all treasury pools

#### Financial Operations
- **`DepositToTreasury`**: Handles deposits to a specific pool
- **`WithdrawFromTreasury`**: Processes withdrawals from pools (authority-controlled)
- **`GetPoolBalance`**: Retrieves current balance of a pool
- **`GetAllPoolBalances`**: Lists balances for all pools

#### Reserve Management
- **`SetPoolReserves`**: Sets minimum reserve requirements for pools
- **`GetPoolReserves`**: Retrieves reserve settings for a pool
- **`ValidateReserves`**: Ensures pools meet minimum reserve requirements

#### Administrative Functions
- **`GetParams`**: Retrieves module parameters
- **`SetParams`**: Updates module parameters
- **`GetAuthority`**: Gets the module authority address

## Message Types

### Transaction Messages

#### MsgCreateTreasuryPool
Creates a new treasury pool for managing insurance funds.

```go
type MsgCreateTreasuryPool struct {
    Authority   string   // Module authority address
    PoolId      string   // Unique pool identifier
    Description string   // Pool description
    Manager     string   // Pool manager address
    PolicyTypes []string // Supported policy types
}
```

#### MsgUpdateTreasuryPool
Updates an existing treasury pool configuration.

```go
type MsgUpdateTreasuryPool struct {
    Authority   string   // Module authority address
    PoolId      string   // Pool to update
    Description string   // New description
    Manager     string   // New manager address
    PolicyTypes []string // Updated policy types
}
```

#### MsgDepositToTreasury
Deposits funds into a treasury pool.

```go
type MsgDepositToTreasury struct {
    Depositor string     // Address making the deposit
    PoolId    string     // Target pool ID
    Amount    types.Coin // Amount to deposit
}
```

#### MsgWithdrawFromTreasury
Withdraws funds from a treasury pool (authority-controlled).

```go
type MsgWithdrawFromTreasury struct {
    Authority string     // Module authority address
    PoolId    string     // Source pool ID
    Recipient string     // Withdrawal recipient
    Amount    types.Coin // Amount to withdraw
}
```

#### MsgSetPoolReserves
Sets minimum reserve requirements for a pool.

```go
type MsgSetPoolReserves struct {
    Authority string         // Module authority address
    PoolId    string         // Target pool ID
    Reserves  []PoolReserves // Reserve requirements
}
```

### Data Structures

#### TreasuryPool
Represents a treasury pool configuration.

```go
type TreasuryPool struct {
    Id          string   // Unique pool identifier
    Description string   // Pool description
    Manager     string   // Pool manager address
    PolicyTypes []string // Supported policy types
}
```

#### PoolBalance
Tracks the balance of a treasury pool.

```go
type PoolBalance struct {
    PoolId  string     // Pool identifier
    Balance types.Coin // Current balance
}
```

#### PoolReserves
Defines minimum reserve requirements.

```go
type PoolReserves struct {
    PoolId           string // Pool identifier
    Denom            string // Token denomination
    MinReserveRatio  string // Minimum reserve ratio
}
```

## CLI Commands

### Transaction Commands

#### Create Treasury Pool
```bash
nuahchaind tx treasury create-pool [pool-id] [description] [manager] [policy-types] --from [authority]
```

#### Update Treasury Pool
```bash
nuahchaind tx treasury update-pool [pool-id] [description] [manager] [policy-types] --from [authority]
```

#### Deposit to Treasury
```bash
nuahchaind tx treasury deposit [pool-id] [amount] --from [depositor]
```

#### Withdraw from Treasury
```bash
nuahchaind tx treasury withdraw [pool-id] [recipient] [amount] --from [authority]
```

#### Set Pool Reserves
```bash
nuahchaind tx treasury set-reserves [pool-id] [reserves-json] --from [authority]
```

### Query Commands

#### Query Pool
```bash
nuahchaind query treasury pool [pool-id]
```

#### Query All Pools
```bash
nuahchaind query treasury pools
```

#### Query Pool Balance
```bash
nuahchaind query treasury balance [pool-id]
```

#### Query Pool Reserves
```bash
nuahchaind query treasury reserves [pool-id]
```

## Usage Examples

### Example 1: Creating and Funding a Treasury Pool

From the comprehensive integration test:

```go
// Step 2: Create Treasury Pool
createPoolMsg := &treasurytypes.MsgCreateTreasuryPool{
    Authority:   authority,
    PoolId:      "health-insurance-pool",
    Description: "Pool for health insurance policies",
    Manager:     treasuryManager,
    PolicyTypes: []string{"health", "dental"},
}

// Step 3: Fund Treasury Pool
depositMsg := &treasurytypes.MsgDepositToTreasury{
    Depositor: treasuryManager,
    PoolId:    "health-insurance-pool",
    Amount:    sdk.NewCoin("stake", sdk.NewInt(1000000)),
}
```

### Example 2: Setting Pool Reserves

```go
// Step 8: Set Pool Reserves
reserves := []treasurytypes.PoolReserves{
    {
        PoolId:          "health-insurance-pool",
        Denom:           "stake",
        MinReserveRatio: "0.20", // 20% minimum reserve
    },
}

setReservesMsg := &treasurytypes.MsgSetPoolReserves{
    Authority: authority,
    PoolId:    "health-insurance-pool",
    Reserves:  reserves,
}
```

### Example 3: Processing Claim Payout

```go
// Step 7: Process claim payout from treasury
payoutMsg := &treasurytypes.MsgWithdrawFromTreasury{
    Authority: authority,
    PoolId:    "health-insurance-pool",
    Recipient: policyHolder,
    Amount:    sdk.NewCoin("stake", sdk.NewInt(50000)),
}
```

## Integration with Other Modules

### Roles Module Integration
The Treasury module integrates with the Roles module for access control:
- **Treasury Manager Role**: Required for pool management operations
- **Authority Validation**: Uses roles for permission checking
- **Secure Operations**: Role-based access to sensitive functions

### Bank Module Integration
- **Coin Transfers**: Uses bank module for secure fund transfers
- **Account Management**: Integrates with module accounts
- **Balance Tracking**: Leverages bank module balance queries

### Policy Module Integration
- **Pool Association**: Links treasury pools to policy types
- **Premium Collection**: Receives premium payments
- **Claim Payouts**: Processes approved claim payments

## Events

The Treasury module emits the following events:

### TreasuryPoolCreated
```go
sdk.NewEvent(
    "treasury_pool_created",
    sdk.NewAttribute("pool_id", poolId),
    sdk.NewAttribute("manager", manager),
)
```

### TreasuryDeposit
```go
sdk.NewEvent(
    "treasury_deposit",
    sdk.NewAttribute("pool_id", poolId),
    sdk.NewAttribute("depositor", depositor),
    sdk.NewAttribute("amount", amount.String()),
)
```

### TreasuryWithdrawal
```go
sdk.NewEvent(
    "treasury_withdrawal",
    sdk.NewAttribute("pool_id", poolId),
    sdk.NewAttribute("recipient", recipient),
    sdk.NewAttribute("amount", amount.String()),
)
```

## Error Handling

Common error scenarios and their handling:

### Pool Not Found
```go
if pool == nil {
    return nil, sdkerrors.Wrapf(types.ErrPoolNotFound, "pool %s not found", poolId)
}
```

### Insufficient Funds
```go
if balance.IsLT(amount) {
    return nil, sdkerrors.Wrapf(types.ErrInsufficientFunds, "pool %s has insufficient funds", poolId)
}
```

### Reserve Violation
```go
if newBalance.IsLT(minReserve) {
    return nil, sdkerrors.Wrapf(types.ErrReserveViolation, "withdrawal would violate minimum reserves")
}
```

## Best Practices

### 1. Pool Management
- Use descriptive pool IDs and descriptions
- Assign dedicated managers for each pool
- Regularly review and update policy type associations

### 2. Financial Operations
- Always validate balances before withdrawals
- Implement proper reserve management
- Monitor pool health and liquidity

### 3. Security
- Use role-based access control consistently
- Validate all authority addresses
- Implement proper audit trails

### 4. Integration
- Coordinate with other modules for seamless operations
- Handle cross-module dependencies properly
- Maintain data consistency across modules

## Security Considerations

### Access Control
- All sensitive operations require proper authority validation
- Role-based permissions prevent unauthorized access
- Multi-signature support for high-value operations

### Fund Security
- Module accounts provide secure fund isolation
- Reserve requirements ensure liquidity
- Audit trails enable transaction tracking

### Operational Security
- Parameter validation prevents invalid configurations
- Error handling prevents system failures
- Event emission enables monitoring and alerting

The Treasury module provides a robust foundation for financial management in the insurance system, ensuring secure, transparent, and efficient handling of insurance funds while maintaining strict access controls and operational integrity.
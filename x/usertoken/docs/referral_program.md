# Referral Program Documentation

## Overview

The Referral Program is a feature within the usertoken module that allows token creators to establish referral systems for their tokens. This system enables users to earn rewards by referring others to participate in token-related activities.

## Key Components

### 1. ReferralProgram Structure

```go
type ReferralProgram struct {
    Creator        string // Address of the program creator
    TokenDenom     string // Token denomination for the program
    AvailableLinks uint64 // Number of available referral links
    UsedLinks      uint64 // Number of used referral links
    LastResetTime  int64  // Unix timestamp of last reset
    IsActive       bool   // Whether the program is active
}
```

### 2. ReferralActivation Structure

```go
type ReferralActivation struct {
    ReferralCode   string // Referral code (token denom)
    Referee        string // Address of the person being referred
    Referrer       string // Address of the referrer
    TokenDenom     string // Token denomination
    ActivationTime int64  // Unix timestamp of activation
}
```

## Messages

### CreateReferralProgram

Creates a new referral program for a specific token.

**Message Structure:**
```go
type MsgCreateReferralProgram struct {
    Creator    string // Address of the program creator
    TokenDenom string // Token denomination
}
```

**CLI Command:**
```bash
osmosisd tx usertoken create-referral-program [token-denom] --from [creator]
```

**Example:**
```bash
osmosisd tx usertoken create-referral-program factory/osmo1abc.../mytoken --from creator
```

**Default Settings:**
- Available Links: 3
- Used Links: 0
- Is Active: true
- Last Reset Time: Current block time

### ActivateReferral

Activates a referral link for a user.

**Message Structure:**
```go
type MsgActivateReferral struct {
    ReferralCode string // Referral code (token denom)
    Referee      string // Address of the person being referred
}
```

**CLI Command:**
```bash
osmosisd tx usertoken activate-referral [referral-code] [referee-address] --from [referee]
```

**Example:**
```bash
osmosisd tx usertoken activate-referral factory/osmo1abc.../mytoken osmo1def456... --from referee
```

## Queries

### Query Referral Program

Retrieves information about a specific referral program.

**CLI Command:**
```bash
osmosisd query usertoken referral-program [token-denom]
```

**Example:**
```bash
osmosisd query usertoken referral-program factory/osmo1abc.../mytoken
```

**Response:**
```json
{
  "referral_program": {
    "creator": "osmo1abc123...",
    "token_denom": "factory/osmo1abc.../mytoken",
    "available_links": "3",
    "used_links": "2",
    "last_reset_time": "1640995200",
    "is_active": true
  }
}
```

## Business Logic

### 1. Program Creation

- Only token creators can create referral programs for their tokens
- Each token can have only one referral program
- Programs start with 3 available links by default
- Programs are automatically activated upon creation

### 2. Weekly Link Replenishment System

- **Initial State**: Programs start with 3 available links
- **Full Utilization Reward**: When all available links are used, +3 new links are added weekly
- **Growth Pattern**: 3 → 6 → 9 → 12 → 15... (increments of 3)
- **Reset Timing**: Replenishment occurs at weekly intervals based on LastResetTime
- **Condition**: New links are only added if the previous batch was fully utilized

### 2. Referral Activation

- Users can activate referral links using the token denomination as referral code
- Each user can only activate one referral per program
- Activation reduces available links and increases used links
- Inactive programs cannot accept new referral activations

### 3. Link Management

- Available links: Current number of referrals allowed
- Used links: Current number of activated referrals
- **Weekly Link Replenishment**: When all available links are used, the program automatically adds +3 new links at the end of each week
- **Progressive Growth**: Starting with 3 links, programs can grow to 6, 9, 12, etc. as users fully utilize available slots
- Link replenishment is controlled by LastResetTime (weekly intervals)
- Programs remain active and continue growing as long as there's demand

## Integration with Token Economics

The referral program is designed to work independently of token purchase mechanics. It focuses on:

1. **User Acquisition**: Incentivizing users to bring new participants
2. **Community Building**: Creating networks around specific tokens
3. **Reward Distribution**: Enabling creators to reward successful referrers

## Storage Keys

- **ReferralProgram**: `0x03 + tokenDenom`
- **ReferralActivation**: `0x04 + linkId`

## Events

### ReferralProgramCreated
```json
{
  "type": "referral_program_created",
  "attributes": [
    {"key": "creator", "value": "osmo1abc123..."},
    {"key": "token_denom", "value": "factory/osmo1abc.../mytoken"},
    {"key": "available_links", "value": "3"}
  ]
}
```

### ReferralActivated
```json
{
  "type": "referral_activated",
  "attributes": [
    {"key": "referral_code", "value": "factory/osmo1abc.../mytoken"},
    {"key": "referee", "value": "osmo1def456..."},
    {"key": "referrer", "value": "osmo1abc123..."}
  ]
}
```

## Error Handling

### Common Errors

1. **Invalid Address**: `invalid creator/referee address`
2. **Program Not Found**: `referral program not found: [token-denom]`
3. **Inactive Program**: `referral program is not active`
4. **Already Activated**: `referral already activated for this user`
5. **No Available Links**: `no available referral links remaining`

## Usage Examples

### Complete Workflow Example

1. **Create a token:**
```bash
osmosisd tx usertoken create-user-token mytoken "My Token" MTK 6 --from creator
```

2. **Create referral program:**
```bash
osmosisd tx usertoken create-referral-program factory/osmo1creator.../mytoken --from creator
```

3. **User activates referral:**
```bash
osmosisd tx usertoken activate-referral factory/osmo1creator.../mytoken osmo1user123... --from user
```

4. **Query program status:**
```bash
osmosisd query usertoken referral-program factory/osmo1creator.../mytoken
```

### Testing Scenarios

#### Scenario 1: Basic Program Creation
```go
// Create referral program
msg := types.NewMsgCreateReferralProgram(
    "osmo1creator123...",
    "factory/osmo1creator123.../testtoken",
)

// Verify program exists
program, found := keeper.GetReferralProgram(ctx, "factory/osmo1creator123.../testtoken")
require.True(t, found)
require.Equal(t, uint64(3), program.AvailableLinks)
require.Equal(t, uint64(0), program.UsedLinks)
require.True(t, program.IsActive)
```

#### Scenario 2: Referral Activation
```go
// Activate referral
msg := types.NewMsgActivateReferral(
    "factory/osmo1creator123.../testtoken", // referral code
    "osmo1user456...",                      // referee
)

// Verify activation
activation, found := keeper.GetReferralActivation(ctx, "factory/osmo1creator123.../testtoken")
require.True(t, found)
require.Equal(t, "osmo1user456...", activation.Referee)

// Check updated program state
program, _ := keeper.GetReferralProgram(ctx, "factory/osmo1creator123.../testtoken")
require.Equal(t, uint64(3), program.AvailableLinks)
require.Equal(t, uint64(1), program.UsedLinks)
```

#### Scenario 3: Weekly Link Replenishment
```go
// Simulate full utilization (3 activations)
for i := 0; i < 3; i++ {
    msg := types.NewMsgActivateReferral(
        "factory/osmo1creator123.../testtoken",
        fmt.Sprintf("osmo1user%d...", i),
    )
    _, err := msgServer.ActivateReferral(ctx, msg)
    require.NoError(t, err)
}

// Verify all links used
program, _ := keeper.GetReferralProgram(ctx, "factory/osmo1creator123.../testtoken")
require.Equal(t, uint64(3), program.AvailableLinks)
require.Equal(t, uint64(3), program.UsedLinks)

// Simulate weekly reset (after 7 days)
ctx = ctx.WithBlockTime(ctx.BlockTime().Add(7 * 24 * time.Hour))
keeper.ProcessWeeklyReset(ctx)

// Verify link replenishment (+3 new links)
program, _ = keeper.GetReferralProgram(ctx, "factory/osmo1creator123.../testtoken")
require.Equal(t, uint64(6), program.AvailableLinks) // 3 + 3 new
require.Equal(t, uint64(3), program.UsedLinks)      // previous activations remain
```

## Best Practices

1. **Program Management**:
   - Monitor link utilization to track program success
   - Encourage full utilization to unlock weekly growth
   - Track weekly replenishment cycles for optimal timing
   - Programs automatically grow based on demand

2. **Growth Strategy**:
   - Start with 3 links to test market demand
   - Full utilization triggers automatic expansion (+3 weekly)
   - Progressive growth rewards active communities
   - No manual intervention needed for scaling

3. **Security**:
   - Validate all addresses before processing
   - Check program status before allowing activations
   - Prevent duplicate activations
   - Monitor for abuse patterns in rapid growth

4. **User Experience**:
   - Provide clear error messages
   - Display current available/used link counts
   - Show next replenishment date when links are exhausted
   - Notify users of successful activations and program growth

## Future Enhancements

Potential improvements to the referral system:

1. **Configurable Parameters**: Allow creators to set custom link limits
2. **Reward Automation**: Automatic reward distribution for successful referrals
3. **Analytics**: Detailed tracking of referral performance
4. **Tiered Programs**: Multiple referral levels with different rewards
5. **Time-based Limits**: Configurable reset periods beyond weekly

## Technical Notes

- The referral system uses token denomination as the referral code for simplicity
- Programs are stored using the token denomination as the key
- Weekly resets are handled through the LastResetTime field
- All operations are atomic and state-consistent
- The system is designed to be gas-efficient and scalable

---

*This documentation covers the current implementation of the referral program within the usertoken module. For the latest updates and changes, please refer to the module's changelog and test files.*

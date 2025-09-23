package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v30/x/policy/types"
	rolestypes "github.com/osmosis-labs/osmosis/v30/x/roles/types"
)

// Keeper maintains module state and orchestrates policy operations.
type Keeper struct {
	cdc         codec.BinaryCodec
	storeKey    storetypes.StoreKey
	paramstore  paramtypes.Subspace
	rolesKeeper types.RolesKeeper
}

// NewKeeper constructs a new policy keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	roles types.RolesKeeper,
) Keeper {
	if ps.Name() != "" && !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		paramstore:  ps,
		rolesKeeper: roles,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams returns the current module parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	if k.paramstore.Name() == "" {
		return types.DefaultParams()
	}

	var params types.Params
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams stores the supplied parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if err := params.Validate(); err != nil {
		panic(err)
	}

	if k.paramstore.Name() == "" {
		return
	}

	k.paramstore.SetParamSet(ctx, &params)
}

// CreatePolicy records a new policy entry.
func (k Keeper) CreatePolicy(
	ctx sdk.Context,
	owner sdk.AccAddress,
	policyType string,
	attributes []types.PolicyAttribute,
	startTime, endTime *time.Time,
	treasuryPoolID string,
	tags []string,
) (types.Policy, error) {
	params := k.GetParams(ctx)
	if !k.isPolicyTypeAllowed(params, policyType) {
		return types.Policy{}, types.ErrInvalidPolicyType
	}

	policyID := k.nextPolicyID(ctx)

	start := ctx.BlockTime().UTC()
	if startTime != nil {
		start = startTime.UTC()
	}

	startCopy := start
	policy := types.Policy{
		Id:             policyID,
		Owner:          owner.String(),
		PolicyType:     policyType,
		Status:         types.PolicyStatus_POLICY_STATUS_ACTIVE,
		TreasuryPoolId: treasuryPoolID,
		Tags:           append([]string(nil), tags...),
		StartTime:      &startCopy,
	}

	if endTime != nil {
		end := endTime.UTC()
		if end.Before(startCopy) {
			return types.Policy{}, fmt.Errorf("end time precedes start time")
		}
		endCopy := end
		policy.EndTime = &endCopy
	}

	if len(attributes) > 0 {
		policy.Attributes = append([]types.PolicyAttribute(nil), attributes...)
	}

	k.setPolicy(ctx, policy)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePolicyCreated,
		sdk.NewAttribute(types.AttributeKeyPolicyID, fmt.Sprintf("%d", policyID)),
		sdk.NewAttribute(types.AttributeKeyOwner, owner.String()),
		sdk.NewAttribute(types.AttributeKeyPolicyType, policyType),
	))

	return policy, nil
}

// UpdatePolicyAttributes mutates the policy metadata.
func (k Keeper) UpdatePolicyAttributes(ctx sdk.Context, authority sdk.AccAddress, policyID uint64, attrs []types.PolicyAttribute, replace bool) (types.Policy, error) {
	policy, found := k.GetPolicy(ctx, policyID)
	if !found {
		return types.Policy{}, types.ErrPolicyNotFound
	}

	if err := k.assertPolicyEditable(ctx, authority, policy); err != nil {
		return types.Policy{}, err
	}

	if replace {
		policy.Attributes = nil
	}

	if len(attrs) > 0 {
		if replace {
			policy.Attributes = append([]types.PolicyAttribute(nil), attrs...)
		} else {
			policy.Attributes = append(policy.Attributes, attrs...)
		}
	}

	k.setPolicy(ctx, policy)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePolicyUpdated,
		sdk.NewAttribute(types.AttributeKeyPolicyID, fmt.Sprintf("%d", policy.Id)),
		sdk.NewAttribute(types.AttributeKeyOwner, policy.Owner),
	))

	return policy, nil
}

// CancelPolicy moves a policy to the cancelled state.
func (k Keeper) CancelPolicy(ctx sdk.Context, authority sdk.AccAddress, policyID uint64, _ string) (types.Policy, error) {
	policy, found := k.GetPolicy(ctx, policyID)
	if !found {
		return types.Policy{}, types.ErrPolicyNotFound
	}

	if err := k.assertPolicyEditable(ctx, authority, policy); err != nil {
		return types.Policy{}, err
	}

	if policy.Status == types.PolicyStatus_POLICY_STATUS_CANCELLED {
		return types.Policy{}, types.ErrPolicyAlreadyClosed
	}

	policy.Status = types.PolicyStatus_POLICY_STATUS_CANCELLED
	k.setPolicy(ctx, policy)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePolicyCancelled,
		sdk.NewAttribute(types.AttributeKeyPolicyID, fmt.Sprintf("%d", policy.Id)),
		sdk.NewAttribute(types.AttributeKeyStatus, policy.Status.String()),
	))

	return policy, nil
}

// UpdatePolicyStatus sets the policy status via insurer authority.
func (k Keeper) UpdatePolicyStatus(ctx sdk.Context, authority sdk.AccAddress, policyID uint64, status types.PolicyStatus) (types.Policy, error) {
	policy, found := k.GetPolicy(ctx, policyID)
	if !found {
		return types.Policy{}, types.ErrPolicyNotFound
	}

	if !k.hasRole(ctx, authority, rolestypes.Role_ROLE_INSURER) {
		return types.Policy{}, types.ErrUnauthorized
	}

	switch policy.Status {
	case types.PolicyStatus_POLICY_STATUS_CANCELLED, types.PolicyStatus_POLICY_STATUS_CLAIMED:
		return types.Policy{}, types.ErrPolicyAlreadyClosed
	}

	policy.Status = status
	k.setPolicy(ctx, policy)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePolicyStatusUpdated,
		sdk.NewAttribute(types.AttributeKeyPolicyID, fmt.Sprintf("%d", policy.Id)),
		sdk.NewAttribute(types.AttributeKeyStatus, status.String()),
	))

	return policy, nil
}

// GetPolicy fetches a single policy by identifier.
func (k Keeper) GetPolicy(ctx sdk.Context, policyID uint64) (types.Policy, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PolicyKey(policyID))
	if bz == nil {
		return types.Policy{}, false
	}

	var policy types.Policy
	k.cdc.MustUnmarshal(bz, &policy)
	return policy, true
}

// IteratePolicies walks all stored policies.
func (k Keeper) IteratePolicies(ctx sdk.Context, cb func(types.Policy) bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PolicyKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var policy types.Policy
		k.cdc.MustUnmarshal(iterator.Value(), &policy)
		if cb(policy) {
			return
		}
	}
}

// GetPolicies returns policies matching an optional filter with pagination.
func (k Keeper) GetPolicies(ctx sdk.Context, filter *types.PolicyFilter, pageReq *query.PageRequest) ([]types.Policy, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PolicyKeyPrefix)

	policies := make([]types.Policy, 0)
	pageRes, err := query.FilteredPaginate(store, pageReq, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var policy types.Policy
		k.cdc.MustUnmarshal(value, &policy)

		if !k.matchesFilter(policy, filter) {
			return false, nil
		}

		if accumulate {
			policies = append(policies, policy)
		}

		return true, nil
	})
	if err != nil {
		return nil, nil, err
	}

	return policies, pageRes, nil
}

// ExportPolicies collects all stored policies.
func (k Keeper) ExportPolicies(ctx sdk.Context) []types.Policy {
	policies := make([]types.Policy, 0)
	k.IteratePolicies(ctx, func(policy types.Policy) bool {
		policies = append(policies, policy)
		return false
	})
	return policies
}

func (k Keeper) matchesFilter(policy types.Policy, filter *types.PolicyFilter) bool {
	if filter == nil {
		return true
	}
	if filter.Owner != "" && policy.Owner != filter.Owner {
		return false
	}
	if filter.PolicyType != "" && policy.PolicyType != filter.PolicyType {
		return false
	}
	if filter.Status != types.PolicyStatus_POLICY_STATUS_UNSPECIFIED && policy.Status != filter.Status {
		return false
	}
	return true
}

func (k Keeper) assertPolicyEditable(ctx sdk.Context, authority sdk.AccAddress, policy types.Policy) error {
	if authority.String() == policy.Owner {
		return nil
	}

	if k.hasRole(ctx, authority, rolestypes.Role_ROLE_INSURER) {
		return nil
	}

	return types.ErrUnauthorized
}

func (k Keeper) hasRole(ctx sdk.Context, addr sdk.AccAddress, role rolestypes.Role) bool {
	if k.rolesKeeper == nil {
		return true
	}
	return k.rolesKeeper.HasRole(ctx, addr, role)
}

func (k Keeper) isPolicyTypeAllowed(params types.Params, policyType string) bool {
	if len(params.AllowedPolicyTypes) == 0 {
		return true
	}

	for _, allowed := range params.AllowedPolicyTypes {
		if allowed == policyType {
			return true
		}
	}
	return false
}

func (k Keeper) nextPolicyID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextPolicyIDKey)
	var id uint64 = 1
	if bz != nil {
		id = types.BytesToUint64(bz)
	}

	store.Set(types.NextPolicyIDKey, types.Uint64ToBytes(id+1))
	return id
}

func (k Keeper) setPolicy(ctx sdk.Context, policy types.Policy) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&policy)
	store.Set(types.PolicyKey(policy.Id), bz)
}

// SetPolicy exposes policy persistence for genesis/state sync usage.
func (k Keeper) SetPolicy(ctx sdk.Context, policy types.Policy) {
	k.setPolicy(ctx, policy)
}

// SetNextPolicyID sets the next policy identifier counter.
func (k Keeper) SetNextPolicyID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NextPolicyIDKey, types.Uint64ToBytes(id))
}

// GetNextPolicyID returns the current next policy identifier.
func (k Keeper) GetNextPolicyID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextPolicyIDKey)
	if bz == nil {
		return 1
	}
	return types.BytesToUint64(bz)
}

// GetAllPolicies returns all stored policies.
func (k Keeper) GetAllPolicies(ctx sdk.Context) []types.Policy {
	return k.ExportPolicies(ctx)
}

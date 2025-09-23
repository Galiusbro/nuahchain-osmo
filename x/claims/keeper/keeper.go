package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v30/x/claims/types"
	policytypes "github.com/osmosis-labs/osmosis/v30/x/policy/types"
	rolestypes "github.com/osmosis-labs/osmosis/v30/x/roles/types"
)

// Keeper manages claim state and orchestrates review workflows.
type Keeper struct {
	cdc            codec.BinaryCodec
	storeKey       storetypes.StoreKey
	paramstore     paramtypes.Subspace
	rolesKeeper    types.RolesKeeper
	policyKeeper   types.PolicyKeeper
	treasuryKeeper types.TreasuryKeeper
}

// NewKeeper creates a new claims keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	ps paramtypes.Subspace,
	roles types.RolesKeeper,
	policy types.PolicyKeeper,
	treasury types.TreasuryKeeper,
) Keeper {
	if ps.Name() != "" && !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:            cdc,
		storeKey:       key,
		paramstore:     ps,
		rolesKeeper:    roles,
		policyKeeper:   policy,
		treasuryKeeper: treasury,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams retrieves module params.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	if k.paramstore.Name() == "" {
		return types.DefaultParams()
	}

	var params types.Params
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams stores module params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if err := params.Validate(); err != nil {
		panic(err)
	}

	if k.paramstore.Name() == "" {
		return
	}

	k.paramstore.SetParamSet(ctx, &params)
}

// SubmitClaim records a new claim and returns the created claim.
func (k Keeper) SubmitClaim(ctx sdk.Context, claimant sdk.AccAddress, policyID uint64, amount sdk.Coin, description string, evidence []types.ClaimEvidence) (types.Claim, error) {
	if amount.Amount.IsZero() || amount.Amount.IsNegative() {
		return types.Claim{}, fmt.Errorf("claim amount must be positive")
	}

	params := k.GetParams(ctx)

	if params.MaxOpenClaimsPerPolicy > 0 {
		open := k.countOpenClaims(ctx, policyID)
		if open >= params.MaxOpenClaimsPerPolicy {
			return types.Claim{}, types.ErrMaxOpenClaimsExceeded
		}
	}

	var policy policytypes.Policy
	var policyFound bool
	if k.policyKeeper != nil {
		policy, policyFound = k.policyKeeper.GetPolicy(ctx, policyID)
		if !policyFound {
			return types.Claim{}, types.ErrInvalidPolicy
		}
		if policy.Status != policytypes.PolicyStatus_POLICY_STATUS_ACTIVE {
			return types.Claim{}, types.ErrInvalidPolicy
		}
		if policy.Owner != claimant.String() {
			// Allow claimants other than owner, but ensure policy is active; optionally skip owner check.
		}
	}

	claimID := k.nextClaimID(ctx)
	submittedAt := ctx.BlockTime().UTC()

	claim := types.Claim{
		Id:             claimID,
		PolicyId:       policyID,
		Claimant:       claimant.String(),
		Reporter:       claimant.String(),
		Amount:         amount,
		Description:    description,
		Evidence:       append([]types.ClaimEvidence(nil), evidence...),
		Status:         types.ClaimStatus_CLAIM_STATUS_PENDING,
		SubmittedAt:    &submittedAt,
		TreasuryPoolId: "",
	}

	if policyFound {
		claim.TreasuryPoolId = policy.TreasuryPoolId
	}

	if k.shouldAutoApprove(params, policy) {
		resolvedAt := submittedAt
		claim.Status = types.ClaimStatus_CLAIM_STATUS_APPROVED
		claim.Decision = &types.ClaimDecision{
			Reviewer:  "auto",
			Status:    types.ClaimStatus_CLAIM_STATUS_APPROVED,
			Reason:    "auto-approved",
			DecidedAt: &resolvedAt,
		}
	}

	k.setClaim(ctx, claim)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeClaimSubmitted,
		sdk.NewAttribute(types.AttributeKeyClaimID, fmt.Sprintf("%d", claimID)),
		sdk.NewAttribute(types.AttributeKeyPolicyID, fmt.Sprintf("%d", policyID)),
		sdk.NewAttribute(types.AttributeKeyStatus, claim.Status.String()),
	))

	return claim, nil
}

// ReviewClaim reviews a pending claim and updates its status.
func (k Keeper) ReviewClaim(ctx sdk.Context, reviewer sdk.AccAddress, claimID uint64, decision types.ClaimStatus, reason string) (types.Claim, error) {
	if !k.hasAnyRole(ctx, reviewer, rolestypes.Role_ROLE_CLAIMS_REVIEWER, rolestypes.Role_ROLE_INSURER) {
		return types.Claim{}, types.ErrUnauthorized
	}

	claim, found := k.GetClaim(ctx, claimID)
	if !found {
		return types.Claim{}, types.ErrClaimNotFound
	}

	switch claim.Status {
	case types.ClaimStatus_CLAIM_STATUS_PENDING:
		if decision != types.ClaimStatus_CLAIM_STATUS_APPROVED && decision != types.ClaimStatus_CLAIM_STATUS_REJECTED {
			return types.Claim{}, types.ErrInvalidStatusChange
		}
	case types.ClaimStatus_CLAIM_STATUS_APPROVED:
		if decision != types.ClaimStatus_CLAIM_STATUS_REJECTED {
			return types.Claim{}, types.ErrInvalidStatusChange
		}
	default:
		return types.Claim{}, types.ErrClaimAlreadyResolved
	}

	decidedAt := ctx.BlockTime().UTC()
	claim.Status = decision
	claim.Decision = &types.ClaimDecision{
		Reviewer:  reviewer.String(),
		Status:    decision,
		Reason:    reason,
		DecidedAt: &decidedAt,
	}
	if decision == types.ClaimStatus_CLAIM_STATUS_REJECTED {
		claim.ResolvedAt = &decidedAt
	}

	k.setClaim(ctx, claim)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeClaimReviewed,
		sdk.NewAttribute(types.AttributeKeyClaimID, fmt.Sprintf("%d", claimID)),
		sdk.NewAttribute(types.AttributeKeyReviewer, reviewer.String()),
		sdk.NewAttribute(types.AttributeKeyStatus, decision.String()),
	))

	return claim, nil
}

// AddClaimEvidence appends new evidence to a claim.
func (k Keeper) AddClaimEvidence(ctx sdk.Context, authority sdk.AccAddress, claimID uint64, evidence types.ClaimEvidence) (types.Claim, error) {
	if !k.hasAnyRole(ctx, authority, rolestypes.Role_ROLE_CLAIMS_REVIEWER, rolestypes.Role_ROLE_INSURER) {
		return types.Claim{}, types.ErrUnauthorized
	}

	claim, found := k.GetClaim(ctx, claimID)
	if !found {
		return types.Claim{}, types.ErrClaimNotFound
	}

	claim.Evidence = append(claim.Evidence, evidence)
	k.setClaim(ctx, claim)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeClaimEvidence,
		sdk.NewAttribute(types.AttributeKeyClaimID, fmt.Sprintf("%d", claimID)),
	))

	return claim, nil
}

// ExecuteClaimPayout processes an approved claim and marks it as paid.
func (k Keeper) ExecuteClaimPayout(ctx sdk.Context, authority sdk.AccAddress, claimID uint64, recipient sdk.AccAddress) (types.Claim, error) {
	if !k.hasAnyRole(ctx, authority, rolestypes.Role_ROLE_TREASURY_MANAGER, rolestypes.Role_ROLE_INSURER) {
		return types.Claim{}, types.ErrUnauthorized
	}

	claim, found := k.GetClaim(ctx, claimID)
	if !found {
		return types.Claim{}, types.ErrClaimNotFound
	}

	if claim.Status != types.ClaimStatus_CLAIM_STATUS_APPROVED {
		return types.Claim{}, types.ErrInvalidStatusChange
	}

	if k.treasuryKeeper == nil {
		return types.Claim{}, types.ErrUnsupportedPayout
	}

	if err := k.treasuryKeeper.DisburseClaim(ctx, claim.TreasuryPoolId, recipient, claim.Amount); err != nil {
		return types.Claim{}, err
	}

	now := ctx.BlockTime().UTC()
	claim.Status = types.ClaimStatus_CLAIM_STATUS_PAID
	claim.ResolvedAt = &now
	if claim.Decision != nil {
		claim.Decision.Status = types.ClaimStatus_CLAIM_STATUS_PAID
		claim.Decision.DecidedAt = &now
	}

	k.setClaim(ctx, claim)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeClaimPaid,
		sdk.NewAttribute(types.AttributeKeyClaimID, fmt.Sprintf("%d", claimID)),
		sdk.NewAttribute(types.AttributeKeyStatus, claim.Status.String()),
	))

	return claim, nil
}

// GetClaim fetches a claim by identifier.
func (k Keeper) GetClaim(ctx sdk.Context, claimID uint64) (types.Claim, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ClaimKey(claimID))
	if bz == nil {
		return types.Claim{}, false
	}

	var claim types.Claim
	k.cdc.MustUnmarshal(bz, &claim)
	return claim, true
}

// GetClaims returns claims matching the provided filter and pagination.
func (k Keeper) GetClaims(ctx sdk.Context, filter *types.QueryClaimsRequest, pageReq *query.PageRequest) ([]types.Claim, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ClaimKeyPrefix)

	claims := make([]types.Claim, 0)
	pageRes, err := query.FilteredPaginate(store, pageReq, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var claim types.Claim
		k.cdc.MustUnmarshal(value, &claim)

		if !matchesFilter(claim, filter) {
			return false, nil
		}

		if accumulate {
			claims = append(claims, claim)
		}
		return true, nil
	})
	if err != nil {
		return nil, nil, err
	}

	return claims, pageRes, nil
}

// ExportClaims returns all stored claims.
func (k Keeper) ExportClaims(ctx sdk.Context) []types.Claim {
	claims := make([]types.Claim, 0)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ClaimKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var claim types.Claim
		k.cdc.MustUnmarshal(iterator.Value(), &claim)
		claims = append(claims, claim)
	}

	return claims
}

func matchesFilter(claim types.Claim, filter *types.QueryClaimsRequest) bool {
	if filter == nil {
		return true
	}
	if filter.PolicyId != 0 && claim.PolicyId != filter.PolicyId {
		return false
	}
	if filter.Claimant != "" && claim.Claimant != filter.Claimant {
		return false
	}
	if filter.Status != types.ClaimStatus_CLAIM_STATUS_UNSPECIFIED && claim.Status != filter.Status {
		return false
	}
	return true
}

func (k Keeper) hasAnyRole(ctx sdk.Context, addr sdk.AccAddress, roles ...rolestypes.Role) bool {
	if k.rolesKeeper == nil {
		return true
	}
	for _, role := range roles {
		if k.rolesKeeper.HasRole(ctx, addr, role) {
			return true
		}
	}
	return false
}

func (k Keeper) shouldAutoApprove(params types.Params, policy policytypes.Policy) bool {
	if len(params.AutoApprovalPolicyTypes) == 0 {
		return false
	}
	for _, t := range params.AutoApprovalPolicyTypes {
		if policy.PolicyType == t {
			return true
		}
	}
	return false
}

func (k Keeper) countOpenClaims(ctx sdk.Context, policyID uint64) uint64 {
	var count uint64
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ClaimKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var claim types.Claim
		k.cdc.MustUnmarshal(iterator.Value(), &claim)
		if claim.PolicyId == policyID && claim.Status == types.ClaimStatus_CLAIM_STATUS_PENDING {
			count++
		}
	}
	return count
}

func (k Keeper) nextClaimID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextClaimIDKey)
	var id uint64 = 1
	if bz != nil {
		id = types.BytesToUint64(bz)
	}

	store.Set(types.NextClaimIDKey, types.Uint64ToBytes(id+1))
	return id
}

func (k Keeper) setClaim(ctx sdk.Context, claim types.Claim) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ClaimKey(claim.Id), k.cdc.MustMarshal(&claim))
}

// SetClaim is exposed for genesis initialization.
func (k Keeper) SetClaim(ctx sdk.Context, claim types.Claim) {
	k.setClaim(ctx, claim)
}

// SetNextClaimID sets the next claim identifier (genesis helper).
func (k Keeper) SetNextClaimID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NextClaimIDKey, types.Uint64ToBytes(id))
}

// GetNextClaimID returns the persisted next claim identifier.
func (k Keeper) GetNextClaimID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextClaimIDKey)
	if bz == nil {
		return 1
	}
	return types.BytesToUint64(bz)
}

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

	"github.com/osmosis-labs/osmosis/v30/x/premium/types"
	rolestypes "github.com/osmosis-labs/osmosis/v30/x/roles/types"
)

// Keeper controls premium plan state.
type Keeper struct {
	cdc         codec.BinaryCodec
	storeKey    storetypes.StoreKey
	paramstore  paramtypes.Subspace
	rolesKeeper types.RolesKeeper
}

// NewKeeper creates a premium keeper instance.
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

// Logger returns a module logger.
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

// CreatePremiumPlan creates a new premium plan.
func (k Keeper) CreatePremiumPlan(ctx sdk.Context, authority sdk.AccAddress, policyID uint64, payer sdk.AccAddress, amount sdk.Coin, schedule types.PremiumSchedule, treasuryPoolID string) (types.PremiumPlan, error) {
	if !k.hasRole(ctx, authority, rolestypes.Role_ROLE_INSURER) {
		return types.PremiumPlan{}, types.ErrPremiumUnauthorized
	}

	if err := validateSchedule(schedule); err != nil {
		return types.PremiumPlan{}, err
	}

	if amount.Amount.IsZero() || amount.Amount.IsNegative() {
		return types.PremiumPlan{}, fmt.Errorf("premium amount must be positive")
	}

	planID := k.nextPlanID(ctx)

	nextDue := ctx.BlockTime().Add(time.Duration(schedule.PeriodSeconds) * time.Second).UTC()

	plan := types.PremiumPlan{
		Id:             planID,
		PolicyId:       policyID,
		Payer:          payer.String(),
		Amount:         amount,
		Schedule:       schedule,
		Status:         types.PremiumPlanStatus_PREMIUM_PLAN_STATUS_ACTIVE,
		NextDueTime:    &nextDue,
		PaymentsMade:   0,
		TreasuryPoolId: treasuryPoolID,
	}

	k.setPlan(ctx, plan)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePremiumPlanCreated,
		sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprintf("%d", planID)),
		sdk.NewAttribute(types.AttributeKeyPayer, payer.String()),
	))

	return plan, nil
}

// RecordPremiumPayment stores a payment and updates plan state.
func (k Keeper) RecordPremiumPayment(ctx sdk.Context, payer sdk.AccAddress, planID uint64, amount sdk.Coin) (types.PremiumPayment, error) {
	plan, found := k.GetPremiumPlan(ctx, planID)
	if !found {
		return types.PremiumPayment{}, types.ErrPremiumPlanNotFound
	}

	if payer.String() != plan.Payer {
		return types.PremiumPayment{}, types.ErrPremiumUnauthorized
	}

	if plan.Amount.Amount.IsZero() {
		return types.PremiumPayment{}, fmt.Errorf("plan amount not configured")
	}

	if amount.Denom != plan.Amount.Denom {
		return types.PremiumPayment{}, fmt.Errorf("payment denom mismatch: expected %s", plan.Amount.Denom)
	}

	paymentID := k.nextPaymentID(ctx)
	paidAt := ctx.BlockTime().UTC()

	payment := types.PremiumPayment{
		Id:     paymentID,
		PlanId: planID,
		Payer:  payer.String(),
		Amount: amount,
		PaidAt: &paidAt,
	}

	k.setPayment(ctx, payment)

	plan.PaymentsMade++
	paidAtCopy := paidAt
	plan.LastPaymentTime = &paidAtCopy

	if plan.Schedule.PeriodSeconds > 0 {
		nextDue := ctx.BlockTime().Add(time.Duration(plan.Schedule.PeriodSeconds) * time.Second).UTC()
		plan.NextDueTime = &nextDue
	}

	if plan.Schedule.MaxPayments > 0 && plan.PaymentsMade >= plan.Schedule.MaxPayments {
		plan.Status = types.PremiumPlanStatus_PREMIUM_PLAN_STATUS_COMPLETED
	}

	// Clear overdue marker
	k.deleteOverdue(ctx, planID)

	k.setPlan(ctx, plan)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePremiumPaymentMade,
		sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprintf("%d", planID)),
		sdk.NewAttribute(types.AttributeKeyPaymentID, fmt.Sprintf("%d", paymentID)),
		sdk.NewAttribute(types.AttributeKeyPayer, payer.String()),
	))

	return payment, nil
}

// MarkPremiumOverdue flags a plan as overdue.
func (k Keeper) MarkPremiumOverdue(ctx sdk.Context, authority sdk.AccAddress, planID uint64, reason string) (types.PremiumOverdue, error) {
	if !k.hasRole(ctx, authority, rolestypes.Role_ROLE_INSURER) {
		return types.PremiumOverdue{}, types.ErrPremiumUnauthorized
	}

	plan, found := k.GetPremiumPlan(ctx, planID)
	if !found {
		return types.PremiumOverdue{}, types.ErrPremiumPlanNotFound
	}

	overdue := types.PremiumOverdue{
		PlanId: planID,
		DueAt:  plan.NextDueTime,
		Reason: reason,
	}

	k.setOverdue(ctx, overdue)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePremiumMarkedOverdue,
		sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprintf("%d", planID)),
	))

	return overdue, nil
}

// GetPremiumPlan retrieves a plan by id.
func (k Keeper) GetPremiumPlan(ctx sdk.Context, planID uint64) (types.PremiumPlan, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PremiumPlanKey(planID))
	if bz == nil {
		return types.PremiumPlan{}, false
	}

	var plan types.PremiumPlan
	k.cdc.MustUnmarshal(bz, &plan)
	return plan, true
}

// ListPremiumPlans returns plans with optional filters and pagination.
func (k Keeper) ListPremiumPlans(ctx sdk.Context, policyID uint64, payer string, pageReq *query.PageRequest) ([]types.PremiumPlan, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PremiumPlanKeyPrefix)

	plans := make([]types.PremiumPlan, 0)
	pageRes, err := query.FilteredPaginate(store, pageReq, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var plan types.PremiumPlan
		k.cdc.MustUnmarshal(value, &plan)

		if policyID != 0 && plan.PolicyId != policyID {
			return false, nil
		}

		if payer != "" && plan.Payer != payer {
			return false, nil
		}

		if accumulate {
			plans = append(plans, plan)
		}
		return true, nil
	})
	if err != nil {
		return nil, nil, err
	}

	return plans, pageRes, nil
}

// ListPremiumPayments returns payments for a plan.
func (k Keeper) ListPremiumPayments(ctx sdk.Context, planID uint64, pageReq *query.PageRequest) ([]types.PremiumPayment, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PremiumPaymentKeyPrefix)

	payments := make([]types.PremiumPayment, 0)
	pageRes, err := query.FilteredPaginate(store, pageReq, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var payment types.PremiumPayment
		k.cdc.MustUnmarshal(value, &payment)

		if planID != 0 && payment.PlanId != planID {
			return false, nil
		}

		if accumulate {
			payments = append(payments, payment)
		}
		return true, nil
	})
	if err != nil {
		return nil, nil, err
	}

	return payments, pageRes, nil
}

// ExportPlans exports stored plans.
func (k Keeper) ExportPlans(ctx sdk.Context) []types.PremiumPlan {
	plans := make([]types.PremiumPlan, 0)
	k.IteratePlans(ctx, func(plan types.PremiumPlan) bool {
		plans = append(plans, plan)
		return false
	})
	return plans
}

// ExportPayments exports stored payments.
func (k Keeper) ExportPayments(ctx sdk.Context) []types.PremiumPayment {
	payments := make([]types.PremiumPayment, 0)
	k.IteratePayments(ctx, func(payment types.PremiumPayment) bool {
		payments = append(payments, payment)
		return false
	})
	return payments
}

// ExportOverdue exports overdue records.
func (k Keeper) ExportOverdue(ctx sdk.Context) []types.PremiumOverdue {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OverdueKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	overdue := make([]types.PremiumOverdue, 0)
	for ; iterator.Valid(); iterator.Next() {
		var record types.PremiumOverdue
		k.cdc.MustUnmarshal(iterator.Value(), &record)
		overdue = append(overdue, record)
	}

	return overdue
}

// IteratePlans iterates over plans.
func (k Keeper) IteratePlans(ctx sdk.Context, cb func(types.PremiumPlan) bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PremiumPlanKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var plan types.PremiumPlan
		k.cdc.MustUnmarshal(iterator.Value(), &plan)
		if cb(plan) {
			return
		}
	}
}

// IteratePayments iterates over payments.
func (k Keeper) IteratePayments(ctx sdk.Context, cb func(types.PremiumPayment) bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PremiumPaymentKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var payment types.PremiumPayment
		k.cdc.MustUnmarshal(iterator.Value(), &payment)
		if cb(payment) {
			return
		}
	}
}

func (k Keeper) hasRole(ctx sdk.Context, addr sdk.AccAddress, role rolestypes.Role) bool {
	if k.rolesKeeper == nil {
		return true
	}
	return k.rolesKeeper.HasRole(ctx, addr, role)
}

func validateSchedule(schedule types.PremiumSchedule) error {
	if schedule.PeriodSeconds == 0 {
		return types.ErrInvalidPremiumSchedule
	}
	return nil
}

func (k Keeper) nextPlanID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextPlanIDKey)
	var id uint64 = 1
	if bz != nil {
		id = types.BytesToUint64(bz)
	}

	store.Set(types.NextPlanIDKey, types.Uint64ToBytes(id+1))
	return id
}

func (k Keeper) nextPaymentID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextPaymentIDKey)
	var id uint64 = 1
	if bz != nil {
		id = types.BytesToUint64(bz)
	}

	store.Set(types.NextPaymentIDKey, types.Uint64ToBytes(id+1))
	return id
}

func (k Keeper) setPlan(ctx sdk.Context, plan types.PremiumPlan) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.PremiumPlanKey(plan.Id), k.cdc.MustMarshal(&plan))
}

func (k Keeper) setPayment(ctx sdk.Context, payment types.PremiumPayment) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.PremiumPaymentKey(payment.Id), k.cdc.MustMarshal(&payment))
}

func (k Keeper) setOverdue(ctx sdk.Context, overdue types.PremiumOverdue) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.OverdueKey(overdue.PlanId), k.cdc.MustMarshal(&overdue))
}

func (k Keeper) deleteOverdue(ctx sdk.Context, planID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.OverdueKey(planID))
}

// SetPremiumPlan persists a plan (genesis helper).
func (k Keeper) SetPremiumPlan(ctx sdk.Context, plan types.PremiumPlan) {
	k.setPlan(ctx, plan)
}

// SetPremiumPayment persists a payment (genesis helper).
func (k Keeper) SetPremiumPayment(ctx sdk.Context, payment types.PremiumPayment) {
	k.setPayment(ctx, payment)
}

// SetPremiumOverdue persists overdue info (genesis helper).
func (k Keeper) SetPremiumOverdue(ctx sdk.Context, overdue types.PremiumOverdue) {
	k.setOverdue(ctx, overdue)
}

// SetNextPlanID sets next plan id (genesis helper).
func (k Keeper) SetNextPlanID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NextPlanIDKey, types.Uint64ToBytes(id))
}

// SetNextPaymentID sets next payment id (genesis helper).
func (k Keeper) SetNextPaymentID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NextPaymentIDKey, types.Uint64ToBytes(id))
}

// GetNextPlanID returns next plan id.
func (k Keeper) GetNextPlanID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextPlanIDKey)
	if bz == nil {
		return 1
	}
	return types.BytesToUint64(bz)
}

// GetNextPaymentID returns next payment id.
func (k Keeper) GetNextPaymentID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextPaymentIDKey)
	if bz == nil {
		return 1
	}
	return types.BytesToUint64(bz)
}

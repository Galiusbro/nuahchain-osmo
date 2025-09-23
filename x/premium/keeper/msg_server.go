package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/premium/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl creates a new Msg server instance.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return msgServer{Keeper: keeper}
}

func (m msgServer) CreatePremiumPlan(goCtx context.Context, msg *types.MsgCreatePremiumPlan) (*types.MsgCreatePremiumPlanResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	payer, err := sdk.AccAddressFromBech32(msg.Payer)
	if err != nil {
		return nil, err
	}

	amount := sdk.NewCoin(msg.Amount.Denom, msg.Amount.Amount)

	plan, err := m.Keeper.CreatePremiumPlan(sdkCtx, authority, msg.PolicyId, payer, amount, msg.Schedule, msg.TreasuryPoolId)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreatePremiumPlanResponse{PlanId: plan.Id}, nil
}

func (m msgServer) RecordPremiumPayment(goCtx context.Context, msg *types.MsgRecordPremiumPayment) (*types.MsgRecordPremiumPaymentResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	payer, err := sdk.AccAddressFromBech32(msg.Payer)
	if err != nil {
		return nil, err
	}

	amount := sdk.NewCoin(msg.Amount.Denom, msg.Amount.Amount)

	payment, err := m.Keeper.RecordPremiumPayment(sdkCtx, payer, msg.PlanId, amount)
	if err != nil {
		return nil, err
	}

	plan, _ := m.Keeper.GetPremiumPlan(sdkCtx, msg.PlanId)

	return &types.MsgRecordPremiumPaymentResponse{
		PaymentId:   payment.Id,
		NextDueTime: plan.NextDueTime,
	}, nil
}

func (m msgServer) MarkPremiumOverdue(goCtx context.Context, msg *types.MsgMarkPremiumOverdue) (*types.MsgMarkPremiumOverdueResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	_, err = m.Keeper.MarkPremiumOverdue(sdkCtx, authority, msg.PlanId, msg.Reason)
	if err != nil {
		return nil, err
	}

	return &types.MsgMarkPremiumOverdueResponse{}, nil
}

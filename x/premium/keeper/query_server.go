package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/premium/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	Keeper
}

// NewQueryServer creates a new query server instance.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return queryServer{Keeper: keeper}
}

func (q queryServer) PremiumPlan(goCtx context.Context, req *types.QueryPremiumPlanRequest) (*types.QueryPremiumPlanResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	plan, found := q.GetPremiumPlan(sdkCtx, req.PlanId)
	if !found {
		return nil, status.Error(codes.NotFound, "plan not found")
	}

	return &types.QueryPremiumPlanResponse{Plan: &plan}, nil
}

func (q queryServer) PremiumPlans(goCtx context.Context, req *types.QueryPremiumPlansRequest) (*types.QueryPremiumPlansResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	plans, pageRes, err := q.ListPremiumPlans(sdkCtx, req.PolicyId, req.Payer, req.Pagination)
	if err != nil {
		return nil, err
	}

	// convert []types.PremiumPlan to []*types.PremiumPlan
	planPtrs := make([]*types.PremiumPlan, len(plans))
	for i := range plans {
		planPtrs[i] = &plans[i]
	}

	return &types.QueryPremiumPlansResponse{
		Plans:      planPtrs,
		Pagination: pageRes,
	}, nil
}

func (q queryServer) PremiumPayments(goCtx context.Context, req *types.QueryPremiumPaymentsRequest) (*types.QueryPremiumPaymentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	payments, pageRes, err := q.ListPremiumPayments(sdkCtx, req.PlanId, req.Pagination)
	if err != nil {
		return nil, err
	}

	// convert []types.PremiumPayment to []*types.PremiumPayment
	paymentPtrs := make([]*types.PremiumPayment, len(payments))
	for i := range payments {
		paymentPtrs[i] = &payments[i]
	}

	return &types.QueryPremiumPaymentsResponse{
		Payments:   paymentPtrs,
		Pagination: pageRes,
	}, nil
}

func (q queryServer) Params(goCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	params := q.GetParams(sdkCtx)
	return &types.QueryParamsResponse{Params: &params}, nil
}

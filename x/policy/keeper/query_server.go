package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/policy/types"
)

var _ types.QueryServer = queryServer{}

// queryServer implements the Query service.
type queryServer struct {
	Keeper
}

// NewQueryServer returns a new query server instance.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return queryServer{Keeper: keeper}
}

func (q queryServer) Policy(goCtx context.Context, req *types.QueryPolicyRequest) (*types.QueryPolicyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	policy, found := q.GetPolicy(sdkCtx, req.PolicyId)
	if !found {
		return nil, status.Error(codes.NotFound, "policy not found")
	}

	return &types.QueryPolicyResponse{Policy: &policy}, nil
}

func (q queryServer) Policies(goCtx context.Context, req *types.QueryPoliciesRequest) (*types.QueryPoliciesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	policies, pageRes, err := q.GetPolicies(sdkCtx, req.Filter, req.Pagination)
	if err != nil {
		return nil, err
	}

	// convert []types.Policy to []*types.Policy to match response type
	policyPtrs := make([]*types.Policy, len(policies))
	for i := range policies {
		policyPtrs[i] = &policies[i]
	}

	return &types.QueryPoliciesResponse{
		Policies:   policyPtrs,
		Pagination: pageRes,
	}, nil
}

func (q queryServer) Params(goCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	params := q.GetParams(sdkCtx)
	return &types.QueryParamsResponse{Params: &params}, nil
}

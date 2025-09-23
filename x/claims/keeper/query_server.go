package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/claims/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	Keeper
}

// NewQueryServer creates a new query service instance.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return queryServer{Keeper: keeper}
}

func (q queryServer) Claim(goCtx context.Context, req *types.QueryClaimRequest) (*types.QueryClaimResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	claim, found := q.GetClaim(sdkCtx, req.ClaimId)
	if !found {
		return nil, status.Error(codes.NotFound, "claim not found")
	}

	return &types.QueryClaimResponse{Claim: &claim}, nil
}

func (q queryServer) Claims(goCtx context.Context, req *types.QueryClaimsRequest) (*types.QueryClaimsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	claims, pageRes, err := q.GetClaims(sdkCtx, req, req.Pagination)
	if err != nil {
		return nil, err
	}

	// convert []types.Claim to []*types.Claim to match response type
	claimPtrs := make([]*types.Claim, len(claims))
	for i := range claims {
		claimPtrs[i] = &claims[i]
	}

	return &types.QueryClaimsResponse{
		Claims:     claimPtrs,
		Pagination: pageRes,
	}, nil
}

func (q queryServer) Params(goCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	params := q.GetParams(sdkCtx)
	return &types.QueryParamsResponse{Params: &params}, nil
}

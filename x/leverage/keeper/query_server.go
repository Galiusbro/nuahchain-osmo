package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/leverage/types"
)

type queryServer struct {
	Keeper
}

var _ types.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{Keeper: k}
}

func (q queryServer) Position(goCtx context.Context, req *types.QueryPositionRequest) (*types.QueryPositionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id must be positive")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	position, found := q.GetPosition(ctx, req.Id)
	if !found {
		return nil, status.Error(codes.NotFound, "position not found")
	}
	return &types.QueryPositionResponse{Position: position}, nil
}

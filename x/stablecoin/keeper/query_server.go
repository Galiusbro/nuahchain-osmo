package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/stablecoin/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	Keeper
}

// NewQueryServer returns a new QueryServer backed by the keeper.
func NewQueryServer(k Keeper) types.QueryServer {
	return queryServer{Keeper: k}
}

// Stats handles the Query/Stats RPC.
func (q queryServer) Stats(goCtx context.Context, _ *types.QueryStatsRequest) (*types.QueryStatsResponse, error) {
	if goCtx == nil {
		return nil, status.Error(codes.Internal, "context cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	stats := q.GetStats(ctx)
	return &types.QueryStatsResponse{Stats: &stats}, nil
}

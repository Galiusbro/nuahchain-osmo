package keeper

import (
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/risk/types"
)

type queryServer struct {
	Keeper
}

var _ types.QueryServer = queryServer{}

// NewQueryServerImpl returns a query server backed by the keeper.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{Keeper: k}
}

// RiskParams returns stored leverage limits for the given symbol.
func (q queryServer) RiskParams(goCtx context.Context, req *types.QueryRiskParamsRequest) (*types.QueryRiskParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if strings.TrimSpace(req.Symbol) == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params, found := q.GetRiskParams(ctx, req.Symbol)
	if !found {
		return nil, status.Error(codes.NotFound, "risk params not found")
	}

	return &types.QueryRiskParamsResponse{Params: params}, nil
}

package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
)

type queryServer struct {
	Keeper
}

var _ types.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the QueryServer interface.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{Keeper: k}
}

func (s queryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := s.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

func (s queryServer) GlobalPause(goCtx context.Context, req *types.QueryGlobalPauseRequest) (*types.QueryGlobalPauseResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	info, found := s.getGlobalPause(ctx)
	if !found {
		return &types.QueryGlobalPauseResponse{}, nil
	}
	return &types.QueryGlobalPauseResponse{Status: &info}, nil
}

func (s queryServer) TokenPause(goCtx context.Context, req *types.QueryTokenPauseRequest) (*types.QueryTokenPauseResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if err := sdk.ValidateDenom(req.Denom); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	info, found := s.getTokenPause(ctx, req.Denom)
	if !found {
		return &types.QueryTokenPauseResponse{}, nil
	}
	return &types.QueryTokenPauseResponse{Status: &info}, nil
}

func (s queryServer) Freeze(goCtx context.Context, req *types.QueryFreezeRequest) (*types.QueryFreezeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.TargetType == types.FreezeTargetType_FREEZE_TARGET_TYPE_UNSPECIFIED || req.Target == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid freeze target")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	info, found := s.getFreezeInfo(ctx, req.TargetType, req.Target)
	if !found {
		return &types.QueryFreezeResponse{}, nil
	}
	return &types.QueryFreezeResponse{Status: &info}, nil
}

func (s queryServer) PendingParams(goCtx context.Context, req *types.QueryPendingParamsRequest) (*types.QueryPendingParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	pending, found := s.getPendingParams(ctx)
	if !found {
		return &types.QueryPendingParamsResponse{}, nil
	}
	return &types.QueryPendingParamsResponse{PendingParams: &pending}, nil
}

func (s queryServer) EmergencyActions(goCtx context.Context, req *types.QueryEmergencyActionsRequest) (*types.QueryEmergencyActionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	actions := make([]*types.EmergencyAction, 0)
	limit := req.Limit
	s.iterateEmergencyActions(ctx, limit, func(action types.EmergencyAction) bool {
		actionCopy := action
		actions = append(actions, &actionCopy)
		return false
	})
	return &types.QueryEmergencyActionsResponse{Actions: actions}, nil
}

func (s queryServer) EmergencyConfig(goCtx context.Context, req *types.QueryEmergencyConfigRequest) (*types.QueryEmergencyConfigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	config := s.getEmergencyConfig(ctx)
	return &types.QueryEmergencyConfigResponse{Config: &config}, nil
}

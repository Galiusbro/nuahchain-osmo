package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/pegkeeper/types"
)

var _ types.QueryServer = Keeper{}

// Params returns the total set of pegkeeper parameters.
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

// PegState returns the current peg state.
func (k Keeper) PegState(c context.Context, req *types.QueryPegStateRequest) (*types.QueryPegStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	params := k.GetParams(ctx)

	// Get current price from USD Oracle
	currentPrice := "1.0" // Default value
	if usdPrice, found := k.usdOracleKeeper.GetCurrentPrice(ctx); found {
		currentPrice = usdPrice.Price.String()
	}

	// Calculate deviation
	deviation := "0.0" // Simplified calculation

	pegState := &types.PegState{
		TargetDenom:        params.TargetDenom,
		ReferenceDenom:     params.ReferenceDenom,
		CurrentPrice:       currentPrice,
		TargetPrice:        params.TargetPrice,
		Deviation:          deviation,
		IsActive:           params.Enabled,
		LastAdjustmentTime: time.Time{}, // Initialize with zero time
	}

	return &types.QueryPegStateResponse{PegState: pegState}, nil
}

// AdjustmentHistory returns the supply adjustment history.
func (k Keeper) AdjustmentHistory(c context.Context, req *types.QueryAdjustmentHistoryRequest) (*types.QueryAdjustmentHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	// ctx := sdk.UnwrapSDKContext(c)

	// For now, return empty history
	// TODO: implement actual adjustment history storage and retrieval
	adjustments := []*types.SupplyAdjustment{}

	return &types.QueryAdjustmentHistoryResponse{
		Adjustments: adjustments,
		Total:       0,
	}, nil
}

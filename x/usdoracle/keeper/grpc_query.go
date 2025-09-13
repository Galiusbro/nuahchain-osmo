package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/usdoracle/types"
)

var _ types.QueryServer = Keeper{}

// Params returns the module parameters
func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// GetUSDPrice returns the current USD price
func (k Keeper) GetUSDPrice(goCtx context.Context, req *types.QueryGetUSDPriceRequest) (*types.QueryGetUSDPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	price, found := k.GetCurrentPrice(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "USD price not found")
	}

	return &types.QueryGetUSDPriceResponse{Price: price}, nil
}

// GetPriceHistory returns the price history
func (k Keeper) GetPriceHistory(goCtx context.Context, req *types.QueryGetPriceHistoryRequest) (*types.QueryGetPriceHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	history := k.GetPriceHistoryList(ctx, 100) // default limit

	return &types.QueryGetPriceHistoryResponse{Prices: history}, nil
}

// GetPriceDeviation returns the current price deviation from target (1.0)
func (k Keeper) GetPriceDeviation(goCtx context.Context, req *types.QueryGetPriceDeviationRequest) (*types.QueryGetPriceDeviationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	deviation, _ := k.CalculatePriceDeviation(ctx)

	return &types.QueryGetPriceDeviationResponse{Deviation: deviation}, nil
}

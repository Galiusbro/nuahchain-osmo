package keeper

import (
	"context"

	"cosmossdk.io/math"
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

func (k Keeper) GetAllTokenPrices(goCtx context.Context, req *types.QueryGetAllTokenPricesRequest) (*types.QueryGetAllTokenPricesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Implement get all token prices
	_ = ctx

	return &types.QueryGetAllTokenPricesResponse{
		Prices:     []types.TokenPrice{},
		Pagination: nil,
	}, nil
}

func (k Keeper) GetTokenPrice(goCtx context.Context, req *types.QueryGetTokenPriceRequest) (*types.QueryGetTokenPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Get token price using the same logic as GetTokenPriceForExchange
	tokenPrice, found := k.GetTokenPriceForExchange(goCtx, req.Denom)
	if !found {
		return nil, status.Error(codes.NotFound, "token price not found")
	}

	return &types.QueryGetTokenPriceResponse{
		Price: tokenPrice,
	}, nil
}

func (k Keeper) GetSupportedTokens(goCtx context.Context, req *types.QueryGetSupportedTokensRequest) (*types.QueryGetSupportedTokensResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Implement get supported tokens
	_ = ctx

	return &types.QueryGetSupportedTokensResponse{
		Tokens:     []types.SupportedToken{},
		Pagination: nil,
	}, nil
}

func (k Keeper) GetTokenInfo(goCtx context.Context, req *types.QueryGetTokenInfoRequest) (*types.QueryGetTokenInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Implement logic to get token info
	_ = ctx

	return &types.QueryGetTokenInfoResponse{
		Token: types.SupportedToken{},
	}, nil
}

func (k Keeper) GetTokenPriceDeviation(goCtx context.Context, req *types.QueryGetTokenPriceDeviationRequest) (*types.QueryGetTokenPriceDeviationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Implement logic to get token price deviation
	_ = ctx

	return &types.QueryGetTokenPriceDeviationResponse{
		Deviation:         math.LegacyZeroDec(),
		IsWithinThreshold: true,
		Threshold:         math.LegacyZeroDec(),
	}, nil
}

func (k Keeper) GetPriceAggregation(goCtx context.Context, req *types.QueryGetPriceAggregationRequest) (*types.QueryGetPriceAggregationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Implement logic to get price aggregation
	_ = ctx

	return &types.QueryGetPriceAggregationResponse{
		Aggregation: types.PriceAggregation{},
	}, nil
}

func (k Keeper) GetTokenPriceHistory(goCtx context.Context, req *types.QueryGetTokenPriceHistoryRequest) (*types.QueryGetTokenPriceHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Implement get token price history
	_ = ctx

	return &types.QueryGetTokenPriceHistoryResponse{
		Prices:     []types.TokenPrice{},
		Pagination: nil,
	}, nil
}

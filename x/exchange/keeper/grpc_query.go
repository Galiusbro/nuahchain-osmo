package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/exchange/types"
)

// grpcQueryServer wraps the keeper to implement the gRPC QueryServer interface
type grpcQueryServer struct {
	Keeper
}

// NewQueryServer returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return &grpcQueryServer{Keeper: keeper}
}

var _ types.QueryServer = grpcQueryServer{}

// Params returns the module parameters
func (q grpcQueryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return q.Keeper.Params(ctx, req)
}

// ExchangeRate returns the exchange rate for a specific token
func (q grpcQueryServer) ExchangeRate(goCtx context.Context, req *types.QueryExchangeRateRequest) (*types.QueryExchangeRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return q.Keeper.ExchangeRate(ctx, req)
}

// ExchangeRates returns all exchange rates with pagination
func (q grpcQueryServer) ExchangeRates(goCtx context.Context, req *types.QueryExchangeRatesRequest) (*types.QueryExchangeRatesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return q.Keeper.ExchangeRates(ctx, req)
}

// DailyLimit returns the daily exchange limit for a specific address
func (q grpcQueryServer) DailyLimit(goCtx context.Context, req *types.QueryDailyLimitRequest) (*types.QueryDailyLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return q.Keeper.DailyLimit(ctx, req)
}
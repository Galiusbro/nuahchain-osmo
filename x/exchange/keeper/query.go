package keeper

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/exchange/types"
)

// Query methods for the keeper (used by grpcQueryServer)

// Params returns the module parameters
func (k Keeper) Params(ctx sdk.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

// ExchangeRate returns the exchange rate for a specific token
func (k Keeper) ExchangeRate(ctx sdk.Context, req *types.QueryExchangeRateRequest) (*types.QueryExchangeRateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "denom cannot be empty")
	}

	exchangeRate, err := k.GetExchangeRate(ctx, req.Denom)
	if err != nil {
		return nil, status.Error(codes.NotFound, "exchange rate not found")
	}

	return &types.QueryExchangeRateResponse{ExchangeRate: exchangeRate}, nil
}

// ExchangeRates returns all exchange rates with pagination
func (k Keeper) ExchangeRates(ctx sdk.Context, req *types.QueryExchangeRatesRequest) (*types.QueryExchangeRatesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Get all exchange rates using keeper method
	exchangeRates, err := k.GetAllExchangeRates(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryExchangeRatesResponse{
		ExchangeRates: exchangeRates,
		Pagination:    nil, // Simplified without pagination for now
	}, nil
}

// DailyLimit returns the daily exchange limit for a specific address
func (k Keeper) DailyLimit(ctx sdk.Context, req *types.QueryDailyLimitRequest) (*types.QueryDailyLimitResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	// Validate address format
	if _, err := sdk.AccAddressFromBech32(req.Address); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address format")
	}

	// Get today's date
	today := k.GetTodayString(ctx)

	// Get daily limit for today
	dailyLimit, err := k.GetDailyLimit(ctx, req.Address, today)
	if err != nil {
		// If not found, return zero limit
		dailyLimit = types.DailyLimit{
			Address:            req.Address,
			TotalExchangedUsd:  math.LegacyZeroDec(),
			Date:               today,
		}
	}

	return &types.QueryDailyLimitResponse{DailyLimit: dailyLimit}, nil
}
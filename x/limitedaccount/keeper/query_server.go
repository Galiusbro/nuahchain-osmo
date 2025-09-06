package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v30/x/limitedaccount/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// Params returns the module parameters
func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: *params,
	}, nil
}

// LimitedAccount returns a limited account by address
func (k Keeper) LimitedAccount(goCtx context.Context, req *types.QueryLimitedAccountRequest) (*types.QueryLimitedAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	account, found := k.GetLimitedAccount(ctx, req.Address)
	if !found {
		return nil, status.Errorf(codes.NotFound, "limited account %s not found", req.Address)
	}

	return &types.QueryLimitedAccountResponse{
		LimitedAccount: account,
	}, nil
}

// AllLimitedAccounts returns all limited accounts
func (k Keeper) AllLimitedAccounts(goCtx context.Context, req *types.QueryAllLimitedAccountsRequest) (*types.QueryAllLimitedAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	accounts := k.GetAllLimitedAccounts(ctx)

	return &types.QueryAllLimitedAccountsResponse{
		LimitedAccounts: accounts,
	}, nil
}

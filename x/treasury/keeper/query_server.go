package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/treasury/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	Keeper
}

// NewQueryServer creates a new query service instance.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return queryServer{Keeper: keeper}
}

func (q queryServer) TreasuryPool(goCtx context.Context, req *types.QueryTreasuryPoolRequest) (*types.QueryTreasuryPoolResponse, error) {
	if req == nil || req.PoolId == "" {
		return nil, status.Error(codes.InvalidArgument, "pool id required")
	}

	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	pool, found := q.GetTreasuryPool(sdkCtx, req.PoolId)
	if !found {
		return nil, status.Error(codes.NotFound, "treasury pool not found")
	}

	return &types.QueryTreasuryPoolResponse{Pool: &pool}, nil
}

func (q queryServer) TreasuryPools(goCtx context.Context, req *types.QueryTreasuryPoolsRequest) (*types.QueryTreasuryPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	pools, pageRes, err := q.ListTreasuryPools(sdkCtx, req.Pagination)
	if err != nil {
		return nil, err
	}

	ptrPools := make([]*types.TreasuryPool, 0, len(pools))
	for i := range pools {
		ptrPools = append(ptrPools, &pools[i])
	}

	return &types.QueryTreasuryPoolsResponse{
		Pools:      ptrPools,
		Pagination: pageRes,
	}, nil
}

func (q queryServer) PoolBalances(goCtx context.Context, req *types.QueryPoolBalancesRequest) (*types.QueryPoolBalancesResponse, error) {
	if req == nil || req.PoolId == "" {
		return nil, status.Error(codes.InvalidArgument, "pool id required")
	}

	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	if _, found := q.GetTreasuryPool(sdkCtx, req.PoolId); !found {
		return nil, status.Error(codes.NotFound, "treasury pool not found")
	}

	balances := q.GetPoolBalances(sdkCtx, req.PoolId)
	reserves := q.GetPoolReserves(sdkCtx, req.PoolId)

	ptrBalances := make([]*types.PoolBalance, 0, len(balances))
	for i := range balances {
		ptrBalances = append(ptrBalances, &balances[i])
	}
	ptrReserves := make([]*types.PoolReserves, 0, len(reserves))
	for i := range reserves {
		ptrReserves = append(ptrReserves, &reserves[i])
	}

	return &types.QueryPoolBalancesResponse{
		Balances: ptrBalances,
		Reserves: ptrReserves,
	}, nil
}

func (q queryServer) Params(goCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	params := q.GetParams(sdkCtx)
	return &types.QueryParamsResponse{Params: &params}, nil
}

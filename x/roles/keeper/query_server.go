package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/roles/types"
)

var _ types.QueryServer = queryServer{}

// queryServer implements the Query service.
type queryServer struct {
	Keeper
}

// NewQueryServer creates a new query service instance.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return queryServer{Keeper: keeper}
}

func (q queryServer) RolesByAddress(goCtx context.Context, req *types.QueryRolesByAddressRequest) (*types.QueryRolesByAddressResponse, error) {
	if req == nil || req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address required")
	}

	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	binding, found := q.Keeper.GetRoleBinding(sdkCtx, addr)
	if !found {
		return &types.QueryRolesByAddressResponse{}, nil
	}

	return &types.QueryRolesByAddressResponse{Binding: &binding}, nil
}

func (q queryServer) AllRoleBindings(goCtx context.Context, req *types.QueryAllRoleBindingsRequest) (*types.QueryAllRoleBindingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	store := prefix.NewStore(sdkCtx.KVStore(q.storeKey), types.RoleBindingKeyPrefix)

	bindings := []*types.RoleBinding{}
	pageRes, err := query.Paginate(store, req.Pagination, func(_ []byte, value []byte) error {
		var binding types.RoleBinding
		q.cdc.MustUnmarshal(value, &binding)
		bindings = append(bindings, &binding)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryAllRoleBindingsResponse{
		Bindings:   bindings,
		Pagination: pageRes,
	}, nil
}

func (q queryServer) Params(goCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	params := q.GetParams(sdkCtx)
	return &types.QueryParamsResponse{Params: &params}, nil
}

package keeper

import (
	"bytes"
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	Keeper
}

// NewQueryServer returns a new QueryServer implementation backed by the keeper.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return queryServer{Keeper: keeper}
}

func (q queryServer) Asset(goCtx context.Context, req *types.QueryAssetRequest) (*types.QueryAssetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	symbol := strings.TrimSpace(req.Symbol)
	if symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	asset, found := q.GetAsset(ctx, symbol)
	if !found {
		return nil, status.Error(codes.NotFound, "asset not found")
	}

	return &types.QueryAssetResponse{Asset: asset}, nil
}

func (q queryServer) Assets(goCtx context.Context, req *types.QueryAssetsRequest) (*types.QueryAssetsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(q.storeKey)

	var assets []*types.Asset
	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		if !bytes.HasPrefix(key, types.AssetKeyPrefix) {
			return false, nil
		}
		if !accumulate {
			return true, nil
		}

		var asset types.Asset
		if err := q.cdc.Unmarshal(value, &asset); err != nil {
			return false, err
		}
		assetCopy := asset
		assets = append(assets, &assetCopy)
		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAssetsResponse{
		Assets:     assets,
		Pagination: pageRes,
	}, nil
}

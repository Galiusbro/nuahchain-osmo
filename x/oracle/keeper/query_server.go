package keeper

import (
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	Keeper
}

// NewQueryServer creates a new query server instance.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return queryServer{Keeper: keeper}
}

func (q queryServer) Price(goCtx context.Context, req *types.QueryPriceRequest) (*types.QueryPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	symbol := strings.TrimSpace(req.Symbol)
	if symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	price, found := q.GetPrice(ctx, symbol)
	if !found {
		return nil, status.Error(codes.NotFound, "price not found")
	}

	return &types.QueryPriceResponse{Price: price}, nil
}

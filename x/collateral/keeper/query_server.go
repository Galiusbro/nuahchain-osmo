package keeper

import (
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/collateral/types"
)

type queryServer struct {
	Keeper
}

var _ types.QueryServer = queryServer{}

// NewQueryServerImpl returns a new query server instance.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{Keeper: k}
}

// Collateral returns all collateral positions for the specified owner.
func (q queryServer) Collateral(goCtx context.Context, req *types.QueryCollateralRequest) (*types.QueryCollateralResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if strings.TrimSpace(req.Owner) == "" {
		return nil, status.Error(codes.InvalidArgument, "owner cannot be empty")
	}
	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner address")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	positions := q.GetPositions(ctx, owner)
	resp := &types.QueryCollateralResponse{}
	for _, pos := range positions {
		posCopy := pos
		resp.Positions = append(resp.Positions, &posCopy)
	}
	return resp, nil
}

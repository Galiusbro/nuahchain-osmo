package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/freeaccount/types"
)

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// IsFreeAccount implements the Query/IsFreeAccount gRPC method
func (k queryServer) IsFreeAccount(goCtx context.Context, req *types.QueryIsFreeAccountRequest) (*types.QueryIsFreeAccountResponse, error) {
	// Parse address
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	// Check if account is free
	isFree := k.Keeper.IsFreeAccount(goCtx, addr)

	return &types.QueryIsFreeAccountResponse{
		IsFree: isFree,
	}, nil
}
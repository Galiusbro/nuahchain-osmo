package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v30/x/pegkeeper/types"
)

type msgServer struct {
	Keeper
	authority string
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper, authority string) types.MsgServer {
	return &msgServer{Keeper: keeper, authority: authority}
}

var _ types.MsgServer = msgServer{}

// UpdateParams updates the module parameters
func (k msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetParams(ctx, req.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}

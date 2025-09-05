package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v30/x/freeaccount/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// CreateFreeAccount creates a new free account
func (k msgServer) CreateFreeAccount(goCtx context.Context, msg *types.MsgCreateFreeAccount) (*types.MsgCreateFreeAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority
	if k.GetAuthority() != msg.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s, got %s", k.GetAuthority(), msg.Authority)
	}

	// Parse address
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	// Create free account
	err = k.Keeper.CreateFreeAccount(goCtx, addr)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"free_account_created",
			sdk.NewAttribute("address", msg.Address),
			sdk.NewAttribute("authority", msg.Authority),
		),
	)

	return &types.MsgCreateFreeAccountResponse{}, nil
}
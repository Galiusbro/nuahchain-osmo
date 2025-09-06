package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v30/x/limitedaccount/types"
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

// CreateLimitedAccount creates a new limited account
func (k msgServer) CreateLimitedAccount(goCtx context.Context, req *types.MsgCreateLimitedAccount) (*types.MsgCreateLimitedAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if account already exists
	if k.IsLimitedAccount(ctx, req.Address) {
		return nil, types.ErrAccountAlreadyExists
	}

	// Create new limited account
	account := &types.LimitedAccount{
		Address:       req.Address,
		DailyTxCount:  0,
		LastResetTime: ctx.BlockTime(),
		MaxDailyTxs:   3, // Default limit
	}

	k.SetLimitedAccount(ctx, account)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateLimitedAccount,
			sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
		),
	)

	return &types.MsgCreateLimitedAccountResponse{}, nil
}

// RemoveLimitedAccount removes a limited account
func (k msgServer) RemoveLimitedAccount(goCtx context.Context, req *types.MsgRemoveLimitedAccount) (*types.MsgRemoveLimitedAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if account exists
	if !k.IsLimitedAccount(ctx, req.Address) {
		return nil, types.ErrAccountNotFound
	}

	k.Keeper.RemoveLimitedAccount(ctx, req.Address)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRemoveLimitedAccount,
			sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
		),
	)

	return &types.MsgRemoveLimitedAccountResponse{}, nil
}

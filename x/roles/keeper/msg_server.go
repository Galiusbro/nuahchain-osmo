package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/roles/types"
)

var _ types.MsgServer = msgServer{}

// msgServer implements the Msg service for the roles module.
type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return msgServer{Keeper: keeper}
}

func (m msgServer) AssignRoles(goCtx context.Context, msg *types.MsgAssignRoles) (*types.MsgAssignRolesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	if err := m.Keeper.AssignRoles(sdkCtx, authority, addr, msg.Roles); err != nil {
		return nil, err
	}

	return &types.MsgAssignRolesResponse{}, nil
}

func (m msgServer) RevokeRoles(goCtx context.Context, msg *types.MsgRevokeRoles) (*types.MsgRevokeRolesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	if err := m.Keeper.RevokeRoles(sdkCtx, authority, addr, msg.Roles); err != nil {
		return nil, err
	}

	return &types.MsgRevokeRolesResponse{}, nil
}

func (m msgServer) UpdateAuthority(goCtx context.Context, msg *types.MsgUpdateAuthority) (*types.MsgUpdateAuthorityResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	if err := m.Keeper.UpdateAuthority(sdkCtx, authority, msg.NewAuthority); err != nil {
		return nil, err
	}

	return &types.MsgUpdateAuthorityResponse{}, nil
}

package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/policy/types"
)

var _ types.MsgServer = msgServer{}

// msgServer implements the Msg service for the policy module.
type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns a new Msg server instance.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return msgServer{Keeper: keeper}
}

func (m msgServer) CreatePolicy(goCtx context.Context, msg *types.MsgCreatePolicy) (*types.MsgCreatePolicyResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	var start *time.Time
	if msg.StartTime != nil {
		start = msg.StartTime
	}

	var end *time.Time
	if msg.EndTime != nil {
		end = msg.EndTime
	}

	policy, err := m.Keeper.CreatePolicy(sdkCtx, owner, msg.PolicyType, msg.Attributes, start, end, msg.TreasuryPoolId, msg.Tags)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreatePolicyResponse{PolicyId: policy.Id}, nil
}

func (m msgServer) UpdatePolicyAttributes(goCtx context.Context, msg *types.MsgUpdatePolicyAttributes) (*types.MsgUpdatePolicyAttributesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	_, err = m.Keeper.UpdatePolicyAttributes(sdkCtx, authority, msg.PolicyId, msg.Attributes, msg.Replace)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdatePolicyAttributesResponse{}, nil
}

func (m msgServer) CancelPolicy(goCtx context.Context, msg *types.MsgCancelPolicy) (*types.MsgCancelPolicyResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	_, err = m.Keeper.CancelPolicy(sdkCtx, authority, msg.PolicyId, msg.Reason)
	if err != nil {
		return nil, err
	}

	return &types.MsgCancelPolicyResponse{}, nil
}

func (m msgServer) UpdatePolicyStatus(goCtx context.Context, msg *types.MsgUpdatePolicyStatus) (*types.MsgUpdatePolicyStatusResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	_, err = m.Keeper.UpdatePolicyStatus(sdkCtx, authority, msg.PolicyId, msg.Status)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdatePolicyStatusResponse{}, nil
}

package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/risk/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns a new MsgServer implementation.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

// SetRiskParams handles MsgSetRiskParams requests.
func (m msgServer) SetRiskParams(goCtx context.Context, msg *types.MsgSetRiskParams) (*types.MsgSetRiskParamsResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if msg.Authority != m.authority {
		return nil, fmt.Errorf("unauthorized: expected %s", m.authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err := m.Keeper.SetRiskParams(ctx, msg.Params)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetRiskParamsResponse{}, nil
}

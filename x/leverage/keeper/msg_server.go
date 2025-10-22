package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/leverage/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

func (m msgServer) OpenPosition(goCtx context.Context, msg *types.MsgOpenPosition) (*types.MsgOpenPositionResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	position, err := m.Keeper.OpenPosition(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgOpenPositionResponse{Position: position}, nil
}

func (m msgServer) ClosePosition(goCtx context.Context, msg *types.MsgClosePosition) (*types.MsgClosePositionResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	pnl, err := m.Keeper.ClosePosition(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgClosePositionResponse{Pnl: pnl}, nil
}

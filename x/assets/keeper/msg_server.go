package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServer returns a new Msg service implementation.
func NewMsgServer(keeper Keeper) types.MsgServer {
	return msgServer{Keeper: keeper}
}

func (m msgServer) EnsureAsset(goCtx context.Context, msg *types.MsgEnsureAsset) (*types.MsgEnsureAssetResponse, error) {
	if msg == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("message cannot be nil")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	asset, created, err := m.Keeper.EnsureAsset(ctx, msg.Symbol)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	if created {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeAssetCreated,
				sdk.NewAttribute(types.AttributeKeySymbol, asset.Symbol),
				sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			),
		)
	}

	return &types.MsgEnsureAssetResponse{Asset: asset}, nil
}

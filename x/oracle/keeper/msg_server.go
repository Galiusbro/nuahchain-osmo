package keeper

import (
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServer returns a new Msg service implementation.
func NewMsgServer(keeper Keeper) types.MsgServer {
	return msgServer{Keeper: keeper}
}

func (m msgServer) SetPrice(goCtx context.Context, msg *types.MsgSetPrice) (*types.MsgSetPriceResponse, error) {
	if msg == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("message cannot be nil")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if msg.Authority != m.authority {
		return nil, sdkerrors.ErrUnauthorized.Wrap("only authority can set prices")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	symbol := EnsureSymbol(msg.Symbol)
	if symbol == "" {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("symbol cannot be empty")
	}

	price := &types.Price{
		Symbol: symbol,
		Value:  strings.TrimSpace(msg.Value),
	}

	m.Keeper.SetPrice(ctx, price)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePriceUpdated,
			sdk.NewAttribute(types.AttributeKeySymbol, price.Symbol),
			sdk.NewAttribute(types.AttributeKeyValue, price.Value),
			sdk.NewAttribute(types.AttributeKeyAuthority, msg.Authority),
		),
	)

	return &types.MsgSetPriceResponse{}, nil
}

package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v30/x/stablecoin/types"
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

// BuyNDollar converts unuah to NDOLLAR at 1:1 ratio
func (m msgServer) BuyNDollar(goCtx context.Context, msg *types.MsgBuyNDollar) (*types.MsgBuyNDollarResponse, error) {
	if msg == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("message cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate buyer address
	buyer, err := sdk.AccAddressFromBech32(msg.Buyer)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid buyer: %v", err)
	}

	// Parse amount
	amount, ok := sdkmath.NewIntFromString(msg.Amount)
	if !ok || !amount.IsPositive() {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("amount must be a positive integer")
	}

	// Execute the conversion
	ndollarDenom, ndollarAmount, err := m.Keeper.BuyNDollar(ctx, buyer, amount)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"buy_ndollar",
			sdk.NewAttribute("buyer", msg.Buyer),
			sdk.NewAttribute("unuah_amount", amount.String()),
			sdk.NewAttribute("ndollar_amount", ndollarAmount.String()),
			sdk.NewAttribute("ndollar_denom", ndollarDenom),
		),
	)

	return &types.MsgBuyNDollarResponse{
		NdollarAmount: ndollarAmount.String(),
		NdollarDenom:  ndollarDenom,
	}, nil
}

// SellNDollar converts NDOLLAR back to unuah at 1:1 ratio
func (m msgServer) SellNDollar(goCtx context.Context, msg *types.MsgSellNDollar) (*types.MsgSellNDollarResponse, error) {
	if msg == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("message cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate seller address
	seller, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid seller: %v", err)
	}

	// Parse amount
	amount, ok := sdkmath.NewIntFromString(msg.Amount)
	if !ok || !amount.IsPositive() {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("amount must be a positive integer")
	}

	// Execute the conversion
	unuahAmount, err := m.Keeper.SellNDollar(ctx, seller, amount)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"sell_ndollar",
			sdk.NewAttribute("seller", msg.Seller),
			sdk.NewAttribute("ndollar_amount", amount.String()),
			sdk.NewAttribute("unuah_amount", unuahAmount.String()),
		),
	)

	return &types.MsgSellNDollarResponse{
		UnuahAmount: unuahAmount.String(),
	}, nil
}

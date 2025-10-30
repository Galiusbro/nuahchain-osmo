package keeper

import (
	"context"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	osmomath "github.com/osmosis-labs/osmosis/osmomath"

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

func (m msgServer) BuyAsset(goCtx context.Context, msg *types.MsgBuyAsset) (*types.MsgBuyAssetResponse, error) {
	if msg == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("message cannot be nil")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	buyer, err := sdk.AccAddressFromBech32(msg.Buyer)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid buyer: %v", err)
	}

	var payment sdk.Coin
	var amountNDStr string

	// Support both old (amount_NDOLLAR) and new (denom + amount) format
	if msg.Denom != "" && msg.Amount != "" {
		// New format: denom + amount
		amount, ok := sdkmath.NewIntFromString(msg.Amount)
		if !ok {
			return nil, sdkerrors.ErrInvalidRequest.Wrap("amount must be an integer")
		}
		payment = sdk.NewCoin(msg.Denom, amount)
		amountNDStr = msg.Amount
	} else if msg.Amount_NDOLLAR != "" {
		// Old format: amount_NDOLLAR (deprecated but still supported)
		amountND, ok := sdkmath.NewIntFromString(msg.Amount_NDOLLAR)
		if !ok {
			return nil, sdkerrors.ErrInvalidRequest.Wrap("amount_ndollar must be an integer")
		}
		payment = sdk.NewCoin(types.NDollarDenom, amountND)
		amountNDStr = msg.Amount_NDOLLAR
	} else {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("either amount_NDOLLAR (deprecated) or denom+amount must be provided")
	}

	var baseAmountDec osmomath.Dec
	// Check if payment is NDOLLAR (either "NDOLLAR" or factory/*/ndollar format)
	isNDollar := payment.Denom == types.NDollarDenom || (strings.HasPrefix(payment.Denom, "factory/") && strings.HasSuffix(payment.Denom, "/ndollar"))

	if isNDollar {
		// Use old method for backward compatibility when denom is exactly "NDOLLAR"
		if payment.Denom == types.NDollarDenom {
			_, baseAmountDec, err = m.Keeper.BuyAsset(ctx, buyer, msg.Symbol, payment.Amount)
		} else {
			// Use new method for factory/*/ndollar format
			_, baseAmountDec, err = m.Keeper.BuyAssetWithPayment(ctx, buyer, msg.Symbol, payment)
		}
		if err != nil {
			return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
		}
	} else {
		// Use new method for unuah or other denoms
		_, baseAmountDec, err = m.Keeper.BuyAssetWithPayment(ctx, buyer, msg.Symbol, payment)
		if err != nil {
			return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAssetBought,
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
			sdk.NewAttribute(types.AttributeKeyBuyer, msg.Buyer),
			sdk.NewAttribute(types.AttributeKeyAmountNDOLLAR, amountNDStr),
			sdk.NewAttribute(types.AttributeKeyBaseAmount, baseAmountDec.String()),
		),
	)

	return &types.MsgBuyAssetResponse{BaseAmount: baseAmountDec.String()}, nil
}

func (m msgServer) SellAsset(goCtx context.Context, msg *types.MsgSellAsset) (*types.MsgSellAssetResponse, error) {
	if msg == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("message cannot be nil")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	baseAmountDec, err := osmomath.NewDecFromStr(msg.BaseAmount)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid base amount: %v", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	seller, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid seller: %v", err)
	}

	ndCoin, payoutInt, err := m.Keeper.SellAsset(ctx, seller, msg.Symbol, baseAmountDec)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAssetSold,
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
			sdk.NewAttribute(types.AttributeKeySeller, msg.Seller),
			sdk.NewAttribute(types.AttributeKeyBaseAmount, msg.BaseAmount),
			sdk.NewAttribute(types.AttributeKeyPayoutNDOLLAR, payoutInt.String()),
		),
	)

	return &types.MsgSellAssetResponse{Payout_NDOLLAR: ndCoin.Amount.String()}, nil
}

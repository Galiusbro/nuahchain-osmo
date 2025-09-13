package keeper

import (
	"context"
	"time"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v30/x/usdoracle/types"
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

// UpdateParams updates the module parameters
func (k msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetParams(ctx, req.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}

// UpdateUSDPrice updates the USD price
func (k msgServer) UpdateUSDPrice(goCtx context.Context, req *types.MsgUpdateUSDPrice) (*types.MsgUpdateUSDPriceResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Create USD price entry
	usdPrice := types.USDPrice{
		Price:     req.Price,
		Source:    req.Source,
		Timestamp: time.Now(),
	}

	// Update current price
	k.SetCurrentPrice(ctx, usdPrice)

	// Add to price history
	k.AddPriceHistory(ctx, usdPrice)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUSDPriceUpdate,
			sdk.NewAttribute(types.AttributeKeyPrice, req.Price.String()),
			sdk.NewAttribute(types.AttributeKeySource, req.Source),
			sdk.NewAttribute(types.AttributeKeyTimestamp, usdPrice.Timestamp.String()),
		),
	)

	return &types.MsgUpdateUSDPriceResponse{}, nil
}

// UpdateTokenPrice updates the USD price for a specific token
func (k msgServer) UpdateTokenPrice(goCtx context.Context, req *types.MsgUpdateTokenPrice) (*types.MsgUpdateTokenPriceResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Update token price logic here
	_ = ctx

	return &types.MsgUpdateTokenPriceResponse{}, nil
}

// AddSupportedToken adds a new supported token
func (k msgServer) AddSupportedToken(goCtx context.Context, req *types.MsgAddSupportedToken) (*types.MsgAddSupportedTokenResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Add supported token logic here
	_ = ctx

	return &types.MsgAddSupportedTokenResponse{}, nil
}

// RemoveSupportedToken removes a supported token
func (k msgServer) RemoveSupportedToken(goCtx context.Context, req *types.MsgRemoveSupportedToken) (*types.MsgRemoveSupportedTokenResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Remove supported token logic here
	_ = ctx

	return &types.MsgRemoveSupportedTokenResponse{}, nil
}

// UpdateSupportedToken updates configuration for a supported token
func (k msgServer) UpdateSupportedToken(goCtx context.Context, req *types.MsgUpdateSupportedToken) (*types.MsgUpdateSupportedTokenResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Update supported token logic here
	_ = ctx

	return &types.MsgUpdateSupportedTokenResponse{}, nil
}

// SetPriceSources sets the price sources configuration
func (k msgServer) SetPriceSources(goCtx context.Context, req *types.MsgSetPriceSources) (*types.MsgSetPriceSourcesResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Set price sources
	for _, source := range req.Sources {
		k.SetPriceSource(ctx, source)
	}

	return &types.MsgSetPriceSourcesResponse{}, nil
}

// UpdatePriceSource updates a specific price source configuration
func (k msgServer) UpdatePriceSource(goCtx context.Context, req *types.MsgUpdatePriceSource) (*types.MsgUpdatePriceSourceResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Update price source logic here
	_ = ctx

	return &types.MsgUpdatePriceSourceResponse{}, nil
}
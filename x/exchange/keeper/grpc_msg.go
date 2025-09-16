package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/exchange/types"
)

// grpcMsgServer wraps the msgServer to implement the gRPC MsgServer interface
type grpcMsgServer struct {
	msgServer
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &grpcMsgServer{
		msgServer: msgServer{Keeper: keeper},
	}
}

var _ types.MsgServer = grpcMsgServer{}

// ExchangeTokens handles token exchange transactions
func (m grpcMsgServer) ExchangeTokens(goCtx context.Context, msg *types.MsgExchangeTokens) (*types.MsgExchangeTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return m.msgServer.ExchangeTokens(ctx, msg)
}

// UpdateParams handles parameter update transactions
func (m grpcMsgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return m.msgServer.UpdateParams(ctx, msg)
}

// AddSupportedToken handles adding a new supported token to the registry
func (m grpcMsgServer) AddSupportedToken(goCtx context.Context, msg *types.MsgAddSupportedToken) (*types.MsgAddSupportedTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return m.msgServer.AddSupportedToken(ctx, msg)
}

// RemoveSupportedToken handles removing a supported token from the registry
func (m grpcMsgServer) RemoveSupportedToken(goCtx context.Context, msg *types.MsgRemoveSupportedToken) (*types.MsgRemoveSupportedTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return m.msgServer.RemoveSupportedToken(ctx, msg)
}
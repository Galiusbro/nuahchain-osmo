package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/treasury/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns a new Msg server.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return msgServer{Keeper: keeper}
}

func (m msgServer) CreateTreasuryPool(goCtx context.Context, msg *types.MsgCreateTreasuryPool) (*types.MsgCreateTreasuryPoolResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	pool := types.TreasuryPool{
		Id:          msg.PoolId,
		Description: msg.Description,
		Manager:     msg.Manager,
		PolicyTypes: append([]string(nil), msg.PolicyTypes...),
	}

	if err := m.Keeper.CreateTreasuryPool(sdkCtx, authority, pool); err != nil {
		return nil, err
	}

	return &types.MsgCreateTreasuryPoolResponse{}, nil
}

func (m msgServer) UpdateTreasuryPool(goCtx context.Context, msg *types.MsgUpdateTreasuryPool) (*types.MsgUpdateTreasuryPoolResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	pool := types.TreasuryPool{
		Id:          msg.PoolId,
		Description: msg.Description,
		Manager:     msg.Manager,
		PolicyTypes: append([]string(nil), msg.PolicyTypes...),
	}

	if err := m.Keeper.UpdateTreasuryPool(sdkCtx, authority, pool); err != nil {
		return nil, err
	}

	return &types.MsgUpdateTreasuryPoolResponse{}, nil
}

func (m msgServer) DepositToTreasury(goCtx context.Context, msg *types.MsgDepositToTreasury) (*types.MsgDepositToTreasuryResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, err
	}

	amount := sdk.NewCoin(msg.Amount.Denom, msg.Amount.Amount)

	if err := m.Keeper.DepositToTreasury(sdkCtx, depositor, msg.PoolId, amount); err != nil {
		return nil, err
	}

	return &types.MsgDepositToTreasuryResponse{}, nil
}

func (m msgServer) WithdrawFromTreasury(goCtx context.Context, msg *types.MsgWithdrawFromTreasury) (*types.MsgWithdrawFromTreasuryResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	recipient, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		return nil, err
	}

	amount := sdk.NewCoin(msg.Amount.Denom, msg.Amount.Amount)

	if err := m.Keeper.WithdrawFromTreasury(sdkCtx, authority, msg.PoolId, recipient, amount); err != nil {
		return nil, err
	}

	return &types.MsgWithdrawFromTreasuryResponse{}, nil
}

func (m msgServer) SetPoolReserves(goCtx context.Context, msg *types.MsgSetPoolReserves) (*types.MsgSetPoolReservesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	if err := m.Keeper.SetPoolReserves(sdkCtx, authority, msg.PoolId, msg.Reserves); err != nil {
		return nil, err
	}

	return &types.MsgSetPoolReservesResponse{}, nil
}

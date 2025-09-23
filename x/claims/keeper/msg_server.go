package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/claims/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return msgServer{Keeper: keeper}
}

func (m msgServer) SubmitClaim(goCtx context.Context, msg *types.MsgSubmitClaim) (*types.MsgSubmitClaimResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	claimant, err := sdk.AccAddressFromBech32(msg.Claimant)
	if err != nil {
		return nil, err
	}

	amount := sdk.NewCoin(msg.Amount.Denom, msg.Amount.Amount)

	claim, err := m.Keeper.SubmitClaim(sdkCtx, claimant, msg.PolicyId, amount, msg.Description, msg.Evidence)
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitClaimResponse{ClaimId: claim.Id}, nil
}

func (m msgServer) ReviewClaim(goCtx context.Context, msg *types.MsgReviewClaim) (*types.MsgReviewClaimResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	claim, err := m.Keeper.ReviewClaim(sdkCtx, authority, msg.ClaimId, msg.Decision, msg.Reason)
	if err != nil {
		return nil, err
	}

	return &types.MsgReviewClaimResponse{Claim: &claim}, nil
}

func (m msgServer) AddClaimEvidence(goCtx context.Context, msg *types.MsgAddClaimEvidence) (*types.MsgAddClaimEvidenceResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	claim, err := m.Keeper.AddClaimEvidence(sdkCtx, authority, msg.ClaimId, msg.Evidence)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddClaimEvidenceResponse{Claim: &claim}, nil
}

func (m msgServer) ExecuteClaimPayout(goCtx context.Context, msg *types.MsgExecuteClaimPayout) (*types.MsgExecuteClaimPayoutResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	recipient, err := sdk.AccAddressFromBech32(msg.PayoutAddress)
	if err != nil {
		return nil, err
	}

	claim, err := m.Keeper.ExecuteClaimPayout(sdkCtx, authority, msg.ClaimId, recipient)
	if err != nil {
		return nil, err
	}

	return &types.MsgExecuteClaimPayoutResponse{Claim: &claim}, nil
}

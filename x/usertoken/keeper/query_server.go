package keeper

import (
	"context"
	"fmt"

	cosmossdk_io_math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

func (q queryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := q.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (q queryServer) UserToken(goCtx context.Context, req *types.QueryUserTokenRequest) (*types.QueryUserTokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "denom cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get user token info from store
	userToken, found := q.GetUserToken(ctx, req.Denom)
	if !found {
		return nil, status.Error(codes.NotFound, "user token not found")
	}

	return &types.QueryUserTokenResponse{UserToken: &userToken}, nil
}

func (q queryServer) UserTokenMetadata(goCtx context.Context, req *types.QueryUserTokenMetadataRequest) (*types.QueryUserTokenMetadataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "denom cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get user token info from store
	userToken, found := q.GetUserToken(ctx, req.Denom)
	if !found {
		return nil, status.Error(codes.NotFound, "user token not found")
	}

	return &types.QueryUserTokenMetadataResponse{
		Name:          userToken.Name,
		Symbol:        userToken.Symbol,
		Description:   userToken.Description,
		Creator:       userToken.Creator,
		MaxSupply:     userToken.MaxSupply,
		CurrentSupply: userToken.CurrentSupply,
	}, nil
}

func (q queryServer) UserTokens(goCtx context.Context, req *types.QueryUserTokensRequest) (*types.QueryUserTokensResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get all user tokens from store
	userTokens := q.GetAllUserTokens(ctx)

	return &types.QueryUserTokensResponse{
		UserTokens: userTokens,
		Pagination: nil, // TODO: implement pagination if needed
	}, nil
}

func (q queryServer) BondingCurvePrice(goCtx context.Context, req *types.QueryBondingCurvePriceRequest) (*types.QueryBondingCurvePriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "denom cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get bonding curve supply using the correct calculation
	bondingCurveSupply, err := q.GetBondingCurveSupply(ctx, req.Denom)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	price := q.CalculateBondingCurvePrice(ctx, req.Denom, bondingCurveSupply)

	return &types.QueryBondingCurvePriceResponse{
		Price: price,
	}, nil
}

func (q queryServer) EstimateTokens(goCtx context.Context, req *types.QueryEstimateTokensRequest) (*types.QueryEstimateTokensResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate denom
	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "denom cannot be empty")
	}

	// Parse payment amount from string
	paymentAmount, ok := cosmossdk_io_math.NewIntFromString(req.PaymentAmount)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid payment amount format")
	}

	// Validate payment amount
	if paymentAmount.IsZero() || paymentAmount.IsNegative() {
		return nil, status.Error(codes.InvalidArgument, "payment amount must be positive")
	}

	// Check if token exists
	_, found := q.GetUserToken(ctx, req.Denom)
	if !found {
		return nil, status.Error(codes.NotFound, "token not found")
	}

	// Get current bonding curve supply
	bondingCurveSupply, err := q.GetBondingCurveSupply(ctx, req.Denom)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get bonding curve supply: %v", err))
	}

	// Calculate tokens using the same method as BuyTokens
	tokensReceived := q.CalculateTokensFromPayment(ctx, req.Denom, bondingCurveSupply, paymentAmount)

	return &types.QueryEstimateTokensResponse{TokensReceived: tokensReceived}, nil
}

func (q queryServer) ReferralProgram(goCtx context.Context, req *types.QueryReferralProgramRequest) (*types.QueryReferralProgramResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.TokenDenom == "" {
		return nil, status.Error(codes.InvalidArgument, "token denom cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get referral program from store
	referralProgram, found := q.GetReferralProgram(ctx, req.TokenDenom)
	if !found {
		return nil, status.Error(codes.NotFound, "referral program not found")
	}

	return &types.QueryReferralProgramResponse{
		ReferralProgram: &referralProgram,
	}, nil
}

func (q queryServer) ReferralPrograms(goCtx context.Context, req *types.QueryReferralProgramsRequest) (*types.QueryReferralProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get all referral programs from store
	referralPrograms := q.GetAllReferralPrograms(ctx)

	return &types.QueryReferralProgramsResponse{
		ReferralPrograms: referralPrograms,
	}, nil
}

func (q queryServer) UserTokensByCreator(goCtx context.Context, req *types.QueryUserTokensByCreatorRequest) (*types.QueryUserTokensByCreatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Creator == "" {
		return nil, status.Error(codes.InvalidArgument, "creator cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get all user tokens from store
	allUserTokens := q.GetAllUserTokens(ctx)

	// Filter by creator
	var userTokens []*types.UserToken
	for _, token := range allUserTokens {
		if token.Creator == req.Creator {
			userTokens = append(userTokens, token)
		}
	}

	return &types.QueryUserTokensByCreatorResponse{
		UserTokens: userTokens,
		Pagination: nil, // TODO: implement pagination if needed
	}, nil
}

func (q queryServer) ReferralActivations(goCtx context.Context, req *types.QueryReferralActivationsRequest) (*types.QueryReferralActivationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.User == "" {
		return nil, status.Error(codes.InvalidArgument, "user cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get all referral activations from store
	allActivations := q.GetAllReferralActivations(ctx)

	// Filter by user if specified
	var referralActivations []*types.ReferralActivation
	for _, activation := range allActivations {
		if activation.Referee == req.User {
			referralActivations = append(referralActivations, activation)
		}
	}

	return &types.QueryReferralActivationsResponse{
		ReferralActivations: referralActivations,
	}, nil
}

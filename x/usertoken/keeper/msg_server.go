package keeper

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/osmosis-labs/osmosis/osmomath"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

var (
	allowedSubdenomRegexp = regexp.MustCompile(`[^a-z0-9._/-]`)
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the MsgServer interface for the provided keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (s msgServer) CreateToken(goCtx context.Context, msg *types.MsgCreateToken) (*types.MsgCreateTokenResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	params := s.GetParams(ctx)

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	if s.isNameTaken(ctx, msg.Name) {
		return nil, types.ErrNameExists
	}

	if s.isSymbolTaken(ctx, msg.Symbol) {
		return nil, types.ErrSymbolExists
	}

	subdenom, err := sanitizeSubdenom(msg.Symbol)
	if err != nil {
		return nil, err
	}

	denom, err := s.tokenFactoryKeeper.CreateDenom(ctx, msg.Creator, subdenom)
	if err != nil {
		return nil, types.ErrTokenFactory.Wrap(err.Error())
	}

	// Set denom metadata to include human-readable unit with 6 decimals for better UX
	// Base is the factory denom (exponent 0), display uses the token symbol (lowercased) with exponent 6
	displayUnit := strings.ToLower(strings.TrimSpace(msg.Symbol))
	if displayUnit == "" {
		displayUnit = denom
	}
	metadata := banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: denom, Exponent: 0},
			{Denom: displayUnit, Exponent: 6},
		},
		Base:    denom,
		Display: displayUnit,
		Name:    strings.TrimSpace(msg.Name),
		Symbol:  strings.ToUpper(strings.TrimSpace(msg.Symbol)),
	}
	s.bankKeeper.SetDenomMetaData(ctx, metadata)

	if err := s.chargeCreationFee(ctx, creatorAddr, params.TokenCreationFee); err != nil {
		return nil, err
	}

	if err := ensureDistributionAddresses(params); err != nil {
		return nil, err
	}

	if err := s.mintDistributions(ctx, msg.Creator, denom, params); err != nil {
		return nil, err
	}

	token := types.Token{
		Creator:     msg.Creator,
		Denom:       denom,
		Name:        strings.TrimSpace(msg.Name),
		Symbol:      strings.TrimSpace(msg.Symbol),
		Image:       strings.TrimSpace(msg.Image),
		Description: msg.Description,
		CreatedAt:   uint64(ctx.BlockTime().Unix()),
		Distribution: types.TokenDistribution{
			FounderClaimed: false,
		},
	}

	token.Distribution.SetTotalSupply(osmomath.NewInt(100_000_000))
	token.Distribution.SetBondingCurveSupply(osmomath.NewInt(30_000_000))
	token.Distribution.SetPlatformAllocation(osmomath.NewInt(10_000_000))
	token.Distribution.SetReferralAllocation(osmomath.NewInt(10_000_000))
	token.Distribution.SetAiCeoAllocation(osmomath.NewInt(40_000_000))
	token.Distribution.SetFounderReserved(osmomath.NewInt(10_000_000))

	state := types.TokenState{}
	state.BondingCurveSold = "0"
	state.CurrentPrice = "0"
	state.CurveCompleted = false
	state.DexTradingEnabled = false
	state.SoftLockEnabled = true
	token.State = state

	deadline := ctx.BlockTime().Add(time.Duration(params.FounderClaimPeriod) * time.Second)
	token.Distribution.FounderClaimDeadline = uint64(deadline.Unix())

	s.setToken(ctx, token)
	s.queueFounderDeadline(ctx, token.Distribution.FounderClaimDeadline, denom)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCreateToken,
		sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
		sdk.NewAttribute(types.AttributeKeyDenom, denom),
		sdk.NewAttribute(types.AttributeKeyName, token.Name),
		sdk.NewAttribute(types.AttributeKeySymbol, token.Symbol),
	))

	return &types.MsgCreateTokenResponse{Denom: denom}, nil
}

func (s msgServer) FounderClaim(goCtx context.Context, msg *types.MsgFounderClaim) (*types.MsgFounderClaimResponse, error) {
	return nil, types.ErrFounderClaimExpired
}

func (s msgServer) chargeCreationFee(ctx sdk.Context, creator sdk.AccAddress, fee sdk.Coin) error {
	if fee.IsNil() || fee.IsZero() {
		return nil
	}

	coins := sdk.NewCoins(fee)
	if err := s.bankKeeper.SendCoinsFromAccountToModule(sdk.WrapSDKContext(ctx), creator, types.ModuleName, coins); err != nil {
		return err
	}
	return nil
}

func (s msgServer) mintDistributions(ctx sdk.Context, creator string, denom string, params types.Params) error {
	targets := []struct {
		address string
		amount  osmomath.Int
	}{
		{address: params.BondingCurveWallet, amount: osmomath.NewInt(30_000_000)},
		{address: params.PlatformWallet, amount: osmomath.NewInt(10_000_000)},
		{address: params.ReferralWallet, amount: osmomath.NewInt(10_000_000)},
		{address: params.AiCeoWallet, amount: osmomath.NewInt(40_000_000)},
	}

	for _, target := range targets {
		if err := s.mintToAddress(ctx, creator, denom, target.address, target.amount); err != nil {
			return err
		}
	}

	return nil
}

func ensureDistributionAddresses(params types.Params) error {
	if params.BondingCurveWallet == "" || params.PlatformWallet == "" || params.ReferralWallet == "" || params.AiCeoWallet == "" {
		return types.ErrParamAddresses
	}
	if _, err := sdk.AccAddressFromBech32(params.BondingCurveWallet); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(params.PlatformWallet); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(params.ReferralWallet); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(params.AiCeoWallet); err != nil {
		return err
	}
	return nil
}

func sanitizeSubdenom(symbol string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(symbol))
	if trimmed == "" {
		return "", types.ErrInvalidSymbol
	}

	sanitized := allowedSubdenomRegexp.ReplaceAllString(trimmed, "")
	if sanitized == "" {
		return "", types.ErrInvalidSymbol
	}

	if len(sanitized) > 44 {
		sanitized = sanitized[:44]
	}

	return sanitized, nil
}

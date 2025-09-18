package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/errors"
	cosmossdk_io_math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	balancertypes "github.com/osmosis-labs/osmosis/v30/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

type msgServer struct {
	Keeper
	authority string
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper, authority string) types.MsgServer {
	return &msgServer{Keeper: keeper, authority: authority}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) CreateUserToken(goCtx context.Context, msg *types.MsgCreateUserToken) (*types.MsgCreateUserTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Create the token via tokenfactory
	denom, err := k.tokenfactoryKeeper.CreateDenom(ctx, msg.Creator, msg.Subdenom)
	if err != nil {
		return nil, err
	}

	// Transfer admin rights to usertoken module
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if moduleAddr == nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "module address not found")
	}

	// Note: ChangeAdmin method needs to be implemented in tokenfactory keeper
	// For now, we'll skip this step and implement it later
	_ = moduleAddr // suppress unused variable warning

	// Mint 100M tokens total
	totalSupply := cosmossdk_io_math.NewInt(100_000_000)
	tokensToMint := sdk.NewCoin(denom, totalSupply)
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(tokensToMint))
	if err != nil {
		return nil, fmt.Errorf("failed to mint tokens: %w", err)
	}

	// Distribute tokens according to tokenomics:
	// 30M to bonding curve (stays in module)
	// 10M to platform
	// 10M to referral wallet
	// 40M to AI CEO
	// 10M reserved for founder offer
	err = k.distributeInitialTokens(ctx, denom, totalSupply)
	if err != nil {
		return nil, fmt.Errorf("failed to distribute tokens: %w", err)
	}

	// Create and store user token metadata
	params := k.GetParams(ctx)
	maxSupply := params.BondingCurveMaxSupply

	userToken := types.UserToken{
		Denom:                denom,
		Creator:              msg.Creator,
		Name:                 msg.Name,
		Symbol:               msg.Symbol,
		MaxSupply:            maxSupply,
		CurrentSupply:        totalSupply,                                  // 100M tokens minted and distributed
		ReserveRatio:         cosmossdk_io_math.LegacyNewDecWithPrec(5, 1), // 0.5
		InitialPrice:         params.BondingCurveStartPrice,
		FounderPercentage:    cosmossdk_io_math.LegacyNewDecWithPrec(1, 1), // 0.1 (10%)
		FounderTokensClaimed: cosmossdk_io_math.NewInt(0),
		LbpActive:            false,
		LbpStartTime:         0,
	}

	k.SetUserToken(ctx, userToken)

	// Automatically create N$/UserToken pool
	err = k.createNuahUserTokenPool(ctx, denom, msg.Creator)
	if err != nil {
		// Log error but don't fail token creation
		ctx.Logger().Error("Failed to create N$/UserToken pool", "denom", denom, "error", err)
	}

	// Emit optimized event with pre-computed decimals string
	decimalsStr := fmt.Sprintf("%d", msg.Decimals)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"create_user_token",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("denom", denom),
			sdk.NewAttribute("subdenom", msg.Subdenom),
			sdk.NewAttribute("name", msg.Name),
			sdk.NewAttribute("symbol", msg.Symbol),
			sdk.NewAttribute("decimals", decimalsStr),
		),
	})

	return &types.MsgCreateUserTokenResponse{
		Denom: denom,
	}, nil
}

func (k msgServer) BuyTokens(goCtx context.Context, msg *types.MsgBuyTokens) (*types.MsgBuyTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate input - msg.Amount is the payment amount in N$
	if msg.Amount.Amount.IsZero() || msg.Amount.Amount.IsNegative() {
		return nil, fmt.Errorf("invalid payment amount: %s", msg.Amount.Amount.String())
	}

	// Get user token to verify it exists
	userToken, found := k.GetUserToken(ctx, msg.Denom)
	if !found {
		return nil, types.ErrTokenNotFound
	}

	// Get bonding curve supply (excludes founder tokens)
	bondingCurveSupply, err := k.GetBondingCurveSupply(ctx, msg.Denom)
	if err != nil {
		// Fallback to stored supply minus founder tokens
		bondingCurveSupply = userToken.CurrentSupply.Sub(userToken.FounderTokensClaimed)
		if bondingCurveSupply.IsNegative() {
			bondingCurveSupply = cosmossdk_io_math.ZeroInt()
		}
	}

	// Calculate how many tokens can be bought with the payment amount (using 30% for bonding curve)
	tokensReceived := k.CalculateTokensFromPayment(ctx, bondingCurveSupply, msg.Amount.Amount)

	// Check if tokens received meets minimum requirement
	if tokensReceived.LT(msg.MinTokens) {
		return nil, fmt.Errorf("tokens received %s is less than minimum %s", tokensReceived.String(), msg.MinTokens.String())
	}

	// Transfer payment from buyer to module
	buyerAddr, err := sdk.AccAddressFromBech32(msg.Buyer)
	if err != nil {
		return nil, fmt.Errorf("invalid buyer address: %w", err)
	}

	// Transfer payment from buyer to module first
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if moduleAddr == nil {
		return nil, fmt.Errorf("module address not found")
	}

	err = k.bankKeeper.SendCoins(ctx, buyerAddr, moduleAddr, sdk.NewCoins(msg.Amount))
	if err != nil {
		return nil, fmt.Errorf("failed to transfer payment: %w", err)
	}

	// Distribute payment according to tokenomics
	err = k.DistributeTokenPurchasePayment(ctx, msg.Denom, msg.Amount.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to distribute payment: %w", err)
	}

	// Mint tokens to buyer via tokenfactory
	tokensToMint := sdk.NewCoin(msg.Denom, tokensReceived)
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(tokensToMint))
	if err != nil {
		return nil, fmt.Errorf("failed to mint tokens: %w", err)
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, buyerAddr, sdk.NewCoins(tokensToMint))
	if err != nil {
		return nil, fmt.Errorf("failed to send minted tokens: %w", err)
	}

	// Update supply tracking in store
	userToken.CurrentSupply = userToken.CurrentSupply.Add(tokensReceived)
	k.SetUserToken(ctx, userToken)

	// Emit optimized event with pre-computed strings
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"buy_tokens",
			sdk.NewAttribute("buyer", msg.Buyer),
			sdk.NewAttribute("denom", msg.Denom),
			sdk.NewAttribute("payment_amount", msg.Amount.Amount.String()),
			sdk.NewAttribute("tokens_received", tokensReceived.String()),
			sdk.NewAttribute("new_supply", userToken.CurrentSupply.String()),
		),
	})

	return &types.MsgBuyTokensResponse{
		TokensReceived: tokensReceived,
	}, nil
}

func (k msgServer) SellTokens(goCtx context.Context, msg *types.MsgSellTokens) (*types.MsgSellTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate input
	if msg.Amount.Amount.IsZero() || msg.Amount.Amount.IsNegative() {
		return nil, fmt.Errorf("invalid token amount: %s", msg.Amount.Amount.String())
	}

	// Get user token to verify it exists
	userToken, found := k.GetUserToken(ctx, msg.Amount.Denom)
	if !found {
		return nil, types.ErrTokenNotFound
	}

	// Get bonding curve supply (excludes founder tokens)
	bondingCurveSupply, err := k.GetBondingCurveSupply(ctx, msg.Amount.Denom)
	if err != nil {
		// Fallback to stored supply minus founder tokens
		bondingCurveSupply = userToken.CurrentSupply.Sub(userToken.FounderTokensClaimed)
		if bondingCurveSupply.IsNegative() {
			bondingCurveSupply = cosmossdk_io_math.ZeroInt()
		}
	}

	// Validate seller has enough tokens
	sellerAddr, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		return nil, fmt.Errorf("invalid seller address: %w", err)
	}

	sellerBalance := k.bankKeeper.GetBalance(ctx, sellerAddr, msg.Amount.Denom)
	if sellerBalance.Amount.LT(msg.Amount.Amount) {
		return nil, fmt.Errorf("insufficient tokens to sell: available %s, requested %s", sellerBalance.Amount.String(), msg.Amount.Amount.String())
	}

	// Calculate total payout using optimized mathematical approach
	totalPayout := k.CalculatePayoutFromTokens(ctx, bondingCurveSupply, msg.Amount.Amount)

	// Check if payout meets minimum price
	if totalPayout.LT(msg.MinPrice) {
		return nil, fmt.Errorf("price received %s is less than minimum price %s", totalPayout.String(), msg.MinPrice.String())
	}

	// Burn tokens from seller
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sellerAddr, types.ModuleName, sdk.NewCoins(msg.Amount))
	if err != nil {
		return nil, fmt.Errorf("failed to send tokens to module: %w", err)
	}

	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(msg.Amount))
	if err != nil {
		return nil, fmt.Errorf("failed to burn tokens: %w", err)
	}

	// Transfer payout from module to seller
	payoutCoin := sdk.NewCoin("unuah", totalPayout) // N$ denomination
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sellerAddr, sdk.NewCoins(payoutCoin))
	if err != nil {
		return nil, fmt.Errorf("failed to send payout: %w", err)
	}

	// Update supply tracking in store
	userToken.CurrentSupply = userToken.CurrentSupply.Sub(msg.Amount.Amount)
	k.SetUserToken(ctx, userToken)

	// Emit optimized event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"sell_tokens",
			sdk.NewAttribute("seller", msg.Seller),
			sdk.NewAttribute("denom", msg.Amount.Denom),
			sdk.NewAttribute("tokens_sold", msg.Amount.Amount.String()),
			sdk.NewAttribute("payout_received", totalPayout.String()),
			sdk.NewAttribute("new_supply", userToken.CurrentSupply.String()),
		),
	})

	return &types.MsgSellTokensResponse{
		PriceReceived: totalPayout,
	}, nil
}

func (k msgServer) BuyFounderTokens(goCtx context.Context, msg *types.MsgBuyFounderTokens) (*types.MsgBuyFounderTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get user token to verify it exists and get creator
	userToken, found := k.GetUserToken(ctx, msg.Denom)
	if !found {
		return nil, types.ErrTokenNotFound
	}

	// Verify buyer is the token creator
	if userToken.Creator != msg.Buyer {
		return nil, fmt.Errorf("only token creator can buy founder tokens")
	}

	// Check if founder tokens already claimed
	if !userToken.FounderTokensClaimed.IsZero() {
		return nil, fmt.Errorf("founder tokens already claimed")
	}

	// Check if founder offer has expired (7 days from token creation)
	currentTime := ctx.BlockTime().Unix()
	founderOfferExpiry := ctx.BlockTime().AddDate(0, 0, 7).Unix() // 7 days from creation
	if currentTime >= founderOfferExpiry {
		// Founder offer has expired, automatically transfer tokens to AI CEO
		err := k.TransferExpiredFounderTokensToAiCeo(ctx, msg.Denom)
		if err != nil {
			return nil, fmt.Errorf("failed to transfer expired founder tokens to AI CEO: %w", err)
		}
		return nil, fmt.Errorf("founder offer has expired, tokens transferred to AI CEO")
	}

	// Fixed amount: 10M tokens (already reserved in module)
	founderTokenAmount := cosmossdk_io_math.NewInt(10_000_000)

	// Fixed price: 500 N$ total (0.00005 N$ per token)
	totalCost := cosmossdk_io_math.NewInt(500)

	// Transfer payment from buyer to module
	buyerAddr, err := sdk.AccAddressFromBech32(msg.Buyer)
	if err != nil {
		return nil, fmt.Errorf("invalid buyer address: %w", err)
	}

	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if moduleAddr == nil {
		return nil, fmt.Errorf("module address not found")
	}

	paymentCoin := sdk.NewCoin("unuah", totalCost) // N$ denomination
	err = k.bankKeeper.SendCoins(ctx, buyerAddr, moduleAddr, sdk.NewCoins(paymentCoin))
	if err != nil {
		return nil, fmt.Errorf("failed to transfer payment: %w", err)
	}

	// Update founder tokens claimed in state
	userToken.FounderTokensClaimed = founderTokenAmount
	k.SetUserToken(ctx, userToken)

	// Transfer reserved founder tokens from module to buyer (no minting needed)
	founderTokens := sdk.NewCoin(msg.Denom, founderTokenAmount)

	// Create vesting account for 1 year lock period
	endTime := ctx.BlockTime().AddDate(1, 0, 0).Unix() // 1 year from now

	// Check if buyer already has an account
	if !k.accountKeeper.HasAccount(ctx, buyerAddr) {
		// Create new vesting account
		baseAccount := k.accountKeeper.NewAccountWithAddress(ctx, buyerAddr)
		k.accountKeeper.SetAccount(ctx, baseAccount)
	}

	// Get existing account and convert to vesting if needed
	existingAccount := k.accountKeeper.GetAccount(ctx, buyerAddr)

	// Check if it's already a vesting account
	if _, isVesting := existingAccount.(*vestingtypes.ContinuousVestingAccount); !isVesting {
		if _, isDelayedVesting := existingAccount.(*vestingtypes.DelayedVestingAccount); !isDelayedVesting {
			// Convert to base account for vesting
			baseAccount, ok := existingAccount.(*authtypes.BaseAccount)
			if !ok {
				return nil, fmt.Errorf("cannot convert account to vesting account")
			}

			// Create continuous vesting account with 1 year lock
			baseVestingAccount, err := vestingtypes.NewBaseVestingAccount(baseAccount, sdk.NewCoins(founderTokens), endTime)
			if err != nil {
				return nil, fmt.Errorf("failed to create base vesting account: %w", err)
			}

			vestingAccount := vestingtypes.NewContinuousVestingAccountRaw(baseVestingAccount, ctx.BlockTime().Unix())
			k.accountKeeper.SetAccount(ctx, vestingAccount)
		}
	}

	// Send tokens to the vesting account
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, buyerAddr, sdk.NewCoins(founderTokens))
	if err != nil {
		return nil, fmt.Errorf("failed to send founder tokens to vesting account: %w", err)
	}

	// Emit event for successful founder token purchase
	founderPrice := cosmossdk_io_math.LegacyMustNewDecFromStr("0.00005")
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"buy_founder_tokens",
			sdk.NewAttribute("buyer", msg.Buyer),
			sdk.NewAttribute("denom", msg.Denom),
			sdk.NewAttribute("amount", founderTokenAmount.String()),
			sdk.NewAttribute("total_cost", totalCost.String()),
			sdk.NewAttribute("founder_price", founderPrice.String()),
			sdk.NewAttribute("vesting_end_time", fmt.Sprintf("%d", endTime)),
			sdk.NewAttribute("lock_period", "1_year"),
		),
	)

	return &types.MsgBuyFounderTokensResponse{
		TokensReceived: founderTokenAmount,
	}, nil
}

func (k msgServer) ClaimFounderTokens(goCtx context.Context, msg *types.MsgClaimFounderTokens) (*types.MsgClaimFounderTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate input
	if msg.Amount.IsZero() || msg.Amount.IsNegative() {
		return nil, fmt.Errorf("invalid amount: %s", msg.Amount.String())
	}

	// Get params for distribution logic
	params := k.GetParams(ctx)
	founderPrice := params.FounderTranchePrice

	// Calculate total cost (amount * founder_tranche_price)
	totalCost := founderPrice.MulInt(msg.Amount).TruncateInt()

	// Transfer payment from claimer to module
	claimerAddr, err := sdk.AccAddressFromBech32(msg.Claimer)
	if err != nil {
		return nil, fmt.Errorf("invalid claimer address: %w", err)
	}

	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if moduleAddr == nil {
		return nil, fmt.Errorf("module address not found")
	}

	paymentCoin := sdk.NewCoin("unuah", totalCost) // N$ denomination
	err = k.bankKeeper.SendCoins(ctx, claimerAddr, moduleAddr, sdk.NewCoins(paymentCoin))
	if err != nil {
		return nil, fmt.Errorf("failed to transfer payment: %w", err)
	}

	// Check minimum purchase requirement before claiming
	minCreatorPurchase := params.MinCreatorPurchase
	minCreatorPurchaseInt := minCreatorPurchase.TruncateInt()

	if totalCost.LT(minCreatorPurchaseInt) {
		// Creator doesn't meet minimum - send tokens to AI CEO wallet
		if params.AiCeoWallet == "" {
			return nil, fmt.Errorf("AI CEO wallet not configured")
		}

		aiCeoAddr, err := sdk.AccAddressFromBech32(params.AiCeoWallet)
		if err != nil {
			return nil, fmt.Errorf("invalid AI CEO wallet address: %w", err)
		}

		// Update founder tokens claimed in state
		userToken, found := k.GetUserToken(ctx, msg.Denom)
		if !found {
			return nil, fmt.Errorf("user token not found: %s", msg.Denom)
		}

		userToken.FounderTokensClaimed = userToken.FounderTokensClaimed.Add(msg.Amount)
		k.SetUserToken(ctx, userToken)

		// Mint tokens directly to AI CEO wallet (not affecting bonding curve)
		founderTokens := sdk.NewCoin(msg.Denom, msg.Amount)
		err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(founderTokens))
		if err != nil {
			return nil, fmt.Errorf("failed to mint founder tokens for AI CEO: %w", err)
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, aiCeoAddr, sdk.NewCoins(founderTokens))
		if err != nil {
			return nil, fmt.Errorf("failed to send founder tokens to AI CEO: %w", err)
		}

		// Emit event for AI CEO allocation
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"founder_tokens_to_ai_ceo",
				sdk.NewAttribute("original_claimer", msg.Claimer),
				sdk.NewAttribute("ai_ceo_wallet", params.AiCeoWallet),
				sdk.NewAttribute("denom", msg.Denom),
				sdk.NewAttribute("amount", msg.Amount.String()),
				sdk.NewAttribute("total_cost", totalCost.String()),
				sdk.NewAttribute("reason", "minimum_purchase_not_met"),
			),
		)

		return &types.MsgClaimFounderTokensResponse{}, nil
	}

	// Claim founder tokens through keeper (normal case - minimum purchase met)
	err = k.Keeper.ClaimFounderTokens(ctx, msg.Denom, msg.Claimer, msg.Amount)
	if err != nil {
		return nil, err
	}

	// Mint founder tokens to claimer (normal case)
	founderTokens := sdk.NewCoin(msg.Denom, msg.Amount)
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(founderTokens))
	if err != nil {
		return nil, fmt.Errorf("failed to mint founder tokens: %w", err)
	}

	// Create vesting account for 1 year lock period
	endTime := ctx.BlockTime().AddDate(1, 0, 0).Unix() // 1 year from now

	// Check if claimer already has an account
	if !k.accountKeeper.HasAccount(ctx, claimerAddr) {
		// Create new vesting account
		baseAccount := k.accountKeeper.NewAccountWithAddress(ctx, claimerAddr)
		k.accountKeeper.SetAccount(ctx, baseAccount)
	}

	// Get existing account and convert to vesting if needed
	existingAccount := k.accountKeeper.GetAccount(ctx, claimerAddr)

	// Check if it's already a vesting account
	if _, isVesting := existingAccount.(*vestingtypes.ContinuousVestingAccount); !isVesting {
		if _, isDelayedVesting := existingAccount.(*vestingtypes.DelayedVestingAccount); !isDelayedVesting {
			// Convert to base account for vesting
			baseAccount, ok := existingAccount.(*authtypes.BaseAccount)
			if !ok {
				return nil, fmt.Errorf("cannot convert account to vesting account")
			}

			// Create continuous vesting account with 1 year lock
			baseVestingAccount, err := vestingtypes.NewBaseVestingAccount(baseAccount, sdk.NewCoins(founderTokens), endTime)
			if err != nil {
				return nil, fmt.Errorf("failed to create base vesting account: %w", err)
			}

			vestingAccount := vestingtypes.NewContinuousVestingAccountRaw(baseVestingAccount, ctx.BlockTime().Unix())
			k.accountKeeper.SetAccount(ctx, vestingAccount)
		}
	}

	// Send tokens to the vesting account
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, claimerAddr, sdk.NewCoins(founderTokens))
	if err != nil {
		return nil, fmt.Errorf("failed to send founder tokens to vesting account: %w", err)
	}

	// Emit event for successful claim with vesting
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"claim_founder_tokens",
			sdk.NewAttribute("claimer", msg.Claimer),
			sdk.NewAttribute("denom", msg.Denom),
			sdk.NewAttribute("amount", msg.Amount.String()),
			sdk.NewAttribute("total_cost", totalCost.String()),
			sdk.NewAttribute("founder_price", founderPrice.String()),
			sdk.NewAttribute("vesting_end_time", fmt.Sprintf("%d", endTime)),
			sdk.NewAttribute("lock_period", "1_year"),
		),
	)

	return &types.MsgClaimFounderTokensResponse{}, nil
}

func (k msgServer) StartLBP(goCtx context.Context, msg *types.MsgStartLBP) (*types.MsgStartLBPResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get user token
	userToken, found := k.GetUserToken(ctx, msg.Denom)
	if !found {
		return nil, types.ErrTokenNotFound
	}

	// Check if creator is the token creator
	if userToken.Creator != msg.Creator {
		return nil, types.ErrUnauthorized
	}

	// Check if LBP is already active
	if userToken.LbpActive {
		return nil, fmt.Errorf("LBP already active for token %s", msg.Denom)
	}

	// Create LBP pool with 90/10 initial weights (90% N$, 10% UserToken)
	// Pool will transition to 50/50 over 72 hours
	startTime := ctx.BlockTime()
	endTime := startTime.Add(72 * time.Hour)

	// Initial weights: 90% N$, 10% UserToken
	initialWeights := []cosmossdk_io_math.Int{
		cosmossdk_io_math.NewInt(900), // 90% for N$
		cosmossdk_io_math.NewInt(100), // 10% for UserToken
	}

	// Target weights: 50% N$, 50% UserToken
	targetWeights := []cosmossdk_io_math.Int{
		cosmossdk_io_math.NewInt(500), // 50% for N$
		cosmossdk_io_math.NewInt(500), // 50% for UserToken
	}

	// Pool assets
	poolAssets := []balancertypes.PoolAsset{
		{
			Token:  sdk.NewCoin("unuah", cosmossdk_io_math.NewInt(1000000)), // 1M N$ initial liquidity
			Weight: initialWeights[0],
		},
		{
			Token:  sdk.NewCoin(msg.Denom, cosmossdk_io_math.NewInt(100000)), // 100K UserToken initial liquidity
			Weight: initialWeights[1],
		},
	}

	// Smooth weight change parameters
	smoothWeightChangeParams := &balancertypes.SmoothWeightChangeParams{
		StartTime:          startTime,
		Duration:           72 * time.Hour,
		InitialPoolWeights: poolAssets,
		TargetPoolWeights: []balancertypes.PoolAsset{
			{
				Token:  sdk.NewCoin("unuah", cosmossdk_io_math.NewInt(1000000)),
				Weight: targetWeights[0],
			},
			{
				Token:  sdk.NewCoin(msg.Denom, cosmossdk_io_math.NewInt(100000)),
				Weight: targetWeights[1],
			},
		},
	}

	// Create balancer pool with LBP parameters
	poolParams := balancertypes.PoolParams{
		SwapFee:                  cosmossdk_io_math.LegacyNewDecWithPrec(3, 3), // 0.3% swap fee
		ExitFee:                  cosmossdk_io_math.LegacyNewDecWithPrec(0, 3), // 0% exit fee
		SmoothWeightChangeParams: smoothWeightChangeParams,
	}

	// Create LBP pool using poolmanager

	// Create the balancer pool message
	createPoolMsg := balancertypes.NewMsgCreateBalancerPool(
		sdk.MustAccAddressFromBech32(msg.Creator),
		poolParams,
		poolAssets,
		msg.Creator, // Pool governor
	)

	// Create pool through poolmanager
	poolId, err := k.poolManagerKeeper.CreatePool(ctx, createPoolMsg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create LBP pool")
	}

	// Update user token with LBP info
	userToken.LbpActive = true
	userToken.LbpStartTime = startTime.Unix()
	k.SetUserToken(ctx, userToken)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"start_lbp",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("denom", msg.Denom),
			sdk.NewAttribute("pool_id", fmt.Sprintf("%d", poolId)),
			sdk.NewAttribute("start_time", startTime.String()),
			sdk.NewAttribute("end_time", endTime.String()),
			sdk.NewAttribute("initial_weights", fmt.Sprintf("%d/%d", initialWeights[0], initialWeights[1])),
			sdk.NewAttribute("target_weights", fmt.Sprintf("%d/%d", targetWeights[0], targetWeights[1])),
			sdk.NewAttribute("duration", "72h"),
		),
	)

	return &types.MsgStartLBPResponse{}, nil
}

func (k msgServer) CreateVestingAccount(goCtx context.Context, msg *types.MsgCreateVestingAccount) (*types.MsgCreateVestingAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "invalid creator address: %s", err)
	}

	// Validate to_address
	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "invalid to_address: %s", err)
	}

	// Validate end_time is in the future
	if msg.EndTime <= ctx.BlockTime().Unix() {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "end_time must be in the future")
	}

	// Validate amount is not empty
	if len(msg.Amount) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "amount cannot be empty")
	}

	// Check if account already exists
	if k.accountKeeper.HasAccount(ctx, toAddr) {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "account %s already exists", msg.ToAddress)
	}

	// Create base account first using AccountKeeper to get proper account number
	accountI := k.accountKeeper.NewAccountWithAddress(ctx, toAddr)
	baseAccount, ok := accountI.(*authtypes.BaseAccount)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "failed to cast account to BaseAccount")
	}

	// Create vesting account based on type
	var vestingAccount sdk.AccountI
	if msg.Delayed {
		// Create delayed vesting account
		baseVestingAccount, err := vestingtypes.NewBaseVestingAccount(baseAccount, msg.Amount, msg.EndTime)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create base vesting account")
		}
		vestingAccount = vestingtypes.NewDelayedVestingAccountRaw(baseVestingAccount)
	} else {
		// Create continuous vesting account
		baseVestingAccount, err := vestingtypes.NewBaseVestingAccount(baseAccount, msg.Amount, msg.EndTime)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create base vesting account")
		}
		vestingAccount = vestingtypes.NewContinuousVestingAccountRaw(baseVestingAccount, ctx.BlockTime().Unix())
	}

	// Set the vesting account
	k.accountKeeper.SetAccount(ctx, vestingAccount)

	// Transfer tokens from creator to vesting account
	err = k.bankKeeper.SendCoins(ctx, creatorAddr, toAddr, msg.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to send coins to vesting account")
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"create_vesting_account",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("to_address", msg.ToAddress),
			sdk.NewAttribute("amount", sdk.NewCoins(msg.Amount...).String()),
			sdk.NewAttribute("end_time", fmt.Sprintf("%d", msg.EndTime)),
			sdk.NewAttribute("delayed", fmt.Sprintf("%t", msg.Delayed)),
		),
	)

	return &types.MsgCreateVestingAccountResponse{}, nil
}

// createNuahUserTokenPool creates a basic N$/UserToken pool for trading
func (k msgServer) createNuahUserTokenPool(ctx sdk.Context, userTokenDenom string, creator string) error {
	// Define initial pool assets with equal weights (50/50)
	poolAssets := []balancertypes.PoolAsset{
		{
			Token:  sdk.NewCoin("unuah", cosmossdk_io_math.NewInt(1000000)), // 1M N$ initial liquidity
			Weight: cosmossdk_io_math.NewInt(500),                           // 50%
		},
		{
			Token:  sdk.NewCoin(userTokenDenom, cosmossdk_io_math.NewInt(1000000)), // 1M UserToken initial liquidity
			Weight: cosmossdk_io_math.NewInt(500),                                  // 50%
		},
	}

	// Pool parameters
	poolParams := balancertypes.PoolParams{
		SwapFee: cosmossdk_io_math.LegacyNewDecWithPrec(3, 3), // 0.3% swap fee
		ExitFee: cosmossdk_io_math.LegacyNewDecWithPrec(0, 3), // 0% exit fee
	}

	// Create the balancer pool
	pool, err := balancertypes.NewBalancerPool(
		0, // pool ID will be assigned by gamm keeper
		poolParams,
		poolAssets,
		"", // future governor
		ctx.BlockTime(),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create balancer pool")
	}

	// TODO: Implement pool creation
	// This will be implemented when integrating with Osmosis pools
	_ = pool // Avoid unused variable error

	return nil
}

func (k msgServer) CreateReferralProgram(goCtx context.Context, msg *types.MsgCreateReferralProgram) (*types.MsgCreateReferralProgramResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "invalid creator address: %s", err)
	}

	// Check if user has available quota slots
	quota, found := k.GetUserReferralQuota(ctx, msg.Creator)
	if !found {
		// Initialize quota for new user with 3 slots
		quota = types.UserReferralQuota{
			User:          msg.Creator,
			TotalSlots:    3,
			UsedSlots:     0,
			LastResetTime: ctx.BlockTime().Unix(),
		}
	}

	// Check if user has available slots
	if quota.UsedSlots >= quota.TotalSlots {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "no available referral slots remaining: %d/%d used", quota.UsedSlots, quota.TotalSlots)
	}

	// Check if referral program already exists for this token
	_, found = k.GetReferralProgram(ctx, msg.TokenDenom)
	if found {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "referral program already exists for token: %s", msg.TokenDenom)
	}

	// Create referral program
	program := types.ReferralProgram{
		Creator:        msg.Creator,
		TokenDenom:     msg.TokenDenom,
		AvailableLinks: 3, // Default value: start with 3 links
		UsedLinks:      0,
		LastResetTime:  ctx.BlockTime().Unix(),
		IsActive:       true,
	}

	// Update user quota
	quota.UsedSlots++
	k.SetUserReferralQuota(ctx, quota)

	// Store the program
	k.SetReferralProgram(ctx, program)

	return &types.MsgCreateReferralProgramResponse{}, nil
}

func (k msgServer) ActivateReferral(goCtx context.Context, msg *types.MsgActivateReferral) (*types.MsgActivateReferralResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate referee address
	if _, err := sdk.AccAddressFromBech32(msg.Referee); err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "invalid referee address: %s", err)
	}

	// Check if referral code exists and is valid
	program, found := k.GetReferralProgram(ctx, msg.ReferralCode)
	if !found {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "referral program not found: %s", msg.ReferralCode)
	}

	if !program.IsActive {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "referral program is not active")
	}

	// Check if referee already activated this referral
	_, found = k.GetReferralActivation(ctx, msg.ReferralCode)
	if found {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "referral already activated for this user")
	}

	// Check if program has available links
	if program.UsedLinks >= program.AvailableLinks {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "no available referral links remaining")
	}

	// Create activation record
	activation := types.ReferralActivation{
		ReferralCode:   msg.ReferralCode,
		Referee:        msg.Referee,
		Referrer:       program.Creator,
		TokenDenom:     program.TokenDenom,
		ActivationTime: ctx.BlockTime().Unix(),
	}

	// Store activation
	k.SetReferralActivation(ctx, activation)

	// Update program usage
	program.UsedLinks++
	k.SetReferralProgram(ctx, program)

	// Reward both participants with 1 token each
	rewardAmount := cosmossdk_io_math.NewInt(1000000) // 1 token with 6 decimals
	rewardCoin := sdk.NewCoin(program.TokenDenom, rewardAmount)

	// Mint tokens for rewards (2 tokens total)
	totalRewardAmount := rewardAmount.Mul(cosmossdk_io_math.NewInt(2)) // 2 tokens
	totalRewardCoin := sdk.NewCoin(program.TokenDenom, totalRewardAmount)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(totalRewardCoin)); err != nil {
		return nil, errors.Wrapf(err, "failed to mint reward coins")
	}

	// Send reward to referrer
	referrerAddr, err := sdk.AccAddressFromBech32(activation.Referrer)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "invalid referrer address: %s", err)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, referrerAddr, sdk.NewCoins(rewardCoin)); err != nil {
		return nil, errors.Wrapf(err, "failed to send reward coins to referrer")
	}

	// Send reward to referee
	refereeAddr, err := sdk.AccAddressFromBech32(activation.Referee)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "invalid referee address: %s", err)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, refereeAddr, sdk.NewCoins(rewardCoin)); err != nil {
		return nil, errors.Wrapf(err, "failed to send reward coins to referee")
	}

	return &types.MsgActivateReferralResponse{}, nil
}

// UpdateParams updates the module parameters
func (k msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetParams(ctx, req.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}

// distributeInitialTokens distributes 100M tokens according to tokenomics
func (k msgServer) distributeInitialTokens(ctx sdk.Context, denom string, totalSupply cosmossdk_io_math.Int) error {
	params := k.GetParams(ctx)

	// Calculate distribution amounts
	bondingCurveAmount := cosmossdk_io_math.NewInt(30_000_000) // 30M to bonding curve
	platformAmount := cosmossdk_io_math.NewInt(10_000_000)     // 10M to platform
	referralAmount := cosmossdk_io_math.NewInt(10_000_000)     // 10M to referral
	aiCeoAmount := cosmossdk_io_math.NewInt(40_000_000)        // 40M to AI CEO
	founderOfferAmount := cosmossdk_io_math.NewInt(10_000_000) // 10M reserved for founder offer

	// 30M stays in module for bonding curve - no transfer needed

	// 10M to platform wallet
	if !platformAmount.IsZero() && params.PlatformFeeWallet != "" {
		platformAddr, err := sdk.AccAddressFromBech32(params.PlatformFeeWallet)
		if err == nil {
			platformCoin := sdk.NewCoin(denom, platformAmount)
			err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, platformAddr, sdk.NewCoins(platformCoin))
			if err != nil {
				return fmt.Errorf("failed to send tokens to platform: %w", err)
			}
		}
	}

	// 10M to referral wallet
	if !referralAmount.IsZero() && params.ReferralWallet != "" {
		referralAddr, err := sdk.AccAddressFromBech32(params.ReferralWallet)
		if err == nil {
			referralCoin := sdk.NewCoin(denom, referralAmount)
			err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, referralAddr, sdk.NewCoins(referralCoin))
			if err != nil {
				return fmt.Errorf("failed to send tokens to referral: %w", err)
			}
		}
	}

	// 40M to AI CEO wallet
	if !aiCeoAmount.IsZero() && params.AiCeoWallet != "" {
		aiCeoAddr, err := sdk.AccAddressFromBech32(params.AiCeoWallet)
		if err == nil {
			aiCeoCoin := sdk.NewCoin(denom, aiCeoAmount)
			err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, aiCeoAddr, sdk.NewCoins(aiCeoCoin))
			if err != nil {
				return fmt.Errorf("failed to send tokens to AI CEO: %w", err)
			}
		}
	}

	// 10M reserved in module for founder offer - no transfer needed
	_ = founderOfferAmount
	_ = bondingCurveAmount

	return nil
}

// TransferExpiredFounderTokensToAiCeo transfers unclaimed founder tokens to AI CEO after expiry
func (k msgServer) TransferExpiredFounderTokensToAiCeo(ctx sdk.Context, denom string) error {
	userToken, found := k.GetUserToken(ctx, denom)
	if !found {
		return fmt.Errorf("user token not found: %s", denom)
	}

	// Check if founder offer has expired (7 days from token creation)
	currentTime := ctx.BlockTime().Unix()
	founderOfferExpiry := ctx.BlockTime().AddDate(0, 0, 7).Unix() // 7 days from creation
	if currentTime < founderOfferExpiry {
		return fmt.Errorf("founder offer has not expired yet")
	}

	// Check if founder tokens are still unclaimed
	founderOfferAmount := cosmossdk_io_math.NewInt(10_000_000)
	if !userToken.FounderTokensClaimed.IsZero() {
		return fmt.Errorf("founder tokens already claimed")
	}

	params := k.GetParams(ctx)
	if params.AiCeoWallet == "" {
		return fmt.Errorf("AI CEO wallet not configured")
	}

	aiCeoAddr, err := sdk.AccAddressFromBech32(params.AiCeoWallet)
	if err != nil {
		return fmt.Errorf("invalid AI CEO wallet address: %w", err)
	}

	// Transfer the 10M founder tokens to AI CEO
	founderTokens := sdk.NewCoin(denom, founderOfferAmount)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, aiCeoAddr, sdk.NewCoins(founderTokens))
	if err != nil {
		return fmt.Errorf("failed to send expired founder tokens to AI CEO: %w", err)
	}

	// Update founder tokens claimed to prevent double claiming
	userToken.FounderTokensClaimed = founderOfferAmount
	k.SetUserToken(ctx, userToken)

	// Emit event for expired founder tokens transfer
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"expired_founder_tokens_to_ai_ceo",
			sdk.NewAttribute("denom", denom),
			sdk.NewAttribute("ai_ceo_wallet", params.AiCeoWallet),
			sdk.NewAttribute("amount", founderOfferAmount.String()),
			sdk.NewAttribute("reason", "founder_offer_expired"),
			sdk.NewAttribute("expired_at", fmt.Sprintf("%d", founderOfferExpiry)),
		),
	)

	return nil
}

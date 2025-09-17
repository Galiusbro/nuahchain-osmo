package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v30/x/tokenfactory/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

type (
	Keeper struct {
		cdc       codec.BinaryCodec
		storeKey  storetypes.StoreKey
		memKey    storetypes.StoreKey
		authority string

		// External keepers
		tokenfactoryKeeper tokenfactorykeeper.Keeper
		bankKeeper         types.BankKeeper
		accountKeeper      types.AccountKeeper
		gammKeeper         types.GammKeeper
		poolManagerKeeper  types.PoolManagerKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	authority string,
	tokenfactoryKeeper tokenfactorykeeper.Keeper,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	gammKeeper types.GammKeeper,
	poolManagerKeeper types.PoolManagerKeeper,
) *Keeper {
	return &Keeper{
		cdc:                cdc,
		storeKey:           storeKey,
		memKey:             memKey,
		authority:          authority,
		tokenfactoryKeeper: tokenfactoryKeeper,
		bankKeeper:         bankKeeper,
		accountKeeper:      accountKeeper,
		gammKeeper:         gammKeeper,
		poolManagerKeeper:  poolManagerKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte("params"))
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set([]byte("params"), bz)
}

// GetTokenSupply gets the current supply of a token
func (k Keeper) GetTokenSupply(ctx sdk.Context, denom string) (math.Int, error) {
	// Get token supply from bank keeper
	supply := k.bankKeeper.GetSupply(ctx, denom)
	return supply.Amount, nil
}

// GetBondingCurveSupply returns the supply that should be used for bonding curve calculations
// For new token distribution model, bonding curve gets 30M tokens out of 100M total
func (k Keeper) GetBondingCurveSupply(ctx sdk.Context, denom string) (math.Int, error) {
	// Get user token to check if it exists
	_, found := k.GetUserToken(ctx, denom)
	if !found {
		return math.ZeroInt(), fmt.Errorf("token not found: %s", denom)
	}

	// For the new distribution model:
	// - Total supply: 100M tokens
	// - Bonding curve gets: 30M tokens
	// - Platform gets: 10M tokens
	// - Referral wallet gets: 10M tokens
	// - AI CEO gets: 40M tokens (+ 10M more if founder doesn't buy)
	// - Founder can buy: 10M tokens

	// The bonding curve supply should track how many tokens have been sold from the 30M allocation
	// We can calculate this by checking the actual token supply in circulation vs the initial distribution
	totalSupply, err := k.GetTokenSupply(ctx, denom)
	if err != nil {
		return math.ZeroInt(), err
	}

	// Initial distribution is 100M, bonding curve starts with 30M available
	bondingCurveMaxSupply := math.NewInt(30_000_000) // 30M tokens for bonding curve
	initialTotalSupply := math.NewInt(100_000_000)   // 100M total initial supply

	// If total supply is still at initial 100M, no tokens have been sold from bonding curve
	if totalSupply.Equal(initialTotalSupply) {
		return math.ZeroInt(), nil // No tokens sold yet from bonding curve
	}

	// Calculate how many tokens have been sold from bonding curve
	// This assumes any increase in supply beyond 100M comes from bonding curve sales
	bondingCurveSold := totalSupply.Sub(initialTotalSupply)
	if bondingCurveSold.IsNegative() {
		bondingCurveSold = math.ZeroInt()
	}

	// Ensure we don't exceed the bonding curve max supply
	if bondingCurveSold.GT(bondingCurveMaxSupply) {
		bondingCurveSold = bondingCurveMaxSupply
	}

	return bondingCurveSold, nil
}

// CalculateBondingCurvePrice calculates the current price based on supply
func (k Keeper) CalculateBondingCurvePrice(ctx sdk.Context, currentSupply math.Int) math.LegacyDec {
	// Linear bonding curve: price = 0.0002 + (currentSupply / 30M) * (1.0 - 0.0002)
	// Price ranges from 0.0002 N$ to 1.0 N$ for 30M tokens
	maxSupply := math.NewInt(30_000_000)        // 30M tokens
	minPrice := math.LegacyNewDecWithPrec(2, 4) // 0.0002
	maxPrice := math.LegacyOneDec()             // 1.0

	if currentSupply.GTE(maxSupply) {
		return maxPrice
	}

	// Calculate linear progression
	progress := math.LegacyNewDecFromInt(currentSupply).Quo(math.LegacyNewDecFromInt(maxSupply))
	priceRange := maxPrice.Sub(minPrice)
	currentPrice := minPrice.Add(progress.Mul(priceRange))

	return currentPrice
}

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// Set params
	k.SetParams(ctx, genState.Params)

	// Initialize user tokens
	for _, userToken := range genState.UserTokens {
		// TODO: Store user token data
		_ = userToken
	}
}

// ExportGenesis returns the module's exported genesis
// GetUserToken retrieves a user token by denom
func (k Keeper) GetUserToken(ctx sdk.Context, denom string) (types.UserToken, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.UserTokenKey(denom))
	if bz == nil {
		return types.UserToken{}, false
	}

	var userToken types.UserToken
	k.cdc.MustUnmarshal(bz, &userToken)
	return userToken, true
}

// SetUserToken stores a user token
func (k Keeper) SetUserToken(ctx sdk.Context, userToken types.UserToken) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&userToken)
	store.Set(types.UserTokenKey(userToken.Denom), bz)
}

// GetAllUserTokens returns all user tokens
func (k Keeper) GetAllUserTokens(ctx sdk.Context) []*types.UserToken {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.UserTokenKeyPrefix)
	defer iterator.Close()

	var userTokens []*types.UserToken
	for ; iterator.Valid(); iterator.Next() {
		var userToken types.UserToken
		k.cdc.MustUnmarshal(iterator.Value(), &userToken)
		userTokens = append(userTokens, &userToken)
	}

	return userTokens
}

// GetFounderTokensRemaining returns the remaining founder tokens that can be claimed
func (k Keeper) GetFounderTokensRemaining(ctx sdk.Context, denom string) math.Int {
	userToken, found := k.GetUserToken(ctx, denom)
	if !found {
		return math.ZeroInt()
	}

	params := k.GetParams(ctx)
	founderTrancheAmount := params.FounderTrancheAmount

	return founderTrancheAmount.Sub(userToken.FounderTokensClaimed)
}

// ClaimFounderTokens claims founder tokens for a user token with new distribution logic
func (k Keeper) ClaimFounderTokens(ctx sdk.Context, denom string, claimer string, amount math.Int) error {
	userToken, found := k.GetUserToken(ctx, denom)
	if !found {
		return fmt.Errorf("user token not found: %s", denom)
	}

	// Only the creator can claim founder tokens
	if userToken.Creator != claimer {
		return fmt.Errorf("only token creator can claim founder tokens")
	}

	// Check if there are enough founder tokens remaining
	remainingTokens := k.GetFounderTokensRemaining(ctx, denom)
	if amount.GT(remainingTokens) {
		return fmt.Errorf("insufficient founder tokens remaining: requested %s, available %s", amount.String(), remainingTokens.String())
	}

	// Minimum purchase check is now handled in msg_server.go
	// This function only handles the actual claiming logic

	// Update founder tokens claimed
	userToken.FounderTokensClaimed = userToken.FounderTokensClaimed.Add(amount)
	k.SetUserToken(ctx, userToken)

	return nil
}

// DistributeTokenPurchasePayment distributes payment according to new tokenomics:
// 30% to bonding curve, 10% to creator, 10% to referral, 40% to AI CEO, 10% to platform fee
func (k Keeper) DistributeTokenPurchasePayment(ctx sdk.Context, denom string, totalPayment math.Int) error {
	params := k.GetParams(ctx)
	userToken, found := k.GetUserToken(ctx, denom)
	if !found {
		return fmt.Errorf("user token not found: %s", denom)
	}

	// Calculate distribution amounts
	bondingCurveAmount := totalPayment.MulRaw(30).QuoRaw(100) // 30%
	creatorAmount := totalPayment.MulRaw(10).QuoRaw(100)      // 10%
	referralAmount := totalPayment.MulRaw(10).QuoRaw(100)     // 10%
	aiCeoAmount := totalPayment.MulRaw(40).QuoRaw(100)        // 40%
	platformFeeAmount := totalPayment.MulRaw(10).QuoRaw(100)  // 10%

	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if moduleAddr == nil {
		return fmt.Errorf("module address not found")
	}

	// 30% stays in module for bonding curve (already transferred)
	_ = bondingCurveAmount // This amount stays in the module for bonding curve calculations
	// Distribute other portions

	// 10% to creator
	if creatorAmount.IsPositive() {
		creatorAddr, err := sdk.AccAddressFromBech32(userToken.Creator)
		if err == nil {
			creatorCoin := sdk.NewCoin("unuah", creatorAmount)
			k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, sdk.NewCoins(creatorCoin))
		}
	}

	// 10% to referral wallet
	if referralAmount.IsPositive() && params.ReferralWallet != "" {
		referralAddr, err := sdk.AccAddressFromBech32(params.ReferralWallet)
		if err == nil {
			referralCoin := sdk.NewCoin("unuah", referralAmount)
			k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, referralAddr, sdk.NewCoins(referralCoin))
		}
	}

	// 40% to AI CEO wallet
	if aiCeoAmount.IsPositive() && params.AiCeoWallet != "" {
		aiCeoAddr, err := sdk.AccAddressFromBech32(params.AiCeoWallet)
		if err == nil {
			aiCeoCoin := sdk.NewCoin("unuah", aiCeoAmount)
			k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, aiCeoAddr, sdk.NewCoins(aiCeoCoin))
		}
	}

	// 10% to platform fee wallet
	if platformFeeAmount.IsPositive() && params.PlatformFeeWallet != "" {
		platformFeeAddr, err := sdk.AccAddressFromBech32(params.PlatformFeeWallet)
		if err == nil {
			platformFeeCoin := sdk.NewCoin("unuah", platformFeeAmount)
			k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, platformFeeAddr, sdk.NewCoins(platformFeeCoin))
		}
	}

	return nil
}

// CalculateTokensFromPayment calculates how many tokens can be bought with a given payment amount
// Uses optimized mathematical approach to avoid iterative calculations
// Now uses only 30% of payment for bonding curve calculation
func (k Keeper) CalculateTokensFromPayment(ctx sdk.Context, currentSupply math.Int, paymentAmount math.Int) math.Int {
	params := k.GetParams(ctx)
	startPrice := params.BondingCurveStartPrice
	endPrice := params.BondingCurveEndPrice
	maxSupply := params.BondingCurveMaxSupply

	// Only 30% of payment goes to bonding curve
	bondingCurvePayment := paymentAmount.MulRaw(30).QuoRaw(100)

	// Calculate linear coefficient: (endPrice - startPrice) / maxSupply
	priceRange := endPrice.Sub(startPrice)
	linearCoeff := priceRange.QuoInt(maxSupply)

	// Current price at supply: startPrice + linearCoeff * currentSupply
	currentPrice := startPrice.Add(linearCoeff.MulInt(currentSupply))

	// For small amounts, use simple approximation to avoid expensive calculations
	paymentDec := math.LegacyNewDecFromInt(bondingCurvePayment)

	// Estimate tokens using average price method (more gas efficient)
	// avgPrice ≈ currentPrice + (linearCoeff * estimatedTokens / 2)
	// Solving: paymentAmount = estimatedTokens * (currentPrice + linearCoeff * estimatedTokens / 2)

	if linearCoeff.IsZero() {
		// Constant price case
		return paymentDec.Quo(currentPrice).TruncateInt()
	}

	// Quadratic formula: a*n^2 + b*n - c = 0
	// where a = linearCoeff/2, b = currentPrice, c = paymentAmount
	a := linearCoeff.QuoInt64(2)
	b := currentPrice
	c := paymentDec

	// Discriminant = b^2 + 4ac (note: +4ac because c is positive in our equation)
	discriminant := b.Mul(b).Add(a.Mul(c).MulInt64(4))

	// n = (-b + sqrt(discriminant)) / (2a)
	sqrtDiscriminant, err := discriminant.ApproxSqrt()
	if err != nil {
		// Fallback to simple linear approximation
		return paymentDec.Quo(currentPrice).TruncateInt()
	}

	numerator := b.Neg().Add(sqrtDiscriminant)
	denominator := a.MulInt64(2)

	tokens := numerator.Quo(denominator).TruncateInt()

	// Ensure non-negative result and within reasonable bounds
	if tokens.IsNegative() || tokens.IsZero() {
		return math.ZeroInt()
	}

	// Cap at remaining supply to max
	remainingSupply := maxSupply.Sub(currentSupply)
	if tokens.GT(remainingSupply) {
		tokens = remainingSupply
	}

	return tokens
}

// CalculatePayoutFromTokens calculates payout when selling tokens
// Uses optimized mathematical approach instead of iterative calculation
func (k Keeper) CalculatePayoutFromTokens(ctx sdk.Context, currentSupply math.Int, tokensToSell math.Int) math.Int {
	params := k.GetParams(ctx)
	startPrice := params.BondingCurveStartPrice
	endPrice := params.BondingCurveEndPrice
	maxSupply := params.BondingCurveMaxSupply

	// Calculate linear coefficient
	priceRange := endPrice.Sub(startPrice)
	linearCoeff := priceRange.QuoInt(maxSupply)

	// Calculate integral of price function from (currentSupply - tokensToSell) to currentSupply
	// Integral of (startPrice + linearCoeff * x) dx = startPrice * x + linearCoeff * x^2 / 2

	startSupply := currentSupply.Sub(tokensToSell)
	endSupply := currentSupply

	// Ensure non-negative supply
	if startSupply.IsNegative() {
		startSupply = math.ZeroInt()
	}

	// Calculate area under curve
	// Area = integral(endSupply) - integral(startSupply)
	integralEnd := k.calculatePriceIntegral(startPrice, linearCoeff, endSupply)
	integralStart := k.calculatePriceIntegral(startPrice, linearCoeff, startSupply)

	payout := integralEnd.Sub(integralStart)

	// Convert to integer (payout is already in unuah units)
	payoutInt := payout.TruncateInt()

	// Ensure non-negative result
	if payoutInt.IsNegative() {
		return math.ZeroInt()
	}

	return payoutInt
}

// calculatePriceIntegral calculates the integral of the price function at given supply
func (k Keeper) calculatePriceIntegral(startPrice math.LegacyDec, linearCoeff math.LegacyDec, supply math.Int) math.LegacyDec {
	supplyDec := math.LegacyNewDecFromInt(supply)

	// Integral = startPrice * supply + linearCoeff * supply^2 / 2
	linearTerm := startPrice.Mul(supplyDec)
	quadraticTerm := linearCoeff.Mul(supplyDec).Mul(supplyDec).QuoInt64(2)

	return linearTerm.Add(quadraticTerm)
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:              k.GetParams(ctx),
		UserTokens:          k.GetAllUserTokens(ctx),
		ReferralPrograms:    k.GetAllReferralPrograms(ctx),
		ReferralActivations: k.GetAllReferralActivations(ctx),
	}
}

// Referral Program functions

// GetReferralProgram retrieves a referral program by token denom
func (k Keeper) GetReferralProgram(ctx sdk.Context, tokenDenom string) (types.ReferralProgram, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ReferralProgramKey(tokenDenom))
	if bz == nil {
		return types.ReferralProgram{}, false
	}

	var program types.ReferralProgram
	k.cdc.MustUnmarshal(bz, &program)
	return program, true
}

// SetReferralProgram stores a referral program
func (k Keeper) SetReferralProgram(ctx sdk.Context, program types.ReferralProgram) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&program)
	store.Set(types.ReferralProgramKey(program.TokenDenom), bz)
}

// GetAllReferralPrograms returns all referral programs
func (k Keeper) GetAllReferralPrograms(ctx sdk.Context) []*types.ReferralProgram {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ReferralProgramKeyPrefix)
	defer iterator.Close()

	var programs []*types.ReferralProgram
	for ; iterator.Valid(); iterator.Next() {
		var program types.ReferralProgram
		k.cdc.MustUnmarshal(iterator.Value(), &program)
		programs = append(programs, &program)
	}

	return programs
}

// GetReferralActivation retrieves a referral activation by referral code
func (k Keeper) GetReferralActivation(ctx sdk.Context, referralCode string) (types.ReferralActivation, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ReferralActivationKey(referralCode))
	if bz == nil {
		return types.ReferralActivation{}, false
	}

	var activation types.ReferralActivation
	k.cdc.MustUnmarshal(bz, &activation)
	return activation, true
}

// SetReferralActivation stores a referral activation
func (k Keeper) SetReferralActivation(ctx sdk.Context, activation types.ReferralActivation) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&activation)
	store.Set(types.ReferralActivationKey(activation.ReferralCode), bz)
}

// GetAllReferralActivations returns all referral activations
func (k Keeper) GetAllReferralActivations(ctx sdk.Context) []*types.ReferralActivation {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ReferralActivationKeyPrefix)
	defer iterator.Close()

	var activations []*types.ReferralActivation
	for ; iterator.Valid(); iterator.Next() {
		var activation types.ReferralActivation
		k.cdc.MustUnmarshal(iterator.Value(), &activation)
		activations = append(activations, &activation)
	}

	return activations
}

// ResetWeeklyLimits processes weekly link replenishment for all referral programs
func (k Keeper) ResetWeeklyLimits(ctx sdk.Context) {
	programs := k.GetAllReferralPrograms(ctx)
	for _, program := range programs {
		// If all available links were used, add 3 more links
		if program.UsedLinks >= program.AvailableLinks {
			program.AvailableLinks += 3
		}
		// Reset used links counter for the new week
		program.UsedLinks = 0
		program.LastResetTime = ctx.BlockTime().Unix()
		k.SetReferralProgram(ctx, *program)
	}
}

// ProcessWeeklyReset is an alias for ResetWeeklyLimits for backward compatibility
func (k Keeper) ProcessWeeklyReset(ctx sdk.Context) {
	k.ResetWeeklyLimits(ctx)
}

// ProcessReferralReward processes referral reward when someone uses a referral link
// This is separate from token purchases and should be called when referral conditions are met
func (k Keeper) ProcessReferralReward(ctx sdk.Context, referralCode string, rewardAmount math.Int, tokenDenom string) error {
	activation, found := k.GetReferralActivation(ctx, referralCode)
	if !found {
		return fmt.Errorf("referral activation not found: %s", referralCode)
	}

	program, found := k.GetReferralProgram(ctx, activation.TokenDenom)
	if !found {
		return fmt.Errorf("referral program not found for token: %s", activation.TokenDenom)
	}

	// Check if program is active
	if !program.IsActive {
		return fmt.Errorf("referral program is not active")
	}

	// Check if we haven't exceeded available links
	if program.UsedLinks >= program.AvailableLinks {
		return fmt.Errorf("referral program has no available links")
	}

	// Update used links
	program.UsedLinks++
	k.SetReferralProgram(ctx, program)

	// Send reward to referrer
	referrerAddr, err := sdk.AccAddressFromBech32(activation.Referrer)
	if err != nil {
		return fmt.Errorf("invalid referrer address: %w", err)
	}

	rewardCoin := sdk.NewCoin(tokenDenom, rewardAmount)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(rewardCoin)); err != nil {
		return fmt.Errorf("failed to mint reward coins: %w", err)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, referrerAddr, sdk.NewCoins(rewardCoin)); err != nil {
		return fmt.Errorf("failed to send reward coins: %w", err)
	}

	return nil
}

// EndBlocker is called at the end of every block
func (k Keeper) EndBlocker(ctx sdk.Context) {
	// Check if it's time to reset weekly limits (every 7 days)
	currentTime := ctx.BlockTime().Unix()
	weekInSeconds := int64(7 * 24 * 60 * 60) // 7 days in seconds

	// Get all referral programs and check if they need weekly reset
	programs := k.GetAllReferralPrograms(ctx)
	for _, program := range programs {
		if currentTime-program.LastResetTime >= weekInSeconds {
			program.UsedLinks = 0
			program.LastResetTime = currentTime
			k.SetReferralProgram(ctx, *program)
		}
	}
}

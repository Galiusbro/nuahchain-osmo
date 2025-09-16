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
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		memKey   storetypes.StoreKey

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

// ClaimFounderTokens claims founder tokens for a user token
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

	// Update founder tokens claimed
	userToken.FounderTokensClaimed = userToken.FounderTokensClaimed.Add(amount)
	k.SetUserToken(ctx, userToken)

	return nil
}

// CalculateTokensFromPayment calculates how many tokens can be bought with given payment
// Uses optimized mathematical approach instead of iterative calculation
func (k Keeper) CalculateTokensFromPayment(ctx sdk.Context, currentSupply math.Int, paymentAmount math.Int) math.Int {
	params := k.GetParams(ctx)
	startPrice := params.BondingCurveStartPrice
	endPrice := params.BondingCurveEndPrice
	maxSupply := params.BondingCurveMaxSupply

	// Calculate linear coefficient: (endPrice - startPrice) / maxSupply
	priceRange := endPrice.Sub(startPrice)
	linearCoeff := priceRange.QuoInt(maxSupply)

	// Current price at supply: startPrice + linearCoeff * currentSupply
	currentPrice := startPrice.Add(linearCoeff.MulInt(currentSupply))

	// For small amounts, use simple approximation to avoid expensive calculations
	paymentDec := math.LegacyNewDecFromInt(paymentAmount)

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

	// Convert to integer (multiply by 1e6 for N$ units)
	payoutInt := payout.MulInt64(1000000).TruncateInt()

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
		Params:     k.GetParams(ctx),
		UserTokens: k.GetAllUserTokens(ctx),
	}
}

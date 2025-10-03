package keeper

import (
	"fmt"
	"strings"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

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

func (k Keeper) getTokenScaleFactors(ctx sdk.Context, denom string) (math.Int, math.LegacyDec, error) {
	scale := math.NewInt(1)
	scaleDec := math.LegacyNewDecFromInt(scale)

	metadata, found := k.bankKeeper.GetDenomMetaData(ctx, denom)
	if !found {
		return scale, scaleDec, fmt.Errorf("metadata not found for denom: %s", denom)
	}

	var decimals uint32
	for _, unit := range metadata.DenomUnits {
		if unit.Denom == metadata.Display {
			decimals = unit.Exponent
			break
		}
	}

	if decimals == 0 {
		return scale, scaleDec, nil
	}

	for i := uint32(0); i < decimals; i++ {
		scale = scale.MulRaw(10)
	}

	scaleDec = math.LegacyNewDecFromInt(scale)
	return scale, scaleDec, nil
}

func (k Keeper) mustGetTokenScaleFactors(ctx sdk.Context, denom string) (math.Int, math.LegacyDec) {
	// For user tokens (factory denoms), always use 6 decimals for now
	// TODO: Fix metadata retrieval to get the actual decimals
	if strings.Contains(denom, "factory/") {
		defaultScale := math.NewInt(1_000_000) // 6 decimals
		defaultScaleDec := math.LegacyNewDecFromInt(defaultScale)
		return defaultScale, defaultScaleDec
	}

	scale, scaleDec, err := k.getTokenScaleFactors(ctx, denom)
	if err != nil {
		// Default to 6 decimals (1,000,000 scale) for other tokens when metadata is not found
		defaultScale := math.NewInt(1_000_000)
		defaultScaleDec := math.LegacyNewDecFromInt(defaultScale)
		return defaultScale, defaultScaleDec
	}
	if scale.IsZero() {
		// Default to 6 decimals if scale is zero
		defaultScale := math.NewInt(1_000_000)
		defaultScaleDec := math.LegacyNewDecFromInt(defaultScale)
		return defaultScale, defaultScaleDec
	}
	return scale, scaleDec
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
	userToken, found := k.GetUserToken(ctx, denom)
	if !found {
		return math.ZeroInt(), fmt.Errorf("token not found: %s", denom)
	}

	scale, _ := k.mustGetTokenScaleFactors(ctx, denom)

	// Simple approach: track bonding curve supply based on CurrentSupply changes
	// Initial circulating supply is 60M (distributed immediately)
	// 10M platform + 10M referral + 40M AI CEO = 60M circulating
	// Any increase in CurrentSupply above 60M represents tokens sold from bonding curve
	initialCirculatingSupply := math.NewInt(60_000_000).Mul(scale)

	bondingCurveSold := userToken.CurrentSupply.Sub(initialCirculatingSupply)
	if bondingCurveSold.IsNegative() {
		bondingCurveSold = math.ZeroInt()
	}

	return bondingCurveSold, nil
}

// CalculateBondingCurvePrice calculates the current price based on supply
func (k Keeper) CalculateBondingCurvePrice(ctx sdk.Context, denom string, currentSupply math.Int) math.LegacyDec {
	// Linear bonding curve: price = 0.0002 + (currentSupply / 30M) * (1.0 - 0.0002)
	// Price ranges from 0.0002 N$ to 1.0 N$ for 30M tokens
	minPrice := math.LegacyNewDecWithPrec(2, 4) // 0.0002
	maxPrice := math.LegacyOneDec()             // 1.0

	scale, scaleDec := k.mustGetTokenScaleFactors(ctx, denom)
	if scale.IsZero() {
		scale = math.NewInt(1)
		scaleDec = math.LegacyOneDec()
	}

	maxSupplyTokens := math.LegacyNewDecFromInt(math.NewInt(30_000_000))
	currentSupplyTokens := math.LegacyNewDecFromInt(currentSupply).Quo(scaleDec)

	if currentSupplyTokens.GTE(maxSupplyTokens) {
		return maxPrice
	}

	progress := currentSupplyTokens.Quo(maxSupplyTokens)
	priceRange := maxPrice.Sub(minPrice)
	return minPrice.Add(progress.Mul(priceRange))
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

	// Ensure metadata exists (default to 6 decimals when unknown)
	k.ensureTokenMetadata(ctx, userToken, 6)
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

func buildTokenMetadata(token types.UserToken, decimals uint32) banktypes.Metadata {
	trimmedSymbol := strings.TrimSpace(token.Symbol)
	trimmedName := strings.TrimSpace(token.Name)

	displayDenom := strings.ToLower(trimmedSymbol)
	if displayDenom == "" {
		displayDenom = token.Denom
	}

	if decimals == 0 {
		decimals = 6
	}

	denomUnits := []*banktypes.DenomUnit{{
		Denom:    token.Denom,
		Exponent: 0,
	}}
	if displayDenom != token.Denom {
		aliases := []string{}
		if trimmedSymbol != "" {
			aliases = append(aliases, trimmedSymbol)
		}
		upperSymbol := strings.ToUpper(trimmedSymbol)
		if upperSymbol != "" && upperSymbol != trimmedSymbol {
			aliases = append(aliases, upperSymbol)
		}

		denomUnits = append(denomUnits, &banktypes.DenomUnit{
			Denom:    displayDenom,
			Exponent: decimals,
			Aliases:  aliases,
		})
	}

	return banktypes.Metadata{
		Description: fmt.Sprintf("%s user token", trimmedName),
		DenomUnits:  denomUnits,
		Base:        token.Denom,
		Display:     displayDenom,
		Name:        trimmedName,
		Symbol:      strings.ToUpper(trimmedSymbol),
	}
}

func (k Keeper) ensureTokenMetadata(ctx sdk.Context, token types.UserToken, decimals uint32) {
	metadata, found := k.bankKeeper.GetDenomMetaData(ctx, token.Denom)
	if found && metadata.Base == token.Denom && len(metadata.DenomUnits) > 0 {
		return
	}

	derived := buildTokenMetadata(token, decimals)
	k.bankKeeper.SetDenomMetaData(ctx, derived)
}

// GetFounderTokensRemaining returns the remaining founder tokens that can be claimed
func (k Keeper) GetFounderTokensRemaining(ctx sdk.Context, denom string) math.Int {
	userToken, found := k.GetUserToken(ctx, denom)
	if !found {
		return math.ZeroInt()
	}

	params := k.GetParams(ctx)
	scaleInt, _ := k.mustGetTokenScaleFactors(ctx, denom)
	founderTrancheAmount := params.FounderTrancheAmount.Mul(scaleInt)

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
func (k Keeper) DistributeTokenPurchasePayment(ctx sdk.Context, denom string, totalPayment math.Int, paymentDenom string) error {
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
			creatorCoin := sdk.NewCoin(paymentDenom, creatorAmount)
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, sdk.NewCoins(creatorCoin)); err != nil {
				return err
			}
		}
	}

	// 10% to referral wallet
	if referralAmount.IsPositive() && params.ReferralWallet != "" {
		referralAddr, err := sdk.AccAddressFromBech32(params.ReferralWallet)
		if err == nil {
			referralCoin := sdk.NewCoin(paymentDenom, referralAmount)
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, referralAddr, sdk.NewCoins(referralCoin)); err != nil {
				return err
			}
		}
	}

	// 40% to AI CEO wallet
	if aiCeoAmount.IsPositive() && params.AiCeoWallet != "" {
		aiCeoAddr, err := sdk.AccAddressFromBech32(params.AiCeoWallet)
		if err == nil {
			aiCeoCoin := sdk.NewCoin(paymentDenom, aiCeoAmount)
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, aiCeoAddr, sdk.NewCoins(aiCeoCoin)); err != nil {
				return err
			}
		}
	}

	// 10% to platform fee wallet
	if platformFeeAmount.IsPositive() && params.PlatformFeeWallet != "" {
		platformFeeAddr, err := sdk.AccAddressFromBech32(params.PlatformFeeWallet)
		if err == nil {
			platformFeeCoin := sdk.NewCoin(paymentDenom, platformFeeAmount)
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, platformFeeAddr, sdk.NewCoins(platformFeeCoin)); err != nil {
				return err
			}
		}
	}

	return nil
}

// CalculateTokensFromPayment converts a payment (in base currency units) into newly minted tokens.
// The function assumes 30% of the payment funds the bonding curve reserve and integrates the
// linear price curve P(s) = P0 + k*s, where k = (P1 - P0) / Smax. All supply math is handled in
// whole token units (no reliance on denom metadata).
func (k Keeper) CalculateTokensFromPayment(ctx sdk.Context, denom string, currentSupply math.Int, paymentAmount math.Int) math.Int {
	logger := k.Logger(ctx)
	if !paymentAmount.IsPositive() {
		logger.Debug("bonding curve buy calc skipped", "denom", denom, "payment_total", paymentAmount.String(), "reason", "non-positive payment")
		return math.ZeroInt()
	}

	scaleInt, scaleDec := k.mustGetTokenScaleFactors(ctx, denom)
	params := k.GetParams(ctx)
	startPrice := params.BondingCurveStartPrice
	endPrice := params.BondingCurveEndPrice
	maxSupplyTokens := math.LegacyNewDecFromInt(params.BondingCurveMaxSupply)

	if maxSupplyTokens.IsZero() {
		logger.Debug("bonding curve buy calc skipped", "denom", denom, "reason", "max supply is zero")
		return math.ZeroInt()
	}

	if currentSupply.IsNegative() {
		currentSupply = math.ZeroInt()
	}

	currentSupplyTokens := math.LegacyNewDecFromInt(currentSupply).Quo(scaleDec)
	if currentSupplyTokens.GTE(maxSupplyTokens) {
		logger.Debug("bonding curve buy calc skipped", "denom", denom, "s_before", currentSupply.String(), "reason", "curve exhausted")
		return math.ZeroInt()
	}

	remainingCapacityTokens := maxSupplyTokens.Sub(currentSupplyTokens)
	if remainingCapacityTokens.IsNegative() || remainingCapacityTokens.IsZero() {
		return math.ZeroInt()
	}

	// Calculate tokens based on full payment, but only 30% actually goes to curve reserve
	// The distribution happens separately in msg_server.go
	bondingCurvePayment := paymentAmount
	if !bondingCurvePayment.IsPositive() {
		logger.Debug("bonding curve buy calc skipped", "denom", denom, "payment_total", paymentAmount.String(), "payment_to_curve", bondingCurvePayment.String(), "reason", "insufficient payment for curve share")
		return math.ZeroInt()
	}

	// Convert payment from micro-NUAH to NUAH (divide by 1,000,000) to match price units
	nuahScale := math.LegacyNewDecWithPrec(1, 6) // 0.000001 = 1/1,000,000
	paymentDec := math.LegacyNewDecFromInt(bondingCurvePayment).Mul(nuahScale)
	priceRange := endPrice.Sub(startPrice)
	linearCoeff := math.LegacyZeroDec()
	if !priceRange.IsZero() {
		linearCoeff = priceRange.Quo(maxSupplyTokens)
	}

	tokensDec := math.LegacyZeroDec()
	if linearCoeff.IsZero() {
		if startPrice.IsZero() {
			logger.Debug("bonding curve buy calc skipped", "denom", denom, "reason", "zero price with zero slope")
			return math.ZeroInt()
		}
		tokensDec = paymentDec.Quo(startPrice)
	} else {
		half := math.LegacyNewDecWithPrec(5, 1)
		a := linearCoeff.Mul(half)
		b := startPrice.Add(linearCoeff.Mul(currentSupplyTokens))

		discriminant := b.Mul(b).Add(a.Mul(paymentDec).MulInt64(4))
		sqrtDiscriminant, err := discriminant.ApproxSqrt()
		if err != nil || sqrtDiscriminant.LT(b) {
			tokensDec = paymentDec.Quo(b)
		} else {
			numerator := sqrtDiscriminant.Sub(b)
			denominator := a.MulInt64(2)
			if !denominator.IsZero() {
				tokensDec = numerator.Quo(denominator)
			}
		}
	}

	if tokensDec.IsNegative() {
		tokensDec = math.LegacyZeroDec()
	}
	if tokensDec.GT(remainingCapacityTokens) {
		tokensDec = remainingCapacityTokens
	}

	maxSupplyBase := params.BondingCurveMaxSupply.Mul(scaleInt)
	remainingCapacityBase := maxSupplyBase.Sub(currentSupply)
	if remainingCapacityBase.IsNegative() || remainingCapacityBase.IsZero() {
		return math.ZeroInt()
	}

	tokensBaseDec := tokensDec.Mul(scaleDec)
	tokens := tokensBaseDec.TruncateInt()
	if tokens.IsZero() && tokensDec.IsPositive() {
		tokens = math.OneInt()
	}
	if tokens.GT(remainingCapacityBase) {
		tokens = remainingCapacityBase
	}

	tokensDecFinal := math.LegacyNewDecFromInt(tokens).Quo(scaleDec)
	costDec := math.LegacyZeroDec()
	if tokensDecFinal.IsPositive() {
		if linearCoeff.IsZero() {
			costDec = startPrice.Mul(tokensDecFinal)
		} else {
			costDec = k.priceIntegralDelta(startPrice, linearCoeff, currentSupplyTokens, tokensDecFinal)
		}
	}

	for costDec.GT(paymentDec) && tokens.GT(math.ZeroInt()) {
		tokens = tokens.Sub(math.OneInt())
		if tokens.IsNegative() || tokens.IsZero() {
			tokens = math.ZeroInt()
			costDec = math.LegacyZeroDec()
			break
		}
		tokensDecFinal = math.LegacyNewDecFromInt(tokens).Quo(scaleDec)
		if tokensDecFinal.IsPositive() {
			if linearCoeff.IsZero() {
				costDec = startPrice.Mul(tokensDecFinal)
			} else {
				costDec = k.priceIntegralDelta(startPrice, linearCoeff, currentSupplyTokens, tokensDecFinal)
			}
		} else {
			costDec = math.LegacyZeroDec()
		}
	}

	remainingAfterBase := maxSupplyBase.Sub(currentSupply.Add(tokens))
	if remainingAfterBase.IsNegative() {
		remainingAfterBase = math.ZeroInt()
	}

	avgPriceStr := "0"
	if tokensDecFinal.IsPositive() {
		avgPrice := costDec
		if !tokensDecFinal.IsZero() {
			avgPrice = costDec.Quo(tokensDecFinal)
		}
		avgPriceStr = avgPrice.String()
	}

	logger.Debug(
		"bonding curve buy calc",
		"denom", denom,
		"s_before", currentSupply.String(),
		"s_after", currentSupply.Add(tokens).String(),
		"s_max", maxSupplyBase.String(),
		"s_remaining_before", remainingCapacityBase.String(),
		"s_remaining_after", remainingAfterBase.String(),
		"delta_tokens", tokens.String(),
		"payment_total", paymentAmount.String(),
		"payment_to_curve", bondingCurvePayment.String(),
		"cost_to_curve", costDec.String(),
		"avg_price", avgPriceStr,
	)

	return tokens
}

// CalculatePayoutFromTokens determines the base currency payout for selling tokens back to the curve.
// It integrates the same linear price function while clamping supply to the valid bonding curve range.
func (k Keeper) CalculatePayoutFromTokens(ctx sdk.Context, denom string, currentSupply math.Int, tokensToSell math.Int) math.Int {
	logger := k.Logger(ctx)
	if !tokensToSell.IsPositive() {
		logger.Debug("bonding curve sell calc skipped", "denom", denom, "delta_tokens", tokensToSell.String(), "reason", "non-positive amount")
		return math.ZeroInt()
	}

	scaleInt, scaleDec := k.mustGetTokenScaleFactors(ctx, denom)
	params := k.GetParams(ctx)
	startPrice := params.BondingCurveStartPrice
	endPrice := params.BondingCurveEndPrice
	maxSupplyBase := params.BondingCurveMaxSupply.Mul(scaleInt)

	if currentSupply.IsNegative() {
		currentSupply = math.ZeroInt()
	}
	if currentSupply.GT(maxSupplyBase) {
		currentSupply = maxSupplyBase
	}

	if tokensToSell.GT(currentSupply) {
		tokensToSell = currentSupply
	}
	if tokensToSell.IsZero() {
		logger.Debug("bonding curve sell calc skipped", "denom", denom, "delta_tokens", tokensToSell.String(), "reason", "nothing available on curve")
		return math.ZeroInt()
	}

	currentSupplyTokens := math.LegacyNewDecFromInt(currentSupply).Quo(scaleDec)
	tokensToSellTokens := math.LegacyNewDecFromInt(tokensToSell).Quo(scaleDec)
	if tokensToSellTokens.GT(currentSupplyTokens) {
		tokensToSellTokens = currentSupplyTokens
	}

	priceRange := endPrice.Sub(startPrice)
	linearCoeff := math.LegacyZeroDec()
	maxSupplyTokens := math.LegacyNewDecFromInt(params.BondingCurveMaxSupply)
	if !priceRange.IsZero() {
		linearCoeff = priceRange.Quo(maxSupplyTokens)
	}

	startingSupplyTokens := currentSupplyTokens.Sub(tokensToSellTokens)
	if startingSupplyTokens.IsNegative() {
		startingSupplyTokens = math.LegacyZeroDec()
	}

	payoutDec := math.LegacyZeroDec()
	if tokensToSellTokens.IsPositive() {
		if linearCoeff.IsZero() {
			payoutDec = startPrice.Mul(tokensToSellTokens)
		} else {
			payoutDec = k.priceIntegralDelta(startPrice, linearCoeff, startingSupplyTokens, tokensToSellTokens)
		}
	}
	if payoutDec.IsNegative() {
		payoutDec = math.LegacyZeroDec()
	}

	// Convert payout from NUAH back to micro-NUAH (multiply by 1,000,000)
	nuahToMicroScale := math.LegacyNewDecFromInt(math.NewInt(1_000_000))
	payout := payoutDec.Mul(nuahToMicroScale).TruncateInt()
	avgPriceStr := "0"
	if tokensToSellTokens.IsPositive() {
		avgPrice := payoutDec
		if !tokensToSellTokens.IsZero() {
			avgPrice = payoutDec.Quo(tokensToSellTokens)
		}
		avgPriceStr = avgPrice.String()
	}

	sRemainingBefore := currentSupply
	sRemainingAfter := currentSupply.Sub(tokensToSell)
	if sRemainingAfter.IsNegative() {
		sRemainingAfter = math.ZeroInt()
	}

	logger.Debug(
		"bonding curve sell calc",
		"denom", denom,
		"s_before", currentSupply.String(),
		"s_after", currentSupply.Sub(tokensToSell).String(),
		"s_max", maxSupplyBase.String(),
		"s_remaining_before", sRemainingBefore.String(),
		"s_remaining_after", sRemainingAfter.String(),
		"delta_tokens", tokensToSell.String(),
		"payout_from_curve", payout.String(),
		"payout_curve_dec", payoutDec.String(),
		"avg_price", avgPriceStr,
	)

	return payout
}

// priceIntegralDelta computes the integral of the linear price function over [s, s+delta]
// using token units (expressed as decimals) and returns the area in base currency units.
func (k Keeper) priceIntegralDelta(startPrice, linearCoeff, supplyStart, delta math.LegacyDec) math.LegacyDec {
	if delta.IsNegative() || delta.IsZero() {
		return math.LegacyZeroDec()
	}

	supplyEnd := supplyStart.Add(delta)
	half := math.LegacyNewDecWithPrec(5, 1)

	endIntegral := startPrice.Mul(supplyEnd).Add(linearCoeff.Mul(supplyEnd).Mul(supplyEnd).Mul(half))
	startIntegral := startPrice.Mul(supplyStart).Add(linearCoeff.Mul(supplyStart).Mul(supplyStart).Mul(half))

	return endIntegral.Sub(startIntegral)
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// Convert slices of pointers to slices of values
	userTokens := make([]types.UserToken, 0)
	for _, token := range k.GetAllUserTokens(ctx) {
		userTokens = append(userTokens, *token)
	}

	referralPrograms := make([]types.ReferralProgram, 0)
	for _, program := range k.GetAllReferralPrograms(ctx) {
		referralPrograms = append(referralPrograms, *program)
	}

	referralActivations := make([]types.ReferralActivation, 0)
	for _, activation := range k.GetAllReferralActivations(ctx) {
		referralActivations = append(referralActivations, *activation)
	}

	userReferralQuotas := make([]types.UserReferralQuota, 0)
	for _, quota := range k.GetAllUserReferralQuotas(ctx) {
		userReferralQuotas = append(userReferralQuotas, *quota)
	}

	return &types.GenesisState{
		Params:              k.GetParams(ctx),
		UserTokens:          userTokens,
		ReferralPrograms:    referralPrograms,
		ReferralActivations: referralActivations,
		UserReferralQuotas:  userReferralQuotas,
	}
}

// Referral Program functions

// GetUserReferralQuota retrieves a user's referral quota
func (k Keeper) GetUserReferralQuota(ctx sdk.Context, user string) (types.UserReferralQuota, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.UserReferralQuotaKey(user))
	if bz == nil {
		return types.UserReferralQuota{}, false
	}

	var quota types.UserReferralQuota
	k.cdc.MustUnmarshal(bz, &quota)
	return quota, true
}

// SetUserReferralQuota stores a user's referral quota
func (k Keeper) SetUserReferralQuota(ctx sdk.Context, quota types.UserReferralQuota) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&quota)
	store.Set(types.UserReferralQuotaKey(quota.User), bz)
}

// GetAllUserReferralQuotas returns all user referral quotas
func (k Keeper) GetAllUserReferralQuotas(ctx sdk.Context) []*types.UserReferralQuota {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.UserReferralQuotaKeyPrefix)
	defer iterator.Close()

	var quotas []*types.UserReferralQuota
	for ; iterator.Valid(); iterator.Next() {
		var quota types.UserReferralQuota
		k.cdc.MustUnmarshal(iterator.Value(), &quota)
		quotas = append(quotas, &quota)
	}

	return quotas
}

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

	// Process user referral quotas - expand only if fully utilized
	quotas := k.GetAllUserReferralQuotas(ctx)
	for _, quota := range quotas {
		// If user used all available slots, expand by 3
		if quota.UsedSlots >= quota.TotalSlots {
			quota.TotalSlots += 3
		}
		// Don't reset used slots - they carry over to show actual usage
		quota.LastResetTime = ctx.BlockTime().Unix()
		k.SetUserReferralQuota(ctx, *quota)
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

// findPaymentCurrency finds a suitable payment currency for the user
// It looks for NDollar tokens first, then falls back to unuah
func (k Keeper) findPaymentCurrency(ctx sdk.Context, userAddr sdk.AccAddress, requiredAmount math.Int) (string, error) {
	// Get all spendable balances for the user
	allBalances := k.bankKeeper.SpendableCoins(ctx, userAddr)

	// Look for NDollar tokens (factory denoms with "ndollar" subdenom)
	for _, balance := range allBalances {
		if strings.Contains(balance.Denom, "factory/") && strings.HasSuffix(balance.Denom, "/ndollar") {
			if balance.Amount.GTE(requiredAmount) {
				return balance.Denom, nil
			}
		}
	}

	// Fall back to unuah if no NDollar found or insufficient balance
	unuahBalance := k.bankKeeper.GetBalance(ctx, userAddr, "unuah")
	if unuahBalance.Amount.GTE(requiredAmount) {
		return "unuah", nil
	}

	// If neither currency has sufficient balance, return error with details
	ndollarBalances := []string{}
	for _, balance := range allBalances {
		if strings.Contains(balance.Denom, "factory/") && strings.HasSuffix(balance.Denom, "/ndollar") {
			ndollarBalances = append(ndollarBalances, fmt.Sprintf("%s: %s", balance.Denom, balance.Amount.String()))
		}
	}

	if len(ndollarBalances) > 0 {
		return "", fmt.Errorf("insufficient balance: need %s, have NDollar balances: %s, unuah balance: %s",
			requiredAmount.String(), strings.Join(ndollarBalances, ", "), unuahBalance.Amount.String())
	}

	return "", fmt.Errorf("insufficient balance: need %s, have unuah balance: %s (no NDollar tokens found)",
		requiredAmount.String(), unuahBalance.Amount.String())
}

// findPayoutCurrency finds a suitable currency for payout from module reserves
// It looks for NDollar tokens first, then falls back to unuah
func (k Keeper) findPayoutCurrency(ctx sdk.Context, requiredAmount math.Int) (string, error) {
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if moduleAddr == nil {
		return "", fmt.Errorf("module address not found")
	}

	// Get all spendable balances for the module
	allBalances := k.bankKeeper.SpendableCoins(ctx, moduleAddr)

	// Look for NDollar tokens (factory denoms with "ndollar" subdenom)
	for _, balance := range allBalances {
		if strings.Contains(balance.Denom, "factory/") && strings.HasSuffix(balance.Denom, "/ndollar") {
			if balance.Amount.GTE(requiredAmount) {
				return balance.Denom, nil
			}
		}
	}

	// Fall back to unuah if no NDollar found or insufficient balance
	unuahBalance := k.bankKeeper.GetBalance(ctx, moduleAddr, "unuah")
	if unuahBalance.Amount.GTE(requiredAmount) {
		return "unuah", nil
	}

	// If neither currency has sufficient balance, return error with details
	ndollarBalances := []string{}
	for _, balance := range allBalances {
		if strings.Contains(balance.Denom, "factory/") && strings.HasSuffix(balance.Denom, "/ndollar") {
			ndollarBalances = append(ndollarBalances, fmt.Sprintf("%s: %s", balance.Denom, balance.Amount.String()))
		}
	}

	if len(ndollarBalances) > 0 {
		return "", fmt.Errorf("insufficient module balance for payout: need %s, have NDollar balances: %s, unuah balance: %s",
			requiredAmount.String(), strings.Join(ndollarBalances, ", "), unuahBalance.Amount.String())
	}

	return "", fmt.Errorf("insufficient module balance for payout: need %s, have unuah balance: %s (no NDollar tokens found)",
		requiredAmount.String(), unuahBalance.Amount.String())
}

// findPreferredBaseCurrency finds the preferred base currency for pools and trading
// It looks for NDollar tokens first, then falls back to unuah
func (k Keeper) findPreferredBaseCurrency(ctx sdk.Context) string {
	// Check module balance for available currencies
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if moduleAddr == nil {
		return "unuah" // Fallback if module address not found
	}

	// Get all spendable balances for the module
	allBalances := k.bankKeeper.SpendableCoins(ctx, moduleAddr)

	// Look for NDollar tokens (factory denoms with "ndollar" subdenom)
	for _, balance := range allBalances {
		if strings.Contains(balance.Denom, "factory/") && strings.HasSuffix(balance.Denom, "/ndollar") {
			// Found NDollar, use it as preferred base currency
			return balance.Denom
		}
	}

	// No NDollar found, fall back to unuah
	return "unuah"
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

	// Ensure bank metadata exists for all user tokens
	for _, token := range k.GetAllUserTokens(ctx) {
		k.ensureTokenMetadata(ctx, *token, 6)
	}
}

// GetTokenPrice returns the current price of a token from bonding curve
func (k Keeper) GetTokenPrice(ctx sdk.Context, denom string) (math.LegacyDec, error) {
	// Get current supply on bonding curve
	supply, err := k.GetBondingCurveSupply(ctx, denom)
	if err != nil {
		return math.LegacyZeroDec(), err
	}

	// Calculate price using bonding curve formula: price = supply / scale
	scale, _ := k.mustGetTokenScaleFactors(ctx, denom)
	if supply.IsZero() {
		// If supply is zero, return base price
		return math.LegacyNewDecFromInt(math.NewInt(200)), nil // 0.0002 unuah = 200 unuah
	}

	// Linear bonding curve: price = base_price + (supply / max_supply) × (max_price - base_price)
	// Price ranges from 0.0002 to 1.0 unuah for 30M tokens
	basePrice := math.LegacyNewDecFromInt(math.NewInt(200))    // 0.0002 unuah = 200 unuah
	maxPrice := math.LegacyNewDecFromInt(math.NewInt(1000000)) // 1.0 unuah = 1,000,000 unuah
	maxSupply := math.LegacyNewDecFromInt(math.NewInt(30_000_000).Mul(scale))

	supplyDec := math.LegacyNewDecFromInt(supply)
	if supplyDec.GTE(maxSupply) {
		return maxPrice, nil
	}

	progress := supplyDec.Quo(maxSupply)
	priceRange := maxPrice.Sub(basePrice)
	price := basePrice.Add(progress.Mul(priceRange))
	return price, nil
}

// CheckTokenExists verifies if a user token exists
func (k Keeper) CheckTokenExists(ctx sdk.Context, denom string) bool {
	// Check if it's a factory token
	if !strings.Contains(denom, "factory/") {
		return false
	}

	// Check if token exists in our store
	_, found := k.GetUserToken(ctx, denom)
	return found
}

// ExecuteBuyTokens executes a token purchase through bonding curve
func (k Keeper) ExecuteBuyTokens(ctx sdk.Context, buyer sdk.AccAddress, denom string, paymentAmount math.Int, paymentDenom string) (math.Int, error) {
	// Get current supply
	currentSupply, err := k.GetBondingCurveSupply(ctx, denom)
	if err != nil {
		return math.ZeroInt(), err
	}

	// Calculate tokens to receive
	tokensToReceive := k.CalculateTokensFromPayment(ctx, denom, currentSupply, paymentAmount)
	if tokensToReceive.IsZero() {
		return math.ZeroInt(), fmt.Errorf("calculated zero tokens for payment amount %s", paymentAmount.String())
	}

	// Transfer payment from buyer to module
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	paymentCoin := sdk.NewCoin(paymentDenom, paymentAmount)
	err = k.bankKeeper.SendCoins(ctx, buyer, moduleAddr, sdk.NewCoins(paymentCoin))
	if err != nil {
		return math.ZeroInt(), fmt.Errorf("failed to transfer payment: %w", err)
	}

	// Mint tokens to buyer
	tokenCoin := sdk.NewCoin(denom, tokensToReceive)
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(tokenCoin))
	if err != nil {
		return math.ZeroInt(), fmt.Errorf("failed to mint tokens: %w", err)
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, buyer, sdk.NewCoins(tokenCoin))
	if err != nil {
		return math.ZeroInt(), fmt.Errorf("failed to send tokens to buyer: %w", err)
	}

	return tokensToReceive, nil
}

// ExecuteSellTokens executes a token sale through bonding curve
func (k Keeper) ExecuteSellTokens(ctx sdk.Context, seller sdk.AccAddress, denom string, tokenAmount math.Int) (math.Int, string, error) {
	// Get user token to get current supply
	userToken, found := k.GetUserToken(ctx, denom)
	if !found {
		return math.ZeroInt(), "", fmt.Errorf("token not found: %s", denom)
	}

	// Get bonding curve supply for validation
	bondingCurveSupply, err := k.GetBondingCurveSupply(ctx, denom)
	if err != nil {
		return math.ZeroInt(), "", err
	}

	// Validate that we're not trying to sell more than available on bonding curve
	if tokenAmount.GT(bondingCurveSupply) {
		return math.ZeroInt(), "", fmt.Errorf("trying to sell %s tokens, but only %s available on bonding curve", tokenAmount.String(), bondingCurveSupply.String())
	}

	// Calculate payout using the total current supply (not just bonding curve supply)
	payout := k.CalculatePayoutFromTokens(ctx, denom, userToken.CurrentSupply, tokenAmount)
	if payout.IsZero() {
		return math.ZeroInt(), "", fmt.Errorf("calculated zero payout for token amount %s", tokenAmount.String())
	}

	// Determine payout currency (prefer NDollar, fallback to unuah)
	payoutDenom, err := k.findPayoutCurrency(ctx, payout)
	if err != nil {
		// If module doesn't have funds, use unuah as fallback
		payoutDenom = "unuah"
	}

	// Burn tokens from seller
	tokenCoin := sdk.NewCoin(denom, tokenAmount)
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, seller, types.ModuleName, sdk.NewCoins(tokenCoin))
	if err != nil {
		return math.ZeroInt(), "", fmt.Errorf("failed to transfer tokens from seller: %w", err)
	}
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(tokenCoin))
	if err != nil {
		return math.ZeroInt(), "", fmt.Errorf("failed to burn tokens: %w", err)
	}

	// Transfer payout from module to seller
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	payoutCoin := sdk.NewCoin(payoutDenom, payout)
	err = k.bankKeeper.SendCoins(ctx, moduleAddr, seller, sdk.NewCoins(payoutCoin))
	if err != nil {
		return math.ZeroInt(), "", fmt.Errorf("failed to transfer payout: %w", err)
	}

	return payout, payoutDenom, nil
}

package keeper

import (
	"fmt"
	"strconv"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/leverage/types"
)

// LendingKeeper handles all lending operations
type LendingKeeper struct {
	keeper *Keeper
}

// NewLendingKeeper creates a new lending keeper
func NewLendingKeeper(k *Keeper) LendingKeeper {
	return LendingKeeper{keeper: k}
}

// =================== LENDING POOL MANAGEMENT ===================

// CreateLendingPool creates a new lending pool for a token
func (lk LendingKeeper) CreateLendingPool(ctx sdk.Context, denom string) error {
	// Check if pool already exists
	if _, found := lk.GetLendingPool(ctx, denom); found {
		return fmt.Errorf("lending pool for %s already exists", denom)
	}

	// Create new pool
	pool := types.LendingPool{
		Denom:              denom,
		TotalSupply:        math.ZeroInt(),
		TotalBorrowed:      math.ZeroInt(),
		AvailableLiquidity: math.ZeroInt(),
		InterestRate:       math.LegacyZeroDec(),
		UtilizationRate:    math.LegacyZeroDec(),
		LastUpdateTime:     ctx.BlockTime().Unix(),
	}

	lk.SetLendingPool(ctx, pool)
	return nil
}

// InitializePoolWithLiquidity initializes a lending pool with initial liquidity from module reserves
func (lk LendingKeeper) InitializePoolWithLiquidity(ctx sdk.Context, denom string) error {
	// Get module address
	moduleAddr := lk.keeper.accountKeeper.GetModuleAddress(types.ModuleName)
	if moduleAddr == nil {
		return fmt.Errorf("module address not found")
	}

	// Get bonding curve supply from usertoken module
	bondingCurveSupply, err := lk.keeper.userTokenKeeper.GetBondingCurveSupply(ctx, denom)
	if err != nil {
		return fmt.Errorf("failed to get bonding curve supply: %w", err)
	}

	// If no bonding curve supply, create empty pool
	if bondingCurveSupply.IsZero() {
		return nil
	}

	// Get the pool
	pool, found := lk.GetLendingPool(ctx, denom)
	if !found {
		return fmt.Errorf("pool not found after creation")
	}

	// Use bonding curve supply as available liquidity
	// This represents tokens that can be borrowed for SHORT positions
	pool.AvailableLiquidity = pool.AvailableLiquidity.Add(bondingCurveSupply)
	pool.LastUpdateTime = ctx.BlockTime().Unix()

	lk.SetLendingPool(ctx, pool)

	return nil
}

// GetLendingPool retrieves a lending pool by denom
func (lk LendingKeeper) GetLendingPool(ctx sdk.Context, denom string) (types.LendingPool, bool) {
	store := ctx.KVStore(lk.keeper.storeKey)
	bz := store.Get(types.LendingPoolKey(denom))
	if bz == nil {
		return types.LendingPool{}, false
	}

	var pool types.LendingPool
	lk.keeper.cdc.MustUnmarshal(bz, &pool)
	return pool, true
}

// SetLendingPool stores a lending pool
func (lk LendingKeeper) SetLendingPool(ctx sdk.Context, pool types.LendingPool) {
	store := ctx.KVStore(lk.keeper.storeKey)
	bz := lk.keeper.cdc.MustMarshal(&pool)
	store.Set(types.LendingPoolKey(pool.Denom), bz)
}

// GetAllLendingPools returns all lending pools
func (lk LendingKeeper) GetAllLendingPools(ctx sdk.Context) []types.LendingPool {
	store := ctx.KVStore(lk.keeper.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.LendingPoolKeyPrefix)
	defer iterator.Close()

	var pools []types.LendingPool
	for ; iterator.Valid(); iterator.Next() {
		var pool types.LendingPool
		lk.keeper.cdc.MustUnmarshal(iterator.Value(), &pool)
		pools = append(pools, pool)
	}

	return pools
}

// =================== LIQUIDITY PROVISION ===================

// ProvideLiquidity allows users to provide liquidity to a lending pool
func (lk LendingKeeper) ProvideLiquidity(ctx sdk.Context, provider sdk.AccAddress, amount sdk.Coin) error {
	// Get or create lending pool
	pool, found := lk.GetLendingPool(ctx, amount.Denom)
	if !found {
		if err := lk.CreateLendingPool(ctx, amount.Denom); err != nil {
			return err
		}
		pool, _ = lk.GetLendingPool(ctx, amount.Denom)
	}

	// Transfer tokens from provider to module
	err := lk.keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, provider, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		return fmt.Errorf("failed to transfer liquidity: %w", err)
	}

	// Calculate share tokens to mint
	var shareTokens math.Int
	if pool.TotalSupply.IsZero() {
		// First provider gets 1:1 ratio
		shareTokens = amount.Amount
	} else {
		// shareTokens = (amount * totalShareTokens) / totalSupply
		totalShareTokens := lk.GetTotalShareTokens(ctx, amount.Denom)
		shareTokens = amount.Amount.Mul(totalShareTokens).Quo(pool.TotalSupply)
	}

	// Update pool
	pool.TotalSupply = pool.TotalSupply.Add(amount.Amount)
	pool.AvailableLiquidity = pool.AvailableLiquidity.Add(amount.Amount)
	pool.UtilizationRate = pool.CalculateUtilizationRate()
	pool.LastUpdateTime = ctx.BlockTime().Unix()

	// Update interest rate based on new utilization
	lk.UpdateInterestRate(ctx, &pool)
	lk.SetLendingPool(ctx, pool)

	// Create or update liquidity provider record
	lp := types.LiquidityProvider{
		Provider:    provider.String(),
		TokenDenom:  amount.Denom,
		Amount:      amount.Amount,
		ShareTokens: shareTokens,
		ProvidedAt:  ctx.BlockTime().Unix(),
	}

	// Check if provider already exists
	if existingLP, found := lk.GetLiquidityProvider(ctx, provider.String(), amount.Denom); found {
		lp.Amount = existingLP.Amount.Add(amount.Amount)
		lp.ShareTokens = existingLP.ShareTokens.Add(shareTokens)
		lp.ProvidedAt = existingLP.ProvidedAt // Keep original time
	}

	lk.SetLiquidityProvider(ctx, lp)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"provide_liquidity",
			sdk.NewAttribute("provider", provider.String()),
			sdk.NewAttribute("denom", amount.Denom),
			sdk.NewAttribute("amount", amount.Amount.String()),
			sdk.NewAttribute("share_tokens", shareTokens.String()),
		),
	})

	return nil
}

// WithdrawLiquidity allows users to withdraw their provided liquidity
func (lk LendingKeeper) WithdrawLiquidity(ctx sdk.Context, provider sdk.AccAddress, denom string, shareTokens math.Int) error {
	// Get liquidity provider record
	lp, found := lk.GetLiquidityProvider(ctx, provider.String(), denom)
	if !found {
		return types.ErrLiquidityProviderNotFound
	}

	if lp.ShareTokens.LT(shareTokens) {
		return types.ErrInsufficientShares
	}

	// Get lending pool
	pool, found := lk.GetLendingPool(ctx, denom)
	if !found {
		return types.ErrLendingPoolNotFound
	}

	// Calculate withdrawal amount
	totalShareTokens := lk.GetTotalShareTokens(ctx, denom)
	withdrawAmount := shareTokens.Mul(pool.TotalSupply).Quo(totalShareTokens)

	// Check if there's enough available liquidity
	if pool.AvailableLiquidity.LT(withdrawAmount) {
		return types.ErrInsufficientLiquidity
	}

	// Transfer tokens from module to provider
	withdrawCoin := sdk.NewCoin(denom, withdrawAmount)
	err := lk.keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, provider, sdk.NewCoins(withdrawCoin))
	if err != nil {
		return fmt.Errorf("failed to withdraw liquidity: %w", err)
	}

	// Update pool
	pool.TotalSupply = pool.TotalSupply.Sub(withdrawAmount)
	pool.AvailableLiquidity = pool.AvailableLiquidity.Sub(withdrawAmount)
	pool.UtilizationRate = pool.CalculateUtilizationRate()
	pool.LastUpdateTime = ctx.BlockTime().Unix()

	lk.UpdateInterestRate(ctx, &pool)
	lk.SetLendingPool(ctx, pool)

	// Update liquidity provider record
	lp.Amount = lp.Amount.Sub(withdrawAmount)
	lp.ShareTokens = lp.ShareTokens.Sub(shareTokens)

	if lp.ShareTokens.IsZero() {
		lk.DeleteLiquidityProvider(ctx, provider.String(), denom)
	} else {
		lk.SetLiquidityProvider(ctx, lp)
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"withdraw_liquidity",
			sdk.NewAttribute("provider", provider.String()),
			sdk.NewAttribute("denom", denom),
			sdk.NewAttribute("amount", withdrawAmount.String()),
			sdk.NewAttribute("share_tokens", shareTokens.String()),
		),
	})

	return nil
}

// =================== BORROWING ===================

// BorrowTokens allows borrowing tokens from a lending pool
func (lk LendingKeeper) BorrowTokens(ctx sdk.Context, borrower sdk.AccAddress, denom string, amount math.Int, leveragePositionID string) (string, error) {
	// Get lending pool, create if doesn't exist
	pool, found := lk.GetLendingPool(ctx, denom)
	if !found {
		// Auto-create lending pool for supported tokens
		if err := lk.CreateLendingPool(ctx, denom); err != nil {
			return "", fmt.Errorf("failed to create lending pool: %w", err)
		}
		// Initialize pool with some liquidity from module reserves
		if err := lk.InitializePoolWithLiquidity(ctx, denom); err != nil {
			return "", fmt.Errorf("failed to initialize pool liquidity: %w", err)
		}
		pool, _ = lk.GetLendingPool(ctx, denom)
	}

	// Check if there's enough available liquidity
	if pool.AvailableLiquidity.LT(amount) {
		return "", types.ErrInsufficientLiquidity
	}

	// Generate borrow position ID
	borrowID := lk.GetNextBorrowID(ctx)

	// Create borrow position
	borrowPosition := types.BorrowPosition{
		Id:                 borrowID,
		Borrower:           borrower.String(),
		TokenDenom:         denom,
		BorrowedAmount:     amount,
		AccruedInterest:    math.ZeroInt(),
		InterestRate:       pool.InterestRate,
		CreatedAt:          ctx.BlockTime().Unix(),
		LastInterestTime:   ctx.BlockTime().Unix(),
		LeveragePositionId: leveragePositionID,
	}

	// For SHORT positions, we don't actually transfer tokens to borrower
	// Instead, we simulate borrowing by allowing the borrower to sell tokens from bonding curve
	// The actual token transfer happens in the SHORT position logic via ExecuteSellTokens
	// This is a conceptual borrow - the borrower gets the right to sell tokens from bonding curve

	// Update pool
	pool.TotalBorrowed = pool.TotalBorrowed.Add(amount)
	pool.AvailableLiquidity = pool.AvailableLiquidity.Sub(amount)
	pool.UtilizationRate = pool.CalculateUtilizationRate()
	pool.LastUpdateTime = ctx.BlockTime().Unix()

	lk.UpdateInterestRate(ctx, &pool)
	lk.SetLendingPool(ctx, pool)

	// Store borrow position
	lk.SetBorrowPosition(ctx, borrowPosition)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"borrow_tokens",
			sdk.NewAttribute("borrower", borrower.String()),
			sdk.NewAttribute("denom", denom),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("borrow_id", borrowID),
			sdk.NewAttribute("leverage_position_id", leveragePositionID),
		),
	})

	return borrowID, nil
}

// RepayTokens allows repaying borrowed tokens
func (lk LendingKeeper) RepayTokens(ctx sdk.Context, borrower sdk.AccAddress, borrowID string, amount math.Int) error {
	// Get borrow position
	borrowPos, found := lk.GetBorrowPosition(ctx, borrowID)
	if !found {
		return types.ErrBorrowPositionNotFound
	}

	// Check ownership
	if borrowPos.Borrower != borrower.String() {
		return types.ErrUnauthorized
	}

	// Update accrued interest
	lk.UpdateBorrowInterest(ctx, &borrowPos)

	// Calculate total debt
	totalDebt := borrowPos.GetTotalDebt()

	// Check if repayment amount is valid
	if amount.GT(totalDebt) {
		return types.ErrRepaymentExceedsDebt
	}

	// Transfer tokens from borrower to module
	repayCoin := sdk.NewCoin(borrowPos.TokenDenom, amount)
	err := lk.keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, borrower, types.ModuleName, sdk.NewCoins(repayCoin))
	if err != nil {
		return fmt.Errorf("failed to transfer repayment: %w", err)
	}

	// Update borrow position
	if amount.GTE(totalDebt) {
		// Full repayment - delete position
		lk.DeleteBorrowPosition(ctx, borrowPos)
		amount = totalDebt // Don't over-repay
	} else {
		// Partial repayment - update position
		// First pay off interest, then principal
		if amount.GTE(borrowPos.AccruedInterest) {
			// Pay all interest and some principal
			principalPayment := amount.Sub(borrowPos.AccruedInterest)
			borrowPos.AccruedInterest = math.ZeroInt()
			borrowPos.BorrowedAmount = borrowPos.BorrowedAmount.Sub(principalPayment)
		} else {
			// Only pay part of interest
			borrowPos.AccruedInterest = borrowPos.AccruedInterest.Sub(amount)
		}
		borrowPos.LastInterestTime = ctx.BlockTime().Unix()
		lk.SetBorrowPosition(ctx, borrowPos)
	}

	// Update lending pool
	pool, _ := lk.GetLendingPool(ctx, borrowPos.TokenDenom)
	pool.TotalBorrowed = pool.TotalBorrowed.Sub(amount)
	pool.AvailableLiquidity = pool.AvailableLiquidity.Add(amount)
	pool.UtilizationRate = pool.CalculateUtilizationRate()
	pool.LastUpdateTime = ctx.BlockTime().Unix()

	lk.UpdateInterestRate(ctx, &pool)
	lk.SetLendingPool(ctx, pool)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"repay_tokens",
			sdk.NewAttribute("borrower", borrower.String()),
			sdk.NewAttribute("borrow_id", borrowID),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("remaining_debt", borrowPos.GetTotalDebt().String()),
		),
	})

	return nil
}

// =================== INTEREST CALCULATION ===================

// UpdateInterestRate updates the interest rate for a lending pool
func (lk LendingKeeper) UpdateInterestRate(ctx sdk.Context, pool *types.LendingPool) {
	params := lk.keeper.GetParams(ctx)

	// Use a simple interest rate model
	// Base rate + (utilization rate * multiplier)
	baseRate := params.BaseInterestRate
	multiplier := params.InterestRateMultiplier

	newRate := baseRate.Add(pool.UtilizationRate.Mul(multiplier))

	// Cap the interest rate
	if newRate.GT(params.MaxInterestRate) {
		newRate = params.MaxInterestRate
	}

	pool.InterestRate = newRate
}

// UpdateBorrowInterest updates the accrued interest for a borrow position
func (lk LendingKeeper) UpdateBorrowInterest(ctx sdk.Context, borrowPos *types.BorrowPosition) {
	currentTime := ctx.BlockTime().Unix()
	timeDiff := currentTime - borrowPos.LastInterestTime

	if timeDiff <= 0 {
		return
	}

	// Calculate interest: interest = principal * rate * time
	// Rate is annual, so divide by seconds in a year
	secondsInYear := int64(365 * 24 * 3600)
	timeRatio := math.LegacyNewDec(timeDiff).Quo(math.LegacyNewDec(secondsInYear))

	interestAmount := math.LegacyNewDecFromInt(borrowPos.BorrowedAmount).
		Mul(borrowPos.InterestRate).
		Mul(timeRatio).
		TruncateInt()

	borrowPos.AccruedInterest = borrowPos.AccruedInterest.Add(interestAmount)
	borrowPos.LastInterestTime = currentTime
}

// UpdateAllBorrowInterests updates interest for all borrow positions
func (lk LendingKeeper) UpdateAllBorrowInterests(ctx sdk.Context) {
	borrowPositions := lk.GetAllBorrowPositions(ctx)

	for _, borrowPos := range borrowPositions {
		lk.UpdateBorrowInterest(ctx, &borrowPos)
		lk.SetBorrowPosition(ctx, borrowPos)
	}
}

// =================== HELPER FUNCTIONS ===================

// GetNextBorrowID returns the next borrow position ID
func (lk LendingKeeper) GetNextBorrowID(ctx sdk.Context) string {
	store := ctx.KVStore(lk.keeper.storeKey)
	bz := store.Get(types.NextBorrowIDKeyPrefix)

	var nextID uint64 = 1
	if bz != nil {
		nextID = sdk.BigEndianToUint64(bz)
	}

	// Increment and store the next ID
	store.Set(types.NextBorrowIDKeyPrefix, sdk.Uint64ToBigEndian(nextID+1))

	return "borrow_" + strconv.FormatUint(nextID, 10)
}

// GetTotalShareTokens calculates total share tokens for a denom
func (lk LendingKeeper) GetTotalShareTokens(ctx sdk.Context, denom string) math.Int {
	providers := lk.GetLiquidityProvidersByDenom(ctx, denom)
	total := math.ZeroInt()

	for _, provider := range providers {
		total = total.Add(provider.ShareTokens)
	}

	return total
}

// Storage functions for borrow positions
func (lk LendingKeeper) GetBorrowPosition(ctx sdk.Context, borrowID string) (types.BorrowPosition, bool) {
	store := ctx.KVStore(lk.keeper.storeKey)
	bz := store.Get(types.BorrowPositionKey(borrowID))
	if bz == nil {
		return types.BorrowPosition{}, false
	}

	var borrowPos types.BorrowPosition
	lk.keeper.cdc.MustUnmarshal(bz, &borrowPos)
	return borrowPos, true
}

func (lk LendingKeeper) SetBorrowPosition(ctx sdk.Context, borrowPos types.BorrowPosition) {
	store := ctx.KVStore(lk.keeper.storeKey)
	bz := lk.keeper.cdc.MustMarshal(&borrowPos)
	store.Set(types.BorrowPositionKey(borrowPos.Id), bz)

	// Also store by borrower for efficient queries
	store.Set(types.BorrowerPositionKey(borrowPos.Borrower, borrowPos.Id), []byte(borrowPos.Id))
}

func (lk LendingKeeper) DeleteBorrowPosition(ctx sdk.Context, borrowPos types.BorrowPosition) {
	store := ctx.KVStore(lk.keeper.storeKey)
	store.Delete(types.BorrowPositionKey(borrowPos.Id))
	store.Delete(types.BorrowerPositionKey(borrowPos.Borrower, borrowPos.Id))
}

func (lk LendingKeeper) GetAllBorrowPositions(ctx sdk.Context) []types.BorrowPosition {
	store := ctx.KVStore(lk.keeper.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.BorrowPositionKeyPrefix)
	defer iterator.Close()

	var positions []types.BorrowPosition
	for ; iterator.Valid(); iterator.Next() {
		var pos types.BorrowPosition
		lk.keeper.cdc.MustUnmarshal(iterator.Value(), &pos)
		positions = append(positions, pos)
	}

	return positions
}

// Storage functions for liquidity providers
func (lk LendingKeeper) GetLiquidityProvider(ctx sdk.Context, provider, denom string) (types.LiquidityProvider, bool) {
	store := ctx.KVStore(lk.keeper.storeKey)
	bz := store.Get(types.LiquidityProviderKey(provider, denom))
	if bz == nil {
		return types.LiquidityProvider{}, false
	}

	var lp types.LiquidityProvider
	lk.keeper.cdc.MustUnmarshal(bz, &lp)
	return lp, true
}

func (lk LendingKeeper) SetLiquidityProvider(ctx sdk.Context, lp types.LiquidityProvider) {
	store := ctx.KVStore(lk.keeper.storeKey)
	bz := lk.keeper.cdc.MustMarshal(&lp)
	store.Set(types.LiquidityProviderKey(lp.Provider, lp.TokenDenom), bz)
}

func (lk LendingKeeper) DeleteLiquidityProvider(ctx sdk.Context, provider, denom string) {
	store := ctx.KVStore(lk.keeper.storeKey)
	store.Delete(types.LiquidityProviderKey(provider, denom))
}

func (lk LendingKeeper) GetLiquidityProvidersByDenom(ctx sdk.Context, denom string) []types.LiquidityProvider {
	store := ctx.KVStore(lk.keeper.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.LiquidityProviderKeyPrefix)
	defer iterator.Close()

	var providers []types.LiquidityProvider
	for ; iterator.Valid(); iterator.Next() {
		var lp types.LiquidityProvider
		lk.keeper.cdc.MustUnmarshal(iterator.Value(), &lp)
		if lp.TokenDenom == denom {
			providers = append(providers, lp)
		}
	}

	return providers
}

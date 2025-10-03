package keeper

import (
	"fmt"
	"strconv"
	"strings"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/leverage/types"
)

type (
	Keeper struct {
		cdc       codec.BinaryCodec
		storeKey  storetypes.StoreKey
		memKey    storetypes.StoreKey
		authority string

		// External keepers
		accountKeeper   types.AccountKeeper
		bankKeeper      types.BankKeeper
		userTokenKeeper types.UserTokenKeeper

		// Internal keepers
		lendingKeeper LendingKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	authority string,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	userTokenKeeper types.UserTokenKeeper,
) *Keeper {
	k := &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		memKey:          memKey,
		authority:       authority,
		accountKeeper:   accountKeeper,
		bankKeeper:      bankKeeper,
		userTokenKeeper: userTokenKeeper,
	}

	// Initialize lending keeper
	k.lendingKeeper = NewLendingKeeper(k)

	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetLendingKeeper returns the lending keeper.
func (k Keeper) GetLendingKeeper() LendingKeeper {
	return k.lendingKeeper
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.LeverageParams {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKeyPrefix)
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.LeverageParams
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.LeverageParams) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKeyPrefix, bz)
}

// GetNextPositionID returns the next position ID and increments it
func (k Keeper) GetNextPositionID(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextPositionIDKeyPrefix)

	var nextID uint64 = 1
	if bz != nil {
		nextID = sdk.BigEndianToUint64(bz)
	}

	// Increment and store the next ID
	store.Set(types.NextPositionIDKeyPrefix, sdk.Uint64ToBigEndian(nextID+1))

	return strconv.FormatUint(nextID, 10)
}

// GetPosition retrieves a position by ID
func (k Keeper) GetPosition(ctx sdk.Context, positionID string) (types.Position, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PositionKey(positionID))
	if bz == nil {
		return types.Position{}, false
	}

	var position types.Position
	k.cdc.MustUnmarshal(bz, &position)
	return position, true
}

// SetPosition stores a position
func (k Keeper) SetPosition(ctx sdk.Context, position types.Position) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&position)
	store.Set(types.PositionKey(position.Id), bz)

	// Also store by trader for efficient queries
	store.Set(types.TraderPositionKey(position.Trader, position.Id), []byte(position.Id))

	// Store by token denom for efficient queries
	store.Set(types.TokenPositionKey(position.TokenDenom, position.Id), []byte(position.Id))
}

// DeletePosition removes a position from the store
func (k Keeper) DeletePosition(ctx sdk.Context, position types.Position) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.PositionKey(position.Id))
	store.Delete(types.TraderPositionKey(position.Trader, position.Id))
	store.Delete(types.TokenPositionKey(position.TokenDenom, position.Id))
}

// GetAllPositions returns all positions
func (k Keeper) GetAllPositions(ctx sdk.Context) []types.Position {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PositionKeyPrefix)
	defer iterator.Close()

	var positions []types.Position
	for ; iterator.Valid(); iterator.Next() {
		var position types.Position
		k.cdc.MustUnmarshal(iterator.Value(), &position)
		positions = append(positions, position)
	}

	return positions
}

// GetPositionsByTrader returns all positions for a specific trader
func (k Keeper) GetPositionsByTrader(ctx sdk.Context, trader string) []types.Position {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.TraderPositionPrefix(trader))
	defer iterator.Close()

	var positions []types.Position
	for ; iterator.Valid(); iterator.Next() {
		positionID := string(iterator.Value())
		if position, found := k.GetPosition(ctx, positionID); found {
			positions = append(positions, position)
		}
	}

	return positions
}

// GetPositionsByToken returns all positions for a specific token
func (k Keeper) GetPositionsByToken(ctx sdk.Context, tokenDenom string) []types.Position {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.TokenPositionPrefix(tokenDenom))
	defer iterator.Close()

	var positions []types.Position
	for ; iterator.Valid(); iterator.Next() {
		positionID := string(iterator.Value())
		if position, found := k.GetPosition(ctx, positionID); found {
			positions = append(positions, position)
		}
	}

	return positions
}

// CalculateLiquidationPrice calculates the liquidation price for a position
func (k Keeper) CalculateLiquidationPrice(ctx sdk.Context, collateral, positionSize math.Int, entryPrice math.LegacyDec, side types.PositionSide) math.LegacyDec {
	params := k.GetParams(ctx)

	// Convert amounts to decimal for calculation
	collateralDec := math.LegacyNewDecFromInt(collateral)
	positionSizeDec := math.LegacyNewDecFromInt(positionSize)

	// Calculate maintenance margin requirement
	maintenanceMarginAmount := positionSizeDec.Mul(entryPrice).Mul(params.MaintenanceMargin)

	// Calculate liquidation price based on position side
	if side == types.PositionSideLong {
		// For LONG: liquidationPrice = entryPrice - (collateral - maintenanceMargin) / positionSize
		numerator := collateralDec.Sub(maintenanceMarginAmount)
		if numerator.LTE(math.LegacyZeroDec()) {
			return math.LegacyZeroDec() // Position would be immediately liquidated
		}
		priceChange := numerator.Quo(positionSizeDec)
		return entryPrice.Sub(priceChange)
	} else {
		// For SHORT: liquidationPrice = entryPrice + (collateral - maintenanceMargin) / positionSize
		numerator := collateralDec.Sub(maintenanceMarginAmount)
		if numerator.LTE(math.LegacyZeroDec()) {
			return math.LegacyMaxDec(math.LegacyZeroDec(), math.LegacyZeroDec()) // Position would be immediately liquidated
		}
		priceChange := numerator.Quo(positionSizeDec)
		return entryPrice.Add(priceChange)
	}
}

// UpdatePositionPnL updates the unrealized PnL of a position based on current price
func (k Keeper) UpdatePositionPnL(ctx sdk.Context, position *types.Position, currentPrice math.LegacyDec) {
	positionSizeDec := math.LegacyNewDecFromInt(position.Size_)

	var pnlDec math.LegacyDec
	if position.Side == types.PositionSideLong {
		// For LONG: PnL = (currentPrice - entryPrice) * positionSize
		pnlDec = currentPrice.Sub(position.EntryPrice).Mul(positionSizeDec)
	} else {
		// For SHORT: PnL = (entryPrice - currentPrice) * positionSize
		pnlDec = position.EntryPrice.Sub(currentPrice).Mul(positionSizeDec)
	}

	position.UnrealizedPnl = pnlDec.TruncateInt()
	position.UpdatedAt = ctx.BlockTime().Unix()
}

// IsPositionLiquidatable checks if a position should be liquidated
func (k Keeper) IsPositionLiquidatable(ctx sdk.Context, position types.Position, currentPrice math.LegacyDec) bool {
	if position.Status != types.PositionStatusOpen {
		return false
	}

	if position.Side == types.PositionSideLong {
		return currentPrice.LTE(position.LiquidationPrice)
	} else {
		return currentPrice.GTE(position.LiquidationPrice)
	}
}

// ValidateCollateralDenom checks if a denomination can be used as collateral
func (k Keeper) ValidateCollateralDenom(ctx sdk.Context, denom string) bool {
	// Accept NDollar (factory denoms ending with /ndollar) and unuah
	if denom == "unuah" {
		return true
	}

	if strings.Contains(denom, "factory/") && strings.HasSuffix(denom, "/ndollar") {
		return true
	}

	return false
}

// ValidateTokenDenom checks if a token can be traded with leverage
func (k Keeper) ValidateTokenDenom(ctx sdk.Context, denom string) bool {
	// Only allow user tokens (factory denoms that are not ndollar)
	if !strings.Contains(denom, "factory/") {
		return false
	}

	if strings.HasSuffix(denom, "/ndollar") {
		return false
	}

	// Check if the token exists in usertoken module
	return k.userTokenKeeper.CheckTokenExists(ctx, denom)
}

// GetTokenPrice gets the current price of a token from usertoken module
func (k Keeper) GetTokenPrice(ctx sdk.Context, denom string) (math.LegacyDec, error) {
	return k.userTokenKeeper.GetTokenPrice(ctx, denom)
}

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// Set params
	k.SetParams(ctx, genState.Params)

	// Initialize positions
	for _, position := range genState.Positions {
		k.SetPosition(ctx, position)
	}

	// Initialize lending pools
	for _, pool := range genState.LendingPools {
		k.lendingKeeper.SetLendingPool(ctx, pool)
	}

	// Initialize borrow positions
	for _, borrowPos := range genState.BorrowPositions {
		k.lendingKeeper.SetBorrowPosition(ctx, borrowPos)
	}

	// Initialize liquidity providers
	for _, lp := range genState.LiquidityProviders {
		k.lendingKeeper.SetLiquidityProvider(ctx, lp)
	}

	// Set next position ID
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NextPositionIDKeyPrefix, sdk.Uint64ToBigEndian(genState.NextPositionId))

	// Set next borrow ID
	store.Set(types.NextBorrowIDKeyPrefix, sdk.Uint64ToBigEndian(genState.NextBorrowId))
}

// ExportGenesis returns the module's exported genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// Get next position ID
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextPositionIDKeyPrefix)
	var nextID uint64 = 1
	if bz != nil {
		nextID = sdk.BigEndianToUint64(bz)
	}

	// Get next borrow ID
	bzBorrow := store.Get(types.NextBorrowIDKeyPrefix)
	var nextBorrowID uint64 = 1
	if bzBorrow != nil {
		nextBorrowID = sdk.BigEndianToUint64(bzBorrow)
	}

	return &types.GenesisState{
		Params:             k.GetParams(ctx),
		Positions:          k.GetAllPositions(ctx),
		NextPositionId:     nextID,
		LendingPools:       k.lendingKeeper.GetAllLendingPools(ctx),
		BorrowPositions:    k.lendingKeeper.GetAllBorrowPositions(ctx),
		LiquidityProviders: k.GetAllLiquidityProviders(ctx),
		NextBorrowId:       nextBorrowID,
	}
}

// GetAllLiquidityProviders returns all liquidity providers
func (k Keeper) GetAllLiquidityProviders(ctx sdk.Context) []types.LiquidityProvider {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.LiquidityProviderKeyPrefix)
	defer iterator.Close()

	var providers []types.LiquidityProvider
	for ; iterator.Valid(); iterator.Next() {
		var lp types.LiquidityProvider
		k.cdc.MustUnmarshal(iterator.Value(), &lp)
		providers = append(providers, lp)
	}

	return providers
}

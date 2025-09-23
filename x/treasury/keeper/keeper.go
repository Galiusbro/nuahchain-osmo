package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/osmomath"

	rolestypes "github.com/osmosis-labs/osmosis/v30/x/roles/types"
	"github.com/osmosis-labs/osmosis/v30/x/treasury/types"
)

// Keeper manages treasury pools and balances.
type Keeper struct {
	cdc         codec.BinaryCodec
	storeKey    storetypes.StoreKey
	paramstore  paramtypes.Subspace
	bankKeeper  types.BankKeeper
	rolesKeeper types.RolesKeeper
	authority   string
}

// NewKeeper constructs a new treasury keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	ps paramtypes.Subspace,
	bank types.BankKeeper,
	roles types.RolesKeeper,
	authority string,
) Keeper {
	if ps.Name() != "" && !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:         cdc,
		storeKey:    key,
		paramstore:  ps,
		bankKeeper:  bank,
		rolesKeeper: roles,
		authority:   authority,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams fetches the module parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	if k.paramstore.Name() == "" {
		params := types.DefaultParams()
		params.Authority = k.authority
		return params
	}

	var params types.Params
	k.paramstore.GetParamSet(ctx, &params)
	if params.Authority == "" {
		params.Authority = k.authority
	}
	return params
}

// SetParams sets the module params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if k.paramstore.Name() == "" {
		return
	}
	k.paramstore.SetParamSet(ctx, &params)
}

// GetAuthority returns the configured authority address (if any).
func (k Keeper) GetAuthority(ctx sdk.Context) string {
	params := k.GetParams(ctx)
	if params.Authority != "" {
		return params.Authority
	}
	return k.authority
}

// CreateTreasuryPool registers a new treasury pool.
func (k Keeper) CreateTreasuryPool(ctx sdk.Context, authority sdk.AccAddress, pool types.TreasuryPool) error {
	if err := k.assertAuthority(ctx, authority); err != nil {
		return err
	}

	if pool.Id == "" {
		return types.ErrInvalidPoolRequest
	}

	if _, found := k.GetTreasuryPool(ctx, pool.Id); found {
		return fmt.Errorf("treasury pool %s already exists", pool.Id)
	}

	k.setPool(ctx, pool)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePoolCreated,
		sdk.NewAttribute(types.AttributeKeyPoolID, pool.Id),
	))

	return nil
}

// UpdateTreasuryPool updates pool metadata.
func (k Keeper) UpdateTreasuryPool(ctx sdk.Context, authority sdk.AccAddress, pool types.TreasuryPool) error {
	if err := k.assertAuthority(ctx, authority); err != nil {
		return err
	}

	if pool.Id == "" {
		return types.ErrInvalidPoolRequest
	}

	if _, found := k.GetTreasuryPool(ctx, pool.Id); !found {
		return types.ErrPoolNotFound
	}

	k.setPool(ctx, pool)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePoolUpdated,
		sdk.NewAttribute(types.AttributeKeyPoolID, pool.Id),
	))

	return nil
}

// DepositToTreasury records a deposit into a pool and moves funds into the module account.
func (k Keeper) DepositToTreasury(ctx sdk.Context, depositor sdk.AccAddress, poolID string, amount sdk.Coin) error {
	if amount.Amount.IsZero() || amount.Amount.IsNegative() {
		return types.ErrInvalidDeposit
	}

	pool, found := k.GetTreasuryPool(ctx, poolID)
	if !found {
		return types.ErrPoolNotFound
	}

	if k.bankKeeper != nil {
		if err := k.bankKeeper.SendCoinsFromAccountToModule(sdk.WrapSDKContext(ctx), depositor, types.ModuleName, sdk.NewCoins(amount)); err != nil {
			return err
		}
	}

	current := k.getPoolBalance(ctx, poolID, amount.Denom)
	newBalance := current.Add(amount)
	k.setPoolBalance(ctx, poolID, newBalance)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeDeposit,
		sdk.NewAttribute(types.AttributeKeyPoolID, pool.Id),
		sdk.NewAttribute(types.AttributeKeyDenom, amount.Denom),
		sdk.NewAttribute(types.AttributeKeyAmount, amount.Amount.String()),
		sdk.NewAttribute(types.AttributeKeySender, depositor.String()),
	))

	return nil
}

// WithdrawFromTreasury withdraws funds from a pool to a recipient account.
func (k Keeper) WithdrawFromTreasury(ctx sdk.Context, authority sdk.AccAddress, poolID string, recipient sdk.AccAddress, amount sdk.Coin) error {
	if err := k.assertAuthority(ctx, authority); err != nil {
		return err
	}

	if amount.Amount.IsZero() || amount.Amount.IsNegative() {
		return types.ErrInvalidDeposit
	}

	if _, found := k.GetTreasuryPool(ctx, poolID); !found {
		return types.ErrPoolNotFound
	}

	if err := k.performWithdrawal(ctx, poolID, recipient, amount); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeWithdrawal,
		sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
		sdk.NewAttribute(types.AttributeKeyDenom, amount.Denom),
		sdk.NewAttribute(types.AttributeKeyAmount, amount.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyReceiver, recipient.String()),
	))

	return nil
}

// SetPoolReserves replaces the stored reserve targets for a pool.
func (k Keeper) SetPoolReserves(ctx sdk.Context, authority sdk.AccAddress, poolID string, reserves []types.PoolReserves) error {
	if err := k.assertAuthority(ctx, authority); err != nil {
		return err
	}

	if _, found := k.GetTreasuryPool(ctx, poolID); !found {
		return types.ErrPoolNotFound
	}

	k.clearReserves(ctx, poolID)
	for _, reserve := range reserves {
		if reserve.Denom == "" {
			continue
		}
		k.setReserve(ctx, poolID, reserve)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeReserveUpdate,
		sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
	))

	return nil
}

// DisburseClaim performs a payout for the claims module.
func (k Keeper) DisburseClaim(ctx sdk.Context, poolID string, recipient sdk.AccAddress, amount sdk.Coin) error {
	if amount.Amount.IsZero() || amount.Amount.IsNegative() {
		return types.ErrInvalidDeposit
	}

	if _, found := k.GetTreasuryPool(ctx, poolID); !found {
		return types.ErrPoolNotFound
	}

	if err := k.performWithdrawal(ctx, poolID, recipient, amount); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePayout,
		sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
		sdk.NewAttribute(types.AttributeKeyDenom, amount.Denom),
		sdk.NewAttribute(types.AttributeKeyAmount, amount.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyReceiver, recipient.String()),
	))

	return nil
}

// GetTreasuryPool returns a pool by id.
func (k Keeper) GetTreasuryPool(ctx sdk.Context, poolID string) (types.TreasuryPool, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TreasuryPoolKey(poolID))
	if bz == nil {
		return types.TreasuryPool{}, false
	}

	var pool types.TreasuryPool
	k.cdc.MustUnmarshal(bz, &pool)
	return pool, true
}

// ListTreasuryPools returns all pools with pagination support.
func (k Keeper) ListTreasuryPools(ctx sdk.Context, pageReq *query.PageRequest) ([]types.TreasuryPool, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PoolKeyPrefix)

	pools := make([]types.TreasuryPool, 0)
	pageRes, err := query.Paginate(store, pageReq, func(_ []byte, value []byte) error {
		var pool types.TreasuryPool
		k.cdc.MustUnmarshal(value, &pool)
		pools = append(pools, pool)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return pools, pageRes, nil
}

// ExportPools exports stored pools.
func (k Keeper) ExportPools(ctx sdk.Context) []types.TreasuryPool {
	pools := []types.TreasuryPool{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PoolKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var pool types.TreasuryPool
		k.cdc.MustUnmarshal(iterator.Value(), &pool)
		pools = append(pools, pool)
	}

	return pools
}

// ExportBalances exports pool balances.
func (k Keeper) ExportBalances(ctx sdk.Context) []types.PoolBalance {
	balances := []types.PoolBalance{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.BalanceKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var balance types.PoolBalance
		k.cdc.MustUnmarshal(iterator.Value(), &balance)
		balances = append(balances, balance)
	}

	return balances
}

// ExportReserves exports pool reserve targets.
func (k Keeper) ExportReserves(ctx sdk.Context) []types.PoolReserves {
	reserves := []types.PoolReserves{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ReserveKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var reserve types.PoolReserves
		k.cdc.MustUnmarshal(iterator.Value(), &reserve)
		reserves = append(reserves, reserve)
	}

	return reserves
}

// GetPoolBalances returns balances for a specific pool.
func (k Keeper) GetPoolBalances(ctx sdk.Context, poolID string) []types.PoolBalance {
	prefixKey := append(append([]byte{}, types.BalanceKeyPrefix...), []byte(poolID)...)
	prefixKey = append(prefixKey, byte(0))
	store := prefix.NewStore(ctx.KVStore(k.storeKey), prefixKey)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	balances := []types.PoolBalance{}
	for ; iterator.Valid(); iterator.Next() {
		var balance types.PoolBalance
		k.cdc.MustUnmarshal(iterator.Value(), &balance)
		balances = append(balances, balance)
	}

	return balances
}

// GetPoolReserves returns reserve configuration for a pool.
func (k Keeper) GetPoolReserves(ctx sdk.Context, poolID string) []types.PoolReserves {
	prefixKey := append(append([]byte{}, types.ReserveKeyPrefix...), []byte(poolID)...)
	prefixKey = append(prefixKey, byte(0))
	store := prefix.NewStore(ctx.KVStore(k.storeKey), prefixKey)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	reserves := []types.PoolReserves{}
	for ; iterator.Valid(); iterator.Next() {
		var reserve types.PoolReserves
		k.cdc.MustUnmarshal(iterator.Value(), &reserve)
		reserves = append(reserves, reserve)
	}

	return reserves
}

func (k Keeper) performWithdrawal(ctx sdk.Context, poolID string, recipient sdk.AccAddress, amount sdk.Coin) error {
	current := k.getPoolBalance(ctx, poolID, amount.Denom)
	if current.Amount.LT(amount.Amount) {
		return types.ErrInsufficientFunds
	}

	newBalance := current.Sub(amount)
	k.setPoolBalance(ctx, poolID, newBalance)

	if k.bankKeeper != nil {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(sdk.WrapSDKContext(ctx), types.ModuleName, recipient, sdk.NewCoins(amount)); err != nil {
			// revert balance update if transfer fails
			k.setPoolBalance(ctx, poolID, current)
			return err
		}
	}

	return nil
}

func (k Keeper) getPoolBalance(ctx sdk.Context, poolID, denom string) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PoolBalanceKey(poolID, denom))
	if bz == nil {
		return sdk.NewCoin(denom, osmomath.NewInt(0))
	}

	var balance types.PoolBalance
	k.cdc.MustUnmarshal(bz, &balance)
	return sdk.NewCoin(balance.Balance.Denom, balance.Balance.Amount)
}

func (k Keeper) setPoolBalance(ctx sdk.Context, poolID string, coin sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	entry := types.PoolBalance{
		PoolId:  poolID,
		Balance: coin,
	}
	key := types.PoolBalanceKey(poolID, coin.Denom)
	if coin.Amount.IsZero() {
		store.Delete(key)
		return
	}
	store.Set(key, k.cdc.MustMarshal(&entry))
}

func (k Keeper) setPool(ctx sdk.Context, pool types.TreasuryPool) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.TreasuryPoolKey(pool.Id), k.cdc.MustMarshal(&pool))
}

func (k Keeper) setReserve(ctx sdk.Context, poolID string, reserve types.PoolReserves) {
	store := ctx.KVStore(k.storeKey)
	reserve.PoolId = poolID
	store.Set(types.PoolReserveKey(poolID, reserve.Denom), k.cdc.MustMarshal(&reserve))
}

func (k Keeper) clearReserves(ctx sdk.Context, poolID string) {
	prefixKey := append(append([]byte{}, types.ReserveKeyPrefix...), []byte(poolID)...)
	prefixKey = append(prefixKey, byte(0))
	store := prefix.NewStore(ctx.KVStore(k.storeKey), prefixKey)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

func (k Keeper) assertAuthority(ctx sdk.Context, authority sdk.AccAddress) error {
	expected := k.GetAuthority(ctx)
	if expected != "" {
		if authority.String() != expected {
			return types.ErrUnauthorized
		}
		return nil
	}

	if k.rolesKeeper != nil && k.rolesKeeper.HasRole(ctx, authority, rolestypes.Role_ROLE_TREASURY_MANAGER) {
		return nil
	}

	return types.ErrUnauthorized
}

// SetGenesisPool stores a pool during genesis initialization.
func (k Keeper) SetGenesisPool(ctx sdk.Context, pool types.TreasuryPool) {
	k.setPool(ctx, pool)
}

// SetGenesisPoolBalance stores a balance entry during genesis initialization.
func (k Keeper) SetGenesisPoolBalance(ctx sdk.Context, poolID string, coin sdk.Coin) {
	k.setPoolBalance(ctx, poolID, coin)
}

// SetGenesisReserve stores a reserve entry during genesis initialization.
func (k Keeper) SetGenesisReserve(ctx sdk.Context, reserve types.PoolReserves) {
	k.setReserve(ctx, reserve.PoolId, reserve)
}

// Keeper of the limitedaccount store
package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/limitedaccount/types"
)

type (
	// Keeper of the limitedaccount store
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

// NewKeeper creates a new limitedaccount Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,
) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		logger:       logger,
		authority:    authority,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// IsLimitedAccount checks if an address is a limited account
func (k Keeper) IsLimitedAccount(ctx context.Context, address string) bool {
	_, found := k.GetLimitedAccount(ctx, address)
	return found
}

// GetLimitedAccount retrieves a limited account by address
func (k Keeper) GetLimitedAccount(ctx context.Context, address string) (*types.LimitedAccount, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.LimitedAccountKey(address)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return nil, false
	}

	var account types.LimitedAccount
	k.cdc.MustUnmarshal(bz, &account)
	return &account, true
}

// SetLimitedAccount stores a limited account
func (k Keeper) SetLimitedAccount(ctx context.Context, account *types.LimitedAccount) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.LimitedAccountKey(account.Address)
	bz := k.cdc.MustMarshal(account)
	if err := store.Set(key, bz); err != nil {
		panic(fmt.Sprintf("failed to set limited account: %v", err))
	}
}

// RemoveLimitedAccount removes a limited account
func (k Keeper) RemoveLimitedAccount(ctx context.Context, address string) {
	storeAdapter := k.storeService.OpenKVStore(ctx)
	key := types.LimitedAccountKey(address)
	if err := storeAdapter.Delete(key); err != nil {
		panic(fmt.Sprintf("failed to delete limited account: %v", err))
	}
}

// GetAllLimitedAccounts returns all limited accounts
func (k Keeper) GetAllLimitedAccounts(ctx context.Context) []*types.LimitedAccount {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.LimitedAccountPrefix, append(types.LimitedAccountPrefix, 0xFF))
	if err != nil {
		return []*types.LimitedAccount{}
	}
	defer iterator.Close()

	var accounts []*types.LimitedAccount
	for ; iterator.Valid(); iterator.Next() {
		var account types.LimitedAccount
		k.cdc.MustUnmarshal(iterator.Value(), &account)
		accounts = append(accounts, &account)
	}
	return accounts
}

// CanTransact checks if a limited account can make a transaction
func (k Keeper) CanTransact(ctx context.Context, address string) bool {
	account, found := k.GetLimitedAccount(ctx, address)
	if !found {
		return false // not a limited account
	}

	return account.CanTransact()
}

// ProcessTransaction processes a transaction for a limited account
func (k Keeper) ProcessTransaction(ctx context.Context, address string) error {
	account, found := k.GetLimitedAccount(ctx, address)
	if !found {
		// Create new limited account with default parameters
		params := k.GetParams(ctx)
		account = types.NewLimitedAccount(address, params.DefaultDailyLimit)
	}

	// Check if we need to reset the daily counter
	now := time.Now()
	if k.shouldResetCounter(account.LastResetTime, now, k.GetParams(ctx).ResetHour) {
		account.DailyTxCount = 0
		account.LastResetTime = now
	}

	// Check if limit is exceeded
	if account.DailyTxCount >= account.MaxDailyTxs {
		return types.ErrDailyLimitExceeded
	}

	// Increment counter and save
	account.DailyTxCount++
	k.SetLimitedAccount(ctx, account)
	return nil
}

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx context.Context) *types.Params {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil || bz == nil {
		return types.DefaultParams()
	}

	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return &params
}

// SetParams sets the module parameters. Returns error if store.Set fails.
func (k Keeper) SetParams(ctx context.Context, params *types.Params) {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(params)
	if err := store.Set(types.ParamsKey, bz); err != nil {
		panic(fmt.Sprintf("failed to set params: %v", err))
	}
}

// shouldResetCounter checks if the daily counter should be reset
func (k Keeper) shouldResetCounter(lastReset time.Time, now time.Time, resetHour uint32) bool {
	// Check if it's a new day or if we've passed the reset hour
	lastResetDay := lastReset.Truncate(24 * time.Hour)
	nowDay := now.Truncate(24 * time.Hour)

	// If it's a different day, reset
	if !lastResetDay.Equal(nowDay) {
		return true
	}

	// If it's the same day, check if we've passed the reset hour
	resetTime := time.Date(now.Year(), now.Month(), now.Day(), int(resetHour), 0, 0, 0, now.Location())
	return lastReset.Before(resetTime) && now.After(resetTime)
}

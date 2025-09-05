package keeper

import (
	"context"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	freeaccounttypes "github.com/osmosis-labs/osmosis/v30/x/freeaccount/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	logger       log.Logger

	// the address capable of executing a MsgCreateFreeAccount message. Typically, this
	// should be the x/gov module account.
	authority string

	accountKeeper freeaccounttypes.AccountKeeper
}

// NewKeeper creates a new freeaccount Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,
	accountKeeper freeaccounttypes.AccountKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		storeService:  storeService,
		logger:        logger,
		authority:     authority,
		accountKeeper: accountKeeper,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", "x/"+freeaccounttypes.ModuleName)
}

// IsFreeAccount checks if an account is marked as fee-exempt
func (k Keeper) IsFreeAccount(ctx context.Context, addr sdk.AccAddress) bool {
	store := k.storeService.OpenKVStore(ctx)
	key := append(freeaccounttypes.FreeAccountPrefix, addr.Bytes()...)
	has, _ := store.Has(key)
	return has
}

// SetFreeAccount marks an account as fee-exempt
func (k Keeper) SetFreeAccount(ctx context.Context, addr sdk.AccAddress) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(freeaccounttypes.FreeAccountPrefix, addr.Bytes()...)
	return store.Set(key, []byte{1})
}

// RemoveFreeAccount removes the fee-exempt status from an account
func (k Keeper) RemoveFreeAccount(ctx context.Context, addr sdk.AccAddress) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(freeaccounttypes.FreeAccountPrefix, addr.Bytes()...)
	return store.Delete(key)
}

// CreateFreeAccount creates a new free account or converts an existing account to free
func (k Keeper) CreateFreeAccount(ctx context.Context, addr sdk.AccAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check if account already exists
	existingAccount := k.accountKeeper.GetAccount(sdkCtx, addr)

	if existingAccount == nil {
		// Create new free account
		freeAccount := freeaccounttypes.NewFreeAccountWithAddress(addr)
		k.accountKeeper.SetAccount(sdkCtx, freeAccount)
	}

	// Mark account as fee-exempt
	return k.SetFreeAccount(ctx, addr)
}

// InitGenesis initializes the freeaccount module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState freeaccounttypes.GenesisState) {
	for _, addrStr := range genState.FreeAccounts {
		addr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			panic(err)
		}
		err = k.SetFreeAccount(ctx, addr)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the freeaccount module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *freeaccounttypes.GenesisState {
	var freeAccounts []string

	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(freeaccounttypes.FreeAccountPrefix, storetypes.PrefixEndBytes(freeaccounttypes.FreeAccountPrefix))
	if err != nil {
		panic(err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		// Remove prefix to get address bytes
		addrBytes := key[len(freeaccounttypes.FreeAccountPrefix):]
		addr := sdk.AccAddress(addrBytes)
		freeAccounts = append(freeAccounts, addr.String())
	}

	return &freeaccounttypes.GenesisState{
		FreeAccounts: freeAccounts,
	}
}

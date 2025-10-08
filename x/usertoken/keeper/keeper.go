package keeper

import (
	"encoding/binary"
	"fmt"
	"strings"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v30/x/tokenfactory/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

// Keeper orchestrates usertoken state transitions.
type Keeper struct {
	cdc                codec.BinaryCodec
	storeKey           storetypes.StoreKey
	paramstore         paramtypes.Subspace
	accountKeeper      types.AccountKeeper
	bankKeeper         types.BankKeeper
	tokenFactoryKeeper *tokenfactorykeeper.Keeper
}

// NewKeeper constructs a usertoken keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	tokenFactoryKeeper *tokenfactorykeeper.Keeper,
) Keeper {
	if ps.Name() != "" && !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:                cdc,
		storeKey:           storeKey,
		paramstore:         ps,
		accountKeeper:      accountKeeper,
		bankKeeper:         bankKeeper,
		tokenFactoryKeeper: tokenFactoryKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams fetches module parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	if k.paramstore.Name() == "" {
		return types.DefaultParams()
	}

	var params types.Params
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams updates module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if err := params.Validate(); err != nil {
		panic(err)
	}
	if k.paramstore.Name() == "" {
		return
	}
	k.paramstore.SetParamSet(ctx, &params)
}

func (k Keeper) getStore(ctx sdk.Context) storetypes.KVStore {
	return ctx.KVStore(k.storeKey)
}

func (k Keeper) setToken(ctx sdk.Context, token types.Token) {
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(&token)
	store.Set(types.TokenKey(token.Denom), bz)

	nameKey := types.NameIndexKey(normalizeName(token.Name))
	store.Set(nameKey, []byte(token.Denom))

	symbolKey := types.SymbolIndexKey(normalizeSymbol(token.Symbol))
	store.Set(symbolKey, []byte(token.Denom))

	creatorKey := types.CreatorIndexKey(token.Creator, token.Denom)
	store.Set(creatorKey, []byte{1})
}

func (k Keeper) UpdateToken(ctx sdk.Context, token types.Token) error {
	k.setToken(ctx, token)
	return nil
}

func (k Keeper) getToken(ctx sdk.Context, denom string) (types.Token, bool) {
	store := k.getStore(ctx)
	bz := store.Get(types.TokenKey(denom))
	if bz == nil {
		return types.Token{}, false
	}
	var token types.Token
	k.cdc.MustUnmarshal(bz, &token)
	return token, true
}

// GetToken exposes token retrieval for external callers.
func (k Keeper) GetToken(ctx sdk.Context, denom string) (types.Token, bool) {
	return k.getToken(ctx, denom)
}

func (k Keeper) hasToken(ctx sdk.Context, denom string) bool {
	store := k.getStore(ctx)
	return store.Has(types.TokenKey(denom))
}

func (k Keeper) isNameTaken(ctx sdk.Context, name string) bool {
	store := k.getStore(ctx)
	return store.Has(types.NameIndexKey(normalizeName(name)))
}

func (k Keeper) isSymbolTaken(ctx sdk.Context, symbol string) bool {
	store := k.getStore(ctx)
	return store.Has(types.SymbolIndexKey(normalizeSymbol(symbol)))
}

// IterateTokens iterates over all tokens, executing the provided callback.
func (k Keeper) IterateTokens(ctx sdk.Context, cb func(token types.Token) bool) {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.TokenKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var token types.Token
		k.cdc.MustUnmarshal(iterator.Value(), &token)
		if cb(token) {
			break
		}
	}
}

func (k Keeper) queueFounderDeadline(ctx sdk.Context, deadline uint64, denom string) {
	store := k.getStore(ctx)
	store.Set(types.FounderDeadlineKey(deadline, denom), []byte{1})
}

func (k Keeper) removeFounderDeadline(ctx sdk.Context, deadline uint64, denom string) {
	store := k.getStore(ctx)
	store.Delete(types.FounderDeadlineKey(deadline, denom))
}

func (k Keeper) iterateFounderDeadline(ctx sdk.Context, cb func(deadline uint64, denom string) bool) {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.FounderDeadlineQueueKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		if len(key) <= len(types.FounderDeadlineQueueKey)+9 {
			continue
		}
		deadline := binary.BigEndian.Uint64(key[len(types.FounderDeadlineQueueKey) : len(types.FounderDeadlineQueueKey)+8])
		denom := string(key[len(types.FounderDeadlineQueueKey)+9:])
		if cb(deadline, denom) {
			break
		}
	}
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func normalizeSymbol(symbol string) string {
	return strings.ToLower(strings.TrimSpace(symbol))
}

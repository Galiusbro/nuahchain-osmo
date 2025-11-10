package keeper

import (
	"encoding/binary"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/leverage/types"
)

// Keeper manages leverage positions.
type Keeper struct {
	cdc              codec.BinaryCodec
	storeKey         storetypes.StoreKey
	bankKeeper       types.BankKeeper
	oracleKeeper     types.OracleKeeper
	stablecoinKeeper types.StablecoinKeeper
}

// NewKeeper constructs a new keeper.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, bankKeeper types.BankKeeper, oracleKeeper types.OracleKeeper, stablecoinKeeper types.StablecoinKeeper) Keeper {
	if bankKeeper == nil {
		panic("bank keeper cannot be nil")
	}
	if oracleKeeper == nil {
		panic("oracle keeper cannot be nil")
	}
	if stablecoinKeeper == nil {
		panic("stablecoin keeper cannot be nil")
	}
	return Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		bankKeeper:       bankKeeper,
		oracleKeeper:     oracleKeeper,
		stablecoinKeeper: stablecoinKeeper,
	}
}

// OpenPosition records a new leveraged position.
func (k Keeper) OpenPosition(ctx sdk.Context, msg *types.MsgOpenPosition) (*types.Position, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	owner, _ := sdk.AccAddressFromBech32(msg.Owner)
	amount, ok := sdkmath.NewIntFromString(msg.Quote_NDOLLAR)
	if !ok {
		return nil, fmt.Errorf("invalid quote amount")
	}
	if !amount.IsPositive() {
		return nil, fmt.Errorf("quote must be positive")
	}

	// Get real NDOLLAR denom from stablecoin keeper
	ndollarDenom := k.stablecoinKeeper.GetNDollarDenom(ctx)
	quoteCoin := sdk.NewCoin(ndollarDenom, amount)

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, sdk.NewCoins(quoteCoin)); err != nil {
		return nil, err
	}

	price, found := k.oracleKeeper.GetPrice(ctx, msg.Symbol)
	if !found {
		return nil, fmt.Errorf("price for %s not found", msg.Symbol)
	}
	priceDec, err := sdkmath.LegacyNewDecFromStr(price.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid price value: %w", err)
	}
	if priceDec.IsZero() {
		return nil, fmt.Errorf("price cannot be zero")
	}

	quoteDec := sdkmath.LegacyNewDecFromInt(quoteCoin.Amount)
	baseQty := quoteDec.Quo(priceDec)

	id := k.nextPositionID(ctx)

	position := &types.Position{
		Id:         id,
		Owner:      msg.Owner,
		Symbol:     msg.Symbol,
		Side:       msg.Side,
		BaseQty:    baseQty.String(),
		EntryPrice: price.Value,
		Leverage:   msg.Leverage,
	}

	k.setPosition(ctx, position)
	k.addOwnerIndex(ctx, owner, id)
	k.setQuoteAmount(ctx, id, quoteCoin)

	return position, nil
}

// ClosePosition removes an existing position and returns funds.
func (k Keeper) ClosePosition(ctx sdk.Context, msg *types.MsgClosePosition) (string, error) {
	if err := msg.ValidateBasic(); err != nil {
		return "", err
	}

	position, found := k.GetPosition(ctx, msg.Id)
	if !found {
		return "", fmt.Errorf("position not found")
	}

	if position.Owner != msg.Owner {
		return "", fmt.Errorf("unauthorized")
	}

	owner, _ := sdk.AccAddressFromBech32(msg.Owner)
	quoteCoin, foundQuote := k.getQuoteAmount(ctx, msg.Id)
	if !foundQuote {
		// Get real NDOLLAR denom from stablecoin keeper
		ndollarDenom := k.stablecoinKeeper.GetNDollarDenom(ctx)
		quoteCoin = sdk.NewCoin(ndollarDenom, sdkmath.ZeroInt())
	}

	k.deletePosition(ctx, owner, msg.Id)

	if quoteCoin.Amount.IsPositive() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, sdk.NewCoins(quoteCoin)); err != nil {
			return "", err
		}
	}

	// Simplified PnL calculation: zero.
	return "0", nil
}

// GetPosition retrieves a position by id.
func (k Keeper) GetPosition(ctx sdk.Context, id uint64) (*types.Position, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PositionKeyPrefix)
	bz := store.Get(uint64ToBytes(id))
	if len(bz) == 0 {
		return nil, false
	}
	var position types.Position
	k.cdc.MustUnmarshal(bz, &position)
	return &position, true
}

func (k Keeper) setPosition(ctx sdk.Context, position *types.Position) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PositionKeyPrefix)
	bz := k.cdc.MustMarshal(position)
	store.Set(uint64ToBytes(position.Id), bz)
}

func (k Keeper) deletePosition(ctx sdk.Context, owner sdk.AccAddress, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PositionKeyPrefix)
	store.Delete(uint64ToBytes(id))

	ownerStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.OwnerPositionsPrefix)
	ownerPrefixed := prefix.NewStore(ownerStore, owner.Bytes())
	ownerPrefixed.Delete(uint64ToBytes(id))

	quoteStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.PositionQuotePrefix)
	quoteStore.Delete(uint64ToBytes(id))
}

func (k Keeper) addOwnerIndex(ctx sdk.Context, owner sdk.AccAddress, id uint64) {
	ownerStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.OwnerPositionsPrefix)
	store := prefix.NewStore(ownerStore, owner.Bytes())
	store.Set(uint64ToBytes(id), []byte{1})
}

func (k Keeper) setQuoteAmount(ctx sdk.Context, id uint64, coin sdk.Coin) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PositionQuotePrefix)
	store.Set(uint64ToBytes(id), []byte(coin.Amount.String()))
}

func (k Keeper) getQuoteAmount(ctx sdk.Context, id uint64) (sdk.Coin, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PositionQuotePrefix)
	bz := store.Get(uint64ToBytes(id))
	if len(bz) == 0 {
		return sdk.Coin{}, false
	}
	amount, ok := sdkmath.NewIntFromString(string(bz))
	if !ok {
		panic("invalid stored quote amount")
	}
	// Get real NDOLLAR denom from stablecoin keeper
	ndollarDenom := k.stablecoinKeeper.GetNDollarDenom(ctx)
	return sdk.NewCoin(ndollarDenom, amount), true
}

func (k Keeper) nextPositionID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextPositionIDKey)
	var id uint64 = 1
	if len(bz) == 8 {
		id = binary.BigEndian.Uint64(bz)
	}
	next := make([]byte, 8)
	binary.BigEndian.PutUint64(next, id+1)
	store.Set(types.NextPositionIDKey, next)
	return id
}

func uint64ToBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}

// InitGenesis initializes state.
func (k Keeper) InitGenesis(ctx sdk.Context, state *types.GenesisState) {
	if state == nil {
		return
	}
	for _, position := range state.Positions {
		k.setPosition(ctx, position)
		owner, err := sdk.AccAddressFromBech32(position.Owner)
		if err != nil {
			panic(err)
		}
		k.addOwnerIndex(ctx, owner, position.Id)
	}
	nextID := state.NextPositionId
	if nextID == 0 {
		nextID = uint64(len(state.Positions)) + 1
	}
	ctx.KVStore(k.storeKey).Set(types.NextPositionIDKey, uint64ToBytes(nextID))
}

// ExportGenesis exports current state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	state := types.DefaultGenesis()

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PositionKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var position types.Position
		k.cdc.MustUnmarshal(iterator.Value(), &position)
		posCopy := position
		state.Positions = append(state.Positions, &posCopy)
	}

	next := ctx.KVStore(k.storeKey).Get(types.NextPositionIDKey)
	if len(next) == 8 {
		state.NextPositionId = binary.BigEndian.Uint64(next)
	}
	return state
}

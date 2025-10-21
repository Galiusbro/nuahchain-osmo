package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/collateral/types"
)

// Keeper provides state management for collateral positions.
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	bankKeeper types.BankKeeper
}

// NewKeeper creates a new collateral keeper instance.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, bankKeeper types.BankKeeper) Keeper {
	if bankKeeper == nil {
		panic("bank keeper cannot be nil")
	}
	return Keeper{cdc: cdc, storeKey: storeKey, bankKeeper: bankKeeper}
}

// DepositCollateral transfers funds into the module account and records collateral.
func (k Keeper) DepositCollateral(ctx sdk.Context, depositor sdk.AccAddress, coin sdk.Coin) error {
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleName, sdk.NewCoins(coin)); err != nil {
		return err
	}
	k.addCollateral(ctx, depositor, coin)
	return nil
}

// WithdrawCollateral releases collateral back to the owner if sufficient balance exists.
func (k Keeper) WithdrawCollateral(ctx sdk.Context, owner sdk.AccAddress, coin sdk.Coin) error {
	current := k.getCollateralAmount(ctx, owner, coin.Denom)
	if current.LT(coin.Amount) {
		return fmt.Errorf("insufficient collateral: have %s need %s", current.String(), coin.Amount.String())
	}
	k.setCollateral(ctx, owner, coin.Denom, current.Sub(coin.Amount))
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, sdk.NewCoins(coin))
}

// GetPositions returns the list of collateral positions for the provided owner address.
func (k Keeper) GetPositions(ctx sdk.Context, owner sdk.AccAddress) []types.Position {
	ownerStore := k.ownerStore(ctx, owner)
	iterator := ownerStore.Iterator(nil, nil)
	defer iterator.Close()

	var positions []types.Position
	for ; iterator.Valid(); iterator.Next() {
		amount, ok := sdkmath.NewIntFromString(string(iterator.Value()))
		if !ok {
			panic("invalid stored amount")
		}
		denom := string(iterator.Key())
		positions = append(positions, types.Position{
			Owner:  owner.String(),
			Denom:  denom,
			Amount: amount.String(),
		})
	}

	return positions
}

// InitGenesis initializes state from genesis data.
func (k Keeper) InitGenesis(ctx sdk.Context, state *types.GenesisState) {
	if state == nil {
		return
	}
	for _, position := range state.Positions {
		if position == nil {
			continue
		}
		owner, err := sdk.AccAddressFromBech32(position.Owner)
		if err != nil {
			panic(err)
		}
		amount, ok := sdkmath.NewIntFromString(position.Amount)
		if !ok {
			panic("invalid genesis amount")
		}
		k.setCollateral(ctx, owner, position.Denom, amount)
	}
}

// ExportGenesis exports current module state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	state := types.DefaultGenesis()
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CollateralKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		// address.Len is the length of Bech32 address string (32), but we need the raw address bytes length (20)
		addrLen := 20 // Standard address length in bytes
		if len(key) <= addrLen {
			panic(fmt.Sprintf("invalid collateral key length: expected > %d, got %d", addrLen, len(key)))
		}
		ownerBytes := key[:addrLen]
		denom := string(key[addrLen:])
		amount, ok := sdkmath.NewIntFromString(string(iterator.Value()))
		if !ok {
			panic(fmt.Sprintf("invalid stored amount: %s", string(iterator.Value())))
		}
		state.Positions = append(state.Positions, &types.Position{
			Owner:  sdk.AccAddress(ownerBytes).String(),
			Denom:  denom,
			Amount: amount.String(),
		})
	}

	return state
}

func (k Keeper) addCollateral(ctx sdk.Context, owner sdk.AccAddress, coin sdk.Coin) {
	current := k.getCollateralAmount(ctx, owner, coin.Denom)
	k.setCollateral(ctx, owner, coin.Denom, current.Add(coin.Amount))
}

func (k Keeper) getCollateralAmount(ctx sdk.Context, owner sdk.AccAddress, denom string) sdkmath.Int {
	ownerStore := k.ownerStore(ctx, owner)
	bz := ownerStore.Get([]byte(types.NormalizeDenom(denom)))
	if len(bz) == 0 {
		return sdkmath.ZeroInt()
	}
	amount, ok := sdkmath.NewIntFromString(string(bz))
	if !ok {
		panic("invalid stored amount")
	}
	return amount
}

func (k Keeper) setCollateral(ctx sdk.Context, owner sdk.AccAddress, denom string, amount sdkmath.Int) {
	denom = types.NormalizeDenom(denom)
	ownerStore := k.ownerStore(ctx, owner)
	if !amount.IsPositive() {
		ownerStore.Delete([]byte(denom))
		return
	}
	ownerStore.Set([]byte(denom), []byte(amount.String()))
}

func (k Keeper) ownerStore(ctx sdk.Context, owner sdk.AccAddress) prefix.Store {
	base := prefix.NewStore(ctx.KVStore(k.storeKey), types.CollateralKeyPrefix)
	return prefix.NewStore(base, owner.Bytes())
}

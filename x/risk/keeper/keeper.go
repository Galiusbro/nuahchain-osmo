package keeper

import (
	"fmt"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/risk/types"
)

// Keeper manages risk parameters for leveraged assets.
type Keeper struct {
	cdc       codec.BinaryCodec
	storeKey  storetypes.StoreKey
	authority string
}

// NewKeeper creates a new risk keeper instance.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, authority string) Keeper {
	if authority == "" {
		panic("authority cannot be empty")
	}

	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,
	}
}

// GetAuthority returns the configured authority address.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetRiskParams stores the provided risk parameters.
func (k Keeper) SetRiskParams(ctx sdk.Context, params *types.RiskParams) error {
	if err := types.ValidateRiskParams(params); err != nil {
		return err
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RiskParamsKeyPrefix)
	symbol := types.NormalizeSymbol(params.Symbol)
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	copy := *params
	copy.Symbol = symbol
	bz := k.cdc.MustMarshal(&copy)
	store.Set([]byte(symbol), bz)
	return nil
}

// GetRiskParams returns risk parameters for the provided symbol.
func (k Keeper) GetRiskParams(ctx sdk.Context, symbol string) (*types.RiskParams, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RiskParamsKeyPrefix)
	bz := store.Get([]byte(types.NormalizeSymbol(symbol)))
	if len(bz) == 0 {
		return nil, false
	}

	var params types.RiskParams
	k.cdc.MustUnmarshal(bz, &params)
	return &params, true
}

// IterateRiskParams iterates over all stored risk parameters.
func (k Keeper) IterateRiskParams(ctx sdk.Context, cb func(*types.RiskParams) bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RiskParamsKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var params types.RiskParams
		k.cdc.MustUnmarshal(iterator.Value(), &params)
		if cb(&params) {
			break
		}
	}
}

// InitGenesis initializes module state from genesis data.
func (k Keeper) InitGenesis(ctx sdk.Context, state *types.GenesisState) {
	if state == nil {
		return
	}

	for _, params := range state.RiskParams {
		if err := k.SetRiskParams(ctx, params); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports the module state into a genesis structure.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()

	k.IterateRiskParams(ctx, func(params *types.RiskParams) bool {
		copy := *params
		genesis.RiskParams = append(genesis.RiskParams, &copy)
		return false
	})

	return genesis
}

package keeper

import (
	"fmt"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v30/x/risk/types"
)

// Keeper manages risk parameters for leveraged assets.
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	authority  string
	paramstore paramtypes.Subspace
}

// NewKeeper creates a new risk keeper instance.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, authority string, ps paramtypes.Subspace) Keeper {
	if authority == "" {
		panic("authority cannot be empty")
	}

	if ps.Name() != "" && !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		authority:  authority,
		paramstore: ps,
	}
}

// GetAuthority returns the configured authority address.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetParams returns module parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	if k.paramstore.Name() == "" {
		return types.DefaultParams()
	}

	var params types.Params
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams stores module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if err := params.Validate(); err != nil {
		panic(err)
	}

	if k.paramstore.Name() == "" {
		return
	}

	k.paramstore.SetParamSet(ctx, &params)
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
		defaultParams := types.DefaultParams()
		k.SetParams(ctx, defaultParams)
		return
	}

	if state.Params != nil {
		k.SetParams(ctx, *state.Params)
	} else {
		k.SetParams(ctx, types.DefaultParams())
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
	params := k.GetParams(ctx)
	genesis.Params = &params

	k.IterateRiskParams(ctx, func(params *types.RiskParams) bool {
		copy := *params
		genesis.RiskParams = append(genesis.RiskParams, &copy)
		return false
	})

	return genesis
}

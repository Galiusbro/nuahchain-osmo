package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v30/x/fees/types"

	osmomath "github.com/osmosis-labs/osmosis/osmomath"
)

// Keeper provides access to fee parameters.
type Keeper struct {
	cdc        codec.BinaryCodec
	paramstore paramtypes.Subspace
}

// NewKeeper creates a new fees keeper instance.
func NewKeeper(cdc codec.BinaryCodec, ps paramtypes.Subspace) Keeper {
	if ps.Name() != "" && !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:        cdc,
		paramstore: ps,
	}
}

// GetParams returns the current module parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	if k.paramstore.Name() == "" {
		return types.DefaultParams()
	}

	var params types.Params
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams stores the provided parameters after validation.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	if k.paramstore.Name() == "" {
		return nil
	}

	k.paramstore.SetParamSet(ctx, &params)
	return nil
}

// GetTradeFeeRate returns the trade fee rate as a decimal.
func (k Keeper) GetTradeFeeRate(ctx sdk.Context) osmomath.Dec {
	return k.GetParams(ctx).TradeFeeRateDec()
}

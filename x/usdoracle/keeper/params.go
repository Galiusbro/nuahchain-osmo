package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/usdoracle/types"
)

// GetParams get all parameters as types.Params (context.Context version)
func (k Keeper) GetParams(ctx context.Context) types.Params {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return k.GetParamsSDK(sdkCtx)
}

// GetParamsSDK get all parameters as types.Params (sdk.Context version)
func (k Keeper) GetParamsSDK(ctx sdk.Context) types.Params {
	var p types.Params
	k.paramstore.GetParamSet(ctx, &p)
	return p
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/risk/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/risk/types"
)

func TestMsgSetRiskParams(t *testing.T) {
	k, ctx := setupKeeper(t)
	server := keeper.NewMsgServerImpl(k)
	goCtx := sdk.WrapSDKContext(ctx)

	msg := types.NewMsgSetRiskParams(authority, &types.RiskParams{
		Symbol:            "btc",
		MaxLeverage:       "2.5",
		MaintenanceMargin: "0.1",
		InitialMargin:     "0.2",
	})

	_, err := server.SetRiskParams(goCtx, msg)
	require.NoError(t, err)
	stored, found := k.GetRiskParams(ctx, "BTC")
	require.True(t, found)
	require.Equal(t, "BTC", stored.Symbol)

	_, err = server.SetRiskParams(goCtx, types.NewMsgSetRiskParams("osmo1notreal", msg.Params))
	require.Error(t, err)
}

func TestMsgSetRiskParamsInvalid(t *testing.T) {
	k, ctx := setupKeeper(t)
	server := keeper.NewMsgServerImpl(k)

	_, err := server.SetRiskParams(context.Background(), nil)
	require.Error(t, err)

	_, err = server.SetRiskParams(sdk.WrapSDKContext(ctx), &types.MsgSetRiskParams{
		Authority: authority,
		Params:    &types.RiskParams{Symbol: "", MaxLeverage: "1", MaintenanceMargin: "0.1", InitialMargin: "0.2"},
	})
	require.Error(t, err)
}

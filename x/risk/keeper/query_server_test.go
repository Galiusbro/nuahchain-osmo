package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/risk/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/risk/types"
)

func TestQueryRiskParams(t *testing.T) {
	k, ctx := setupKeeper(t)
	require.NoError(t, k.SetRiskParams(ctx, &types.RiskParams{
		Symbol:            "btc",
		MaxLeverage:       "4",
		MaintenanceMargin: "0.12",
		InitialMargin:     "0.18",
	}))

	server := keeper.NewQueryServerImpl(k)
	resp, err := server.RiskParams(sdk.WrapSDKContext(ctx), &types.QueryRiskParamsRequest{Symbol: "BTC"})
	require.NoError(t, err)
	require.NotNil(t, resp.Params)
	require.Equal(t, "BTC", resp.Params.Symbol)
}

func TestQueryRiskParamsMissing(t *testing.T) {
	k, ctx := setupKeeper(t)
	server := keeper.NewQueryServerImpl(k)

	_, err := server.RiskParams(sdk.WrapSDKContext(ctx), &types.QueryRiskParamsRequest{Symbol: "ETH"})
	require.Error(t, err)
}

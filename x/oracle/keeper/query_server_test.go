package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

func TestQueryPrice(t *testing.T) {
	k, ctx := setupKeeper(t)
	srv := keeper.NewQueryServer(k)
	goCtx := sdk.WrapSDKContext(ctx)

	k.SetPrice(ctx, &types.Price{Symbol: "GOLD", Value: "2000"})

	resp, err := srv.Price(goCtx, &types.QueryPriceRequest{Symbol: "GOLD"})
	require.NoError(t, err)
	require.Equal(t, "2000", resp.Price.Value)

	_, err = srv.Price(goCtx, &types.QueryPriceRequest{Symbol: "SILVER"})
	require.Error(t, err)
	require.Equal(t, codes.NotFound, status.Code(err))

	_, err = srv.Price(goCtx, &types.QueryPriceRequest{Symbol: " "})
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

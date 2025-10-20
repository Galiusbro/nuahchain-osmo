package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/assets/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
)

func TestQueryAsset(t *testing.T) {
	k, ctx := setupKeeper(t)
	srv := keeper.NewQueryServer(k)
	goCtx := sdk.WrapSDKContext(ctx)

	asset := &types.Asset{
		Symbol:   "EUR",
		Name:     "Euro",
		Type:     "fiat",
		Decimals: 2,
		Status:   "active",
	}
	k.SetAsset(ctx, asset)

	resp, err := srv.Asset(goCtx, &types.QueryAssetRequest{Symbol: "EUR"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, asset, resp.Asset)

	_, err = srv.Asset(goCtx, &types.QueryAssetRequest{Symbol: "USD"})
	require.Error(t, err)
	require.Equal(t, codes.NotFound, status.Code(err))

	_, err = srv.Asset(goCtx, &types.QueryAssetRequest{Symbol: "   "})
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestQueryAssetsPagination(t *testing.T) {
	k, ctx := setupKeeper(t)
	srv := keeper.NewQueryServer(k)
	goCtx := sdk.WrapSDKContext(ctx)

	symbols := []string{"AAA", "BBB", "CCC"}
	for _, sym := range symbols {
		k.SetAsset(ctx, &types.Asset{
			Symbol:   sym,
			Name:     sym,
			Type:     "test",
			Decimals: 2,
			Status:   "active",
		})
	}

	firstPage, err := srv.Assets(goCtx, &types.QueryAssetsRequest{
		Pagination: &sdkquery.PageRequest{Limit: 2},
	})
	require.NoError(t, err)
	require.Len(t, firstPage.Assets, 2)
	require.NotNil(t, firstPage.Pagination)
	require.NotZero(t, len(firstPage.Pagination.NextKey))

	secondPage, err := srv.Assets(goCtx, &types.QueryAssetsRequest{
		Pagination: &sdkquery.PageRequest{Key: firstPage.Pagination.NextKey},
	})
	require.NoError(t, err)
	require.Len(t, secondPage.Assets, 1)
	require.Equal(t, 0, len(secondPage.Pagination.NextKey))
}

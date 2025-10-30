package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

func TestSetAndGetLatestPriceHistory(t *testing.T) {
	k, ctx := setupKeeper(t)

	symbol := "BTC-USD"

	// Insert 5 entries with increasing timestamps
	for i := int64(1); i <= 5; i++ {
		entry := &types.PriceHistoryEntry{
			Symbol:      symbol,
			Value:       "100",
			Source:      "test",
			Timestamp:   1000 + i,
			Confidence:  1.0,
			BlockHeight: 10 + i,
		}
		k.SetPriceHistory(ctx, entry)
	}

	// Fetch latest 3 entries — should be timestamps 1005, 1004, 1003
	latest, err := k.GetLatestPriceHistory(ctx, symbol, 3)
	require.NoError(t, err)
	require.Len(t, latest, 3)
	require.Equal(t, int64(1005), latest[0].Timestamp)
	require.Equal(t, int64(1004), latest[1].Timestamp)
	require.Equal(t, int64(1003), latest[2].Timestamp)
}

func TestGetPriceHistoryByRange(t *testing.T) {
	k, ctx := setupKeeper(t)

	symbol := "ETH-USD"

	// Insert entries at timestamps 2001..2005
	for i := int64(1); i <= 5; i++ {
		entry := &types.PriceHistoryEntry{
			Symbol:      symbol,
			Value:       "50",
			Source:      "test",
			Timestamp:   2000 + i,
			Confidence:  1.0,
			BlockHeight: 20 + i,
		}
		k.SetPriceHistory(ctx, entry)
	}

	// Query from 2002 to 2004 inclusive, limit 10
	entries, err := k.GetPriceHistory(ctx, symbol, 2002, 2004, 10)
	require.NoError(t, err)
	require.Len(t, entries, 3)
	require.Equal(t, int64(2002), entries[0].Timestamp)
	require.Equal(t, int64(2003), entries[1].Timestamp)
	require.Equal(t, int64(2004), entries[2].Timestamp)
}

func TestQueryServerPriceHistory(t *testing.T) {
	baseKeeper, ctx := setupKeeper(t)

	// Populate some history using the concrete keeper
	symbol := "GC=F"
	for i := int64(0); i < 3; i++ {
		baseKeeper.SetPriceHistory(ctx, &types.PriceHistoryEntry{
			Symbol:      symbol,
			Value:       "2000",
			Source:      "test",
			Timestamp:   3000 + i,
			Confidence:  1.0,
			BlockHeight: 30 + i,
		})
	}

	// Create query server and request
	qs := keeper.NewQueryServer(baseKeeper)
	res, err := qs.PriceHistory(sdk.WrapSDKContext(ctx), &types.QueryPriceHistoryRequest{
		Symbol: symbol,
		Limit:  2,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Entries, 2)
	// Latest first: timestamps 3002, 3001
	require.Equal(t, int64(3002), res.Entries[0].Timestamp)
	require.Equal(t, int64(3001), res.Entries[1].Timestamp)
}

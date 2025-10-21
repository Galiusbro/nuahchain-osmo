package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

func TestKeeperSetGetPrice(t *testing.T) {
	k, ctx := setupKeeper(t)

	price := &types.Price{
		Symbol: "GOLD",
		Value:  "2000",
	}

	k.SetPrice(ctx, price)

	got, found := k.GetPrice(ctx, "GOLD")
	require.True(t, found)
	require.Equal(t, price, got)

	_, found = k.GetPrice(ctx, "SILVER")
	require.False(t, found)
}

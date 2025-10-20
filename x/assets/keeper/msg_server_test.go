package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/assets/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
)

func TestMsgEnsureAssetCreates(t *testing.T) {
	k, ctx := setupKeeper(t)
	srv := keeper.NewMsgServer(k)
	goCtx := sdk.WrapSDKContext(ctx)

	priv := secp256k1.GenPrivKey()
	creator := sdk.AccAddress(priv.PubKey().Address()).String()

	resp, err := srv.EnsureAsset(goCtx, types.NewMsgEnsureAsset(creator, "EUR"))
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "EUR", resp.Asset.Symbol)
	require.Equal(t, "EUR", resp.Asset.Name)
	require.Equal(t, "unknown", resp.Asset.Type)
	require.Equal(t, uint32(2), resp.Asset.Decimals)
	require.Equal(t, "active", resp.Asset.Status)

	stored, found := k.GetAsset(ctx, "EUR")
	require.True(t, found)
	require.Equal(t, resp.Asset, stored)

	events := ctx.EventManager().Events()
	require.Len(t, events, 1)
	require.Equal(t, types.EventTypeAssetCreated, events[0].Type)
}

func TestMsgEnsureAssetIdempotent(t *testing.T) {
	k, ctx := setupKeeper(t)
	srv := keeper.NewMsgServer(k)
	goCtx := sdk.WrapSDKContext(ctx)

	priv := secp256k1.GenPrivKey()
	creator := sdk.AccAddress(priv.PubKey().Address()).String()

	_, err := srv.EnsureAsset(goCtx, types.NewMsgEnsureAsset(creator, "GOLD"))
	require.NoError(t, err)
	initialEventCount := len(ctx.EventManager().Events())

	resp, err := srv.EnsureAsset(goCtx, types.NewMsgEnsureAsset(creator, "GOLD"))
	require.NoError(t, err)
	require.NotNil(t, resp.Asset)
	require.Equal(t, "GOLD", resp.Asset.Symbol)
	require.Equal(t, initialEventCount, len(ctx.EventManager().Events()), "no new event expected for existing asset")
}

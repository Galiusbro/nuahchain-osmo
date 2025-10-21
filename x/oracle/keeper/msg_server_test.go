package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

func TestMsgSetPriceAuthorized(t *testing.T) {
	k, ctx := setupKeeper(t)
	authority := k.Authority()

	authorityAddr, err := sdk.AccAddressFromBech32(authority)
	require.NoError(t, err)

	srv := keeper.NewMsgServer(k)
	goCtx := sdk.WrapSDKContext(ctx)

	resp, err := srv.SetPrice(goCtx, types.NewMsgSetPrice(authorityAddr.String(), "GOLD", "2000"))
	require.NoError(t, err)
	require.NotNil(t, resp)

	stored, found := k.GetPrice(ctx, "GOLD")
	require.True(t, found)
	require.Equal(t, "2000", stored.Value)
}

func TestMsgSetPriceUnauthorized(t *testing.T) {
	k, ctx := setupKeeper(t)
	srv := keeper.NewMsgServer(k)
	goCtx := sdk.WrapSDKContext(ctx)

	priv := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	_, err := srv.SetPrice(goCtx, types.NewMsgSetPrice(addr.String(), "GOLD", "2000"))
	require.Error(t, err)
}

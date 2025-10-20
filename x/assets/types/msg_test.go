package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
)

func TestMsgEnsureAssetValidateBasic(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address()).String()

	msg := types.NewMsgEnsureAsset(addr, "EUR")
	require.NoError(t, msg.ValidateBasic())

	invalidAddrMsg := types.NewMsgEnsureAsset("invalid", "EUR")
	require.Error(t, invalidAddrMsg.ValidateBasic())

	emptySymbolMsg := types.NewMsgEnsureAsset(addr, "   ")
	require.Error(t, emptySymbolMsg.ValidateBasic())
}

func TestMsgEnsureAssetGetSigners(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	msg := types.NewMsgEnsureAsset(addr.String(), "EUR")
	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr.String(), signers[0].String())
}

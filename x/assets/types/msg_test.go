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

func TestMsgBuyAssetValidateBasic(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address()).String()

	msg := types.NewMsgBuyAsset(addr, "GOLD", "1000")
	require.NoError(t, msg.ValidateBasic())
	require.Equal(t, "1000", msg.Amount_NDOLLAR)

	invalidAddrMsg := types.NewMsgBuyAsset("invalid", "GOLD", "1000")
	require.Error(t, invalidAddrMsg.ValidateBasic())

	emptySymbol := types.NewMsgBuyAsset(addr, " ", "1000")
	require.Error(t, emptySymbol.ValidateBasic())

	emptyAmount := types.NewMsgBuyAsset(addr, "GOLD", " ")
	require.Error(t, emptyAmount.ValidateBasic())

	badAmount := types.NewMsgBuyAsset(addr, "GOLD", "1.5")
	require.Error(t, badAmount.ValidateBasic())
}

func TestMsgBuyAssetGetSigners(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	msg := types.NewMsgBuyAsset(addr.String(), "GOLD", "1000")
	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr.String(), signers[0].String())
}

func TestMsgSellAssetValidateBasic(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address()).String()

	msg := types.NewMsgSellAsset(addr, "GOLD", "0.5")
	require.NoError(t, msg.ValidateBasic())

	invalidAddr := types.NewMsgSellAsset("invalid", "GOLD", "0.5")
	require.Error(t, invalidAddr.ValidateBasic())

	emptySymbol := types.NewMsgSellAsset(addr, " ", "0.5")
	require.Error(t, emptySymbol.ValidateBasic())

	emptyAmount := types.NewMsgSellAsset(addr, "GOLD", " ")
	require.Error(t, emptyAmount.ValidateBasic())

	negativeAmount := types.NewMsgSellAsset(addr, "GOLD", "-1")
	require.Error(t, negativeAmount.ValidateBasic())
}

func TestMsgSellAssetGetSigners(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	msg := types.NewMsgSellAsset(addr.String(), "GOLD", "0.5")
	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr.String(), signers[0].String())
}

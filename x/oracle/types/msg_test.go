package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

func TestMsgSetPriceValidateBasic(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address()).String()

	msg := types.NewMsgSetPrice(addr, "GOLD", "2000")
	require.NoError(t, msg.ValidateBasic())

	invalidAuthority := types.NewMsgSetPrice("invalid", "GOLD", "2000")
	require.Error(t, invalidAuthority.ValidateBasic())

	emptySymbol := types.NewMsgSetPrice(addr, " ", "2000")
	require.Error(t, emptySymbol.ValidateBasic())

	emptyValue := types.NewMsgSetPrice(addr, "GOLD", " ")
	require.Error(t, emptyValue.ValidateBasic())
}

func TestMsgSetPriceGetSigners(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	msg := types.NewMsgSetPrice(addr.String(), "GOLD", "2000")
	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr.String(), signers[0].String())
}

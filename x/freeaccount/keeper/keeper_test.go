package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	freeaccounttypes "github.com/osmosis-labs/osmosis/v30/x/freeaccount/types"
)

func TestBasicFunctionality(t *testing.T) {
	// Test that module name is correct
	require.Equal(t, "freeaccount", freeaccounttypes.ModuleName)

	// Test that store key is correct
	require.Equal(t, "freeaccount", freeaccounttypes.StoreKey)

	// Create test addresses
	authorityAddr := sdk.AccAddress("authority_address_1")
	testAddr := sdk.AccAddress("test_address_12345")

	// Test message creation
	msg := freeaccounttypes.NewMsgCreateFreeAccount(authorityAddr.String(), testAddr.String())
	require.NotNil(t, msg)
	require.Equal(t, authorityAddr.String(), msg.Authority)
	require.Equal(t, testAddr.String(), msg.Address)

	// Test message validation
	err := msg.ValidateBasic()
	require.NoError(t, err)
}

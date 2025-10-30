package authz_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/authz"
	assetstypes "github.com/osmosis-labs/osmosis/v30/x/assets/types"
)

const testAddrLen = 20

func testAddress(b byte) string {
	addr := sdk.AccAddress(bytes.Repeat([]byte{b}, testAddrLen))
	return addr.String()
}

func TestAuthzClient_ValidateExecRequest(t *testing.T) {
	tests := []struct {
		name          string
		request       *authz.ExecRequest
		expectedError bool
	}{
		{
			name: "valid request",
			request: &authz.ExecRequest{
				Grantee: testAddress(1),
				Msgs:    []sdk.Msg{assetstypes.NewMsgBuyAsset(testAddress(2), "BTC", "1000000")},
			},
			expectedError: false,
		},
		{
			name:          "nil request",
			request:       nil,
			expectedError: true,
		},
		{
			name: "empty grantee",
			request: &authz.ExecRequest{
				Grantee: "",
				Msgs:    []sdk.Msg{assetstypes.NewMsgBuyAsset(testAddress(2), "BTC", "1000000")},
			},
			expectedError: true,
		},
		{
			name: "empty messages",
			request: &authz.ExecRequest{
				Grantee: testAddress(1),
				Msgs:    []sdk.Msg{},
			},
			expectedError: true,
		},
		{
			name: "invalid grantee address",
			request: &authz.ExecRequest{
				Grantee: "invalid_address",
				Msgs:    []sdk.Msg{assetstypes.NewMsgBuyAsset(testAddress(2), "BTC", "1000000")},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &authz.Client{}
			err := client.ValidateExecRequest(tt.request)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

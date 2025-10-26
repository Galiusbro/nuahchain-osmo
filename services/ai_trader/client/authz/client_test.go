package authz_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/authz"
	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
)

func TestAuthzClient_ValidateExecRequest(t *testing.T) {
	tests := []struct {
		name          string
		request       *authz.ExecRequest
		expectedError bool
	}{
		{
			name: "valid request",
			request: &authz.ExecRequest{
				Grantee: "cosmos1qk93t4j0yyzgqgt6k5qf8deh8fq6smpn3ntu3x",
				Msgs:    []sdk.Msg{types.NewMsgBuyAsset("cosmos1p9qh4ldfd6n0qehujsal4k7g0e37kel90rc4ts", "BTC", "1000000")},
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
				Msgs:    []sdk.Msg{types.NewMsgBuyAsset("cosmos1p9qh4ldfd6n0qehujsal4k7g0e37kel90rc4ts", "BTC", "1000000")},
			},
			expectedError: true,
		},
		{
			name: "empty messages",
			request: &authz.ExecRequest{
				Grantee: "cosmos1qk93t4j0yyzgqgt6k5qf8deh8fq6smpn3ntu3x",
				Msgs:    []sdk.Msg{},
			},
			expectedError: true,
		},
		{
			name: "invalid grantee address",
			request: &authz.ExecRequest{
				Grantee: "invalid_address",
				Msgs:    []sdk.Msg{types.NewMsgBuyAsset("cosmos1p9qh4ldfd6n0qehujsal4k7g0e37kel90rc4ts", "BTC", "1000000")},
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

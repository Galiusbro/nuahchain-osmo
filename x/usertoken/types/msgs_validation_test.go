package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMsgCreateUserToken_ValidateBasic(t *testing.T) {
	tests := []struct {
		name      string
		msg       MsgCreateUserToken
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid message",
			msg: MsgCreateUserToken{
				Creator:  "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				Subdenom: "mytoken",
				Name:     "My Token",
				Symbol:   "MTK",
				Decimals: 6,
			},
			expectErr: false,
		},
		{
			name: "invalid creator address",
			msg: MsgCreateUserToken{
				Creator:  "invalid-address",
				Subdenom: "mytoken",
				Name:     "My Token",
				Symbol:   "MTK",
				Decimals: 6,
			},
			expectErr: true,
			errMsg:    "invalid creator address",
		},
		{
			name: "empty subdenom",
			msg: MsgCreateUserToken{
				Creator:  "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				Subdenom: "",
				Name:     "My Token",
				Symbol:   "MTK",
				Decimals: 6,
			},
			expectErr: true,
			errMsg:    "subdenom cannot be empty",
		},
		{
			name: "subdenom too long",
			msg: MsgCreateUserToken{
				Creator:  "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				Subdenom: strings.Repeat("a", 45), // 45 characters
				Name:     "My Token",
				Symbol:   "MTK",
				Decimals: 6,
			},
			expectErr: true,
			errMsg:    "subdenom cannot be longer than 44 characters",
		},
		{
			name: "subdenom with invalid characters",
			msg: MsgCreateUserToken{
				Creator:  "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				Subdenom: "my@token",
				Name:     "My Token",
				Symbol:   "MTK",
				Decimals: 6,
			},
			expectErr: true,
			errMsg:    "subdenom can only contain alphanumeric characters and hyphens",
		},
		{
			name: "empty name",
			msg: MsgCreateUserToken{
				Creator:  "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				Subdenom: "mytoken",
				Name:     "",
				Symbol:   "MTK",
				Decimals: 6,
			},
			expectErr: true,
			errMsg:    "name cannot be empty",
		},
		{
			name: "name too long",
			msg: MsgCreateUserToken{
				Creator:  "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				Subdenom: "mytoken",
				Name:     strings.Repeat("a", 129), // 129 characters
				Symbol:   "MTK",
				Decimals: 6,
			},
			expectErr: true,
			errMsg:    "name cannot be longer than 128 characters",
		},
		{
			name: "empty symbol",
			msg: MsgCreateUserToken{
				Creator:  "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				Subdenom: "mytoken",
				Name:     "My Token",
				Symbol:   "",
				Decimals: 6,
			},
			expectErr: true,
			errMsg:    "symbol cannot be empty",
		},
		{
			name: "symbol too long",
			msg: MsgCreateUserToken{
				Creator:  "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				Subdenom: "mytoken",
				Name:     "My Token",
				Symbol:   strings.Repeat("A", 33), // 33 characters
				Decimals: 6,
			},
			expectErr: true,
			errMsg:    "symbol cannot be longer than 32 characters",
		},
		{
			name: "symbol with invalid characters",
			msg: MsgCreateUserToken{
				Creator:  "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				Subdenom: "mytoken",
				Name:     "My Token",
				Symbol:   "MT@K",
				Decimals: 6,
			},
			expectErr: true,
			errMsg:    "symbol can only contain alphanumeric characters",
		},
		{
			name: "decimals too high",
			msg: MsgCreateUserToken{
				Creator:  "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				Subdenom: "mytoken",
				Name:     "My Token",
				Symbol:   "MTK",
				Decimals: 19,
			},
			expectErr: true,
			errMsg:    "decimals cannot be greater than 18",
		},
		{
			name: "valid subdenom with hyphens",
			msg: MsgCreateUserToken{
				Creator:  "nuah1uwjmdjzspz2xa2g60dh6hq2h30jjmufutf099d",
				Subdenom: "my-token-123",
				Name:     "My Token",
				Symbol:   "MTK123",
				Decimals: 18,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgBuyTokens_ValidateBasic(t *testing.T) {
	tests := []struct {
		name      string
		msg       MsgBuyTokens
		expectErr bool
		errMsg    string
	}{
		{
			name: "invalid buyer address",
			msg: MsgBuyTokens{
				Buyer: "invalid-address",
				Denom: "factory/osmo1abc/mytoken",
			},
			expectErr: true,
			errMsg:    "invalid buyer address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

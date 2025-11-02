package authz

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	assetstypes "github.com/osmosis-labs/osmosis/v30/x/assets/types"
	bondingtypes "github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
)

type stubMsgClient struct {
	grantedTypes []string
	grantErr     error
}

func (s *stubMsgClient) Grant(ctx context.Context, msg *authztypes.MsgGrant, _ ...grpc.CallOption) (*authztypes.MsgGrantResponse, error) {
	authorization, err := msg.GetAuthorization()
	if err != nil {
		return nil, err
	}
	gen, ok := authorization.(*authztypes.GenericAuthorization)
	if !ok {
		return nil, fmt.Errorf("unexpected authorization type %T", authorization)
	}
	s.grantedTypes = append(s.grantedTypes, gen.Msg)
	if s.grantErr != nil {
		return nil, s.grantErr
	}
	return &authztypes.MsgGrantResponse{}, nil
}

func (s *stubMsgClient) Exec(context.Context, *authztypes.MsgExec, ...grpc.CallOption) (*authztypes.MsgExecResponse, error) {
	return nil, nil
}

func (s *stubMsgClient) Revoke(context.Context, *authztypes.MsgRevoke, ...grpc.CallOption) (*authztypes.MsgRevokeResponse, error) {
	return nil, nil
}

func testBech32Address(b byte) string {
	addr := sdk.AccAddress(bytes.Repeat([]byte{b}, 20))
	return addr.String()
}

func TestGrantTradingAuthorizations(t *testing.T) {
	granter := testBech32Address(1)
	grantee := testBech32Address(2)

	t.Run("assets only", func(t *testing.T) {
		stub := &stubMsgClient{}
		client := &Client{client: stub}
		granted, err := client.GrantTradingAuthorizations(context.Background(), granter, grantee, time.Time{}, false)
		require.NoError(t, err)
		require.Equal(t, []string{sdk.MsgTypeURL(&assetstypes.MsgBuyAsset{}), sdk.MsgTypeURL(&assetstypes.MsgSellAsset{})}, granted)
		require.Equal(t, granted, stub.grantedTypes)
	})

	t.Run("error when grant fails", func(t *testing.T) {
		stub := &stubMsgClient{grantErr: errors.New("boom")}
		client := &Client{client: stub}
		granted, err := client.GrantTradingAuthorizations(context.Background(), granter, grantee, time.Now().Add(48*time.Hour), false)
		require.Error(t, err)
		require.Len(t, granted, 0)
	})

	t.Run("include bonding", func(t *testing.T) {
		stub := &stubMsgClient{}
		client := &Client{client: stub}
		granted, err := client.GrantTradingAuthorizations(context.Background(), granter, grantee, time.Time{}, true)
		require.NoError(t, err)
		expected := []string{
			sdk.MsgTypeURL(&assetstypes.MsgBuyAsset{}),
			sdk.MsgTypeURL(&assetstypes.MsgSellAsset{}),
			sdk.MsgTypeURL(&bondingtypes.MsgBuyFromCurve{}),
			sdk.MsgTypeURL(&bondingtypes.MsgSellToCurve{}),
		}
		require.Equal(t, expected, granted)
		require.Equal(t, expected, stub.grantedTypes)
	})

	t.Run("invalid addresses", func(t *testing.T) {
		client := &Client{}
		_, err := client.GrantTradingAuthorizations(context.Background(), "", grantee, time.Time{}, false)
		require.Error(t, err)
		_, err = client.GrantTradingAuthorizations(context.Background(), granter, "bad", time.Time{}, false)
		require.Error(t, err)
	})
}

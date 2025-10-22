package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	assettypes "github.com/osmosis-labs/osmosis/v30/x/assets/types"
	leveragekeeper "github.com/osmosis-labs/osmosis/v30/x/leverage/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/leverage/types"
	oracletypes "github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

func TestQueryPosition(t *testing.T) {
	s := new(apptesting.KeeperTestHelper)
	s.SetT(t)
	s.Setup()

	ctx := s.Ctx
	owner := s.TestAccs[0]
	amount := sdkmath.NewInt(200)
	require.NoError(t, banktestutil.FundAccount(ctx, s.App.BankKeeper, owner, sdk.NewCoins(sdk.NewCoin(assettypes.NDollarDenom, amount))))
	s.App.OracleKeeper.SetPrice(ctx, &oracletypes.Price{Symbol: "ATOM", Value: "20"})

	msgSrv := leveragekeeper.NewMsgServerImpl(*s.App.LeverageKeeper)
	openResp, err := msgSrv.OpenPosition(sdk.WrapSDKContext(ctx), types.NewMsgOpenPosition(owner.String(), "ATOM", types.Side_SIDE_LONG, "200", "3"))
	require.NoError(t, err)

	querySrv := leveragekeeper.NewQueryServerImpl(*s.App.LeverageKeeper)
	resp, err := querySrv.Position(sdk.WrapSDKContext(ctx), &types.QueryPositionRequest{Id: openResp.Position.Id})
	require.NoError(t, err)
	require.Equal(t, openResp.Position.Id, resp.Position.Id)
}

func TestQueryPositionNotFound(t *testing.T) {
	s := new(apptesting.KeeperTestHelper)
	s.SetT(t)
	s.Setup()

	querySrv := leveragekeeper.NewQueryServerImpl(*s.App.LeverageKeeper)
	_, err := querySrv.Position(sdk.WrapSDKContext(s.Ctx), &types.QueryPositionRequest{Id: 42})
	require.Error(t, err)
}

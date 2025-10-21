package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/collateral/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/collateral/types"
)

func TestQueryCollateral(t *testing.T) {
	s := new(apptesting.KeeperTestHelper)
	s.SetT(t)
	s.Setup()

	ctx := s.Ctx
	acc := s.TestAccs[0]
	coin := sdk.NewCoin("uosmo", sdkmath.NewInt(500))
	require.NoError(t, banktestutil.FundAccount(ctx, s.App.BankKeeper, acc, sdk.NewCoins(coin)))

	msgSrv := keeper.NewMsgServerImpl(*s.App.CollateralKeeper)
	_, err := msgSrv.Deposit(sdk.WrapSDKContext(ctx), types.NewMsgDeposit(acc.String(), "500uosmo"))
	require.NoError(t, err)

	querySrv := keeper.NewQueryServerImpl(*s.App.CollateralKeeper)
	resp, err := querySrv.Collateral(sdk.WrapSDKContext(ctx), &types.QueryCollateralRequest{Owner: acc.String()})
	require.NoError(t, err)
	require.Len(t, resp.Positions, 1)
	require.Equal(t, "500", resp.Positions[0].Amount)
}

func TestQueryCollateralNoPositions(t *testing.T) {
	s := new(apptesting.KeeperTestHelper)
	s.SetT(t)
	s.Setup()

	ctx := s.Ctx
	acc := s.TestAccs[0]

	querySrv := keeper.NewQueryServerImpl(*s.App.CollateralKeeper)
	resp, err := querySrv.Collateral(sdk.WrapSDKContext(ctx), &types.QueryCollateralRequest{Owner: acc.String()})
	require.NoError(t, err)
	require.Len(t, resp.Positions, 0)
}

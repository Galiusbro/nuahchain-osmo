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

func TestMsgServerDepositWithdraw(t *testing.T) {
	s := new(apptesting.KeeperTestHelper)
	s.SetT(t)
	s.Setup()

	ctx := s.Ctx
	acc := s.TestAccs[0]
	coin := sdk.NewCoin("uosmo", sdkmath.NewInt(1_000))
	require.NoError(t, banktestutil.FundAccount(ctx, s.App.BankKeeper, acc, sdk.NewCoins(coin)))

	srv := keeper.NewMsgServerImpl(*s.App.CollateralKeeper)
	goCtx := sdk.WrapSDKContext(ctx)

	_, err := srv.Deposit(goCtx, types.NewMsgDeposit(acc.String(), "1000uosmo"))
	require.NoError(t, err)

	_, err = srv.Withdraw(goCtx, types.NewMsgWithdraw(acc.String(), "400uosmo"))
	require.NoError(t, err)

	positions := s.App.CollateralKeeper.GetPositions(ctx, acc)
	require.Len(t, positions, 1)
	require.Equal(t, "600", positions[0].Amount)
}

func TestMsgServerWithdrawInsufficient(t *testing.T) {
	s := new(apptesting.KeeperTestHelper)
	s.SetT(t)
	s.Setup()

	ctx := s.Ctx
	acc := s.TestAccs[0]
	coin := sdk.NewCoin("uosmo", sdkmath.NewInt(200))
	require.NoError(t, banktestutil.FundAccount(ctx, s.App.BankKeeper, acc, sdk.NewCoins(coin)))

	srv := keeper.NewMsgServerImpl(*s.App.CollateralKeeper)
	goCtx := sdk.WrapSDKContext(ctx)

	_, err := srv.Deposit(goCtx, types.NewMsgDeposit(acc.String(), "200uosmo"))
	require.NoError(t, err)

	_, err = srv.Withdraw(goCtx, types.NewMsgWithdraw(acc.String(), "300uosmo"))
	require.Error(t, err)
}

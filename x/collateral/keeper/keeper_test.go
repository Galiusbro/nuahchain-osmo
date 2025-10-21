package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/collateral/types"
)

func TestKeeperDepositWithdraw(t *testing.T) {
	s := new(apptesting.KeeperTestHelper)
	s.SetT(t)
	s.Setup()

	ctx := s.Ctx
	acc := s.TestAccs[0]
	coin := sdk.NewCoin("uosmo", sdkmath.NewInt(1_000))

	require.NoError(t, banktestutil.FundAccount(ctx, s.App.BankKeeper, acc, sdk.NewCoins(coin)))

	k := *s.App.CollateralKeeper
	require.NoError(t, k.DepositCollateral(ctx, acc, coin))

	positions := k.GetPositions(ctx, acc)
	require.Len(t, positions, 1)
	require.Equal(t, coin.Amount.String(), positions[0].Amount)
	require.Equal(t, coin.Denom, positions[0].Denom)

	moduleAddr := s.App.AccountKeeper.GetModuleAddress(types.ModuleName)
	moduleBalance := s.App.BankKeeper.GetBalance(ctx, moduleAddr, coin.Denom)
	require.Equal(t, coin.Amount, moduleBalance.Amount)

	withdrawCoin := sdk.NewCoin("uosmo", sdkmath.NewInt(400))
	require.NoError(t, k.WithdrawCollateral(ctx, acc, withdrawCoin))

	positions = k.GetPositions(ctx, acc)
	require.Len(t, positions, 1)
	require.Equal(t, sdkmath.NewInt(600).String(), positions[0].Amount)

	// withdrawing more than available should fail
	tooMuch := sdk.NewCoin("uosmo", sdkmath.NewInt(700))
	require.Error(t, k.WithdrawCollateral(ctx, acc, tooMuch))
}

func TestKeeperGenesis(t *testing.T) {
	s := new(apptesting.KeeperTestHelper)
	s.SetT(t)
	s.Setup()

	ctx := s.Ctx
	k := *s.App.CollateralKeeper

	owner := s.TestAccs[0]
	gen := &types.GenesisState{
		Positions: []*types.Position{
			{Owner: owner.String(), Denom: "uosmo", Amount: "250"},
		},
	}

	k.InitGenesis(ctx, gen)

	// Check that the position was actually stored
	positions := k.GetPositions(ctx, owner)
	require.Len(t, positions, 1)
	require.Equal(t, "250", positions[0].Amount)

	exported := k.ExportGenesis(ctx)
	require.Len(t, exported.Positions, 1)
	require.Equal(t, "250", exported.Positions[0].Amount)
}

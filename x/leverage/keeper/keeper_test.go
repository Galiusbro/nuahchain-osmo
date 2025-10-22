package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	leveragekeeper "github.com/osmosis-labs/osmosis/v30/x/leverage/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/leverage/types"
	oracletypes "github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

func TestKeeperOpenClosePosition(t *testing.T) {
	s := new(apptesting.KeeperTestHelper)
	s.SetT(t)
	s.Setup()

	ctx := s.Ctx
	owner := s.TestAccs[0]

	amount := sdkmath.NewInt(1000)
	require.NoError(t, banktestutil.FundAccount(ctx, s.App.BankKeeper, owner, sdk.NewCoins(sdk.NewCoin(types.NDollarDenom, amount))))

	s.App.OracleKeeper.SetPrice(ctx, &oracletypes.Price{Symbol: "BTC", Value: "100"})

	msgSrv := leveragekeeper.NewMsgServerImpl(*s.App.LeverageKeeper)
	openResp, err := msgSrv.OpenPosition(sdk.WrapSDKContext(ctx), types.NewMsgOpenPosition(owner.String(), "BTC", types.Side_SIDE_LONG, "1000", "2"))
	require.NoError(t, err)
	position := openResp.Position
	require.Equal(t, types.Side_SIDE_LONG, position.Side)
	baseQty, err := sdkmath.LegacyNewDecFromStr(position.BaseQty)
	require.NoError(t, err)
	expectedQty, err := sdkmath.LegacyNewDecFromStr("10")
	require.NoError(t, err)
	require.True(t, baseQty.Equal(expectedQty))

	moduleAddr := s.App.AccountKeeper.GetModuleAddress(types.ModuleName)
	moduleBalance := s.App.BankKeeper.GetBalance(ctx, moduleAddr, types.NDollarDenom)
	require.Equal(t, amount, moduleBalance.Amount)

	closeResp, err := msgSrv.ClosePosition(sdk.WrapSDKContext(ctx), types.NewMsgClosePosition(owner.String(), position.Id))
	require.NoError(t, err)
	require.Equal(t, "0", closeResp.Pnl)

	_, found := s.App.LeverageKeeper.GetPosition(ctx, position.Id)
	require.False(t, found)

	moduleBalance = s.App.BankKeeper.GetBalance(ctx, moduleAddr, types.NDollarDenom)
	require.True(t, moduleBalance.Amount.IsZero())
}

func TestKeeperClosePositionUnauthorized(t *testing.T) {
	s := new(apptesting.KeeperTestHelper)
	s.SetT(t)
	s.Setup()

	ctx := s.Ctx
	owner := s.TestAccs[0]
	other := s.TestAccs[1]

	amount := sdkmath.NewInt(500)
	require.NoError(t, banktestutil.FundAccount(ctx, s.App.BankKeeper, owner, sdk.NewCoins(sdk.NewCoin(types.NDollarDenom, amount))))

	s.App.OracleKeeper.SetPrice(ctx, &oracletypes.Price{Symbol: "ETH", Value: "50"})

	msgSrv := leveragekeeper.NewMsgServerImpl(*s.App.LeverageKeeper)
	openResp, err := msgSrv.OpenPosition(sdk.WrapSDKContext(ctx), types.NewMsgOpenPosition(owner.String(), "ETH", types.Side_SIDE_SHORT, "500", "1.5"))
	require.NoError(t, err)

	_, err = msgSrv.ClosePosition(sdk.WrapSDKContext(ctx), types.NewMsgClosePosition(other.String(), openResp.Position.Id))
	require.Error(t, err)
}

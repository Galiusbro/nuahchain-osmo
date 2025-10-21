package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	osmomath "github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/x/assets/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
	feetypes "github.com/osmosis-labs/osmosis/v30/x/fees/types"
	oracletypes "github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

func TestBuyAssetAppliesFeeAndBurnsRemainder(t *testing.T) {
	s := new(apptesting.KeeperTestHelper)
	s.SetT(t)
	s.Setup()

	ctx := s.Ctx
	buyer := s.TestAccs[0]

	startBalance := sdkmath.NewInt(5000)
	require.NoError(t, banktestutil.FundAccount(ctx, s.App.BankKeeper, buyer, sdk.NewCoins(sdk.NewCoin(types.NDollarDenom, startBalance))))

	require.NoError(t, s.App.FeesKeeper.SetParams(ctx, feetypes.NewParams("0.1")))
	s.App.OracleKeeper.SetPrice(ctx, &oracletypes.Price{Symbol: "GOLD", Value: "2000"})

	srv := keeper.NewMsgServer(*s.App.AssetsKeeper)
	resp, err := srv.BuyAsset(sdk.WrapSDKContext(ctx), types.NewMsgBuyAsset(buyer.String(), "GOLD", "1000"))
	require.NoError(t, err)

	feeRate := osmomath.MustNewDecFromStr("0.1")
	amountND := sdkmath.NewInt(1000)
	feeInt := feeRate.Mul(osmomath.NewDecFromInt(amountND)).TruncateInt() // 100
	netND := amountND.Sub(feeInt)                                         // 900

	baseDec := osmomath.MustNewDecFromStr(resp.BaseAmount)
	expectedBase := osmomath.NewDecFromInt(netND).Quo(osmomath.MustNewDecFromStr("2000"))
	tolerance := osmomath.MustNewDecFromStr("0.000000000000000001")
	require.True(t, baseDec.Sub(expectedBase).Abs().LTE(tolerance))

	expectedAsset := expectedBase.Mul(osmomath.NewDecFromInt(types.AssetPrecisionFactor())).TruncateInt()
	assetBalance := s.App.BankKeeper.GetBalance(ctx, buyer, types.AssetDenom("GOLD"))
	require.Equal(t, expectedAsset, assetBalance.Amount)

	remainingND := s.App.BankKeeper.GetBalance(ctx, buyer, types.NDollarDenom)
	require.Equal(t, startBalance.Sub(amountND), remainingND.Amount)

	feeAddr := s.App.AccountKeeper.GetModuleAddress(feetypes.ModuleName)
	feeBalance := s.App.BankKeeper.GetBalance(ctx, feeAddr, types.NDollarDenom)
	require.Equal(t, feeInt, feeBalance.Amount)

	stats := s.App.StablecoinKeeper.GetStats(ctx)
	minted := mustInt(stats.TotalMinted)
	burned := mustInt(stats.TotalBurned)
	require.True(t, minted.IsZero())
	require.Equal(t, netND, burned)
}

func TestSellAssetAppliesFeeAndRewardsUser(t *testing.T) {
	s := new(apptesting.KeeperTestHelper)
	s.SetT(t)
	s.Setup()

	ctx := s.Ctx
	trader := s.TestAccs[0]

	startBalance := sdkmath.NewInt(5000)
	require.NoError(t, banktestutil.FundAccount(ctx, s.App.BankKeeper, trader, sdk.NewCoins(sdk.NewCoin(types.NDollarDenom, startBalance))))

	require.NoError(t, s.App.FeesKeeper.SetParams(ctx, feetypes.NewParams("0.1")))
	s.App.OracleKeeper.SetPrice(ctx, &oracletypes.Price{Symbol: "GOLD", Value: "2000"})

	srv := keeper.NewMsgServer(*s.App.AssetsKeeper)
	buyResp, err := srv.BuyAsset(sdk.WrapSDKContext(ctx), types.NewMsgBuyAsset(trader.String(), "GOLD", "1000"))
	require.NoError(t, err)

	s.App.OracleKeeper.SetPrice(ctx, &oracletypes.Price{Symbol: "GOLD", Value: "3000"})
	sellResp, err := srv.SellAsset(sdk.WrapSDKContext(ctx), types.NewMsgSellAsset(trader.String(), "GOLD", buyResp.BaseAmount))
	require.NoError(t, err)

	feeRate := osmomath.MustNewDecFromStr("0.1")
	amountND := sdkmath.NewInt(1000)
	feeBuy := feeRate.Mul(osmomath.NewDecFromInt(amountND)).TruncateInt() // 100
	netBuy := amountND.Sub(feeBuy)                                        // 900

	baseDec := osmomath.MustNewDecFromStr(buyResp.BaseAmount)
	expectedBase := osmomath.NewDecFromInt(netBuy).Quo(osmomath.MustNewDecFromStr("2000"))
	tolerance := osmomath.MustNewDecFromStr("0.000000000000000001")
	require.True(t, baseDec.Sub(expectedBase).Abs().LTE(tolerance))

	payoutDec := expectedBase.Mul(osmomath.MustNewDecFromStr("3000"))
	payoutInt := payoutDec.TruncateInt()            // 1350
	feeSell := payoutDec.Mul(feeRate).TruncateInt() // 135
	netSell := payoutInt.Sub(feeSell)               // 1215
	require.Equal(t, netSell.String(), sellResp.Payout_NDOLLAR)

	finalNDBalance := s.App.BankKeeper.GetBalance(ctx, trader, types.NDollarDenom)
	expectedND := startBalance.Sub(amountND).Add(netSell)
	require.Equal(t, expectedND, finalNDBalance.Amount)

	assetBalance := s.App.BankKeeper.GetBalance(ctx, trader, types.AssetDenom("GOLD"))
	require.True(t, assetBalance.Amount.IsZero())

	feeAddr := s.App.AccountKeeper.GetModuleAddress(feetypes.ModuleName)
	feeBalance := s.App.BankKeeper.GetBalance(ctx, feeAddr, types.NDollarDenom)
	require.Equal(t, feeBuy.Add(feeSell), feeBalance.Amount)

	stats := s.App.StablecoinKeeper.GetStats(ctx)
	minted := mustInt(stats.TotalMinted)
	burned := mustInt(stats.TotalBurned)
	require.Equal(t, payoutInt, minted)
	require.Equal(t, sdkmath.NewInt(900), burned)
}

func mustInt(value string) sdkmath.Int {
	if value == "" {
		return sdkmath.ZeroInt()
	}
	intVal, ok := sdkmath.NewIntFromString(value)
	if !ok {
		panic("invalid integer string")
	}
	return intVal
}

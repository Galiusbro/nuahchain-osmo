package keeper_test

import (
	"fmt"
	"strings"
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
	stablecoinkeeper "github.com/osmosis-labs/osmosis/v30/x/stablecoin/keeper"
	stablecointypes "github.com/osmosis-labs/osmosis/v30/x/stablecoin/types"
)

type mockOracleKeeper struct {
	prices map[string]oracletypes.Price
}

func newMockOracleKeeper() *mockOracleKeeper {
	return &mockOracleKeeper{prices: make(map[string]oracletypes.Price)}
}

func (m *mockOracleKeeper) SetStaticPrice(ctx sdk.Context, symbol, value string) {
	clean := strings.TrimSpace(symbol)
	if clean == "" {
		panic("mock oracle symbol cannot be empty")
	}
	m.prices[clean] = oracletypes.Price{
		Symbol:     clean,
		Value:      value,
		Source:     "mock",
		Timestamp:  ctx.BlockTime().Unix(),
		Confidence: 1,
	}
}

func (m *mockOracleKeeper) GetPrice(_ sdk.Context, symbol string) (*oracletypes.Price, bool) {
	price, ok := m.prices[strings.TrimSpace(symbol)]
	if !ok {
		return nil, false
	}
	priceCopy := price
	return &priceCopy, true
}

func (m *mockOracleKeeper) GetPriceWithFallback(ctx sdk.Context, symbol string) (*oracletypes.Price, bool) {
	return m.GetPrice(ctx, symbol)
}

func (m *mockOracleKeeper) EnsureFreshPrice(ctx sdk.Context, symbol string) (*oracletypes.Price, error) {
	clean := strings.TrimSpace(symbol)
	if clean == "" {
		return nil, fmt.Errorf("mock oracle: invalid symbol")
	}
	price, ok := m.prices[clean]
	if !ok {
		return nil, fmt.Errorf("mock oracle: price for %s not set", clean)
	}
	price.Timestamp = ctx.BlockTime().Unix()
	m.prices[clean] = price
	priceCopy := price
	return &priceCopy, nil
}

func TestBuyAssetAppliesFeeAndBurnsRemainder(t *testing.T) {
	s := new(apptesting.KeeperTestHelper)
	s.SetT(t)
	s.Setup()

	ctx := s.Ctx
	buyer := s.TestAccs[0]

	startBalance := sdkmath.NewInt(5000)
	require.NoError(t, banktestutil.FundAccount(ctx, s.App.BankKeeper, buyer, sdk.NewCoins(sdk.NewCoin(types.NDollarDenom, startBalance))))

	require.NoError(t, s.App.FeesKeeper.SetParams(ctx, feetypes.NewParams("0.1")))
	feeRate := s.App.FeesKeeper.GetTradeFeeRate(ctx)

	oracleKeeper := newMockOracleKeeper()
	oracleKeeper.SetStaticPrice(ctx, "GOLD", "2000")
	assetsKeeper := keeper.NewKeeper(
		s.App.AppCodec(),
		s.App.GetKey(types.StoreKey),
		s.App.BankKeeper,
		oracleKeeper,
		s.App.FeesKeeper,
		s.App.StablecoinKeeper,
	)

	srv := keeper.NewMsgServer(assetsKeeper)
	resp, err := srv.BuyAsset(ctx, types.NewMsgBuyAsset(buyer.String(), "GOLD", "1000"))
	require.NoError(t, err)

	amountND := sdkmath.NewInt(1000)
	feeInt := feeRate.Mul(osmomath.NewDecFromInt(amountND)).TruncateInt()
	netND := amountND.Sub(feeInt)
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
	feeRate := s.App.FeesKeeper.GetTradeFeeRate(ctx)

	oracleKeeper := newMockOracleKeeper()
	oracleKeeper.SetStaticPrice(ctx, "GOLD", "2000")
	assetsKeeper := keeper.NewKeeper(
		s.App.AppCodec(),
		s.App.GetKey(types.StoreKey),
		s.App.BankKeeper,
		oracleKeeper,
		s.App.FeesKeeper,
		s.App.StablecoinKeeper,
	)

	srv := keeper.NewMsgServer(assetsKeeper)
	buyResp, err := srv.BuyAsset(ctx, types.NewMsgBuyAsset(trader.String(), "GOLD", "1000"))
	require.NoError(t, err)

	oracleKeeper.SetStaticPrice(ctx, "GOLD", "3000")
	sellResp, err := srv.SellAsset(ctx, types.NewMsgSellAsset(trader.String(), "GOLD", buyResp.BaseAmount))
	require.NoError(t, err)

	amountND := sdkmath.NewInt(1000)
	feeBuy := feeRate.Mul(osmomath.NewDecFromInt(amountND)).TruncateInt()
	netBuy := amountND.Sub(feeBuy)

	baseDec := osmomath.MustNewDecFromStr(buyResp.BaseAmount)
	expectedBase := osmomath.NewDecFromInt(netBuy).Quo(osmomath.MustNewDecFromStr("2000"))
	tolerance := osmomath.MustNewDecFromStr("0.000000000000000001")
	require.True(t, baseDec.Sub(expectedBase).Abs().LTE(tolerance))

	payoutDec := expectedBase.Mul(osmomath.MustNewDecFromStr("3000"))
	payoutInt := payoutDec.TruncateInt()
	feeSell := payoutDec.Mul(feeRate).TruncateInt()
	netSell := payoutInt.Sub(feeSell)
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
	require.Equal(t, netBuy, burned)
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

func TestStablecoinCoverageTracksTrades(t *testing.T) {
	s := new(apptesting.KeeperTestHelper)
	s.SetT(t)
	s.Setup()

	ctx := s.Ctx
	trader := s.TestAccs[0]

	startBalance := sdkmath.NewInt(5000)
	require.NoError(t, banktestutil.FundAccount(ctx, s.App.BankKeeper, trader, sdk.NewCoins(sdk.NewCoin(types.NDollarDenom, startBalance))))

	require.NoError(t, s.App.FeesKeeper.SetParams(ctx, feetypes.NewParams("0.1")))
	feeRate := s.App.FeesKeeper.GetTradeFeeRate(ctx)

	oracleKeeper := newMockOracleKeeper()
	oracleKeeper.SetStaticPrice(ctx, "GOLD", "2000")
	assetsKeeper := keeper.NewKeeper(
		s.App.AppCodec(),
		s.App.GetKey(types.StoreKey),
		s.App.BankKeeper,
		oracleKeeper,
		s.App.FeesKeeper,
		s.App.StablecoinKeeper,
	)

	assetSrv := keeper.NewMsgServer(assetsKeeper)
	amountND := sdkmath.NewInt(1000)
	buyResp, err := assetSrv.BuyAsset(ctx, types.NewMsgBuyAsset(trader.String(), "GOLD", amountND.String()))
	require.NoError(t, err)

	stableQuery := stablecoinkeeper.NewQueryServer(*s.App.StablecoinKeeper)
	coverageAfterBuy, err := stableQuery.Coverage(ctx, &stablecointypes.QueryCoverageRequest{})
	require.NoError(t, err)
	feeBuy := feeRate.Mul(osmomath.NewDecFromInt(amountND)).TruncateInt()
	netBuy := amountND.Sub(feeBuy)
	require.Equal(t, netBuy.Neg().String(), coverageAfterBuy.Outstanding)
	require.Equal(t, feeBuy.String(), coverageAfterBuy.ReserveBalance)
	require.Equal(t, sdkmath.LegacyZeroDec().String(), coverageAfterBuy.CoverageRatio)

	oracleKeeper.SetStaticPrice(ctx, "GOLD", "3000")
	_, err = assetSrv.SellAsset(ctx, types.NewMsgSellAsset(trader.String(), "GOLD", buyResp.BaseAmount))
	require.NoError(t, err)

	coverageAfterSell, err := stableQuery.Coverage(ctx, &stablecointypes.QueryCoverageRequest{})
	require.NoError(t, err)
	baseDec := osmomath.MustNewDecFromStr(buyResp.BaseAmount)
	payoutDec := baseDec.Mul(osmomath.MustNewDecFromStr("3000"))
	payoutInt := payoutDec.TruncateInt()
	feeSell := payoutDec.Mul(feeRate).TruncateInt()
	expectedOutstanding := payoutInt.Sub(netBuy)
	expectedReserve := feeBuy.Add(feeSell)
	require.Equal(t, expectedOutstanding.String(), coverageAfterSell.Outstanding)
	require.Equal(t, expectedReserve.String(), coverageAfterSell.ReserveBalance)
	outstandingInt, ok := sdkmath.NewIntFromString(coverageAfterSell.Outstanding)
	require.True(t, ok)
	reserveInt, ok := sdkmath.NewIntFromString(coverageAfterSell.ReserveBalance)
	require.True(t, ok)
	expectedRatio := osmomath.NewDecFromInt(reserveInt).Quo(osmomath.NewDecFromInt(outstandingInt)).String()
	require.Equal(t, expectedRatio, coverageAfterSell.CoverageRatio)
}

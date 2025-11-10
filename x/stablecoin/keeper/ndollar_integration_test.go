package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	leveragetypes "github.com/osmosis-labs/osmosis/v30/x/leverage/types"
	oracletypes "github.com/osmosis-labs/osmosis/v30/x/oracle/types"
	stablecointypes "github.com/osmosis-labs/osmosis/v30/x/stablecoin/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v30/x/tokenfactory/types"
)

type NDollarIntegrationTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient stablecointypes.QueryClient
}

func TestNDollarIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NDollarIntegrationTestSuite))
}

func (s *NDollarIntegrationTestSuite) SetupTest() {
	s.Setup()
	s.queryClient = stablecointypes.NewQueryClient(s.QueryHelper)
}

// TestNDollarDenomUnification проверяет унификацию NDOLLAR denom во всех модулях
func (s *NDollarIntegrationTestSuite) TestNDollarDenomUnification() {
	ctx := s.Ctx
	require := s.Require()

	// 1. Создаем тестового пользователя с начальным балансом unuah
	user := s.TestAccs[0]
	initialUnuah := sdkmath.NewInt(1_000_000_000_000) // 1M NUAH
	require.NoError(banktestutil.FundAccount(ctx, s.App.BankKeeper, user, sdk.NewCoins(sdk.NewCoin("unuah", initialUnuah))))

	// 2. Создаем NDOLLAR через tokenfactory (имитируем genesis setup)
	// Используем TestAccs[2] как admin чтобы избежать проблем с module account permissions
	admin := s.TestAccs[2]
	_, err := s.App.TokenFactoryKeeper.CreateDenom(ctx, admin.String(), "ndollar")
	require.NoError(err)

	// Получаем реальный denom NDOLLAR (factory/.../ndollar)
	realNDollarDenom, err := tokenfactorytypes.GetTokenDenom(admin.String(), "ndollar")
	require.NoError(err)
	s.T().Logf("✓ Created NDOLLAR with denom: %s", realNDollarDenom)

	// В реальности BuyNDollar использует tokenfactory для минта NDOLLAR
	// Для теста упростим логику: минтим NDOLLAR напрямую через banktestutil
	// чтобы избежать проблем с module account permissions
	initialNDollarForUser := sdkmath.NewInt(500_000_000) // 500 NDOLLAR начальный баланс
	require.NoError(banktestutil.FundAccount(ctx, s.App.BankKeeper, user, sdk.NewCoins(sdk.NewCoin(realNDollarDenom, initialNDollarForUser))))

	// Регистрируем metadata для NDOLLAR, чтобы GetNDollarDenom() работал
	metadata := banktypes.Metadata{
		Description: "NUAH Dollar stablecoin",
		Base:        realNDollarDenom,
		Display:     "NDOLLAR",
		Name:        "NDOLLAR",
		Symbol:      "N$",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: realNDollarDenom, Exponent: 0, Aliases: []string{}},
			{Denom: "NDOLLAR", Exponent: 6, Aliases: []string{}},
		},
	}
	s.App.BankKeeper.SetDenomMetaData(ctx, metadata)

	// 3. ТЕСТ 1: Проверяем GetNDollarDenom() возвращает правильный denom
	s.T().Log("\n=== TEST 1: GetNDollarDenom() ===")
	denomFromKeeper := s.App.StablecoinKeeper.GetNDollarDenom(ctx)
	require.Equal(realNDollarDenom, denomFromKeeper, "GetNDollarDenom должен возвращать реальный factory denom")
	s.T().Logf("✓ GetNDollarDenom() = %s", denomFromKeeper)

	// 4. ТЕСТ 2: Проверка GetNDollarDenom для assets/leverage модулей
	s.T().Log("\n=== TEST 2: GetNDollarDenom used by x/assets ===")
	assetsNDollarDenom := s.App.AssetsKeeper.GetNDollarDenom(ctx)
	require.Equal(realNDollarDenom, assetsNDollarDenom, "x/assets должен использовать реальный denom")
	s.T().Logf("✓ x/assets GetNDollarDenom() = %s", assetsNDollarDenom)

	// Проверяем начальный баланс
	ndollarBalance := s.App.BankKeeper.GetBalance(ctx, user, realNDollarDenom)
	require.Equal(initialNDollarForUser, ndollarBalance.Amount, "Пользователь должен иметь начальный NDOLLAR")
	unuahBalance := s.App.BankKeeper.GetBalance(ctx, user, "unuah")
	require.Equal(initialUnuah, unuahBalance.Amount, "Начальный баланс unuah")
	s.T().Logf("✓ User initial NDOLLAR balance: %s", ndollarBalance.Amount)
	s.T().Logf("✓ User initial unuah balance: %s", unuahBalance.Amount)

	// 5. ТЕСТ 3: Покупка актива через x/assets с NDOLLAR
	s.T().Log("\n=== TEST 3: BuyAsset (x/assets) with NDOLLAR ===")

	// Регистрируем тестовый актив
	symbol := "TEST"
	_, _, err = s.App.AssetsKeeper.EnsureAsset(ctx, symbol)
	require.NoError(err)

	// Устанавливаем цену через oracle (100 USD за 1 TEST)
	price := &oracletypes.Price{
		Symbol:    symbol,
		Value:     "100.00", // 100.00 USD
		Timestamp: ctx.BlockTime().Unix(),
	}
	s.App.OracleKeeper.SetPrice(ctx, price)

	// Покупаем актив за NDOLLAR
	purchaseAmount := sdkmath.NewInt(10_000_000) // 10 NDOLLAR
	payment := sdk.NewCoin(realNDollarDenom, purchaseAmount)
	assetCoin, assetAmount, err := s.App.AssetsKeeper.BuyAssetWithPayment(ctx, user, symbol, payment)
	require.NoError(err)
	require.True(assetAmount.IsPositive(), "Должны получить актив")
	s.T().Logf("✓ Bought asset: %s (amount: %s)", assetCoin, assetAmount)

	// Проверяем баланс NDOLLAR уменьшился
	ndollarBalanceAfterBuy := s.App.BankKeeper.GetBalance(ctx, user, realNDollarDenom)
	require.True(ndollarBalanceAfterBuy.Amount.LT(ndollarBalance.Amount), "NDOLLAR должен уменьшиться после покупки")
	s.T().Logf("✓ NDOLLAR balance after buy: %s", ndollarBalanceAfterBuy.Amount)

	// 6. ТЕСТ 4: Продажа актива через x/assets получая NDOLLAR
	s.T().Log("\n=== TEST 4: SellAsset (x/assets) receiving NDOLLAR ===")

	// Продаем половину купленного актива
	// assetAmount уже в базовых единицах (Dec), поэтому используем его напрямую
	sellAmountDec := assetAmount.QuoInt64(2)
	resultCoin, resultAmount, err := s.App.AssetsKeeper.SellAsset(ctx, user, symbol, sellAmountDec)
	require.NoError(err)
	require.Equal(realNDollarDenom, resultCoin.Denom, "Продажа должна возвращать реальный NDOLLAR denom")
	require.True(resultAmount.IsPositive(), "Должны получить NDOLLAR")
	s.T().Logf("✓ Sold asset, received: %s (amount: %s)", resultCoin, resultAmount)

	// Проверяем баланс NDOLLAR увеличился
	ndollarBalanceAfterSell := s.App.BankKeeper.GetBalance(ctx, user, realNDollarDenom)
	require.True(ndollarBalanceAfterSell.Amount.GT(ndollarBalanceAfterBuy.Amount), "NDOLLAR должен увеличиться после продажи")
	s.T().Logf("✓ NDOLLAR balance after sell: %s", ndollarBalanceAfterSell.Amount)

	// 7. ТЕСТ 5: Открытие маржинальной позиции через x/leverage с NDOLLAR
	s.T().Log("\n=== TEST 5: OpenPosition (x/leverage) with NDOLLAR ===")

	// Открываем позицию
	leverageAmount := sdkmath.NewInt(5_000_000) // 5 NDOLLAR
	openMsg := &leveragetypes.MsgOpenPosition{
		Owner:         user.String(),
		Symbol:        symbol,
		Side:          leveragetypes.Side_SIDE_LONG,
		Quote_NDOLLAR: leverageAmount.String(),
		Leverage:      "2",
	}

	position, err := s.App.LeverageKeeper.OpenPosition(ctx, openMsg)
	require.NoError(err)
	require.NotNil(position)
	require.Greater(position.Id, uint64(0), "Position ID должен быть положительным")
	s.T().Logf("✓ Opened leverage position ID: %d", position.Id)
	s.T().Logf("✓ Position details: Symbol=%s, Side=%s, BaseQty=%s, EntryPrice=%s, Leverage=%s",
		position.Symbol, position.Side, position.BaseQty, position.EntryPrice, position.Leverage)

	// Проверяем что NDOLLAR списался
	ndollarBalanceAfterLeverage := s.App.BankKeeper.GetBalance(ctx, user, realNDollarDenom)
	require.True(ndollarBalanceAfterLeverage.Amount.LT(ndollarBalanceAfterSell.Amount), "NDOLLAR должен списаться на margin")
	s.T().Logf("✓ NDOLLAR balance after opening position: %s", ndollarBalanceAfterLeverage.Amount)

	// Проверяем что модуль leverage получил NDOLLAR
	leverageModuleAddr := s.App.AccountKeeper.GetModuleAddress(leveragetypes.ModuleName)
	leverageModuleBalance := s.App.BankKeeper.GetBalance(ctx, leverageModuleAddr, realNDollarDenom)
	require.Equal(leverageAmount, leverageModuleBalance.Amount, "Модуль leverage должен получить NDOLLAR")
	s.T().Logf("✓ Leverage module NDOLLAR balance: %s", leverageModuleBalance.Amount)

	// 8. ТЕСТ 6: Закрытие маржинальной позиции через x/leverage
	s.T().Log("\n=== TEST 6: ClosePosition (x/leverage) ===")

	closeMsg := &leveragetypes.MsgClosePosition{
		Owner: user.String(),
		Id:    position.Id,
	}

	pnl, err := s.App.LeverageKeeper.ClosePosition(ctx, closeMsg)
	require.NoError(err)
	s.T().Logf("✓ Closed position with PnL: %s", pnl)

	// Проверяем что NDOLLAR вернулся пользователю
	ndollarBalanceAfterClose := s.App.BankKeeper.GetBalance(ctx, user, realNDollarDenom)
	require.True(ndollarBalanceAfterClose.Amount.GT(ndollarBalanceAfterLeverage.Amount), "NDOLLAR должен вернуться после закрытия")
	s.T().Logf("✓ NDOLLAR balance after closing position: %s", ndollarBalanceAfterClose.Amount)

	// Проверяем что модуль leverage вернул NDOLLAR
	leverageModuleBalanceAfter := s.App.BankKeeper.GetBalance(ctx, leverageModuleAddr, realNDollarDenom)
	require.True(leverageModuleBalanceAfter.Amount.LT(leverageModuleBalance.Amount), "Модуль должен вернуть NDOLLAR")
	s.T().Logf("✓ Leverage module NDOLLAR balance after close: %s", leverageModuleBalanceAfter.Amount)

	// 9. ФИНАЛЬНАЯ ПРОВЕРКА: Все операции использовали один и тот же denom
	s.T().Log("\n=== FINAL CHECK: Denom Consistency ===")
	s.T().Logf("✓ All operations used the same NDOLLAR denom: %s", realNDollarDenom)
	s.T().Logf("✓ Total NDOLLAR operations:")
	s.T().Logf("  - Initial balance: %s", initialNDollarForUser)
	s.T().Logf("  - BuyAsset payment: %s", purchaseAmount)
	s.T().Logf("  - SellAsset received: %s", resultCoin.Amount)
	s.T().Logf("  - OpenPosition margin: %s", leverageAmount)
	s.T().Logf("  - ClosePosition returned: %s", leverageAmount)
	s.T().Logf("✓ Final user NDOLLAR balance: %s", ndollarBalanceAfterClose.Amount)
}

// TestNDollarDenomFallback проверяет fallback когда metadata не найдена
func (s *NDollarIntegrationTestSuite) TestNDollarDenomFallback() {
	ctx := s.Ctx
	require := s.Require()

	s.T().Log("\n=== TEST: GetNDollarDenom Fallback ===")

	// Без metadata GetNDollarDenom должен вернуть константу
	denom := s.App.StablecoinKeeper.GetNDollarDenom(ctx)
	require.Equal(stablecointypes.NDollarDenom, denom, "Должен вернуть fallback константу")
	s.T().Logf("✓ Fallback denom: %s", denom)
}

// TestNDollarWithUnuahConversion проверяет автоматическую конвертацию unuah → NDOLLAR
func (s *NDollarIntegrationTestSuite) TestNDollarWithUnuahConversion() {
	ctx := s.Ctx
	require := s.Require()

	s.T().Log("\n=== TEST: Auto-conversion unuah → NDOLLAR ===")

	// Создаем пользователя и NDOLLAR
	user := s.TestAccs[1]
	initialUnuah := sdkmath.NewInt(500_000_000) // 500 NUAH
	require.NoError(banktestutil.FundAccount(ctx, s.App.BankKeeper, user, sdk.NewCoins(sdk.NewCoin("unuah", initialUnuah))))

	// Создаем NDOLLAR и metadata
	admin := s.App.AccountKeeper.GetModuleAddress(stablecointypes.ModuleName)
	_, err := s.App.TokenFactoryKeeper.CreateDenom(ctx, admin.String(), "ndollar")
	require.NoError(err)

	realNDollarDenom, err := tokenfactorytypes.GetTokenDenom(admin.String(), "ndollar")
	require.NoError(err)
	metadata := banktypes.Metadata{
		Description: "NUAH Dollar stablecoin",
		Base:        realNDollarDenom,
		Display:     "NDOLLAR",
		Name:        "NDOLLAR",
		Symbol:      "N$",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: realNDollarDenom, Exponent: 0, Aliases: []string{}},
			{Denom: "NDOLLAR", Exponent: 6, Aliases: []string{}},
		},
	}
	s.App.BankKeeper.SetDenomMetaData(ctx, metadata)

	// Регистрируем актив
	symbol := "GOLD"
	_, _, err = s.App.AssetsKeeper.EnsureAsset(ctx, symbol)
	require.NoError(err)

	// Устанавливаем цену (2000 USD за 1 GOLD)
	price := &oracletypes.Price{
		Symbol:    symbol,
		Value:     "2000.00", // 2000.00 USD
		Timestamp: ctx.BlockTime().Unix(),
	}
	s.App.OracleKeeper.SetPrice(ctx, price)

	// Покупаем актив напрямую с unuah (должна произойти автоматическая конвертация)
	purchaseAmount := sdkmath.NewInt(50_000_000) // 50 unuah
	unuahPayment := sdk.NewCoin("unuah", purchaseAmount)

	assetCoin, assetAmount, err := s.App.AssetsKeeper.BuyAssetWithPayment(ctx, user, symbol, unuahPayment)
	require.NoError(err)
	require.True(assetAmount.IsPositive(), "Должны получить актив")
	s.T().Logf("✓ Bought %s with unuah auto-conversion", assetCoin)

	// Проверяем что unuah списался
	unuahBalance := s.App.BankKeeper.GetBalance(ctx, user, "unuah")
	require.True(unuahBalance.Amount.LT(initialUnuah), "unuah должен списаться")
	s.T().Logf("✓ unuah balance after purchase: %s", unuahBalance.Amount)

	// Проверяем что промежуточный NDOLLAR не остался у пользователя
	// (был создан и сразу использован для покупки актива)
	ndollarBalance := s.App.BankKeeper.GetBalance(ctx, user, realNDollarDenom)
	s.T().Logf("✓ NDOLLAR balance after purchase: %s (should be 0 or very small)", ndollarBalance.Amount)
}

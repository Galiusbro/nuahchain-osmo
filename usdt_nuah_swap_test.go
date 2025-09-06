package poolmanager_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v30/app/params"
	"github.com/osmosis-labs/osmosis/v30/x/poolmanager/types"
)

// USDTNuahSwapTestSuite тестирует функциональность обмена USDT на NUAH
type USDTNuahSwapTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestUSDTNuahSwapTestSuite(t *testing.T) {
	suite.Run(t, new(USDTNuahSwapTestSuite))
}

const (
	// USDT IBC denom из assetlist
	USDT_IBC_DENOM = "ibc/8242AD24008032E457D2E12D46588FD39FB54FB29680C6C7663D296B383C37C4"
	// NUAH - базовая валюта сети
	NUAH = appparams.BaseCoinUnit
	// Тестовые суммы
	DEFAULT_USDT_AMOUNT = int64(1000000) // 1 USDT (6 decimals)
	DEFAULT_NUAH_AMOUNT = int64(1000000) // 1 NUAH (6 decimals для соотношения 1:1)
)

var (
	defaultUSDTCoin = sdk.NewCoin(USDT_IBC_DENOM, osmomath.NewInt(DEFAULT_USDT_AMOUNT))
	defaultNuahCoin = sdk.NewCoin(NUAH, osmomath.NewInt(DEFAULT_NUAH_AMOUNT))

	initialPoolLiquidity = sdk.NewCoins(
		sdk.NewCoin(USDT_IBC_DENOM, osmomath.NewInt(5000000000)), // 5000 USDT (6 decimals)
		sdk.NewCoin(NUAH, osmomath.NewInt(5000000000)),           // 5000 NUAH (6 decimals, соотношение 1:1)
	)
)

// TestSwapUSDTForNuah тестирует обмен USDT на NUAH
func (s *USDTNuahSwapTestSuite) TestSwapUSDTForNuah() {
	tests := []struct {
		name              string
		tokenIn           sdk.Coin
		tokenOutDenom     string
		tokenOutMinAmount osmomath.Int
		expectError       bool
		description       string
	}{
		{
			name:              "successful USDT to NUAH swap - small amount",
			tokenIn:           sdk.NewCoin(USDT_IBC_DENOM, osmomath.NewInt(1000000)), // 1 USDT
			tokenOutDenom:     NUAH,
			tokenOutMinAmount: osmomath.NewInt(1), // минимальное количество NUAH
			expectError:       false,
			description:       "Обмен 1 USDT на NUAH должен пройти успешно",
		},
		{
			name:              "successful USDT to NUAH swap - medium amount",
			tokenIn:           sdk.NewCoin(USDT_IBC_DENOM, osmomath.NewInt(10000000)), // 10 USDT
			tokenOutDenom:     NUAH,
			tokenOutMinAmount: osmomath.NewInt(1),
			expectError:       false,
			description:       "Обмен 10 USDT на NUAH должен пройти успешно",
		},
		{
			name:              "failed swap - insufficient USDT balance",
			tokenIn:           sdk.NewCoin(USDT_IBC_DENOM, osmomath.NewInt(1000000000)), // 1000 USDT (больше чем у пользователя)
			tokenOutDenom:     NUAH,
			tokenOutMinAmount: osmomath.NewInt(1),
			expectError:       true,
			description:       "Обмен должен провалиться из-за недостатка USDT",
		},
		{
			name:              "failed swap - slippage protection",
			tokenIn:           sdk.NewCoin(USDT_IBC_DENOM, osmomath.NewInt(1000000)), // 1 USDT
			tokenOutDenom:     NUAH,
			tokenOutMinAmount: osmomath.NewInt(1000000000000000), // нереально высокое минимальное количество
			expectError:       true,
			description:       "Обмен должен провалиться из-за защиты от проскальзывания",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Настройка теста
			s.Setup()

			// Создание пула USDT/NUAH
			s.FundAcc(s.TestAccs[0], initialPoolLiquidity)
			poolId := s.PrepareBalancerPoolWithCoins(initialPoolLiquidity...)

			// Финансирование тестового аккаунта USDT для обмена
			if !tc.expectError || tc.name != "failed swap - insufficient USDT balance" {
				s.FundAcc(s.TestAccs[1], sdk.NewCoins(tc.tokenIn))
			}

			// Получение пула
			pool, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
			s.Require().NoError(err)
			spreadFactor := pool.GetSpreadFactor(s.Ctx)

			// Выполнение обмена
			tokenOutAmount, err := s.App.GAMMKeeper.SwapExactAmountIn(
				s.Ctx,
				s.TestAccs[1],
				pool,
				tc.tokenIn,
				tc.tokenOutDenom,
				tc.tokenOutMinAmount,
				spreadFactor,
			)

			if tc.expectError {
				s.Require().Error(err, tc.description)
				s.T().Logf("Ожидаемая ошибка: %v", err)
			} else {
				s.Require().NoError(err, tc.description)
				s.Require().True(tokenOutAmount.GT(osmomath.ZeroInt()), "Количество полученных NUAH должно быть больше нуля")
				s.T().Logf("Успешный обмен: %s USDT -> %s NUAH", tc.tokenIn.Amount.String(), tokenOutAmount.String())

				// Проверка баланса пользователя после обмена
				userBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[1], NUAH)
				s.Require().Equal(tokenOutAmount, userBalance.Amount, "Баланс NUAH пользователя должен соответствовать полученной сумме")
			}
		})
	}
}

// TestSwapNuahForUSDT тестирует обратный обмен NUAH на USDT
func (s *USDTNuahSwapTestSuite) TestSwapNuahForUSDT() {
	tests := []struct {
		name              string
		tokenIn           sdk.Coin
		tokenOutDenom     string
		tokenOutMinAmount osmomath.Int
		expectError       bool
		description       string
	}{
		{
			name:              "successful NUAH to USDT swap",
			tokenIn:           sdk.NewCoin(NUAH, osmomath.NewInt(1000000)), // 1 NUAH
			tokenOutDenom:     USDT_IBC_DENOM,
			tokenOutMinAmount: osmomath.NewInt(1), // минимальное количество USDT
			expectError:       false,
			description:       "Обмен 1 NUAH на USDT должен пройти успешно",
		},
		{
			name:              "failed swap - insufficient NUAH balance",
			tokenIn:           sdk.NewCoin(NUAH, osmomath.NewInt(1000000000000000)), // 1,000,000 NUAH
			tokenOutDenom:     USDT_IBC_DENOM,
			tokenOutMinAmount: osmomath.NewInt(1),
			expectError:       true,
			description:       "Обмен должен провалиться из-за недостатка NUAH",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Настройка теста
			s.Setup()

			// Создание пула USDT/NUAH
			s.FundAcc(s.TestAccs[0], initialPoolLiquidity)
			poolId := s.PrepareBalancerPoolWithCoins(initialPoolLiquidity...)

			// Финансирование тестового аккаунта NUAH для обмена
			if !tc.expectError {
				s.FundAcc(s.TestAccs[1], sdk.NewCoins(tc.tokenIn))
			}

			// Получение пула
			pool, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
			s.Require().NoError(err)
			spreadFactor := pool.GetSpreadFactor(s.Ctx)

			// Выполнение обмена
			tokenOutAmount, err := s.App.GAMMKeeper.SwapExactAmountIn(
				s.Ctx,
				s.TestAccs[1],
				pool,
				tc.tokenIn,
				tc.tokenOutDenom,
				tc.tokenOutMinAmount,
				spreadFactor,
			)

			if tc.expectError {
				s.Require().Error(err, tc.description)
				s.T().Logf("Ожидаемая ошибка: %v", err)
			} else {
				s.Require().NoError(err, tc.description)
				s.Require().True(tokenOutAmount.GT(osmomath.ZeroInt()), "Количество полученных USDT должно быть больше нуля")
				s.T().Logf("Успешный обмен: %s NUAH -> %s USDT", tc.tokenIn.Amount.String(), tokenOutAmount.String())

				// Проверка баланса пользователя после обмена
				userBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[1], USDT_IBC_DENOM)
				s.Require().Equal(tokenOutAmount, userBalance.Amount, "Баланс USDT пользователя должен соответствовать полученной сумме")
			}
		})
	}
}

// TestMultihopSwapWithUSDT тестирует многошаговый обмен через USDT
func (s *USDTNuahSwapTestSuite) TestMultihopSwapWithUSDT() {
	s.Setup()

	// Создание нескольких пулов для многошагового обмена
	// Пул 1: USDT/NUAH
	s.FundAcc(s.TestAccs[0], initialPoolLiquidity)
	usdtNuahPoolId := s.PrepareBalancerPoolWithCoins(initialPoolLiquidity...)

	// Пул 2: NUAH/FOO (тестовый токен)
	fooCoins := sdk.NewCoins(
		sdk.NewCoin(NUAH, osmomath.NewInt(50000000000000)), // 50,000 NUAH
		sdk.NewCoin("foo", osmomath.NewInt(50000000000)),   // 50,000 FOO
	)
	s.FundAcc(s.TestAccs[0], fooCoins)
	nuahFooPoolId := s.PrepareBalancerPoolWithCoins(fooCoins...)

	// Настройка маршрута: USDT -> NUAH -> FOO
	routes := []types.SwapAmountInRoute{
		{
			PoolId:        usdtNuahPoolId,
			TokenOutDenom: NUAH,
		},
		{
			PoolId:        nuahFooPoolId,
			TokenOutDenom: "foo",
		},
	}

	// Финансирование пользователя USDT
	tokenIn := sdk.NewCoin(USDT_IBC_DENOM, osmomath.NewInt(5000000)) // 5 USDT
	s.FundAcc(s.TestAccs[1], sdk.NewCoins(tokenIn))

	// Выполнение многошагового обмена
	tokenOutAmount, err := s.App.PoolManagerKeeper.RouteExactAmountIn(
		s.Ctx,
		s.TestAccs[1],
		routes,
		tokenIn,
		osmomath.NewInt(1), // минимальное количество FOO
	)

	s.Require().NoError(err, "Многошаговый обмен USDT -> NUAH -> FOO должен пройти успешно")
	s.Require().True(tokenOutAmount.GT(osmomath.ZeroInt()), "Количество полученных FOO должно быть больше нуля")
	s.T().Logf("Многошаговый обмен: %s USDT -> %s FOO", tokenIn.Amount.String(), tokenOutAmount.String())

	// Проверка баланса пользователя
	userFooBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[1], "foo")
	s.Require().Equal(tokenOutAmount, userFooBalance.Amount, "Баланс FOO пользователя должен соответствовать полученной сумме")
}

// TestPoolLiquidityWithUSDT тестирует создание пула и проверку ликвидности USDT/NUAH
func (s *USDTNuahSwapTestSuite) TestPoolLiquidityWithUSDT() {
	s.Setup()

	// Создание пула USDT/NUAH
	s.FundAcc(s.TestAccs[0], initialPoolLiquidity)
	poolId := s.PrepareBalancerPoolWithCoins(initialPoolLiquidity...)

	// Проверка общей ликвидности пула
	totalLiquidity, err := s.App.GAMMKeeper.GetTotalPoolLiquidity(s.Ctx, poolId)
	s.Require().NoError(err, "Получение ликвидности пула должно пройти успешно")
	s.Require().True(totalLiquidity.IsAllPositive(), "Общая ликвидность пула должна быть положительной")
	s.T().Logf("Общая ликвидность пула: %s", totalLiquidity.String())

	// Проверка, что пул содержит правильные токены
	usdtInPool := false
	nuahInPool := false
	for _, coin := range totalLiquidity {
		if coin.Denom == USDT_IBC_DENOM {
			usdtInPool = true
		}
		if coin.Denom == NUAH {
			nuahInPool = true
		}
	}
	s.Require().True(usdtInPool, "Пул должен содержать USDT")
	s.Require().True(nuahInPool, "Пул должен содержать NUAH")
}

// TestUSDTDenomValidation тестирует валидацию USDT denom
func (s *USDTNuahSwapTestSuite) TestUSDTDenomValidation() {
	s.Setup()

	// Тест с правильным USDT IBC denom
	validUSDTCoin := sdk.NewCoin(USDT_IBC_DENOM, osmomath.NewInt(1000000))
	err := validUSDTCoin.Validate()
	s.Require().NoError(err, "Валидный USDT IBC denom должен проходить валидацию")

	// Тест с неправильным denom
	invalidUSDTCoin := sdk.NewCoin("invalid-usdt-denom", osmomath.NewInt(1000000))
	err = invalidUSDTCoin.Validate()
	s.Require().NoError(err, "Coin.Validate() не проверяет формат denom, только базовую структуру")

	s.T().Logf("USDT IBC denom: %s", USDT_IBC_DENOM)
	s.T().Logf("Длина USDT IBC denom: %d", len(USDT_IBC_DENOM))
}
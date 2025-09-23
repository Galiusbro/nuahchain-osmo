package apptesting

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"

	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	"github.com/osmosis-labs/osmosis/v30/app/params"
	claimstypes "github.com/osmosis-labs/osmosis/v30/x/claims/types"
	policytypes "github.com/osmosis-labs/osmosis/v30/x/policy/types"
	premiumtypes "github.com/osmosis-labs/osmosis/v30/x/premium/types"
	rolestypes "github.com/osmosis-labs/osmosis/v30/x/roles/types"
	treasurykeeper "github.com/osmosis-labs/osmosis/v30/x/treasury/keeper"
	treasurytypes "github.com/osmosis-labs/osmosis/v30/x/treasury/types"
)

type ComprehensiveInsuranceTestSuite struct {
	apptesting.KeeperTestHelper

	// Тестовые аккаунты
	insuranceAuthority sdk.AccAddress
	insurer            sdk.AccAddress
	policyHolder       sdk.AccAddress
	claimsReviewer     sdk.AccAddress
	treasuryManager    sdk.AccAddress
	oracle             sdk.AccAddress
}

func TestComprehensiveInsuranceTestSuite(t *testing.T) {
	suite.Run(t, new(ComprehensiveInsuranceTestSuite))
}

func (s *ComprehensiveInsuranceTestSuite) SetupTest() {
	s.Setup()

	// Создание тестовых аккаунтов
	testAccounts := apptesting.CreateRandomAccounts(6)
	s.insuranceAuthority = testAccounts[0]
	s.insurer = testAccounts[1]
	s.policyHolder = testAccounts[2]
	s.claimsReviewer = testAccounts[3]
	s.treasuryManager = testAccounts[4]
	s.oracle = testAccounts[5]

	// Финансирование аккаунтов
	baseCoin := sdk.NewInt64Coin(params.BaseCoinUnit, 10_000_000) // Увеличиваем до 10M unuah
	s.FundAcc(s.insuranceAuthority, sdk.NewCoins(baseCoin))
	s.FundAcc(s.insurer, sdk.NewCoins(baseCoin))
	s.FundAcc(s.policyHolder, sdk.NewCoins(baseCoin))
	s.FundAcc(s.claimsReviewer, sdk.NewCoins(baseCoin))
	s.FundAcc(s.treasuryManager, sdk.NewCoins(baseCoin))
	s.FundAcc(s.oracle, sdk.NewCoins(baseCoin))

	// Настройка параметров модулей
	ctx := s.Ctx.WithBlockTime(time.Now().UTC())
	s.Ctx = ctx

	// Настройка roles keeper
	rolesKeeper := s.App.RolesKeeper
	rolesParams := rolestypes.NewParams(s.insuranceAuthority.String())
	rolesKeeper.SetParams(ctx, rolesParams)

	// Настройка treasury keeper с пустым authority для использования ролей
	newTreasuryKeeper := treasurykeeper.NewKeeper(
		s.App.AppCodec(),
		s.App.GetKey(treasurytypes.StoreKey),
		s.App.GetSubspace(treasurytypes.ModuleName),
		s.App.BankKeeper,
		s.App.RolesKeeper,
		"", // Пустой authority для использования ролей
	)
	s.App.TreasuryKeeper = &newTreasuryKeeper
	
	treasuryParams := treasurytypes.NewParams("", "", false)
	newTreasuryKeeper.SetParams(ctx, treasuryParams)
	
	// Дополнительная отладочная информация для проверки authority
	s.T().Logf("Treasury params authority: %s", treasuryParams.Authority)
	s.T().Logf("Treasury keeper authority after: %s", newTreasuryKeeper.GetAuthority(ctx))

	// Настройка premium keeper
	premiumKeeper := s.App.PremiumKeeper
	premiumKeeper.SetParams(ctx, premiumtypes.NewParams([]string{params.BaseCoinUnit}, 0, false))

	// Настройка claims keeper
	claimsKeeper := s.App.ClaimsKeeper
	claimsKeeper.SetParams(ctx, claimstypes.Params{MaxOpenClaimsPerPolicy: 5})

	// Настройка policy keeper
	policyKeeper := s.App.PolicyKeeper
	policyParams := policytypes.NewParams([]string{"auto", "home", "health", "custom"}, 365, "")
	policyKeeper.SetParams(ctx, policyParams)
}

func (s *ComprehensiveInsuranceTestSuite) TestComprehensiveInsuranceWorkflow() {
	ctx := s.Ctx

	// 1. Настройка ролей - назначение ролей участникам
	s.T().Log("=== Шаг 1: Настройка ролей ===")

	// Назначение роли страховщика
	err := s.App.RolesKeeper.AssignRoles(ctx, s.insuranceAuthority, s.insurer, []rolestypes.Role{rolestypes.Role_ROLE_INSURER})
	s.Require().NoError(err)

	// Назначение роли держателя полиса
	err = s.App.RolesKeeper.AssignRoles(ctx, s.insuranceAuthority, s.policyHolder, []rolestypes.Role{rolestypes.Role_ROLE_POLICY_HOLDER})
	s.Require().NoError(err)

	// Назначение роли рецензента претензий
	err = s.App.RolesKeeper.AssignRoles(ctx, s.insuranceAuthority, s.claimsReviewer, []rolestypes.Role{rolestypes.Role_ROLE_CLAIMS_REVIEWER})
	s.Require().NoError(err)

	// Назначение роли менеджера казначейства
	err = s.App.RolesKeeper.AssignRoles(ctx, s.insuranceAuthority, s.treasuryManager, []rolestypes.Role{rolestypes.Role_ROLE_TREASURY_MANAGER})
	s.Require().NoError(err)

	// Назначение роли оракула
	err = s.App.RolesKeeper.AssignRoles(ctx, s.insuranceAuthority, s.oracle, []rolestypes.Role{rolestypes.Role_ROLE_ORACLE})
	s.Require().NoError(err)

	// Проверка назначения ролей
	s.Require().True(s.App.RolesKeeper.HasRole(ctx, s.insurer, rolestypes.Role_ROLE_INSURER))
	s.Require().True(s.App.RolesKeeper.HasRole(ctx, s.policyHolder, rolestypes.Role_ROLE_POLICY_HOLDER))
	s.Require().True(s.App.RolesKeeper.HasRole(ctx, s.claimsReviewer, rolestypes.Role_ROLE_CLAIMS_REVIEWER))
	s.Require().True(s.App.RolesKeeper.HasRole(ctx, s.treasuryManager, rolestypes.Role_ROLE_TREASURY_MANAGER))
	s.Require().True(s.App.RolesKeeper.HasRole(ctx, s.oracle, rolestypes.Role_ROLE_ORACLE))

	// Отладочная информация
	s.T().Logf("TreasuryManager address: %s", s.treasuryManager.String())
	s.T().Logf("HasRole check for treasuryManager: %v", s.App.RolesKeeper.HasRole(ctx, s.treasuryManager, rolestypes.Role_ROLE_TREASURY_MANAGER))
	
	// Проверим, что RolesKeeper не nil в TreasuryKeeper
	treasuryKeeper := s.App.TreasuryKeeper
	s.T().Logf("TreasuryKeeper authority: %s", treasuryKeeper.GetAuthority(ctx))
	
	// Получим все роли для treasuryManager
	binding, found := s.App.RolesKeeper.GetRoleBinding(ctx, s.treasuryManager)
	s.T().Logf("Role binding found: %v", found)
	if found {
		s.T().Logf("Roles for treasuryManager: %v", binding.Roles)
	}

	// 2. Создание пула казначейства
	s.T().Log("=== Шаг 2: Создание пула казначейства ===")

	treasuryPool := treasurytypes.TreasuryPool{
		Id:          "pool-001",
		Description: "Основной пул для общих страховых полисов",
		Manager:     s.treasuryManager.String(),
		PolicyTypes: []string{"auto", "home", "health"},
	}

	err = s.App.TreasuryKeeper.CreateTreasuryPool(ctx, s.treasuryManager, treasuryPool)
	s.Require().NoError(err)

	// Финансирование пула казначейства
	poolFunding := sdk.NewCoin(params.BaseCoinUnit, osmomath.NewInt(500000)) // Используем unuah
	err = s.App.TreasuryKeeper.DepositToTreasury(ctx, s.insurer, "pool-001", poolFunding)
	s.Require().NoError(err)

	// 3. Создание полиса
	s.T().Log("=== Шаг 3: Создание полиса ===")

	startTime := time.Now()
	endTime := startTime.Add(365 * 24 * time.Hour) // Полис на 1 год

	policyAttrs := []policytypes.PolicyAttribute{
		{Key: "coverage_amount", Value: "100000"},
		{Key: "deductible", Value: "1000"},
		{Key: "policy_type", Value: "auto"},
	}

	createdPolicy, err := s.App.PolicyKeeper.CreatePolicy(
		ctx,
		s.policyHolder,
		"auto",
		policyAttrs,
		&startTime,
		&endTime,
		"pool-001",
		[]string{"comprehensive", "collision"},
	)
	s.Require().NoError(err)
	s.Require().NotEmpty(createdPolicy.Id)

	// 4. Создание плана премий
	s.T().Log("=== Шаг 4: Создание плана премий ===")

	premiumSchedule := premiumtypes.PremiumSchedule{
		ScheduleType:  "monthly",
		PeriodSeconds: 30 * 24 * 60 * 60, // 30 дней в секундах
		MaxPayments:   12,                // 12 платежей в год
	}

	premiumAmount := sdk.NewCoin(params.BaseCoinUnit, osmomath.NewInt(1000)) // Используем unuah
	premiumPlan, err := s.App.PremiumKeeper.CreatePremiumPlan(ctx, s.insurer, createdPolicy.Id, s.policyHolder, premiumAmount, premiumSchedule, "pool-001")
	s.Require().NoError(err)
	s.Require().NotEmpty(premiumPlan.Id)

	// 5. Платежи по премиям
	s.T().Log("=== Шаг 5: Платежи по премиям ===")

	// Запись первого платежа по премии
	payment1Amount := sdk.NewCoin(params.BaseCoinUnit, osmomath.NewInt(1000)) // Используем unuah
	payment1, err := s.App.PremiumKeeper.RecordPremiumPayment(ctx, s.policyHolder, premiumPlan.Id, payment1Amount)
	s.Require().NoError(err)
	s.Require().NotEmpty(payment1.Id)

	// Запись второго платежа по премии
	payment2Amount := sdk.NewCoin(params.BaseCoinUnit, osmomath.NewInt(1000)) // Используем unuah
	payment2, err := s.App.PremiumKeeper.RecordPremiumPayment(ctx, s.policyHolder, premiumPlan.Id, payment2Amount)
	s.Require().NoError(err)
	s.Require().NotEmpty(payment2.Id)

	// 6. Подача претензии
	s.T().Log("=== Шаг 6: Подача претензии ===")

	claimEvidence := []claimstypes.ClaimEvidence{
		{
			Uri:   "https://example.com/evidence/photos.zip",
			Notes: "Фотографии места аварии",
		},
	}

	claimAmount := sdk.NewCoin(params.BaseCoinUnit, osmomath.NewInt(15000)) // Используем unuah
	submittedClaim, err := s.App.ClaimsKeeper.SubmitClaim(ctx, s.policyHolder, createdPolicy.Id, claimAmount, "Повреждение автомобиля при столкновении", claimEvidence)
	s.Require().NoError(err)
	s.Require().NotEmpty(submittedClaim.Id)

	// 7. Добавление доказательств
	s.T().Log("=== Шаг 7: Добавление доказательств ===")

	additionalEvidence := claimstypes.ClaimEvidence{
		Uri:   "https://example.com/evidence/police-report.pdf",
		Notes: "Отчет полиции о ДТП",
	}

	updatedClaim, err := s.App.ClaimsKeeper.AddClaimEvidence(ctx, s.claimsReviewer, submittedClaim.Id, additionalEvidence)
	s.Require().NoError(err)
	s.Require().Len(updatedClaim.Evidence, 2)

	// 8. Рассмотрение претензии
	s.T().Log("=== Шаг 8: Рассмотрение претензии ===")

	reviewedClaim, err := s.App.ClaimsKeeper.ReviewClaim(ctx, s.claimsReviewer, submittedClaim.Id, claimstypes.ClaimStatus_CLAIM_STATUS_APPROVED, "Претензия одобрена после тщательного рассмотрения доказательств")
	s.Require().NoError(err)
	s.Require().Equal(claimstypes.ClaimStatus_CLAIM_STATUS_APPROVED, reviewedClaim.Status)

	// 9. Выплата по претензии
	s.T().Log("=== Шаг 9: Выплата по претензии ===")

	paidClaim, err := s.App.ClaimsKeeper.ExecuteClaimPayout(ctx, s.treasuryManager, submittedClaim.Id, s.policyHolder)
	s.Require().NoError(err)
	s.Require().Equal(claimstypes.ClaimStatus_CLAIM_STATUS_PAID, paidClaim.Status)

	// 10. Управление пулом казначейства
	s.T().Log("=== Шаг 10: Управление пулом казначейства ===")

	// Установка резервов для пула
	reserves := []treasurytypes.PoolReserves{
		{
			PoolId:          "pool-001",
			Denom:           params.BaseCoinUnit, // Используем unuah
			MinReserveRatio: "0.1", // 10% минимальный резерв
		},
	}

	err = s.App.TreasuryKeeper.SetPoolReserves(ctx, s.treasuryManager, "pool-001", reserves)
	s.Require().NoError(err)

	// 11. Обновление полиса
	s.T().Log("=== Шаг 11: Обновление полиса ===")

	updatedAttrs := []policytypes.PolicyAttribute{
		{Key: "coverage_amount", Value: "120000"}, // Увеличенное покрытие
		{Key: "deductible", Value: "500"},         // Уменьшенная франшиза
		{Key: "policy_type", Value: "auto"},
	}

	_, err = s.App.PolicyKeeper.UpdatePolicyAttributes(ctx, s.insurer, createdPolicy.Id, updatedAttrs, false)
	s.Require().NoError(err)

	// 12. Просроченная премия
	s.T().Log("=== Шаг 12: Просроченная премия ===")

	overdueReason := "Платеж не получен в течение льготного периода"
	_, err = s.App.PremiumKeeper.MarkPremiumOverdue(ctx, s.insurer, premiumPlan.Id, overdueReason)
	s.Require().NoError(err)

	// 13. Отзыв роли
	s.T().Log("=== Шаг 13: Отзыв роли ===")

	err = s.App.RolesKeeper.RevokeRoles(ctx, s.insuranceAuthority, s.oracle, []rolestypes.Role{rolestypes.Role_ROLE_ORACLE})
	s.Require().NoError(err)

	// Проверка отзыва роли
	s.Require().False(s.App.RolesKeeper.HasRole(ctx, s.oracle, rolestypes.Role_ROLE_ORACLE))

	// 14. Отмена полиса
	s.T().Log("=== Шаг 14: Отмена полиса ===")

	_, err = s.App.PolicyKeeper.CancelPolicy(ctx, s.insurer, createdPolicy.Id, "Полис отменен из-за неуплаты")
	s.Require().NoError(err)

	// 15. Финальная проверка
	s.T().Log("=== Шаг 15: Финальная проверка ===")

	// Проверка финального статуса полиса
	finalPolicy, found := s.App.PolicyKeeper.GetPolicy(ctx, createdPolicy.Id)
	s.Require().True(found)
	s.Require().Equal(policytypes.PolicyStatus_POLICY_STATUS_CANCELLED, finalPolicy.Status)

	// Проверка, что пул казначейства все еще существует
	finalPool, found := s.App.TreasuryKeeper.GetTreasuryPool(ctx, "pool-001")
	s.Require().True(found)
	s.Require().Equal("pool-001", finalPool.Id)

	s.T().Log("=== Комплексный тест страхового рабочего процесса успешно завершен ===")
}
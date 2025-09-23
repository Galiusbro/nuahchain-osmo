package e2e

import (
	"fmt"
	"strings"
	"time"

	"github.com/osmosis-labs/osmosis/v30/app/params"
	"github.com/osmosis-labs/osmosis/v30/tests/e2e/initialization"
	claimstypes "github.com/osmosis-labs/osmosis/v30/x/claims/types"
)

func (s *IntegrationTestSuite) InsuranceLifecycle() {
	chainCfg, node := s.getChainCfgs()
	s.Require().NotNil(chainCfg)
	s.Require().NotNil(node)

	authorityWallet := initialization.ValidatorWalletName
	authorityAddr := strings.TrimSpace(node.GetWallet(authorityWallet))

	// assign required roles to the authority account
	node.AssignRoles(authorityWallet, authorityAddr, authorityAddr, "insurer", "treasury_manager", "claims_reviewer")

	// create treasury pool and fund it
	poolID := fmt.Sprintf("insurance-%s-%d", chainCfg.Id, time.Now().UnixNano())
	node.CreateTreasuryPool(authorityWallet, authorityAddr, poolID, "integration insurance pool", authorityAddr, []string{"custom"})

	depositAmount := fmt.Sprintf("1000000%s", params.BaseCoinUnit)
	node.TreasuryDeposit(authorityWallet, authorityAddr, poolID, depositAmount)

	// create policy owned by the authority address
	policyID := node.CreatePolicy(authorityWallet, authorityAddr, "custom", poolID, map[string]string{"coverage": "basic"}, []string{"integration"}, 0, 0)

	// create premium plan and record first payment
	planAmount := fmt.Sprintf("50000%s", params.BaseCoinUnit)
	planID := node.CreatePremiumPlan(authorityWallet, authorityAddr, policyID, authorityAddr, planAmount, 24*60*60, 0, "periodic", poolID)
	node.RecordPremiumPayment(authorityWallet, authorityAddr, planID, planAmount)

	// submit, approve, and pay out a claim
	claimAmount := fmt.Sprintf("25000%s", params.BaseCoinUnit)
	claimID := node.SubmitClaim(authorityWallet, authorityAddr, policyID, claimAmount, "integration test claim", nil)
	node.ReviewClaim(authorityWallet, authorityAddr, claimID, "approved", "approved during integration test")
	node.ExecuteClaimPayout(authorityWallet, authorityAddr, claimID, authorityAddr)

	// verify the claim status is paid
	claim := node.QueryClaim(claimID)
	s.Require().Equal(claimstypes.ClaimStatus_CLAIM_STATUS_PAID, claim.Status)
}

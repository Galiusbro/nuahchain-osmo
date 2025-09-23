package premium

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/premium/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/premium/types"
)

// InitGenesis initializes state from genesis data.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState *types.GenesisState) {
	if genState == nil {
		genState = types.DefaultGenesis()
	}

	k.SetParams(ctx, genState.Params)
	if genState.NextPlanId == 0 {
		genState.NextPlanId = 1
	}
	if genState.NextPaymentId == 0 {
		genState.NextPaymentId = 1
	}

	k.SetNextPlanID(ctx, genState.NextPlanId)
	k.SetNextPaymentID(ctx, genState.NextPaymentId)

	for _, plan := range genState.Plans {
		k.SetPremiumPlan(ctx, plan)
	}

	for _, payment := range genState.Payments {
		k.SetPremiumPayment(ctx, payment)
	}

	for _, overdue := range genState.Overdue {
		k.SetPremiumOverdue(ctx, overdue)
	}
}

// ExportGenesis exports the module state to genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:        k.GetParams(ctx),
		NextPlanId:    k.GetNextPlanID(ctx),
		NextPaymentId: k.GetNextPaymentID(ctx),
		Plans:         k.ExportPlans(ctx),
		Payments:      k.ExportPayments(ctx),
		Overdue:       k.ExportOverdue(ctx),
	}
}

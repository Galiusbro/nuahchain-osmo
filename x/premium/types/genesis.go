package types

import "fmt"

// DefaultGenesis returns default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:        DefaultParams(),
		NextPlanId:    1,
		NextPaymentId: 1,
		Plans:         []PremiumPlan{},
		Payments:      []PremiumPayment{},
		Overdue:       []PremiumOverdue{},
	}
}

// Validate performs basic genesis validation.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	planIDs := make(map[uint64]struct{})
	for _, plan := range gs.Plans {
		if plan.Id == 0 {
			return fmt.Errorf("plan id must be positive")
		}
		if _, exists := planIDs[plan.Id]; exists {
			return fmt.Errorf("duplicate plan id %d", plan.Id)
		}
		planIDs[plan.Id] = struct{}{}
	}

	paymentIDs := make(map[uint64]struct{})
	for _, payment := range gs.Payments {
		if payment.Id == 0 {
			return fmt.Errorf("payment id must be positive")
		}
		if _, exists := paymentIDs[payment.Id]; exists {
			return fmt.Errorf("duplicate payment id %d", payment.Id)
		}
	}

	return nil
}

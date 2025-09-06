package types

import (
	"time"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: &Params{
			DefaultDailyLimit: 3,
			ResetHour:         0, // midnight UTC
		},
		LimitedAccounts: []*LimitedAccount{},
	}
}

// DefaultParams returns default parameters
func DefaultParams() *Params {
	return &Params{
		DefaultDailyLimit: 3,
		ResetHour:         0, // midnight UTC
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Validate params
	if gs.Params == nil {
		return ErrInvalidParams
	}
	if gs.Params.DefaultDailyLimit == 0 {
		return ErrInvalidParams
	}
	if gs.Params.ResetHour > 23 {
		return ErrInvalidParams
	}

	// Validate limited accounts
	addressMap := make(map[string]bool)
	for _, account := range gs.LimitedAccounts {
		if account.Address == "" {
			return ErrInvalidAddress
		}
		if addressMap[account.Address] {
			return ErrDuplicateAddress
		}
		addressMap[account.Address] = true
		if account.MaxDailyTxs == 0 {
			account.MaxDailyTxs = gs.Params.DefaultDailyLimit
		}
	}

	return nil
}

// NewLimitedAccount creates a new limited account
func NewLimitedAccount(address string, maxDailyTxs uint32) *LimitedAccount {
	if maxDailyTxs == 0 {
		maxDailyTxs = 3 // default
	}
	return &LimitedAccount{
		Address:       address,
		DailyTxCount:  0,
		LastResetTime: time.Now(),
		MaxDailyTxs:   maxDailyTxs,
	}
}

// CanTransact checks if the account can perform a transaction
func (la *LimitedAccount) CanTransact() bool {
	// Check if we need to reset the daily count
	now := time.Now()
	if now.Sub(la.LastResetTime).Hours() >= 24 {
		la.DailyTxCount = 0
		la.LastResetTime = now
	}

	return la.DailyTxCount < la.MaxDailyTxs
}

// IncrementTxCount increments the daily transaction count
func (la *LimitedAccount) IncrementTxCount() {
	la.DailyTxCount++
}

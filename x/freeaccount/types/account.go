package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// FreeAccount represents an account that doesn't pay transaction fees
type FreeAccount struct {
	*authtypes.BaseAccount
	
	// IsFreeAccount marks this account as fee-exempt
	IsFreeAccount bool `json:"is_free_account,omitempty"`
}

// NewFreeAccount creates a new FreeAccount
func NewFreeAccount(baseAccount *authtypes.BaseAccount) *FreeAccount {
	return &FreeAccount{
		BaseAccount:   baseAccount,
		IsFreeAccount: true,
	}
}

// NewFreeAccountWithAddress creates a new FreeAccount with the given address
func NewFreeAccountWithAddress(addr sdk.AccAddress) *FreeAccount {
	return &FreeAccount{
		BaseAccount:   authtypes.NewBaseAccountWithAddress(addr),
		IsFreeAccount: true,
	}
}

// GetType returns the account type
func (fa FreeAccount) GetType() string {
	return "FreeAccount"
}

// Validate checks if the account is valid
func (fa FreeAccount) Validate() error {
	if fa.BaseAccount == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "base account cannot be nil")
	}
	return fa.BaseAccount.Validate()
}

// String returns a string representation of the account
func (fa FreeAccount) String() string {
	return fa.BaseAccount.String()
}
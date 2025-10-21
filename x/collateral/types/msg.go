package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewMsgDeposit creates a new MsgDeposit instance.
func NewMsgDeposit(depositor string, amount string) *MsgDeposit {
	return &MsgDeposit{Depositor: depositor, Amount: amount}
}

// ValidateBasic performs stateless message validation.
func (msg *MsgDeposit) ValidateBasic() error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Depositor); err != nil {
		return fmt.Errorf("invalid depositor address: %w", err)
	}
	coin, err := sdk.ParseCoinNormalized(msg.Amount)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	if !coin.Amount.IsPositive() {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}

// GetSigners returns the expected signers.
func (msg *MsgDeposit) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewMsgWithdraw creates a new MsgWithdraw instance.
func NewMsgWithdraw(owner string, amount string) *MsgWithdraw {
	return &MsgWithdraw{Owner: owner, Amount: amount}
}

// ValidateBasic performs stateless validation for MsgWithdraw.
func (msg *MsgWithdraw) ValidateBasic() error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %w", err)
	}
	coin, err := sdk.ParseCoinNormalized(msg.Amount)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	if !coin.Amount.IsPositive() {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}

// GetSigners returns the withdraw message signers.
func (msg *MsgWithdraw) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

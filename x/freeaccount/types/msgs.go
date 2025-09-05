package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewMsgCreateFreeAccount creates a new MsgCreateFreeAccount instance
func NewMsgCreateFreeAccount(authority, address string) *MsgCreateFreeAccount {
	return &MsgCreateFreeAccount{
		Authority: authority,
		Address:   address,
	}
}

// Route returns the route of MsgCreateFreeAccount
func (msg MsgCreateFreeAccount) Route() string { return RouterKey }

// Type returns the type of MsgCreateFreeAccount
func (msg MsgCreateFreeAccount) Type() string { return "create_free_account" }

// GetSigners returns the signers of MsgCreateFreeAccount
func (msg MsgCreateFreeAccount) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes returns the sign bytes of MsgCreateFreeAccount
func (msg MsgCreateFreeAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic validates the MsgCreateFreeAccount
func (msg MsgCreateFreeAccount) ValidateBasic() error {
	if msg.Authority == "" {
		return errors.New("authority cannot be empty")
	}
	if msg.Address == "" {
		return errors.New("address cannot be empty")
	}
	return nil
}

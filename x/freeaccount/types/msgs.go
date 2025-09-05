package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgCreateFreeAccount defines a message to create a free account
type MsgCreateFreeAccount struct {
	// Authority is the address that controls the module (gov module account)
	Authority string `json:"authority,omitempty"`
	// Address is the address to make fee-exempt
	Address string `json:"address,omitempty"`
}

// NewMsgCreateFreeAccount creates a new MsgCreateFreeAccount instance
func NewMsgCreateFreeAccount(authority, address string) *MsgCreateFreeAccount {
	return &MsgCreateFreeAccount{
		Authority: authority,
		Address:   address,
	}
}

// Route returns the name of the module
func (msg MsgCreateFreeAccount) Route() string { return ModuleName }

// Type returns the action
func (msg MsgCreateFreeAccount) Type() string { return "create_free_account" }

// GetSigners returns the expected signers for the message
func (msg *MsgCreateFreeAccount) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// GetSignBytes encodes the message for signing
func (msg *MsgCreateFreeAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(Amino.MustMarshalJSON(msg))
}

// ValidateBasic performs basic validation of the message
func (msg *MsgCreateFreeAccount) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}

	return nil
}

// ProtoMessage implements proto.Message interface
func (msg *MsgCreateFreeAccount) ProtoMessage() {}

// Reset implements proto.Message interface
func (msg *MsgCreateFreeAccount) Reset() {
	*msg = MsgCreateFreeAccount{}
}

// String implements proto.Message interface
func (msg *MsgCreateFreeAccount) String() string {
	return "MsgCreateFreeAccount{Authority: " + msg.Authority + ", Address: " + msg.Address + "}"
}
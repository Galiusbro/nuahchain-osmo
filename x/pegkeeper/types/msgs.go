package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Message types for the pegkeeper module
const (
	TypeMsgUpdateParams = "update_params"
)

// NewMsgUpdateParams creates a new MsgUpdateParams instance
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// Route returns the message route
func (msg MsgUpdateParams) Route() string {
	return RouterKey
}

// Type returns the message type
func (msg MsgUpdateParams) Type() string {
	return TypeMsgUpdateParams
}

// GetSigners returns the signers of the message
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// GetSignBytes returns the sign bytes for the message
func (msg *MsgUpdateParams) GetSignBytes() []byte {
	bz := sdk.MustSortJSON([]byte{})
	return bz
}

// ValidateBasic performs basic validation of the message
func (msg *MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address (%s)", err)
	}
	return msg.Params.Validate()
}

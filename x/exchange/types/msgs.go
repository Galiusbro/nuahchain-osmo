package types

import (
	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgExchangeTokens = "exchange_tokens"
	TypeMsgUpdateParams   = "update_params"
)

var (
	_ sdk.Msg = &MsgExchangeTokens{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgExchangeTokens creates a new MsgExchangeTokens instance
func NewMsgExchangeTokens(sender string, tokenIn sdk.Coin, minNuahOut math.Int) *MsgExchangeTokens {
	return &MsgExchangeTokens{
		Sender:     sender,
		TokenIn:    tokenIn,
		MinNuahOut: minNuahOut,
	}
}

// Route implements the sdk.Msg interface
func (msg MsgExchangeTokens) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface
func (msg MsgExchangeTokens) Type() string {
	return TypeMsgExchangeTokens
}

// GetSigners implements the sdk.Msg interface
func (msg MsgExchangeTokens) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// GetSignBytes implements the sdk.Msg interface
func (msg MsgExchangeTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface
func (msg MsgExchangeTokens) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address (%s)", err)
	}

	if !msg.TokenIn.IsValid() || msg.TokenIn.IsZero() {
		return errors.Wrap(ErrInvalidExchangeAmount, "token in must be valid and positive")
	}

	if msg.MinNuahOut.IsNegative() {
		return errors.Wrap(ErrInvalidMinOutput, "minimum nuah out cannot be negative")
	}

	return nil
}

// NewMsgUpdateParams creates a new MsgUpdateParams instance
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// Route implements the sdk.Msg interface
func (msg MsgUpdateParams) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface
func (msg MsgUpdateParams) Type() string {
	return TypeMsgUpdateParams
}

// GetSigners implements the sdk.Msg interface
func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// GetSignBytes implements the sdk.Msg interface
func (msg MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface
func (msg MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	return msg.Params.Validate()
}
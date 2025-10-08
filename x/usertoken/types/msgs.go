package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCreateToken{}
var _ sdk.Msg = &MsgFounderClaim{}

func NewMsgCreateToken(creator, name, symbol, image, description string) *MsgCreateToken {
	return &MsgCreateToken{
		Creator:     creator,
		Name:        name,
		Symbol:      symbol,
		Image:       image,
		Description: description,
	}
}

func (msg *MsgCreateToken) Route() string { return RouterKey }

func (msg *MsgCreateToken) Type() string { return "create_token" }

func (msg *MsgCreateToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrInvalidAddress
	}
	if err := ValidateName(msg.Name); err != nil {
		return err
	}
	if err := ValidateSymbol(msg.Symbol); err != nil {
		return err
	}
	if err := ValidateImage(msg.Image); err != nil {
		return err
	}
	if err := ValidateDescription(msg.Description); err != nil {
		return err
	}
	return nil
}

func (msg *MsgCreateToken) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{addr}
}

func (msg *MsgFounderClaim) Route() string { return RouterKey }

func (msg *MsgFounderClaim) Type() string { return "founder_claim" }

func (msg *MsgFounderClaim) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Founder); err != nil {
		return ErrInvalidAddress
	}
	if strings.TrimSpace(msg.Denom) == "" {
		return ErrTokenNotFound
	}
	return nil
}

func (msg *MsgFounderClaim) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Founder)
	return []sdk.AccAddress{addr}
}

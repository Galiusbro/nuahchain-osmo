package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewMsgSetPrice constructs a new MsgSetPrice.
func NewMsgSetPrice(authority, symbol, value string) *MsgSetPrice {
	return &MsgSetPrice{
		Authority: authority,
		Symbol:    symbol,
		Value:     value,
	}
}

// Route implements sdk.Msg.
func (m *MsgSetPrice) Route() string { return RouterKey }

// Type implements sdk.Msg.
func (m *MsgSetPrice) Type() string { return TypeMsgSetPrice }

// GetSigners implements sdk.Msg.
func (m *MsgSetPrice) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg.
func (m *MsgSetPrice) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// ValidateBasic implements sdk.Msg.
func (m *MsgSetPrice) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority: %v", err)
	}
	if strings.TrimSpace(m.Symbol) == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("symbol cannot be empty")
	}
	if strings.TrimSpace(m.Value) == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("value cannot be empty")
	}
	return nil
}

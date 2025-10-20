package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewMsgEnsureAsset creates a new MsgEnsureAsset instance.
func NewMsgEnsureAsset(creator, symbol string) *MsgEnsureAsset {
	return &MsgEnsureAsset{
		Creator: creator,
		Symbol:  symbol,
	}
}

// Route implements sdk.Msg.
func (m *MsgEnsureAsset) Route() string { return RouterKey }

// Type implements sdk.Msg.
func (m *MsgEnsureAsset) Type() string { return TypeMsgEnsureAsset }

// GetSigners implements sdk.Msg.
func (m *MsgEnsureAsset) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes implements sdk.Msg.
func (m *MsgEnsureAsset) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// ValidateBasic implements sdk.Msg.
func (m *MsgEnsureAsset) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Creator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid creator address: %v", err)
	}

	if strings.TrimSpace(m.Symbol) == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("symbol cannot be empty")
	}

	return nil
}

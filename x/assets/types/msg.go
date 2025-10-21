package types

import (
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	osmomath "github.com/osmosis-labs/osmosis/osmomath"
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

// NewMsgBuyAsset creates a new MsgBuyAsset instance.
func NewMsgBuyAsset(buyer, symbol, amountNDOLLAR string) *MsgBuyAsset {
	return &MsgBuyAsset{
		Buyer:          buyer,
		Symbol:         symbol,
		Amount_NDOLLAR: amountNDOLLAR,
	}
}

// Route implements sdk.Msg.
func (m *MsgBuyAsset) Route() string { return RouterKey }

// Type implements sdk.Msg.
func (m *MsgBuyAsset) Type() string { return TypeMsgBuyAsset }

// GetSigners implements sdk.Msg.
func (m *MsgBuyAsset) GetSigners() []sdk.AccAddress {
	buyer, err := sdk.AccAddressFromBech32(m.Buyer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{buyer}
}

// GetSignBytes implements sdk.Msg.
func (m *MsgBuyAsset) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// ValidateBasic implements sdk.Msg.
func (m *MsgBuyAsset) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Buyer); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid buyer address: %v", err)
	}

	if strings.TrimSpace(m.Symbol) == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("symbol cannot be empty")
	}

	if strings.TrimSpace(m.Amount_NDOLLAR) == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("amount_ndollar cannot be empty")
	}

	if _, ok := sdkmath.NewIntFromString(m.Amount_NDOLLAR); !ok {
		return sdkerrors.ErrInvalidRequest.Wrap("amount_ndollar must be an integer")
	}

	return nil
}

// NewMsgSellAsset creates a new MsgSellAsset instance.
func NewMsgSellAsset(seller, symbol, baseAmount string) *MsgSellAsset {
	return &MsgSellAsset{
		Seller:     seller,
		Symbol:     symbol,
		BaseAmount: baseAmount,
	}
}

// Route implements sdk.Msg.
func (m *MsgSellAsset) Route() string { return RouterKey }

// Type implements sdk.Msg.
func (m *MsgSellAsset) Type() string { return TypeMsgSellAsset }

// GetSigners implements sdk.Msg.
func (m *MsgSellAsset) GetSigners() []sdk.AccAddress {
	seller, err := sdk.AccAddressFromBech32(m.Seller)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{seller}
}

// GetSignBytes implements sdk.Msg.
func (m *MsgSellAsset) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// ValidateBasic implements sdk.Msg.
func (m *MsgSellAsset) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Seller); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid seller address: %v", err)
	}

	if strings.TrimSpace(m.Symbol) == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("symbol cannot be empty")
	}

	if strings.TrimSpace(m.BaseAmount) == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("base_amount cannot be empty")
	}

	dec, err := osmomath.NewDecFromStr(m.BaseAmount)
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid base_amount: %v", err)
	}
	if !dec.IsPositive() {
		return sdkerrors.ErrInvalidRequest.Wrap("base_amount must be positive")
	}

	return nil
}

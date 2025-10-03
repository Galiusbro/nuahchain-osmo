package types

import (
	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgOpenPosition      = "open_position"
	TypeMsgClosePosition     = "close_position"
	TypeMsgAddCollateral     = "add_collateral"
	TypeMsgRemoveCollateral  = "remove_collateral"
	TypeMsgLiquidatePosition = "liquidate_position"
	TypeMsgUpdateParams      = "update_params"
	TypeMsgProvideLiquidity  = "provide_liquidity"
)

var (
	_ sdk.Msg = &MsgOpenPosition{}
	_ sdk.Msg = &MsgClosePosition{}
	_ sdk.Msg = &MsgAddCollateral{}
	_ sdk.Msg = &MsgRemoveCollateral{}
	_ sdk.Msg = &MsgLiquidatePosition{}
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgProvideLiquidity{}
)

// MsgOpenPosition
func NewMsgOpenPosition(trader, tokenDenom string, collateral sdk.Coin, leverage math.LegacyDec, side PositionSide, minPrice, maxPrice math.LegacyDec) *MsgOpenPosition {
	return &MsgOpenPosition{
		Trader:     trader,
		TokenDenom: tokenDenom,
		Collateral: collateral,
		Leverage:   leverage,
		Side:       side,
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
	}
}

func (msg *MsgOpenPosition) Route() string {
	return RouterKey
}

func (msg *MsgOpenPosition) Type() string {
	return TypeMsgOpenPosition
}

func (msg *MsgOpenPosition) GetSigners() []sdk.AccAddress {
	trader, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{trader}
}

func (msg *MsgOpenPosition) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgOpenPosition) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid trader address (%s)", err)
	}

	if msg.TokenDenom == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "token denom cannot be empty")
	}

	if !msg.Collateral.IsValid() || !msg.Collateral.IsPositive() {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, "collateral must be positive")
	}

	if msg.Leverage.IsNil() || msg.Leverage.LTE(math.LegacyOneDec()) {
		return errors.Wrap(ErrInvalidLeverage, "leverage must be greater than 1")
	}

	if msg.Side == PositionSideUnspecified {
		return errors.Wrap(ErrInvalidPositionSide, "position side must be specified")
	}

	if msg.MinPrice.IsNil() || msg.MinPrice.IsNegative() {
		return errors.Wrap(ErrInvalidPrice, "min price must be non-negative")
	}

	if msg.MaxPrice.IsNil() || msg.MaxPrice.IsNegative() {
		return errors.Wrap(ErrInvalidPrice, "max price must be non-negative")
	}

	return nil
}

// MsgClosePosition
func NewMsgClosePosition(trader, positionID string, minPrice, maxPrice math.LegacyDec) *MsgClosePosition {
	return &MsgClosePosition{
		Trader:     trader,
		PositionId: positionID,
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
	}
}

func (msg *MsgClosePosition) Route() string {
	return RouterKey
}

func (msg *MsgClosePosition) Type() string {
	return TypeMsgClosePosition
}

func (msg *MsgClosePosition) GetSigners() []sdk.AccAddress {
	trader, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{trader}
}

func (msg *MsgClosePosition) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgClosePosition) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid trader address (%s)", err)
	}

	if msg.PositionId == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "position ID cannot be empty")
	}

	if msg.MinPrice.IsNil() || msg.MinPrice.IsNegative() {
		return errors.Wrap(ErrInvalidPrice, "min price must be non-negative")
	}

	if msg.MaxPrice.IsNil() || msg.MaxPrice.IsNegative() {
		return errors.Wrap(ErrInvalidPrice, "max price must be non-negative")
	}

	return nil
}

// MsgAddCollateral
func NewMsgAddCollateral(trader, positionID string, amount sdk.Coin) *MsgAddCollateral {
	return &MsgAddCollateral{
		Trader:     trader,
		PositionId: positionID,
		Amount:     amount,
	}
}

func (msg *MsgAddCollateral) Route() string {
	return RouterKey
}

func (msg *MsgAddCollateral) Type() string {
	return TypeMsgAddCollateral
}

func (msg *MsgAddCollateral) GetSigners() []sdk.AccAddress {
	trader, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{trader}
}

func (msg *MsgAddCollateral) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddCollateral) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid trader address (%s)", err)
	}

	if msg.PositionId == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "position ID cannot be empty")
	}

	if !msg.Amount.IsValid() || !msg.Amount.IsPositive() {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, "amount must be positive")
	}

	return nil
}

// MsgRemoveCollateral
func NewMsgRemoveCollateral(trader, positionID string, amount sdk.Coin) *MsgRemoveCollateral {
	return &MsgRemoveCollateral{
		Trader:     trader,
		PositionId: positionID,
		Amount:     amount,
	}
}

func (msg *MsgRemoveCollateral) Route() string {
	return RouterKey
}

func (msg *MsgRemoveCollateral) Type() string {
	return TypeMsgRemoveCollateral
}

func (msg *MsgRemoveCollateral) GetSigners() []sdk.AccAddress {
	trader, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{trader}
}

func (msg *MsgRemoveCollateral) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveCollateral) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid trader address (%s)", err)
	}

	if msg.PositionId == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "position ID cannot be empty")
	}

	if !msg.Amount.IsValid() || !msg.Amount.IsPositive() {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, "amount must be positive")
	}

	return nil
}

// MsgLiquidatePosition
func NewMsgLiquidatePosition(liquidator, positionID string) *MsgLiquidatePosition {
	return &MsgLiquidatePosition{
		Liquidator: liquidator,
		PositionId: positionID,
	}
}

func (msg *MsgLiquidatePosition) Route() string {
	return RouterKey
}

func (msg *MsgLiquidatePosition) Type() string {
	return TypeMsgLiquidatePosition
}

func (msg *MsgLiquidatePosition) GetSigners() []sdk.AccAddress {
	liquidator, err := sdk.AccAddressFromBech32(msg.Liquidator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{liquidator}
}

func (msg *MsgLiquidatePosition) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgLiquidatePosition) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Liquidator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid liquidator address (%s)", err)
	}

	if msg.PositionId == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "position ID cannot be empty")
	}

	return nil
}

// MsgUpdateParams
func NewMsgUpdateParams(authority string, params LeverageParams) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

func (msg *MsgUpdateParams) Route() string {
	return RouterKey
}

func (msg *MsgUpdateParams) Type() string {
	return TypeMsgUpdateParams
}

func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

func (msg *MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	return msg.Params.Validate()
}

func (msg *MsgProvideLiquidity) Route() string {
	return RouterKey
}

func (msg *MsgProvideLiquidity) Type() string {
	return TypeMsgProvideLiquidity
}

func (msg *MsgProvideLiquidity) GetSigners() []sdk.AccAddress {
	provider, err := sdk.AccAddressFromBech32(msg.Provider)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{provider}
}

func (msg *MsgProvideLiquidity) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgProvideLiquidity) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Provider)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid provider address (%s)", err)
	}

	if msg.Amount.Amount.IsZero() || msg.Amount.Amount.IsNegative() {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid amount: %s", msg.Amount.Amount.String())
	}

	if msg.Amount.Denom == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "denom cannot be empty")
	}

	return nil
}

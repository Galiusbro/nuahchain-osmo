package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgCreateUserToken    = "create_user_token"
	TypeMsgBuyTokens          = "buy_tokens"
	TypeMsgSellTokens         = "sell_tokens"
	TypeMsgClaimFounderTokens = "claim_founder_tokens"
	TypeMsgStartLBP           = "start_lbp"
)

// Message type assertions will be handled by protobuf generation

func NewMsgCreateUserToken(creator, subdenom, name, symbol string, decimals uint32) *MsgCreateUserToken {
	return &MsgCreateUserToken{
		Creator:  creator,
		Subdenom: subdenom,
		Name:     name,
		Symbol:   symbol,
		Decimals: decimals,
	}
}

func (msg *MsgCreateUserToken) Route() string {
	return RouterKey
}

func (msg *MsgCreateUserToken) Type() string {
	return TypeMsgCreateUserToken
}

func (msg *MsgCreateUserToken) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateUserToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateUserToken) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Validate subdenom
	if len(msg.Subdenom) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "subdenom cannot be empty")
	}
	if len(msg.Subdenom) > 44 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "subdenom cannot be longer than 44 characters")
	}
	// Check for valid characters (alphanumeric and hyphens only)
	for _, char := range msg.Subdenom {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-') {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "subdenom can only contain alphanumeric characters and hyphens")
		}
	}

	// Validate name
	if len(msg.Name) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "name cannot be empty")
	}
	if len(msg.Name) > 128 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "name cannot be longer than 128 characters")
	}

	// Validate symbol
	if len(msg.Symbol) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "symbol cannot be empty")
	}
	if len(msg.Symbol) > 32 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "symbol cannot be longer than 32 characters")
	}
	// Check for valid characters (alphanumeric only)
	for _, char := range msg.Symbol {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "symbol can only contain alphanumeric characters")
		}
	}

	// Validate decimals
	if msg.Decimals > 18 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "decimals cannot be greater than 18")
	}

	return nil
}

func NewMsgBuyTokens(buyer, denom string, amount sdk.Coin, minTokens string) *MsgBuyTokens {
	minTokensInt, _ := math.NewIntFromString(minTokens)
	return &MsgBuyTokens{
		Buyer:     buyer,
		Denom:     denom,
		Amount:    amount,
		MinTokens: minTokensInt,
	}
}

func (msg *MsgBuyTokens) Route() string {
	return RouterKey
}

func (msg *MsgBuyTokens) Type() string {
	return TypeMsgBuyTokens
}

func (msg *MsgBuyTokens) GetSigners() []sdk.AccAddress {
	buyer, err := sdk.AccAddressFromBech32(msg.Buyer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{buyer}
}

func (msg *MsgBuyTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgBuyTokens) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Buyer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid buyer address (%s)", err)
	}

	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "invalid amount")
	}

	return nil
}

func NewMsgSellTokens(seller, denom string, amount sdk.Coin, minPrice string) *MsgSellTokens {
	minPriceInt, _ := math.NewIntFromString(minPrice)
	return &MsgSellTokens{
		Seller:   seller,
		Denom:    denom,
		Amount:   amount,
		MinPrice: minPriceInt,
	}
}

func (msg *MsgSellTokens) Route() string {
	return RouterKey
}

func (msg *MsgSellTokens) Type() string {
	return TypeMsgSellTokens
}

func (msg *MsgSellTokens) GetSigners() []sdk.AccAddress {
	seller, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{seller}
}

func (msg *MsgSellTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSellTokens) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid seller address (%s)", err)
	}

	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "invalid amount")
	}

	return nil
}

func NewMsgClaimFounderTokens(claimer, denom, amount string) *MsgClaimFounderTokens {
	amountInt, _ := math.NewIntFromString(amount)
	return &MsgClaimFounderTokens{
		Claimer: claimer,
		Denom:   denom,
		Amount:  amountInt,
	}
}

func (msg *MsgClaimFounderTokens) Route() string {
	return RouterKey
}

func (msg *MsgClaimFounderTokens) Type() string {
	return TypeMsgClaimFounderTokens
}

func (msg *MsgClaimFounderTokens) GetSigners() []sdk.AccAddress {
	claimer, err := sdk.AccAddressFromBech32(msg.Claimer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{claimer}
}

func (msg *MsgClaimFounderTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgClaimFounderTokens) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Claimer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid claimer address (%s)", err)
	}

	return nil
}

func NewMsgStartLBP(creator, denom string) *MsgStartLBP {
	return &MsgStartLBP{
		Creator: creator,
		Denom:   denom,
	}
}

func (msg *MsgStartLBP) Route() string {
	return RouterKey
}

func (msg *MsgStartLBP) Type() string {
	return TypeMsgStartLBP
}

func (msg *MsgStartLBP) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgStartLBP) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgStartLBP) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.Denom == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "denom cannot be empty")
	}

	return nil
}

// NewMsgCreateVestingAccount creates a new MsgCreateVestingAccount instance
func NewMsgCreateVestingAccount(creator, toAddress string, amount sdk.Coins, endTime int64, delayed bool) *MsgCreateVestingAccount {
	return &MsgCreateVestingAccount{
		Creator:   creator,
		ToAddress: toAddress,
		Amount:    amount,
		EndTime:   endTime,
		Delayed:   delayed,
	}
}

func (msg *MsgCreateVestingAccount) Route() string {
	return RouterKey
}

func (msg *MsgCreateVestingAccount) Type() string {
	return "create_vesting_account"
}

func (msg *MsgCreateVestingAccount) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateVestingAccount) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateVestingAccount) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid to address (%s)", err)
	}

	amount := sdk.Coins(msg.Amount)
	if !amount.IsValid() || amount.Empty() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "amount must be valid and not empty")
	}

	if msg.EndTime <= 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "end time must be positive")
	}

	return nil
}

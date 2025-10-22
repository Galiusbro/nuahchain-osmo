package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewMsgOpenPosition(owner, symbol string, side Side, quote string, leverage string) *MsgOpenPosition {
	return &MsgOpenPosition{
		Owner:         owner,
		Symbol:        symbol,
		Side:          side,
		Quote_NDOLLAR: quote,
		Leverage:      leverage,
	}
}

func (msg *MsgOpenPosition) ValidateBasic() error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %w", err)
	}
	if msg.Side != Side_SIDE_LONG && msg.Side != Side_SIDE_SHORT {
		return fmt.Errorf("invalid side")
	}
	if msg.Symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	amount, ok := sdkmath.NewIntFromString(msg.Quote_NDOLLAR)
	if !ok {
		return fmt.Errorf("invalid quote amount")
	}
	if !amount.IsPositive() {
		return fmt.Errorf("quote amount must be positive")
	}
	if _, err := sdkmath.LegacyNewDecFromStr(msg.Leverage); err != nil {
		return fmt.Errorf("invalid leverage: %w", err)
	}
	return nil
}

func (msg *MsgOpenPosition) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func NewMsgClosePosition(owner string, id uint64) *MsgClosePosition {
	return &MsgClosePosition{Owner: owner, Id: id}
}

func (msg *MsgClosePosition) ValidateBasic() error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %w", err)
	}
	if msg.Id == 0 {
		return fmt.Errorf("id must be positive")
	}
	return nil
}

func (msg *MsgClosePosition) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

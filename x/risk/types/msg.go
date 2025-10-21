package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewMsgSetRiskParams creates a new MsgSetRiskParams instance.
func NewMsgSetRiskParams(authority string, params *RiskParams) *MsgSetRiskParams {
	return &MsgSetRiskParams{
		Authority: authority,
		Params:    params,
	}
}

// ValidateBasic performs stateless validation on the message fields.
func (msg *MsgSetRiskParams) ValidateBasic() error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}
	if msg.Params == nil {
		return fmt.Errorf("params cannot be nil")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return fmt.Errorf("invalid authority address: %w", err)
	}

	return ValidateRiskParams(msg.Params)
}

// GetSigners returns the expected message signers.
func (msg *MsgSetRiskParams) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ValidateBasic performs basic validation for MsgUpdateParams
func (msg *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	return msg.Params.Validate()
}

// ValidateBasic performs basic validation for MsgUpdateUSDPrice
func (msg *MsgUpdateUSDPrice) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	if msg.Price.IsNil() || msg.Price.IsNegative() {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "price must be positive")
	}

	if len(msg.Source) == 0 {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "source cannot be empty")
	}

	return nil
}

// ValidateBasic performs basic validation for MsgSetPriceSources
func (msg *MsgSetPriceSources) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	if len(msg.Sources) == 0 {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "sources cannot be empty")
	}

	for _, source := range msg.Sources {
		if len(source.Name) == 0 {
			return errors.Wrap(sdkerrors.ErrInvalidRequest, "source name cannot be empty")
		}
		if source.Weight.IsNil() || source.Weight.IsNegative() {
			return errors.Wrap(sdkerrors.ErrInvalidRequest, "source weight must be positive")
		}
	}

	return nil
}

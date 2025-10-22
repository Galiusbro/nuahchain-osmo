package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgOpenPosition{}, "leverage/MsgOpenPosition", nil)
	cdc.RegisterConcrete(&MsgClosePosition{}, "leverage/MsgClosePosition", nil)
}

func RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	reg.RegisterImplementations((*sdk.Msg)(nil), &MsgOpenPosition{}, &MsgClosePosition{})

	// Register query service
	reg.RegisterImplementations((*sdk.Msg)(nil))
}

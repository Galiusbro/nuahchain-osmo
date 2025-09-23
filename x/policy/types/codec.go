package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterCodec registers msgs with the legacy Amino codec.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreatePolicy{}, "policy/MsgCreatePolicy", nil)
	cdc.RegisterConcrete(&MsgUpdatePolicyAttributes{}, "policy/MsgUpdatePolicyAttributes", nil)
	cdc.RegisterConcrete(&MsgCancelPolicy{}, "policy/MsgCancelPolicy", nil)
	cdc.RegisterConcrete(&MsgUpdatePolicyStatus{}, "policy/MsgUpdatePolicyStatus", nil)
}

// RegisterInterfaces registers sdk.Msg implementations.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreatePolicy{},
		&MsgUpdatePolicyAttributes{},
		&MsgCancelPolicy{},
		&MsgUpdatePolicyStatus{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(Amino)
	Amino.Seal()
}

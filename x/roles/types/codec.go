package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterCodec registers concrete message types on the Amino codec.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgAssignRoles{}, "roles/MsgAssignRoles", nil)
	cdc.RegisterConcrete(&MsgRevokeRoles{}, "roles/MsgRevokeRoles", nil)
	cdc.RegisterConcrete(&MsgUpdateAuthority{}, "roles/MsgUpdateAuthority", nil)
}

// RegisterInterfaces registers the sdk.Msg implementations.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAssignRoles{},
		&MsgRevokeRoles{},
		&MsgUpdateAuthority{},
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

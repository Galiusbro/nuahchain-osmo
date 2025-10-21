package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())

// RegisterCodec registers the amino interfaces for the module.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgDeposit{}, "collateral/MsgDeposit", nil)
	cdc.RegisterConcrete(&MsgWithdraw{}, "collateral/MsgWithdraw", nil)
}

// RegisterInterfaces registers interface implementations with the interface registry.
func RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	reg.RegisterImplementations((*sdk.Msg)(nil), &MsgDeposit{}, &MsgWithdraw{})
}

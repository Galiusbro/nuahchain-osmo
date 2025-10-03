package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgOpenPosition{}, "leverage/MsgOpenPosition", nil)
	cdc.RegisterConcrete(&MsgClosePosition{}, "leverage/MsgClosePosition", nil)
	cdc.RegisterConcrete(&MsgAddCollateral{}, "leverage/MsgAddCollateral", nil)
	cdc.RegisterConcrete(&MsgRemoveCollateral{}, "leverage/MsgRemoveCollateral", nil)
	cdc.RegisterConcrete(&MsgLiquidatePosition{}, "leverage/MsgLiquidatePosition", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "leverage/MsgUpdateParams", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgOpenPosition{},
		&MsgClosePosition{},
		&MsgAddCollateral{},
		&MsgRemoveCollateral{},
		&MsgLiquidatePosition{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(Amino)
	Amino.Seal()
}

// RegisterLegacyAminoCodec registers the necessary x/leverage interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	RegisterCodec(cdc)
}

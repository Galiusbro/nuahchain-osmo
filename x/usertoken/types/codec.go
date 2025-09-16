package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateUserToken{}, "usertoken/MsgCreateUserToken", nil)
	cdc.RegisterConcrete(&MsgBuyTokens{}, "usertoken/MsgBuyTokens", nil)
	cdc.RegisterConcrete(&MsgSellTokens{}, "usertoken/MsgSellTokens", nil)
	cdc.RegisterConcrete(&MsgClaimFounderTokens{}, "usertoken/MsgClaimFounderTokens", nil)
	cdc.RegisterConcrete(&MsgStartLBP{}, "usertoken/MsgStartLBP", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateUserToken{},
		&MsgBuyTokens{},
		&MsgSellTokens{},
		&MsgClaimFounderTokens{},
		&MsgStartLBP{},
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

// RegisterLegacyAminoCodec registers the necessary x/usertoken interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	RegisterCodec(cdc)
}

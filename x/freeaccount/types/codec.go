package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// RegisterInterfaces registers the FreeAccount type with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*authtypes.AccountI)(nil),
		&FreeAccount{},
	)
	
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateFreeAccount{},
	)
}

// RegisterLegacyAminoCodec registers the FreeAccount type with the amino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&FreeAccount{}, "osmosis/FreeAccount", nil)
	cdc.RegisterConcrete(&MsgCreateFreeAccount{}, "osmosis/MsgCreateFreeAccount", nil)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(Amino)
	registry := codectypes.NewInterfaceRegistry()
	RegisterInterfaces(registry)
	ModuleCdc = codec.NewProtoCodec(registry)
	Amino.Seal()
}
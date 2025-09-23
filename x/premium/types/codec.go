package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreatePremiumPlan{}, "premium/MsgCreatePremiumPlan", nil)
	cdc.RegisterConcrete(&MsgRecordPremiumPayment{}, "premium/MsgRecordPremiumPayment", nil)
	cdc.RegisterConcrete(&MsgMarkPremiumOverdue{}, "premium/MsgMarkPremiumOverdue", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreatePremiumPlan{},
		&MsgRecordPremiumPayment{},
		&MsgMarkPremiumOverdue{},
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

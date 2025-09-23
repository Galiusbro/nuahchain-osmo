package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterCodec registers module messages on the legacy Amino codec.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateTreasuryPool{}, "treasury/MsgCreateTreasuryPool", nil)
	cdc.RegisterConcrete(&MsgUpdateTreasuryPool{}, "treasury/MsgUpdateTreasuryPool", nil)
	cdc.RegisterConcrete(&MsgDepositToTreasury{}, "treasury/MsgDepositToTreasury", nil)
	cdc.RegisterConcrete(&MsgWithdrawFromTreasury{}, "treasury/MsgWithdrawFromTreasury", nil)
	cdc.RegisterConcrete(&MsgSetPoolReserves{}, "treasury/MsgSetPoolReserves", nil)
}

// RegisterInterfaces registers sdk.Msg implementations.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateTreasuryPool{},
		&MsgUpdateTreasuryPool{},
		&MsgDepositToTreasury{},
		&MsgWithdrawFromTreasury{},
		&MsgSetPoolReserves{},
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

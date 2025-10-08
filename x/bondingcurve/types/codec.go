package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgBuyFromCurve{}, "osmosis/bondingcurve/buy")
	legacy.RegisterAminoMsg(cdc, &MsgSellToCurve{}, "osmosis/bondingcurve/sell")
	legacy.RegisterAminoMsg(cdc, &MsgOpenMarginPosition{}, "osmosis/bondingcurve/open-margin")
	legacy.RegisterAminoMsg(cdc, &MsgCloseMarginPosition{}, "osmosis/bondingcurve/close-margin")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgBuyFromCurve{},
		&MsgSellToCurve{},
		&MsgOpenMarginPosition{},
		&MsgCloseMarginPosition{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(Amino)
	Amino.Seal()
}

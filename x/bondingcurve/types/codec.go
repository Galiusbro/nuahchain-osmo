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
	legacy.RegisterAminoMsg(cdc, &MsgLiquidateMarginPosition{}, "osmosis/bondingcurve/liquidate-margin")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "osmosis/bondingcurve/update-params")
	legacy.RegisterAminoMsg(cdc, &MsgSetEmergencyPause{}, "osmosis/bondingcurve/emergency-pause")
	legacy.RegisterAminoMsg(cdc, &MsgSetTokenPause{}, "osmosis/bondingcurve/token-pause")
	legacy.RegisterAminoMsg(cdc, &MsgForceLiquidation{}, "osmosis/bondingcurve/force-liquidation")
	legacy.RegisterAminoMsg(cdc, &MsgSetFreeze{}, "osmosis/bondingcurve/set-freeze")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgBuyFromCurve{},
		&MsgSellToCurve{},
		&MsgOpenMarginPosition{},
		&MsgCloseMarginPosition{},
		&MsgLiquidateMarginPosition{},
		&MsgUpdateParams{},
		&MsgSetEmergencyPause{},
		&MsgSetTokenPause{},
		&MsgForceLiquidation{},
		&MsgSetFreeze{},
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

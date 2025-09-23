package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterCodec registers concrete types on the Amino codec.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSubmitClaim{}, "claims/MsgSubmitClaim", nil)
	cdc.RegisterConcrete(&MsgReviewClaim{}, "claims/MsgReviewClaim", nil)
	cdc.RegisterConcrete(&MsgAddClaimEvidence{}, "claims/MsgAddClaimEvidence", nil)
	cdc.RegisterConcrete(&MsgExecuteClaimPayout{}, "claims/MsgExecuteClaimPayout", nil)
}

// RegisterInterfaces registers interface implementations.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitClaim{},
		&MsgReviewClaim{},
		&MsgAddClaimEvidence{},
		&MsgExecuteClaimPayout{},
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

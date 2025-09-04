//go:build test_amino
// +build test_amino

package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// MakeEncodingConfig creates an EncodingConfig for an amino based test configuration.
func MakeEncodingConfig() EncodingConfig {
	amino := codec.NewLegacyAmino()
	interfaceRegistry := testutil.CodecOptions{AccAddressPrefix: "nuah", ValAddressPrefix: "nuahvaloper"}.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(marshaler, tx.DefaultSignModes)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "bondingcurve"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName

	MemStoreKey = "mem_bondingcurve"
)

var (
	TokenPoolKeyPrefix      = []byte{0x01}
	MarginPoolKeyPrefix     = []byte{0x02}
	MarginPositionKeyPrefix = []byte{0x03}
	MarginTraderIndexPrefix = []byte{0x04}
	GlobalPauseKey          = []byte{0x05}
	TokenPauseKeyPrefix     = []byte{0x06}
	FreezeKeyPrefix         = []byte{0x07}
	PendingParamsKey        = []byte{0x08}
	EmergencyActionSeqKey   = []byte{0x09}
	EmergencyActionKeyPrefix = []byte{0x0A}
	EmergencyConfigKey      = []byte{0x0B}
)

func TokenPoolKey(denom string) []byte {
	return append(TokenPoolKeyPrefix, []byte(denom)...)
}

func MarginPoolKey(denom string) []byte {
	return append(MarginPoolKeyPrefix, []byte(denom)...)
}

func MarginPositionKey(id uint64) []byte {
	return append(MarginPositionKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func MarginPositionsByTraderPrefix(trader sdk.AccAddress) []byte {
	return append(MarginTraderIndexPrefix, trader.Bytes()...)
}

func MarginPositionTraderIndexKey(trader sdk.AccAddress, id uint64) []byte {
	return append(MarginPositionsByTraderPrefix(trader), sdk.Uint64ToBigEndian(id)...)
}

func TokenPauseKey(denom string) []byte {
	return append(TokenPauseKeyPrefix, []byte(denom)...)
}

func FreezeKey(targetType FreezeTargetType, target string) []byte {
	typeKey := []byte{byte(targetType)}
	return append(append(FreezeKeyPrefix, typeKey...), []byte(target)...)
}

func EmergencyActionKey(id uint64) []byte {
	return append(EmergencyActionKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

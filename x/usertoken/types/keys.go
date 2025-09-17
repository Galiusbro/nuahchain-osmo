package types

const (
	// ModuleName defines the module name
	ModuleName = "usertoken"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_usertoken"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// Store key prefixes
var (
	// UserTokenKeyPrefix is the prefix to retrieve all UserToken
	UserTokenKeyPrefix = []byte{0x01}
	// FounderTrancheKeyPrefix is the prefix to retrieve all FounderTranche
	FounderTrancheKeyPrefix = []byte{0x02}
	// ReferralProgramKeyPrefix is the prefix to retrieve all ReferralProgram
	ReferralProgramKeyPrefix = []byte{0x03}
	// ReferralActivationKeyPrefix is the prefix to retrieve all ReferralActivation
	ReferralActivationKeyPrefix = []byte{0x04}
)

// UserTokenKey returns the store key to retrieve a UserToken from the index fields
func UserTokenKey(denom string) []byte {
	return append(UserTokenKeyPrefix, []byte(denom)...)
}

// FounderTrancheKey returns the store key to retrieve a FounderTranche from the index fields
func FounderTrancheKey(denom string) []byte {
	return append(FounderTrancheKeyPrefix, []byte(denom)...)
}

// ReferralProgramKey returns the store key to retrieve a ReferralProgram from the token denom
func ReferralProgramKey(tokenDenom string) []byte {
	return append(ReferralProgramKeyPrefix, []byte(tokenDenom)...)
}

// ReferralActivationKey returns the store key to retrieve a ReferralActivation from the link ID
func ReferralActivationKey(linkId string) []byte {
	return append(ReferralActivationKeyPrefix, []byte(linkId)...)
}

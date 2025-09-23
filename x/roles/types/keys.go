package types

const (
	ModuleName  = "roles"
	StoreKey    = ModuleName
	RouterKey   = ModuleName
	MemStoreKey = "mem_roles"
)

var (
	RoleBindingKeyPrefix = []byte{0x01}
)

// RoleBindingKey builds the key used to store role bindings.
func RoleBindingKey(address []byte) []byte {
	return append(RoleBindingKeyPrefix, address...)
}

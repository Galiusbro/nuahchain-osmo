package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v30/x/roles/types"
)

// Keeper manages role assignments and provides role lookup helpers to other modules.
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramstore paramtypes.Subspace
	authority  string
}

// NewKeeper constructs a new roles keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	ps paramtypes.Subspace,
	authority string,
) Keeper {
	if ps.Name() != "" && !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:        cdc,
		storeKey:   key,
		paramstore: ps,
		authority:  authority,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams fetches the module parameters from the store.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	if k.paramstore.Name() == "" {
		return types.DefaultParams()
	}

	// store := ctx.KVStore(k.storeKey)
	// if !store.Has(types.KeyAuthority) {
	// 	params := types.DefaultParams()
	// 	if params.Authority == "" {
	// 		params.Authority = k.authority
	// 	}
	// 	return params
	// }

	var params types.Params
	// Check if parameters exist in the store before trying to get them
	if k.paramstore.Has(ctx, types.KeyAuthority) {
		k.paramstore.GetParamSet(ctx, &params)
	} else {
		// If no parameters exist, return default params with keeper's authority
		params = types.DefaultParams()
	}

	if params.Authority == "" {
		params.Authority = k.authority
	}

	return params
}

// SetParams updates the module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if k.paramstore.Name() == "" {
		return
	}
	k.paramstore.SetParamSet(ctx, &params)
}

// GetAuthority resolves the operative authority account.
func (k Keeper) GetAuthority(ctx sdk.Context) string {
	params := k.GetParams(ctx)
	if params.Authority != "" {
		return params.Authority
	}
	return k.authority
}

// setAuthority writes the provided authority into the params store.
func (k Keeper) setAuthority(ctx sdk.Context, authority string) {
	params := k.GetParams(ctx)
	params.Authority = authority
	k.SetParams(ctx, params)
}

// AssignRoles assigns the provided roles to an address.
func (k Keeper) AssignRoles(ctx sdk.Context, authority sdk.AccAddress, addr sdk.AccAddress, roles []types.Role) error {
	if authority.String() != k.GetAuthority(ctx) {
		return types.ErrUnauthorized
	}

	if len(roles) == 0 {
		return fmt.Errorf("no roles supplied")
	}

	binding, found := k.getRoleBinding(ctx, addr)
	if !found {
		binding = types.RoleBinding{
			Address: addr.String(),
		}
	}

	existing := make(map[types.Role]struct{}, len(binding.Roles))
	for _, role := range binding.Roles {
		existing[role] = struct{}{}
	}

	modified := false
	for _, role := range roles {
		if role == types.Role_ROLE_UNSPECIFIED {
			return types.ErrUnknownRole(role)
		}
		if _, ok := existing[role]; ok {
			continue
		}
		existing[role] = struct{}{}
		binding.Roles = append(binding.Roles, role)
		modified = true
	}

	if !modified {
		return nil
	}

	k.setRoleBinding(ctx, binding)

	em := ctx.EventManager()
	for _, role := range roles {
		em.EmitEvent(sdk.NewEvent(
			types.EventTypeRoleAssigned,
			sdk.NewAttribute(types.AttributeKeyAuthority, authority.String()),
			sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
			sdk.NewAttribute(types.AttributeKeyRole, role.String()),
		))
	}

	return nil
}

// RevokeRoles removes roles from an address.
func (k Keeper) RevokeRoles(ctx sdk.Context, authority sdk.AccAddress, addr sdk.AccAddress, roles []types.Role) error {
	if authority.String() != k.GetAuthority(ctx) {
		return types.ErrUnauthorized
	}

	if len(roles) == 0 {
		return fmt.Errorf("no roles supplied")
	}

	binding, found := k.getRoleBinding(ctx, addr)
	if !found {
		return types.ErrRoleNotFound
	}

	roleSet := make(map[types.Role]struct{}, len(binding.Roles))
	for _, role := range binding.Roles {
		roleSet[role] = struct{}{}
	}

	changed := false
	for _, role := range roles {
		if _, ok := roleSet[role]; !ok {
			continue
		}
		delete(roleSet, role)
		changed = true
	}

	if !changed {
		return nil
	}

	binding.Roles = binding.Roles[:0]
	for role := range roleSet {
		binding.Roles = append(binding.Roles, role)
	}

	if len(binding.Roles) == 0 {
		k.deleteRoleBinding(ctx, addr)
	} else {
		k.setRoleBinding(ctx, binding)
	}

	em := ctx.EventManager()
	for _, role := range roles {
		em.EmitEvent(sdk.NewEvent(
			types.EventTypeRoleRevoked,
			sdk.NewAttribute(types.AttributeKeyAuthority, authority.String()),
			sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
			sdk.NewAttribute(types.AttributeKeyRole, role.String()),
		))
	}

	return nil
}

// HasRole returns whether the address is associated with the provided role.
func (k Keeper) HasRole(ctx sdk.Context, addr sdk.AccAddress, role types.Role) bool {
	binding, found := k.getRoleBinding(ctx, addr)
	if !found {
		return false
	}

	for _, r := range binding.Roles {
		if r == role {
			return true
		}
	}

	return false
}

// GetRoleBinding retrieves the stored role binding for an address.
func (k Keeper) GetRoleBinding(ctx sdk.Context, addr sdk.AccAddress) (types.RoleBinding, bool) {
	return k.getRoleBinding(ctx, addr)
}

// GetAllRoleBindings returns all role bindings.
func (k Keeper) GetAllRoleBindings(ctx sdk.Context) []types.RoleBinding {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.RoleBindingKeyPrefix)
	defer iterator.Close()

	bindings := []types.RoleBinding{}

	for ; iterator.Valid(); iterator.Next() {
		var binding types.RoleBinding
		k.cdc.MustUnmarshal(iterator.Value(), &binding)
		bindings = append(bindings, binding)
	}

	return bindings
}

// UpdateAuthority changes the module authority.
func (k Keeper) UpdateAuthority(ctx sdk.Context, authority sdk.AccAddress, newAuthority string) error {
	if authority.String() != k.GetAuthority(ctx) {
		return types.ErrUnauthorized
	}

	if err := types.ValidateAuthority(newAuthority); err != nil {
		return err
	}

	k.setAuthority(ctx, newAuthority)
	return nil
}

func (k Keeper) getRoleBinding(ctx sdk.Context, addr sdk.AccAddress) (types.RoleBinding, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.RoleBindingKey(addr))
	if bz == nil {
		return types.RoleBinding{}, false
	}

	var binding types.RoleBinding
	k.cdc.MustUnmarshal(bz, &binding)
	return binding, true
}

func (k Keeper) setRoleBinding(ctx sdk.Context, binding types.RoleBinding) {
	store := ctx.KVStore(k.storeKey)
	addr, err := sdk.AccAddressFromBech32(binding.Address)
	if err != nil {
		panic(fmt.Errorf("invalid address stored: %s", err))
	}

	bz := k.cdc.MustMarshal(&binding)
	store.Set(types.RoleBindingKey(addr), bz)
}

func (k Keeper) SetRoleBinding(ctx sdk.Context, binding types.RoleBinding) error {
	if binding.Address == "" {
		return fmt.Errorf("binding address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(binding.Address); err != nil {
		return err
	}
	k.setRoleBinding(ctx, binding)
	return nil
}

func (k Keeper) deleteRoleBinding(ctx sdk.Context, addr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.RoleBindingKey(addr))
}

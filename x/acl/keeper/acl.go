package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/acl/types"
)

func (k Keeper) SetAllowed(ctx sdk.Context, addr sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllowed)
	store.Set(addr.Bytes(), []byte{0}) //TODO empty struct
}

func (k Keeper) Allowed(ctx sdk.Context, addr sdk.AccAddress) bool {
	var enabled bool
	k.paramSpace.Get(ctx, types.ParamStoreKeyEnable, &enabled)
	if !enabled {
		return true
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllowed)
	return store.Has(addr.Bytes())
}

func (k Keeper) ExportAllowed(ctx sdk.Context) (addresses []string) {
	iterator := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllowed).Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		addr := sdk.AccAddress(iterator.Key())
		addresses = append(addresses, addr.String())
	}
	return addresses
}

func (k Keeper) SetAdmin(ctx sdk.Context, addr sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdmins)
	store.Set(addr.Bytes(), []byte{0}) //TODO
}

func (k Keeper) IsAdmin(ctx sdk.Context, addr sdk.AccAddress) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdmins)
	return store.Has(addr.Bytes())
}

func (k Keeper) ExportAdmins(ctx sdk.Context) (addresses []string) {
	iterator := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdmins).Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		addr := sdk.AccAddress(iterator.Key())
		addresses = append(addresses, addr.String())
	}
	return addresses
}

func (k Keeper) Enabled(ctx sdk.Context) (enable bool) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyEnable, &enable)
	return
}

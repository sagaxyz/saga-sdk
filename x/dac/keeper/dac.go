package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/sagaxyz/sagaevm/v8/x/dac/types"
)

func (k Keeper) SetAllowed(ctx sdk.Context, addr common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllowed)
	store.Set(addr.Bytes(), []byte{})
}

func (k Keeper) Allowed(ctx sdk.Context, addr common.Address) bool {
	var enabled bool
	k.paramSpace.Get(ctx, types.ParamStoreKeyEnable, &enabled)
	if !enabled {
		return true
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllowed)
	return store.Has(addr.Bytes())
}

func (k Keeper) ExportAllowed(ctx sdk.Context) []string {
	iterator := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllowed).Iterator(nil, nil)
	defer iterator.Close()

	var addresses []string
	for ; iterator.Valid(); iterator.Next() {
		addr := common.BytesToAddress(iterator.Key())
		addresses = append(addresses, addr.Hex())
	}
	return addresses
}

func (k Keeper) SetAdmin(ctx sdk.Context, addr sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdmins)
	store.Set(addr.Bytes(), []byte{})
}

func (k Keeper) Admin(ctx sdk.Context, addr sdk.AccAddress) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdmins)
	return store.Has(addr.Bytes())
}

func (k Keeper) ExportAdmins(ctx sdk.Context) []string {
	iterator := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdmins).Iterator(nil, nil)
	defer iterator.Close()

	var addresses []string
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

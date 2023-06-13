package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/dac/types"
)

func (k Keeper) SetAllowed(ctx sdk.Context, addr *types.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllowed)
	store.Set(addr.Bytes(), []byte{byte(addr.Format)})
}

func (k Keeper) Allowed(ctx sdk.Context, addr *types.Address) bool {
	var enabled bool
	k.paramSpace.Get(ctx, types.ParamStoreKeyEnable, &enabled)
	if !enabled {
		return true
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllowed)
	return store.Has(addr.Bytes())
}

func (k Keeper) ExportAllowed(ctx sdk.Context) []*types.Address {
	iterator := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllowed).Iterator(nil, nil)
	defer iterator.Close()

	var addresses []*types.Address
	for ; iterator.Valid(); iterator.Next() {
		format := types.AddressFormat(iterator.Value()[0]) //TODO
		addr, err := types.LoadAddress(format, iterator.Key())
		if err != nil {
			panic(fmt.Sprintf("store contains an invalid address '%s' (%s)", iterator.Key(), format))
		}
		addresses = append(addresses, addr)
	}
	return addresses
}

func (k Keeper) SetAdmin(ctx sdk.Context, addr sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdmins)
	store.Set(addr.Bytes(), []byte{byte(types.AddressFormat_ADDRESS_BECH32)})
}

func (k Keeper) Admin(ctx sdk.Context, addr sdk.AccAddress) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdmins)
	return store.Has(addr.Bytes())
}

func (k Keeper) ExportAdmins(ctx sdk.Context) []*types.Address {
	iterator := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdmins).Iterator(nil, nil)
	defer iterator.Close()

	var addresses []*types.Address
	for ; iterator.Valid(); iterator.Next() {
		format := types.AddressFormat(iterator.Value()[0]) //TODO
		addr, err := types.LoadAddress(format, iterator.Key())
		if err != nil {
			panic(fmt.Sprintf("store contains an invalid address '%s' (%s)", iterator.Key(), format))
		}
		addresses = append(addresses, addr)
	}
	return addresses
}

func (k Keeper) Enabled(ctx sdk.Context) (enable bool) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyEnable, &enable)
	return
}

package keeper

import (
	"github.com/sagaxyz/saga-sdk/x/feedistribution/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns the total set of feedistribution parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.ParamsKey)
	k.cdc.MustUnmarshal(bz, &params)

	return
}

// SetParams sets the feedistribution params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.storeKey)

	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.ParamsKey, bz)

	return nil
}

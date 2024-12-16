package keeper

import (
	"github.com/sagaxyz/saga-sdk/x/filter/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns the total set of filter parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if len(bz) == 0 {
		var p types.Params
		k.ss.GetParamSetIfExists(ctx, &p)
		return p
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the filter params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.ParamsKey, bz)

	return nil
}

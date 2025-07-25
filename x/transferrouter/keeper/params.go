package keeper

import (
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns the module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	// TODO: once KVStoreService schema finalized, retrieve parameters
	return types.Params{}
}

// SetParams sets the module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	// NOTE: Params persistence not yet implemented with StoreService.
	return nil
}

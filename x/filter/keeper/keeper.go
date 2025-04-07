package keeper

import (
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/sagaxyz/saga-sdk/x/filter/types"
)

// Keeper grants access to the Filter module state.
type Keeper struct {
	// Protobuf codec
	cdc codec.BinaryCodec
	// Store key required for the Filter Prefix KVStore.
	storeKey storetypes.StoreKey
	// Legacy subspace
	ss paramstypes.Subspace
	// Authority to change params
	authority string
}

// New generates new filter module keeper
func New(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, ss paramstypes.Subspace, authority string) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		ss:        ss,
		authority: authority,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

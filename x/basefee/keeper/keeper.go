package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	ethermintevmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/sagaxyz/saga-sdk/x/basefee/types"
)

// Keeper grants access to the Fee Market module state.
type Keeper struct {
	// Protobuf codec
	cdc codec.BinaryCodec
	// Store key required for the Fee Market Prefix KVStore.
	storeKey     storetypes.StoreKey
	transientKey storetypes.StoreKey
	// the address capable of executing a MsgUpdateParams message. Typically, this should be the x/gov module account.
	authority sdk.AccAddress
	// Legacy subspace
	ss paramstypes.Subspace

	feeMarketKeeper ethermintevmtypes.FeeMarketKeeper
}

// New generates new basefee module keeper
func New(cdc codec.BinaryCodec, authority sdk.AccAddress, storeKey, transientKey storetypes.StoreKey, ss paramstypes.Subspace, fmk ethermintevmtypes.FeeMarketKeeper) Keeper {
	// ensure authority account is correctly formatted
	if err := sdk.VerifyAddressFormat(authority); err != nil {
		panic(err)
	}

	return Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		authority:       authority,
		transientKey:    transientKey,
		ss:              ss,
		feeMarketKeeper: fmk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

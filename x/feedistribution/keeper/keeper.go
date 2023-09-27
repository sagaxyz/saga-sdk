package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/sagaxyz/saga-sdk/x/feedistribution/types"
)

// Keeper grants access to the Fee Market module state.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	// The address (e.g. x/gov module account) capable of changing the params
	authority sdk.AccAddress

	authKeeper types.AccountKeeper
	bankKeeper types.BankKeeper

	// Name of the FeeCollector ModuleAccount
	feeCollectorName string
}

// New generates a new feedistribution module keeper
func New(cdc codec.BinaryCodec, authority sdk.AccAddress, storeKey storetypes.StoreKey,
	ak types.AccountKeeper, bk types.BankKeeper, feeCollectorName string) Keeper {
	// Ensure authority account is correctly formatted
	if err := sdk.VerifyAddressFormat(authority); err != nil {
		panic(err)
	}

	return Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		authority:        authority,
		authKeeper:       ak,
		bankKeeper:       bk,
		feeCollectorName: feeCollectorName,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

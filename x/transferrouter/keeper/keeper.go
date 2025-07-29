package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	corestore "cosmossdk.io/core/store"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

var (
	ParamsPrefix           = collections.NewPrefix(0) // Stores params
	CallQueuePrefix        = collections.NewPrefix(1) // Stores the call queue
	LastCallSequencePrefix = collections.NewPrefix(2) // Stores the last call sequence
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService corestore.KVStoreService
	authority    string

	Schema           collections.Schema
	Params           collections.Item[types.Params]
	CallQueue        collections.Map[uint64, types.CallQueueItem]
	LastCallSequence collections.Item[uint64]

	ics4Wrapper porttypes.ICS4Wrapper
}

// New returns a new Keeper instance.
func NewKeeper(cdc codec.BinaryCodec, storeSvc corestore.KVStoreService, authority string) Keeper {

	sb := collections.NewSchemaBuilder(storeSvc)
	k := Keeper{
		cdc:          cdc,
		storeService: storeSvc,
		authority:    authority,
		Params: collections.NewItem(
			sb,
			ParamsPrefix,
			"params",
			codec.CollValue[types.Params](cdc),
		),
		CallQueue: collections.NewMap(
			sb,
			CallQueuePrefix,
			"call_queue",
			collections.Uint64Key,
			codec.CollValue[types.CallQueueItem](cdc),
		),
		LastCallSequence: collections.NewItem(
			sb,
			LastCallSequencePrefix,
			"last_call_sequence",
			collections.Uint64Value,
		),
	}

	var err error
	k.Schema, err = sb.Build()
	if err != nil {
		panic(err)
	}

	return k
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	corestore "cosmossdk.io/core/store"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

var (
	ParamsPrefix           = collections.NewPrefix(0) // Stores params
	CallQueuePrefix        = collections.NewPrefix(1) // Stores the call queue
	CallQueueHashPrefix    = collections.NewPrefix(2) // Stores the call queue hash
	LastCallSequencePrefix = collections.NewPrefix(3) // Stores the last call sequence
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService corestore.KVStoreService
	authority    string

	Schema           collections.Schema
	Params           collections.Item[types.Params]
	CallQueue        *collections.IndexedMap[uint64, types.CallQueueItem, CallQueIndexes]
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
		CallQueue: collections.NewIndexedMap(
			sb,
			CallQueuePrefix,
			"call_queue",
			collections.Uint64Key,
			codec.CollValue[types.CallQueueItem](cdc),
			NewCallQueIndexes(sb),
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

// GetCallQueueItemByHash returns the call queue item by hash (in hex format)
func (k Keeper) GetCallQueueItemByHash(ctx sdk.Context, hash string) (seq uint64, item types.CallQueueItem, found bool) {
	seq, err := k.CallQueue.Indexes.Hash.MatchExact(ctx, hash)
	if err != nil {
		return 0, types.CallQueueItem{}, false
	}

	item, err = k.CallQueue.Get(ctx, seq)
	if err != nil {
		return 0, types.CallQueueItem{}, false
	}

	return seq, item, true
}

// WriteIBCAcknowledgment writes the IBC acknowledgment for the call queue item
func (k Keeper) WriteIBCAcknowledgment(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

// Indexes
type CallQueIndexes struct {
	Hash *indexes.Unique[string, uint64, types.CallQueueItem]
}

func (c CallQueIndexes) IndexesList() []collections.Index[uint64, types.CallQueueItem] {
	return []collections.Index[uint64, types.CallQueueItem]{c.Hash}
}

func NewCallQueIndexes(sb *collections.SchemaBuilder) CallQueIndexes {
	return CallQueIndexes{
		Hash: indexes.NewUnique[string, uint64, types.CallQueueItem](sb, CallQueueHashPrefix, "callqueue_hash", collections.StringKey, collections.Uint64Key, func(pk uint64, v types.CallQueueItem) (string, error) {
			return v.ToMsgEthereumTx().Hash, nil
		}),
	}
}

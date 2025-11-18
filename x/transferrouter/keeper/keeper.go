package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/collections/corecompat"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"

	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

var (
	ParamsPrefix              = collections.NewPrefix(0) // Stores params
	PacketQueuePrefix         = collections.NewPrefix(2) // Stores the packets
	PacketResultPrefix        = collections.NewPrefix(3) // Stores the packet results
	SrcCallbackQueuePrefix    = collections.NewPrefix(4) // Stores the src callback queue
	ErrorOrTimeoutQueuePrefix = collections.NewPrefix(5) // Stores the error or timeout queue
	GlobalPacketSequenceKey   = collections.NewPrefix(6) // Stores the global packet sequence
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService corecompat.KVStoreService
	authority    string

	Schema               collections.Schema
	Params               collections.Item[types.Params]
	GlobalPacketSequence collections.Sequence
	PacketQueue          collections.Map[uint64, types.PacketQueueItem]
	SrcCallbackQueue     collections.Map[uint64, types.PacketQueueItem]
	ErrorOrTimeoutQueue  collections.Map[uint64, types.PacketQueueItem]

	Erc20Keeper    types.ERC20Keeper
	ChannelKeeper  types.ChannelKeeper
	TransferKeeper types.TransferKeeper
	BankKeeper     types.BankKeeper
	AccountKeeper  types.AccountKeeper
	EVMKeeper      types.EVMKeeper

	ics4Wrapper porttypes.ICS4Wrapper
}

// New returns a new Keeper instance.
func NewKeeper(cdc codec.BinaryCodec,
	storeSvc corecompat.KVStoreService,
	erc20Keeper types.ERC20Keeper,
	ics4Wrapper porttypes.ICS4Wrapper,
	channelKeeper types.ChannelKeeper,
	transferKeeper types.TransferKeeper,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	evmKeeper types.EVMKeeper,
	authority string) Keeper {

	sb := collections.NewSchemaBuilder(storeSvc)
	k := Keeper{
		cdc:            cdc,
		storeService:   storeSvc,
		authority:      authority,
		Erc20Keeper:    erc20Keeper,
		ChannelKeeper:  channelKeeper,
		TransferKeeper: transferKeeper,
		BankKeeper:     bankKeeper,
		AccountKeeper:  accountKeeper,
		EVMKeeper:      evmKeeper,
		ics4Wrapper:    ics4Wrapper,
		Params: collections.NewItem(
			sb,
			ParamsPrefix,
			"params",
			codec.CollValue[types.Params](cdc),
		),
		PacketQueue: collections.NewMap(
			sb,
			PacketQueuePrefix,
			"packet_queue",
			collections.Uint64Key,
			codec.CollValue[types.PacketQueueItem](cdc),
		),
		SrcCallbackQueue: collections.NewMap(
			sb,
			SrcCallbackQueuePrefix,
			"src_callback_queue",
			collections.Uint64Key,
			codec.CollValue[types.PacketQueueItem](cdc),
		),
		ErrorOrTimeoutQueue: collections.NewMap(
			sb,
			ErrorOrTimeoutQueuePrefix,
			"error_or_timeout_queue",
			collections.Uint64Key,
			codec.CollValue[types.PacketQueueItem](cdc),
		),
		GlobalPacketSequence: collections.NewSequence(
			sb,
			GlobalPacketSequenceKey,
			"global_packet_sequence",
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

// WriteIBCAcknowledgment writes the IBC acknowledgment for the call queue item.
// As we don't modify outgoing txs, we just pass this call to the original transferkeeper.
func (k Keeper) WriteIBCAcknowledgment(ctx sdk.Context, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, packet, ack)
}

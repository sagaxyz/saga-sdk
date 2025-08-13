package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	corestore "cosmossdk.io/core/store"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
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

type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetPacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64) []byte
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error)
}

type ERC20Keeper interface {
	GetCoinAddress(ctx sdk.Context, denom string) (common.Address, error)
}

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService corestore.KVStoreService
	authority    string

	Schema    collections.Schema
	Params    collections.Item[types.Params]
	CallQueue collections.Map[uint64, types.CallQueueItem]
	NextNonce collections.Item[uint64]

	Erc20Keeper   ERC20Keeper
	ChannelKeeper ChannelKeeper

	ics4Wrapper porttypes.ICS4Wrapper
}

// New returns a new Keeper instance.
func NewKeeper(cdc codec.BinaryCodec,
	storeSvc corestore.KVStoreService,
	erc20Keeper ERC20Keeper,
	ics4Wrapper porttypes.ICS4Wrapper,
	channelKeeper ChannelKeeper,
	authority string) Keeper {

	sb := collections.NewSchemaBuilder(storeSvc)
	k := Keeper{
		cdc:           cdc,
		storeService:  storeSvc,
		authority:     authority,
		Erc20Keeper:   erc20Keeper,
		ChannelKeeper: channelKeeper,
		ics4Wrapper:   ics4Wrapper,
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
		NextNonce: collections.NewItem(
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

// WriteIBCAcknowledgment writes the IBC acknowledgment for the call queue item
func (k Keeper) WriteIBCAcknowledgment(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

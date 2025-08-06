package middlewares

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	evmostypes "github.com/evmos/evmos/v20/x/evm/types"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

var _ sdk.PostDecorator = PostHandler{}

type PostHandler struct {
	keeper keeper.Keeper
}

func NewPostHandler(k keeper.Keeper) PostHandler {
	return PostHandler{
		keeper: k,
	}
}

// PostHandle implements types.PostDecorator.
func (p PostHandler) PostHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, success bool, next sdk.PostHandler) (newCtx sdk.Context, err error) {
	// Here we need to find the corresponding call in the call queue and remove it, also we need to write the acknowledgment if needed

	// 1. Find the corresponding call in the call queue
	msgs := tx.GetMsgs()
	var (
		callQueueItem types.CallQueueItem
		seq           uint64
		found         bool
	)

	for _, msg := range msgs {
		if sdk.MsgTypeURL(msg) == "/cosmos.evm.v1.MsgEthereumTx" {
			msgEthereumTx := msg.(*evmostypes.MsgEthereumTx)
			seq, callQueueItem, found = p.keeper.GetCallQueueItemByHash(ctx, msgEthereumTx.Hash)
			if found {
				break
			}
		}
	}

	// not a call queue item, continue
	if !found {
		return next(ctx, tx, simulate, success)
	}

	// 2. Write the IBC acknowledgment if needed
	p.keeper.Logger(ctx).Debug("writing IBC acknowledgment for call queue item", "seq", seq, "call_queue_item", callQueueItem)

	packet := channeltypes.Packet{
		Sequence:           seq,
		SourceChannel:      callQueueItem.InFlightPacket.PacketSrcChannelId,
		SourcePort:         callQueueItem.InFlightPacket.PacketSrcPortId,
		DestinationChannel: callQueueItem.InFlightPacket.RefundChannelId,
		DestinationPort:    callQueueItem.InFlightPacket.RefundPortId,
		Data:               callQueueItem.Call.Data,
		TimeoutHeight:      clienttypes.MustParseHeight(callQueueItem.InFlightPacket.PacketTimeoutHeight),
		TimeoutTimestamp:   callQueueItem.InFlightPacket.PacketTimeoutTimestamp,
	}

	ack := channeltypes.NewResultAcknowledgement([]byte{1})

	// took from ibc-go:
	// // Lookup module by channel capability
	// module, capability, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.Packet.DestinationPort, msg.Packet.DestinationChannel)
	// if err != nil {
	// 	ctx.Logger().Error("receive packet failed", "port-id", msg.Packet.SourcePort, "channel-id", msg.Packet.SourceChannel, "error", errorsmod.Wrap(err, "could not retrieve module from port-id"))
	// 	return nil, errorsmod.Wrap(err, "could not retrieve module from port-id")
	// }

	// TODO: missing chanCapability, but it has been removed in v10, so we might not need it
	err = p.keeper.WriteIBCAcknowledgment(ctx, nil, packet, ack)
	if err != nil {
		ctx.Logger().Error("failed to write IBC acknowledgment", "error", err)
		return newCtx, err
	}

	// 3. Remove the call from the call queue
	p.keeper.CallQueue.Remove(ctx, seq)

	return next(ctx, tx, simulate, success)
}

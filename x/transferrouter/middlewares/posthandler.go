package middlewares

import (
	"bytes"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	evmostypes "github.com/evmos/evmos/v20/x/evm/types"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

type PostHandler struct {
	keeper keeper.Keeper
}

func NewPostHandler(k keeper.Keeper) PostHandler {
	return PostHandler{
		keeper: k,
	}
}

// PostHandle implements types.PostDecorator.
func (p PostHandler) PostHandler() sdk.PostHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool, success bool) (newCtx sdk.Context, err error) {
		fmt.Println("PostHandler!!!! =================================")
		// Here we need to find the corresponding call in the call queue and remove it, also we need to write the acknowledgment if needed

		ctx.Logger().Info("=== POSTHANDLER START ===", "simulate", simulate, "success", success)
		ctx.Logger().Info("Transaction details", "height", ctx.BlockHeight())

		// 1. Find the corresponding call in the call queue
		msgs := tx.GetMsgs()
		ctx.Logger().Info("Processing messages", "num_msgs", len(msgs))

		var (
			callQueueItem types.CallQueueItem
			seq           uint64
			found         bool
		)

		// TODO: check performance
		for i, msg := range msgs {
			msgType := sdk.MsgTypeURL(msg)
			ctx.Logger().Info("Processing message", "index", i, "type", msgType)

			if msgType == "/ethermint.evm.v1.MsgEthereumTx" {
				ctx.Logger().Info("Found Ethereum transaction message")
				msgEthereumTx := msg.(*evmostypes.MsgEthereumTx)
				ctx.Logger().Info("Ethereum tx details", "hash", msgEthereumTx.Hash, "from", msgEthereumTx.From)
				// seq, callQueueItem, found = p.keeper.GetCallQueueItemByHash(ctx, msgEthereumTx.Hash)

				// increment nonce
				nonce := msgEthereumTx.AsTransaction().Nonce()
				err = p.keeper.NextNonce.Set(ctx, nonce+1)
				if err != nil {
					ctx.Logger().Error("failed to set last nonce", "error", err)
					return ctx, err
				}

				err = p.keeper.CallQueue.Walk(ctx, nil, func(key uint64, value types.CallQueueItem) (stop bool, err error) {
					// TODO: for now we do it with data?
					if bytes.Equal(value.ToMsgEthereumTx(nonce, big.NewInt(1234)).AsTransaction().Data(), msgEthereumTx.AsTransaction().Data()) {
						found = true
						seq = key
						callQueueItem = value
						return true, nil
					}
					return false, nil
				})
				if err != nil {
					ctx.Logger().Error("failed to walk call queue", "error", err)
					return ctx, err
				}
				ctx.Logger().Info("Call queue lookup result", "found", found, "seq", seq, "hash", msgEthereumTx.Hash)

				if found {
					ctx.Logger().Info("Found call queue item", "seq", seq, "call_queue_item", callQueueItem)
					break
				}
			} else {
				ctx.Logger().Info("Skipping non-Ethereum message", "type", msgType)
			}
		}

		// not a call queue item, continue
		if !found {
			ctx.Logger().Info("No call queue item found, returning early")
			return ctx, nil
		}

		ctx.Logger().Info("=== PROCESSING CALL QUEUE ITEM ===", "seq", seq)

		// 2. Write the IBC acknowledgment if needed
		p.keeper.Logger(ctx).Debug("writing IBC acknowledgment for call queue item", "seq", seq, "call_queue_item", callQueueItem)

		ctx.Logger().Info("Creating IBC packet from call queue item")
		ctx.Logger().Info("Call queue item details",
			"seq", seq,
			"src_channel", callQueueItem.InFlightPacket.PacketSrcChannelId,
			"src_port", callQueueItem.InFlightPacket.PacketSrcPortId,
			"dest_channel", callQueueItem.InFlightPacket.RefundChannelId,
			"dest_port", callQueueItem.InFlightPacket.RefundPortId,
			"timeout_height", callQueueItem.InFlightPacket.PacketTimeoutHeight,
			"timeout_timestamp", callQueueItem.InFlightPacket.PacketTimeoutTimestamp,
		)

		packet := channeltypes.Packet{
			Sequence:           seq,
			SourceChannel:      callQueueItem.InFlightPacket.PacketSrcChannelId,
			SourcePort:         callQueueItem.InFlightPacket.PacketSrcPortId,
			DestinationChannel: callQueueItem.InFlightPacket.RefundChannelId,
			DestinationPort:    callQueueItem.InFlightPacket.RefundPortId,
			Data:               callQueueItem.InFlightPacket.PacketData,
			TimeoutHeight:      clienttypes.MustParseHeight(callQueueItem.InFlightPacket.PacketTimeoutHeight),
			TimeoutTimestamp:   callQueueItem.InFlightPacket.PacketTimeoutTimestamp,
		}

		ctx.Logger().Info("Created IBC packet",
			"sequence", packet.Sequence,
			"source_channel", packet.SourceChannel,
			"source_port", packet.SourcePort,
			"dest_channel", packet.DestinationChannel,
			"dest_port", packet.DestinationPort,
			"data_length", len(packet.Data),
		)

		ack := channeltypes.NewResultAcknowledgement([]byte{1})
		ctx.Logger().Info("Created acknowledgment", "ack", ack)

		// took from ibc-go:
		// // Lookup module by channel capability
		_, capability, err := p.keeper.ChannelKeeper.LookupModuleByChannel(ctx, packet.DestinationPort, packet.DestinationChannel)
		if err != nil {
			ctx.Logger().Error("receive packet failed", "port-id", packet.DestinationPort, "channel-id", packet.DestinationChannel, "error", err)
			return ctx, err
		}

		// TODO: missing chanCapability, but it has been removed in v10, so we might not need it
		ctx.Logger().Info("Writing IBC acknowledgment...")
		err = p.keeper.WriteIBCAcknowledgment(ctx, capability, packet, ack)
		if err != nil {
			ctx.Logger().Error("failed to write IBC acknowledgment", "error", err)
			return newCtx, err
		}
		ctx.Logger().Info("Successfully wrote IBC acknowledgment")

		// 3. Remove the call from the call queue
		ctx.Logger().Info("Removing call from queue", "seq", seq)
		p.keeper.CallQueue.Remove(ctx, seq)
		ctx.Logger().Info("Successfully removed call from queue")

		ctx.Logger().Info("=== POSTHANDLER COMPLETE ===")
		return ctx, nil
	}

}

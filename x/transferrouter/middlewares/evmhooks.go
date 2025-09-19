package middlewares

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
)

type EvmHooks struct {
	k keeper.Keeper
}

func NewEvmHooks(k keeper.Keeper) evmtypes.EvmHooks {
	return &EvmHooks{k: k}
}

func (h *EvmHooks) PostTxProcessing(ctx sdk.Context, sender common.Address, msg core.Message, receipt *ethtypes.Receipt) error {
	h.k.Logger(ctx).Info("PostTxProcessing", "sender", sender, "msg", msg, "receipt", receipt)

	// chainId, err := evmostypes.ParseChainID(ctx.ChainID())
	// if err != nil {
	// 	h.k.Logger(ctx).Error("failed to parse chain id", "error", err)
	// 	return err
	// }

	// // only perform the following checks if the sender is the known signer
	// params, err := h.k.Params.Get(ctx)
	// if err != nil {
	// 	return err
	// }

	// knownSignerAddress, err := sdk.AccAddressFromBech32(params.KnownSignerAddress)
	// if err != nil {
	// 	return err
	// }

	// // Return early if the sender is not the known signer
	// if !bytes.Equal(sender.Bytes(), knownSignerAddress.Bytes()) {
	// 	h.k.Logger(ctx).Info("Sender is not the known signer, skipping post tx processing")
	// 	return nil
	// }

	// // find call

	// var (
	// 	callQueueItem types.CallQueueItem
	// 	seq           uint64
	// 	found         bool
	// )

	// err = h.k.CallQueue.Walk(ctx, nil, func(key uint64, value types.CallQueueItem) (stop bool, err error) {
	// 	// TODO: for now we do it with data, but we must make sure this tx is the one we are looking for
	// 	if bytes.Equal(value.ToMsgEthereumTx(msg.Nonce(), chainId).AsTransaction().Data(), msg.Data()) {
	// 		found = true
	// 		seq = key
	// 		callQueueItem = value
	// 		return true, nil
	// 	}
	// 	return false, nil
	// })

	// if !found {
	// 	h.k.Logger(ctx).Error("Call not found in call queue, reverting tx")
	// 	return errors.New("call not found in call queue, reverting tx")
	// }

	// // send IBC acknowledgement
	// packet := channeltypes.Packet{
	// 	Sequence:           seq,
	// 	SourceChannel:      callQueueItem.InFlightPacket.PacketSrcChannelId,
	// 	SourcePort:         callQueueItem.InFlightPacket.PacketSrcPortId,
	// 	DestinationChannel: callQueueItem.InFlightPacket.RefundChannelId,
	// 	DestinationPort:    callQueueItem.InFlightPacket.RefundPortId,
	// 	Data:               callQueueItem.InFlightPacket.PacketData,
	// 	TimeoutHeight:      clienttypes.MustParseHeight(callQueueItem.InFlightPacket.PacketTimeoutHeight),
	// 	TimeoutTimestamp:   callQueueItem.InFlightPacket.PacketTimeoutTimestamp,
	// }

	// var ack channeltypes.Acknowledgement
	// if receipt.Status == coretypes.ReceiptStatusSuccessful {
	// 	h.k.Logger(ctx).Info("Receipt status successful, creating result acknowledgement")
	// 	ack = channeltypes.NewResultAcknowledgement([]byte{1})
	// } else {
	// 	h.k.Logger(ctx).Info("Receipt status unsuccessful, creating error acknowledgement")
	// 	ack = channeltypes.NewErrorAcknowledgement(errors.New("failed to execute call"))
	// }
	// h.k.Logger(ctx).Info("Created acknowledgment", "ack", ack, "receipt", receipt)

	// h.k.Logger(ctx).Info("Writing IBC acknowledgment...")
	// var data transfertypes.FungibleTokenPacketData
	// err = transfertypes.ModuleCdc.UnmarshalJSON(callQueueItem.InFlightPacket.PacketData, &data)
	// if err != nil {
	// 	h.k.Logger(ctx).Error("failed to unmarshal packet data", "error", err)
	// 	return err
	// }
	// err = h.k.WriteAcknowledgementForPacket(ctx, packet, data, callQueueItem.InFlightPacket, ack)
	// if err != nil {
	// 	h.k.Logger(ctx).Error("failed to write IBC acknowledgment", "error", err)
	// 	return err
	// }
	// h.k.Logger(ctx).Info("Successfully wrote IBC acknowledgment")

	// // remove call from call queue
	// h.k.Logger(ctx).Info("Removing call from queue", "seq", seq)
	// err = h.k.CallQueue.Remove(ctx, seq)
	// if err != nil {
	// 	h.k.Logger(ctx).Error("failed to remove call from queue", "error", err)
	// 	return err
	// }
	// h.k.Logger(ctx).Info("Successfully removed call from queue")

	return nil
}

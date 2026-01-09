package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	callbacktypes "github.com/cosmos/ibc-go/v10/modules/apps/callbacks/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v10/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/precompiles/gateway"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/utils"
)

var _ porttypes.IBCModule = IBCMiddleware{}

type IBCMiddleware struct {
	app                   porttypes.IBCModule
	k                     keeper.Keeper
	packetDataUnmarshaler porttypes.PacketDataUnmarshaler
}

func NewIBCMiddleware(app porttypes.IBCModule, packetDataUnmarshaler porttypes.PacketDataUnmarshaler, k keeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app:                   app,
		k:                     k,
		packetDataUnmarshaler: packetDataUnmarshaler,
	}
}

// OnAcknowledgementPacket implements types.IBCModule.
func (i IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	params, err := i.k.Params.Get(ctx)
	if err != nil {
		i.k.Logger(ctx).Error("failed to get params", "error", err)
		return i.app.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer)
	}
	if !params.Enabled {
		// if the transferrouter module is disabled, we let the underlying module handle the acknowledgement packet
		return i.app.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer)
	}

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		// not a transfer packet, let the underlying module handle the acknowledgement packet
		return i.app.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer)
	}

	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return err
	}

	if !ack.Success() {
		globalPacketSequence, err := i.k.GlobalPacketSequence.Next(ctx)
		if err != nil {
			i.k.Logger(ctx).Error("failed to get next packet sequence", "error", err)
			return err
		}

		// if it's not a success, we must send the tokens back to the sender, either from the escrow address or by minting them
		err = i.k.ErrorOrTimeoutQueue.Set(ctx, globalPacketSequence, types.PacketQueueItem{
			Packet:          &packet,
			OriginalTxHash:  tmhash.Sum(ctx.TxBytes()),
			IsTimeout:       false,
			Acknowledgement: acknowledgement,
		})
		if err != nil {
			i.k.Logger(ctx).Error("failed to set error or timeout queue", "error", err)
			return err
		}
	}

	err = i.addSrcCallbackToQueue(ctx, params, packet, acknowledgement, false)
	if err != nil {
		i.k.Logger(ctx).Error("failed to add src callback to queue on acknowledgement packet", "error", err)
		return err
	}

	return nil
}

// OnTimeoutPacket implements types.IBCModule.
func (i IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	params, err := i.k.Params.Get(ctx)
	if err != nil {
		i.k.Logger(ctx).Error("failed to get params", "error", err)
		return i.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	}
	if !params.Enabled {
		// if the transferrouter module is disabled, we let the underlying module handle the timeout packet
		return i.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	}

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		// not a transfer packet, let the underlying module handle the timeout
		return i.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	}

	globalPacketSequence, err := i.k.GlobalPacketSequence.Next(ctx)
	if err != nil {
		i.k.Logger(ctx).Error("failed to get next packet sequence", "error", err)
		return err
	}

	err = i.k.ErrorOrTimeoutQueue.Set(ctx, globalPacketSequence, types.PacketQueueItem{
		Packet:          &packet,
		OriginalTxHash:  tmhash.Sum(ctx.TxBytes()),
		IsTimeout:       true,
		Acknowledgement: nil,
	})
	if err != nil {
		i.k.Logger(ctx).Error("failed to set error or timeout queue", "error", err)
		return err
	}

	err = i.addSrcCallbackToQueue(ctx, params, packet, nil, true)
	if err != nil {
		i.k.Logger(ctx).Error("failed to add src callback to queue on timeout packet", "error", err)
		return err
	}

	return nil
}

// OnRecvPacket implements types.IBCModule.
func (i IBCMiddleware) OnRecvPacket(ctx sdk.Context, channelVersion string, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	logger := i.k.Logger(ctx)
	params, err := i.k.Params.Get(ctx)
	if err != nil {
		i.k.Logger(ctx).Error("failed to get params in OnRecvPacket", "error", err)
		return i.app.OnRecvPacket(ctx, channelVersion, packet, relayer)
	}

	if !params.Enabled {
		// if the transferrouter module is disabled, we let the underlying module handle the recv packet
		return i.app.OnRecvPacket(ctx, channelVersion, packet, relayer)
	}

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		logger.Debug(fmt.Sprintf("OnRecvPacket payload is not a FungibleTokenPacketData: %s", err.Error()))
		return i.app.OnRecvPacket(ctx, channelVersion, packet, relayer)
	}

	// If it's a PFM packet meant to be forwarded, we return early as we won't handle it here
	d := make(map[string]interface{})
	err = json.Unmarshal([]byte(data.Memo), &d)
	if err == nil && d["forward"] != nil {
		logger.Debug("Packet handled by PFM")
		// a packet meant to be forwarded, let the PFM module handle it
		return i.app.OnRecvPacket(ctx, channelVersion, packet, relayer)
	}

	// Override the receiver address to the gateway contract address
	overrideReceiver := sdk.AccAddress(gateway.PrecompileAddress.Bytes())

	// If it's a callback packet, we perform a check to ensure the receiver address is the expected one,
	// and we set it as the receiver of the funds
	cbData, isCbPacket, err := callbacktypes.GetDestCallbackData(
		ctx, i.packetDataUnmarshaler, packet, params.MaxCallbackGas,
	)

	if isCbPacket {
		// if the packet does opt-in to callbacks but the callback data is malformed,
		// then the packet receive is rejected.
		if err != nil {
			logger.Error("transferrouter OnRecvPacket err", "err", err)
			return channeltypes.NewErrorAcknowledgement(err)
		}

		// if it's a callback packet, we need to receive tokens in the expected address
		receiver, err := sdk.AccAddressFromBech32(data.Receiver)
		if err != nil {
			i.k.Logger(ctx).Error("acc addr from bech32 conversion failed for receiver address", "error", err)
			return i.app.OnRecvPacket(ctx, channelVersion, packet, relayer)
		}
		receiverHex := common.BytesToAddress(receiver.Bytes())

		// Generate secure isolated address from sender.
		isolatedAddr := utils.GenerateIsolatedAddress(packet.GetDestChannel(), data.Sender)
		isolatedAddrHex := common.BytesToAddress(isolatedAddr.Bytes())

		overrideReceiver = isolatedAddr

		// Ensure receiver address is equal to the isolated address.
		if !bytes.Equal(receiverHex.Bytes(), isolatedAddrHex.Bytes()) {
			return newErrorAcknowledgement(fmt.Errorf("expected %s, got %s", isolatedAddrHex.String(), receiverHex.String()))
		}

		if i.k.AccountKeeper.GetAccount(ctx, receiver) == nil {
			acc := i.k.AccountKeeper.NewAccountWithAddress(ctx, receiver)
			i.k.AccountKeeper.SetAccount(ctx, acc)
		}

		contractAddr := common.HexToAddress(cbData.CallbackAddress)

		// Check if the contract address contains code.
		// This check is required because if there is no code, the call will still pass on the EVM side,
		// but it will ignore the calldata and funds may get stuck.
		if !i.k.EVMKeeper.IsContract(ctx, contractAddr) {
			return newErrorAcknowledgement(fmt.Errorf("provided contract address is not a contract: %s", contractAddr))
		}
	}

	// 1. Store the packet in the call queue
	txHash := tmhash.Sum(ctx.TxBytes())
	packetQueueItem := types.PacketQueueItem{
		Packet:         &packet,
		OriginalTxHash: txHash,
	}

	// get a uniquely identifiable packet sequence, to retain order among multiple channels
	globalPacketSequence, err := i.k.GlobalPacketSequence.Next(ctx)
	if err != nil {
		i.k.Logger(ctx).Error("failed to get next packet sequence", "error", err)
		return newErrorAcknowledgement(err)
	}

	err = i.k.PacketQueue.Set(ctx, globalPacketSequence, packetQueueItem)
	if err != nil {
		i.k.Logger(ctx).Error("failed to set packet in call queue", "error", err)
		return newErrorAcknowledgement(err)
	}

	// Move tokens to an escrow account (gateway contract or the isolated address for callback packets)
	err = i.receiveFunds(ctx, channelVersion, packet, data, overrideReceiver.String(), relayer)
	if err != nil {
		i.k.Logger(ctx).Error("failed to receive funds", "error", err)
		return newErrorAcknowledgement(err)
	}

	return nil
}

// OnChanCloseConfirm implements types.IBCModule.
func (i IBCMiddleware) OnChanCloseConfirm(ctx sdk.Context, portID string, channelID string) error {
	return i.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements types.IBCModule.
func (i IBCMiddleware) OnChanCloseInit(ctx sdk.Context, portID string, channelID string) error {
	return i.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanOpenAck implements types.IBCModule.
func (i IBCMiddleware) OnChanOpenAck(ctx sdk.Context, portID string, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
	return i.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements types.IBCModule.
func (i IBCMiddleware) OnChanOpenConfirm(ctx sdk.Context, portID string, channelID string) error {
	return i.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanOpenInit implements types.IBCModule.
func (i IBCMiddleware) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, counterparty channeltypes.Counterparty, version string) (string, error) {
	return i.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, counterparty, version)
}

// OnChanOpenTry implements types.IBCModule.
func (i IBCMiddleware) OnChanOpenTry(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, counterparty channeltypes.Counterparty, counterpartyVersion string) (version string, err error) {
	return i.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, counterparty, counterpartyVersion)
}

// helper functions

func (i IBCMiddleware) addSrcCallbackToQueue(ctx sdk.Context, params types.Params, packet channeltypes.Packet, acknowledgement []byte, isTimeout bool) error {

	// get callback data
	_, isCbPacket, err := callbacktypes.GetSourceCallbackData(ctx, i.packetDataUnmarshaler, packet, params.MaxCallbackGas)
	if isCbPacket {
		if err != nil {
			i.k.Logger(ctx).Error("failed to get callback data", "error", err)
			return err
		}

		// get a uniquely identifiable packet sequence, to retain order among multiple channels
		globalPacketSequence, err := i.k.GlobalPacketSequence.Next(ctx)
		if err != nil {
			i.k.Logger(ctx).Error("failed to get next packet sequence", "error", err)
			return err
		}

		// add the callback data to the callback queue
		err = i.k.SrcCallbackQueue.Set(ctx, globalPacketSequence, types.PacketQueueItem{
			Packet:          &packet,
			OriginalTxHash:  tmhash.Sum(ctx.TxBytes()),
			IsTimeout:       isTimeout,
			Acknowledgement: acknowledgement,
		})
		if err != nil {
			i.k.Logger(ctx).Error("failed to set callback queue", "error", err)
		}
		return err
	}
	return nil
}

// receiveFunds receives funds from the packet into the override receiver
// address and returns an error if the funds cannot be received. (from PFM, thank you!)
func (i IBCMiddleware) receiveFunds(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
	overrideReceiver string,
	relayer sdk.AccAddress,
) error {
	overrideData := transfertypes.FungibleTokenPacketData{
		Denom:    data.Denom,
		Amount:   data.Amount,
		Sender:   data.Sender,
		Receiver: overrideReceiver, // override receiver
		// Memo explicitly zeroed
	}
	overrideDataBz := transfertypes.ModuleCdc.MustMarshalJSON(&overrideData)
	overridePacket := channeltypes.Packet{
		Sequence:           packet.Sequence,
		SourcePort:         packet.SourcePort,
		SourceChannel:      packet.SourceChannel,
		DestinationPort:    packet.DestinationPort,
		DestinationChannel: packet.DestinationChannel,
		Data:               overrideDataBz, // override data
		TimeoutHeight:      packet.TimeoutHeight,
		TimeoutTimestamp:   packet.TimeoutTimestamp,
	}

	ack := i.app.OnRecvPacket(ctx, channelVersion, overridePacket, relayer)

	if ack == nil {
		return fmt.Errorf("ack is nil")
	}

	if !ack.Success() {
		return fmt.Errorf("ack error: %s", string(ack.Acknowledgement()))
	}

	return nil
}

// newErrorAcknowledgement returns an error that identifies PFM and provides the error.
// It's okay if these errors are non-deterministic, because they will not be committed to state, only emitted as events.
func newErrorAcknowledgement(err error) channeltypes.Acknowledgement {
	return channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Error{
			Error: fmt.Sprintf("transfer-router error: %s", err.Error()),
		},
	}
}

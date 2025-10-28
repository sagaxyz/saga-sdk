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
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/utils"
)

var _ porttypes.IBCModule = IBCMiddleware{}

type IBCMiddleware struct {
	app                   porttypes.IBCModule
	k                     keeper.Keeper
	maxCallbackGas        uint64
	packetDataUnmarshaler porttypes.PacketDataUnmarshaler
}

func NewIBCMiddleware(app porttypes.IBCModule, packetDataUnmarshaler porttypes.PacketDataUnmarshaler, maxCallbackGas uint64, k keeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app:                   app,
		k:                     k,
		packetDataUnmarshaler: packetDataUnmarshaler,
		maxCallbackGas:        maxCallbackGas,
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
	err := i.addSrcCallbackToQueue(ctx, packet, acknowledgement, false)
	if err != nil {
		i.k.Logger(ctx).Error("failed to add src callback to queue on acknowledgement packet", "error", err)
	}

	return i.app.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements types.IBCModule.
func (i IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	err := i.addSrcCallbackToQueue(ctx, packet, nil, true)
	if err != nil {
		i.k.Logger(ctx).Error("failed to add src callback to queue on timeout packet", "error", err)
	}
	return i.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
}

// OnRecvPacket implements types.IBCModule.
func (i IBCMiddleware) OnRecvPacket(ctx sdk.Context, channelVersion string, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	logger := i.k.Logger(ctx)

	logger.Info("transferrouter OnRecvPacket", "packet", packet)

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		logger.Debug(fmt.Sprintf("OnRecvPacket payload is not a FungibleTokenPacketData: %s", err.Error()))
		return i.app.OnRecvPacket(ctx, channelVersion, packet, relayer)
	}

	logger.Info("transferrouter OnRecvPacket data", "data", data)
	// If it's a PFM packet meant to be forwarded, we return early as we won't handle it here
	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(data.Memo), &d)
	if err == nil && d["forward"] != nil {
		logger.Debug("Packet handled by PFM")
		// a packet meant to be forwarded, let the PFM module handle it
		return i.app.OnRecvPacket(ctx, channelVersion, packet, relayer)
	}

	params, err := i.k.Params.Get(ctx)
	if err != nil {
		i.k.Logger(ctx).Error("failed to get params", "error", err)
		return newErrorAcknowledgement(err)
	}

	logger.Info("transferrouter OnRecvPacket params", "params", params)

	// Override the receiver address to the gateway contract address
	gatewayAddr := common.HexToAddress(params.GatewayContractAddress)
	logger.Info("transferrouter OnRecvPacket gatewayAddr", "gatewayAddr", gatewayAddr)
	overrideReceiver := sdk.AccAddress(gatewayAddr.Bytes())

	// If it's a callback packet, we perform a check to ensure the receiver address is the expected one,
	// and we set it as the receiver of the funds
	cbData, isCbPacket, err := callbacktypes.GetDestCallbackData(
		ctx, i.packetDataUnmarshaler, packet, i.maxCallbackGas,
	)
	logger.Info("transferrouter OnRecvPacket cbData", "cbData", cbData)
	logger.Info("transferrouter OnRecvPacket isCbPacket", "isCbPacket", isCbPacket)

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
		contractAccount := i.k.EVMKeeper.GetAccountOrEmpty(ctx, contractAddr)

		// Check if the contract address contains code.
		// This check is required because if there is no code, the call will still pass on the EVM side,
		// but it will ignore the calldata and funds may get stuck.
		if !contractAccount.IsContract() {
			return newErrorAcknowledgement(fmt.Errorf("provided contract address is not a contract: %s", contractAddr))
		}
	}

	logger.Info("transferrouter OnRecvPacket before store packet in call queue")

	// 1. Store the packet in the call queue
	txHash := tmhash.Sum(ctx.TxBytes())
	packetQueueItem := types.PacketQueueItem{
		Packet:         &packet,
		OriginalTxHash: txHash,
	}
	err = i.k.PacketQueue.Set(ctx, packet.Sequence, packetQueueItem)
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

	// print current balance of the override receiver
	balances := i.k.BankKeeper.GetAllBalances(ctx, overrideReceiver)

	i.k.Logger(ctx).Info("transferrouter OnRecvPacket balance", "overrideReceiver", overrideReceiver, "balance", balances)

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

func (i IBCMiddleware) addSrcCallbackToQueue(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, isTimeout bool) error {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		i.k.Logger(ctx).Error("transferrouter error parsing packet data from ack packet",
			"sequence", packet.Sequence,
			"src-channel", packet.SourceChannel, "src-port", packet.SourcePort,
			"dst-channel", packet.DestinationChannel, "dst-port", packet.DestinationPort,
			"error", err,
		)

		// do not return an error, just log it
		return nil
	}

	// get callback data
	_, isCbPacket, err := callbacktypes.GetSourceCallbackData(ctx, i.packetDataUnmarshaler, packet, i.maxCallbackGas)
	if isCbPacket {
		if err != nil {
			i.k.Logger(ctx).Error("failed to get callback data", "error", err)
		}

		// add the callback data to the callback queue
		err = i.k.SrcCallbackQueue.Set(ctx, packet.Sequence, types.PacketQueueItem{
			Packet:          &packet,
			OriginalTxHash:  tmhash.Sum(ctx.TxBytes()),
			IsTimeout:       isTimeout,
			Acknowledgement: acknowledgement,
		})
		if err != nil {
			i.k.Logger(ctx).Error("failed to set callback queue", "error", err)
		}
		return nil
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
	i.k.Logger(ctx).Info("transferrouter receiveFunds overrideDataBz", "overrideDataBz", string(overrideDataBz))
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

	i.k.Logger(ctx).Info("transferrouter OnRecvPacket overridePacket", "overrideReceiver", overrideReceiver)

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

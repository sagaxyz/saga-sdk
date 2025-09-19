package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	callbacktypes "github.com/sagaxyz/saga-sdk/x/transferrouter/v10types"
)

var _ porttypes.IBCModule = IBCMiddleware{}

type IBCMiddleware struct {
	app porttypes.IBCModule
	k   keeper.Keeper
}

func NewIBCMiddleware(app porttypes.IBCModule, k keeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app: app,
		k:   k,
	}
}

// OnAcknowledgementPacket implements types.IBCModule.
func (i IBCMiddleware) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		i.k.Logger(ctx).Error("transferrouter error parsing packet data from ack packet",
			"sequence", packet.Sequence,
			"src-channel", packet.SourceChannel, "src-port", packet.SourcePort,
			"dst-channel", packet.DestinationChannel, "dst-port", packet.DestinationPort,
			"error", err,
		)
		return i.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	ack := channeltypes.Acknowledgement{}
	err := json.Unmarshal(acknowledgement, &ack)
	if err != nil {
		return err
	}

	if ack.Success() {
		return i.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	// if the acknowledgement is an error, we need to refund the tokens to the sender
	// TODO: implement refund by adding a call to the call queue
	// callData, err := CreateGatewayERC20TransferCallData(
	// 	ctx, i.k, data.Denom, data.Amount, data.Sender, nil,
	// )
	// if err != nil {
	// 	i.k.Logger(ctx).Error("failed to create gateway execute call data", "error", err)
	// 	return err
	// }

	// params, err := i.k.Params.Get(ctx)
	// if err != nil {
	// 	i.k.Logger(ctx).Error("failed to get params", "error", err)
	// 	return err
	// }

	// // Parse the configured private key (in hex format) and derive the corresponding
	// // Ethereum address of the known signer.
	// knownSignerAddress, err := sdk.AccAddressFromBech32(params.KnownSignerAddress)
	// if err != nil {
	// 	i.k.Logger(ctx).Error("failed to parse known signer private key", "error", err)
	// 	return err
	// }
	// gatewayAddr := common.HexToAddress("0x5A6A8Ce46E34c2cd998129d013fA0253d3892345")

	// err = i.k.CallQueue.Set(ctx, packet.Sequence, types.CallQueueItem{
	// 	Call: &types.Call{
	// 		From:     knownSignerAddress.Bytes(),
	// 		Contract: gatewayAddr.Bytes(),
	// 		Data:     callData,
	// 		Commit:   true,
	// 	},
	// })
	// if err != nil {
	// 	i.k.Logger(ctx).Error("failed to set call queue", "error", err)
	// 	return err
	// }

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
func (i IBCMiddleware) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string) (string, error) {
	return i.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, version)
}

// OnChanOpenTry implements types.IBCModule.
func (i IBCMiddleware) OnChanOpenTry(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string) (version string, err error) {
	return i.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, counterpartyVersion)
}

// OnRecvPacket implements types.IBCModule.
func (i IBCMiddleware) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	logger := i.k.Logger(ctx)

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		logger.Debug(fmt.Sprintf("OnRecvPacket payload is not a FungibleTokenPacketData: %s", err.Error()))
		return i.app.OnRecvPacket(ctx, packet, relayer)
	}

	logger.Info("OnRecvPacket called with memo!!!asdasdasdasdasasd21345678976543======++++++======", "memo", data.Memo)

	// If it's a PFM packet meant to be forwarded, we return early as we won't handle it here
	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(data.Memo), &d)
	if err == nil && d["forward"] != nil {
		logger.Info("Packet handled by PFM")
		// a packet meant to be forwarded, let the PFM module handle it
		return i.app.OnRecvPacket(ctx, packet, relayer)
	}

	// Override the receiver address to the gateway contract address
	gatewayAddr := common.HexToAddress("0x5A6A8Ce46E34c2cd998129d013fA0253d3892345") // TODO: make this configurable
	overrideReceiver := sdk.AccAddress(gatewayAddr.Bytes())

	// Generate secure isolated address from sender.
	isolatedAddr := GenerateIsolatedAddress(packet.GetDestChannel(), data.Sender)
	isolatedAddrHex := common.BytesToAddress(isolatedAddr.Bytes())

	logger.Info("OnRecvPacket called with isolatedAddrHex", "isolatedAddrHex", isolatedAddrHex)

	// If it's a callback packet, we perform a check to ensure the receiver address is the expected one,
	// and we set it as the receiver of the funds
	cbData, isCbPacket, err := callbacktypes.GetCallbackData(data, callbacktypes.V1, packet.GetDestPort(), ctx.GasMeter().GasRemaining(), ctx.GasMeter().GasRemaining(), callbacktypes.DestinationCallbackKey)
	logger.Info("OnRecvPacket called with cbData", "cbData", cbData, "isCbPacket", isCbPacket, "err", err)

	if isCbPacket {
		if err != nil {
			// if isCbPacket is true and the error != nil, we have a malformed packet
			i.k.Logger(ctx).Error("failed to get callback data", "error", err)
			return newErrorAcknowledgement(err)
		}
		// if it's a callback packet, we need to receive tokens in the expected address
		receiver, err := sdk.AccAddressFromBech32(data.Receiver)
		if err != nil {
			i.k.Logger(ctx).Error("acc addr from bech32 conversion failed for receiver address", "error", err)
			return i.app.OnRecvPacket(ctx, packet, relayer)
		}
		receiverHex := common.BytesToAddress(receiver.Bytes())

		// Generate secure isolated address from sender.
		isolatedAddr := GenerateIsolatedAddress(packet.GetDestChannel(), data.Sender)
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

	// Move tokens to an escrow account (gateway contract or the isolated address for callback packets)
	err = i.receiveFunds(ctx, packet, data, overrideReceiver.String(), relayer)
	if err != nil {
		i.k.Logger(ctx).Error("failed to receive funds", "error", err)
		return newErrorAcknowledgement(err)
	}

	// params, err := i.k.Params.Get(ctx)
	// if err != nil {
	// 	i.k.Logger(ctx).Error("failed to get params", "error", err)
	// 	return newErrorAcknowledgement(err)
	// }

	// assemble the call data, erc20 transfer for now
	// callData, err := CreateGatewayERC20TransferExecuteCallDataFromPacket(ctx, i.k, packet, data)
	// if err != nil {
	// 	i.k.Logger(ctx).Error("failed to create gateway execute call data", "error", err)
	// 	return newErrorAcknowledgement(err)
	// }

	// knownSignerAddress, err := sdk.AccAddressFromBech32(params.KnownSignerAddress)
	// if err != nil {
	// 	i.k.Logger(ctx).Error("failed to parse known signer address", "error", err)
	// 	return newErrorAcknowledgement(err)
	// }

	// 1. Store the packet in the call queue
	i.k.PacketQueue.Set(ctx, packet.Sequence, packet)

	// i.k.CallQueue.Set(ctx, packet.Sequence, types.CallQueueItem{
	// 	Call: &types.Call{
	// 		From:     knownSignerAddress.Bytes(),
	// 		Contract: gatewayAddr.Bytes(),
	// 		Data:     callData,
	// 		Commit:   true,
	// 	},
	// 	InFlightPacket: &types.InFlightPacket{
	// 		OriginalSenderAddress:  data.Sender,
	// 		RefundChannelId:        packet.SourceChannel,
	// 		RefundPortId:           packet.SourcePort,
	// 		PacketSrcChannelId:     packet.SourceChannel,
	// 		PacketSrcPortId:        packet.SourcePort,
	// 		PacketTimeoutTimestamp: packet.TimeoutTimestamp,
	// 		PacketTimeoutHeight:    packet.TimeoutHeight.String(),
	// 		PacketData:             packet.Data,
	// 		RefundSequence:         packet.Sequence,
	// 		RetriesRemaining:       0,
	// 		Timeout:                0,
	// 		Nonrefundable:          false,
	// 	},
	// })

	// Do not return the acknowledgement, we will write it in the post handler
	return nil
}

// OnTimeoutPacket implements types.IBCModule.
func (i IBCMiddleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	return i.app.OnTimeoutPacket(ctx, packet, relayer)
}

// receiveFunds receives funds from the packet into the override receiver
// address and returns an error if the funds cannot be received. (from PFM, thank you!)
func (i IBCMiddleware) receiveFunds(
	ctx sdk.Context,
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

	ack := i.app.OnRecvPacket(ctx, overridePacket, relayer)

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

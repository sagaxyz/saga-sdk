package middlewares

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/abi"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
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

	// If it's a PFM packet meant to be forwarded, we return early as we won't handle it here
	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(data.Memo), &d)
	if err == nil && d["forward"] != nil {
		// a packet meant to be forwarded, let the PFM module handle it
		return i.app.OnRecvPacket(ctx, packet, relayer)
	}

	// Move tokens to an escrow account by replacing the destination address in the packet data
	params, err := i.k.Params.Get(ctx)
	if err != nil {
		i.k.Logger(ctx).Error("failed to get params", "error", err)
		return newErrorAcknowledgement(err)
	}

	// Override the receiver address to the gateway contract address
	gatewayAddr := common.HexToAddress("0x0000000000000000000000000000000000006a7e") // TODO: make this configurable
	gatewayCosmosAddr := sdk.AccAddress(gatewayAddr.Bytes())

	err = i.receiveFunds(ctx, packet, data, gatewayCosmosAddr.String(), relayer)
	if err != nil {
		i.k.Logger(ctx).Error("failed to receive funds", "error", err)
		return newErrorAcknowledgement(err)
	}

	// TODO: now only a simple transfer is supported, we need to add support for other stuff?

	// assemble the call data, erc20 transfer for now

	// TODO: remember to handle denoms differently if this chain was the sender
	// see ReceiverChainIsSource in transfer keeper relay.go
	// since SendPacket did not prefix the denomination, we must prefix denomination here
	sourcePrefix := transfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
	// NOTE: sourcePrefix contains the trailing "/"
	prefixedDenom := sourcePrefix + data.Denom
	denomTrace := transfertypes.ParseDenomTrace(prefixedDenom)

	receiverAccAddr, err := sdk.AccAddressFromBech32(data.Receiver)
	if err != nil {
		i.k.Logger(ctx).Error("failed to parse receiver address", "error", err)
		return newErrorAcknowledgement(err)
	}
	recipientAddrHex := common.BytesToAddress(receiverAccAddr.Bytes())

	// get the coin address
	coinAddr, err := i.k.Erc20Keeper.GetCoinAddress(ctx, denomTrace.IBCDenom())
	if err != nil {
		i.k.Logger(ctx).Error("failed to get coin address", "error", err)
		return newErrorAcknowledgement(err)
	}

	amount, ok := new(big.Int).SetString(data.Amount, 10)
	if !ok {
		i.k.Logger(ctx).Error("failed to parse amount", "error", err)
		return newErrorAcknowledgement(err)
	}

	// transfer(address recipient, uint256 amount) → bool
	callData, err := abi.ERC20ABI.Pack("transfer", recipientAddrHex, amount)
	if err != nil {
		i.k.Logger(ctx).Error("failed to pack call data", "error", err)
		return newErrorAcknowledgement(err)
	}

	txHash := tmhash.Sum(ctx.TxBytes())
	txHashHex := hex.EncodeToString(txHash)
	memo, err := json.Marshal(map[string]interface{}{
		"txHash": txHashHex,
	})
	if err != nil {
		i.k.Logger(ctx).Error("failed to marshal memo", "error", err)
		return newErrorAcknowledgement(err)
	}

	// Now assemble the call data for the gateway
	// function execute(address target,uint256 value, bytes calldata data, bytes calldata note)
	// note should contain data on the original packet, for now we just use the packet data
	callData, err = abi.GatewayABI.Pack("execute", coinAddr, big.NewInt(0), callData, memo)
	if err != nil {
		i.k.Logger(ctx).Error("failed to pack call data", "error", err)
		return newErrorAcknowledgement(err)
	}

	// Parse the configured private key (in hex format) and derive the corresponding
	// Ethereum address of the known signer.
	privKey, err := crypto.HexToECDSA(params.KnownSignerPrivateKey)
	if err != nil {
		i.k.Logger(ctx).Error("failed to parse known signer private key", "error", err)
		return newErrorAcknowledgement(err)
	}
	knownSignerAddress := sdk.AccAddress(crypto.PubkeyToAddress(privKey.PublicKey).Bytes())

	// 1. Store the packet in the call queue
	i.k.CallQueue.Set(ctx, packet.Sequence, types.CallQueueItem{
		Call: &types.Call{
			From:     knownSignerAddress.Bytes(),
			Contract: gatewayAddr.Bytes(),
			Data:     callData,
			Commit:   true,
		},
		InFlightPacket: &types.InFlightPacket{
			OriginalSenderAddress:  data.Sender,
			RefundChannelId:        packet.SourceChannel,
			RefundPortId:           packet.SourcePort,
			PacketSrcChannelId:     packet.SourceChannel,
			PacketSrcPortId:        packet.SourcePort,
			PacketTimeoutTimestamp: packet.TimeoutTimestamp,
			PacketTimeoutHeight:    packet.TimeoutHeight.String(),
			PacketData:             packet.Data,
			RefundSequence:         packet.Sequence,
			RetriesRemaining:       0,
			Timeout:                0,
			Nonrefundable:          false,
		},
	})

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

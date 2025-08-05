package middlewares

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/ethereum/go-ethereum/crypto"
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
	return i.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
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

	// If it's a PFM packet meant to be forwarded, we return early
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
		return i.app.OnRecvPacket(ctx, packet, relayer)
	}

	// Parse the configured private key (in hex format) and derive the corresponding
	// Ethereum address of the known signer.
	privKey, err := crypto.HexToECDSA(params.KnownSignerPrivateKey)
	if err != nil {
		i.k.Logger(ctx).Error("failed to parse known signer private key", "error", err)
		return i.app.OnRecvPacket(ctx, packet, relayer)
	}

	knownSignerAddress := crypto.PubkeyToAddress(privKey.PublicKey)
	newRecAddr := sdk.AccAddress(knownSignerAddress.Bytes())
	data.Receiver = newRecAddr.String()

	// update the packet data
	packet.Data, err = json.Marshal(data)
	if err != nil {
		i.k.Logger(ctx).Error("failed to marshal packet data", "error", err)
		return i.app.OnRecvPacket(ctx, packet, relayer)
	}

	// 1. Store the packet in the call queue
	i.k.CallQueue.Set(ctx, packet.Sequence, types.CallQueueItem{
		Call: &types.Call{
			Data: packet.Data, // TODO: this is not right, we need to parse it and make the call
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

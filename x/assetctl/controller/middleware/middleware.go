package middleware

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/types"
	pfmtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/keeper"
)

type Middleware struct {
	logger log.Logger
	app    porttypes.IBCModule
	k      keeper.Keeper
}

// OnAcknowledgementPacket implements types.IBCModule.
func (m *Middleware) OnAcknowledgementPacket(ctx types.Context, packet channeltypes.Packet, acknowledgement []byte, relayer types.AccAddress) error {
	return m.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnChanCloseConfirm implements types.IBCModule.
func (m *Middleware) OnChanCloseConfirm(ctx types.Context, portID string, channelID string) error {
	return m.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements types.IBCModule.
func (m *Middleware) OnChanCloseInit(ctx types.Context, portID string, channelID string) error {
	return m.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanOpenAck implements types.IBCModule.
func (m *Middleware) OnChanOpenAck(ctx types.Context, portID string, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
	return m.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements types.IBCModule.
func (m *Middleware) OnChanOpenConfirm(ctx types.Context, portID string, channelID string) error {
	return m.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanOpenInit implements types.IBCModule.
func (m *Middleware) OnChanOpenInit(ctx types.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string) (string, error) {
	return m.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, version)
}

// OnChanOpenTry implements types.IBCModule.
func (m *Middleware) OnChanOpenTry(ctx types.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string) (version string, err error) {
	return m.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, counterpartyVersion)
}

// OnRecvPacket implements types.IBCModule.
func (m *Middleware) OnRecvPacket(ctx types.Context, packet channeltypes.Packet, relayer types.AccAddress) exported.Acknowledgement {
	// This is the only method we need to implement, the rest are just passthrough.
	var data transfertypes.FungibleTokenPacketData
	// TODO: use a cdc that we pass in the constructor
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return m.app.OnRecvPacket(ctx, packet, relayer)
	}

	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(data.Memo), &d)
	if err != nil || d["forward"] == nil {
		// not a packet that should be forwarded
		return m.app.OnRecvPacket(ctx, packet, relayer)
	}

	pfmmetadata := &pfmtypes.PacketMetadata{}
	err = json.Unmarshal([]byte(data.Memo), pfmmetadata)
	if err != nil {
		m.logger.Error("packetForwardMiddleware OnRecvPacket error parsing forward metadata", "error", err)
		return newErrorAcknowledgement(fmt.Sprintf("error parsing forward metadata: %s", err.Error()))
	}

	has, err := m.k.SupportedAssets.Has(ctx, collections.Join(pfmmetadata.Forward.Channel, data.Denom))
	if err != nil {
		return newErrorAcknowledgement("error checking if asset is supported")
	}
	if !has {
		return newErrorAcknowledgement("asset not supported by destination chain")
	}

	return m.app.OnRecvPacket(ctx, packet, relayer)
}

// OnTimeoutPacket implements types.IBCModule.
func (m *Middleware) OnTimeoutPacket(ctx types.Context, packet channeltypes.Packet, relayer types.AccAddress) error {
	return m.app.OnTimeoutPacket(ctx, packet, relayer)
}

var _ porttypes.IBCModule = &Middleware{}

// newErrorAcknowledgement returns an error that identifies PFM and provides the error.
// It's okay if these errors are non-deterministic, because they will not be committed to state, only emitted as events.
func newErrorAcknowledgement(err string) channeltypes.Acknowledgement {
	return channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Error{
			Error: fmt.Sprintf("assetctl error: %s", err),
		},
	}
}

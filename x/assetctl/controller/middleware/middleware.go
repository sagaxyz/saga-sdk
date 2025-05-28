package middleware

import (
	"github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/keeper"
)

type middlewareInterface interface {
	porttypes.IBCModule
	porttypes.ICS4Wrapper
}

var _ middlewareInterface = &Middleware{}

type Middleware struct {
	k keeper.Keeper
}

// GetAppVersion implements middlewareInterface.
func (m *Middleware) GetAppVersion(ctx types.Context, portID string, channelID string) (string, bool) {
	panic("unimplemented")
}

// OnAcknowledgementPacket implements middlewareInterface.
func (m *Middleware) OnAcknowledgementPacket(ctx types.Context, packet channeltypes.Packet, acknowledgement []byte, relayer types.AccAddress) error {
	panic("unimplemented")
}

// OnChanCloseConfirm implements middlewareInterface.
func (m *Middleware) OnChanCloseConfirm(ctx types.Context, portID string, channelID string) error {
	panic("unimplemented")
}

// OnChanCloseInit implements middlewareInterface.
func (m *Middleware) OnChanCloseInit(ctx types.Context, portID string, channelID string) error {
	panic("unimplemented")
}

// OnChanOpenAck implements middlewareInterface.
func (m *Middleware) OnChanOpenAck(ctx types.Context, portID string, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
	panic("unimplemented")
}

// OnChanOpenConfirm implements middlewareInterface.
func (m *Middleware) OnChanOpenConfirm(ctx types.Context, portID string, channelID string) error {
	panic("unimplemented")
}

// OnChanOpenInit implements middlewareInterface.
func (m *Middleware) OnChanOpenInit(ctx types.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string) (string, error) {
	panic("unimplemented")
}

// OnChanOpenTry implements middlewareInterface.
func (m *Middleware) OnChanOpenTry(ctx types.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string) (version string, err error) {
	panic("unimplemented")
}

// OnRecvPacket implements middlewareInterface.
func (m *Middleware) OnRecvPacket(ctx types.Context, packet channeltypes.Packet, relayer types.AccAddress) exported.Acknowledgement {
	panic("unimplemented")
}

// OnTimeoutPacket implements middlewareInterface.
func (m *Middleware) OnTimeoutPacket(ctx types.Context, packet channeltypes.Packet, relayer types.AccAddress) error {
	panic("unimplemented")
}

// SendPacket implements middlewareInterface.
func (m *Middleware) SendPacket(ctx types.Context, chanCap *capabilitytypes.Capability, sourcePort string, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, data []byte) (sequence uint64, err error) {
	panic("unimplemented")
}

// WriteAcknowledgement implements middlewareInterface.
func (m *Middleware) WriteAcknowledgement(ctx types.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	panic("unimplemented")
}

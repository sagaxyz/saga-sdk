package middleware

import (
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	controllertypes "github.com/sagaxyz/saga-sdk/x/assetctl/controller/types"
	"github.com/sagaxyz/saga-sdk/x/assetctl/host/keeper"
)

type IBCMiddleware struct {
	logger log.Logger
	app    porttypes.IBCModule
	k      keeper.Keeper
}

func NewIBCMiddleware(app porttypes.IBCModule, k keeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app: app,
		k:   k,
	}
}

// OnAcknowledgementPacket implements types.IBCModule.
func (m *IBCMiddleware) OnAcknowledgementPacket(ctx types.Context, packet channeltypes.Packet, acknowledgement []byte, relayer types.AccAddress) error {
	msgType, err := m.k.InFlightRequests.Get(ctx, packet.Sequence)
	if err != nil {
		// we don't have a request for this sequence, so we skip it with no errors as it
		// must have been launched by a different module
		return m.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	// mark the request as completed
	err = m.k.InFlightRequests.Remove(ctx, packet.Sequence)
	if err != nil {
		return err
	}

	ack := channeltypes.Acknowledgement{}
	err = ack.Unmarshal(acknowledgement)
	if err != nil {
		return err
	}

	if ack.Response == nil {
		return errors.ErrInvalidRequest.Wrap("acknowledgement is nil")
	}

	if ack.GetError() != "" {
		return errors.ErrInvalidRequest.Wrapf("acknowledgement error: %s", ack.GetError())
	}

	if ack.GetResult() == nil {
		return errors.ErrInvalidRequest.Wrap("acknowledgement result is nil")
	}

	// check if the packet request is an icatypes.InterchainAccountPacketData,
	// here we can error as the packet sequence was in our store, so it must be an
	// interchain account packet
	var packetData icatypes.InterchainAccountPacketData
	if err := packetData.Unmarshal(packet.Data); err != nil {
		return err
	}

	// now we check what kind of request it was, could be controllertypes.MsgManageRegisteredAssets or
	// controllertypes.MsgManageSupportedAssets
	switch msgType {
	case sdk.MsgTypeURL(&controllertypes.MsgManageRegisteredAssets{}):
		msgRegAssets := &controllertypes.MsgManageRegisteredAssets{}
		err = msgRegAssets.Unmarshal(packetData.Data)
		if err != nil {
			return err
		}

		response := &controllertypes.MsgManageRegisteredAssetsResponse{}
		err = response.Unmarshal(ack.GetResult())
		if err != nil {
			return err
		}

		// nothing to do here, we just need to ack the packet
	case sdk.MsgTypeURL(&controllertypes.MsgManageSupportedAssets{}):
		msgSuppAssets := &controllertypes.MsgManageSupportedAssets{}
		err = msgSuppAssets.Unmarshal(packetData.Data)
		if err != nil {
			return err
		}

		response := &controllertypes.MsgManageSupportedAssetsResponse{}
		err = response.Unmarshal(ack.GetResult())
		if err != nil {
			return err
		}

		// now we need to register the assets
		for _, asset := range response.AddedAssets {
			// we need to register the asset
			tokenPair, err := m.k.Erc20Keeper.RegisterERC20Extension(ctx, asset.Base)
			if err != nil {
				return err
			}

			// base denomination, change the base denom to the erc20 denom
			erc20Denom := erc20types.CreateDenom(tokenPair.GetERC20Contract().String())
			asset.Base = erc20Denom
			for i := range asset.DenomUnits {
				if asset.DenomUnits[i].Exponent == 0 {
					asset.DenomUnits[i].Denom = erc20Denom
					break
				}
			}

			m.k.BankKeeper.SetDenomMetaData(ctx, asset)
		}
	default:
		return errors.ErrInvalidRequest.Wrapf("unknown message type: %s", msgType)
	}

	return m.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnChanCloseConfirm implements types.IBCModule.
func (m *IBCMiddleware) OnChanCloseConfirm(ctx types.Context, portID string, channelID string) error {
	return m.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements types.IBCModule.
func (m *IBCMiddleware) OnChanCloseInit(ctx types.Context, portID string, channelID string) error {
	return m.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanOpenAck implements types.IBCModule.
func (m *IBCMiddleware) OnChanOpenAck(ctx types.Context, portID string, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
	return m.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements types.IBCModule.
func (m *IBCMiddleware) OnChanOpenConfirm(ctx types.Context, portID string, channelID string) error {
	return m.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanOpenInit implements types.IBCModule.
func (m *IBCMiddleware) OnChanOpenInit(ctx types.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string) (string, error) {
	return m.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, version)
}

// OnChanOpenTry implements types.IBCModule.
func (m *IBCMiddleware) OnChanOpenTry(ctx types.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string) (version string, err error) {
	return m.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, counterpartyVersion)
}

// OnRecvPacket implements types.IBCModule.
func (m *IBCMiddleware) OnRecvPacket(ctx types.Context, packet channeltypes.Packet, relayer types.AccAddress) exported.Acknowledgement {
	return m.app.OnRecvPacket(ctx, packet, relayer)
}

// OnTimeoutPacket implements types.IBCModule.
func (m *IBCMiddleware) OnTimeoutPacket(ctx types.Context, packet channeltypes.Packet, relayer types.AccAddress) error {
	return m.app.OnTimeoutPacket(ctx, packet, relayer)
}

var _ porttypes.IBCModule = &IBCMiddleware{}

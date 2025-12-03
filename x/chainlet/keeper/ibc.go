package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"

	"github.com/sagaxyz/saga-sdk/x/chainlet/types"
)

// Verifies channel matches the client ID of the provider
func (k Keeper) verifyChannel(ctx sdk.Context, clientID string, channelID string) error {
	channel, found := k.channelKeeper.GetChannel(ctx, types.PortID, channelID)
	if !found {
		return fmt.Errorf("channel %s not found (port: %s)", channelID, types.PortID)
	}
	if len(channel.ConnectionHops) == 0 {
		return fmt.Errorf("no connection hops for channel %s", channelID)
	}
	connectionID := channel.ConnectionHops[0]
	connection, found := k.connectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return fmt.Errorf("connection not found: %s", connectionID)
	}
	if connection.ClientId != clientID {
		return fmt.Errorf("client ID of the provided channel does not match provider client id (%s != %s)", connection.ClientId, clientID)
	}

	return nil
}

// TransmitConfirmUpgradePacket transmits the packet over IBC with the specified source port and source channel
func (k Keeper) TransmitConfirmUpgradePacket(
	ctx sdk.Context,
	packetData types.ConfirmUpgradePacketData,
	sourcePort,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) (uint64, error) {
	packetBytes, err := packetData.GetBytes()
	if err != nil {
		return 0, errorsmod.Wrapf(sdkerrors.ErrJSONMarshal, "cannot marshal the packet: %s", err)
	}

	return k.channelKeeper.SendPacket(ctx, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, packetBytes)
}

// OnAcknowledgementConfirmUpgradePacket responds to the success or failure of a packet
// acknowledgement written on the receiving chain.
func (k Keeper) OnAcknowledgementConfirmUpgradePacket(ctx sdk.Context, packet channeltypes.Packet, data types.ConfirmUpgradePacketData, ack channeltypes.Acknowledgement) error {
	switch dispatchedAck := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:
		return nil
	case *channeltypes.Acknowledgement_Result:
		// Decode the packet acknowledgment
		var packetAck types.ConfirmUpgradePacketAck

		if err := types.ModuleCdc.UnmarshalJSON(dispatchedAck.Result, &packetAck); err != nil {
			// The counter-party module doesn't implement the correct acknowledgment format
			return errors.New("cannot unmarshal acknowledgment")
		}
		return nil
	default:
		// The counter-party module doesn't implement the correct acknowledgment format
		return errors.New("invalid acknowledgment format")
	}
}

// OnTimeoutConfirmUpgradePacket responds to the case where a packet has not been transmitted because of a timeout
func (k Keeper) OnTimeoutConfirmUpgradePacket(ctx sdk.Context, packet channeltypes.Packet, data types.ConfirmUpgradePacketData) error {
	return nil
}

// OnRecvCreateUpgradePacket processes packet reception
func (k Keeper) OnRecvCreateUpgradePacket(ctx sdk.Context, packet channeltypes.Packet, data types.CreateUpgradePacketData) (packetAck types.CreateUpgradePacketAck, err error) {
	// validate packet data upon receiving
	if err := data.ValidateBasic(); err != nil {
		return packetAck, err
	}

	// Verify the channel connects to the provider
	clientID, err := k.consumerKeeper.GetProviderClientID(ctx)
	if err != nil {
		return
	}
	err = k.verifyChannel(ctx, clientID, packet.DestinationChannel)
	if err != nil {
		return
	}

	_, err = k.upgradeKeeper.GetUpgradePlan(ctx)
	if err == nil || !errors.Is(err, upgradetypes.ErrNoUpgradePlanFound) {
		return packetAck, errors.New("existing upgrade plan found")
	}
	err = k.upgradeKeeper.ScheduleUpgrade(ctx, upgradetypes.Plan{
		Name:   data.Name,
		Height: int64(data.Height),
		Info:   data.Info,
	})
	if err != nil {
		return packetAck, err
	}

	return packetAck, nil
}

// OnRecvCancelUpgradePacket processes packet reception
func (k Keeper) OnRecvCancelUpgradePacket(ctx sdk.Context, packet channeltypes.Packet, data types.CancelUpgradePacketData) (packetAck types.CancelUpgradePacketAck, err error) {
	// validate packet data upon receiving
	if err := data.ValidateBasic(); err != nil {
		return packetAck, err
	}

	// Verify the channel connects to the provider
	clientID, err := k.consumerKeeper.GetProviderClientID(ctx)
	if err != nil {
		return
	}
	err = k.verifyChannel(ctx, clientID, packet.DestinationChannel)
	if err != nil {
		return
	}

	plan, err := k.upgradeKeeper.GetUpgradePlan(ctx)
	if err != nil {
		if errors.Is(err, upgradetypes.ErrNoUpgradePlanFound) {
			//NOTE: Returning a nil error allows to clear an invalid upgrade on the provider.
			return packetAck, nil
		}
		return packetAck, err
	}
	if plan.Name != data.Plan {
		return packetAck, fmt.Errorf("plan does not match: %s != %s", plan.Name, data.Plan)
	}

	err = k.upgradeKeeper.ClearUpgradePlan(ctx)
	if err != nil {
		return packetAck, err
	}
	k.Logger(ctx).Debug(fmt.Sprintf("upgrade plan %s canceled: %+v", plan.Name, plan))

	return packetAck, nil
}

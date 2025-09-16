package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"

	"github.com/sagaxyz/saga-sdk/x/chainlet/types"
)

// TransmitConfirmUpgradePacket transmits the packet over IBC with the specified source port and source channel
func (k Keeper) TransmitConfirmUpgradePacket(
	ctx sdk.Context,
	packetData types.ConfirmUpgradePacketData,
	sourcePort,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) (uint64, error) {
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return 0, errorsmod.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	packetBytes, err := packetData.GetBytes()
	if err != nil {
		return 0, errorsmod.Wrapf(sdkerrors.ErrJSONMarshal, "cannot marshal the packet: %s", err)
	}

	return k.ibcKeeperFn().ChannelKeeper.SendPacket(ctx, channelCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, packetBytes)
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

	cstore := ctx.KVStore(k.storeKey)
	cstore.Set([]byte("test-key"), []byte("123"))

	return packetAck, nil

	_, err = k.upgradeKeeper.GetUpgradePlan(ctx)
	if err == nil || !errors.Is(err, upgradetypes.ErrNoUpgradePlanFound) {
		return packetAck, errors.New("existing upgrade plan found")
	}
	//err = k.upgradeKeeper.ScheduleUpgrade(ctx, upgradetypes.Plan{
	err = k.ScheduleUpgrade(ctx, upgradetypes.Plan{
		Name:   data.Name,
		Height: int64(data.Height),
		Info:   data.Info,
	})
	if err != nil {
		return packetAck, err
	}
	plan, err := k.upgradeKeeper.GetUpgradePlan(ctx)
	if err != nil {
		return packetAck, errors.New("upgrade plan not found")
	}
	k.Logger(ctx).Debug(fmt.Sprintf("upgrade plan %s created: %+v", plan.Name, plan))

	return packetAck, nil
}

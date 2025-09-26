package keeper

import (
	"context"
	"errors"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ccvtypes "github.com/cosmos/interchain-security/v7/x/ccv/types"

	"github.com/sagaxyz/saga-sdk/x/chainlet/types"
)

func (k *Keeper) getConsumerConnectionID(ctx sdk.Context) (connectionID string, err error) {
	ccvChannelID, found := k.consumerKeeper.GetProviderChannel(ctx)
	if !found {
		err = errors.New("channel ID for consumer not found")
		return
	}
	ccvChannel, found := k.channelKeeper.GetChannel(ctx, ccvtypes.ConsumerPortID, ccvChannelID)
	if !found {
		err = fmt.Errorf("consumer channel %s not found", ccvChannelID)
		return
	}
	if len(ccvChannel.ConnectionHops) == 0 {
		err = fmt.Errorf("no connections for channel %s", ccvChannelID)
		return
	}
	connectionID = ccvChannel.ConnectionHops[0]
	return
}

func (k Keeper) Send(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Send only for the last block before an upgrade
	plan, err := k.upgradeKeeper.GetUpgradePlan(ctx)
	if err != nil {
		if errors.Is(err, upgradetypes.ErrNoUpgradePlanFound) {
			return nil
		}
		return err
	}
	if sdkCtx.BlockHeight() < plan.Height-1 {
		k.Logger(sdkCtx).Debug(fmt.Sprintf("skipping until the upgrade height is reached: %d >= %d", plan.Height-1, sdkCtx.BlockHeight()))
		return nil
	}

	// Find a channel for the provider chain
	var sourceChannel *channeltypes.IdentifiedChannel
	ccvConnectionID, err := k.getConsumerConnectionID(sdkCtx)
	if err != nil {
		return err
	}
	channels := k.channelKeeper.GetAllChannelsWithPortPrefix(sdkCtx, types.PortID)
	for _, channel := range channels {
		if channel.State != channeltypes.OPEN {
			continue
		}
		if len(channel.ConnectionHops) == 0 || channel.ConnectionHops[0] != ccvConnectionID {
			continue
		}

		sourceChannel = &channel
	}
	if sourceChannel == nil {
		return errors.New("no channel open")
	}

	// Create the packet data
	packetData := types.ConfirmUpgradePacketData{
		ChainId: sdkCtx.ChainID(),
		Height:  uint64(sdkCtx.BlockHeight()),
		Plan:    plan.Name,
	}
	err = packetData.ValidateBasic()
	if err != nil {
		return err
	}

	// Timeout
	connEnd, found := k.connectionKeeper.GetConnection(sdkCtx, ccvConnectionID)
	if !found {
		return fmt.Errorf("connection %s not found", ccvConnectionID)
	}
	latestHeight := k.clientKeeper.GetClientLatestHeight(sdkCtx, connEnd.ClientId)
	p := k.GetParams(sdkCtx)
	timeoutHeight := clienttypes.Height{
		RevisionNumber: latestHeight.GetRevisionNumber(),
		RevisionHeight: latestHeight.GetRevisionHeight() + p.TimeoutHeight,
	}
	timeoutTimestamp := uint64(sdkCtx.BlockTime().Add(p.TimeoutTime).UnixNano())

	_, err = k.TransmitConfirmUpgradePacket(sdkCtx, packetData, types.PortID, sourceChannel.ChannelId, timeoutHeight, timeoutTimestamp)
	if err != nil {
		return err
	}
	k.Logger(sdkCtx).Info("sent IBC message about reaching the upgrade height for the current plan")
	return nil
}

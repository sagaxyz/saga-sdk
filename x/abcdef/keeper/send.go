package keeper

import (
	"context"
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	"github.com/sagaxyz/saga-sdk/x/abcdef/types"
)

func (k Keeper) Send(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	////
	channels := k.channelKeeper.GetAllChannelsWithPortPrefix(sdkCtx, types.PortID)
	if len(channels) == 0 {
		return errors.New("no channels")
	}
	//TODO check correct client for the provider chain
	//TODO check open
	sourceChannel := channels[0]

	timeoutHeight := clienttypes.Height{
		RevisionNumber: 1,
		RevisionHeight: uint64(sdkCtx.BlockHeight()) + 1000,
	}
	timeoutTimestamp := uint64(sdkCtx.BlockTime().Add(24 * time.Hour).Unix())
	////

	height := sdkCtx.BlockHeight()
	packetData := types.ConfirmUpgradePacketData{
		ChainId: sdkCtx.ChainID(),
		Height:  uint64(height),
		//Plan: //TODO
	}
	err := packetData.ValidateBasic()
	if err != nil {
		return err
	}

	_, err = k.TransmitConfirmUpgradePacket(sdkCtx, packetData, types.PortID, sourceChannel.ChannelId, timeoutHeight, timeoutTimestamp)
	if err != nil {
		return err
	}

	return nil
}

package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/sagaxyz/saga-sdk/x/abcdef/types"
)

func (k msgServer) SendUpgrade(goCtx context.Context, msg *types.MsgSendUpgrade) (*types.MsgSendUpgradeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: logic before transmitting the packet

	// Construct the packet
	packet := types.UpgradePacketData{
		Height: msg.Height,
	}

	// Transmit the packet
	_, err := k.TransmitUpgradePacket(
		ctx,
		packet,
		msg.Port,
		msg.ChannelID,
		clienttypes.ZeroHeight(),
		msg.TimeoutTimestamp,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgSendUpgradeResponse{}, nil
}

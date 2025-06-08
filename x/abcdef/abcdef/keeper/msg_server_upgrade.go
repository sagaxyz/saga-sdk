package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"

	"github.com/sagaxyz/saga-sdk/x/abcdef/types"
)

func (k msgServer) SendUpgrade(goCtx context.Context, msg *types.MsgSendUpgrade) (*types.MsgSendUpgradeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.icaKeeper.RegisterInterchainAccount(ctx, connectionID, authtypes.NewModuleAddress(types.ModuleName).String(), "1")
	if err != nil {
		return nil, err
	}

	// Construct the packet
	packet := types.UpgradePacketData{
		Height: msg.Height,
	}

	// Transmit the packet
	_, err = k.TransmitUpgradePacket(
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

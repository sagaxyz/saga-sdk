package types

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
)

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(sdk.Context, []byte, interface{})
	Set(sdk.Context, []byte, interface{})
}

type UpgradeKeeper interface {
	GetUpgradePlan(context.Context) (upgradetypes.Plan, error)
	ScheduleUpgrade(context.Context, upgradetypes.Plan) error
	ClearUpgradePlan(context.Context) error
}

type ConsumerKeeper interface {
	GetProviderChannel(ctx sdk.Context) (string, bool)
	GetProviderClientID(ctx sdk.Context) (string, error)
}

type ClientKeeper interface {
	GetClientState(sdk.Context, string) (ibcexported.ClientState, bool)
	GetClientLatestHeight(sdk.Context, string) ibcclienttypes.Height
}

type ConnectionKeeper interface {
	GetConnection(sdk.Context, string) (ibcconnectiontypes.ConnectionEnd, bool)
}

// ChannelKeeper defines the expected IBC channel keeper.
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool)
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	SendPacket(
		ctx sdk.Context,
		sourcePort string,
		sourceChannel string,
		timeoutHeight clienttypes.Height,
		timeoutTimestamp uint64,
		data []byte,
	) (uint64, error)
	ChanCloseInit(ctx sdk.Context, portID, channelID string) error
	GetAllChannelsWithPortPrefix(ctx sdk.Context, portPrefix string) []channeltypes.IdentifiedChannel
}

package types

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
)

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(sdk.Context, []byte, interface{})
	Set(sdk.Context, []byte, interface{})
}

type UpgradeKeeper interface {
	GetUpgradePlan(context.Context) (upgradetypes.Plan, error)
	//ScheduleUpgrade(context.Context, upgradetypes.Plan) error
	GetDoneHeight(context.Context, string) (int64, error)
	ClearUpgradePlan(context.Context) error
}

type ConsumerKeeper interface {
	GetProviderChannel(ctx sdk.Context) (string, bool)
}

type ClientKeeper interface {
	GetClientState(sdk.Context, string) (ibcexported.ClientState, bool)
	GetClientLatestHeight(sdk.Context, string) ibcclienttypes.Height
}

type ConnectionKeeper interface {
	GetConnection(sdk.Context, string) (ibcconnectiontypes.ConnectionEnd, bool)
}

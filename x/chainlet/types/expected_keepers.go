package types

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(sdk.Context, []byte, interface{})
	Set(sdk.Context, []byte, interface{})
}

type UpgradeKeeper interface {
	GetUpgradePlan(ctx context.Context) (plan upgradetypes.Plan, err error)
}

type ConsumerKeeper interface {
	GetProviderChannel(ctx sdk.Context) (string, bool)
}

type ClientKeeper interface {
	GetClientState(sdk.Context, string) (ibcexported.ClientState, bool)
}

type ConnectionKeeper interface {
	GetConnection(sdk.Context, string) (ibcconnectiontypes.ConnectionEnd, bool)
}

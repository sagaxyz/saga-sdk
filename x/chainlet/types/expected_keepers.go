package types

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

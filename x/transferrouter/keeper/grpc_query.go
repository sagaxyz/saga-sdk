package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

// Ensure Keeper implements the expected gRPC query interface (currently empty).
var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

// Params returns the current module parameters. This is compatible with the common
// `{module}/Params` query pattern used across the SDK.
func (k Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

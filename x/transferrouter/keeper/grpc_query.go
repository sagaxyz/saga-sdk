package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

// Params returns the current module parameters.
func (k Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

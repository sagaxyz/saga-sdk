package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	k Keeper
}

// NewQuerier creates a new Querier instance
func NewQuerier(k Keeper) Querier {
	return Querier{k: k}
}

// Params returns the current module parameters.
func (q Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params, err := q.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: params}, nil
}

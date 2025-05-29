package keeper

import (
	"context"

	"github.com/sagaxyz/saga-sdk/x/assetctl/host/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, so server doesn't implement their interface
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &Querier{Keeper: keeper}
}

func (k Querier) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := k.Keeper.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: params}, nil
}

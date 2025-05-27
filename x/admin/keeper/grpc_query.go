package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/admin/types"
)

var _ types.QueryServer = Keeper{}

// Params returns the params of the claim module
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

func (k Keeper) Superuser(c context.Context, _ *types.QuerySuperuserRequest) (*types.QuerySuperuserResponse, error) {
	return &types.QuerySuperuserResponse{
		Superuser: k.GetAuthority(),
	}, nil
}

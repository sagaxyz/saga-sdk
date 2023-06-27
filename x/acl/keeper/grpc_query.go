package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sagaxyz/saga-sdk/x/acl/types"
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

func (k Keeper) ListAdmins(c context.Context, req *types.QueryListAdminsRequest) (*types.QueryListAdminsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryListAdminsResponse{
		Admins: k.ExportAdmins(ctx),
	}, nil
}

func (k Keeper) ListAllowed(c context.Context, req *types.QueryListAllowedRequest) (*types.QueryListAllowedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryListAllowedResponse{
		Allowed: k.ExportAllowed(ctx),
	}, nil
}

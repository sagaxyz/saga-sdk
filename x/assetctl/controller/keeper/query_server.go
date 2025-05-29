package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// AssetDirectory implements types.QueryServer.
func (k Querier) AssetDirectory(ctx context.Context, req *types.QueryAssetDirectoryRequest) (*types.QueryAssetDirectoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	_ = sdkCtx

	return &types.QueryAssetDirectoryResponse{
		// Asset: posts,
		// Pagination: pageRes,
	}, nil
}

// ChainletRegistryStatus implements types.QueryServer.
func (k Querier) ChainletRegistryStatus(context.Context, *types.QueryChainletRegistryStatusRequest) (*types.QueryChainletRegistryStatusResponse, error) {
	panic("unimplemented")
}

// Params implements types.QueryServer.
func (k Querier) Params(context.Context, *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	panic("unimplemented")
}

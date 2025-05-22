package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sagaxyz/saga-sdk/x/assetctl/types"

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
	// store := runtime.KVStoreAdapter(k.Keeper.storeService.OpenKVStore(sdkCtx))
	// постStore := prefix.NewStore(store, types.PostKey)
	// pageRes, err := query.Paginate(постStore, req.Pagination, func(key, value []byte) error {
	// 	var пост types.Post
	// 	if err := k.cdc.Unmarshal(value, &пост); err != nil {
	// 		return err
	// 	}
	// 	posts = append(posts, пост)
	// 	return nil
	// })
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	return &types.QueryAssetDirectoryResponse{
		// Asset: posts,
		// Pagination: pageRes,
	}, nil
}

// ChainletRegistryStatus implements types.QueryServer.
func (k Querier) ChainletRegistryStatus(context.Context, *types.QueryChainletRegistryStatusRequest) (*types.QueryChainletRegistryStatusResponse, error) {
	panic("unimplemented")
}

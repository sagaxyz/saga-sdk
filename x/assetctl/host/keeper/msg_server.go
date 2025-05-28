package keeper

import (
	"context"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/sagaxyz/saga-sdk/x/assetctl/host/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// RegisterAssets implements types.MsgServer.
func (k msgServer) RegisterAssets(goCtx context.Context, msg *types.MsgRegisterAssets) (*types.MsgRegisterAssetsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Iterate through msg.AssetsToRegister
	// For each asset:
	// 1. Determine its IBC denom (e.g., types.GetIBCDenom)
	// 2. Check for uniqueness in the keeper's asset directory (EnabledList)
	// 3. If unique, create a types.RegisteredAsset from msg.AssetDetails
	// 4. Store it: k.EnabledList.Set(ctx, ibcDenom)
	//    (Note: EnabledList currently stores KeySet[string]. You might need to store the full RegisteredAsset. This might mean changing EnabledList to a collections.Map[string, types.RegisteredAsset] or creating a new Map for the full asset details and keeping EnabledList for quick lookups of allowed denoms.)
	// 5. Emit event
	_ = ctx

	return &types.MsgRegisterAssetsResponse{}, nil
}

// UnregisterAssets implements types.MsgServer.
func (k msgServer) UnregisterAssets(goCtx context.Context, msg *types.MsgUnregisterAssets) (*types.MsgUnregisterAssetsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Iterate through msg.IbcDenomsToUnregister
	// For each ibcDenom:
	// 1. Check if it exists in k.EnabledList
	// 2. If it exists, remove it: k.EnabledList.Delete(ctx, ibcDenom)
	//    (And remove from the full RegisteredAsset map if you created one)
	// 3. Emit event
	_ = ctx

	return &types.MsgUnregisterAssetsResponse{}, nil
}

// ToggleChainletRegistry implements types.MsgServer.
func (k msgServer) ToggleChainletRegistry(ctx context.Context, msg *types.MsgToggleChainletRegistry) (*types.MsgToggleChainletRegistryResponse, error) {
	// TODO: Add authority check - who can toggle this? Chainlet's ICA controller? Specific admin key?
	// For now, assuming the msg.Creator is authorized and msg.ChainletId is the target.

	if msg.ChainletId == "" {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("chainlet_id cannot be empty")
	}

	// TODO: emit an event

	if msg.Enable {
		err := k.EnabledList.Set(ctx, msg.ChainletId)
		return &types.MsgToggleChainletRegistryResponse{}, err
	}

	err := k.EnabledList.Remove(ctx, msg.ChainletId)
	return &types.MsgToggleChainletRegistryResponse{}, err
}

// SupportAsset implements types.MsgServer.
func (k msgServer) SupportAsset(context.Context, *types.MsgSupportAsset) (*types.MsgSupportAssetResponse, error) {
	panic("unimplemented")
}

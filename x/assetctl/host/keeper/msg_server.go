package keeper

import (
	"context"

	"github.com/sagaxyz/saga-sdk/x/assetctl/host/types"
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

// RegisterDenom implements types.MsgServer.
func (k msgServer) RegisterDenom(context.Context, *types.MsgRegisterDenom) (*types.MsgRegisterDenomResponse, error) {
	panic("unimplemented")
}

// UpdateParams implements types.MsgServer.
func (k msgServer) UpdateParams(context.Context, *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	panic("unimplemented")
}

// SupportAsset implements types.MsgServer.
func (k msgServer) SupportAsset(context.Context, *types.MsgSupportAsset) (*types.MsgSupportAssetResponse, error) {
	panic("unimplemented")
}

package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/errors"
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
func (k msgServer) RegisterDenom(ctx context.Context, msg *types.MsgRegisterDenom) (*types.MsgRegisterDenomResponse, error) {
	if msg.Authority != k.Authority {
		return nil, errors.ErrUnauthorized.Wrap("the signer is not the authority")
	}

	if msg.IbcDenom == "" {
		return nil, errors.ErrInvalidRequest.Wrap("ibc_denom cannot be empty")
	}

	// TODO: Add logic to register the denom as a native asset
	// This would typically involve storing the denom in the keeper's state

	return &types.MsgRegisterDenomResponse{}, nil
}

// UpdateParams implements types.MsgServer.
func (k msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if msg.Authority != k.Authority {
		return nil, errors.ErrUnauthorized.Wrap("the signer is not the authority")
	}

	if msg.Params == nil {
		return nil, errors.ErrInvalidRequest.Wrap("params cannot be nil")
	}

	err := k.Params.Set(ctx, *msg.Params)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// SupportAsset implements types.MsgServer.
func (k msgServer) SupportAsset(ctx context.Context, msg *types.MsgSupportAsset) (*types.MsgSupportAssetResponse, error) {
	if msg.Authority != k.Authority {
		return nil, errors.ErrUnauthorized.Wrap("the signer is not the authority")
	}

	if msg.IbcDenom == "" {
		return nil, errors.ErrInvalidRequest.Wrap("ibc_denom cannot be empty")
	}

	// TODO: Add logic to notify the controller that this host supports the asset
	// This would typically involve sending a message to the controller module

	return &types.MsgSupportAssetResponse{}, nil
}

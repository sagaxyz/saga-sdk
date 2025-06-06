package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	ibccontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"

	controllertypes "github.com/sagaxyz/saga-sdk/x/assetctl/controller/types"
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
func (k msgServer) RegisterDenoms(ctx context.Context, msg *types.MsgRegisterDenoms) (*types.MsgRegisterDenomsResponse, error) {
	addr, err := k.addressCodec.StringToBytes(msg.Authority)
	if err != nil {
		return nil, err
	}

	if !k.aclKeeper.Admin(sdk.UnwrapSDKContext(ctx), addr) {
		return nil, errors.ErrUnauthorized.Wrap("the signer is not an admin")
	}

	if len(msg.IbcDenoms) == 0 {
		return nil, errors.ErrInvalidRequest.Wrap("ibc_denoms cannot be empty")
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	assetsToRegister := make([]controllertypes.AssetDetails, len(msg.IbcDenoms))
	for i, denom := range msg.IbcDenoms {
		assetsToRegister[i] = controllertypes.AssetDetails{
			// TODO: Add logic to get the asset details
			Denom: denom,
		}
	}

	controllerMsg := &controllertypes.MsgRegisterAssets{
		Authority:        k.Authority,
		AssetsToRegister: assetsToRegister,
	}

	controllerMsgBytes, err := controllerMsg.Marshal()
	if err != nil {
		return nil, err
	}

	wrapperMsg := &ibccontrollertypes.MsgSendTx{
		Owner:           k.Authority,
		ConnectionId:    params.HubConnectionId,
		RelativeTimeout: icatypes.DefaultRelativePacketTimeoutTimestamp,
		PacketData: icatypes.InterchainAccountPacketData{
			Type: icatypes.EXECUTE_TX,
			Data: controllerMsgBytes,
		},
	}

	handler := k.router.Handler(wrapperMsg)
	if handler == nil {
		return nil, errors.ErrUnknownRequest.Wrapf("unrecognized message route: %s", sdk.MsgTypeURL(msg))
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	msgResp, err := handler(sdkCtx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to execute message; message %v", msg)
	}

	// We expect a single response, check if there is only one and return it
	if len(msgResp.MsgResponses) != 1 {
		return nil, fmt.Errorf("expected a single response, got %d", len(msgResp.MsgResponses))
	}

	return &types.MsgRegisterDenomsResponse{MsgResponse: msgResp.MsgResponses[0]}, nil
}

// UpdateParams implements types.MsgServer.
func (k msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	addr, err := k.addressCodec.StringToBytes(msg.Authority)
	if err != nil {
		return nil, err
	}

	if !k.aclKeeper.Admin(sdk.UnwrapSDKContext(ctx), addr) {
		return nil, errors.ErrUnauthorized.Wrap("the signer is not an admin")
	}

	if msg.Params == nil {
		return nil, errors.ErrInvalidRequest.Wrap("params cannot be nil")
	}

	err = k.Params.Set(ctx, *msg.Params)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// SupportAsset implements types.MsgServer.
func (k msgServer) SupportAssets(ctx context.Context, msg *types.MsgSupportAssets) (*types.MsgSupportAssetsResponse, error) {
	addr, err := k.addressCodec.StringToBytes(msg.Authority)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if !k.aclKeeper.Admin(sdkCtx, addr) {
		return nil, errors.ErrUnauthorized.Wrap("the signer is not an admin")
	}

	if len(msg.IbcDenoms) == 0 {
		return nil, errors.ErrInvalidRequest.Wrap("ibc_denoms cannot be empty")
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	controllerMsg := &controllertypes.MsgSupportAssets{
		Authority: k.Authority,
		ChannelId: params.HubChannelId,
		IbcDenoms: msg.IbcDenoms,
	}

	controllerMsgBytes, err := controllerMsg.Marshal()
	if err != nil {
		return nil, err
	}

	wrapperMsg := &ibccontrollertypes.MsgSendTx{
		Owner:           k.Authority,
		ConnectionId:    params.HubConnectionId,
		RelativeTimeout: icatypes.DefaultRelativePacketTimeoutTimestamp,
		PacketData: icatypes.InterchainAccountPacketData{
			Type: icatypes.EXECUTE_TX,
			Data: controllerMsgBytes,
		},
	}

	handler := k.router.Handler(wrapperMsg)
	if handler == nil {
		return nil, errors.ErrUnknownRequest.Wrapf("unrecognized message route: %s", sdk.MsgTypeURL(msg))
	}

	msgResp, err := handler(sdkCtx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to execute message; message %v", msg)
	}

	// We expect a single response, check if there is only one and return it
	if len(msgResp.MsgResponses) != 1 {
		return nil, fmt.Errorf("expected a single response, got %d", len(msgResp.MsgResponses))
	}

	return &types.MsgSupportAssetsResponse{MsgResponse: msgResp.MsgResponses[0]}, nil
}

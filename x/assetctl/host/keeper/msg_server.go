package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	ibccontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	controllertypes "github.com/sagaxyz/saga-sdk/x/assetctl/controller/types"
	"github.com/sagaxyz/saga-sdk/x/assetctl/host/types"
	assetctltypes "github.com/sagaxyz/saga-sdk/x/assetctl/types"
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

	owner := k.accountKeeper.GetModuleAddress(assetctltypes.ModuleName)
	if owner == nil {
		return nil, fmt.Errorf("owner address not found")
	}

	wrapperMsg := &ibccontrollertypes.MsgSendTx{
		Owner:           owner.String(),
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

	owner := k.accountKeeper.GetModuleAddress(assetctltypes.ModuleName)
	if owner == nil {
		return nil, fmt.Errorf("owner address not found")
	}

	wrapperMsg := &ibccontrollertypes.MsgSendTx{
		Owner:           owner.String(),
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

// CreateICAOnHub is a helper function to create an ICA on the hub, should be used only once.
func (k msgServer) CreateICAOnHub(ctx context.Context, msg *types.MsgCreateICAOnHub) (*types.MsgCreateICAOnHubResponse, error) {
	// check if the ica on hub already exists, if it does, return an error
	has, err := k.ICAOnHub.Has(ctx)
	if err != nil {
		return nil, err
	}
	if has {
		return nil, errors.ErrInvalidRequest.Wrap("ICA on hub already exists")
	}

	// check if the signer is an admin
	addr, err := k.addressCodec.StringToBytes(msg.Authority)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if !k.aclKeeper.Admin(sdkCtx, addr) {
		return nil, errors.ErrUnauthorized.Wrap("the signer is not an admin")
	}

	// get the assetctl module address
	hostAddress := k.accountKeeper.GetModuleAddress(assetctltypes.ModuleName)
	if hostAddress == nil {
		return nil, fmt.Errorf("host address not found")
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	registerMsg := &icacontrollertypes.MsgRegisterInterchainAccount{
		Owner:        hostAddress.String(),
		ConnectionId: params.HubConnectionId,
		Version:      icatypes.Version,
		Ordering:     channeltypes.UNORDERED,
	}

	handler := k.router.Handler(registerMsg)
	res, err := handler(sdkCtx, registerMsg)
	if err != nil {
		return nil, err
	}

	if len(res.MsgResponses) != 1 {
		return nil, fmt.Errorf("expected a single response, got %d", len(res.MsgResponses))
	}

	// get the response as icacontrollertypes.MsgRegisterInterchainAccountResponse
	registerResponse := icacontrollertypes.MsgRegisterInterchainAccountResponse{}
	err = k.cdc.UnpackAny(res.MsgResponses[0], &registerResponse)
	if err != nil {
		return nil, err
	}

	err = k.ICAOnHub.Set(sdkCtx, types.ICAOnHub{
		ChannelId: registerResponse.ChannelId,
		PortId:    registerResponse.PortId,
	})

	if err != nil {
		return nil, err
	}

	return &types.MsgCreateICAOnHubResponse{}, nil
}

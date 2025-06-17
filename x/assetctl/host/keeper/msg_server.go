package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	proto "github.com/cosmos/gogoproto/proto"
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
func (k msgServer) ManageRegisteredAssets(ctx context.Context, msg *types.MsgManageRegisteredAssets) (*types.MsgManageRegisteredAssetsResponse, error) {
	addr, err := k.addressCodec.StringToBytes(msg.Authority)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if !k.aclKeeper.Admin(sdkCtx, addr) {
		return nil, errors.ErrUnauthorized.Wrap("the signer is not an admin")
	}

	if len(msg.AssetsToRegister) == 0 && len(msg.AssetsToUnregister) == 0 {
		return nil, errors.ErrInvalidRequest.Wrap("assets_to_register and assets_to_unregister cannot be empty")
	}

	assetsToRegister := make([]banktypes.Metadata, len(msg.AssetsToRegister))
	for i, denom := range msg.AssetsToRegister {
		meta, ok := k.BankKeeper.GetDenomMetaData(sdkCtx, denom)
		if !ok {
			return nil, errors.ErrInvalidRequest.Wrapf("denom %s not found", denom)
		}
		assetsToRegister[i] = meta
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	moduleAddress := k.AccountKeeper.GetModuleAddress(assetctltypes.ModuleName)
	if moduleAddress == nil {
		return nil, fmt.Errorf("module address not found")
	}

	sequence, err := k.sendMsgThroughICA(ctx, &controllertypes.MsgManageRegisteredAssets{
		Authority:          moduleAddress.String(),
		ChannelId:          params.HubChannelId,
		AssetsToRegister:   assetsToRegister,
		AssetsToUnregister: msg.AssetsToUnregister, // we don't need to send any metadata here, just the denom
	})
	if err != nil {
		return nil, err
	}

	err = k.InFlightRequests.Set(ctx, sequence, sdk.MsgTypeURL(msg))
	if err != nil {
		return nil, err
	}

	return &types.MsgManageRegisteredAssetsResponse{Sequence: sequence}, nil
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
func (k msgServer) ManageSupportedAssets(ctx context.Context, msg *types.MsgManageSupportedAssets) (*types.MsgManageSupportedAssetsResponse, error) {
	addr, err := k.addressCodec.StringToBytes(msg.Authority)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if !k.aclKeeper.Admin(sdkCtx, addr) {
		return nil, errors.ErrUnauthorized.Wrap("the signer is not an admin")
	}

	if len(msg.AddIbcDenoms) == 0 && len(msg.RemoveIbcDenoms) == 0 {
		return nil, errors.ErrInvalidRequest.Wrap("add_ibc_denoms and remove_ibc_denoms cannot be empty")
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	sequence, err := k.sendMsgThroughICA(ctx, &controllertypes.MsgManageSupportedAssets{
		Authority:       k.Authority,
		ChannelId:       params.HubChannelId,
		AddIbcDenoms:    msg.AddIbcDenoms,
		RemoveIbcDenoms: msg.RemoveIbcDenoms,
	})
	if err != nil {
		return nil, err
	}

	err = k.InFlightRequests.Set(ctx, sequence, sdk.MsgTypeURL(msg))
	if err != nil {
		return nil, err
	}

	return &types.MsgManageSupportedAssetsResponse{Sequence: sequence}, nil
}

// CreateICAOnHub is a helper function to create an ICA on the hub, should be used only once.
func (k msgServer) CreateICAOnHub(ctx context.Context, msg *types.MsgCreateICAOnHub) (*types.MsgCreateICAOnHubResponse, error) {
	// check if the signer is an admin
	addr, err := k.addressCodec.StringToBytes(msg.Authority)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if !k.aclKeeper.Admin(sdkCtx, addr) {
		return nil, errors.ErrUnauthorized.Wrap("the signer is not an admin")
	}

	// check if the ica on hub already exists, if it does, return an error
	has, err := k.ICAData.Has(ctx)
	if err != nil {
		return nil, err
	}
	if has {
		return nil, errors.ErrInvalidRequest.Wrap("ICA on hub already exists")
	}

	// get the assetctl module address
	hostAddress := k.AccountKeeper.GetModuleAddress(assetctltypes.ModuleName)
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
	registerResponse := &icacontrollertypes.MsgRegisterInterchainAccountResponse{}
	err = k.cdc.UnpackAny(res.MsgResponses[0], &registerResponse)
	if err != nil {
		return nil, err
	}

	err = k.ICAData.Set(sdkCtx, types.ICAOnHub{
		ChannelId: registerResponse.ChannelId,
		PortId:    registerResponse.PortId,
	})

	if err != nil {
		return nil, err
	}

	return &types.MsgCreateICAOnHubResponse{
		ChannelId: registerResponse.ChannelId,
		PortId:    registerResponse.PortId,
	}, nil
}

func (k msgServer) sendMsgThroughICA(ctx context.Context, msg proto.Marshaler) (uint64, error) {
	// Check if the ICA exists and is active
	params, err := k.Params.Get(ctx)
	if err != nil {
		return 0, err
	}

	icaData, err := k.ICAData.Get(ctx)
	if err != nil {
		return 0, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	_, ok := k.icaControllerKeeper.GetInterchainAccountAddress(sdkCtx, params.HubConnectionId, icaData.PortId)
	if !ok {
		return 0, errors.ErrInvalidRequest.Wrap("ICA has not been created yet")
	}

	if !k.icaControllerKeeper.IsActiveChannel(sdkCtx, params.HubConnectionId, icaData.PortId) {
		return 0, errors.ErrInvalidRequest.Wrap("ICA channel is not active")
	}

	owner := k.AccountKeeper.GetModuleAddress(assetctltypes.ModuleName)
	if owner == nil {
		return 0, fmt.Errorf("owner address not found")
	}

	bz, err := msg.Marshal()
	if err != nil {
		return 0, err
	}

	wrapperMsg := &icacontrollertypes.MsgSendTx{
		Owner:           owner.String(),
		ConnectionId:    params.HubConnectionId,
		RelativeTimeout: icatypes.DefaultRelativePacketTimeoutTimestamp,
		PacketData: icatypes.InterchainAccountPacketData{
			Type: icatypes.EXECUTE_TX,
			Data: bz,
		},
	}

	handler := k.router.Handler(wrapperMsg)
	if handler == nil {
		return 0, errors.ErrUnknownRequest.Wrapf("unrecognized message route: %s", sdk.MsgTypeURL(wrapperMsg))
	}

	msgResp, err := handler(sdkCtx, wrapperMsg)
	if err != nil {
		return 0, fmt.Errorf("failed to execute message; message %v", msg)
	}

	// We expect a single response, check if there is only one and return it
	if len(msgResp.MsgResponses) != 1 {
		return 0, fmt.Errorf("expected a single response, got %d", len(msgResp.MsgResponses))
	}

	// parse response as MsgSendTxResponse
	sendTxResponse := &icacontrollertypes.MsgSendTxResponse{}
	err = k.cdc.UnpackAny(msgResp.MsgResponses[0], &sendTxResponse)
	if err != nil {
		return 0, err
	}

	return sendTxResponse.Sequence, nil
}

package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/types"
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
func (k msgServer) RegisterAssets(ctx context.Context, msg *types.MsgRegisterAssets) (*types.MsgRegisterAssetsResponse, error) {
	if err := k.checkChannelAuthority(ctx, msg.Authority, msg.ChannelId); err != nil {
		return nil, err
	}

	if len(msg.AssetsToRegister) == 0 {
		return nil, fmt.Errorf("assets to register cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, asset := range msg.AssetsToRegister {

		// trim the denomination prefix, by default "ibc/"
		hexHash := asset.IbcDenom[len("ibc/"):]
		hash, err := hex.DecodeString(hexHash)
		if err != nil {
			return nil, err
		}

		denomTrace, found := k.IBCTransferKeeper.GetDenomTrace(sdkCtx, hash)
		if !found {
			return nil, fmt.Errorf("denom trace not found")
		}

		// Then get the channel from the path
		pathParts := strings.Split(denomTrace.Path, "/")
		if len(pathParts) != 2 {
			return nil, fmt.Errorf("denom trace path is not valid, only 1 hop is allowed")
		}

		if pathParts[1] != msg.ChannelId {
			return nil, fmt.Errorf("denom trace channel does not match the channel id")
		}

		// TODO: check if the asset is already registered

		// We allow overwriting the asset metadata
		err = k.AssetMetadata.Set(ctx, asset.IbcDenom, types.RegisteredAsset{
			OriginalDenom: asset.Denom,
			DisplayName:   asset.DisplayName,
			Description:   asset.Description,
			DenomUnits:    asset.DenomUnits,
		})
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgRegisterAssetsResponse{}, nil
}

// UnregisterAssets implements types.MsgServer.
func (k msgServer) UnregisterAssets(ctx context.Context, msg *types.MsgUnregisterAssets) (*types.MsgUnregisterAssetsResponse, error) {
	if err := k.checkChannelAuthority(ctx, msg.Authority, msg.ChannelId); err != nil {
		return nil, err
	}

	if len(msg.IbcDenoms) == 0 {
		return nil, fmt.Errorf("ibc denoms to unregister cannot be empty")
	}

	for _, ibcDenom := range msg.IbcDenoms {
		err := k.AssetMetadata.Remove(ctx, ibcDenom)
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgUnregisterAssetsResponse{}, nil
}

// SupportAssets checks if the sender is the authority of the channel and then
// adds the assets to the supported assets list.
func (k msgServer) SupportAssets(ctx context.Context, msg *types.MsgSupportAssets) (*types.MsgSupportAssetsResponse, error) {
	if err := k.checkChannelAuthority(ctx, msg.Authority, msg.ChannelId); err != nil {
		return nil, err
	}

	if len(msg.IbcDenoms) == 0 {
		return nil, fmt.Errorf("ibc_denoms cannot be empty")
	}

	for _, ibcDenom := range msg.IbcDenoms {
		exists, err := k.SupportedAssets.Has(ctx, collections.Join(msg.ChannelId, ibcDenom))
		if err != nil {
			return nil, err
		}

		// not an error, just a warning
		if exists {
			k.logger.Debug("asset already supported", "ibc_denom", ibcDenom, "channel_id", msg.ChannelId)
			continue
		}

		err = k.SupportedAssets.Set(ctx, collections.Join(msg.ChannelId, ibcDenom))
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgSupportAssetsResponse{}, nil
}

// UpdateParams implements types.MsgServer.
func (k msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if msg.Authority != k.Authority {
		return nil, fmt.Errorf("authority does not match")
	}

	if msg.Params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	err := k.Params.Set(ctx, *msg.Params)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// checkChannelAuthority checks if the address is the authority of the channel.
// Right now it only checks if the address is from the same chainlet, but we should check
// if the address is the authority of the channel (admin).
func (k Keeper) checkChannelAuthority(ctx context.Context, address, channelId string) error {
	// TODO: this is an expensive operation, we should allow pre-registration of the
	// authorities.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	interchainAccounts := k.ICAHostKeeper.GetAllInterchainAccounts(sdkCtx)
	connectionId := ""
	portId := ""
	for _, interchainAccount := range interchainAccounts {
		if interchainAccount.AccountAddress == address {
			connectionId = interchainAccount.ConnectionId
			portId = interchainAccount.PortId
			break
		}
	}

	if connectionId == "" || portId == "" {
		return fmt.Errorf("the signer is not the authority")
	}

	channel, ok := k.IBCChannelKeeper.GetChannel(sdkCtx, portId, channelId)
	if !ok {
		return fmt.Errorf("channel not found")
	}

	if channel.GetConnectionHops()[0] != connectionId {
		return fmt.Errorf("authority does not match")
	}

	return nil
}
